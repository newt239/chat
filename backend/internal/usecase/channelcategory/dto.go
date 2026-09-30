package channelcategory

type ListInput struct {
	WorkspaceID string
	UserID      string
}

type CreateInput struct {
	WorkspaceID string
	UserID      string
	Name        string
}

type UpdateInput struct {
	CategoryID string
	UserID     string
	Name       string
}

type DeleteInput struct {
	CategoryID string
	UserID     string
}

type ReorderInput struct {
	WorkspaceID string
	UserID      string
	CategoryIDs []string
}

type SetChannelInput struct {
	ChannelID  string
	UserID     string
	CategoryID *string
}

type CategoryOutput struct {
	ID         string
	Name       string
	Position   int
	ChannelIDs []string
}
