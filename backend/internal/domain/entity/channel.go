package entity

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrChannelNameRequired       = errors.New("channel name is required")
	ErrChannelWorkspaceIDInvalid = errors.New("workspace ID must be a valid UUID")
	ErrChannelCreatorInvalid     = errors.New("creator ID must be a valid UUID")
)

type Channel struct {
	ID          string
	WorkspaceID string
	Name        string
	Description *string
	IsPrivate   bool
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ChannelParams struct {
	ID          string
	WorkspaceID string
	Name        string
	Description *string
	IsPrivate   bool
	CreatedBy   string
	CreatedAt   time.Time
}

func NewChannel(params ChannelParams) (*Channel, error) {
	workspaceID := strings.TrimSpace(params.WorkspaceID)
	if _, err := uuid.Parse(workspaceID); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrChannelWorkspaceIDInvalid, err)
	}

	creatorID := strings.TrimSpace(params.CreatedBy)
	if _, err := uuid.Parse(creatorID); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrChannelCreatorInvalid, err)
	}

	name := strings.TrimSpace(params.Name)
	if name == "" {
		return nil, ErrChannelNameRequired
	}

	var id string
	if params.ID == "" {
		id = uuid.NewString()
	} else {
		if _, err := uuid.Parse(params.ID); err != nil {
			return nil, fmt.Errorf("channel ID must be a valid UUID: %w", err)
		}
		id = params.ID
	}

	createdAt := params.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	return &Channel{
		ID:          id,
		WorkspaceID: workspaceID,
		Name:        name,
		Description: cloneString(params.Description),
		IsPrivate:   params.IsPrivate,
		CreatedBy:   creatorID,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	}, nil
}

func (c *Channel) ChangeName(newName string) error {
	if c == nil {
		return errors.New("channel is nil")
	}

	name := strings.TrimSpace(newName)
	if name == "" {
		return ErrChannelNameRequired
	}

	if c.Name == name {
		return nil
	}

	c.Name = name
	c.UpdatedAt = time.Now().UTC()
	return nil
}

func cloneString(value *string) *string {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

type ChannelRole string

const (
	ChannelRoleMember ChannelRole = "member"
	ChannelRoleAdmin  ChannelRole = "admin"
)

type ChannelMember struct {
	ChannelID string
	UserID    string
	Role      ChannelRole
	JoinedAt  time.Time
}
