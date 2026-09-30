package registry

import (
	adminuc "github.com/newt239/chat/internal/usecase/admin"
	attachmentuc "github.com/newt239/chat/internal/usecase/attachment"
	"github.com/newt239/chat/internal/usecase/audit"
	authuc "github.com/newt239/chat/internal/usecase/auth"
	bookmarkuc "github.com/newt239/chat/internal/usecase/bookmark"
	channeluc "github.com/newt239/chat/internal/usecase/channel"
	channelcategoryuc "github.com/newt239/chat/internal/usecase/channelcategory"
	channellinkuc "github.com/newt239/chat/internal/usecase/channellink"
	channelmemberuc "github.com/newt239/chat/internal/usecase/channelmember"
	customemojiuc "github.com/newt239/chat/internal/usecase/customemoji"
	dmuc "github.com/newt239/chat/internal/usecase/dm"
	draftuc "github.com/newt239/chat/internal/usecase/draft"
	insightuc "github.com/newt239/chat/internal/usecase/insight"
	invitationuc "github.com/newt239/chat/internal/usecase/invitation"
	linkuc "github.com/newt239/chat/internal/usecase/link"
	mentionuc "github.com/newt239/chat/internal/usecase/mention"
	messageuc "github.com/newt239/chat/internal/usecase/message"
	notificationuc "github.com/newt239/chat/internal/usecase/notification"
	pinuc "github.com/newt239/chat/internal/usecase/pin"
	reactionuc "github.com/newt239/chat/internal/usecase/reaction"
	readstateuc "github.com/newt239/chat/internal/usecase/readstate"
	scheduledmessageuc "github.com/newt239/chat/internal/usecase/scheduledmessage"
	searchuc "github.com/newt239/chat/internal/usecase/search"
	"github.com/newt239/chat/internal/usecase/searchindex"
	systemmsguc "github.com/newt239/chat/internal/usecase/systemmessage"
	threaduc "github.com/newt239/chat/internal/usecase/thread"
	useruc "github.com/newt239/chat/internal/usecase/user"
	usergroupuc "github.com/newt239/chat/internal/usecase/user_group"
	usernoteuc "github.com/newt239/chat/internal/usecase/usernote"
	webhookuc "github.com/newt239/chat/internal/usecase/webhook"
	workspaceuc "github.com/newt239/chat/internal/usecase/workspace"
)

// UseCaseRegistry はユースケース層の依存関係を管理します
type UseCaseRegistry struct {
	domainRegistry         *DomainRegistry
	infrastructureRegistry *InfrastructureRegistry
}

// NewUseCaseRegistry は新しいUseCaseRegistryを作成します
func NewUseCaseRegistry(domainRegistry *DomainRegistry, infrastructureRegistry *InfrastructureRegistry) *UseCaseRegistry {
	return &UseCaseRegistry{
		domainRegistry:         domainRegistry,
		infrastructureRegistry: infrastructureRegistry,
	}
}

// Use Cases
func (r *UseCaseRegistry) NewAuthUseCase() authuc.AuthUseCase {
	return authuc.NewAuthInteractor(
		r.domainRegistry.NewUserRepository(),
		r.domainRegistry.NewSessionRepository(),
		r.domainRegistry.NewWorkspaceRepository(),
		r.domainRegistry.NewInvitationRepository(),
		r.infrastructureRegistry.NewJWTService(),
		r.infrastructureRegistry.NewPasswordService(),
		r.infrastructureRegistry.NewGoogleVerifier(),
		r.infrastructureRegistry.NewGoogleOAuth(),
		r.infrastructureRegistry.NewTransactionManager(),
		r.NewAuditRecorder(),
		r.infrastructureRegistry.NewAuthSettings(),
	)
}

