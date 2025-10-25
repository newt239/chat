package repository

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

type UserGroupRepository interface {
	FindByID(ctx context.Context, id string) (*entity.UserGroup, error)
	FindByWorkspaceID(ctx context.Context, workspaceID string) ([]*entity.UserGroup, error)
	FindByName(ctx context.Context, workspaceID string, name string) (*entity.UserGroup, error)
	Create(ctx context.Context, group *entity.UserGroup) error
	Update(ctx context.Context, group *entity.UserGroup) error
	Delete(ctx context.Context, id string) error
	AddMember(ctx context.Context, member *entity.UserGroupMember) error
	RemoveMember(ctx context.Context, groupID string, userID string) error
	FindMembersByGroupID(ctx context.Context, groupID string) ([]*entity.UserGroupMember, error)
	FindGroupsByUserID(ctx context.Context, userID string) ([]*entity.UserGroup, error)
	IsMember(ctx context.Context, groupID string, userID string) (bool, error)
}
