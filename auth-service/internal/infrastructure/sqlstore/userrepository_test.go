package sqlstore_test

import (
	"auth/internal/config"
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"auth/internal/infrastructure/sqlstore"

	"github.com/BurntSushi/toml"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

var (
	configPath = "../../../../config.toml"
	cfg        config.Config
	ctx        = context.Background()
)

type testUser struct {
	userID          int
	appID           int
	email           string
	password        string
	passHash        []byte
	refreshToken    string
	refreshTokenExp time.Time
}

func init() {
	c := config.NewConfig()
	_, err := toml.DecodeFile(configPath, c)
	if err != nil {
		log.Fatal(err)
	}
	cfg = *c
}

func testDB(t *testing.T, databaseURL string) (*sql.DB, func(...string)) {
	t.Helper()

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		t.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	return db, func(tables ...string) {
		if len(tables) > 0 {
			query := fmt.Sprintf("TRUNCATE %s CASCADE", strings.Join(tables, ", "))
			if _, err := db.Exec(query); err != nil {
				t.Fatal(err)
			}
		}

		_ = db.Close()
	}
}

func newTestUser() testUser {
	user := testUser{
		email:           "user@example.org",
		password:        "password",
		refreshToken:    "123",
		refreshTokenExp: time.Now(),
		appID:           1,
	}
	user.refreshTokenExp = user.refreshTokenExp.Add(time.Hour * 24 * 7)
	_ = bcrypt.CompareHashAndPassword(user.passHash, []byte(user.password))
	return user
}

func TestUserRepository_SaveAndFindUser(t *testing.T) {
	db, teardown := testDB(t, cfg.TestDatabaseURL)
	defer teardown("users", "sessions")
	s := sqlstore.NewUserRepository(db)
	user := newTestUser()

	err := s.SaveUser(ctx, user.email, user.passHash)
	assert.NoError(t, err)

	domainUser, err := s.UserByEmail(ctx, user.email)
	assert.NoError(t, err)
	assert.Equal(t, user.email, domainUser.Email)
}

func TestUserRepository_SaveAndCheckPermisions(t *testing.T) {
	db, teardown := testDB(t, cfg.TestDatabaseURL)
	defer teardown("users", "sessions")
	s := sqlstore.NewUserRepository(db)
	user := newTestUser()

	err := s.SaveUser(ctx, user.email, user.passHash)
	assert.NoError(t, err)

	domainUser, err := s.UserByEmail(ctx, user.email)
	assert.NoError(t, err)
	assert.Equal(t, user.email, domainUser.Email)

	isAdmin, err := s.IsAdmin(ctx, domainUser.ID)
	assert.NoError(t, err)
	assert.Equal(t, false, isAdmin)
}
