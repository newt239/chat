package handler

import (
	"errors"
	"strings"
)

// ErrorResponse represents a generic error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// Auth DTOs

type RegisterRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"displayName" binding:"required,min=1"`
}

func (r *RegisterRequest) Validate() error {
	if strings.TrimSpace(r.Email) == "" {
		return errors.New("email is required")
	}
	if len(r.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if strings.TrimSpace(r.DisplayName) == "" {
		return errors.New("display name is required")
	}
	return nil
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (r *LoginRequest) Validate() error {
	if strings.TrimSpace(r.Email) == "" {
		return errors.New("email is required")
	}
	if strings.TrimSpace(r.Password) == "" {
		return errors.New("password is required")
	}
	return nil
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

func (r *RefreshTokenRequest) Validate() error {
	if strings.TrimSpace(r.RefreshToken) == "" {
		return errors.New("refresh token is required")
	}
	return nil
}

type LogoutRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

func (r *LogoutRequest) Validate() error {
	if strings.TrimSpace(r.RefreshToken) == "" {
		return errors.New("refresh token is required")
	}
	return nil
}
