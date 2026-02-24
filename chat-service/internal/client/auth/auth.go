// Package authclient ...
package authclient

import (
	authv1 "auth/proto/auth/v1"

	"google.golang.org/grpc"
)

// Client ...
type Client struct {
	Api authv1.AuthServiceClient
}

// New ...
func New(conn *grpc.ClientConn) *Client {
	return &Client{Api: authv1.NewAuthServiceClient(conn)}
}
