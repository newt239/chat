// Package audit は各ユースケースから監査ログを記録するための部品を提供します
package audit

import (
	"context"
	"log/slog"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

// ClientInfo は操作元の端末の情報です。インターフェース層がリクエストから取り出して context に載せます
type ClientInfo struct {
	IPAddress string
	UserAgent string
}

type clientInfoKey struct{}

func WithClientInfo(ctx context.Context, info ClientInfo) context.Context {
	return context.WithValue(ctx, clientInfoKey{}, info)
}

func ClientInfoFrom(ctx context.Context) ClientInfo {
	info, _ := ctx.Value(clientInfoKey{}).(ClientInfo)
	return info
}

// Recorder は監査ログを記録します。記録に失敗しても元の操作は取り消さず、エラーログに残します
type Recorder interface {
	Record(ctx context.Context, log entity.AuditLog)
}

type recorder struct {
	repo domainrepository.AuditLogRepository
}

func NewRecorder(repo domainrepository.AuditLogRepository) Recorder {
	return &recorder{repo: repo}
}

func (r *recorder) Record(ctx context.Context, log entity.AuditLog) {
	info := ClientInfoFrom(ctx)
	log.IPAddress = info.IPAddress
	log.UserAgent = info.UserAgent
	if err := r.repo.Create(ctx, &log); err != nil {
		slog.ErrorContext(ctx, "監査ログの記録に失敗しました", "workspaceId", log.WorkspaceID, "action", log.Action, "error", err)
	}
}
