package search

import (
	"time"

	"github.com/newt239/chat/internal/domain/entity"

	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	channeluc "github.com/newt239/chat/internal/usecase/channel"
	messageuc "github.com/newt239/chat/internal/usecase/message"
	workspaceuc "github.com/newt239/chat/internal/usecase/workspace"
)

type SearchTarget string

const (
	SearchTargetAll      SearchTarget = "all"
	SearchTargetMessages SearchTarget = "messages"
	SearchTargetChannels SearchTarget = "channels"
	SearchTargetUsers    SearchTarget = "users"
	SearchTargetGroups   SearchTarget = "groups"
)

var (
	ErrInvalidQuery     = domerr.New(domerr.ErrValidation, "検索キーワードか絞り込み条件を指定してください")
	ErrInvalidDateRange = domerr.New(domerr.ErrValidation, "期間の開始は終了より前にしてください")
)

// MessageFilter はメッセージ検索の絞り込み条件です（条件はすべて AND）
type MessageFilter struct {
	FromUserIDs               []string
	ChannelIDs                []string
	IncludeDescendantChannels bool
	Has                       []domainrepository.MessageContentKind
	PinnedOnly                bool
	ThreadOnly                bool
	MentionsMe                bool
	ExcludeReplies            bool
	After                     *time.Time
	Before                    *time.Time
}

func (f MessageFilter) isEmpty() bool {
	return len(f.FromUserIDs) == 0 && len(f.ChannelIDs) == 0 && len(f.Has) == 0 &&
		!f.PinnedOnly && !f.ThreadOnly && !f.MentionsMe && f.After == nil && f.Before == nil
}

type WorkspaceSearchInput struct {
	WorkspaceID string
	RequesterID string
	Query       string
	Target      SearchTarget
	Filter      MessageFilter
	Sort        domainrepository.MessageSearchSort
	Page        int
	PerPage     int
}

// TextRange は UTF-16 コード単位のオフセットで表した [Start, End) の範囲です
type TextRange struct {
	Start int
	End   int
}

type MessageHit struct {
	Message    messageuc.MessageOutput
	Highlights []TextRange
}

type Paginated[T any] struct {
	Items   []T
	Total   int
	Page    int
	PerPage int
	HasMore bool
}

type WorkspaceSearchOutput struct {
	Messages Paginated[MessageHit]
	Channels Paginated[channeluc.ChannelOutput]
	Users    Paginated[workspaceuc.MemberInfo]
	Groups   Paginated[*entity.UserGroup]
}

// Normalize はサポートされていない検索対象を all に丸めます
func (f SearchTarget) Normalize() SearchTarget {
	switch f {
	case SearchTargetAll, SearchTargetMessages, SearchTargetChannels, SearchTargetUsers, SearchTargetGroups:
		return f
	default:
		return SearchTargetAll
	}
}

func (f SearchTarget) includesMessages() bool {
	return f == SearchTargetAll || f == SearchTargetMessages
}

func (f SearchTarget) includesChannels() bool {
	return f == SearchTargetAll || f == SearchTargetChannels
}

func (f SearchTarget) includesUsers() bool {
	return f == SearchTargetAll || f == SearchTargetUsers
}

func (f SearchTarget) includesGroups() bool {
	return f == SearchTargetAll || f == SearchTargetGroups
}
