package registry

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	goredis "github.com/redis/go-redis/v9"

	"github.com/newt239/chat/ent"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/gen/chat/v1/chatv1connect"
	"github.com/newt239/chat/internal/infrastructure/appwebhook"
	"github.com/newt239/chat/internal/infrastructure/auth"
	"github.com/newt239/chat/internal/infrastructure/config"
	"github.com/newt239/chat/internal/infrastructure/fcm"
	"github.com/newt239/chat/internal/infrastructure/meilisearch"
	"github.com/newt239/chat/internal/infrastructure/ogp"
	"github.com/newt239/chat/internal/infrastructure/redis"
	"github.com/newt239/chat/internal/infrastructure/repository"
	"github.com/newt239/chat/internal/infrastructure/storage/local"
	"github.com/newt239/chat/internal/infrastructure/storage/wasabi"
	"github.com/newt239/chat/internal/infrastructure/transaction"
	"github.com/newt239/chat/internal/interfaces/handler/httpapi"
	"github.com/newt239/chat/internal/interfaces/handler/rpc"
	"github.com/newt239/chat/internal/interfaces/handler/websocket"
	adminuc "github.com/newt239/chat/internal/usecase/admin"
	appuc "github.com/newt239/chat/internal/usecase/app"
	attachmentuc "github.com/newt239/chat/internal/usecase/attachment"
	"github.com/newt239/chat/internal/usecase/audit"
	authuc "github.com/newt239/chat/internal/usecase/auth"
	bookmarkuc "github.com/newt239/chat/internal/usecase/bookmark"
	channeluc "github.com/newt239/chat/internal/usecase/channel"
	channelcategoryuc "github.com/newt239/chat/internal/usecase/channelcategory"
	channellinkuc "github.com/newt239/chat/internal/usecase/channellink"
	channelmemberuc "github.com/newt239/chat/internal/usecase/channelmember"
	commanduc "github.com/newt239/chat/internal/usecase/command"
	customemojiuc "github.com/newt239/chat/internal/usecase/customemoji"
	dmuc "github.com/newt239/chat/internal/usecase/dm"
	draftuc "github.com/newt239/chat/internal/usecase/draft"
	imageuc "github.com/newt239/chat/internal/usecase/image"
	invitationuc "github.com/newt239/chat/internal/usecase/invitation"
	messageuc "github.com/newt239/chat/internal/usecase/message"
	notificationuc "github.com/newt239/chat/internal/usecase/notification"
	pinuc "github.com/newt239/chat/internal/usecase/pin"
	polluc "github.com/newt239/chat/internal/usecase/poll"
	reactionuc "github.com/newt239/chat/internal/usecase/reaction"
	readstateuc "github.com/newt239/chat/internal/usecase/readstate"
	realtimeuc "github.com/newt239/chat/internal/usecase/realtime"
	scheduledmessageuc "github.com/newt239/chat/internal/usecase/scheduledmessage"
	searchuc "github.com/newt239/chat/internal/usecase/search"
	"github.com/newt239/chat/internal/usecase/searchindex"
	useruc "github.com/newt239/chat/internal/usecase/user"
	usergroupuc "github.com/newt239/chat/internal/usecase/usergroup"
	usernoteuc "github.com/newt239/chat/internal/usecase/usernote"
	workspaceuc "github.com/newt239/chat/internal/usecase/workspace"
)

// App はサーバーの起動と定期処理に必要なものです
type App struct {
	Router           *echo.Echo
	Hub              *websocket.Hub
	SearchIndexer    *searchindex.Indexer
	ScheduledMessage *scheduledmessageuc.Interactor
	Command          *commanduc.Interactor
	Sessions         domainrepository.SessionRepository
}

// newPushSender は FIREBASE_PROJECT_ID が未設定か初期化に失敗したら nil を返し、通知を送らない
func newPushSender(projectID string) notificationuc.Sender {
	if projectID == "" {
		return nil
	}
	sender, err := fcm.NewSender(context.Background(), projectID)
	if err != nil {
		slog.Warn("FCM を初期化できないためプッシュ通知を送りません", "error", err)
		return nil
	}
	return sender
}