func (r *UseCaseRegistry) NewInvitationUseCase() *invitationuc.Interactor {
	return invitationuc.NewInteractor(
		r.domainRegistry.NewInvitationRepository(),
		r.domainRegistry.NewWorkspaceRepository(),
		r.domainRegistry.NewUserRepository(),
		r.domainRegistry.NewPermissionService(),
		r.infrastructureRegistry.NewInvitationSender(),
	)
}

func (r *UseCaseRegistry) NewAuditRecorder() audit.Recorder {
	return audit.NewRecorder(r.domainRegistry.NewAuditLogRepository(), r.infrastructureRegistry.NewLogger())
}

func (r *UseCaseRegistry) NewAdminUseCase() *adminuc.Interactor {
	return adminuc.NewInteractor(
		r.domainRegistry.NewWorkspaceRepository(),
		r.domainRegistry.NewUserRepository(),
		r.domainRegistry.NewSessionRepository(),
		r.domainRegistry.NewAuditLogRepository(),
		r.domainRegistry.NewPermissionRepository(),
		r.domainRegistry.NewInsightRepository(),
		r.domainRegistry.NewPermissionService(),
		r.NewAuditRecorder(),
		r.infrastructureRegistry.NewTransactionManager(),
	)
}

func (r *UseCaseRegistry) NewInsightUseCase() *insightuc.Interactor {
	return insightuc.NewInteractor(r.domainRegistry.NewWorkspaceRepository(), r.domainRegistry.NewInsightRepository())
}

func (r *UseCaseRegistry) NewWorkspaceUseCase() workspaceuc.WorkspaceUseCase {
	return workspaceuc.NewWorkspaceInteractor(
		r.domainRegistry.NewWorkspaceRepository(),
		r.domainRegistry.NewUserRepository(),
		r.domainRegistry.NewUserNoteRepository(),
		r.domainRegistry.NewPermissionService(),
		r.NewAuditRecorder(),
	)
}

func (r *UseCaseRegistry) NewChannelUseCase() channeluc.ChannelUseCase {
	return channeluc.NewChannelInteractor(
		r.domainRegistry.NewChannelRepository(),
		r.domainRegistry.NewChannelMemberRepository(),
		r.domainRegistry.NewChannelStarRepository(),
		r.domainRegistry.NewChannelMuteRepository(),
		r.domainRegistry.NewWorkspaceRepository(),
		r.domainRegistry.NewReadStateRepository(),
		r.infrastructureRegistry.NewTransactionManager(),
		r.NewSystemMessageUseCase(),
		r.domainRegistry.NewChannelAccessService(),
		r.domainRegistry.NewPermissionService(),
		r.NewAuditRecorder(),
	)
}

func (r *UseCaseRegistry) NewChannelMemberUseCase() channelmemberuc.ChannelMemberUseCase {
	return channelmemberuc.NewChannelMemberInteractor(
		r.domainRegistry.NewChannelRepository(),
		r.domainRegistry.NewChannelMemberRepository(),
		r.domainRegistry.NewWorkspaceRepository(),
		r.domainRegistry.NewUserRepository(),
		r.NewSystemMessageUseCase(),
	)
}

// NewMessageOutputBuilder はメッセージを返すユースケースで共有する出力の組み立て役です
func (r *UseCaseRegistry) NewMessageOutputBuilder() *messageuc.MessageOutputBuilder {
	return messageuc.NewMessageOutputBuilder(
		r.domainRegistry.NewMessageRepository(),
		r.domainRegistry.NewUserRepository(),
		r.domainRegistry.NewUserGroupRepository(),
		r.domainRegistry.NewMessageUserMentionRepository(),
		r.domainRegistry.NewMessageGroupMentionRepository(),
		r.domainRegistry.NewMessageLinkRepository(),
		r.domainRegistry.NewAttachmentRepository(),
		r.domainRegistry.NewPinRepository(),
		r.domainRegistry.NewChannelAccessService(),
	)
}

