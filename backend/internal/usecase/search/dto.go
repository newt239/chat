package search

import (
	"errors"

	channeluc "github.com/newt239/chat/internal/usecase/channel"
	messageuc "github.com/newt239/chat/internal/usecase/message"
	usergroupuc "github.com/newt239/chat/internal/usecase/user_group"
	workspaceuc "github.com/newt239/chat/internal/usecase/workspace"
)

type SearchFilter string

const (
	SearchFilterAll      SearchFilter = "all"
	SearchFilterMessages SearchFilter = "messages"
	SearchFilterChannels SearchFilter = "channels"
	SearchFilterUsers    SearchFilter = "users"
	SearchFilterGroups   SearchFilter = "groups"
)

var (
	ErrInvalidQuery      = errors.New("検索キーワードを入力してください")
	ErrWorkspaceNotFound = errors.New("ワークスペースが見つかりません")
	ErrUnauthorized      = errors.New("このワークスペースを検索する権限がありません")
)

type WorkspaceSearchInput struct {
	WorkspaceID string
	RequesterID string
	Query       string
	Filter      SearchFilter
	Page        int
	PerPage     int
}

type PaginatedMessages struct {
	Items   []messageuc.MessageOutput `json:"items"`
	Total   int                       `json:"total"`
	Page    int                       `json:"page"`
	PerPage int                       `json:"perPage"`
	HasMore bool                      `json:"hasMore"`
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

// Normalize はサポートされていないフィルターを all に丸めます
func (f SearchFilter) Normalize() SearchFilter {
	switch f {
	case SearchFilterAll, SearchFilterMessages, SearchFilterChannels, SearchFilterUsers, SearchFilterGroups:
		return f
	default:
		return SearchFilterAll
	}
}

func (f SearchFilter) includesMessages() bool {
	return f == SearchFilterAll || f == SearchFilterMessages
}

func (f SearchFilter) includesChannels() bool {
	return f == SearchFilterAll || f == SearchFilterChannels
}

func (f SearchFilter) includesUsers() bool {
	return f == SearchFilterAll || f == SearchFilterUsers
}

func (f SearchFilter) includesGroups() bool {
	return f == SearchFilterAll || f == SearchFilterGroups
}
