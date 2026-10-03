package main

import (
	"context"
	"log"

	"github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/infrastructure/config"
	"github.com/newt239/chat/internal/infrastructure/database"
	"github.com/newt239/chat/internal/infrastructure/meilisearch"
	"github.com/newt239/chat/internal/infrastructure/repository"
	"github.com/newt239/chat/internal/infrastructure/seed"
	"github.com/newt239/chat/internal/usecase/searchindex"
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
	if err := seed.CreateSeedData(ctx, client); err != nil {
		log.Fatalf("シードデータの投入に失敗しました: %v", err)
	}
	// 古いメッセージの検索インデックスを消し、シードしたメッセージで作り直す
	mentionSvc := service.NewMentionService(repository.NewWorkspaceRepository(client), repository.NewUserRepository(client),
		repository.NewUserGroupRepository(client), repository.NewChannelRepository(client))
	indexer := searchindex.NewIndexer(repository.NewMessageRepository(client),
		meilisearch.NewMessageIndex(cfg.Search.MeilisearchURL, cfg.Search.MeilisearchAPIKey), mentionSvc)
	if _, err := indexer.Prepare(ctx, true); err != nil {
		log.Fatalf("検索インデックスの作り直しに失敗しました: %v", err)
	}
	log.Println("データベースと検索インデックスのリセットとシードが完了しました")
}