func (r *UseCaseRegistry) NewWebhookUseCase() *webhookuc.Interactor {
	return webhookuc.NewInteractor(
		r.domainRegistry.NewWebhookRepository(),
		r.domainRegistry.NewUserRepository(),
		r.domainRegistry.NewWorkspaceRepository(),
		r.domainRegistry.NewChannelAccessService(),
		messageuc.NewMessageCreator(
			r.domainRegistry.NewMessageRepository(),
			r.domainRegistry.NewMessageUserMentionRepository(),
			r.domainRegistry.NewMessageGroupMentionRepository(),
			r.domainRegistry.NewMessageLinkRepository(),
			r.domainRegistry.NewThreadRepository(),
			r.domainRegistry.NewAttachmentRepository(),
			r.infrastructureRegistry.NewNotificationService(),
			r.infrastructureRegistry.NewMentionService(),
			r.infrastructureRegistry.NewLinkProcessingService(),
			r.infrastructureRegistry.NewTransactionManager(),
			r.NewMessageOutputBuilder(),
			r.domainRegistry.NewChannelAccessService(),
			r.NewSearchIndexer(),
			r.NewPushDispatcher(),
		),
		r.infrastructureRegistry.NewTransactionManager(),
		r.NewAuditRecorder(),
	)
}

func (r *UseCaseRegistry) NewMessageUseCase() messageuc.MessageUseCase {
	return messageuc.NewMessageUseCase(
		r.domainRegistry.NewMessageRepository(),
		r.domainRegistry.NewSystemMessageRepository(),
		r.domainRegistry.NewChannelRepository(),
		r.domainRegistry.NewChannelMemberRepository(),
		r.domainRegistry.NewWorkspaceRepository(),
		r.domainRegistry.NewUserRepository(),
		r.domainRegistry.NewMessageUserMentionRepository(),
		r.domainRegistry.NewMessageGroupMentionRepository(),
		r.domainRegistry.NewMessageLinkRepository(),
		r.domainRegistry.NewThreadRepository(),
		r.domainRegistry.NewAttachmentRepository(),
		r.NewMessageOutputBuilder(),
		r.infrastructureRegistry.NewNotificationService(),
		r.infrastructureRegistry.NewMentionService(),
		r.infrastructureRegistry.NewLinkProcessingService(),
		r.infrastructureRegistry.NewTransactionManager(),
		r.domainRegistry.NewChannelAccessService(),
		r.domainRegistry.NewPermissionService(),
		r.infrastructureRegistry.NewLogger(),
		r.NewSearchIndexer(),
		r.NewPushDispatcher(),
	)
}

func (r *UseCaseRegistry) NewPushDispatcher() *notificationuc.Dispatcher {
	return notificationuc.NewDispatcher(
		r.domainRegistry.NewUserRepository(),
		r.domainRegistry.NewChannelMemberRepository(),
		r.domainRegistry.NewChannelMuteRepository(),
		r.domainRegistry.NewThreadRepository(),
		r.domainRegistry.NewUserGroupRepository(),
		r.domainRegistry.NewPushTokenRepository(),
		r.domainRegistry.NewChannelAccessService(),
		r.infrastructureRegistry.PushSender(),
		r.infrastructureRegistry.NewLogger(),
	)
}

func (r *UseCaseRegistry) NewNotificationUseCase() notificationuc.UseCase {
	return notificationuc.NewInteractor(r.domainRegistry.NewPushTokenRepository())
}

func (r *UseCaseRegistry) NewDraftUseCase() *draftuc.Interactor {
	return draftuc.NewInteractor(
		r.domainRegistry.NewDraftRepository(),
		r.domainRegistry.NewMessageRepository(),
		r.domainRegistry.NewChannelAccessService(),
	)
}

func (r *UseCaseRegistry) NewScheduledMessageUseCase() *scheduledmessageuc.Interactor {
	return scheduledmessageuc.NewInteractor(
		r.domainRegistry.NewScheduledMessageRepository(),
		r.domainRegistry.NewMessageRepository(),
		r.domainRegistry.NewAttachmentRepository(),
		r.domainRegistry.NewChannelAccessService(),
		r.NewMessageUseCase(),
		r.infrastructureRegistry.NewLogger(),
	)
}

