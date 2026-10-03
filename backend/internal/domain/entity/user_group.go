package entity

type UserGroup struct {
	ID          string
	WorkspaceID string
	Name        string
	Description *string
	CreatedBy   string
}

type UserGroupMember struct {
	GroupID string
	UserID  string
}
