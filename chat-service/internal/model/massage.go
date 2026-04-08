// Package model ...
package model

import "time"

// MessageDTO ...
type MessageDTO struct {
	ID        int
	ChatID    int
	SenderID  int
	Text      string
	CreatedAt *time.Time
}

// ChatPreviewDTO ...
type ChatPreviewDTO struct {
	ChatID        int
	Name          string
	CompanionID   int // только для DM, иначе 0
	LastMessage   string
	UnreadCount   int
	LastMessageAt *time.Time
}
