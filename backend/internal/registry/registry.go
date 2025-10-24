package registry

import (
	"gorm.io/gorm"

	"github.com/example/chat/internal/adapter/controller/http"
	"github.com/example/chat/internal/adapter/controller/http/handler"
	"github.com/example/chat/internal/adapter/controller/websocket"
	"github.com/example/chat/internal/adapter/gateway/persistence"
	"github.com/example/chat/internal/domain/repository"
	"github.com/example/chat/internal/infrastructure/auth"
	"github.com/example/chat/internal/infrastructure/config"
	authuc "github.com/example/chat/internal/usecase/auth"
	channeluc "github.com/example/chat/internal/usecase/channel"
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
}

func NewRegistry(db *gorm.DB, cfg *config.Config) *Registry {
	return &Registry{
		db:     db,
		config: cfg,
	}
}

// Infrastructure Services
func (r *Registry) NewJWTService() authuc.JWTService {
	return auth.NewJWTService(r.config.JWT.Secret)
}

func (r *Registry) NewPasswordService() authuc.PasswordService {
	return auth.NewPasswordService()
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
	)
}

func (r *Registry) NewReadStateUseCase() readstateuc.ReadStateUseCase {
	return readstateuc.NewReadStateInteractor(
		r.NewReadStateRepository(),
		r.NewChannelRepository(),
		r.NewWorkspaceRepository(),
	)
}

func (r *Registry) NewReactionUseCase() reactionuc.ReactionUseCase {
	return reactionuc.NewReactionInteractor(
		r.NewMessageRepository(),
		r.NewChannelRepository(),
		r.NewWorkspaceRepository(),
		r.NewUserRepository(),
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

// Router
func (r *Registry) NewRouter() *echo.Echo {
	routerConfig := http.RouterConfig{
		JWTService:       r.NewJWTService(),
		AllowedOrigins:   r.config.CORS.AllowedOrigins,
		AuthHandler:      r.NewAuthHandler(),
		WorkspaceHandler: r.NewWorkspaceHandler(),
		ChannelHandler:   r.NewChannelHandler(),
		MessageHandler:   r.NewMessageHandler(),
		ReadStateHandler: r.NewReadStateHandler(),
		ReactionHandler:  r.NewReactionHandler(),
		UserGroupHandler: r.NewUserGroupHandler(),
		LinkHandler:      r.NewLinkHandler(),
	}

	return http.NewRouter(routerConfig)
}

// WebSocket Hub
func (r *Registry) NewWebSocketHub() *websocket.Hub {
	return websocket.NewHub()
}
