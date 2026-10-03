package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/newt239/chat/internal/infrastructure/config"
	"github.com/newt239/chat/internal/infrastructure/database"
	"github.com/newt239/chat/internal/infrastructure/redis"
	"github.com/newt239/chat/internal/infrastructure/seed"
	"github.com/newt239/chat/internal/registry"
)

// shutdownDrainDelay は readiness probe の失敗が kube-proxy と cloudflared に伝わるまで待つ時間
const shutdownDrainDelay = 5 * time.Second

func fatal(msg string, err error) {
	slog.Error(msg, "error", err)
	os.Exit(1)
}

// every は ctx が終わるまで interval ごとに task を実行します
func every(ctx context.Context, interval time.Duration, name string, task func(context.Context) error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		if err := task(ctx); err != nil && ctx.Err() == nil {
			slog.Error("定期処理に失敗しました", "task", name, "error", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func main() {
	cfg := config.Load()
	if cfg.Server.Env == "production" {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	}
	if err := cfg.Validate(); err != nil {
		fatal("設定が不正です", err)
	}

	client, db, err := database.InitDB(cfg.Database)
	if err != nil {
		fatal("データベースに接続できません", err)
	}
	ctx := context.Background()
	if err := database.Migrate(ctx, client, db); err != nil {
		fatal("スキーマを移行できません", err)
	}
	// 既知のテストアカウントを作るため本番ではシードしない
	if cfg.Server.Env != "production" {
		if err := seed.AutoSeed(ctx, client); err != nil {
			fatal("シードデータを作成できません", err)
		}
	}

	rdb, err := redis.NewClient(cfg.Redis.URL)
	if err != nil {
		fatal("Redis に接続できません", err)
	}

	var ready atomic.Bool
	ready.Store(true)
	app := registry.New(client, cfg, rdb, ready.Load)

	// 検索インデックスの準備に失敗しても検索以外の機能は止めない
	go func() {
		count, err := app.SearchIndexer.Prepare(ctx, false)
		if err != nil {
			slog.Warn("検索インデックスを準備できません", "error", err)
			return
		}
		if count > 0 {
			slog.Info("検索インデックスに登録しました", "messages", count)
		}
	}()

	// 停止時に DB を閉じる前に終わりを待つ
	var background sync.WaitGroup
	runCtx, stopRun := context.WithCancel(ctx)
	interval := cfg.DispatchInterval
	background.Go(func() {
		every(runCtx, interval, "scheduled_message", func(ctx context.Context) error {
			_, err := app.ScheduledMessage.DispatchDue(ctx)
			return err
		})
	})
	background.Go(func() {
		every(runCtx, interval, "reminder", func(ctx context.Context) error {
			_, err := app.Command.DispatchDue(ctx)
			return err
		})
	})
	background.Go(func() { every(runCtx, time.Hour, "session_cleanup", app.Sessions.DeleteExpired) })
	background.Go(func() { app.Hub.Run(runCtx) })

	addr := ":" + cfg.Server.Port
	slog.Info("サーバーを起動します", "addr", addr)
	go func() {
		if err := app.Router.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fatal("サーバーが停止しました", err)
		}
	}()

	sigCtx, stopSignal := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	<-sigCtx.Done()
	stopSignal()
	slog.Info("サーバーを停止します")

	// readiness を落としてから Service の宛先から外れるまで待ち、新しい接続を他のレプリカへ向ける
	ready.Store(false)
	time.Sleep(shutdownDrainDelay)

	shutdownCtx, cancel := context.WithTimeout(ctx, 40*time.Second)
	defer cancel()
	app.Hub.Shutdown(shutdownCtx)
	if err := app.Router.Shutdown(shutdownCtx); err != nil {
		slog.Error("接続を閉じきれずに停止します", "error", err)
	}
	stopRun()
	background.Wait()
	_ = rdb.Close()
	_ = client.Close()
	slog.Info("サーバーを停止しました")
}
