package message

import (
	"context"
	"log/slog"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

// SystemMessages はチャンネルの変更をタイムラインに残して購読者へ配信します。失敗しても元の操作は取り消さずログに残す
type SystemMessages struct {
	repo     domainrepository.SystemMessageRepository
	notifier Notifier
}

func NewSystemMessages(repo domainrepository.SystemMessageRepository, notifier Notifier) *SystemMessages {
	return &SystemMessages{repo: repo, notifier: notifier}
}

func (s *SystemMessages) Record(ctx context.Context, ch *entity.Channel, kind entity.SystemMessageKind, actorID string, payload map[string]any) {
	msg := &entity.SystemMessage{ChannelID: ch.ID, Kind: kind, Payload: payload, ActorID: &actorID}
	if err := s.repo.Create(ctx, msg); err != nil {
		slog.WarnContext(ctx, "システムメッセージを残せません", "channelId", ch.ID, "kind", kind, "error", err)
		return
	}
	s.notifier.NotifySystemMessageCreated(ch.WorkspaceID, ch.ID, msg)
}
