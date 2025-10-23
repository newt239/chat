package http

import (
	"github.com/gin-gonic/gin"

	"github.com/example/chat/internal/infrastructure/auth"
	"github.com/example/chat/internal/interface/http/handler"
	"github.com/example/chat/internal/interface/http/middleware"
)

// RegisterRoutes registers all HTTP routes.
func RegisterRoutes(
	r *gin.Engine,
	jwtService *auth.JWTService,
	authHandler *handler.AuthHandler,
	workspaceHandler *handler.WorkspaceHandler,
	channelHandler *handler.ChannelHandler,
	messageHandler *handler.MessageHandler,
	readStateHandler *handler.ReadStateHandler,
) {
	api := r.Group("/api")
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

		att := api.Group("/attachments")
		{
			att.POST("/presign", func(c *gin.Context) { c.Status(501) })
			att.GET(":id", func(c *gin.Context) { c.Status(501) })
			att.GET(":id/download", func(c *gin.Context) { c.Status(501) })
		}
	}
}
