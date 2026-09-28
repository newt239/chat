package message

import (
	"errors"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
)

var (
	ErrChannelNotFound       = errors.New("チャンネルが見つかりません")
	ErrUnauthorized          = errors.New("この操作を行う権限がありません")
	ErrParentMessageNotFound = errors.New("親メッセージが見つかりません")
	ErrMessageNotFound       = errors.New("メッセージが見つかりません")
	ErrMessageAlreadyDeleted = errors.New("メッセージは既に削除されています")
	ErrCannotEditDeleted     = errors.New("削除済みメッセージは編集できません")
	ErrAttachmentNotFound    = errors.New("添付ファイルが見つかりません")
)

const (
	defaultMessageLimit = 50
	maxMessageLimit     = 100
)

type ListMessagesInput struct {
	ChannelID          string
	UserID             string
	Limit              int
	Since              *time.Time
	Until              *time.Time
	IncludeDescendants bool
}

type CreateMessageInput struct {
	ChannelID     string
	UserID        string
	Body          string
	ParentID      *string
	AttachmentIDs []string
}

type UpdateMessageInput struct {
	MessageID string
	ChannelID string
	EditorID  string
	Body      string
}

type DeleteMessageInput struct {
	MessageID  string
	ChannelID  string
	ExecutorID string
}

type UserInfo struct {
	ID          string  `json:"id"`
	DisplayName string  `json:"displayName"`
	AvatarURL   *string `json:"avatarUrl,omitempty"`
}

type UserMention struct {
	UserID      string `json:"userId"`
	DisplayName string `json:"displayName"`
}

type GroupMention struct {
	GroupID string `json:"groupId"`
	Name    string `json:"name"`
}

type LinkInfo struct {
	ID              string
	URL             string
	OGP             entity.OGPData
	LinkedMessageID *string
	// 閲覧者が参照できるメッセージリンクのときのみ設定される
	MessagePreview *MessagePreviewOutput
}

// MessagePreviewOutput はメッセージリンクの引用カードの内容です
type MessagePreviewOutput struct {
	MessageID   string
	ChannelID   string
	ChannelName string
	ParentID    *string
	User        UserInfo
	BodyExcerpt string
	CreatedAt   time.Time
}

type PinInfo struct {
	PinnedBy UserInfo
	PinnedAt time.Time
}

type ReactionInfo struct {
	User      UserInfo  `json:"user"`
	Emoji     string    `json:"emoji"`
	CreatedAt time.Time `json:"createdAt"`
}

type AttachmentInfo struct {
	ID        string
	FileName  string
	MimeType  string
	SizeBytes int64
	Media     entity.MediaMetadata
}

type MessageOutput struct {
	ID          string           `json:"id"`
	ChannelID   string           `json:"channelId"`
	UserID      string           `json:"userId"`
	User        UserInfo         `json:"user"`
	ParentID    *string          `json:"parentId"`
	Body        string           `json:"body"`
	Mentions    []UserMention    `json:"mentions"`
	Groups      []GroupMention   `json:"groups"`
	Links       []LinkInfo       `json:"links"`
	Reactions   []ReactionInfo   `json:"reactions"`
	Attachments []AttachmentInfo `json:"attachments"`
	CreatedAt   time.Time        `json:"createdAt"`
	EditedAt    *time.Time       `json:"editedAt"`
	DeletedAt   *time.Time       `json:"deletedAt"`
	IsDeleted   bool             `json:"isDeleted"`
	DeletedBy   *UserInfo        `json:"deletedBy,omitempty"`
	Pin         *PinInfo         `json:"pin,omitempty"`
}

// WithoutMessagePreviews は引用カードを除いたコピーを返します。
// 投稿者の権限で組み立てた引用を、参照権限の異なる購読者へ配信しないために使います
func (m MessageOutput) WithoutMessagePreviews() MessageOutput {
	links := make([]LinkInfo, len(m.Links))
	for i, link := range m.Links {
		link.MessagePreview = nil
		links[i] = link
	}
	m.Links = links
	return m
}

type ListMessagesOutput struct {
	Messages []TimelineItem `json:"messages"`
	HasMore  bool           `json:"hasMore"`
}

type ThreadMetadataOutput struct {
	MessageID          string     `json:"messageId"`
	ReplyCount         int        `json:"replyCount"`
	LastReplyAt        *time.Time `json:"lastReplyAt"`
	LastReplyUser      *UserInfo  `json:"lastReplyUser"`
	ParticipantUserIDs []string   `json:"participantUserIds"`
}

type GetThreadRepliesInput struct {
	MessageID string
	UserID    string
	Limit     int
}

type GetThreadRepliesOutput struct {
	ParentMessage MessageOutput   `json:"parentMessage"`
	Replies       []MessageOutput `json:"replies"`
	HasMore       bool            `json:"hasMore"`
}

type GetMessagePreviewInput struct {
	MessageID string
	UserID    string
}

type GetThreadMetadataInput struct {
	MessageID string
	UserID    string
}

type MessageWithThreadOutput struct {
	MessageOutput
	ThreadMetadata *ThreadMetadataOutput `json:"threadMetadata,omitempty"`
}

// SystemMessageOutput はシステムメッセージの出力です
type SystemMessageOutput struct {
	ID        string         `json:"id"`
	ChannelID string         `json:"channelId"`
	Kind      string         `json:"kind"`
	Payload   map[string]any `json:"payload"`
	ActorID   *string        `json:"actorId,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
}

// TimelineItem はユーザー/システム両メッセージの統合タイムライン項目です
type TimelineItem struct {
	Type          string               `json:"type"` // "user" | "system"
	UserMessage   *MessageOutput       `json:"userMessage,omitempty"`
	SystemMessage *SystemMessageOutput `json:"systemMessage,omitempty"`
	CreatedAt     time.Time            `json:"createdAt"`
}
