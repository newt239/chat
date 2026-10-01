package repository

import (
	"context"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
)

type AppRepository interface {
	FindByID(ctx context.Context, id string) (*entity.App, error)
	// FindByWorkspaceID は作成日時の昇順で返します
	FindByWorkspaceID(ctx context.Context, workspaceID string) ([]*entity.App, error)
	FindOfficial(ctx context.Context, workspaceID string) (*entity.App, error)
	// FindByChannelID はボットユーザーがチャンネルに参加しているアプリを返します
	FindByChannelID(ctx context.Context, channelID string) ([]*entity.App, error)
	Create(ctx context.Context, app *entity.App) error
	// Update は ID・ワークスペース・ボットユーザー・作成者以外を保存します
	Update(ctx context.Context, app *entity.App) error
	MarkUsed(ctx context.Context, id string, usedAt time.Time) error
	Delete(ctx context.Context, id string) error
}
