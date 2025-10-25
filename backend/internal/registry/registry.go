package registry

import (
	"context"

	"gorm.io/gorm"

	"github.com/labstack/echo/v4"
	"github.com/newt239/chat/internal/adapter/controller/http"
	"github.com/newt239/chat/internal/adapter/controller/http/handler"
	"github.com/newt239/chat/internal/adapter/controller/websocket"
	"github.com/newt239/chat/internal/adapter/gateway/persistence"
	"github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
	domaintransaction "github.com/newt239/chat/internal/domain/transaction"
	"github.com/newt239/chat/internal/infrastructure/auth"
	"github.com/newt239/chat/internal/infrastructure/config"
	"github.com/newt239/chat/internal/infrastructure/link"
	"github.com/newt239/chat/internal/infrastructure/mention"
	"github.com/newt239/chat/internal/infrastructure/notification"
	"github.com/newt239/chat/internal/infrastructure/ogp"
	"github.com/newt239/chat/internal/infrastructure/storage/wasabi"
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

func (r *Registry) NewOGPService() service.OGPService {
	return ogp.NewOGPService()
}

func (r *Registry) NewStorageService() service.StorageService {
	client, err := wasabi.NewClient(context.Background(), r.NewWasabiConfig())
	if err != nil {
		// エラーハンドリング: ログ出力してnilを返す
		// 実際のアプリケーションでは適切なエラーハンドリングが必要
		return nil
	}
	return wasabi.NewPresignService(client)
}

func (r *Registry) NewStorageConfig() service.StorageConfig {
	return r.NewWasabiConfig()
}

func (r *Registry) NewWasabiConfig() *wasabi.Config {
	cfg := wasabi.NewConfig()
	cfg.Endpoint = r.config.Wasabi.Endpoint
	cfg.Region = r.config.Wasabi.Region
	cfg.AccessKeyID = r.config.Wasabi.AccessKeyID
	cfg.SecretAccessKey = r.config.Wasabi.SecretAccessKey
	cfg.BucketName = r.config.Wasabi.BucketName
	return cfg
}

func (r *Registry) NewMentionService() service.MentionService {
	return mention.NewMentionService(
		r.NewWorkspaceRepository(),
		r.NewUserRepository(),
		r.NewUserGroupRepository(),
		r.NewMessageUserMentionRepository(),
		r.NewMessageGroupMentionRepository(),
	)
}

func (r *Registry) NewLinkProcessingService() service.LinkProcessingService {
	return link.NewLinkProcessingService(
		r.NewOGPService(),
		r.NewMessageLinkRepository(),
	)
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

func (r *Registry) NewChannelMemberRepository() repository.ChannelMemberRepository {
	return persistence.NewChannelMemberRepository(r.db)
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
	return persistence.NewBookmarkRepository(r.db)
}

func (r *Registry) NewThreadRepository() repository.ThreadRepository {
	return persistence.NewThreadRepository(r.db)
}

func (r *Registry) NewAttachmentRepository() repository.AttachmentRepository {
	return persistence.NewAttachmentRepository(r.db)
}

func (r *Registry) NewTransactionManager() domaintransaction.Manager {
	return persistence.NewTransactionManager(r.db)
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
		r.NewChannelMemberRepository(),
		r.NewWorkspaceRepository(),
		r.NewTransactionManager(),
	)
}

func (r *Registry) NewChannelMemberUseCase() channelmemberuc.ChannelMemberUseCase {
	return channelmemberuc.NewChannelMemberInteractor(
		r.NewChannelRepository(),
		r.NewChannelMemberRepository(),
		r.NewWorkspaceRepository(),
		r.NewUserRepository(),
	)
}

func (r *Registry) NewMessageUseCase() messageuc.MessageUseCase {
	return messageuc.NewMessageInteractor(
		r.NewMessageRepository(),
		r.NewChannelRepository(),
		r.NewChannelMemberRepository(),
		r.NewWorkspaceRepository(),
		r.NewUserRepository(),
		r.NewUserGroupRepository(),
		r.NewMessageUserMentionRepository(),
		r.NewMessageGroupMentionRepository(),
		r.NewMessageLinkRepository(),
		r.NewThreadRepository(),
		r.NewAttachmentRepository(),
		r.NewOGPService(),
		r.NewNotificationService(),
		r.NewMentionService(),
		r.NewLinkProcessingService(),
	)
}

func (r *Registry) NewReadStateUseCase() readstateuc.ReadStateUseCase {
	return readstateuc.NewReadStateInteractor(
		r.NewReadStateRepository(),
		r.NewChannelRepository(),
		r.NewChannelMemberRepository(),
		r.NewWorkspaceRepository(),
		r.NewNotificationService(),
	)
}

func (r *Registry) NewReactionUseCase() reactionuc.ReactionUseCase {
	return reactionuc.NewReactionInteractor(
		r.NewMessageRepository(),
		r.NewChannelRepository(),
		r.NewChannelMemberRepository(),
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
	return linkuc.NewLinkInteractor(r.NewOGPService())
}

func (r *Registry) NewBookmarkUseCase() bookmarkuc.BookmarkUseCase {
	return bookmarkuc.NewBookmarkInteractor(
		r.NewBookmarkRepository(),
		r.NewMessageRepository(),
		r.NewChannelRepository(),
		r.NewChannelMemberRepository(),
		r.NewWorkspaceRepository(),
	)
}

func (r *Registry) NewAttachmentUseCase() *attachmentuc.Interactor {
	return attachmentuc.NewInteractor(
		r.NewAttachmentRepository(),
		r.NewChannelRepository(),
		r.NewChannelMemberRepository(),
		r.NewMessageRepository(),
		r.NewStorageService(),
		r.NewStorageConfig(),
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

func (r *Registry) NewChannelMemberHandler() *handler.ChannelMemberHandler {
	return handler.NewChannelMemberHandler(r.NewChannelMemberUseCase())
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

func (r *Registry) NewBookmarkHandler() *handler.BookmarkHandler {
	return handler.NewBookmarkHandler(r.NewBookmarkUseCase())
}

func (r *Registry) NewAttachmentHandler() *handler.AttachmentHandler {
	return handler.NewAttachmentHandler(r.NewAttachmentUseCase())
}

// Router
func (r *Registry) NewRouter() *echo.Echo {
	routerConfig := http.RouterConfig{
		JWTService:           r.NewJWTService(),
		AllowedOrigins:       r.config.CORS.AllowedOrigins,
		WebSocketHub:         r.hub,
		WorkspaceRepository:  r.NewWorkspaceRepository(),
		AuthHandler:          r.NewAuthHandler(),
		WorkspaceHandler:     r.NewWorkspaceHandler(),
		ChannelHandler:       r.NewChannelHandler(),
		ChannelMemberHandler: r.NewChannelMemberHandler(),
		MessageHandler:       r.NewMessageHandler(),
		ReadStateHandler:     r.NewReadStateHandler(),
		ReactionHandler:      r.NewReactionHandler(),
		UserGroupHandler:     r.NewUserGroupHandler(),
		LinkHandler:          r.NewLinkHandler(),
		BookmarkHandler:      r.NewBookmarkHandler(),
		AttachmentHandler:    r.NewAttachmentHandler(),
	}

	return http.NewRouter(routerConfig)
}

// WebSocket Hub
func (r *Registry) NewWebSocketHub() *websocket.Hub {
	return r.hub
}

// WebSocket Handler
func (r *Registry) NewWebSocketHandler() echo.HandlerFunc {
	return websocket.NewHandler(
		r.hub,
		r.NewJWTService(),
		r.NewWorkspaceRepository(),
		r.NewMessageUseCase(),
		r.NewReadStateUseCase(),
	)
}
