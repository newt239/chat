package entity

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	domerr "github.com/newt239/chat/internal/domain/errors"
)

const (
	// 作成者を含むグループ DM の最大人数
	MaxGroupDMMembers = 10
	// チャンネルパスの最大階層数と 1 階層あたりの最大文字数
	MaxChannelDepth         = 4
	MaxChannelSegmentLength = 32
)

var channelSegmentPattern = regexp.MustCompile(`^[a-z0-9_-]+$`)

var (
	ErrChannelWorkspaceIDInvalid = errors.New("ワークスペースIDの形式が無効です")
	ErrChannelCreatorInvalid     = errors.New("作成者IDはUUID形式で指定してください")
	ErrInvalidChannelType        = errors.New("無効なチャンネル種別です")
	ErrGroupDMMaxMembers         = errors.New("グループDMは自分を含めて10人までです")
	ErrChannelNameInvalid        = fmt.Errorf("%w: チャンネル名は小文字の英数字・ハイフン・アンダースコアをスラッシュで区切った4階層までのパスで指定してください", domerr.ErrValidation)
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
	ParentID    *string
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ArchivedAt  *time.Time
}

type ChannelParams struct {
	ID          string
	WorkspaceID string
	Name        string
	Description *string
	IsPrivate   bool
	Type        ChannelType
	ParentID    *string
	CreatedBy   string
	CreatedAt   time.Time
}

func NewChannel(params ChannelParams) (*Channel, error) {
	workspaceID := strings.TrimSpace(params.WorkspaceID)
	// ワークスペースIDはslug形式（3-12文字の英小文字、数字、ハイフン）またはUUID形式を許可
	if err := ValidateWorkspaceSlug(workspaceID); err != nil {
		// slug形式でない場合、UUID形式かチェック
		if _, err := uuid.Parse(workspaceID); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrChannelWorkspaceIDInvalid, err)
		}
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
	if channelType == ChannelTypePublic || channelType == ChannelTypePrivate {
		normalized, err := NormalizeChannelPath(name)
		if err != nil {
			return nil, err
		}
		name = normalized
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
		ParentID:    cloneString(params.ParentID),
		CreatedBy:   creatorID,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	}, nil
}

// ChangeName はチャンネル名を変更します。階層を移動する変更は受け付けません
func (c *Channel) ChangeName(newName string) error {
	name, err := NormalizeChannelPath(newName)
	if err != nil {
		return err
	}
	if ParentChannelPath(name) != ParentChannelPath(c.Name) {
		return fmt.Errorf("%w: 変更できるのはチャンネル名の末尾の階層のみです", domerr.ErrValidation)
	}
	if c.Name == name {
		return nil
	}

	c.Name = name
	c.UpdatedAt = time.Now().UTC()
	return nil
}

// NormalizeChannelPath はスラッシュ区切りのチャンネルパスを検証し、英字を小文字にそろえます
func NormalizeChannelPath(path string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(path))
	segments := strings.Split(normalized, "/")
	if normalized == "" || len(segments) > MaxChannelDepth {
		return "", ErrChannelNameInvalid
	}
	for _, segment := range segments {
		if len(segment) > MaxChannelSegmentLength || !channelSegmentPattern.MatchString(segment) {
			return "", ErrChannelNameInvalid
		}
	}
	return normalized, nil
}

// ParentChannelPath は親チャンネルのパスを返します。最上位の場合は空文字です
func ParentChannelPath(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx < 0 {
		return ""
	}
	return path[:idx]
}

// AncestorChannelPaths は上位から順に祖先チャンネルのパスを返します
func AncestorChannelPaths(path string) []string {
	segments := strings.Split(path, "/")
	ancestors := make([]string, 0, len(segments)-1)
	for i := 1; i < len(segments); i++ {
		ancestors = append(ancestors, strings.Join(segments[:i], "/"))
	}
	return ancestors
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
