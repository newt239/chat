package repository

import (
	"context"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
)

type ScheduledMessageRepository interface {
	Create(ctx context.Context, message *entity.ScheduledMessage) error
	FindByID(ctx context.Context, id string) (*entity.ScheduledMessage, error)
	// FindByWorkspace はワークスペース内のチャンネル宛ての userID の予約を予約日時の順に返します
	FindByWorkspace(ctx context.Context, userID string, workspaceID string) ([]*entity.ScheduledMessage, error)
	// Reschedule は本文と日時を変え、予約中に戻します
	Reschedule(ctx context.Context, id string, body string, scheduledAt time.Time) error
	Delete(ctx context.Context, id string) error
	// ClaimDue は期限の来た予約を最大 limit 件、送信中にして返します。
	// 行ロックを SKIP LOCKED で取るため、複数のサーバーが同時に動いても同じ予約を二重に取り出さない
	ClaimDue(ctx context.Context, now time.Time, limit int) ([]*entity.ScheduledMessage, error)
	// Claim は予約中か失敗した予約を 1 件だけ送信中にします。取り出せなければ nil を返します
	Claim(ctx context.Context, id string) (*entity.ScheduledMessage, error)
	MarkSent(ctx context.Context, id string, messageID string) error
	MarkFailed(ctx context.Context, id string, reason string) error
}
