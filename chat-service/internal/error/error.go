// Package chaterror ...
package chaterror

import "errors"

var (
	ErrUnauthenticated  = errors.New("unauthenticated")
	ErrPermissionDenied = errors.New("permission denied")
	ErrChatNotFound     = errors.New("chat not found")
)
