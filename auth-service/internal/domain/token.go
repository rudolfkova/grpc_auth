package domain

import "time"

// Token ...
type Token struct {
	AccessToken     string
	RefreshToken    string
	AccessExpireAt  time.Time
	RefreshExpireAt time.Time
}
