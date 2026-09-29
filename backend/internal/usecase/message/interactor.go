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
	GetMessagePreview(ctx context.Context, input GetMessagePreviewInput) (*MessagePreviewOutput, error)
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
	systemMsgRepo domainrepository.SystemMessageRepository,
	channelRepo domainrepository.ChannelRepository,
	channelMemberRepo domainrepository.ChannelMemberRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	userRepo domainrepository.UserRepository,
	userMentionRepo domainrepository.MessageUserMentionRepository,
	groupMentionRepo domainrepository.MessageGroupMentionRepository,
	linkRepo domainrepository.MessageLinkRepository,
	threadRepo domainrepository.ThreadRepository,
	attachmentRepo domainrepository.AttachmentRepository,
	outputBuilder *MessageOutputBuilder,
	notificationSvc Notifier,
	mentionService service.MentionService,
	linkProcessingService service.LinkProcessingService,
	transactionManager transaction.Manager,
	channelAccessSvc service.ChannelAccessService,
	permissionSvc service.PermissionService,
	logger service.Logger,
	searchIndexer SearchIndexer,
) MessageUseCase {
	return &messageInteractor{
		creator: NewMessageCreator(
			messageRepo,
			userMentionRepo,
			groupMentionRepo,
			linkRepo,
			threadRepo,
			attachmentRepo,
			notificationSvc,
			mentionService,
			linkProcessingService,
			transactionManager,
			outputBuilder,
			channelAccessSvc,
			searchIndexer,
		),
		updater: NewMessageUpdater(
			messageRepo,
			workspaceRepo,
			userMentionRepo,
			groupMentionRepo,
			linkRepo,
			notificationSvc,
			mentionService,
			linkProcessingService,
			transactionManager,
			outputBuilder,
			channelAccessSvc,
			searchIndexer,
		),
		deleter: NewMessageDeleter(
			messageRepo,
			channelRepo,
			channelMemberRepo,
			threadRepo,
			notificationSvc,
			channelAccessSvc,
			permissionSvc,
			logger,
			searchIndexer,
		),
		lister: NewMessageLister(
			messageRepo,
			systemMsgRepo,
			userRepo,
			threadRepo,
			outputBuilder,
			channelAccessSvc,
		),
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

// GetMessagePreview はメッセージリンクの引用カードを取得します
func (i *messageInteractor) GetMessagePreview(ctx context.Context, input GetMessagePreviewInput) (*MessagePreviewOutput, error) {
	return i.lister.GetMessagePreview(ctx, input)
}
