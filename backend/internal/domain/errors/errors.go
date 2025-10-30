package errors

import "errors"

var (
    ErrMessageNotFound = errors.New("message not found")
    ErrChannelNotFound = errors.New("channel not found")
    ErrUnauthorized    = errors.New("unauthorized")
    ErrForbidden       = errors.New("forbidden")
)

package errors

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrSessionNotFound    = errors.New("session not found")
	ErrNotFound           = errors.New("resource not found")
	ErrUnauthorized       = errors.New("unauthorized access")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidInput       = errors.New("invalid input")
	ErrConflict           = errors.New("conflict")
	ErrValidation         = errors.New("validation error")
	ErrInternal           = errors.New("internal server error")
)
