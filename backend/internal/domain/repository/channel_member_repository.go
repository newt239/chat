package repository

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

type ChannelMemberRepository interface {
	AddMember(ctx context.Context, member *entity.ChannelMember) error
	RemoveMember(ctx context.Context, channelID string, userID string) error
	FindMembers(ctx context.Context, channelID string) ([]*entity.ChannelMember, error)
	IsMember(ctx context.Context, channelID string, userID string) (bool, error)
	UpdateMemberRole(ctx context.Context, channelID string, userID string, role entity.ChannelRole) error
	CountAdmins(ctx context.Context, channelID string) (int, error)
}
