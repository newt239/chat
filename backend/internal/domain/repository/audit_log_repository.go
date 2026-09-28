package repository

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

// AuditLogRepository は監査ログの保存先です。将来 NoSQL（Firestore など）に移せるよう、
// 他のテーブルとの結合・総件数・オフセット・メインの DB とのトランザクションに依存しない操作だけを持ちます
type AuditLogRepository interface {
	Create(ctx context.Context, log *entity.AuditLog) error
	// List は新しい順にカーソルでページングして返します
	List(ctx context.Context, filter entity.AuditLogFilter) (*entity.AuditLogPage, error)
}
