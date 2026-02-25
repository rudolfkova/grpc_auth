// Package chat ...
package chat

import (
	"chat/internal/grpc/hub"
	"chat/internal/interceptor"
	"chat/internal/model"
	chatv1 "chat/proto/chat/v1"
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Chat ...
type Chat interface {
	GetOrCreateChat(ctx context.Context, initiatorID int, recipientID int) (chatID int, created bool, createdAt time.Time, err error)
	GetMessages(ctx context.Context, chatID int, limit int, cursor string) (massages []model.MassageDTO, nextCursor string, err error)
	GetUserChats(ctx context.Context, userID int, limit int, offset int) (chats []model.ChatPreviewDTO, err error)
	SendMessage(ctx context.Context, chatID int, senderID int, text string) (massageID int, createdAt time.Time, err error)
}

type serverAPI struct {
	chatv1.UnimplementedChatServiceServer
	chat   Chat
	logger *slog.Logger
	hub    *hub.Hub
}

// Register ...
func Register(gRPCServer *grpc.Server, chat Chat, hub *hub.Hub, logger *slog.Logger) {
	chatv1.RegisterChatServiceServer(gRPCServer, &serverAPI{chat: chat, hub: hub, logger: logger})
}

// Ниже бизнес логика сервиса, rpc методы.

// GetOrCreateChat ...
func (s *serverAPI) GetOrCreateChat(ctx context.Context, req *chatv1.GetOrCreateChatRequest) (*chatv1.GetOrCreateChatResponse, error) {
	const op = "serverAPI.GetOrCreateChat"
	log := s.logger.With(
		slog.String("op", op),
	)
	log.Info("GetOrCreateChat")

	userID, ok := ctx.Value(interceptor.UserIDKey).(int)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	if userID != int(req.GetInitiatorId()) {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	chatID, created, createdAt, err := s.chat.GetOrCreateChat(ctx, int(req.GetInitiatorId()), int(req.GetRecipientId()))
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &chatv1.GetOrCreateChatResponse{
		ChatId:    int64(chatID),
		Created:   created,
		CreatedAt: timestamppb.New(createdAt),
	}, nil
}

// GetMassages ...
func (s *serverAPI) GetMessages(ctx context.Context, req *chatv1.GetMessagesRequest) (*chatv1.GetMessagesResponse, error) {
	const op = "serverAPI.GetMassages"
	log := s.logger.With(
		slog.String("op", op),
	)
	log.Info("GetMassages")

	messages, nextCursor, err := s.chat.GetMessages(ctx, int(req.GetChatId()), int(req.GetLimit()), req.GetCursor())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}
	messagesDTO := make([]*chatv1.MessageDTO, len(messages))
	for i := range messages {
		messagesDTO[i] = &chatv1.MessageDTO{
			Id:        int64(messages[i].ID),
			ChatId:    int64(messages[i].ChatID),
			SenderId:  int64(messages[i].SenderID),
			Text:      messages[i].Text,
			CreatedAt: timestamppb.New(*messages[i].CreatedAt),
		}
	}
	return &chatv1.GetMessagesResponse{
		Messages:   messagesDTO,
		NextCursor: nextCursor,
	}, nil
}

// GetUserChats ...
func (s *serverAPI) GetUserChats(ctx context.Context, req *chatv1.GetUserChatsRequest) (*chatv1.GetUserChatsResponse, error) {
	const op = "serverAPI.GetUserChats"
	log := s.logger.With(
		slog.String("op", op),
	)
	log.Info("GetUserChats")

	userID, ok := ctx.Value(interceptor.UserIDKey).(int)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	if userID != int(req.GetUserId()) {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	chatPreview, err := s.chat.GetUserChats(ctx, int(req.GetUserId()), int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}
	chatPreviewDTO := make([]*chatv1.ChatPreviewDTO, len(chatPreview))
	for i := range chatPreview {
		var lastMessageAt *timestamppb.Timestamp
		if chatPreview[i].LastMessageAt != nil {
			lastMessageAt = timestamppb.New(*chatPreview[i].LastMessageAt)
		}

		chatPreviewDTO[i] = &chatv1.ChatPreviewDTO{
			ChatId:        int64(chatPreview[i].ChatID),
			CompanionId:   int64(chatPreview[i].CompanionID),
			LastMessage:   chatPreview[i].LastMessage,
			UnreadCount:   int64(chatPreview[i].UnreadCount),
			LastMessageAt: lastMessageAt,
		}
	}
	return &chatv1.GetUserChatsResponse{
		Chats: chatPreviewDTO,
	}, nil
}

// SendMessage ...
func (s *serverAPI) SendMessage(ctx context.Context, req *chatv1.SendMessageRequest) (*chatv1.SendMessageResponse, error) {
	const op = "serverAPI.SendMessage"
	log := s.logger.With(
		slog.String("op", op),
	)
	log.Info("SendMessage")

	userID, ok := ctx.Value(interceptor.UserIDKey).(int)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	if userID != int(req.GetSenderId()) {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	massageID, createdAt, err := s.chat.SendMessage(ctx, int(req.GetChatId()), int(req.GetSenderId()), req.GetText())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &chatv1.SendMessageResponse{
		MessageId: int64(massageID),
		CreatedAt: timestamppb.New(createdAt),
	}, nil
}

// Subscribe ...
func (s *serverAPI) Subscribe(_ *chatv1.SubscribeRequest, stream chatv1.ChatService_SubscribeServer) error {
	userID, ok := stream.Context().Value(interceptor.UserIDKey).(int)
	if !ok {
		return status.Error(codes.Unauthenticated, "unauthenticated")
	}

	s.hub.Subscribe(userID, stream)
	defer s.hub.Unsubscribe(userID, stream)

	// Держим стрим открытым пока клиент не отключится
	<-stream.Context().Done()
	return nil
}
