package repository

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

// PinRepository はメッセージのピン留めを管理します
type PinRepository interface {
	// Create はピンを作成します。重複する (channel,message) の場合はエラーを返します
	Create(ctx context.Context, pin *entity.MessagePin) error
	// Delete はピンを削除します（存在しなくてもエラーにしない）
	Delete(ctx context.Context, channelID, messageID string) error
	// List は指定チャンネルのピン一覧を新しい順で返します（最大 limit 件）。
	// cursor は pinned_at のカーソル（RFC3339 文字列等）を想定し、cursor より古いものを返します。
	List(ctx context.Context, channelID string, limit int, cursor *string) ([]*entity.MessagePin, *string, error)
}
