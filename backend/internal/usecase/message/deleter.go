package message

import (
	"context"
	"fmt"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
)

// DeleteMessage はメッセージとスレッドの返信を削除します。他人のメッセージは権限設定で許可されたロールだけが削除できます
func (i *Interactor) DeleteMessage(ctx context.Context, input MessageInput) error {
	message, channel, err := i.channelAccessSvc.EnsureMessageAccess(ctx, input.MessageID, input.UserID)
	if err != nil {
		return err
	}
	if message.DeletedAt != nil {
		return ErrMessageAlreadyDeleted
	}
	author, err := i.userRepo.FindByID(ctx, message.UserID)
	if err != nil {
		return fmt.Errorf("failed to load author: %w", err)
	}
	if author != nil && author.IsOfficial {
		return ErrOfficialMessage
	}
	if message.UserID != input.UserID {
		if _, err := i.permissionSvc.Ensure(ctx, channel.WorkspaceID, input.UserID, entity.PermissionDeleteOthersMessages); err != nil {
			return err
		}
	}

	deleteIDs := []string{message.ID}
	if message.ParentID == nil {
		replies, err := i.messageRepo.FindThreadReplies(ctx, message.ID, 0, nil, nil, true)
		if err != nil {
			return fmt.Errorf("failed to fetch replies: %w", err)
		}
		for _, reply := range replies {
			deleteIDs = append(deleteIDs, reply.ID)
		}
	}
	if err := i.messageRepo.SoftDeleteByIDs(ctx, deleteIDs, input.UserID); err != nil {
		return fmt.Errorf("failed to delete messages: %w", err)
	}

	// 返信が消えると親メッセージの「スレッドあり」も変わる
	indexIDs := deleteIDs
	if message.ParentID != nil {
		indexIDs = append(indexIDs, *message.ParentID)
	}
	i.searchIndexer.Sync(ctx, indexIDs...)

	i.notifier.NotifyDeletedMessage(channel.WorkspaceID, channel.ID, MessageDeletion{
		MessageID:  message.ID,
		DeletedIDs: deleteIDs,
		DeletedAt:  time.Now(),
	})
	return nil
}
