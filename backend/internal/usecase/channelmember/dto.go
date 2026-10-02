package channelmember

import (
	"time"

	"github.com/newt239/chat/internal/domain/entity"
)

type ListMembersInput struct {
	ChannelID string
	UserID    string
}

type InviteMemberInput struct {
	ChannelID    string
	OperatorID   string
	TargetUserID string
	Role         entity.ChannelRole
}

type JoinChannelInput struct {
	ChannelID string
	UserID    string
}

type UpdateMemberRoleInput struct {
	ChannelID    string
	OperatorID   string
	TargetUserID string
	Role         entity.ChannelRole
}

type RemoveMemberInput struct {
	ChannelID    string
	OperatorID   string
	TargetUserID string
}

type LeaveChannelInput struct {
	ChannelID string
	UserID    string
}

type MemberInfo struct {
	UserID      string             `json:"userId"`
	Role        entity.ChannelRole `json:"role"`
	JoinedAt    time.Time          `json:"joinedAt"`
	DisplayName string             `json:"displayName"`
	Email       string             `json:"email"`
	AvatarURL   *string            `json:"avatarUrl"`
}

type MemberListOutput struct {
	Members []MemberInfo `json:"members"`
}
