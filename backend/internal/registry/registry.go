package registry

import (
	"context"

	"gorm.io/gorm"

	"github.com/example/chat/internal/adapter/controller/http"
	"github.com/example/chat/internal/adapter/controller/http/handler"
	"github.com/example/chat/internal/adapter/controller/websocket"
	"github.com/example/chat/internal/adapter/gateway/persistence"
	"github.com/example/chat/internal/domain/repository"
	"github.com/example/chat/internal/domain/service"
	"github.com/example/chat/internal/infrastructure/auth"
	"github.com/example/chat/internal/infrastructure/config"
	"github.com/example/chat/internal/infrastructure/notification"
	infrarepository "github.com/example/chat/internal/infrastructure/repository"
	"github.com/example/chat/internal/infrastructure/storage/wasabi"
	interfacehandler "github.com/example/chat/internal/interface/http/handler"
	attachmentuc "github.com/example/chat/internal/usecase/attachment"
	authuc "github.com/example/chat/internal/usecase/auth"
	bookmarkuc "github.com/example/chat/internal/usecase/bookmark"
	channeluc "github.com/example/chat/internal/usecase/channel"
	channelmemberuc "github.com/example/chat/internal/usecase/channelmember"
	linkuc "github.com/example/chat/internal/usecase/link"
	messageuc "github.com/example/chat/internal/usecase/message"
	reactionuc "github.com/example/chat/internal/usecase/reaction"
	readstateuc "github.com/example/chat/internal/usecase/readstate"
	usergroupuc "github.com/example/chat/internal/usecase/user_group"
	workspaceuc "github.com/example/chat/internal/usecase/workspace"
	"github.com/labstack/echo/v4"
)

type Registry struct {
	db     *gorm.DB
	config *config.Config
	hub    *websocket.Hub
}

func NewRegistry(db *gorm.DB, cfg *config.Config) *Registry {
	return &Registry{
		db:     db,
		config: cfg,
		hub:    websocket.NewHub(),
	}
}

// Infrastructure Services
func (r *Registry) NewJWTService() authuc.JWTService {
	return auth.NewJWTService(r.config.JWT.Secret)
}

func (r *Registry) NewPasswordService() authuc.PasswordService {
	return auth.NewPasswordService()
}

func (r *Registry) NewNotificationService() service.NotificationService {
	return notification.NewWebSocketNotificationService(r.hub)
}

// Repositories
func (r *Registry) NewUserRepository() repository.UserRepository {
	return persistence.NewUserRepository(r.db)
}

func (r *Registry) NewSessionRepository() repository.SessionRepository {
	return persistence.NewSessionRepository(r.db)
}

func (r *Registry) NewWorkspaceRepository() repository.WorkspaceRepository {
	return persistence.NewWorkspaceRepository(r.db)
}

func (r *Registry) NewChannelRepository() repository.ChannelRepository {
	return persistence.NewChannelRepository(r.db)
}

func (r *Registry) NewMessageRepository() repository.MessageRepository {
	return persistence.NewMessageRepository(r.db)
}

func (r *Registry) NewReadStateRepository() repository.ReadStateRepository {
	return persistence.NewReadStateRepository(r.db)
}

func (r *Registry) NewUserGroupRepository() repository.UserGroupRepository {
	return persistence.NewUserGroupRepository(r.db)
}

func (r *Registry) NewMessageUserMentionRepository() repository.MessageUserMentionRepository {
	return persistence.NewMessageUserMentionRepository(r.db)
}

func (r *Registry) NewMessageGroupMentionRepository() repository.MessageGroupMentionRepository {
	return persistence.NewMessageGroupMentionRepository(r.db)
}

func (r *Registry) NewMessageLinkRepository() repository.MessageLinkRepository {
	return persistence.NewMessageLinkRepository(r.db)
}

func (r *Registry) NewBookmarkRepository() repository.BookmarkRepository {
	return infrarepository.NewBookmarkRepository(r.db)
}

func (r *Registry) NewThreadRepository() repository.ThreadRepository {
	return persistence.NewThreadRepository(r.db)
}

func (r *Registry) NewAttachmentRepository() repository.AttachmentRepository {
	return persistence.NewAttachmentRepository(r.db)
}

// Use Cases
func (r *Registry) NewAuthUseCase() authuc.AuthUseCase {
	return authuc.NewAuthInteractor(
		r.NewUserRepository(),
		r.NewSessionRepository(),
		r.NewJWTService(),
		r.NewPasswordService(),
	)
}

func (r *Registry) NewWorkspaceUseCase() workspaceuc.WorkspaceUseCase {
	return workspaceuc.NewWorkspaceInteractor(
		r.NewWorkspaceRepository(),
		r.NewUserRepository(),
	)
}

func (r *Registry) NewChannelUseCase() channeluc.ChannelUseCase {
	return channeluc.NewChannelInteractor(
		r.NewChannelRepository(),
		r.NewWorkspaceRepository(),
	)
}

func (r *Registry) NewMessageUseCase() messageuc.MessageUseCase {
	return messageuc.NewMessageInteractor(
		r.NewMessageRepository(),
		r.NewChannelRepository(),
		r.NewWorkspaceRepository(),
		r.NewUserRepository(),
		r.NewUserGroupRepository(),
		r.NewMessageUserMentionRepository(),
		r.NewMessageGroupMentionRepository(),
		r.NewMessageLinkRepository(),
		r.NewThreadRepository(),
		r.NewAttachmentRepository(),
		r.NewNotificationService(),
	)
}

