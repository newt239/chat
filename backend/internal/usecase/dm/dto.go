package dm

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
	ID          string
	WorkspaceID string
	Name        string
	Description *string
	Type        string
	Members     []DMMemberOutput
	CreatedAt   string
	UpdatedAt   string
	IsStarred   bool
	IsMuted     bool
	UnreadCount int
	HasMention  bool
}

type DMMemberOutput struct {
	UserID      string
	DisplayName string
	AvatarURL   *string
}
