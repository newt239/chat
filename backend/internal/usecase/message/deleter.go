package message

import (
	"context"
	"fmt"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
)

// MessageDeleter はメッセージ削除を担当するユースケースです
type MessageDeleter struct {
	messageRepo       domainrepository.MessageRepository
	channelRepo       domainrepository.ChannelRepository
	channelMemberRepo domainrepository.ChannelMemberRepository
	threadRepo        domainrepository.ThreadRepository
	notificationSvc   Notifier
	channelAccessSvc  service.ChannelAccessService
	permissionSvc     service.PermissionService
	logger            service.Logger
	searchIndexer     SearchIndexer
}

// NewMessageDeleter は新しいMessageDeleterを作成します
func NewMessageDeleter(
	messageRepo domainrepository.MessageRepository,
	channelRepo domainrepository.ChannelRepository,
	channelMemberRepo domainrepository.ChannelMemberRepository,
	threadRepo domainrepository.ThreadRepository,
	notificationSvc Notifier,
	channelAccessSvc service.ChannelAccessService,
	permissionSvc service.PermissionService,
	logger service.Logger,
	searchIndexer SearchIndexer,
) *MessageDeleter {
	return &MessageDeleter{
		messageRepo:       messageRepo,
		channelRepo:       channelRepo,
		channelMemberRepo: channelMemberRepo,
		threadRepo:        threadRepo,
		notificationSvc:   notificationSvc,
		channelAccessSvc:  channelAccessSvc,
		permissionSvc:     permissionSvc,
		logger:            logger,
		searchIndexer:     searchIndexer,
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

	// 他人のメッセージは権限設定で許可されたロールだけが削除できる
	if message.UserID != input.ExecutorID {
		if _, err := d.permissionSvc.Ensure(ctx, channel.WorkspaceID, input.ExecutorID, entity.PermissionDeleteOthersMessages); err != nil {
			return err
		}
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
	}

	// ソフトデリート実行
	if err := d.messageRepo.SoftDeleteByIDs(ctx, deleteIDs, input.ExecutorID); err != nil {
		return fmt.Errorf("メッセージの削除に失敗しました: %w", err)
	}

	// 返信が消えると親メッセージの「スレッドあり」も変わる
	indexIDs := deleteIDs
	if message.ParentID != nil {
		indexIDs = append(indexIDs, *message.ParentID)
	}
	d.searchIndexer.Sync(ctx, indexIDs...)

	// WebSocket通知を送信
	if d.notificationSvc != nil {
		d.notificationSvc.NotifyDeletedMessage(channel.WorkspaceID, channel.ID, MessageDeletion{
			MessageID:  message.ID,
			DeletedIDs: deleteIDs,
			DeletedAt:  time.Now(),
		})
	}

	return nil
}
