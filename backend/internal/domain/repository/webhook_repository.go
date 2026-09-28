package repository

import (
	"context"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
)

type WebhookRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Webhook, error)
	// FindByChannelID は作成日時の昇順で返します
	FindByChannelID(ctx context.Context, channelID string) ([]*entity.Webhook, error)
	Create(ctx context.Context, webhook *entity.Webhook) error
	// Update は名前・アイコン・トークンのハッシュを保存します
	Update(ctx context.Context, webhook *entity.Webhook) error
	MarkUsed(ctx context.Context, id string, usedAt time.Time) error
	Delete(ctx context.Context, id string) error
}
