package systemmessage

import (
	"context"
	"fmt"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

type CreateInput struct {
	ChannelID string
	Kind      entity.SystemMessageKind
	Payload   map[string]any
	ActorID   *string
}

type UseCase interface {
	Create(ctx context.Context, input CreateInput) (*entity.SystemMessage, error)
}

type interactor struct {
	systemMsgRepo domainrepository.SystemMessageRepository
	channelRepo   domainrepository.ChannelRepository
	notification  Notifier
}

func New(systemMsgRepo domainrepository.SystemMessageRepository, channelRepo domainrepository.ChannelRepository, notification Notifier) UseCase {
	return &interactor{
		systemMsgRepo: systemMsgRepo,
		channelRepo:   channelRepo,
		notification:  notification,
	}
}

func (i *interactor) Create(ctx context.Context, input CreateInput) (*entity.SystemMessage, error) {
	if input.ChannelID == "" {
		return nil, fmt.Errorf("%w: channel id is required", domerr.ErrValidation)
	}
	if input.Kind == "" {
		return nil, fmt.Errorf("%w: kind is required", domerr.ErrValidation)
	}

	msg := &entity.SystemMessage{
		ChannelID: input.ChannelID,
		Kind:      input.Kind,
		Payload:   input.Payload,
		ActorID:   input.ActorID,
		CreatedAt: time.Now(),
	}

	if err := i.systemMsgRepo.Create(ctx, msg); err != nil {
		return nil, err
	}

	// 通知（workspaceID はチャネルから解決）
	ch, err := i.channelRepo.FindByID(ctx, input.ChannelID)
	if err == nil && ch != nil && i.notification != nil {
		i.notification.NotifySystemMessageCreated(ch.WorkspaceID, input.ChannelID, msg)
	}

	return msg, nil
}
