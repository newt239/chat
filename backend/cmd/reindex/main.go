package main

import (
	"context"
	"log"

	"github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/infrastructure/config"
	"github.com/newt239/chat/internal/infrastructure/database"
	"github.com/newt239/chat/internal/infrastructure/meilisearch"
	"github.com/newt239/chat/internal/infrastructure/repository"
	"github.com/newt239/chat/internal/usecase/searchindex"
)

// 削除されていない全メッセージを Meilisearch に登録し直す
func main() {
	cfg := config.Load()
	client, _, err := database.InitDB(cfg.Database)
	if err != nil {
		log.Fatalf("DB初期化に失敗しました: %v", err)
	}
	defer func() { _ = client.Close() }()

	mentionSvc := service.NewMentionService(repository.NewWorkspaceRepository(client), repository.NewUserRepository(client),
		repository.NewUserGroupRepository(client), repository.NewChannelRepository(client))
	indexer := searchindex.NewIndexer(repository.NewMessageRepository(client),
		meilisearch.NewMessageIndex(cfg.Search.MeilisearchURL, cfg.Search.MeilisearchAPIKey), mentionSvc)
	count, err := indexer.Prepare(context.Background(), true)
	if err != nil {
		log.Fatalf("再インデックスに失敗しました: %v", err)
	}
	log.Printf("%d 件のメッセージを登録しました（反映は Meilisearch 側で順次行われます）", count)
}
