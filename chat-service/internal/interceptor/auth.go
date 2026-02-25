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
func AuthInterceptor(jwtSecret string, authClient *authclient.Client) grpc.UnaryServerInterceptor {
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
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
		}

		// 3. Проверяем, активна ли сессия в auth-сервисе (пользователь не разлогинился)
		resp, err := authClient.API.ValidateSession(ctx, &authv1.ValidateSessionRequest{
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

// AuthStreamInterceptor — то же самое что AuthInterceptor, но для стриминговых методов.
func AuthStreamInterceptor(jwtSecret string, authClient *authclient.Client) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		_ = info
		_ = handler
		// 1. Достаём токен из метаданных
		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return status.Error(codes.Unauthenticated, "missing metadata")
		}
		vals := md.Get("authorization")
		if len(vals) == 0 {
			return status.Error(codes.Unauthenticated, "missing authorization header")
		}
		tokenStr := strings.TrimPrefix(vals[0], "Bearer ")

		// 2. Парсим и валидируем JWT
		claims := &AccessClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, status.Error(codes.Unauthenticated, "unexpected signing method")
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			return status.Error(codes.Unauthenticated, "invalid or expired token")
		}

		// 3. Проверяем сессию
		resp, err := authClient.API.ValidateSession(ss.Context(), &authv1.ValidateSessionRequest{
			SessionId: int64(claims.SessionID),
		})
		if err != nil || !resp.GetActive() {
			return status.Error(codes.Unauthenticated, "session is not active")
		}

		// 4. Кладём user_id в контекст через обёртку стрима
		wrapped := &wrappedStream{ss, context.WithValue(ss.Context(), UserIDKey, claims.UserID)}
		return handler(srv, wrapped)
	}
}

// wrappedStream позволяет подменить контекст у ServerStream.
type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context ...
func (w *wrappedStream) Context() context.Context {
	return w.ctx
}
