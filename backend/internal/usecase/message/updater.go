package message

import (
	"context"
	"fmt"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
)

// UpdateMessage は本文を編集します。編集できるのは投稿者本人だけです
func (i *Interactor) UpdateMessage(ctx context.Context, input UpdateMessageInput) (*MessageOutput, error) {
	message, channel, err := i.channelAccessSvc.EnsureMessageAccess(ctx, input.MessageID, input.EditorID)
	if err != nil {
		return nil, err
	}
	if message.UserID != input.EditorID {
		return nil, domerr.ErrUnauthorized
	}
	if message.DeletedAt != nil {
		return nil, ErrCannotEditDeleted
	}

	previous, _, err := i.mentionRepo.FindByMessageIDs(ctx, []string{message.ID})
	if err != nil {
		return nil, fmt.Errorf("failed to load user mentions: %w", err)
	}
	content, err := i.prepareContent(ctx, input.Body, channel.WorkspaceID, previous)
	if err != nil {
		return nil, err
	}

	message.Body = input.Body
	applyBroadcastMentions(message)
	message.EditedAt = new(time.Now())

	var result *MessageOutput
	err = i.txManager.Do(ctx, func(txCtx context.Context) error {
		if err := i.messageRepo.Update(txCtx, message); err != nil {
			return fmt.Errorf("failed to update message: %w", err)
		}
		if err := i.replaceContent(txCtx, message.ID, content); err != nil {
			return err
		}
		outputs, err := i.outputBuilder.Build(txCtx, input.EditorID, []*entity.Message{message})
		if err != nil {
			return err
		}
		result = &outputs[0]
		return nil
	})
	if err != nil {
		return nil, err
	}

	i.searchIndexer.Sync(ctx, message.ID)
	i.notifier.NotifyUpdatedMessage(channel.WorkspaceID, channel.ID, result.ForBroadcast())
	return result, nil
}
