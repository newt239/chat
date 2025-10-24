package domain

import "time"

type UserGroup struct {
	ID          string
	WorkspaceID string
	Name        string
	Description *string
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type UserGroupMember struct {
	GroupID   string
	UserID    string
	JoinedAt  time.Time
}

type UserGroupRepository interface {
	FindByID(id string) (*UserGroup, error)
	FindByWorkspaceID(workspaceID string) ([]*UserGroup, error)
	FindByName(workspaceID, name string) (*UserGroup, error)
	Create(group *UserGroup) error
	Update(group *UserGroup) error
	Delete(id string) error
	AddMember(member *UserGroupMember) error
	RemoveMember(groupID, userID string) error
	FindMembersByGroupID(groupID string) ([]*UserGroupMember, error)
	FindGroupsByUserID(userID string) ([]*UserGroup, error)
	IsMember(groupID, userID string) (bool, error)
}
