// Package audittest はユースケースのテストで記録された監査ログを検証するための Recorder を提供します
package audittest

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

type Recorder struct {
	Logs []entity.AuditLog
}

func (r *Recorder) Record(_ context.Context, log entity.AuditLog) {
	r.Logs = append(r.Logs, log)
}

func (r *Recorder) Actions() []entity.AuditAction {
	actions := make([]entity.AuditAction, 0, len(r.Logs))
	for _, l := range r.Logs {
		actions = append(actions, l.Action)
	}
	return actions
}
