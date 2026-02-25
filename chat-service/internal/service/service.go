// Package service ...
package service

import (
	chaterror "chat/internal/error"
	"chat/internal/interceptor"
	"chat/internal/model"
	chatv1 "chat/proto/chat/v1"
	"context"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// Service ...
type Service struct {
	chatRepo    ChatRepository
	messageRepo MessageRepository
	hub         Hub
}

// NewService ...
func NewService(chatRepo ChatRepository, messageRepo MessageRepository) *Service {
	return &Service{
		chatRepo:    chatRepo,
		messageRepo: messageRepo,
	}
}

// ChatRepository ...
type ChatRepository interface {
	// GetOrCreateChat ...
	GetOrCreateChat(ctx context.Context, initiatorID int, recipientID int) (chatID int, created bool, createdAt time.Time, err error)
	// IsMember ...
	IsMember(ctx context.Context, chatID int, userID int) (bool, error)
	// GetUserChats ...
	GetUserChats(ctx context.Context, userID int, limit int, offset int) ([]model.ChatPreviewDTO, error)
	// ResetUnread ...
	ResetUnread(ctx context.Context, chatID int, userID int) error
	// GetParticipants ...
	GetParticipants(ctx context.Context, chatID int) (user1ID int, user2ID int, err error)
}

// MessageRepository ...
type MessageRepository interface {
	// GetMessages ...
	GetMessages(ctx context.Context, chatID int, limit int, cursor string) ([]model.MassageDTO, error)
	// SendMessage ...
	SendMessage(ctx context.Context, chatID int, senderID int, text string) (messageID int, createdAt time.Time, err error)
}

// Hub ...
type Hub interface {
	// Push ...
	Push(userID int, msg *chatv1.MessageDTO)
}

// GetOrCreateChat ...
func (s *Service) GetOrCreateChat(ctx context.Context, initiatorID int, recipientID int) (chatID int, created bool, createdAt time.Time, err error) {
	return s.chatRepo.GetOrCreateChat(ctx, initiatorID, recipientID)
}

// GetMessages ...
func (s *Service) GetMessages(ctx context.Context, chatID int, limit int, cursor string) (massages []model.MassageDTO, nextCursor string, err error) {
	callerID, ok := ctx.Value(interceptor.UserIDKey).(int)
	if !ok {
		return nil, "", chaterror.ErrUnauthenticated
	}

	isMember, err := s.chatRepo.IsMember(ctx, chatID, callerID)
	if err != nil {
		return nil, "", err
	}
	if !isMember {
		return nil, "", chaterror.ErrPermissionDenied
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	messages, err := s.messageRepo.GetMessages(ctx, chatID, limit+1, cursor)
	if err != nil {
		return nil, "", err
	}

	var nextCurs string
	if len(messages) > limit {
		nextCurs = messages[limit].CreatedAt.UTC().Format(time.RFC3339Nano)
		messages = messages[:limit]
	}

	if err := s.chatRepo.ResetUnread(ctx, chatID, callerID); err != nil {
		return nil, "", err
	}

	return messages, nextCurs, nil
}

// GetUserChats ...
func (s *Service) GetUserChats(ctx context.Context, userID int, limit int, offset int) (chats []model.ChatPreviewDTO, err error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	chat, err := s.chatRepo.GetUserChats(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	return chat, nil
}

// SendMessage ...
func (s *Service) SendMessage(ctx context.Context, chatID int, senderID int, text string) (int, time.Time, error) {
	isMember, err := s.chatRepo.IsMember(ctx, chatID, senderID)
	if err != nil {
		return 0, time.Time{}, err
	}
	if !isMember {
		return 0, time.Time{}, chaterror.ErrPermissionDenied
	}

	messageID, createdAt, err := s.messageRepo.SendMessage(ctx, chatID, senderID, text)
	if err != nil {
		return 0, time.Time{}, err
	}

	user1ID, user2ID, err := s.chatRepo.GetParticipants(ctx, chatID)
	if err != nil {
		return 0, time.Time{}, err
	}

	msg := &chatv1.MessageDTO{
		Id:        int64(messageID),
		ChatId:    int64(chatID),
		SenderId:  int64(senderID),
		Text:      text,
		CreatedAt: timestamppb.New(createdAt),
	}

	s.hub.Push(user1ID, msg)
	s.hub.Push(user2ID, msg)

	return messageID, createdAt, nil
}
