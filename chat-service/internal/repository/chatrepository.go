// Package repository ...
package repository

import (
	"chat/internal/model"
	"context"
	"database/sql"
	"fmt"
	"time"
)

// ChatRepository ...
type ChatRepository struct {
	db *sql.DB
}

// NewChatRepository ...
func NewChatRepository(db *sql.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

// GetOrCreateChat ...
func (r *ChatRepository) GetOrCreateChat(ctx context.Context, initiatorID int, recipientID int) (int, bool, time.Time, error) {
	const op = "ChatRepository.GetOrCreateChat"

	// Нормализуем порядок — в БД всегда user1_id < user2_id
	user1, user2 := initiatorID, recipientID
	if user1 > user2 {
		user1, user2 = user2, user1
	}

	// Один запрос: пытаемся вставить, при конфликте обновляем id на то же значение.
	// xmax = 0 означает что строка только что создана (не была обновлена).
	const query = `
        INSERT INTO chats (user1_id, user2_id)
        VALUES ($1, $2)
        ON CONFLICT (user1_id, user2_id) DO UPDATE
            SET id = chats.id
        RETURNING id, created_at, (xmax = 0) AS created
    `

	var (
		chatID    int
		createdAt time.Time
		created   bool
	)

	err := r.db.QueryRowContext(ctx, query, user1, user2).Scan(&chatID, &createdAt, &created)
	if err != nil {
		return 0, false, time.Time{}, fmt.Errorf("%s: %w", op, err)
	}

	// Если чат только что создан — вставляем обоих участников в chat_members
	if created {
		const memberQuery = `
            INSERT INTO chat_members (chat_id, user_id)
            VALUES ($1, $2), ($1, $3)
        `
		if _, err := r.db.ExecContext(ctx, memberQuery, chatID, user1, user2); err != nil {
			return 0, false, time.Time{}, fmt.Errorf("%s: insert members: %w", op, err)
		}
	}

	return chatID, created, createdAt, nil
}

// IsMember ...
func (r *ChatRepository) IsMember(ctx context.Context, chatID int, userID int) (bool, error) {
	const op = "ChatRepository.IsMember"

	const query = `
        SELECT EXISTS (
            SELECT 1 FROM chat_members
            WHERE chat_id = $1 AND user_id = $2
        )
    `

	var exists bool
	err := r.db.QueryRowContext(ctx, query, chatID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return exists, nil
}

// GetUserChats ...
func (r *ChatRepository) GetUserChats(ctx context.Context, userID int, limit int, offset int) ([]model.ChatPreviewDTO, error) {
	const op = "ChatRepository.GetUserChats"

	const query = `
        SELECT
            c.id                                            AS chat_id,
            CASE WHEN c.user1_id = $1 THEN c.user2_id
                 ELSE c.user1_id END                       AS companion_id,
            COALESCE(m.text, '')                           AS last_message,
            COALESCE(cm.unread_count, 0)                   AS unread_count,
            m.created_at                                   AS last_message_at
        FROM chats c
        JOIN chat_members cm ON cm.chat_id = c.id AND cm.user_id = $1
        LEFT JOIN LATERAL (
            SELECT text, created_at
            FROM messages
            WHERE chat_id = c.id
            ORDER BY created_at DESC
            LIMIT 1
        ) m ON true
        ORDER BY m.created_at DESC NULLS LAST
        LIMIT $2 OFFSET $3
    `

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	var chats []model.ChatPreviewDTO
	for rows.Next() {
		var chat model.ChatPreviewDTO
		err := rows.Scan(
			&chat.ChatID,
			&chat.CompanionID,
			&chat.LastMessage,
			&chat.UnreadCount,
			&chat.LastMessageAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: scan: %w", op, err)
		}
		chats = append(chats, chat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows: %w", op, err)
	}

	return chats, nil
}

// IncrementUnread ...
func (r *ChatRepository) IncrementUnread(ctx context.Context, chatID int, senderID int) error {
	const op = "ChatRepository.IncrementUnread"

	const query = `
        UPDATE chat_members
        SET unread_count = unread_count + 1
        WHERE chat_id = $1 AND user_id != $2
    `

	if _, err := r.db.ExecContext(ctx, query, chatID, senderID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// ResetUnread ...
func (r *ChatRepository) ResetUnread(ctx context.Context, chatID int, userID int) error {
	const op = "ChatRepository.ResetUnread"

	const query = `
        UPDATE chat_members
        SET unread_count = 0
        WHERE chat_id = $1 AND user_id = $2
    `

	if _, err := r.db.ExecContext(ctx, query, chatID, userID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// GetParticipants ...
func (r *ChatRepository) GetParticipants(ctx context.Context, chatID int) (user1ID int, user2ID int, err error) {
	const op = "ChatRepository.GetParticipants"

	const query = `
        SELECT user1_id, user2_id FROM chats WHERE id = $1
    `

	err = r.db.QueryRowContext(ctx, query, chatID).Scan(&user1ID, &user2ID)
	if err != nil {
		return 0, 0, fmt.Errorf("%s: %w", op, err)
	}

	return user1ID, user2ID, nil
}
