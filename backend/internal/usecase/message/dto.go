package message

import (
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
)

var (
	ErrMessageAlreadyDeleted = domerr.New(domerr.ErrFailedPrecondition, "メッセージは既に削除されています")
	ErrCannotEditDeleted     = domerr.New(domerr.ErrFailedPrecondition, "削除済みメッセージは編集できません")
	ErrEmptyMessage          = domerr.New(domerr.ErrValidation, "本文・添付・位置情報・投票のいずれかが必要です")
	ErrOfficialMessage       = domerr.New(domerr.ErrUnauthorized, "公式アプリの投稿は編集・削除できません")
)

const (
	defaultMessageLimit = 50
	maxMessageLimit     = 100
)

type ListMessagesInput struct {
	ChannelID string
	UserID    string
	Limit     int
	Since     *time.Time
	Until     *time.Time
	// 指定すると since / until より優先し、前後 limit 件ずつ取る
	Around             *time.Time
	IncludeDescendants bool
}

type CreateMessageInput struct {
	ChannelID     string
	UserID        string
	Body          string
	ParentID      *string
	AttachmentIDs []string
	Location      *entity.MessageLocation
	Poll          *PollInput
}

type UpdateMessageInput struct {
	MessageID string
	EditorID  string
	Body      string
}

// MessageInput は 1 件のメッセージに対する操作の入力です
type MessageInput struct {
	MessageID string
	UserID    string
}

type UserInfo struct {
	ID          string
	DisplayName string
	AvatarURL   *string
	IsApp       bool
}

type UserMention struct {
	UserID     string
	ViaGroupID *string
}

type GroupMention struct {
	GroupID string
	Name    string
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
	User      UserInfo
	Emoji     string
	CreatedAt time.Time
}

type AttachmentInfo struct {
	ID        string
	FileName  string
	MimeType  string
	SizeBytes int64
	Media     entity.MediaMetadata
}

type MessageOutput struct {
	ID              string
	ChannelID       string
	UserID          string
	User            UserInfo
	ParentID        *string
	Body            string
	Mentions        []UserMention
	Groups          []GroupMention
	Links           []LinkInfo
	Reactions       []ReactionInfo
	Attachments     []AttachmentInfo
	CreatedAt       time.Time
	EditedAt        *time.Time
	DeletedAt       *time.Time
	DeletedBy       *UserInfo
	Pin             *PinInfo
	Location        *entity.MessageLocation
	MentionsChannel bool
	MentionsHere    bool
	// 公式アプリの投稿。誰も編集・削除できない
	IsOfficial bool
	Poll       *PollOutput
}

// ForBroadcast は閲覧者ごとに変わる内容（引用カードと自分の投票）を除き、他の購読者へ配信できるコピーを返します
func (m MessageOutput) ForBroadcast() MessageOutput {
	links := make([]LinkInfo, len(m.Links))
	for i, link := range m.Links {
		link.MessagePreview = nil
		links[i] = link
	}
	m.Links = links
	if m.Poll != nil {
		poll := *m.Poll
		poll.MyOptionIDs = []string{}
		m.Poll = &poll
	}
	return m
}

type ListMessagesOutput struct {
	Messages []TimelineItem
	HasMore  bool
	HasNewer bool
}

type ThreadMetadataOutput struct {
	MessageID     string
	ReplyCount    int
	LastReplyAt   *time.Time
	LastReplyUser *UserInfo
	IsFollowing   bool
}

type GetThreadRepliesInput struct {
	MessageID string
	UserID    string
	Limit     int
	Since     *time.Time
	Until     *time.Time
	// 指定した返信の前後をまとめて返す。Since / Until より優先する
	AroundReplyID *string
}

type GetThreadRepliesOutput struct {
	ParentMessage MessageOutput
	// 古い順
	Replies    []MessageOutput
	HasMore    bool
	HasNewer   bool
	ReplyCount int
}

type MessageWithThreadOutput struct {
	MessageOutput
	ThreadMetadata *ThreadMetadataOutput
}

type ListMessagesWithThreadOutput struct {
	Messages []MessageWithThreadOutput
	HasMore  bool
}

// TimelineItem はユーザーのメッセージかシステムメッセージのどちらか一方を持ちます
type TimelineItem struct {
	UserMessage   *MessageOutput
	SystemMessage *entity.SystemMessage
	CreatedAt     time.Time
}
