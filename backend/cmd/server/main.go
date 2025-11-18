package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
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
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("config validation failed: %v", err)
	}

	if err := logger.Init(cfg.Server.Env); err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	client, err := database.InitDB(cfg.Database.URL)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	ctx := context.Background()
	if err := client.Schema.Create(
		ctx,
		migrate.WithGlobalUniqueID(true),
		migrate.WithForeignKeys(true),
	); err != nil {
		log.Fatalf("failed to migrate database schema: %v", err)
	}

	if _, err := client.User.Query().Limit(1).All(ctx); err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			log.Fatalf("migration verification failed: users table does not exist after migration. This indicates the migration did not create the tables. Error: %v", err)
		}
		log.Printf("Warning: could not verify migration (non-fatal): %v", err)
	}

	if err := seed.AutoSeed(client); err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			log.Fatalf("database tables do not exist after migration. This indicates a migration failure: %v", err)
		}
		log.Fatalf("failed to auto-seed database: %v", err)
	}

	reg := registry.NewRegistry(client, cfg)

	hub := reg.NewWebSocketHub()
	go hub.Run()

	e := reg.NewRouter()

	if err := setupOpenAPIMiddleware(e); err != nil {
		log.Fatalf("failed to setup OpenAPI middleware: %v", err)
	}

	addr := ":" + cfg.Server.Port
	log.Printf("Starting server on %s", addr)

	go func() {
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}

func setupOpenAPIMiddleware(e *echo.Echo) error {
	specPath := "/app/openapi/openapi.yaml"
	loader := &openapi3.Loader{IsExternalRefsAllowed: true}
	doc, err := loader.LoadFromFile(specPath)
	if err != nil {
		return err
	}
	if err := doc.Validate(loader.Context); err != nil {
		return err
	}

	// OpenAPIバリデーションミドルウェアを作成
	validator := oapimw.OapiRequestValidatorWithOptions(doc, &oapimw.Options{
		Options: openapi3filter.Options{
			AuthenticationFunc: authenticateBearerToken,
		},
	})

	// WebSocketエンドポイントをスキップするカスタムミドルウェア
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// WebSocketエンドポイントはバリデーションをスキップ
			if c.Request().URL.Path == "/ws" {
				return next(c)
			}
			// その他のリクエストはOpenAPIバリデーションを適用
			return validator(next)(c)
		}
	})
	return nil
}

// authenticateBearerToken はBearer認証トークンの存在を確認します
// 実際のトークン検証はcustommw.Authミドルウェアで行われます
func authenticateBearerToken(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
	if input.SecuritySchemeName != "bearerAuth" {
		return input.NewError(openapi3filter.ErrAuthenticationServiceMissing)
	}

	req := input.RequestValidationInput.Request
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		return input.NewError(errors.New("authorization header is required"))
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return input.NewError(errors.New("authorization header must be Bearer token"))
	}

	return nil
}
