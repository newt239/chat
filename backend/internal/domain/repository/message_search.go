package repository

import (
	"context"
	"time"
)

type MessageContentKind string

const (
	MessageContentImage    MessageContentKind = "image"
	MessageContentFile     MessageContentKind = "file"
	MessageContentLink     MessageContentKind = "link"
	MessageContentVideo    MessageContentKind = "video"
	MessageContentLocation MessageContentKind = "location"
)

type MessageSearchSort string

const (
	MessageSearchSortNewest    MessageSearchSort = "newest"
	MessageSearchSortRelevance MessageSearchSort = "relevance"
)

// MessageSearchIndex はメッセージの全文検索インデックスです
type MessageSearchIndex interface {
	Search(ctx context.Context, criteria MessageSearchCriteria) (*MessageSearchResult, error)
	// Upsert は同じ ID の文書を置き換えます
	Upsert(ctx context.Context, documents []MessageSearchDocument) error
	Delete(ctx context.Context, messageIDs []string) error
	DeleteAll(ctx context.Context) error
}

// MessageSearchCriteria は全文検索の条件です（条件はすべて AND）
type MessageSearchCriteria struct {
	WorkspaceID string
	Terms       []string
	// ChannelIDs は閲覧者が見られるものに絞り込み済みの検索対象で、空にしない
	ChannelIDs     []string
	AuthorIDs      []string
	Has            []MessageContentKind
	PinnedOnly     bool
	ThreadOnly     bool
	ExcludeReplies bool
	// Mention を指定すると、その閲覧者宛てのメッセージに絞り込む
	Mention *MessageSearchScope
	After   *time.Time
	Before  *time.Time
	Sort    MessageSearchSort
	Page    int
	PerPage int
}

type MessageSearchResult struct {
	MessageIDs []string
	Total      int
}

// MessageSearchScope は閲覧者が検索できる範囲と、自分宛てかを判定するための所属です
type MessageSearchScope struct {
	UserID             string
	ViewableChannelIDs []string
	// JoinedChannelIDs は参加しているチャンネル（@channel / @here が届く範囲）。グループへのメンションは投稿時点のメンバーに展開済み
	JoinedChannelIDs []string
}

// MessageSearchDocument は検索インデックスに載せる、削除されていないメッセージの内容です
type MessageSearchDocument struct {
	ID                string
	WorkspaceID       string
	ChannelID         string
	SenderID          string
	ParentID          *string
	Body              string
	AttachmentNames   []string
	Has               []MessageContentKind
	MentionedUserIDs  []string
	MentionedGroupIDs []string
	// MentionsChannel は本文が <@channel> / <@here> を含むかどうか
	MentionsChannel bool
	Pinned          bool
	HasReplies      bool
	CreatedAt       time.Time
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
