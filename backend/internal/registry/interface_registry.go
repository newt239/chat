package registry

import (
	"github.com/labstack/echo/v4"
	"github.com/newt239/chat/internal/interfaces/handler/http"
	"github.com/newt239/chat/internal/interfaces/handler/http/handler"
	"github.com/newt239/chat/internal/interfaces/handler/websocket"
)

type InterfaceRegistry struct {
	usecaseRegistry        *UseCaseRegistry
	infrastructureRegistry *InfrastructureRegistry
	domainRegistry         *DomainRegistry
}

func NewInterfaceRegistry(usecaseRegistry *UseCaseRegistry, infrastructureRegistry *InfrastructureRegistry, domainRegistry *DomainRegistry) *InterfaceRegistry {
	return &InterfaceRegistry{
		usecaseRegistry:        usecaseRegistry,
		infrastructureRegistry: infrastructureRegistry,
		domainRegistry:         domainRegistry,
	}
}

func (r *InterfaceRegistry) NewAuthHandler() *handler.AuthHandler {
	return &handler.AuthHandler{
		AuthUC: r.usecaseRegistry.NewAuthUseCase(),
	}
}

func (r *InterfaceRegistry) NewWorkspaceHandler() *handler.WorkspaceHandler {
	return &handler.WorkspaceHandler{
		WorkspaceUC: r.usecaseRegistry.NewWorkspaceUseCase(),
	}
}

func (r *InterfaceRegistry) NewChannelHandler() *handler.ChannelHandler {
	return &handler.ChannelHandler{
		ChannelUC: r.usecaseRegistry.NewChannelUseCase(),
	}
}

func (r *InterfaceRegistry) NewChannelMemberHandler() *handler.ChannelMemberHandler {
	return &handler.ChannelMemberHandler{
		ChannelMemberUseCase: r.usecaseRegistry.NewChannelMemberUseCase(),
		SystemMessageUC:      r.usecaseRegistry.NewSystemMessageUseCase(),
	}
}

func (r *InterfaceRegistry) NewMessageHandler() *handler.MessageHandler {
	return &handler.MessageHandler{
		MessageUC: r.usecaseRegistry.NewMessageUseCase(),
	}
}

func (r *InterfaceRegistry) NewReadStateHandler() *handler.ReadStateHandler {
	return &handler.ReadStateHandler{
		ReadStateUC: r.usecaseRegistry.NewReadStateUseCase(),
	}
}

func (r *InterfaceRegistry) NewReactionHandler() *handler.ReactionHandler {
	return &handler.ReactionHandler{
		ReactionUC: r.usecaseRegistry.NewReactionUseCase(),
	}
}

func (r *InterfaceRegistry) NewUserGroupHandler() *handler.UserGroupHandler {
	return &handler.UserGroupHandler{
		UserGroupUC: r.usecaseRegistry.NewUserGroupUseCase(),
	}
}

func (r *InterfaceRegistry) NewLinkHandler() *handler.LinkHandler {
	return &handler.LinkHandler{
		LinkUC: r.usecaseRegistry.NewLinkUseCase(),
	}
}

func (r *InterfaceRegistry) NewBookmarkHandler() *handler.BookmarkHandler {
	return &handler.BookmarkHandler{
		BookmarkUC: r.usecaseRegistry.NewBookmarkUseCase(),
	}
}

func (r *InterfaceRegistry) NewPinHandler() *handler.PinHandler {
	return &handler.PinHandler{
		UC: r.usecaseRegistry.NewPinUseCase(),
	}
}

func (r *InterfaceRegistry) NewAttachmentHandler() *handler.AttachmentHandler {
	return &handler.AttachmentHandler{
		AttachmentUseCase: r.usecaseRegistry.NewAttachmentUseCase(),
	}
}

func (r *InterfaceRegistry) NewSearchHandler() *handler.SearchHandler {
	return &handler.SearchHandler{
		SearchUC: r.usecaseRegistry.NewSearchUseCase(),
	}
}

func (r *InterfaceRegistry) NewDMHandler() *handler.DMHandler {
	return &handler.DMHandler{
		DMInteractor: r.usecaseRegistry.NewDMInteractor(),
	}
}

func (r *InterfaceRegistry) NewThreadHandler() *handler.ThreadHandler {
	return &handler.ThreadHandler{
		ThreadLister: r.usecaseRegistry.NewThreadLister(),
		ThreadReader: r.usecaseRegistry.NewThreadReader(),
	}
}

func (r *InterfaceRegistry) NewUserHandler() *handler.UserHandler {
	return &handler.UserHandler{
		UC: r.usecaseRegistry.NewUserUseCase(),
	}
}

func (r *InterfaceRegistry) NewRouter() *echo.Echo {
	routerConfig := http.RouterConfig{
		JWTService:           r.infrastructureRegistry.NewJWTService(),
		AllowedOrigins:       r.infrastructureRegistry.config.CORS.AllowedOrigins,
		WebSocketHub:         r.infrastructureRegistry.hub,
		WorkspaceRepository:  r.domainRegistry.NewWorkspaceRepository(),
		MessageUseCase:       r.usecaseRegistry.NewMessageUseCase(),
		ReadStateUseCase:     r.usecaseRegistry.NewReadStateUseCase(),
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
		PinHandler:           r.NewPinHandler(),
		AttachmentHandler:    r.NewAttachmentHandler(),
		SearchHandler:        r.NewSearchHandler(),
		DMHandler:            r.NewDMHandler(),
		ThreadHandler:        r.NewThreadHandler(),
        UserHandler:          r.NewUserHandler(),
	}

	return http.NewRouter(routerConfig)
}

func (r *InterfaceRegistry) NewWebSocketHub() *websocket.Hub {
	return r.infrastructureRegistry.hub
}
