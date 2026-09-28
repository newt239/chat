package repository

import "time"

type MessageContentKind string

const (
	MessageContentImage MessageContentKind = "image"
	MessageContentFile  MessageContentKind = "file"
	MessageContentLink  MessageContentKind = "link"
	MessageContentVideo MessageContentKind = "video"
)

type MessageSearchSort string

const (
	MessageSearchSortNewest    MessageSearchSort = "newest"
	MessageSearchSortRelevance MessageSearchSort = "relevance"
)

// MessageSearchCriteria は閲覧者が見られるチャンネルに限定したメッセージ検索の条件です（条件はすべて AND）
type MessageSearchCriteria struct {
	WorkspaceID    string
	ViewerID       string
	Terms          []string
	ChannelIDs     []string
	AuthorIDs      []string
	Has            []MessageContentKind
	PinnedOnly     bool
	ThreadOnly     bool
	ExcludeReplies bool
	MentionsViewer bool
	After          *time.Time
	Before         *time.Time
	Sort           MessageSearchSort
	Limit          int
	Offset         int
}

type MessageCursor struct {
	CreatedAt time.Time
	MessageID string
}

// FindMentionsInput は UserID 宛てのメンションを含む他人のメッセージを Cursor より古いものから Limit 件探す条件です
type FindMentionsInput struct {
	WorkspaceID string
	UserID      string
	Cursor      *MessageCursor
	Limit       int
}
