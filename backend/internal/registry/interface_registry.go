package registry

import (
	nethttp "net/http"

	"github.com/labstack/echo/v4"

	"github.com/newt239/chat/internal/gen/chat/v1/chatv1connect"
	"github.com/newt239/chat/internal/interfaces/handler/http"
	"github.com/newt239/chat/internal/interfaces/handler/rpc"
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

func (r *InterfaceRegistry) NewRPCHandler() nethttp.Handler {
	uc := r.usecaseRegistry
	return rpc.NewHandler(r.infrastructureRegistry.NewJWTService(), r.infrastructureRegistry.config.CORS.AllowedOrigins,
		rpc.Register(chatv1connect.NewAuthServiceHandler, chatv1connect.AuthServiceHandler(&rpc.AuthServer{UC: uc.NewAuthUseCase()})),
		rpc.Register(chatv1connect.NewInvitationServiceHandler, chatv1connect.InvitationServiceHandler(&rpc.InvitationServer{UC: uc.NewInvitationUseCase()})),
		rpc.Register(chatv1connect.NewUserServiceHandler, chatv1connect.UserServiceHandler(&rpc.UserServer{UC: uc.NewUserUseCase(), NoteUC: uc.NewUserNoteUseCase()})),
		rpc.Register(chatv1connect.NewNotificationServiceHandler, chatv1connect.NotificationServiceHandler(&rpc.NotificationServer{UC: uc.NewNotificationUseCase()})),
		rpc.Register(chatv1connect.NewWorkspaceServiceHandler, chatv1connect.WorkspaceServiceHandler(&rpc.WorkspaceServer{UC: uc.NewWorkspaceUseCase()})),
		rpc.Register(chatv1connect.NewChannelServiceHandler, chatv1connect.ChannelServiceHandler(&rpc.ChannelServer{UC: uc.NewChannelUseCase()})),
		rpc.Register(chatv1connect.NewChannelLinkServiceHandler, chatv1connect.ChannelLinkServiceHandler(&rpc.ChannelLinkServer{UC: uc.NewChannelLinkUseCase()})),
		rpc.Register(chatv1connect.NewChannelCategoryServiceHandler, chatv1connect.ChannelCategoryServiceHandler(&rpc.ChannelCategoryServer{UC: uc.NewChannelCategoryUseCase()})),
		rpc.Register(chatv1connect.NewChannelMemberServiceHandler, chatv1connect.ChannelMemberServiceHandler(&rpc.ChannelMemberServer{UC: uc.NewChannelMemberUseCase()})),
		rpc.Register(chatv1connect.NewReadStateServiceHandler, chatv1connect.ReadStateServiceHandler(&rpc.ReadStateServer{UC: uc.NewReadStateUseCase()})),
		rpc.Register(chatv1connect.NewDirectMessageServiceHandler, chatv1connect.DirectMessageServiceHandler(&rpc.DirectMessageServer{UC: uc.NewDMInteractor()})),
		rpc.Register(chatv1connect.NewUserGroupServiceHandler, chatv1connect.UserGroupServiceHandler(&rpc.UserGroupServer{UC: uc.NewUserGroupUseCase()})),
		rpc.Register(chatv1connect.NewBookmarkServiceHandler, chatv1connect.BookmarkServiceHandler(&rpc.BookmarkServer{UC: uc.NewBookmarkUseCase()})),
		rpc.Register(chatv1connect.NewLinkServiceHandler, chatv1connect.LinkServiceHandler(&rpc.LinkServer{UC: uc.NewLinkUseCase()})),
		rpc.Register(chatv1connect.NewAttachmentServiceHandler, chatv1connect.AttachmentServiceHandler(&rpc.AttachmentServer{UC: uc.NewAttachmentUseCase()})),
		rpc.Register(chatv1connect.NewSearchServiceHandler, chatv1connect.SearchServiceHandler(&rpc.SearchServer{UC: uc.NewSearchUseCase()})),
		rpc.Register(chatv1connect.NewMentionServiceHandler, chatv1connect.MentionServiceHandler(&rpc.MentionServer{Lister: uc.NewMentionLister()})),
		rpc.Register(chatv1connect.NewMessageServiceHandler, chatv1connect.MessageServiceHandler(&rpc.MessageServer{
			Creator: uc.NewMessageCreator(),
			Updater: uc.NewMessageUpdater(),
			Deleter: uc.NewMessageDeleter(),
			Lister:  uc.NewMessageLister(),
		})),
		rpc.Register(chatv1connect.NewDraftServiceHandler, chatv1connect.DraftServiceHandler(&rpc.DraftServer{UC: uc.NewDraftUseCase()})),
		rpc.Register(chatv1connect.NewScheduledMessageServiceHandler, chatv1connect.ScheduledMessageServiceHandler(&rpc.ScheduledMessageServer{UC: uc.NewScheduledMessageUseCase()})),
		rpc.Register(chatv1connect.NewThreadServiceHandler, chatv1connect.ThreadServiceHandler(&rpc.ThreadServer{
			MessageLister: uc.NewMessageLister(),
			ThreadLister:  uc.NewThreadLister(),
			ThreadReader:  uc.NewThreadReader(),
		})),
		rpc.Register(chatv1connect.NewReactionServiceHandler, chatv1connect.ReactionServiceHandler(&rpc.ReactionServer{UC: uc.NewReactionUseCase()})),
		rpc.Register(chatv1connect.NewPinServiceHandler, chatv1connect.PinServiceHandler(&rpc.PinServer{UC: uc.NewPinUseCase()})),
		rpc.Register(chatv1connect.NewAdminServiceHandler, chatv1connect.AdminServiceHandler(&rpc.AdminServer{UC: uc.NewAdminUseCase()})),
		rpc.Register(chatv1connect.NewPermissionServiceHandler, chatv1connect.PermissionServiceHandler(&rpc.PermissionServer{UC: uc.NewAdminUseCase()})),
		rpc.Register(chatv1connect.NewInsightServiceHandler, chatv1connect.InsightServiceHandler(&rpc.InsightServer{UC: uc.NewInsightUseCase()})),
		rpc.Register(chatv1connect.NewPollServiceHandler, chatv1connect.PollServiceHandler(&rpc.PollServer{UC: uc.NewPollUseCase()})),
		rpc.Register(chatv1connect.NewCommandServiceHandler, chatv1connect.CommandServiceHandler(&rpc.CommandServer{UC: uc.NewCommandUseCase()})),
		rpc.Register(chatv1connect.NewAppServiceHandler, chatv1connect.AppServiceHandler(&rpc.AppServer{UC: uc.NewAppUseCase()})),
		rpc.Register(chatv1connect.NewCustomEmojiServiceHandler, chatv1connect.CustomEmojiServiceHandler(&rpc.CustomEmojiServer{UC: uc.NewCustomEmojiUseCase()})),
		rpc.Register(chatv1connect.NewImageServiceHandler, chatv1connect.ImageServiceHandler(&rpc.ImageServer{UC: uc.NewImageUseCase()})),
		rpc.Register(chatv1connect.NewRealtimeServiceHandler, chatv1connect.RealtimeServiceHandler(&rpc.RealtimeServer{UC: uc.NewRealtimeUseCase()})),
	)
}

func (r *InterfaceRegistry) NewRouter() *echo.Echo {
	routerConfig := http.RouterConfig{
		AllowedOrigins:     r.infrastructureRegistry.config.CORS.AllowedOrigins,
		TrustedProxies:     r.infrastructureRegistry.config.Server.TrustedProxies,
		WebSocketHub:       r.infrastructureRegistry.hub,
		TicketConsumer:     r.usecaseRegistry.NewRealtimeUseCase(),
		RPCHandler:         r.NewRPCHandler(),
		WebhookPoster:      r.usecaseRegistry.NewAppUseCase(),
		WebhookRateLimiter: r.infrastructureRegistry.NewWebhookRateLimiter(),
		Ready:              r.infrastructureRegistry.Ready,
		GoogleOAuth:        r.infrastructureRegistry.NewGoogleOAuth(),
		Storage:            r.infrastructureRegistry.NewStorageService(),
		StorageConfig:      r.infrastructureRegistry.NewStorageConfig(),
	}
	if r.infrastructureRegistry.config.Storage.Driver == "local" {
		routerConfig.StorageHandler = r.infrastructureRegistry.NewLocalStorage()
	}

	return http.NewRouter(routerConfig)
}

func (r *InterfaceRegistry) NewWebSocketHub() *websocket.Hub {
	return r.infrastructureRegistry.hub
}
