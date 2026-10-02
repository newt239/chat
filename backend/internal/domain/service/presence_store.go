package service

import "context"

// PresenceEntry は 1 接続が 1 チャンネルを閲覧していることを表します
type PresenceEntry struct {
	WorkspaceID string
	ChannelID   string
	ConnID      string
	UserID      string
}

// PresenceStore はチャンネルの閲覧者を全レプリカで共有します
// 延長されない閲覧は期限で消え、落ちたレプリカの接続が残り続けないようにします
type PresenceStore interface {
	Add(ctx context.Context, e PresenceEntry) error
	Remove(ctx context.Context, e PresenceEntry) error
	Refresh(ctx context.Context, entries []PresenceEntry) error
	// Viewers は閲覧中のユーザー ID を重複なしで昇順に返します
	Viewers(ctx context.Context, workspaceID, channelID string) ([]string, error)
}
