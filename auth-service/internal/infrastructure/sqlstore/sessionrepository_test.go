package sqlstore_test

import (
	"auth/internal/infrastructure/sqlstore"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSessionrepository_CreateAndFindByID(t *testing.T) {
	db, teardown := testDB(t, cfg.TestDatabaseURL)
	defer teardown("users", "sessions")
	s := sqlstore.NewSessionRepository(db)
	u := sqlstore.NewUserRepository(db)
	user := newTestUser()

	err := u.SaveUser(ctx, user.email, user.passHash)
	assert.NoError(t, err)

	domainUser, err := u.UserByEmail(ctx, user.email)
	assert.NoError(t, err)
	user.userID = domainUser.ID

	sessionID, err := s.CreateSession(ctx, user.userID, user.appID, user.refreshToken, user.refreshTokenExp)
	assert.NoError(t, err)

	domainSession, err := s.SessionByID(ctx, sessionID)
	assert.NoError(t, err)
	assert.WithinDuration(t, user.refreshTokenExp, domainSession.RefreshExpiresAt, time.Millisecond)
}

func TestSessionrepository_CreateAndRevoke(t *testing.T) {
	db, teardown := testDB(t, cfg.TestDatabaseURL)
	defer teardown("users", "sessions")
	s := sqlstore.NewSessionRepository(db)
	u := sqlstore.NewUserRepository(db)
	user := newTestUser()

	err := u.SaveUser(ctx, user.email, user.passHash)
	assert.NoError(t, err)

	domainUser, err := u.UserByEmail(ctx, user.email)
	assert.NoError(t, err)
	user.userID = domainUser.ID

	sessionID, err := s.CreateSession(ctx, user.userID, user.appID, user.refreshToken, user.refreshTokenExp)
	assert.NoError(t, err)

	domainSessionByID, err := s.SessionByID(ctx, sessionID)
	assert.NoError(t, err)
	assert.WithinDuration(t, user.refreshTokenExp, domainSessionByID.RefreshExpiresAt, time.Millisecond)

	domainSessionByToken, err := s.SessionByRefreshToken(ctx, user.refreshToken)
	assert.NoError(t, err)
	assert.Equal(t, user.userID, domainSessionByToken.UserID)
	assert.Equal(t, "active", domainSessionByToken.Status)

	ok, err := s.RevokeByRefreshToken(ctx, user.refreshToken)
	assert.NoError(t, err)
	assert.True(t, ok)

	revokedDomainSessionByToken, err := s.SessionByRefreshToken(ctx, user.refreshToken)
	assert.NoError(t, err)
	assert.Equal(t, "revoked", revokedDomainSessionByToken.Status)
}