func (r *UseCaseRegistry) NewSearchIndexer() *searchindex.Indexer {
	return searchindex.NewIndexer(
		r.domainRegistry.NewMessageRepository(),
		r.infrastructureRegistry.MessageSearchIndex(),
		r.infrastructureRegistry.NewLogger(),
	)
}

func (r *UseCaseRegistry) NewSystemMessageUseCase() systemmsguc.UseCase {
	return systemmsguc.New(
		r.domainRegistry.NewSystemMessageRepository(),
		r.domainRegistry.NewChannelRepository(),
		r.infrastructureRegistry.NewNotificationService(),
	)
}

func (r *UseCaseRegistry) NewReadStateUseCase() readstateuc.ReadStateUseCase {
	return readstateuc.NewReadStateInteractor(
		r.domainRegistry.NewReadStateRepository(),
		r.domainRegistry.NewChannelRepository(),
		r.domainRegistry.NewChannelMemberRepository(),
		r.domainRegistry.NewWorkspaceRepository(),
		r.infrastructureRegistry.NewNotificationService(),
		r.domainRegistry.NewChannelAccessService(),
	)
}

func (r *UseCaseRegistry) NewReactionUseCase() reactionuc.ReactionUseCase {
	return reactionuc.NewReactionInteractor(
		r.domainRegistry.NewMessageRepository(),
		r.domainRegistry.NewChannelRepository(),
		r.domainRegistry.NewChannelMemberRepository(),
		r.domainRegistry.NewWorkspaceRepository(),
		r.domainRegistry.NewUserRepository(),
		r.infrastructureRegistry.NewNotificationService(),
		r.domainRegistry.NewChannelAccessService(),
	)
}

func (r *UseCaseRegistry) NewUserGroupUseCase() usergroupuc.UserGroupUseCase {
	return usergroupuc.NewUserGroupInteractor(
		r.domainRegistry.NewUserGroupRepository(),
		r.domainRegistry.NewWorkspaceRepository(),
		r.domainRegistry.NewUserRepository(),
	)
}

func (r *UseCaseRegistry) NewLinkUseCase() linkuc.LinkUseCase {
	return linkuc.NewLinkInteractor(r.infrastructureRegistry.NewOGPService())
}

func (r *UseCaseRegistry) NewBookmarkUseCase() bookmarkuc.BookmarkUseCase {
	return bookmarkuc.NewBookmarkInteractor(
		r.domainRegistry.NewBookmarkRepository(),
		r.domainRegistry.NewMessageRepository(),
		r.NewMessageOutputBuilder(),
		r.domainRegistry.NewChannelAccessService(),
	)
}

func (r *UseCaseRegistry) NewPinUseCase() pinuc.PinUseCase {
	return pinuc.NewPinInteractor(
		r.domainRegistry.NewPinRepository(),
		r.domainRegistry.NewMessageRepository(),
		r.domainRegistry.NewChannelRepository(),
		r.domainRegistry.NewChannelMemberRepository(),
		r.domainRegistry.NewWorkspaceRepository(),
		r.domainRegistry.NewUserRepository(),
		r.infrastructureRegistry.NewNotificationService(),
		r.NewMessageOutputBuilder(),
		r.domainRegistry.NewChannelAccessService(),
		r.NewSystemMessageUseCase(),
		r.domainRegistry.NewPermissionService(),
		r.NewSearchIndexer(),
	)
}

func (r *UseCaseRegistry) NewAttachmentUseCase() *attachmentuc.Interactor {
	return attachmentuc.NewInteractor(
		r.domainRegistry.NewAttachmentRepository(),
		r.domainRegistry.NewMessageRepository(),
		r.domainRegistry.NewChannelAccessService(),
		r.infrastructureRegistry.NewStorageService(),
		r.infrastructureRegistry.NewStorageConfig(),
	)
}

