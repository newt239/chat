package http

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/example/chat/internal/interface/http/handler"
	"github.com/example/chat/internal/interface/http/middleware"
	authuc "github.com/example/chat/internal/usecase/auth"
)

// RegisterRoutes registers all HTTP routes.
func RegisterRoutes(
	e *echo.Echo,
	jwtService authuc.JWTService,
	authHandler *handler.AuthHandler,
	workspaceHandler *handler.WorkspaceHandler,
	channelHandler *handler.ChannelHandler,
	messageHandler *handler.MessageHandler,
	readStateHandler *handler.ReadStateHandler,
	reactionHandler *handler.ReactionHandler,
	userGroupHandler *handler.UserGroupHandler,
	linkHandler *handler.LinkHandler,
) {
	api := e.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", authHandler.Logout)
		}

		// Protected workspace routes
		authMiddleware := middleware.AuthMiddleware(jwtService)
		protected := api.Group("")
		protected.Use(authMiddleware)

		protected.GET("/workspaces", workspaceHandler.GetWorkspaces)
		protected.POST("/workspaces", workspaceHandler.CreateWorkspace)
		protected.GET("/workspaces/:id", workspaceHandler.GetWorkspace)
		protected.PATCH("/workspaces/:id", workspaceHandler.UpdateWorkspace)
		protected.DELETE("/workspaces/:id", workspaceHandler.DeleteWorkspace)
		protected.GET("/workspaces/:id/members", workspaceHandler.ListMembers)
		protected.POST("/workspaces/:id/members", workspaceHandler.AddMember)
		protected.PATCH("/workspaces/:id/members/:userId", workspaceHandler.UpdateMemberRole)
		protected.DELETE("/workspaces/:id/members/:userId", workspaceHandler.RemoveMember)

		ws := protected.Group("/workspaces/:id")
		{
			ws.GET("/channels", channelHandler.ListChannels)
			ws.POST("/channels", channelHandler.CreateChannel)
		}

		ch := protected.Group("/channels/:channelId")
		{
			ch.GET("/messages", messageHandler.ListMessages)
			ch.POST("/messages", messageHandler.CreateMessage)
			ch.GET("/unread_count", readStateHandler.GetUnreadCount)
			ch.POST("/reads", readStateHandler.UpdateReadState)
		}

		msg := protected.Group("/messages/:messageId")
		{
			msg.GET("/reactions", reactionHandler.ListReactions)
			msg.POST("/reactions", reactionHandler.AddReaction)
			msg.DELETE("/reactions/:emoji", reactionHandler.RemoveReaction)
		}

		// User group routes
		groups := protected.Group("/user-groups")
		{
			groups.POST("", userGroupHandler.CreateUserGroup)
			groups.GET("", userGroupHandler.ListUserGroups)
			groups.GET("/:id", userGroupHandler.GetUserGroup)
			groups.PATCH("/:id", userGroupHandler.UpdateUserGroup)
			groups.DELETE("/:id", userGroupHandler.DeleteUserGroup)
			groups.POST("/:id/members", userGroupHandler.AddMember)
			groups.DELETE("/:id/members", userGroupHandler.RemoveMember)
			groups.GET("/:id/members", userGroupHandler.ListMembers)
		}

		// Link routes
		links := protected.Group("/links")
		{
			links.POST("/fetch-ogp", linkHandler.FetchOGP)
		}

		att := api.Group("/attachments")
		{
			att.POST("/presign", func(c echo.Context) error { return c.NoContent(http.StatusNotImplemented) })
			att.GET("/:id", func(c echo.Context) error { return c.NoContent(http.StatusNotImplemented) })
			att.GET("/:id/download", func(c echo.Context) error { return c.NoContent(http.StatusNotImplemented) })
		}
	}
}
