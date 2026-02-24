// Package interceptor ...
package interceptor

import (
	authv1 "auth/proto/auth/v1"
	"context"
	"strings"

	authclient "chat/internal/client/auth"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

// UserIDKey ...
const UserIDKey contextKey = "user_id"

// AccessClaims ...
type AccessClaims struct {
	UserID    int `json:"user_id"`
	SessionID int `json:"session_id"`
	AppID     int `json:"app_id"`
	jwt.RegisteredClaims
}

// AuthInterceptor ...
func AuthInterceptor(jwtSecret []byte, authClient *authclient.Client) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		_ = info
		_ = handler
		// 1. Достаём токен из метаданных
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}
		vals := md.Get("authorization")
		if len(vals) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}
		tokenStr := strings.TrimPrefix(vals[0], "Bearer ")

		// 2. Парсим и валидируем JWT локально (подпись + expiration)
		claims := &AccessClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, status.Error(codes.Unauthenticated, "unexpected signing method")
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
		}

		// 3. Проверяем, активна ли сессия в auth-сервисе (пользователь не разлогинился)
		resp, err := authClient.Api.ValidateSession(ctx, &authv1.ValidateSessionRequest{
			SessionId: int64(claims.SessionID),
		})
		if err != nil || !resp.GetActive() {
			return nil, status.Error(codes.Unauthenticated, "session is not active")
		}

		// 4. Кладём user_id в контекст
		ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
		return handler(ctx, req)
	}
}
