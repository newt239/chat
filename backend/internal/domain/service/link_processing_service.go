package service

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

type LinkProcessingService interface {
	// ProcessLinks は本文中の URL を MessageLink にします。workspaceID は同じワークスペースのメッセージリンクの判定に使います
	ProcessLinks(ctx context.Context, body, workspaceID string) ([]*entity.MessageLink, error)
}
