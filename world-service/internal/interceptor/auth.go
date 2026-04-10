package interceptor

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ServiceAuthUnary проверяет внутренний токен. Пустой expectedToken — проверка отключена.
func ServiceAuthUnary(expectedToken string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if expectedToken == "" {
			return handler(ctx, req)
		}
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}
		tok := first(md, "x-service-token")
		if tok == "" {
			tok = bearerToken(first(md, "authorization"))
		}
		if tok == "" || tok != expectedToken {
			return nil, status.Error(codes.Unauthenticated, "invalid service token")
		}
		return handler(ctx, req)
	}
}

func first(md metadata.MD, key string) string {
	v := md.Get(key)
	if len(v) == 0 {
		return ""
	}
	return strings.TrimSpace(v[0])
}

func bearerToken(auth string) string {
	auth = strings.TrimSpace(auth)
	if !strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		return ""
	}
	return strings.TrimSpace(auth[7:])
}
