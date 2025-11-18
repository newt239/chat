package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	oapimw "github.com/oapi-codegen/echo-middleware"

	"github.com/newt239/chat/ent/migrate"
	"github.com/newt239/chat/internal/infrastructure/config"
	"github.com/newt239/chat/internal/infrastructure/database"
	"github.com/newt239/chat/internal/infrastructure/logger"
	"github.com/newt239/chat/internal/infrastructure/seed"
	"github.com/newt239/chat/internal/registry"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("config validation failed: %v", err)
	}

	// Initialize logger
	if err := logger.Init(cfg.Server.Env); err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Initialize database
	log.Println("Initializing database connection...")
	client, err := database.InitDB(cfg.Database.URL)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	log.Println("✅ Database connection established")

	// Run migration (schema changes are detected and applied automatically)
	log.Println("Running database migration...")
	ctx := context.Background()
	if err := client.Schema.Create(
		ctx,
		migrate.WithGlobalUniqueID(true),
		migrate.WithForeignKeys(true),
	); err != nil {
		log.Fatalf("failed to migrate database schema: %v", err)
	}
	log.Println("✅ Database migration completed successfully!")

	// Verify migration by checking if users table exists
	log.Println("Verifying migration...")
	if _, err := client.User.Query().Limit(1).All(ctx); err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			log.Fatalf("migration verification failed: users table does not exist after migration. This indicates the migration did not create the tables. Error: %v", err)
		}
		log.Printf("Warning: could not verify migration (non-fatal): %v", err)
	} else {
		log.Println("✅ Migration verified: users table exists")
	}

	// Auto-seed database if empty
	// Note: AutoSeed will skip if database already contains data
	if err := seed.AutoSeed(client); err != nil {
		// Check if error is due to missing tables (should not happen after migration)
		if strings.Contains(err.Error(), "does not exist") {
			log.Fatalf("database tables do not exist after migration. This indicates a migration failure: %v", err)
		}
		log.Fatalf("failed to auto-seed database: %v", err)
	}

	// Initialize registry (DI container)
	reg := registry.NewRegistry(client, cfg)

	// Initialize WebSocket hub
	hub := reg.NewWebSocketHub()
	go hub.Run()

	// Setup Echo router
	e := reg.NewRouter()

	// OpenAPI 実行時バリデーション（リクエストのみ）
	if err := setupOpenAPIMiddleware(e); err != nil {
		log.Fatalf("failed to setup OpenAPI middleware: %v", err)
	}

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

func setupOpenAPIMiddleware(e *echo.Echo) error {
	specPath := os.Getenv("OPENAPI_SPEC_PATH")
	if specPath == "" {
		specPath = "openapi/openapi.yaml"
	}
	loader := &openapi3.Loader{IsExternalRefsAllowed: true}
	doc, err := loader.LoadFromFile(specPath)
	if err != nil {
		return err
	}
	if err := doc.Validate(loader.Context); err != nil {
		return err
	}
	e.Use(oapimw.OapiRequestValidatorWithOptions(doc, &oapimw.Options{}))
	return nil
}
