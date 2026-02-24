// Package token_test ...
package tokenjwt_test

import (
	tokenjwt "auth/pkg/token"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	provider = tokenjwt.NewTokenProvider([]byte("123"))
	secret   = []byte("123")
	user     = testUser{
		userID:    42,
		sessionID: 10,
		appID:     1,
		accExp:    time.Now().Add(time.Minute * 15),
	}
)

type testUser struct {
	userID    int
	sessionID int
	appID     int
	accExp    time.Time
}

func TestCreateAccessToken_Success(t *testing.T) {
	accToken, err := provider.CreateAccessToken(user.userID, user.sessionID, user.appID, user.accExp)

	require.NoError(t, err)
	require.NotEmpty(t, accToken)

	token, err := jwt.Parse(accToken, func(token *jwt.Token) (interface{}, error) {
		require.Equal(t, jwt.SigningMethodHS256, token.Method)
		return []byte(secret), nil
	})

	require.NoError(t, err)
	require.True(t, token.Valid)

	claims, ok := token.Claims.(jwt.MapClaims)
	require.True(t, ok)

	exp, ok := claims["exp"].(float64)
	require.True(t, ok)
	userid, ok := claims["user_id"].(float64)
	require.True(t, ok)
	sessionid, ok := claims["session_id"].(float64)
	require.True(t, ok)
	appid, ok := claims["app_id"].(float64)
	require.True(t, ok)

	require.Greater(t, int64(exp), time.Now().Unix())
	assert.Equal(t, float64(user.userID), userid)
	assert.Equal(t, float64(user.sessionID), sessionid)
	assert.Equal(t, float64(user.appID), appid)

}
