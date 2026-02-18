// Package rediscache ...
package rediscache

import (
	"auth/internal/config"
	"auth/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Store ...
type Store struct {
	client *redis.Client
	ttl    time.Duration
}

// NewCacheStore ...
func NewCacheStore(cfg config.Config) (*Store, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	// Макс TTL для кэша (сутки, как ты и хотел)
	return &Store{
		client: rdb,
		ttl:    24 * time.Hour,
	}, nil
}

// Close ...
func (s *Store) Close() error {
	return s.client.Close()
}

func sessionKey(id int) string {
	return fmt.Sprintf("session:%d", id)
}

// SetSession ...
func (s *Store) SetSession(ctx context.Context, keyID int, value domain.Session) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	// TTL: не дольше суток и не дольше жизни refresh
	ttl := time.Until(value.RefreshExpiresAt)
	if ttl <= 0 {
		ttl = time.Minute // чтобы не зависло навсегда
	}
	if ttl > s.ttl {
		ttl = s.ttl
	}

	return s.client.Set(ctx, sessionKey(keyID), data, ttl).Err()
}

// GetSession ...
func (s *Store) GetSession(ctx context.Context, keyID int) (bool, domain.Session, error) {
	b, err := s.client.Get(ctx, sessionKey(keyID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return false, domain.Session{}, nil
		}
		return false, domain.Session{}, err
	}

	var sess domain.Session
	if err := json.Unmarshal(b, &sess); err != nil {
		return false, domain.Session{}, err
	}

	return true, sess, nil
}

// DelSession ...
func (s *Store) DelSession(ctx context.Context, keyID int) error {
	return s.client.Del(ctx, sessionKey(keyID)).Err()
}
