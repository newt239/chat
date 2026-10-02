package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	_ "time/tzdata" // インサイトでクライアントのタイムゾーンを扱うため、tzdata のないイメージでも読み込めるよう埋め込む

	goredis "github.com/redis/go-redis/v9"

	"github.com/newt239/chat/ent/migrate"
	"github.com/newt239/chat/internal/infrastructure/config"
	"github.com/newt239/chat/internal/infrastructure/database"
	"github.com/newt239/chat/internal/infrastructure/database/datamigration"
	"github.com/newt239/chat/internal/infrastructure/logger"
	"github.com/newt239/chat/internal/infrastructure/redis"
	"github.com/newt239/chat/internal/infrastructure/seed"
	"github.com/newt239/chat/internal/registry"
)

// shutdownDrainDelay は readiness probe の失敗が kube-proxy と cloudflared に伝わるまで待つ時間
const shutdownDrainDelay = 5 * time.Second

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

	client, db, err := database.InitDB(cfg.Database)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	ctx := context.Background()
	if err := database.WithMigrationLock(ctx, db, func(ctx context.Context) error {
		if err := client.Schema.Create(
			ctx,
			migrate.WithGlobalUniqueID(true),
			migrate.WithForeignKeys(true),
		); err != nil {
			return err
		}
		return datamigration.Run(ctx, db)
	}); err != nil {
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

	var rdb *goredis.Client
	if cfg.Redis.URL != "" {
		if rdb, err = redis.NewClient(cfg.Redis.URL); err != nil {
			log.Fatalf("failed to connect to redis: %v", err)
		}
	} else {
		log.Println("REDIS_URL is not set: WebSocket events and rate limits are not shared between replicas")
	}

	reg := registry.NewRegistry(client, cfg, rdb)
	go prepareSearchIndex(reg)

	runCtx, stopRun := context.WithCancel(context.Background())
	go reg.UseCase().NewScheduledMessageUseCase().RunDispatcher(runCtx, cfg.ScheduledMessage.DispatchInterval)
	// リマインダーも予約メッセージと同じ間隔で確かめる
	go reg.UseCase().NewCommandUseCase().RunDispatcher(runCtx, cfg.ScheduledMessage.DispatchInterval)

	hub := reg.NewWebSocketHub()
	go hub.Run(runCtx)

	e := reg.NewRouter()

	addr := ":" + cfg.Server.Port
	log.Printf("Starting server on %s", addr)

	go func() {
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	sigCtx, stopSignal := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	<-sigCtx.Done()
	stopSignal()

	log.Println("Shutting down server...")

	// readiness を落としてから Service の宛先から外れるまで待ち、新しい接続を他のレプリカへ向ける
	reg.Infrastructure().SetReady(false)
	time.Sleep(shutdownDrainDelay)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	hub.Shutdown(shutdownCtx)
	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
	stopRun()
	if rdb != nil {
		_ = rdb.Close()
	}
	_ = client.Close()

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
