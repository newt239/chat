package search

import (
	"errors"

	channeluc "github.com/newt239/chat/internal/usecase/channel"
	messageuc "github.com/newt239/chat/internal/usecase/message"
	workspaceuc "github.com/newt239/chat/internal/usecase/workspace"
)

type SearchFilter string

const (
	SearchFilterAll      SearchFilter = "all"
	SearchFilterMessages SearchFilter = "messages"
	SearchFilterChannels SearchFilter = "channels"
	SearchFilterUsers    SearchFilter = "users"
)

var (
	ErrInvalidQuery      = errors.New("query must not be empty")
	ErrWorkspaceNotFound = errors.New("workspace not found")
	ErrUnauthorized      = errors.New("unauthorized to search this workspace")
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

type WorkspaceSearchOutput struct {
	Messages PaginatedMessages `json:"messages"`
	Channels PaginatedChannels `json:"channels"`
	Users    PaginatedUsers    `json:"users"`
}

// Normalize はサポートされていないフィルターを all に丸めます
func (f SearchFilter) Normalize() SearchFilter {
	switch f {
	case SearchFilterAll, SearchFilterMessages, SearchFilterChannels, SearchFilterUsers:
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