func (r *UseCaseRegistry) NewSearchUseCase() searchuc.SearchUseCase {
	return searchuc.NewSearchUseCase(
		r.domainRegistry.NewWorkspaceRepository(),
		r.domainRegistry.NewChannelRepository(),
		r.domainRegistry.NewMessageRepository(),
		r.infrastructureRegistry.MessageSearchIndex(),
		r.domainRegistry.NewUserRepository(),
		r.domainRegistry.NewUserGroupRepository(),
		r.NewMessageOutputBuilder(),
	)
}

func (r *UseCaseRegistry) NewMentionLister() *mentionuc.Lister {
	return mentionuc.NewLister(
		r.domainRegistry.NewWorkspaceRepository(),
		r.domainRegistry.NewMessageRepository(),
		r.NewMessageOutputBuilder(),
	)
}

func (r *UseCaseRegistry) NewDMInteractor() *dmuc.Interactor {
	return dmuc.NewInteractor(
		r.domainRegistry.NewChannelRepository(),
		r.domainRegistry.NewChannelMemberRepository(),
		r.domainRegistry.NewChannelStarRepository(),
		r.domainRegistry.NewChannelMuteRepository(),
		r.domainRegistry.NewReadStateRepository(),
		r.domainRegistry.NewUserRepository(),
		r.domainRegistry.NewWorkspaceRepository(),
	)
}

func (r *UseCaseRegistry) NewThreadLister() *threaduc.ThreadLister {
	return threaduc.NewThreadLister(
		r.domainRegistry.NewThreadRepository(),
		r.NewMessageOutputBuilder(),
	)
}

func (r *UseCaseRegistry) NewThreadReader() *threaduc.ThreadReader {
	return threaduc.NewThreadReader(
		r.domainRegistry.NewThreadRepository(),
		r.domainRegistry.NewMessageRepository(),
		r.domainRegistry.NewChannelAccessService(),
	)
}

func (r *UseCaseRegistry) NewUserUseCase() useruc.UseCase {
	return useruc.NewInteractor(
		r.domainRegistry.NewUserRepository(),
		r.domainRegistry.NewSessionRepository(),
		r.infrastructureRegistry.NewPasswordService(),
	)
}

func (r *UseCaseRegistry) NewChannelLinkUseCase() channellinkuc.UseCase {
	return channellinkuc.NewInteractor(
		r.domainRegistry.NewChannelLinkRepository(),
		r.domainRegistry.NewChannelAccessService(),
		r.domainRegistry.NewPermissionService(),
		r.infrastructureRegistry.NewTransactionManager(),
	)
}

func (r *UseCaseRegistry) NewChannelCategoryUseCase() channelcategoryuc.UseCase {
	return channelcategoryuc.NewInteractor(
		r.domainRegistry.NewChannelCategoryRepository(),
		r.domainRegistry.NewWorkspaceRepository(),
		r.domainRegistry.NewChannelAccessService(),
		r.infrastructureRegistry.NewTransactionManager(),
	)
}

func (r *UseCaseRegistry) NewUserNoteUseCase() usernoteuc.UseCase {
	return usernoteuc.NewInteractor(
		r.domainRegistry.NewUserNoteRepository(),
		r.domainRegistry.NewUserRepository(),
	)
}

func (r *UseCaseRegistry) NewCustomEmojiUseCase() *customemojiuc.Interactor {
	return customemojiuc.NewInteractor(
		r.domainRegistry.NewCustomEmojiRepository(),
		r.domainRegistry.NewUserRepository(),
		r.domainRegistry.NewWorkspaceRepository(),
		r.domainRegistry.NewPermissionService(),
		r.infrastructureRegistry.NewStorageService(),
		r.infrastructureRegistry.NewStorageConfig(),
		r.infrastructureRegistry.NewNotificationService(),
		r.NewAuditRecorder(),
		r.infrastructureRegistry.NewLogger(),
	)
}
