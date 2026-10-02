package systemmessage

import (
	"context"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

type CreateInput struct {
	Channel *entity.Channel
	Kind    entity.SystemMessageKind
	Payload map[string]any
	ActorID *string
}

type UseCase interface {
	Create(ctx context.Context, input CreateInput) (*entity.SystemMessage, error)
}

type interactor struct {
	systemMsgRepo domainrepository.SystemMessageRepository
	notification  Notifier
}

func New(systemMsgRepo domainrepository.SystemMessageRepository, notification Notifier) UseCase {
	return &interactor{systemMsgRepo: systemMsgRepo, notification: notification}
}

// Create はシステムメッセージを保存し、チャンネルの購読者に配信します
func (i *interactor) Create(ctx context.Context, input CreateInput) (*entity.SystemMessage, error) {
	msg := &entity.SystemMessage{
		ChannelID: input.Channel.ID,
		Kind:      input.Kind,
		Payload:   input.Payload,
		ActorID:   input.ActorID,
		CreatedAt: time.Now(),
	}
	if err := i.systemMsgRepo.Create(ctx, msg); err != nil {
		return nil, err
	}
	i.notification.NotifySystemMessageCreated(input.Channel.WorkspaceID, input.Channel.ID, msg)
	return msg, nil
}
