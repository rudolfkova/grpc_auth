// Package model ...
package model

import "time"

// MassageDTO ...
type MassageDTO struct {
	ID        int
	ChatID    int
	SenderID  int
	Text      string
	CreatedAt *time.Time
}

// ChatPreviewDTO ...
type ChatPreviewDTO struct {
	ChatID        int
	CompanionID   int
	LastMessage   string
	UnreadCount   int
	LastMessageAt *time.Time
}
