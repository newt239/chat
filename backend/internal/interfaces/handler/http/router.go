package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/interfaces/handler/http/handler"
	custommw "github.com/newt239/chat/internal/interfaces/handler/http/middleware"
	"github.com/newt239/chat/internal/interfaces/handler/websocket"
	openapi "github.com/newt239/chat/internal/openapi_gen"
	authuc "github.com/newt239/chat/internal/usecase/auth"
)

type RouterConfig struct {
	JWTService     authuc.JWTService
	AllowedOrigins []string

	WebSocketHub        *websocket.Hub
	WorkspaceRepository repository.WorkspaceRepository
	MessageUseCase      websocket.MessageUseCase
	ReadStateUseCase    websocket.ReadStateUseCase

	AuthHandler          *handler.AuthHandler
	WorkspaceHandler     *handler.WorkspaceHandler
	ChannelHandler       *handler.ChannelHandler
	ChannelMemberHandler *handler.ChannelMemberHandler
	MessageHandler       *handler.MessageHandler
	ReadStateHandler     *handler.ReadStateHandler
	ReactionHandler      *handler.ReactionHandler
	UserGroupHandler     *handler.UserGroupHandler
	LinkHandler          *handler.LinkHandler
	BookmarkHandler      *handler.BookmarkHandler
	PinHandler           *handler.PinHandler
	AttachmentHandler    *handler.AttachmentHandler
	SearchHandler        *handler.SearchHandler
	DMHandler            *handler.DMHandler
	ThreadHandler        *handler.ThreadHandler
	UserHandler          *handler.UserHandler
}

type serverImpl struct {
	cfg RouterConfig
}

