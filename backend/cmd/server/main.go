package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/example/chat/internal/adapter/controller/websocket"
	"github.com/example/chat/internal/infrastructure/config"
	"github.com/example/chat/internal/infrastructure/db"
	"github.com/example/chat/internal/infrastructure/logger"
	"github.com/example/chat/internal/infrastructure/seed"
	"github.com/example/chat/internal/registry"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Initialize logger
	if err := logger.Init(cfg.Server.Env); err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Initialize database
	db, err := db.InitDB(cfg.Database.URL)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	// Auto-seed database if empty
	if err := seed.AutoSeed(db); err != nil {
		log.Fatalf("failed to auto-seed database: %v", err)
	}

	// Initialize registry (DI container)
	reg := registry.NewRegistry(db, cfg)

	// Initialize WebSocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// Setup Echo router
	e := reg.NewRouter()

	// WebSocket endpoint
	jwtService := reg.NewJWTService()
	e.GET("/ws", websocket.NewHandler(hub, jwtService))

	// Start server
	addr := ":" + cfg.Server.Port
	log.Printf("Starting server on %s", addr)

	// Graceful shutdown
	go func() {
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
