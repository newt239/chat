package registry

import (
	"github.com/labstack/echo/v4"
	"github.com/newt239/chat/internal/interfaces/handler/http"
	"github.com/newt239/chat/internal/interfaces/handler/http/handler"
	"github.com/newt239/chat/internal/interfaces/handler/websocket"
)

// InterfaceRegistry はインターフェース層の依存関係を管理します
type InterfaceRegistry struct {
	usecaseRegistry        *UseCaseRegistry
	infrastructureRegistry *InfrastructureRegistry
	domainRegistry         *DomainRegistry
}

// NewInterfaceRegistry は新しいInterfaceRegistryを作成します
func NewInterfaceRegistry(usecaseRegistry *UseCaseRegistry, infrastructureRegistry *InfrastructureRegistry, domainRegistry *DomainRegistry) *InterfaceRegistry {
	return &InterfaceRegistry{
		usecaseRegistry:        usecaseRegistry,
		infrastructureRegistry: infrastructureRegistry,
		domainRegistry:         domainRegistry,
	}
}

// Handlers
func (r *InterfaceRegistry) NewAuthHandler() *handler.AuthHandler {
	return handler.NewAuthHandler(r.usecaseRegistry.NewAuthUseCase())
}

func (r *InterfaceRegistry) NewWorkspaceHandler() *handler.WorkspaceHandler {
	return handler.NewWorkspaceHandler(r.usecaseRegistry.NewWorkspaceUseCase())
}

func (r *InterfaceRegistry) NewChannelHandler() *handler.ChannelHandler {
	return handler.NewChannelHandler(r.usecaseRegistry.NewChannelUseCase())
}

func (r *InterfaceRegistry) NewChannelMemberHandler() *handler.ChannelMemberHandler {
	return handler.NewChannelMemberHandler(r.usecaseRegistry.NewChannelMemberUseCase())
}

func (r *InterfaceRegistry) NewMessageHandler() *handler.MessageHandler {
	return handler.NewMessageHandler(r.usecaseRegistry.NewMessageUseCase())
}

func (r *InterfaceRegistry) NewReadStateHandler() *handler.ReadStateHandler {
	return handler.NewReadStateHandler(r.usecaseRegistry.NewReadStateUseCase())
}

func (r *InterfaceRegistry) NewReactionHandler() *handler.ReactionHandler {
	return handler.NewReactionHandler(r.usecaseRegistry.NewReactionUseCase())
}

func (r *InterfaceRegistry) NewUserGroupHandler() *handler.UserGroupHandler {
	return handler.NewUserGroupHandler(r.usecaseRegistry.NewUserGroupUseCase())
}

func (r *InterfaceRegistry) NewLinkHandler() *handler.LinkHandler {
	return handler.NewLinkHandler(r.usecaseRegistry.NewLinkUseCase())
}

func (r *InterfaceRegistry) NewBookmarkHandler() *handler.BookmarkHandler {
	return handler.NewBookmarkHandler(r.usecaseRegistry.NewBookmarkUseCase())
}

func (r *InterfaceRegistry) NewAttachmentHandler() *handler.AttachmentHandler {
	return handler.NewAttachmentHandler(r.usecaseRegistry.NewAttachmentUseCase())
}

// Router
func (r *InterfaceRegistry) NewRouter() *echo.Echo {
	routerConfig := http.RouterConfig{
		JWTService:           r.infrastructureRegistry.NewJWTService(),
		AllowedOrigins:       r.infrastructureRegistry.config.CORS.AllowedOrigins,
		WebSocketHub:         r.infrastructureRegistry.hub,
		WorkspaceRepository:  r.domainRegistry.NewWorkspaceRepository(),
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
func (r *InterfaceRegistry) NewWebSocketHub() *websocket.Hub {
	return r.infrastructureRegistry.hub
}