func (s *serverImpl) Healthz(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (s *serverImpl) PresignUpload(ctx echo.Context) error {
	return s.cfg.AttachmentHandler.PresignUpload(ctx)
}

func (s *serverImpl) GetAttachment(ctx echo.Context, id openapi_types.UUID) error {
	return s.cfg.AttachmentHandler.GetAttachment(ctx, id)
}

func (s *serverImpl) DownloadAttachment(ctx echo.Context, id openapi_types.UUID) error {
	return s.cfg.AttachmentHandler.DownloadAttachment(ctx, id)
}

func (s *serverImpl) Login(ctx echo.Context) error {
	return s.cfg.AuthHandler.Login(ctx)
}

func (s *serverImpl) Logout(ctx echo.Context) error {
	return s.cfg.AuthHandler.Logout(ctx)
}

func (s *serverImpl) Refresh(ctx echo.Context) error {
	return s.cfg.AuthHandler.Refresh(ctx)
}

func (s *serverImpl) Register(ctx echo.Context) error {
	return s.cfg.AuthHandler.Register(ctx)
}

func (s *serverImpl) ListBookmarks(ctx echo.Context) error {
	return s.cfg.BookmarkHandler.ListBookmarks(ctx)
}

func (s *serverImpl) AddBookmark(ctx echo.Context, messageId openapi_types.UUID) error {
	return s.cfg.BookmarkHandler.AddBookmark(ctx, messageId)
}

func (s *serverImpl) RemoveBookmark(ctx echo.Context, messageId openapi_types.UUID) error {
	return s.cfg.BookmarkHandler.RemoveBookmark(ctx, messageId)
}

func (s *serverImpl) UpdateChannel(ctx echo.Context, channelId openapi_types.UUID) error {
	return s.cfg.ChannelHandler.UpdateChannel(ctx, channelId)
}

func (s *serverImpl) ListChannels(ctx echo.Context, id string) error {
	return s.cfg.ChannelHandler.ListChannels(ctx, id)
}

func (s *serverImpl) CreateChannel(ctx echo.Context, id string) error {
	return s.cfg.ChannelHandler.CreateChannel(ctx, id)
}

func (s *serverImpl) ListChannelMembers(ctx echo.Context, channelId openapi_types.UUID) error {
	return s.cfg.ChannelMemberHandler.ListChannelMembers(ctx, channelId)
}

func (s *serverImpl) InviteChannelMember(ctx echo.Context, channelId openapi_types.UUID) error {
	return s.cfg.ChannelMemberHandler.InviteChannelMember(ctx, channelId)
}

func (s *serverImpl) LeaveChannel(ctx echo.Context, channelId openapi_types.UUID) error {
	return s.cfg.ChannelMemberHandler.LeaveChannel(ctx, channelId)
}

func (s *serverImpl) JoinPublicChannel(ctx echo.Context, channelId openapi_types.UUID) error {
	return s.cfg.ChannelMemberHandler.JoinPublicChannel(ctx, channelId)
}

func (s *serverImpl) RemoveChannelMember(ctx echo.Context, channelId openapi_types.UUID, userId openapi_types.UUID) error {
	return s.cfg.ChannelMemberHandler.RemoveChannelMember(ctx, channelId, userId)
}

func (s *serverImpl) UpdateChannelMemberRole(ctx echo.Context, channelId openapi_types.UUID, userId openapi_types.UUID) error {
	return s.cfg.ChannelMemberHandler.UpdateChannelMemberRole(ctx, channelId, userId)
}

func (s *serverImpl) ListDMs(ctx echo.Context, id string) error {
	return s.cfg.DMHandler.ListDMs(ctx, id)
}

func (s *serverImpl) CreateDM(ctx echo.Context, id string) error {
	return s.cfg.DMHandler.CreateDM(ctx, id)
}

func (s *serverImpl) CreateGroupDM(ctx echo.Context, id string) error {
	return s.cfg.DMHandler.CreateGroupDM(ctx, id)
}

func (s *serverImpl) FetchOGP(ctx echo.Context) error {
	return s.cfg.LinkHandler.FetchOGP(ctx)
}

func (s *serverImpl) ListMessages(ctx echo.Context, channelId openapi_types.UUID, params openapi.ListMessagesParams) error {
	return s.cfg.MessageHandler.ListMessages(ctx, channelId, params)
}

func (s *serverImpl) CreateMessage(ctx echo.Context, channelId openapi_types.UUID) error {
	return s.cfg.MessageHandler.CreateMessage(ctx, channelId)
}

func (s *serverImpl) ListMessagesWithThread(ctx echo.Context, channelId openapi_types.UUID, params openapi.ListMessagesWithThreadParams) error {
	return s.cfg.MessageHandler.ListMessagesWithThread(ctx, channelId, params)
}

func (s *serverImpl) DeleteMessage(ctx echo.Context, messageId openapi_types.UUID) error {
	return s.cfg.MessageHandler.DeleteMessage(ctx, messageId)
}

func (s *serverImpl) UpdateMessage(ctx echo.Context, messageId openapi_types.UUID) error {
	return s.cfg.MessageHandler.UpdateMessage(ctx, messageId)
}

func (s *serverImpl) GetThreadReplies(ctx echo.Context, messageId openapi_types.UUID, params openapi.GetThreadRepliesParams) error {
	return s.cfg.MessageHandler.GetThreadReplies(ctx, messageId, params)
}

func (s *serverImpl) GetThreadMetadata(ctx echo.Context, messageId openapi_types.UUID) error {
	return s.cfg.MessageHandler.GetThreadMetadata(ctx, messageId)
}

func (s *serverImpl) ListPins(ctx echo.Context, channelId openapi_types.UUID, params openapi.ListPinsParams) error {
	return s.cfg.PinHandler.ListPins(ctx, channelId, params)
}

func (s *serverImpl) CreatePin(ctx echo.Context, channelId openapi_types.UUID) error {
	return s.cfg.PinHandler.CreatePin(ctx, channelId)
}

func (s *serverImpl) DeletePin(ctx echo.Context, channelId openapi_types.UUID, messageId openapi_types.UUID) error {
	return s.cfg.PinHandler.DeletePin(ctx, channelId, messageId)
}

func (s *serverImpl) UpdateReadState(ctx echo.Context, channelId openapi_types.UUID) error {
	return s.cfg.ReadStateHandler.UpdateReadState(ctx, channelId)
}

func (s *serverImpl) GetUnreadCount(ctx echo.Context, channelId openapi_types.UUID) error {
	return s.cfg.ReadStateHandler.GetUnreadCount(ctx, channelId)
}

func (s *serverImpl) ListReactions(ctx echo.Context, messageId openapi_types.UUID) error {
	return s.cfg.ReactionHandler.ListReactions(ctx, messageId)
}

func (s *serverImpl) AddReaction(ctx echo.Context, messageId openapi_types.UUID) error {
	return s.cfg.ReactionHandler.AddReaction(ctx, messageId)
}

func (s *serverImpl) RemoveReaction(ctx echo.Context, messageId openapi_types.UUID, emoji string) error {
	return s.cfg.ReactionHandler.RemoveReaction(ctx, messageId, emoji)
}

func (s *serverImpl) SearchWorkspace(ctx echo.Context, workspaceId string, params openapi.SearchWorkspaceParams) error {
	return s.cfg.SearchHandler.SearchWorkspace(ctx, workspaceId, params)
}

func (s *serverImpl) MarkThreadRead(ctx echo.Context, threadId openapi_types.UUID) error {
	return s.cfg.ThreadHandler.MarkThreadRead(ctx, threadId)
}

func (s *serverImpl) GetParticipatingThreads(ctx echo.Context, workspaceId string, params openapi.GetParticipatingThreadsParams) error {
	return s.cfg.ThreadHandler.GetParticipatingThreads(ctx, workspaceId, params)
}

func (s *serverImpl) ListUserGroups(ctx echo.Context, params openapi.ListUserGroupsParams) error {
	return s.cfg.UserGroupHandler.ListUserGroups(ctx, params)
}

func (s *serverImpl) CreateUserGroup(ctx echo.Context) error {
	return s.cfg.UserGroupHandler.CreateUserGroup(ctx)
}

func (s *serverImpl) DeleteUserGroup(ctx echo.Context, id openapi_types.UUID) error {
	return s.cfg.UserGroupHandler.DeleteUserGroup(ctx, id)
}

func (s *serverImpl) GetUserGroup(ctx echo.Context, id openapi_types.UUID) error {
	return s.cfg.UserGroupHandler.GetUserGroup(ctx, id)
}

func (s *serverImpl) UpdateUserGroup(ctx echo.Context, id openapi_types.UUID) error {
	return s.cfg.UserGroupHandler.UpdateUserGroup(ctx, id)
}

func (s *serverImpl) RemoveUserGroupMember(ctx echo.Context, id openapi_types.UUID, params openapi.RemoveUserGroupMemberParams) error {
	return s.cfg.UserGroupHandler.RemoveUserGroupMember(ctx, id, params)
}

func (s *serverImpl) ListUserGroupMembers(ctx echo.Context, id openapi_types.UUID) error {
	return s.cfg.UserGroupHandler.ListUserGroupMembers(ctx, id)
}

func (s *serverImpl) AddUserGroupMember(ctx echo.Context, id openapi_types.UUID) error {
	return s.cfg.UserGroupHandler.AddUserGroupMember(ctx, id)
}

func (s *serverImpl) UpdateMe(ctx echo.Context) error {
	return s.cfg.UserHandler.UpdateMe(ctx)
}

func (s *serverImpl) ListWorkspaces(ctx echo.Context) error {
	return s.cfg.WorkspaceHandler.ListWorkspaces(ctx)
}

func (s *serverImpl) CreateWorkspace(ctx echo.Context) error {
	return s.cfg.WorkspaceHandler.CreateWorkspace(ctx)
}

func (s *serverImpl) ListPublicWorkspaces(ctx echo.Context) error {
	return s.cfg.WorkspaceHandler.ListPublicWorkspaces(ctx)
}

func (s *serverImpl) DeleteWorkspace(ctx echo.Context, id string) error {
	return s.cfg.WorkspaceHandler.DeleteWorkspace(ctx, id)
}

func (s *serverImpl) GetWorkspace(ctx echo.Context, id string) error {
	return s.cfg.WorkspaceHandler.GetWorkspace(ctx, id)
}

func (s *serverImpl) UpdateWorkspace(ctx echo.Context, id string) error {
	return s.cfg.WorkspaceHandler.UpdateWorkspace(ctx, id)
}

func (s *serverImpl) JoinPublicWorkspace(ctx echo.Context, id string) error {
	return s.cfg.WorkspaceHandler.JoinPublicWorkspace(ctx, id)
}

func (s *serverImpl) ListMembers(ctx echo.Context, id string) error {
	return s.cfg.WorkspaceHandler.ListMembers(ctx, id)
}

func (s *serverImpl) AddMemberByEmail(ctx echo.Context, id string) error {
	return s.cfg.WorkspaceHandler.AddMemberByEmail(ctx, id)
}

func (s *serverImpl) RemoveMember(ctx echo.Context, id string, userId openapi_types.UUID) error {
	return s.cfg.WorkspaceHandler.RemoveMember(ctx, id, userId)
}

func (s *serverImpl) UpdateMemberRole(ctx echo.Context, id string, userId openapi_types.UUID) error {
	return s.cfg.WorkspaceHandler.UpdateMemberRole(ctx, id, userId)
}

// registerPublicRoutes は認証不要のエンドポイントを登録します
func registerPublicRoutes(e *echo.Echo, wrapper *openapi.ServerInterfaceWrapper) {
	e.POST("/api/auth/login", wrapper.Login)
	e.POST("/api/auth/register", wrapper.Register)
	e.POST("/api/auth/refresh", wrapper.Refresh)
	e.GET("/healthz", wrapper.Healthz)
}

// registerProtectedRoutes は認証が必要なエンドポイントを登録します
func registerProtectedRoutes(protectedAPI *echo.Group, wrapper *openapi.ServerInterfaceWrapper) {
	// アタッチメント
	protectedAPI.POST("/attachments/presign", wrapper.PresignUpload)
	protectedAPI.GET("/attachments/:id", wrapper.GetAttachment)
	protectedAPI.GET("/attachments/:id/download", wrapper.DownloadAttachment)

	// ブックマーク
	protectedAPI.GET("/bookmarks", wrapper.ListBookmarks)
	protectedAPI.POST("/bookmarks/:messageId", wrapper.AddBookmark)
	protectedAPI.DELETE("/bookmarks/:messageId", wrapper.RemoveBookmark)

	// チャンネル
	protectedAPI.PATCH("/channels/:channelId", wrapper.UpdateChannel)
	protectedAPI.GET("/channels/:channelId/members", wrapper.ListChannelMembers)
	protectedAPI.POST("/channels/:channelId/members", wrapper.InviteChannelMember)
	protectedAPI.DELETE("/channels/:channelId/members/self", wrapper.LeaveChannel)
	protectedAPI.POST("/channels/:channelId/members/self", wrapper.JoinPublicChannel)
	protectedAPI.DELETE("/channels/:channelId/members/:userId", wrapper.RemoveChannelMember)
	protectedAPI.PATCH("/channels/:channelId/members/:userId/role", wrapper.UpdateChannelMemberRole)

	// メッセージ
	protectedAPI.GET("/channels/:channelId/messages", wrapper.ListMessages)
	protectedAPI.POST("/channels/:channelId/messages", wrapper.CreateMessage)
	protectedAPI.GET("/channels/:channelId/messages/with-threads", wrapper.ListMessagesWithThread)
	protectedAPI.PATCH("/messages/:messageId", wrapper.UpdateMessage)
	protectedAPI.DELETE("/messages/:messageId", wrapper.DeleteMessage)
	protectedAPI.GET("/messages/:messageId/thread", wrapper.GetThreadReplies)
	protectedAPI.GET("/messages/:messageId/thread/metadata", wrapper.GetThreadMetadata)

	// リアクション
	protectedAPI.GET("/messages/:messageId/reactions", wrapper.ListReactions)
	protectedAPI.POST("/messages/:messageId/reactions", wrapper.AddReaction)
	protectedAPI.DELETE("/messages/:messageId/reactions/:emoji", wrapper.RemoveReaction)

	// ピン
	protectedAPI.GET("/channels/:channelId/pins", wrapper.ListPins)
	protectedAPI.POST("/channels/:channelId/pins", wrapper.CreatePin)
	protectedAPI.DELETE("/channels/:channelId/pins/:messageId", wrapper.DeletePin)

	// 読み取り状態
	protectedAPI.POST("/channels/:channelId/reads", wrapper.UpdateReadState)
	protectedAPI.GET("/channels/:channelId/unread-count", wrapper.GetUnreadCount)

	// DM
	protectedAPI.GET("/workspaces/:id/dms", wrapper.ListDMs)
	protectedAPI.POST("/workspaces/:id/dms", wrapper.CreateDM)
	protectedAPI.POST("/workspaces/:id/group-dms", wrapper.CreateGroupDM)

	// スレッド
	protectedAPI.PATCH("/threads/:threadId/read", wrapper.MarkThreadRead)
	protectedAPI.GET("/workspaces/:workspaceId/threads/participating", wrapper.GetParticipatingThreads)

	// ユーザーグループ
	protectedAPI.GET("/user-groups", wrapper.ListUserGroups)
	protectedAPI.POST("/user-groups", wrapper.CreateUserGroup)
	protectedAPI.GET("/user-groups/:id", wrapper.GetUserGroup)
	protectedAPI.PATCH("/user-groups/:id", wrapper.UpdateUserGroup)
	protectedAPI.DELETE("/user-groups/:id", wrapper.DeleteUserGroup)
	protectedAPI.GET("/user-groups/:id/members", wrapper.ListUserGroupMembers)
	protectedAPI.POST("/user-groups/:id/members", wrapper.AddUserGroupMember)
	protectedAPI.DELETE("/user-groups/:id/members", wrapper.RemoveUserGroupMember)

	// ワークスペース
	protectedAPI.GET("/workspaces", wrapper.ListWorkspaces)
	protectedAPI.POST("/workspaces", wrapper.CreateWorkspace)
	protectedAPI.GET("/workspaces/public", wrapper.ListPublicWorkspaces)
	protectedAPI.GET("/workspaces/:id", wrapper.GetWorkspace)
	protectedAPI.PATCH("/workspaces/:id", wrapper.UpdateWorkspace)
	protectedAPI.DELETE("/workspaces/:id", wrapper.DeleteWorkspace)
	protectedAPI.POST("/workspaces/:id/join", wrapper.JoinPublicWorkspace)
	protectedAPI.GET("/workspaces/:id/channels", wrapper.ListChannels)
	protectedAPI.POST("/workspaces/:id/channels", wrapper.CreateChannel)
	protectedAPI.GET("/workspaces/:id/members", wrapper.ListMembers)
	protectedAPI.POST("/workspaces/:id/members", wrapper.AddMemberByEmail)
	protectedAPI.DELETE("/workspaces/:id/members/:userId", wrapper.RemoveMember)
	protectedAPI.PATCH("/workspaces/:id/members/:userId/role", wrapper.UpdateMemberRole)
	protectedAPI.GET("/workspaces/:workspaceId/search", wrapper.SearchWorkspace)

	// ユーザー
	protectedAPI.PATCH("/users/me", wrapper.UpdateMe)

	// リンク
	protectedAPI.POST("/links/fetch-ogp", wrapper.FetchOGP)
}

func NewRouter(cfg RouterConfig) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.Validator = NewValidator()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
	}))

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// WebSocket
	e.GET("/ws", websocket.Handler(cfg.WebSocketHub, cfg.JWTService, cfg.WorkspaceRepository, cfg.MessageUseCase, cfg.ReadStateUseCase))

	// ServerInterfaceを実装する構造体を作成
	server := &serverImpl{cfg: cfg}

	// oapi-codegenのラッパーを作成
	// ServerInterfaceWrapperはすべてのパラメータバリデーションを自動的に行う
	wrapper := &openapi.ServerInterfaceWrapper{
		Handler: server,
	}

	// 認証不要のエンドポイント
	registerPublicRoutes(e, wrapper)

	// 認証が必要なエンドポイント
	protectedAPI := e.Group("/api")
	protectedAPI.Use(custommw.Auth(cfg.JWTService))
	registerProtectedRoutes(protectedAPI, wrapper)

	return e
}
