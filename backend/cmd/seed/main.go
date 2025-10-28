package main

import (
	"fmt"
	"log"

	"github.com/newt239/chat/internal/infrastructure/config"
	"github.com/newt239/chat/internal/infrastructure/database"
	"github.com/newt239/chat/internal/infrastructure/seed"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("設定の読み込みに失敗しました: %v", err)
	}

	client, err := database.InitDB(cfg.Database.URL)
	if err != nil {
		log.Fatalf("DB初期化に失敗しました: %v", err)
	}

	// DBが空なら自動シード（空でない場合はスキップ）
	if err := seed.AutoSeed(client); err != nil {
		log.Printf("自動シード: %v", err)
	}

	// 明示的投入は重複の原因となるためデフォルトでは実行しない
	// 必要なら手動で AutoSeed を実行せずに CreateSeedData を呼ぶ専用コマンドを用意してください

	fmt.Println("✅ Seed process finished!")
}
