package http

import (
	"github.com/gin-gonic/gin"

	"github.com/example/chat/internal/interface/http/handler"
)

// RegisterRoutes registers all HTTP routes.
func RegisterRoutes(r *gin.Engine, authHandler *handler.AuthHandler) {
	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", authHandler.Logout)
		}

		api.GET("/workspaces", func(c *gin.Context) { c.Status(501) })
		api.POST("/workspaces", func(c *gin.Context) { c.Status(501) })

		ws := api.Group("/workspaces/:id")
		{
			ws.GET("/channels", func(c *gin.Context) { c.Status(501) })
			ws.POST("/channels", func(c *gin.Context) { c.Status(501) })
		}

		ch := api.Group("/channels/:channelId")
		{
			ch.GET("/messages", func(c *gin.Context) { c.Status(501) })
			ch.POST("/messages", func(c *gin.Context) { c.Status(501) })
			ch.GET("/unread_count", func(c *gin.Context) { c.Status(501) })
			ch.POST("/reads", func(c *gin.Context) { c.Status(501) })
		}

		att := api.Group("/attachments")
		{
			att.POST("/presign", func(c *gin.Context) { c.Status(501) })
			att.GET(":id", func(c *gin.Context) { c.Status(501) })
			att.GET(":id/download", func(c *gin.Context) { c.Status(501) })
		}
	}
}
