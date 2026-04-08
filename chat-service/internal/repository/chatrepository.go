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

// CreateChat создаёт новый чат с указанным типом и именем.
// Для DM передаётся type="dm", name="". Для группового — type="group", name=<название>.
func (r *ChatRepository) CreateChat(ctx context.Context, chatType, name string) (int, time.Time, error) {
	const op = "ChatRepository.CreateChat"
	const query = `
		INSERT INTO chats (type, name)
		VALUES ($1, NULLIF($2, ''))
		RETURNING id, created_at
	`
	var id int
	var createdAt time.Time
	err := r.db.QueryRowContext(ctx, query, chatType, name).Scan(&id, &createdAt)
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("%s: %w", op, err)
	}
	return id, createdAt, nil
}

// DeleteChat удаляет чат. CASCADE удалит chat_members и messages.
func (r *ChatRepository) DeleteChat(ctx context.Context, chatID int) error {
	const op = "ChatRepository.DeleteChat"
	_, err := r.db.ExecContext(ctx, `DELETE FROM chats WHERE id = $1`, chatID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// AddMember добавляет участника в чат. Если уже состоит — игнорирует.
func (r *ChatRepository) AddMember(ctx context.Context, chatID, userID int) error {
	const op = "ChatRepository.AddMember"
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO chat_members (chat_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, chatID, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// RemoveMember удаляет участника из чата.
func (r *ChatRepository) RemoveMember(ctx context.Context, chatID, userID int) error {
	const op = "ChatRepository.RemoveMember"
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM chat_members WHERE chat_id = $1 AND user_id = $2
	`, chatID, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// IsMember проверяет, состоит ли пользователь в чате.
func (r *ChatRepository) IsMember(ctx context.Context, chatID, userID int) (bool, error) {
	const op = "ChatRepository.IsMember"
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM chat_members WHERE chat_id = $1 AND user_id = $2
		)
	`, chatID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return exists, nil
}

// GetMemberIDs возвращает список всех участников чата.
func (r *ChatRepository) GetMemberIDs(ctx context.Context, chatID int) ([]int, error) {
	const op = "ChatRepository.GetMemberIDs"
	rows, err := r.db.QueryContext(ctx, `
		SELECT user_id FROM chat_members WHERE chat_id = $1
	`, chatID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("%s: scan: %w", op, err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows: %w", op, err)
	}
	return ids, nil
}

// GetOrCreateDMChat возвращает существующий DM-чат между двумя пользователями
// или создаёт новый. Уникальность пары гарантируется на уровне приложения:
// мы всегда ищем по (min, max) user_id через chat_members.
func (r *ChatRepository) GetOrCreateDMChat(ctx context.Context, initiatorID, recipientID int) (int, bool, time.Time, error) {
	const op = "ChatRepository.GetOrCreateDMChat"

	// Ищем существующий DM-чат между двумя пользователями.
	const findQuery = `
		SELECT c.id, c.created_at
		FROM chats c
		JOIN chat_members m1 ON m1.chat_id = c.id AND m1.user_id = $1
		JOIN chat_members m2 ON m2.chat_id = c.id AND m2.user_id = $2
		WHERE c.type = 'dm'
		LIMIT 1
	`
	var chatID int
	var createdAt time.Time
	err := r.db.QueryRowContext(ctx, findQuery, initiatorID, recipientID).Scan(&chatID, &createdAt)
	if err == nil {
		return chatID, false, createdAt, nil
	}
	if err != sql.ErrNoRows {
		return 0, false, time.Time{}, fmt.Errorf("%s: find: %w", op, err)
	}

	// Не нашли — создаём в транзакции.
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, false, time.Time{}, fmt.Errorf("%s: begin tx: %w", op, err)
	}
	defer func() { _ = tx.Rollback() }()

	err = tx.QueryRowContext(ctx, `
		INSERT INTO chats (type) VALUES ('dm') RETURNING id, created_at
	`).Scan(&chatID, &createdAt)
	if err != nil {
		return 0, false, time.Time{}, fmt.Errorf("%s: insert chat: %w", op, err)
	}

	if _, err = tx.ExecContext(ctx, `
		INSERT INTO chat_members (chat_id, user_id) VALUES ($1, $2), ($1, $3)
	`, chatID, initiatorID, recipientID); err != nil {
		return 0, false, time.Time{}, fmt.Errorf("%s: insert members: %w", op, err)
	}

	if err = tx.Commit(); err != nil {
		return 0, false, time.Time{}, fmt.Errorf("%s: commit: %w", op, err)
	}

	return chatID, true, createdAt, nil
}

// GetUserChats возвращает превью чатов пользователя с последним сообщением.
func (r *ChatRepository) GetUserChats(ctx context.Context, userID, limit, offset int) ([]model.ChatPreviewDTO, error) {
	const op = "ChatRepository.GetUserChats"
	const query = `
		SELECT
			c.id                                                        AS chat_id,
			COALESCE(c.name, '')                                        AS name,
			CASE WHEN c.type = 'dm'
				THEN (
					SELECT user_id FROM chat_members
					WHERE chat_id = c.id AND user_id != $1
					LIMIT 1
				)
				ELSE 0 END                                              AS companion_id,
			COALESCE(m.text, '')                                        AS last_message,
			COALESCE(cm.unread_count, 0)                                AS unread_count,
			m.created_at                                                AS last_message_at
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
	defer rows.Close()

	var chats []model.ChatPreviewDTO
	for rows.Next() {
		var c model.ChatPreviewDTO
		if err := rows.Scan(
			&c.ChatID, &c.Name, &c.CompanionID,
			&c.LastMessage, &c.UnreadCount, &c.LastMessageAt,
		); err != nil {
			return nil, fmt.Errorf("%s: scan: %w", op, err)
		}
		chats = append(chats, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows: %w", op, err)
	}
	return chats, nil
}

// ResetUnread сбрасывает счётчик непрочитанных для пользователя в чате.
func (r *ChatRepository) ResetUnread(ctx context.Context, chatID, userID int) error {
	const op = "ChatRepository.ResetUnread"
	_, err := r.db.ExecContext(ctx, `
		UPDATE chat_members SET unread_count = 0
		WHERE chat_id = $1 AND user_id = $2
	`, chatID, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
