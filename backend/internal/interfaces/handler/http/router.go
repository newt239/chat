package http

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/interfaces/handler/http/handler"
	custommw "github.com/newt239/chat/internal/interfaces/handler/http/middleware"
	"github.com/newt239/chat/internal/interfaces/handler/websocket"
	authuc "github.com/newt239/chat/internal/usecase/auth"
)

type RouterConfig struct {
	JWTService     authuc.JWTService
	AllowedOrigins []string

	// WebSocket
	WebSocketHub        *websocket.Hub
	WorkspaceRepository repository.WorkspaceRepository

	// Handlers
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
	AttachmentHandler    *handler.AttachmentHandler
}

func NewRouter(cfg RouterConfig) *echo.Echo {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(custommw.CORS(cfg.AllowedOrigins))

	// Validator
	e.Validator = NewValidator()

	// Health check
	e.GET("/healthz", func(c echo.Context) error {
		return c.String(200, "ok")
	})

	// WebSocket endpoint
	e.GET("/ws", websocket.NewHandler(cfg.WebSocketHub, cfg.JWTService, cfg.WorkspaceRepository, nil, nil))

	// API routes
	api := e.Group("/api")

	// Auth routes (public)
	auth := api.Group("/auth")
	{
		auth.POST("/register", cfg.AuthHandler.Register)
		auth.POST("/login", cfg.AuthHandler.Login)
		auth.POST("/refresh", cfg.AuthHandler.RefreshToken)
		auth.POST("/logout", cfg.AuthHandler.Logout)
	}

	// Protected routes
	authMw := custommw.Auth(cfg.JWTService)

	// Workspace routes
	api.GET("/workspaces", cfg.WorkspaceHandler.GetWorkspaces, authMw)
	api.POST("/workspaces", cfg.WorkspaceHandler.CreateWorkspace, authMw)
	api.GET("/workspaces/:id", cfg.WorkspaceHandler.GetWorkspace, authMw)
	api.PATCH("/workspaces/:id", cfg.WorkspaceHandler.UpdateWorkspace, authMw)
	api.DELETE("/workspaces/:id", cfg.WorkspaceHandler.DeleteWorkspace, authMw)
	api.GET("/workspaces/:id/members", cfg.WorkspaceHandler.ListMembers, authMw)
	api.POST("/workspaces/:id/members", cfg.WorkspaceHandler.AddMember, authMw)
	api.PATCH("/workspaces/:id/members/:userId", cfg.WorkspaceHandler.UpdateMemberRole, authMw)
	api.DELETE("/workspaces/:id/members/:userId", cfg.WorkspaceHandler.RemoveMember, authMw)

	// Channel routes
	api.GET("/workspaces/:id/channels", cfg.ChannelHandler.ListChannels, authMw)
	api.POST("/workspaces/:id/channels", cfg.ChannelHandler.CreateChannel, authMw)

	// Channel member routes
	api.GET("/channels/:channelId/members", cfg.ChannelMemberHandler.ListMembers, authMw)
	api.POST("/channels/:channelId/members", cfg.ChannelMemberHandler.InviteMember, authMw)
	api.POST("/channels/:channelId/members/self", cfg.ChannelMemberHandler.JoinPublicChannel, authMw)
	api.PATCH("/channels/:channelId/members/:userId/role", cfg.ChannelMemberHandler.UpdateMemberRole, authMw)
	api.DELETE("/channels/:channelId/members/:userId", cfg.ChannelMemberHandler.RemoveMember, authMw)
	api.DELETE("/channels/:channelId/members/self", cfg.ChannelMemberHandler.LeaveChannel, authMw)

	// Message routes
	api.GET("/channels/:channelId/messages", cfg.MessageHandler.ListMessages, authMw)
	api.POST("/channels/:channelId/messages", cfg.MessageHandler.CreateMessage, authMw)
	api.PATCH("/messages/:messageId", cfg.MessageHandler.UpdateMessage, authMw)
	api.DELETE("/messages/:messageId", cfg.MessageHandler.DeleteMessage, authMw)

	// Read state routes
	api.GET("/channels/:channelId/unread_count", cfg.ReadStateHandler.GetUnreadCount, authMw)
	api.POST("/channels/:channelId/reads", cfg.ReadStateHandler.UpdateReadState, authMw)

	// Reaction routes
	api.GET("/messages/:messageId/reactions", cfg.ReactionHandler.ListReactions, authMw)
	api.POST("/messages/:messageId/reactions", cfg.ReactionHandler.AddReaction, authMw)
	api.DELETE("/messages/:messageId/reactions/:emoji", cfg.ReactionHandler.RemoveReaction, authMw)

	// User group routes
	groups := api.Group("/user-groups", authMw)
	{
		groups.POST("", cfg.UserGroupHandler.CreateUserGroup)
		groups.GET("", cfg.UserGroupHandler.ListUserGroups)
		groups.GET("/:id", cfg.UserGroupHandler.GetUserGroup)
		groups.PATCH("/:id", cfg.UserGroupHandler.UpdateUserGroup)
		groups.DELETE("/:id", cfg.UserGroupHandler.DeleteUserGroup)
		groups.POST("/:id/members", cfg.UserGroupHandler.AddMember)
		groups.DELETE("/:id/members/:userId", cfg.UserGroupHandler.RemoveMember)
		groups.GET("/:id/members", cfg.UserGroupHandler.ListMembers)
	}

	// Link routes
	links := api.Group("/links", authMw)
	{
		links.POST("/fetch-ogp", cfg.LinkHandler.FetchOGP)
	}

	// Bookmark routes
	api.GET("/bookmarks", cfg.BookmarkHandler.ListBookmarks, authMw)
	api.POST("/messages/:messageId/bookmarks", cfg.BookmarkHandler.AddBookmark, authMw)
	api.DELETE("/messages/:messageId/bookmarks", cfg.BookmarkHandler.RemoveBookmark, authMw)

	// Attachment routes
	att := api.Group("/attachments", authMw)
	{
		att.POST("/presign", cfg.AttachmentHandler.PresignUpload)
		att.GET("/:attachmentId", cfg.AttachmentHandler.GetMetadata)
		att.GET("/:attachmentId/download", cfg.AttachmentHandler.GetDownloadURL)
	}

	return e
}
