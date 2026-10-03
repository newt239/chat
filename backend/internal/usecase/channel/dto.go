package channel

import (
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

type CreateChannelInput struct {
	WorkspaceID string
	UserID      string
	Name        string
	Description *string
	IsPrivate   bool
	MemberIDs   []string
}

type UpdateChannelInput struct {
	ChannelID   string
	UserID      string
	Name        *string
	Description *string
	IsPrivate   *bool
}

// SetFlagInput はスターやミュートを付け外しします
type SetFlagInput struct {
	ChannelID string
	UserID    string
	Value     bool
}

type ChannelOutput struct {
	*entity.Channel
	UnreadCount  int
	MentionCount int
	IsStarred    bool
	IsMuted      bool
	IsMember     bool
	// ListChannels でだけ設定する
	LastMessageAt *time.Time
}

type BrowsableChannelOutput struct {
	Channel     ChannelOutput
	MemberCount int
}

type SearchBrowsableChannelsInput struct {
	WorkspaceID string
	UserID      string
	Query       string
	Membership  domainrepository.BrowsableChannelMembership
	Sort        domainrepository.BrowsableChannelSort
	// 1 始まり
	Page    int
	PerPage int
}

type SearchBrowsableChannelsOutput struct {
	Channels []BrowsableChannelOutput
	Total    int
}