func (r *Registry) NewReadStateUseCase() readstateuc.ReadStateUseCase {
	return readstateuc.NewReadStateInteractor(
		r.NewReadStateRepository(),
		r.NewChannelRepository(),
		r.NewWorkspaceRepository(),
		r.NewNotificationService(),
	)
}

func (r *Registry) NewReactionUseCase() reactionuc.ReactionUseCase {
	return reactionuc.NewReactionInteractor(
		r.NewMessageRepository(),
		r.NewChannelRepository(),
		r.NewWorkspaceRepository(),
		r.NewUserRepository(),
		r.NewNotificationService(),
	)
}

func (r *Registry) NewUserGroupUseCase() usergroupuc.UserGroupUseCase {
	return usergroupuc.NewUserGroupInteractor(
		r.NewUserGroupRepository(),
		r.NewWorkspaceRepository(),
		r.NewUserRepository(),
	)
}

func (r *Registry) NewLinkUseCase() linkuc.LinkUseCase {
	return linkuc.NewLinkInteractor()
}

func (r *Registry) NewBookmarkUseCase() bookmarkuc.BookmarkUseCase {
	return bookmarkuc.NewBookmarkInteractor(
		r.NewBookmarkRepository(),
		r.NewMessageRepository(),
		r.NewChannelRepository(),
		r.NewWorkspaceRepository(),
	)
}

func (r *Registry) NewAttachmentUseCase() *attachmentuc.Interactor {
	wasabiCfg := wasabi.NewConfig()
	// TODO: 環境変数から設定を読み込む
	wasabiCfg.Endpoint = r.config.Wasabi.Endpoint
	wasabiCfg.Region = r.config.Wasabi.Region
	wasabiCfg.AccessKeyID = r.config.Wasabi.AccessKeyID
	wasabiCfg.SecretAccessKey = r.config.Wasabi.SecretAccessKey
	wasabiCfg.BucketName = r.config.Wasabi.BucketName

	wasabiClient, err := wasabi.NewClient(context.Background(), wasabiCfg)
	if err != nil {
		panic(err) // TODO: エラーハンドリングを改善
	}

	presignService := wasabi.NewPresignService(wasabiClient)

	return attachmentuc.NewInteractor(
		r.NewAttachmentRepository(),
		r.NewChannelRepository(),
		r.NewMessageRepository(),
		presignService,
		wasabiCfg,
	)
}

// Handlers
func (r *Registry) NewAuthHandler() *handler.AuthHandler {
	return handler.NewAuthHandler(r.NewAuthUseCase())
}

func (r *Registry) NewWorkspaceHandler() *handler.WorkspaceHandler {
	return handler.NewWorkspaceHandler(r.NewWorkspaceUseCase())
}

func (r *Registry) NewChannelHandler() *handler.ChannelHandler {
	return handler.NewChannelHandler(r.NewChannelUseCase())
}

func (r *Registry) NewMessageHandler() *handler.MessageHandler {
	return handler.NewMessageHandler(r.NewMessageUseCase())
}

func (r *Registry) NewReadStateHandler() *handler.ReadStateHandler {
	return handler.NewReadStateHandler(r.NewReadStateUseCase())
}

func (r *Registry) NewReactionHandler() *handler.ReactionHandler {
	return handler.NewReactionHandler(r.NewReactionUseCase())
}

func (r *Registry) NewUserGroupHandler() *handler.UserGroupHandler {
	return handler.NewUserGroupHandler(r.NewUserGroupUseCase())
}

func (r *Registry) NewLinkHandler() *handler.LinkHandler {
	return handler.NewLinkHandler(r.NewLinkUseCase())
}

func (r *Registry) NewBookmarkHandler() *interfacehandler.BookmarkHandler {
	return interfacehandler.NewBookmarkHandler(r.NewBookmarkUseCase())
}

func (r *Registry) NewAttachmentHandler() *interfacehandler.AttachmentHandler {
	return interfacehandler.NewAttachmentHandler(r.NewAttachmentUseCase())
}

// Router
func (r *Registry) NewRouter() *echo.Echo {
	routerConfig := http.RouterConfig{
		JWTService:        r.NewJWTService(),
		AllowedOrigins:    r.config.CORS.AllowedOrigins,
		AuthHandler:       r.NewAuthHandler(),
		WorkspaceHandler:  r.NewWorkspaceHandler(),
		ChannelHandler:    r.NewChannelHandler(),
		MessageHandler:    r.NewMessageHandler(),
		ReadStateHandler:  r.NewReadStateHandler(),
		ReactionHandler:   r.NewReactionHandler(),
		UserGroupHandler:  r.NewUserGroupHandler(),
		LinkHandler:       r.NewLinkHandler(),
		BookmarkHandler:   r.NewBookmarkHandler(),
		AttachmentHandler: r.NewAttachmentHandler(),
	}

	return http.NewRouter(routerConfig)
}

// WebSocket Hub
func (r *Registry) NewWebSocketHub() *websocket.Hub {
	return r.hub
}
