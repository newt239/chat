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

func (s *serverImpl) ListDMs(ctx echo.Context, id openapi_types.UUID) error {
	return s.cfg.DMHandler.ListDMs(ctx, id)
}

func (s *serverImpl) CreateDM(ctx echo.Context, id openapi_types.UUID) error {
	return s.cfg.DMHandler.CreateDM(ctx, id)
}

func (s *serverImpl) CreateGroupDM(ctx echo.Context, id openapi_types.UUID) error {
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

	e.GET("/ws/:workspaceId", websocket.Handler(cfg.WebSocketHub, cfg.JWTService, cfg.WorkspaceRepository, cfg.MessageUseCase, cfg.ReadStateUseCase))

	server := &serverImpl{cfg: cfg}

	e.POST("/api/auth/login", server.Login)
	e.POST("/api/auth/register", server.Register)
	e.POST("/api/auth/refresh", server.Refresh)
	e.GET("/healthz", server.Healthz)

	protectedAPI := e.Group("/api")
	protectedAPI.Use(custommw.Auth(cfg.JWTService))

	openapi.RegisterHandlersWithBaseURL(e, server, "/api")

	return e
}
