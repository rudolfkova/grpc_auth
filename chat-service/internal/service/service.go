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

// ChatRepository ...
type ChatRepository interface {
	CreateChat(ctx context.Context, chatType, name string) (chatID int, createdAt time.Time, err error)
	DeleteChat(ctx context.Context, chatID int) error
	AddMember(ctx context.Context, chatID, userID int) error
	RemoveMember(ctx context.Context, chatID, userID int) error
	IsMember(ctx context.Context, chatID, userID int) (bool, error)
	GetMemberIDs(ctx context.Context, chatID int) ([]int, error)
	GetOrCreateDMChat(ctx context.Context, initiatorID, recipientID int) (chatID int, created bool, createdAt time.Time, err error)
	GetUserChats(ctx context.Context, userID, limit, offset int) ([]model.ChatPreviewDTO, error)
	ResetUnread(ctx context.Context, chatID, userID int) error
}

// MessageRepository ...
type MessageRepository interface {
	GetMessages(ctx context.Context, chatID, limit int, cursor string) ([]model.MessageDTO, error)
	SendMessage(ctx context.Context, chatID, senderID int, text string) (messageID int, createdAt time.Time, err error)
}

// Hub ...
type Hub interface {
	Push(userID int, msg *chatv1.MessageDTO)
}

// Service ...
type Service struct {
	chatRepo    ChatRepository
	messageRepo MessageRepository
	hub         Hub
}

// NewService ...
func NewService(chatRepo ChatRepository, messageRepo MessageRepository, hub Hub) *Service {
	return &Service{chatRepo: chatRepo, messageRepo: messageRepo, hub: hub}
}

// CreateChat создаёт новый чат. Вызывающий сам управляет участниками через AddMember.
func (s *Service) CreateChat(ctx context.Context, name string) (int, time.Time, error) {
	chatType := "group"
	if name == "" {
		chatType = "dm"
	}
	return s.chatRepo.CreateChat(ctx, chatType, name)
}

// DeleteChat удаляет чат. Проверка прав — на вызывающем сервисе.
func (s *Service) DeleteChat(ctx context.Context, chatID int) error {
	return s.chatRepo.DeleteChat(ctx, chatID)
}

// AddMember добавляет участника в чат.
func (s *Service) AddMember(ctx context.Context, chatID, userID int) error {
	return s.chatRepo.AddMember(ctx, chatID, userID)
}

// RemoveMember удаляет участника из чата.
func (s *Service) RemoveMember(ctx context.Context, chatID, userID int) error {
	return s.chatRepo.RemoveMember(ctx, chatID, userID)
}

// GetOrCreateChat возвращает или создаёт DM-чат между двумя пользователями.
func (s *Service) GetOrCreateChat(ctx context.Context, initiatorID, recipientID int) (int, bool, time.Time, error) {
	return s.chatRepo.GetOrCreateDMChat(ctx, initiatorID, recipientID)
}

// GetMessages возвращает историю сообщений. Требует членства в чате.
func (s *Service) GetMessages(ctx context.Context, chatID, limit int, cursor string) ([]model.MessageDTO, string, error) {
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

	var nextCursor string
	if len(messages) > limit {
		nextCursor = messages[limit].CreatedAt.UTC().Format(time.RFC3339Nano)
		messages = messages[:limit]
	}

	if err := s.chatRepo.ResetUnread(ctx, chatID, callerID); err != nil {
		return nil, "", err
	}

	return messages, nextCursor, nil
}

// GetUserChats возвращает список чатов пользователя.
func (s *Service) GetUserChats(ctx context.Context, userID, limit, offset int) ([]model.ChatPreviewDTO, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.chatRepo.GetUserChats(ctx, userID, limit, offset)
}

// SendMessage отправляет сообщение и пушит его всем участникам чата через hub.
func (s *Service) SendMessage(ctx context.Context, chatID, senderID int, text string) (int, time.Time, error) {
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

	// Пушим всем участникам чата (включая отправителя — для синхронизации
	// между устройствами, клиент сам решает показывать ли своё сообщение повторно).
	memberIDs, err := s.chatRepo.GetMemberIDs(ctx, chatID)
	if err != nil {
		// Сообщение уже сохранено — не возвращаем ошибку, просто не пушим.
		return messageID, createdAt, nil
	}

	msg := &chatv1.MessageDTO{
		Id:        int64(messageID),
		ChatId:    int64(chatID),
		SenderId:  int64(senderID),
		Text:      text,
		CreatedAt: timestamppb.New(createdAt),
	}
	for _, uid := range memberIDs {
		s.hub.Push(uid, msg)
	}

	return messageID, createdAt, nil
}
