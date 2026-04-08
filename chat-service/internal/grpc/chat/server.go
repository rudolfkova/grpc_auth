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

// Chat is the service interface consumed by the gRPC layer.
type Chat interface {
	CreateChat(ctx context.Context, name string) (chatID int, createdAt time.Time, err error)
	DeleteChat(ctx context.Context, chatID int) error
	AddMember(ctx context.Context, chatID, userID int) error
	RemoveMember(ctx context.Context, chatID, userID int) error
	GetOrCreateChat(ctx context.Context, initiatorID, recipientID int) (chatID int, created bool, createdAt time.Time, err error)
	GetMessages(ctx context.Context, chatID, limit int, cursor string) (messages []model.MessageDTO, nextCursor string, err error)
	GetUserChats(ctx context.Context, userID, limit, offset int) ([]model.ChatPreviewDTO, error)
	SendMessage(ctx context.Context, chatID, senderID int, text string) (messageID int, createdAt time.Time, err error)
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

func callerID(ctx context.Context) (int, error) {
	id, ok := ctx.Value(interceptor.UserIDKey).(int)
	if !ok || id == 0 {
		return 0, status.Error(codes.Unauthenticated, "unauthenticated")
	}
	return id, nil
}

// CreateChat ...
func (s *serverAPI) CreateChat(ctx context.Context, req *chatv1.CreateChatRequest) (*chatv1.CreateChatResponse, error) {
	chatID, createdAt, err := s.chat.CreateChat(ctx, req.GetName())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &chatv1.CreateChatResponse{
		ChatId:    int64(chatID),
		CreatedAt: timestamppb.New(createdAt),
	}, nil
}

// DeleteChat ...
func (s *serverAPI) DeleteChat(ctx context.Context, req *chatv1.DeleteChatRequest) (*chatv1.DeleteChatResponse, error) {
	if err := s.chat.DeleteChat(ctx, int(req.GetChatId())); err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &chatv1.DeleteChatResponse{}, nil
}

// AddMember ...
func (s *serverAPI) AddMember(ctx context.Context, req *chatv1.AddMemberRequest) (*chatv1.AddMemberResponse, error) {
	if err := s.chat.AddMember(ctx, int(req.GetChatId()), int(req.GetUserId())); err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &chatv1.AddMemberResponse{}, nil
}

// RemoveMember ...
func (s *serverAPI) RemoveMember(ctx context.Context, req *chatv1.RemoveMemberRequest) (*chatv1.RemoveMemberResponse, error) {
	if err := s.chat.RemoveMember(ctx, int(req.GetChatId()), int(req.GetUserId())); err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &chatv1.RemoveMemberResponse{}, nil
}

// GetOrCreateChat ...
func (s *serverAPI) GetOrCreateChat(ctx context.Context, req *chatv1.GetOrCreateChatRequest) (*chatv1.GetOrCreateChatResponse, error) {
	userID, err := callerID(ctx)
	if err != nil {
		return nil, err
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

// GetMessages ...
func (s *serverAPI) GetMessages(ctx context.Context, req *chatv1.GetMessagesRequest) (*chatv1.GetMessagesResponse, error) {
	messages, nextCursor, err := s.chat.GetMessages(ctx, int(req.GetChatId()), int(req.GetLimit()), req.GetCursor())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	dtos := make([]*chatv1.MessageDTO, len(messages))
	for i, m := range messages {
		dtos[i] = &chatv1.MessageDTO{
			Id:        int64(m.ID),
			ChatId:    int64(m.ChatID),
			SenderId:  int64(m.SenderID),
			Text:      m.Text,
			CreatedAt: timestamppb.New(*m.CreatedAt),
		}
	}
	return &chatv1.GetMessagesResponse{Messages: dtos, NextCursor: nextCursor}, nil
}

// GetUserChats ...
func (s *serverAPI) GetUserChats(ctx context.Context, req *chatv1.GetUserChatsRequest) (*chatv1.GetUserChatsResponse, error) {
	userID, err := callerID(ctx)
	if err != nil {
		return nil, err
	}
	if userID != int(req.GetUserId()) {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	chats, err := s.chat.GetUserChats(ctx, int(req.GetUserId()), int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	dtos := make([]*chatv1.ChatPreviewDTO, len(chats))
	for i, c := range chats {
		var lastMessageAt *timestamppb.Timestamp
		if c.LastMessageAt != nil {
			lastMessageAt = timestamppb.New(*c.LastMessageAt)
		}
		dtos[i] = &chatv1.ChatPreviewDTO{
			ChatId:        int64(c.ChatID),
			Name:          c.Name,
			CompanionId:   int64(c.CompanionID),
			LastMessage:   c.LastMessage,
			UnreadCount:   int64(c.UnreadCount),
			LastMessageAt: lastMessageAt,
		}
	}
	return &chatv1.GetUserChatsResponse{Chats: dtos}, nil
}

// SendMessage ...
func (s *serverAPI) SendMessage(ctx context.Context, req *chatv1.SendMessageRequest) (*chatv1.SendMessageResponse, error) {
	userID, err := callerID(ctx)
	if err != nil {
		return nil, err
	}
	if userID != int(req.GetSenderId()) {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	messageID, createdAt, err := s.chat.SendMessage(ctx, int(req.GetChatId()), int(req.GetSenderId()), req.GetText())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &chatv1.SendMessageResponse{
		MessageId: int64(messageID),
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

	<-stream.Context().Done()
	return nil
}
