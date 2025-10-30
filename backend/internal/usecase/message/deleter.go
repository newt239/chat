package message

import (
	"context"
	"fmt"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
    "github.com/newt239/chat/internal/domain/service"
    domainservice "github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/infrastructure/logger"
	"go.uber.org/zap"
)

// MessageDeleter はメッセージ削除を担当するユースケースです
type MessageDeleter struct {
	messageRepo       domainrepository.MessageRepository
	channelRepo       domainrepository.ChannelRepository
	channelMemberRepo domainrepository.ChannelMemberRepository
	workspaceRepo     domainrepository.WorkspaceRepository
	threadRepo        domainrepository.ThreadRepository
	notificationSvc   service.NotificationService
    channelAccessSvc  domainservice.ChannelAccessService
}

// NewMessageDeleter は新しいMessageDeleterを作成します
func NewMessageDeleter(
	messageRepo domainrepository.MessageRepository,
	channelRepo domainrepository.ChannelRepository,
	channelMemberRepo domainrepository.ChannelMemberRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	threadRepo domainrepository.ThreadRepository,
	notificationSvc service.NotificationService,
    channelAccessSvc domainservice.ChannelAccessService,
) *MessageDeleter {
	return &MessageDeleter{
		messageRepo:       messageRepo,
		channelRepo:       channelRepo,
		channelMemberRepo: channelMemberRepo,
		workspaceRepo:     workspaceRepo,
		threadRepo:        threadRepo,
		notificationSvc:   notificationSvc,
        channelAccessSvc:  channelAccessSvc,
	}
}

// DeleteMessage はメッセージを削除します
func (d *MessageDeleter) DeleteMessage(ctx context.Context, input DeleteMessageInput) error {
	// メッセージ存在確認
	message, err := d.messageRepo.FindByID(ctx, input.MessageID)
	if err != nil {
		return fmt.Errorf("メッセージの取得に失敗しました: %w", err)
	}
	if message == nil {
		return ErrMessageNotFound
	}

	// チャンネルアクセス確認
    channel, err := d.channelAccessSvc.EnsureChannelAccess(ctx, message.ChannelID, input.ExecutorID)
	if err != nil {
		return err
	}

	// 既に削除済みの場合はエラー
	if message.DeletedAt != nil {
		return ErrMessageAlreadyDeleted
	}

	// 権限確認: 投稿者本人または管理者
	canDelete, err := d.canModifyMessage(ctx, channel.WorkspaceID, message.UserID, input.ExecutorID)
	if err != nil {
		return fmt.Errorf("権限確認に失敗しました: %w", err)
	}
	if !canDelete {
		return ErrUnauthorized
	}

	// 削除対象メッセージIDのリストを作成
	deleteIDs := []string{message.ID}

	// スレッド親メッセージの場合、子メッセージも削除
	if message.ParentID == nil {
		replies, err := d.messageRepo.FindThreadReplies(ctx, message.ID)
		if err != nil {
			return fmt.Errorf("返信の取得に失敗しました: %w", err)
		}
		for _, reply := range replies {
			deleteIDs = append(deleteIDs, reply.ID)
		}

		// スレッドメタデータも削除
		if err := d.threadRepo.DeleteMetadata(ctx, message.ID); err != nil {
			logger.Get().Warn("Failed to delete thread metadata", zap.Error(err))
		}
	}

	// ソフトデリート実行
	if err := d.messageRepo.SoftDeleteByIDs(ctx, deleteIDs, input.ExecutorID); err != nil {
		return fmt.Errorf("メッセージの削除に失敗しました: %w", err)
	}

	// WebSocket通知を送信
	if d.notificationSvc != nil {
		deleteData := map[string]interface{}{
			"messageId":  message.ID,
			"channelId":  message.ChannelID,
			"deletedIds": deleteIDs,
		}
		d.notificationSvc.NotifyDeletedMessage(channel.WorkspaceID, channel.ID, deleteData)
	}

	return nil
}

// ensureChannelAccess は ChannelAccessService に委譲済み

// canModifyMessage はユーザーがメッセージを編集・削除できるかどうかを確認します
func (d *MessageDeleter) canModifyMessage(ctx context.Context, workspaceID, messageOwnerID, executorID string) (bool, error) {
	// 投稿者本人の場合は許可
	if messageOwnerID == executorID {
		return true, nil
	}

	// 管理者権限チェック
	member, err := d.workspaceRepo.FindMember(ctx, workspaceID, executorID)
	if err != nil {
		return false, fmt.Errorf("ワークスペースメンバー情報の取得に失敗しました: %w", err)
	}
	if member == nil {
		return false, nil
	}

	// owner または admin の場合は許可
	if member.Role == entity.WorkspaceRoleOwner || member.Role == entity.WorkspaceRoleAdmin {
		return true, nil
	}

	return false, nil
}
