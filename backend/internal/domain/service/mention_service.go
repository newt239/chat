package service

import "context"

// ResolvedMentions は本文の ID 記法のうち、ワークスペースに実在するものを投稿時点で解決した結果です
type ResolvedMentions struct {
	UserIDs  []string
	GroupIDs []string
	// グループ ID ごとの、解決した時点のメンバー
	GroupMembers map[string][]string
	Channel      bool
	Here         bool
}

type MentionService interface {
	// Resolve は本文の ID 記法を検証して解決します。knownGroups に含まれるグループは展開せず、呼び出し側が以前の展開結果を使います
	Resolve(ctx context.Context, body, workspaceID string, knownGroups []string) (*ResolvedMentions, error)
	// RenderPlain は本文の ID 記法を現在の名前に置き換えた文字列を返します
	RenderPlain(ctx context.Context, body string) (string, error)
}
