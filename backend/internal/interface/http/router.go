package http

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all HTTP routes.
func RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", func(c *gin.Context) { c.Status(501) })
			auth.POST("/login", func(c *gin.Context) { c.Status(501) })
			auth.POST("/refresh", func(c *gin.Context) { c.Status(501) })
			auth.POST("/logout", func(c *gin.Context) { c.Status(501) })
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
