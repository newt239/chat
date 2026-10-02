package service

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

type LinkProcessingService interface {
	// PrepareLinks は本文中の URL をメッセージリンクにし、外部の URL はプレビューを保存してから返します
	// OGP を取りに行くためトランザクションの外で呼びます。workspaceID は同じワークスペースのメッセージリンクの判定に使います
	PrepareLinks(ctx context.Context, body, workspaceID string) ([]*entity.MessageLink, error)
}
