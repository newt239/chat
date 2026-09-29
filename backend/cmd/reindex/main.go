package main

import (
	"context"
	"fmt"
	"log"

	"github.com/newt239/chat/internal/infrastructure/config"
	"github.com/newt239/chat/internal/infrastructure/database"
	"github.com/newt239/chat/internal/registry"
)

// 削除されていない全メッセージを Meilisearch に登録し直す
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("設定の読み込みに失敗しました: %v", err)
	}

	client, _, err := database.InitDB(cfg.Database)
	if err != nil {
		log.Fatalf("DB初期化に失敗しました: %v", err)
	}
	defer func() { _ = client.Close() }()

	ctx := context.Background()
	reg := registry.NewRegistry(client, cfg)
	if err := reg.Infrastructure().MessageSearchIndex().EnsureSettings(ctx); err != nil {
		log.Fatalf("検索インデックスの設定に失敗しました: %v", err)
	}
	count, err := reg.UseCase().NewSearchIndexer().Reindex(ctx)
	if err != nil {
		log.Fatalf("再インデックスに失敗しました: %v", err)
	}
	fmt.Printf("✅ %d 件のメッセージを登録しました（反映は Meilisearch 側で順次行われます）\n", count)
}
