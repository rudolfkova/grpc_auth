// Package service ...
package service

import (
	"chat/internal/model"
	"context"
	"time"
)

// Service ...
type Service struct{}

// NewService ...
func NewService() *Service {
	return &Service{}
}

// GetOrCreateChat ...
func (s *Service) GetOrCreateChat(ctx context.Context, initiatorID int, recipientID int) (chatID int, created bool, createdAt time.Time, err error) {
	_ = ctx
	_ = initiatorID
	_ = recipientID
	return 0, false, time.Now(), nil
}

// GetMessages ...
func (s *Service) GetMessages(ctx context.Context, chatID int, limit int, cursor string) (massages []model.MassageDTO, nextCursor string, err error) {
	_ = ctx
	_ = chatID
	_ = limit
	_ = cursor
	return []model.MassageDTO{}, "", nil
}

// GetUserChats ...
func (s *Service) GetUserChats(ctx context.Context, userID int, limit int, offset int) (chats []model.ChatPreviewDTO, err error) {
	_ = ctx
	_ = userID
	_ = limit
	_ = offset
	return []model.ChatPreviewDTO{}, nil
}

// SendMessage ...
func (s *Service) SendMessage(ctx context.Context, chatID int, senderID int, text string) (massageID int, createdAt time.Time, err error) {
	_ = ctx
	_ = chatID
	_ = senderID
	_ = text
	return 0, time.Now(), nil
}
