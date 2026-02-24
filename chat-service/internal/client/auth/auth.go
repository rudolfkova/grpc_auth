// Package authclient ...
package authclient

import (
	authv1 "auth/proto/auth/v1"

	"google.golang.org/grpc"
)

// Client ...
type Client struct {
	API authv1.AuthServiceClient
}

// New ...
func New(conn *grpc.ClientConn) *Client {
	return &Client{API: authv1.NewAuthServiceClient(conn)}
}
