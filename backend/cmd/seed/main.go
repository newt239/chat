package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/newt239/chat/internal/infrastructure/config"
	"github.com/newt239/chat/internal/infrastructure/database"
	"github.com/newt239/chat/internal/infrastructure/seed"
)

func main() {
	messages := flag.Int("messages", 0, "性能検証用に各チャンネルへ追加するメッセージ数")
	channels := flag.Int("channels", 0, "性能検証用に追加するチャンネル数（-messages と併用）")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("設定の読み込みに失敗しました: %v", err)
	}

	client, _, err := database.InitDB(cfg.Database)
	if err != nil {
		log.Fatalf("DB初期化に失敗しました: %v", err)
	}

	// DBが空なら自動シード（空でない場合はスキップ）
	if err := seed.AutoSeed(client); err != nil {
		log.Printf("自動シード: %v", err)
	}

	if *messages > 0 {
		if err := seed.BulkMessages(context.Background(), client, *channels, *messages); err != nil {
			log.Fatalf("大量データの投入に失敗しました: %v", err)
		}
		fmt.Printf("✅ 各チャンネルに %d 件のメッセージを追加しました\n", *messages)
	}

	fmt.Println("✅ Seed process finished!")
}
