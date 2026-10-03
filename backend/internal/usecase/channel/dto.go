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
	ID           string
	WorkspaceID  string
	Name         string
	Description  *string
	IsPrivate    bool
	CreatedBy    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	UnreadCount  int
	MentionCount int
	ParentID     *string
	IsStarred    bool
	IsMuted      bool
	IsMember     bool
	// ListChannels でだけ設定する
	LastMessageAt *time.Time
}

func NewChannelOutput(ch *entity.Channel) ChannelOutput {
	return ChannelOutput{
		ID:          ch.ID,
		WorkspaceID: ch.WorkspaceID,
		Name:        ch.Name,
		Description: ch.Description,
		IsPrivate:   ch.IsPrivate(),
		CreatedBy:   ch.CreatedBy,
		CreatedAt:   ch.CreatedAt,
		UpdatedAt:   ch.UpdatedAt,
		ParentID:    ch.ParentID,
	}
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
