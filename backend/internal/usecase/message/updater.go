package message

import (
	"context"
	"fmt"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/domain/transaction"
)

// MessageUpdater はメッセージの編集を担当するユースケースです
type MessageUpdater struct {
	messageRepo        domainrepository.MessageRepository
	notificationSvc    Notifier
	recorder           *ContentRecorder
	transactionManager transaction.Manager
	outputBuilder      *MessageOutputBuilder
	channelAccessSvc   service.ChannelAccessService
	searchIndexer      SearchIndexer
}

func NewMessageUpdater(
	messageRepo domainrepository.MessageRepository,
	notificationSvc Notifier,
	recorder *ContentRecorder,
	transactionManager transaction.Manager,
	outputBuilder *MessageOutputBuilder,
	channelAccessSvc service.ChannelAccessService,
	searchIndexer SearchIndexer,
) *MessageUpdater {
	return &MessageUpdater{
		messageRepo:        messageRepo,
		notificationSvc:    notificationSvc,
		recorder:           recorder,
		transactionManager: transactionManager,
		outputBuilder:      outputBuilder,
		channelAccessSvc:   channelAccessSvc,
		searchIndexer:      searchIndexer,
	}
}

// UpdateMessage は本文を編集します。編集できるのは投稿者本人だけで、アーカイブしたチャンネルでは編集できません
func (u *MessageUpdater) UpdateMessage(ctx context.Context, input UpdateMessageInput) (*MessageOutput, error) {
	message, err := u.messageRepo.FindByID(ctx, input.MessageID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch message: %w", err)
	}
	if message == nil {
		return nil, domerr.ErrMessageNotFound
	}
	channel, err := u.channelAccessSvc.EnsureChannelAccess(ctx, message.ChannelID, input.EditorID)
	if err != nil {
		return nil, err
	}
	if message.UserID != input.EditorID {
		return nil, domerr.ErrUnauthorized
	}
	if message.DeletedAt != nil {
		return nil, ErrCannotEditDeleted
	}
	if channel.ArchivedAt != nil {
		return nil, domerr.ErrChannelArchived
	}

	previous, err := u.recorder.previousMentions(ctx, message.ID)
	if err != nil {
		return nil, err
	}
	content, err := u.recorder.prepare(ctx, input.Body, channel.WorkspaceID, previous)
	if err != nil {
		return nil, err
	}

	message.Body = input.Body
	applyBroadcastMentions(message)
	message.EditedAt = new(time.Now())

	var result *MessageOutput
	err = u.transactionManager.Do(ctx, func(txCtx context.Context) error {
		if err := u.messageRepo.Update(txCtx, message); err != nil {
			return fmt.Errorf("failed to update message: %w", err)
		}
		if err := u.recorder.replace(txCtx, message.ID, content); err != nil {
			return err
		}
		outputs, err := u.outputBuilder.Build(txCtx, input.EditorID, []*entity.Message{message})
		if err != nil {
			return err
		}
		result = &outputs[0]
		return nil
	})
	if err != nil {
		return nil, err
	}

	u.searchIndexer.Sync(ctx, message.ID)
	u.notificationSvc.NotifyUpdatedMessage(channel.WorkspaceID, channel.ID, result.ForBroadcast())
	return result, nil
}
