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

// GetMessages ...
func (r *MessageRepository) GetMessages(ctx context.Context, chatID int, limit int, cursor string) ([]model.MassageDTO, error) {
	const op = "MessageRepository.GetMessages"

	var (
		rows *sql.Rows
		err  error
	)

	if cursor == "" {
		const query = `
			SELECT id, chat_id, sender_id, text, created_at
			FROM messages
			WHERE chat_id = $1
			ORDER BY created_at DESC
			LIMIT $2
		`
		rows, err = r.db.QueryContext(ctx, query, chatID, limit)
	} else {
		const query = `
			SELECT id, chat_id, sender_id, text, created_at
			FROM messages
			WHERE chat_id = $1 AND created_at < $2
			ORDER BY created_at DESC
			LIMIT $3
		`
		rows, err = r.db.QueryContext(ctx, query, chatID, cursor, limit)
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	var messages []model.MassageDTO
	for rows.Next() {
		var msg model.MassageDTO
		if err := rows.Scan(
			&msg.ID,
			&msg.ChatID,
			&msg.SenderID,
			&msg.Text,
			&msg.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("%s: scan: %w", op, err)
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows: %w", op, err)
	}

	return messages, nil
}

// SendMessage ...
func (r *MessageRepository) SendMessage(ctx context.Context, chatID int, senderID int, text string) (int, time.Time, error) {
	const op = "MessageRepository.SendMessage"

	const query = `
		INSERT INTO messages (chat_id, sender_id, text)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	var (
		messageID int
		createdAt time.Time
	)

	err := r.db.QueryRowContext(ctx, query, chatID, senderID, text).Scan(&messageID, &createdAt)
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("%s: %w", op, err)
	}

	return messageID, createdAt, nil
}
