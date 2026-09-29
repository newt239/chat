package repository

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

// DraftTarget は下書きの置き場所（ユーザー・チャンネル・スレッド）です
type DraftTarget struct {
	UserID    string
	ChannelID string
	ParentID  *string
}

type DraftRepository interface {
	Find(ctx context.Context, target DraftTarget) (*entity.Draft, error)
	// FindByWorkspace はワークスペース内のチャンネルにある userID の下書きを新しい順に返します
	FindByWorkspace(ctx context.Context, userID string, workspaceID string) ([]*entity.Draft, error)
	// Upsert は同じ置き場所の下書きを本文で置き換え、ID と更新日時を設定します
	Upsert(ctx context.Context, draft *entity.Draft) error
	Delete(ctx context.Context, target DraftTarget) error
}
