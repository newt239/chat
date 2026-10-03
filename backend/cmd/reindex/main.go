package main

import (
	"context"
	"log"

	"github.com/newt239/chat/internal/infrastructure/config"
	"github.com/newt239/chat/internal/infrastructure/database"
	"github.com/newt239/chat/internal/registry"
)

// 削除されていない全メッセージを Meilisearch に登録し直す
func main() {
	cfg := config.Load()
	client, _, err := database.InitDB(cfg.Database)
	if err != nil {
		log.Fatalf("DB初期化に失敗しました: %v", err)
	}
	defer func() { _ = client.Close() }()

	indexer := registry.NewSearchIndexer(client, cfg.Search)
	count, err := indexer.Prepare(context.Background(), true)
	if err != nil {
		log.Fatalf("再インデックスに失敗しました: %v", err)
	}
	log.Printf("%d 件のメッセージを登録しました（反映は Meilisearch 側で順次行われます）", count)
}
