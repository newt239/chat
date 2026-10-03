package entity

import (
	"regexp"
	"strings"
	"time"

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
	ErrInvalidChannelType = domerr.New(domerr.ErrValidation, "無効なチャンネル種別です")
	ErrGroupDMMaxMembers  = domerr.New(domerr.ErrValidation, "グループDMは自分を含めて10人までです")
	ErrChannelNameInvalid = domerr.New(domerr.ErrValidation, "チャンネル名は小文字の英数字・ハイフン・アンダースコアをスラッシュで区切った4階層までのパスで指定してください")
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
	Type        ChannelType
	ParentID    *string
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewChannel は種別を省略したら公開にし、公開・非公開チャンネルの名前をパスとして正規化します
func NewChannel(ch Channel) (*Channel, error) {
	if ch.Type == "" {
		ch.Type = ChannelTypePublic
	}
	if !ch.Type.IsValid() {
		return nil, ErrInvalidChannelType
	}
	if !ch.IsDM() {
		name, err := NormalizeChannelPath(ch.Name)
		if err != nil {
			return nil, err
		}
		ch.Name = name
	}
	return &ch, nil
}

// IsPrivate は参加者だけが閲覧できるチャンネルかを返します。DM とグループ DM も含む
func (c *Channel) IsPrivate() bool {
	return c.Type != ChannelTypePublic
}

// IsDM は 1 対 1 の DM とグループ DM かを返します
func (c *Channel) IsDM() bool {
	return c.Type == ChannelTypeDM || c.Type == ChannelTypeGroupDM
}

// ChangeName はチャンネル名を変更します。階層を移動する変更は受け付けません
func (c *Channel) ChangeName(newName string) error {
	name, err := NormalizeChannelPath(newName)
	if err != nil {
		return err
	}
	if ParentChannelPath(name) != ParentChannelPath(c.Name) {
		return domerr.New(domerr.ErrValidation, "変更できるのはチャンネル名の末尾の階層のみです")
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
