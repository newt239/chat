package registry

import (
	attachmentuc "github.com/newt239/chat/internal/usecase/attachment"
	authuc "github.com/newt239/chat/internal/usecase/auth"
	bookmarkuc "github.com/newt239/chat/internal/usecase/bookmark"
	channeluc "github.com/newt239/chat/internal/usecase/channel"
	channelmemberuc "github.com/newt239/chat/internal/usecase/channelmember"
	linkuc "github.com/newt239/chat/internal/usecase/link"
	messageuc "github.com/newt239/chat/internal/usecase/message"
	reactionuc "github.com/newt239/chat/internal/usecase/reaction"
	readstateuc "github.com/newt239/chat/internal/usecase/readstate"
	usergroupuc "github.com/newt239/chat/internal/usecase/user_group"
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
		r.infrastructureRegistry.NewJWTService(),
		r.infrastructureRegistry.NewPasswordService(),
	)
}

func (r *UseCaseRegistry) NewWorkspaceUseCase() workspaceuc.WorkspaceUseCase {
	return workspaceuc.NewWorkspaceInteractor(
		r.domainRegistry.NewWorkspaceRepository(),
		r.domainRegistry.NewUserRepository(),
	)
}

func (r *UseCaseRegistry) NewChannelUseCase() channeluc.ChannelUseCase {
	return channeluc.NewChannelInteractor(
		r.domainRegistry.NewChannelRepository(),
		r.domainRegistry.NewChannelMemberRepository(),
		r.domainRegistry.NewWorkspaceRepository(),
		r.domainRegistry.NewReadStateRepository(),
		r.infrastructureRegistry.NewTransactionManager(),
	)
}

func (r *UseCaseRegistry) NewChannelMemberUseCase() channelmemberuc.ChannelMemberUseCase {
	return channelmemberuc.NewChannelMemberInteractor(
		r.domainRegistry.NewChannelRepository(),
		r.domainRegistry.NewChannelMemberRepository(),
		r.domainRegistry.NewWorkspaceRepository(),
		r.domainRegistry.NewUserRepository(),
	)
}

func (r *UseCaseRegistry) NewMessageUseCase() messageuc.MessageUseCase {
	return messageuc.NewMessageUseCase(
		r.domainRegistry.NewMessageRepository(),
		r.domainRegistry.NewChannelRepository(),
		r.domainRegistry.NewChannelMemberRepository(),
		r.domainRegistry.NewWorkspaceRepository(),
		r.domainRegistry.NewUserRepository(),
		r.domainRegistry.NewUserGroupRepository(),
		r.domainRegistry.NewMessageUserMentionRepository(),
		r.domainRegistry.NewMessageGroupMentionRepository(),
		r.domainRegistry.NewMessageLinkRepository(),
		r.domainRegistry.NewThreadRepository(),
		r.domainRegistry.NewAttachmentRepository(),
		r.infrastructureRegistry.NewOGPService(),
		r.infrastructureRegistry.NewNotificationService(),
		r.infrastructureRegistry.NewMentionService(),
		r.infrastructureRegistry.NewLinkProcessingService(),
		r.infrastructureRegistry.NewTransactionManager(),
	)
}

func (r *UseCaseRegistry) NewReadStateUseCase() readstateuc.ReadStateUseCase {
	return readstateuc.NewReadStateInteractor(
		r.domainRegistry.NewReadStateRepository(),
		r.domainRegistry.NewChannelRepository(),
		r.domainRegistry.NewChannelMemberRepository(),
		r.domainRegistry.NewWorkspaceRepository(),
		r.infrastructureRegistry.NewNotificationService(),
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
		r.domainRegistry.NewChannelRepository(),
		r.domainRegistry.NewChannelMemberRepository(),
		r.domainRegistry.NewWorkspaceRepository(),
	)
}

func (r *UseCaseRegistry) NewAttachmentUseCase() *attachmentuc.Interactor {
	return attachmentuc.NewInteractor(
		r.domainRegistry.NewAttachmentRepository(),
		r.domainRegistry.NewChannelRepository(),
		r.domainRegistry.NewChannelMemberRepository(),
		r.domainRegistry.NewMessageRepository(),
		r.infrastructureRegistry.NewStorageService(),
		r.infrastructureRegistry.NewStorageConfig(),
	)
}
