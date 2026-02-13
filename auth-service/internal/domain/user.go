// Package domain ...
package domain

// User ...
type User struct {
	ID       int
	Email    string
	PassHash []byte
}
