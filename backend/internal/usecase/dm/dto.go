package dm

import "github.com/newt239/chat/internal/domain/entity"

type CreateDMInput struct {
	WorkspaceID  string
	UserID       string
	TargetUserID string
}

type CreateGroupDMInput struct {
	WorkspaceID string
	CreatorID   string
	MemberIDs   []string
	Name        string
}

type ListDMsInput struct {
	WorkspaceID string
	UserID      string
}

type DMOutput struct {
	*entity.Channel
	// 自分以外の参加者
	Members     []*entity.User
	IsStarred   bool
	IsMuted     bool
	UnreadCount int
}
