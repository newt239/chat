package main

import (
	"context"
	"log"

	"github.com/newt239/chat/internal/infrastructure/auth"
	"github.com/newt239/chat/internal/infrastructure/config"
	"github.com/newt239/chat/internal/infrastructure/database"
	"github.com/newt239/chat/internal/infrastructure/seed"
)

func main() {
	cfg := config.Load()
	if cfg.Server.Env == "production" {
		log.Fatal("本番環境ではデータベースをリセットできません")
	}

	client, db, err := database.InitDB(cfg.Database)
	if err != nil {
		log.Fatalf("データベースへの接続に失敗しました: %v", err)
	}
	defer func() { _ = client.Close() }()

	ctx := context.Background()
	if _, err := db.ExecContext(ctx, "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"); err != nil {
		log.Fatalf("テーブルの削除に失敗しました: %v", err)
	}
	if err := database.Migrate(ctx, client, db); err != nil {
		log.Fatalf("データベーススキーマの作成に失敗しました: %v", err)
	}
	if err := seed.CreateSeedData(ctx, client, auth.PasswordService{}); err != nil {
		log.Fatalf("シードデータの投入に失敗しました: %v", err)
	}
	log.Println("データベースのリセットとシードが完了しました")
}
