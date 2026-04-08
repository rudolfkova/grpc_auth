// Package repository ...
package repository

import (
	"chat/internal/model"
	"context"
	"database/sql"
	"fmt"
	"time"
)

// MessageRepository ...
type MessageRepository struct {
	db *sql.DB
}

// NewMessageRepository ...
func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// GetMessages возвращает сообщения чата с cursor-пагинацией (по created_at DESC).
func (r *MessageRepository) GetMessages(ctx context.Context, chatID, limit int, cursor string) ([]model.MessageDTO, error) {
	const op = "MessageRepository.GetMessages"

	var (
		rows *sql.Rows
		err  error
	)
	if cursor == "" {
		rows, err = r.db.QueryContext(ctx, `
			SELECT id, chat_id, sender_id, text, created_at
			FROM messages
			WHERE chat_id = $1
			ORDER BY created_at DESC
			LIMIT $2
		`, chatID, limit)
	} else {
		rows, err = r.db.QueryContext(ctx, `
			SELECT id, chat_id, sender_id, text, created_at
			FROM messages
			WHERE chat_id = $1 AND created_at < $2
			ORDER BY created_at DESC
			LIMIT $3
		`, chatID, cursor, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var messages []model.MessageDTO
	for rows.Next() {
		var msg model.MessageDTO
		if err := rows.Scan(&msg.ID, &msg.ChatID, &msg.SenderID, &msg.Text, &msg.CreatedAt); err != nil {
			return nil, fmt.Errorf("%s: scan: %w", op, err)
		}
		messages = append(messages, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows: %w", op, err)
	}
	return messages, nil
}

// SendMessage вставляет сообщение и инкрементирует unread всем участникам кроме отправителя.
func (r *MessageRepository) SendMessage(ctx context.Context, chatID, senderID int, text string) (int, time.Time, error) {
	const op = "MessageRepository.SendMessage"

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("%s: begin tx: %w", op, err)
	}
	defer func() { _ = tx.Rollback() }()

	var messageID int
	var createdAt time.Time
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO messages (chat_id, sender_id, text)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`, chatID, senderID, text).Scan(&messageID, &createdAt); err != nil {
		return 0, time.Time{}, fmt.Errorf("%s: insert message: %w", op, err)
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE chat_members
		SET unread_count = unread_count + 1
		WHERE chat_id = $1 AND user_id != $2
	`, chatID, senderID); err != nil {
		return 0, time.Time{}, fmt.Errorf("%s: update unread: %w", op, err)
	}

	if err := tx.Commit(); err != nil {
		return 0, time.Time{}, fmt.Errorf("%s: commit: %w", op, err)
	}

	return messageID, createdAt, nil
}
