// Package grpcauth ...
package grpcauth

import (
	authv1 "auth/proto/auth/v1"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

func ValidateRegisterRequest(req *authv1.RegisterRequest) error {
	return validation.ValidateStruct(
		req,
		validation.Field(&req.Email, validation.Required, is.Email),
		validation.Field(&req.Password, validation.Length(6, 100)),
	)
}
