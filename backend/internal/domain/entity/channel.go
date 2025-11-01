package entity

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	MaxGroupDMMembers = 9
)

var (
	ErrChannelNameRequired       = errors.New("チャンネル名は必須です")
	ErrChannelWorkspaceIDInvalid = errors.New("ワークスペースIDはUUID形式で指定してください")
	ErrChannelCreatorInvalid     = errors.New("作成者IDはUUID形式で指定してください")
	ErrInvalidChannelType        = errors.New("無効なチャンネル種別です")
	ErrGroupDMMaxMembers         = errors.New("グループDMには9人までしか追加できません")
)

type ChannelType string

const (
	ChannelTypePublic  ChannelType = "public"
	ChannelTypePrivate ChannelType = "private"
	ChannelTypeDM      ChannelType = "dm"
	ChannelTypeGroupDM ChannelType = "group_dm"
)

func (t ChannelType) IsValid() bool {
	switch t {
	case ChannelTypePublic, ChannelTypePrivate, ChannelTypeDM, ChannelTypeGroupDM:
		return true
	}
	return false
}

type Channel struct {
	ID          string
	WorkspaceID string
	Name        string
	Description *string
	IsPrivate   bool
	Type        ChannelType
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
	Type        ChannelType
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

	channelType := params.Type
	if channelType == "" {
		channelType = ChannelTypePublic
	}
	if !channelType.IsValid() {
		return nil, ErrInvalidChannelType
	}

	name := strings.TrimSpace(params.Name)
	if name == "" && (channelType == ChannelTypePublic || channelType == ChannelTypePrivate) {
		return nil, ErrChannelNameRequired
	}

	var id string
	if params.ID == "" {
		id = uuid.NewString()
	} else {
		if _, err := uuid.Parse(params.ID); err != nil {
			return nil, fmt.Errorf("チャネルIDがUUID形式ではありません: %w", err)
		}
		id = params.ID
	}

	createdAt := params.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	isPrivate := params.IsPrivate
	if channelType == ChannelTypeDM || channelType == ChannelTypeGroupDM {
		isPrivate = true
	}

	return &Channel{
		ID:          id,
		WorkspaceID: workspaceID,
		Name:        name,
		Description: cloneString(params.Description),
		IsPrivate:   isPrivate,
		Type:        channelType,
		CreatedBy:   creatorID,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	}, nil
}

func (c *Channel) ChangeName(newName string) error {
	if c == nil {
		return errors.New("チャンネルが未初期化です")
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
