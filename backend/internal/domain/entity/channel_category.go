package entity

// ChannelCategory はユーザーが自分のサイドバーに作るチャンネルのカテゴリです
type ChannelCategory struct {
	ID          string
	UserID      string
	WorkspaceID string
	Name        string
	Position    int
	ChannelIDs  []string
}