// NewSearchIndexer は検索インデックスを作り直すコマンド用の Indexer を組み立てます
func NewSearchIndexer(client *ent.Client, cfg config.SearchConfig) *searchindex.Indexer {
	mentionSvc := service.NewMentionService(repository.NewWorkspaceRepository(client), repository.NewUserRepository(client),
		repository.NewUserGroupRepository(client), repository.NewChannelRepository(client))
	return searchindex.NewIndexer(repository.NewMessageRepository(client), meilisearch.NewMessageIndex(cfg.MeilisearchURL, cfg.MeilisearchAPIKey), mentionSvc)
}

// New は依存関係を組み立てます。ready が false を返す間は readiness probe に 503 を返す
func New(client *ent.Client, cfg *config.Config, rdb *goredis.Client, ready func() bool) *App {
	userRepo := repository.NewUserRepository(client)
	sessionRepo := repository.NewSessionRepository(client)
	workspaceRepo := repository.NewWorkspaceRepository(client)
	invitationRepo := repository.NewInvitationRepository(client)
	channelRepo := repository.NewChannelRepository(client)
	channelMemberRepo := repository.NewChannelMemberRepository(client)
	channelStarRepo := repository.NewChannelStarRepository(client)
	channelMuteRepo := repository.NewChannelMuteRepository(client)
	messageRepo := repository.NewMessageRepository(client)
	systemMessageRepo := repository.NewSystemMessageRepository(client)
	readStateRepo := repository.NewReadStateRepository(client)
	userGroupRepo := repository.NewUserGroupRepository(client)
	mentionRepo := repository.NewMessageMentionRepository(client)
	linkRepo := repository.NewLinkRepository(client)
	attachmentRepo := repository.NewAttachmentRepository(client)
	pinRepo := repository.NewPinRepository(client)
	pollRepo := repository.NewPollRepository(client)
	threadRepo := repository.NewThreadRepository(client)
	pushTokenRepo := repository.NewPushTokenRepository(client)
	appRepo := repository.NewAppRepository(client)
	auditLogRepo := repository.NewAuditLogRepository(client)
	permissionRepo := repository.NewPermissionRepository(client)
	userNoteRepo := repository.NewUserNoteRepository(client)

	channelAccess := service.NewChannelAccessService(messageRepo, channelRepo, channelMemberRepo, workspaceRepo)
	permissionSvc := service.NewPermissionService(workspaceRepo, permissionRepo)
	mentionSvc := service.NewMentionService(workspaceRepo, userRepo, userGroupRepo, channelRepo)
	ogpSvc := ogp.NewOGPService()
	txManager := transaction.NewManager(client)
	recorder := audit.NewRecorder(auditLogRepo)
	jwtService := auth.NewJWTService(cfg.JWT.Secret)
	passwordSvc := auth.PasswordService{}
	googleOAuth := auth.NewGoogleOAuth(cfg.Auth.GoogleOAuthClientID, cfg.Auth.GoogleOAuthClientSecret, cfg.Auth.GoogleOAuthRedirectURL, cfg.Auth.NativeAppRedirectURL)
	searchIndex := meilisearch.NewMessageIndex(cfg.Search.MeilisearchURL, cfg.Search.MeilisearchAPIKey)

	var storage service.StorageService
	var storageHandler http.Handler
	if cfg.Storage.Driver == "local" {
		localStorage := local.New("tmp/storage", cfg.Storage.PublicBaseURL, cfg.JWT.Secret)
		storage, storageHandler = localStorage, localStorage
	} else {
		storage = wasabi.NewPresignService(cfg.Wasabi)
	}

	hub := websocket.NewHub(channelAccess, redis.NewBroker(rdb), redis.NewPresenceStore(rdb))
	notifier := websocket.Notifier{Hub: hub}

	outputBuilder := messageuc.NewMessageOutputBuilder(messageRepo, userRepo, userGroupRepo, mentionRepo, linkRepo, attachmentRepo, pinRepo, pollRepo, channelAccess)
	systemMessages := messageuc.NewSystemMessages(systemMessageRepo, notifier)
	indexer := searchindex.NewIndexer(messageRepo, searchIndex, mentionSvc)
	observers := []messageuc.NewMessageObserver{
		notificationuc.NewDispatcher(userRepo, channelMemberRepo, channelMuteRepo, threadRepo, pushTokenRepo, mentionSvc, channelAccess, newPushSender(cfg.Firebase.ProjectID)),
		appuc.NewEventDispatcher(appRepo, mentionSvc, appwebhook.NewSender()),
	}
	linkSvc := service.NewLinkProcessingService(ogpSvc, linkRepo, messageRepo, channelRepo)
	message := messageuc.New(messageRepo, systemMessageRepo, userRepo, workspaceRepo, threadRepo, attachmentRepo, pollRepo, mentionRepo, linkRepo,
		mentionSvc, linkSvc, ogpSvc, channelAccess, permissionSvc, txManager, outputBuilder, notifier, indexer, observers)
	app := appuc.New(appRepo, userRepo, workspaceRepo, channelRepo, channelMemberRepo, messageRepo, channelAccess, message, txManager, recorder)
	admin := adminuc.New(workspaceRepo, userRepo, sessionRepo, auditLogRepo, permissionRepo, permissionSvc, recorder, hub)
	realtime := realtimeuc.New(redis.NewTicketStore(rdb), workspaceRepo, sessionRepo)
	command := commanduc.New(repository.NewReminderRepository(client), userRepo, workspaceRepo, channelRepo, channelAccess, app)
	scheduled := scheduledmessageuc.New(repository.NewScheduledMessageRepository(client), messageRepo, attachmentRepo, channelAccess, message)

	opts := rpc.HandlerOptions(jwtService, cfg.CORS.AllowedOrigins)
	mux := http.NewServeMux()
	mux.Handle(chatv1connect.NewAuthServiceHandler(&rpc.AuthServer{UC: authuc.New(userRepo, sessionRepo, workspaceRepo, invitationRepo, jwtService, passwordSvc,
		auth.GoogleVerifier{ClientID: cfg.Auth.GoogleOAuthClientID}, googleOAuth, txManager, recorder, hub,
		cfg.Auth.PasswordAuthEnabled)}, opts...))
	mux.Handle(chatv1connect.NewInvitationServiceHandler(&rpc.InvitationServer{UC: invitationuc.New(invitationRepo, workspaceRepo, userRepo, permissionSvc)}, opts...))
	mux.Handle(chatv1connect.NewUserServiceHandler(&rpc.UserServer{
		UC:     useruc.New(userRepo, sessionRepo, workspaceRepo, passwordSvc, hub),
		NoteUC: usernoteuc.New(userNoteRepo, userRepo),
	}, opts...))
	mux.Handle(chatv1connect.NewNotificationServiceHandler(&rpc.NotificationServer{UC: notificationuc.New(pushTokenRepo)}, opts...))
	mux.Handle(chatv1connect.NewWorkspaceServiceHandler(&rpc.WorkspaceServer{UC: workspaceuc.New(workspaceRepo, userRepo, userNoteRepo, txManager), AdminUC: admin}, opts...))
	mux.Handle(chatv1connect.NewChannelServiceHandler(&rpc.ChannelServer{UC: channeluc.New(channelRepo, channelMemberRepo, channelStarRepo, channelMuteRepo, workspaceRepo, readStateRepo,
		txManager, systemMessages, channelAccess, permissionSvc, recorder, hub)}, opts...))
	mux.Handle(chatv1connect.NewChannelLinkServiceHandler(&rpc.ChannelLinkServer{UC: channellinkuc.New(repository.NewChannelLinkRepository(client), channelAccess, permissionSvc, txManager)}, opts...))
	mux.Handle(chatv1connect.NewChannelCategoryServiceHandler(&rpc.ChannelCategoryServer{UC: channelcategoryuc.New(repository.NewChannelCategoryRepository(client), workspaceRepo, channelAccess, txManager)}, opts...))
	mux.Handle(chatv1connect.NewChannelMemberServiceHandler(&rpc.ChannelMemberServer{UC: channelmemberuc.New(channelRepo, channelMemberRepo, workspaceRepo, userRepo, systemMessages, channelAccess, txManager, hub)}, opts...))
	mux.Handle(chatv1connect.NewReadStateServiceHandler(&rpc.ReadStateServer{UC: readstateuc.New(readStateRepo, notifier, channelAccess)}, opts...))
	mux.Handle(chatv1connect.NewDirectMessageServiceHandler(&rpc.DirectMessageServer{UC: dmuc.New(channelRepo, channelMemberRepo, channelStarRepo, channelMuteRepo, readStateRepo, userRepo, workspaceRepo)}, opts...))
	mux.Handle(chatv1connect.NewUserGroupServiceHandler(&rpc.UserGroupServer{UC: usergroupuc.New(userGroupRepo, workspaceRepo, userRepo)}, opts...))
	mux.Handle(chatv1connect.NewBookmarkServiceHandler(&rpc.BookmarkServer{UC: bookmarkuc.New(repository.NewBookmarkRepository(client), outputBuilder, channelAccess)}, opts...))
	mux.Handle(chatv1connect.NewLinkServiceHandler(&rpc.LinkServer{UC: message}, opts...))
	mux.Handle(chatv1connect.NewAttachmentServiceHandler(&rpc.AttachmentServer{UC: attachmentuc.New(attachmentRepo, messageRepo, channelAccess, storage)}, opts...))
	mux.Handle(chatv1connect.NewSearchServiceHandler(&rpc.SearchServer{UC: searchuc.New(workspaceRepo, channelRepo, messageRepo, searchIndex, userRepo, userGroupRepo, outputBuilder)}, opts...))
	mux.Handle(chatv1connect.NewMentionServiceHandler(&rpc.MentionServer{UC: message}, opts...))
	mux.Handle(chatv1connect.NewMessageServiceHandler(&rpc.MessageServer{UC: message}, opts...))
	mux.Handle(chatv1connect.NewDraftServiceHandler(&rpc.DraftServer{UC: draftuc.New(repository.NewDraftRepository(client), messageRepo, channelAccess)}, opts...))
	mux.Handle(chatv1connect.NewScheduledMessageServiceHandler(&rpc.ScheduledMessageServer{UC: scheduled}, opts...))
	mux.Handle(chatv1connect.NewThreadServiceHandler(&rpc.ThreadServer{UC: message}, opts...))
	mux.Handle(chatv1connect.NewReactionServiceHandler(&rpc.ReactionServer{UC: reactionuc.New(messageRepo, userRepo, notifier, channelAccess)}, opts...))
	mux.Handle(chatv1connect.NewPinServiceHandler(&rpc.PinServer{UC: pinuc.New(pinRepo, channelMemberRepo, userRepo, notifier, outputBuilder, channelAccess, systemMessages, permissionSvc, indexer)}, opts...))
	mux.Handle(chatv1connect.NewAdminServiceHandler(&rpc.AdminServer{UC: admin}, opts...))
	mux.Handle(chatv1connect.NewPermissionServiceHandler(&rpc.PermissionServer{UC: admin}, opts...))
	mux.Handle(chatv1connect.NewPollServiceHandler(&rpc.PollServer{UC: polluc.New(pollRepo, messageRepo, workspaceRepo, channelAccess, outputBuilder, notifier)}, opts...))
	mux.Handle(chatv1connect.NewCommandServiceHandler(&rpc.CommandServer{UC: command}, opts...))
	mux.Handle(chatv1connect.NewAppServiceHandler(&rpc.AppServer{UC: app}, opts...))
	mux.Handle(chatv1connect.NewCustomEmojiServiceHandler(&rpc.CustomEmojiServer{UC: customemojiuc.New(repository.NewCustomEmojiRepository(client), userRepo, workspaceRepo, permissionSvc, storage, notifier, recorder)}, opts...))
	mux.Handle(chatv1connect.NewImageServiceHandler(&rpc.ImageServer{UC: imageuc.New(workspaceRepo, storage, cfg.Storage.PublicBaseURL)}, opts...))
	mux.Handle(chatv1connect.NewRealtimeServiceHandler(&rpc.RealtimeServer{UC: realtime}, opts...))

	router := httpapi.NewRouter(httpapi.RouterConfig{
		AllowedOrigins:     cfg.CORS.AllowedOrigins,
		WebSocketHub:       hub,
		Tickets:            realtime,
		RPCHandler:         mux,
		WebhookPoster:      app,
		WebhookRateLimiter: redis.NewRateLimiter(rdb, "webhook", httpapi.WebhookRatePerSecond, httpapi.WebhookBurst),
		Ready:              ready,
		GoogleOAuth:        googleOAuth,
		StorageHandler:     storageHandler,
		Storage:            storage,
	})
	return &App{Router: router, Hub: hub, SearchIndexer: indexer, ScheduledMessage: scheduled, Command: command, Sessions: sessionRepo}
}
