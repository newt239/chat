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

// Workspace DTOs - no additional DTOs needed for now, using usecase outputs directly

type CreateWorkspaceRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

func (r *CreateWorkspaceRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return errors.New("name is required")
	}
	return nil
}

type UpdateWorkspaceRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	IconURL     *string `json:"iconUrl"`
}

func (r *UpdateWorkspaceRequest) Validate() error {
	if r.Name == nil && r.Description == nil && r.IconURL == nil {
		return errors.New("at least one field must be provided")
	}

	if r.Name != nil && strings.TrimSpace(*r.Name) == "" {
		return errors.New("name cannot be empty")
	}

	return nil
}

type AddMemberRequest struct {
	UserID string `json:"userId"`
	Role   string `json:"role"`
}

func (r *AddMemberRequest) Validate() error {
	if strings.TrimSpace(r.UserID) == "" {
		return errors.New("userId is required")
	}
	if strings.TrimSpace(r.Role) == "" {
		return errors.New("role is required")
	}
	return nil
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role"`
}

func (r *UpdateMemberRoleRequest) Validate() error {
	if strings.TrimSpace(r.Role) == "" {
		return errors.New("role is required")
	}
	return nil
}

type CreateChannelRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	IsPrivate   bool    `json:"isPrivate"`
}

func (r *CreateChannelRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return errors.New("name is required")
	}
	return nil
}

type CreateMessageRequest struct {
	Body     string  `json:"body"`
	ParentID *string `json:"parentId"`
}

func (r *CreateMessageRequest) Validate() error {
	if strings.TrimSpace(r.Body) == "" {
		return errors.New("body is required")
	}
	return nil
}

type UpdateReadStateRequest struct {
	LastReadAt string `json:"lastReadAt"`
}

func (r *UpdateReadStateRequest) Validate() error {
	if strings.TrimSpace(r.LastReadAt) == "" {
		return errors.New("lastReadAt is required")
	}
	return nil
}
