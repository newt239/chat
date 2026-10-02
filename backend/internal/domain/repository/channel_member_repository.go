package repository

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

type ChannelMemberRepository interface {
	// AddMember は既に参加していれば ErrAlreadyMember を返します
	AddMember(ctx context.Context, member *entity.ChannelMember) error
	// FindMember は参加していなければ nil を返します
	FindMember(ctx context.Context, channelID string, userID string) (*entity.ChannelMember, error)
	RemoveMember(ctx context.Context, channelID string, userID string) error
	FindMembers(ctx context.Context, channelID string) ([]*entity.ChannelMember, error)
	IsMember(ctx context.Context, channelID string, userID string) (bool, error)
	UpdateMemberRole(ctx context.Context, channelID string, userID string, role entity.ChannelRole) error
	CountAdmins(ctx context.Context, channelID string) (int, error)
	// FindJoinedChannelIDs は channelIDs のうち userID が参加しているチャンネルを返します
	FindJoinedChannelIDs(ctx context.Context, userID string, channelIDs []string) (map[string]bool, error)
	// FindMemberIDsIn は userIDs のうちチャンネルに参加しているユーザーを返します
	FindMemberIDsIn(ctx context.Context, channelID string, userIDs []string) (map[string]bool, error)
}
