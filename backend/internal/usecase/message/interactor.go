package message

import (
	"context"

	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/domain/transaction"
)

// MessageUseCase はメッセージ関連のユースケースインターフェースです
type MessageUseCase interface {
	ListMessages(ctx context.Context, input ListMessagesInput) (*ListMessagesOutput, error)
	CreateMessage(ctx context.Context, input CreateMessageInput) (*MessageOutput, error)
	UpdateMessage(ctx context.Context, input UpdateMessageInput) (*MessageOutput, error)
	DeleteMessage(ctx context.Context, input DeleteMessageInput) error
	GetThreadReplies(ctx context.Context, input GetThreadRepliesInput) (*GetThreadRepliesOutput, error)
	GetThreadMetadata(ctx context.Context, input GetThreadMetadataInput) (*ThreadMetadataOutput, error)
	ListMessagesWithThread(ctx context.Context, input ListMessagesInput) ([]MessageWithThreadOutput, error)
}

// messageInteractor は分割されたユースケースを統合するインタラクターです
type messageInteractor struct {
	creator *MessageCreator
	updater *MessageUpdater
	deleter *MessageDeleter
	lister  *MessageLister
}

// NewMessageUseCase は新しいメッセージユースケースを作成します
func NewMessageUseCase(
	messageRepo domainrepository.MessageRepository,
	channelRepo domainrepository.ChannelRepository,
	channelMemberRepo domainrepository.ChannelMemberRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	userRepo domainrepository.UserRepository,
	userGroupRepo domainrepository.UserGroupRepository,
	userMentionRepo domainrepository.MessageUserMentionRepository,
	groupMentionRepo domainrepository.MessageGroupMentionRepository,
	linkRepo domainrepository.MessageLinkRepository,
	threadRepo domainrepository.ThreadRepository,
	attachmentRepo domainrepository.AttachmentRepository,
	ogpService service.OGPService,
	notificationSvc service.NotificationService,
	mentionService service.MentionService,
	linkProcessingService service.LinkProcessingService,
	transactionManager transaction.Manager,
) MessageUseCase {
	// 各機能のユースケースを作成
	creator := NewMessageCreator(
		messageRepo,
		channelRepo,
		channelMemberRepo,
		workspaceRepo,
		userRepo,
		userGroupRepo,
		userMentionRepo,
		groupMentionRepo,
		linkRepo,
		threadRepo,
		attachmentRepo,
		ogpService,
		notificationSvc,
		mentionService,
		linkProcessingService,
		transactionManager,
	)

	updater := NewMessageUpdater(
		messageRepo,
		channelRepo,
		channelMemberRepo,
		workspaceRepo,
		userRepo,
		userGroupRepo,
		userMentionRepo,
		groupMentionRepo,
		linkRepo,
		attachmentRepo,
		notificationSvc,
		mentionService,
		linkProcessingService,
		transactionManager,
	)

	deleter := NewMessageDeleter(
		messageRepo,
		channelRepo,
		channelMemberRepo,
		workspaceRepo,
		threadRepo,
		notificationSvc,
	)

	lister := NewMessageLister(
		messageRepo,
		channelRepo,
		channelMemberRepo,
		workspaceRepo,
		userRepo,
		userGroupRepo,
		userMentionRepo,
		groupMentionRepo,
		linkRepo,
		threadRepo,
		attachmentRepo,
	)

	return &messageInteractor{
		creator: creator,
		updater: updater,
		deleter: deleter,
		lister:  lister,
	}
}

// ListMessages はメッセージ一覧を取得します
func (i *messageInteractor) ListMessages(ctx context.Context, input ListMessagesInput) (*ListMessagesOutput, error) {
	return i.lister.ListMessages(ctx, input)
}

// CreateMessage はメッセージを作成します
func (i *messageInteractor) CreateMessage(ctx context.Context, input CreateMessageInput) (*MessageOutput, error) {
	return i.creator.CreateMessage(ctx, input)
}

// UpdateMessage はメッセージを更新します
func (i *messageInteractor) UpdateMessage(ctx context.Context, input UpdateMessageInput) (*MessageOutput, error) {
	return i.updater.UpdateMessage(ctx, input)
}

// DeleteMessage はメッセージを削除します
func (i *messageInteractor) DeleteMessage(ctx context.Context, input DeleteMessageInput) error {
	return i.deleter.DeleteMessage(ctx, input)
}

// GetThreadReplies はスレッド返信を取得します
func (i *messageInteractor) GetThreadReplies(ctx context.Context, input GetThreadRepliesInput) (*GetThreadRepliesOutput, error) {
	return i.lister.GetThreadReplies(ctx, input)
}

// GetThreadMetadata はスレッドメタデータを取得します
func (i *messageInteractor) GetThreadMetadata(ctx context.Context, input GetThreadMetadataInput) (*ThreadMetadataOutput, error) {
	return i.lister.GetThreadMetadata(ctx, input)
}

// ListMessagesWithThread はスレッド情報付きのメッセージ一覧を取得します
func (i *messageInteractor) ListMessagesWithThread(ctx context.Context, input ListMessagesInput) ([]MessageWithThreadOutput, error) {
	return i.lister.ListMessagesWithThread(ctx, input)
}
