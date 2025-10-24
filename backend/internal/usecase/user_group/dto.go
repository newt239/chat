package user_group

import "time"

// Input DTOs

type CreateUserGroupInput struct {
	WorkspaceID string
	Name        string
	Description *string
	CreatedBy   string
}

type UpdateUserGroupInput struct {
	ID          string
	Name        *string
	Description *string
	UpdatedBy   string
}

type DeleteUserGroupInput struct {
	ID       string
	DeletedBy string
}

type GetUserGroupInput struct {
	ID       string
	UserID   string // For authorization check
}

type ListUserGroupsInput struct {
	WorkspaceID string
	UserID      string // For authorization check
}

type AddMemberInput struct {
	GroupID string
	UserID  string
	AddedBy string // User performing the action
}

type RemoveMemberInput struct {
	GroupID   string
	UserID    string
	RemovedBy string // User performing the action
}

type ListMembersInput struct {
	GroupID string
	UserID  string // For authorization check
}

// Output DTOs

type UserGroupOutput struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspaceId"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	CreatedBy   string    `json:"createdBy"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type CreateUserGroupOutput struct {
	UserGroup UserGroupOutput `json:"userGroup"`
}

type UpdateUserGroupOutput struct {
	UserGroup UserGroupOutput `json:"userGroup"`
}

type DeleteUserGroupOutput struct {
	Success bool `json:"success"`
}

type GetUserGroupOutput struct {
	UserGroup UserGroupOutput `json:"userGroup"`
}

type ListUserGroupsOutput struct {
	UserGroups []UserGroupOutput `json:"userGroups"`
}

type MemberInfo struct {
	UserID      string    `json:"userId"`
	DisplayName string    `json:"displayName"`
	AvatarURL   *string   `json:"avatarUrl"`
	JoinedAt    time.Time `json:"joinedAt"`
}

type AddMemberOutput struct {
	Success bool `json:"success"`
}

type RemoveMemberOutput struct {
	Success bool `json:"success"`
}

type ListMembersOutput struct {
	Members []MemberInfo `json:"members"`
}
