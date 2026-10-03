package main

import (
	"context"
	"flag"
	"log"

	"github.com/newt239/chat/internal/infrastructure/config"
	"github.com/newt239/chat/internal/infrastructure/database"
	"github.com/newt239/chat/internal/infrastructure/seed"
)

// 性能検証用に大量のメッセージを投入する。初期データはサーバーの起動時に作られる
func main() {
	messages := flag.Int("messages", 1000, "各チャンネルへ追加するメッセージ数")
	channels := flag.Int("channels", 0, "追加するチャンネル数")
	flag.Parse()

	client, _, err := database.InitDB(config.Load().Database)
	if err != nil {
		log.Fatalf("DB初期化に失敗しました: %v", err)
	}
	defer func() { _ = client.Close() }()

	if err := seed.BulkMessages(context.Background(), client, *channels, *messages); err != nil {
		log.Fatalf("大量データの投入に失敗しました: %v", err)
	}
	log.Printf("各チャンネルに %d 件のメッセージを追加しました", *messages)
}
