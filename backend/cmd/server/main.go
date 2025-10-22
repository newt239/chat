package main

import (
	"log"
	"net/http"
	"os"

	ginhttp "github.com/example/chat/backend/internal/interface/http"
	"github.com/gin-gonic/gin"
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func setupRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// healthz
	r.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	ginhttp.RegisterRoutes(r)

	return r
}

func main() {
	port := getenv("PORT", "8080")
	r := setupRouter()
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
