package search

import (
	"errors"
	"time"

	domainrepository "github.com/newt239/chat/internal/domain/repository"
	channeluc "github.com/newt239/chat/internal/usecase/channel"
	messageuc "github.com/newt239/chat/internal/usecase/message"
	usergroupuc "github.com/newt239/chat/internal/usecase/user_group"
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
	ErrInvalidQuery      = errors.New("検索キーワードか絞り込み条件を指定してください")
	ErrInvalidDateRange  = errors.New("期間の開始は終了より前にしてください")
	ErrWorkspaceNotFound = errors.New("ワークスペースが見つかりません")
	ErrUnauthorized      = errors.New("このワークスペースを検索する権限がありません")
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

type PaginatedMessages struct {
	Items   []MessageHit
	Total   int
	Page    int
	PerPage int
	HasMore bool
}

type PaginatedChannels struct {
	Items   []channeluc.ChannelOutput `json:"items"`
	Total   int                       `json:"total"`
	Page    int                       `json:"page"`
	PerPage int                       `json:"perPage"`
	HasMore bool                      `json:"hasMore"`
}

type PaginatedUsers struct {
	Items   []workspaceuc.MemberInfo `json:"items"`
	Total   int                      `json:"total"`
	Page    int                      `json:"page"`
	PerPage int                      `json:"perPage"`
	HasMore bool                     `json:"hasMore"`
}

type PaginatedUserGroups struct {
	Items   []usergroupuc.UserGroupOutput `json:"items"`
	Total   int                           `json:"total"`
	Page    int                           `json:"page"`
	PerPage int                           `json:"perPage"`
	HasMore bool                          `json:"hasMore"`
}

type WorkspaceSearchOutput struct {
	Messages PaginatedMessages   `json:"messages"`
	Channels PaginatedChannels   `json:"channels"`
	Users    PaginatedUsers      `json:"users"`
	Groups   PaginatedUserGroups `json:"groups"`
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
