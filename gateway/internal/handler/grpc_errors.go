// Package handler ...
package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// grpcStatusToHTTP конвертирует gRPC-код ошибки в HTTP-статус.
func grpcStatusToHTTP(err error) int {
	s, ok := status.FromError(err)
	if !ok {
		return http.StatusInternalServerError
	}
	switch s.Code() {
	case codes.OK:
		return http.StatusOK
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// grpcMessage извлекает текст ошибки из gRPC-статуса.
func grpcMessage(err error) string {
	s, ok := status.FromError(err)
	if !ok {
		return err.Error()
	}
	return s.Message()
}

// queryInt64 читает int64 параметр из URL query string.
func queryInt64(r *http.Request, key string) (int64, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return 0, fmt.Errorf("missing query param: %s", key)
	}
	return strconv.ParseInt(raw, 10, 64)
}
