package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
	_ "time/tzdata" // インサイトでクライアントのタイムゾーンを扱うため、tzdata のないイメージでも読み込めるよう埋め込む

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

	// 既知のテストアカウントを作るため本番ではシードしない
	if cfg.Server.Env == "production" {
		log.Println("Production environment: skipping auto-seed")
	} else if err := seed.AutoSeed(client); err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			log.Fatalf("database tables do not exist after migration. This indicates a migration failure: %v", err)
		}
		log.Fatalf("failed to auto-seed database: %v", err)
	}

	reg := registry.NewRegistry(client, cfg)
	go prepareSearchIndex(reg)
	go reg.UseCase().NewScheduledMessageUseCase().RunDispatcher(ctx, 10*time.Second)

	hub := reg.NewWebSocketHub()
	go hub.Run()

	e := reg.NewRouter()

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

// prepareSearchIndex は検索インデックスの設定を反映し、空なら全件を登録します。検索以外の機能は止めない
func prepareSearchIndex(reg *registry.Registry) {
	ctx := context.Background()
	index := reg.Infrastructure().MessageSearchIndex()
	if err := index.EnsureSettings(ctx); err != nil {
		log.Printf("Warning: failed to configure the search index: %v", err)
		return
	}
	empty, err := index.IsEmpty(ctx)
	if err != nil || !empty {
		return
	}
	count, err := reg.UseCase().NewSearchIndexer().Reindex(ctx)
	if err != nil {
		log.Printf("Warning: failed to build the search index: %v", err)
		return
	}
	log.Printf("Indexed %d messages for search", count)
}
