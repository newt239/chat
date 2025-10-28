package main

import (
	"fmt"
	"log"

	"github.com/newt239/chat/internal/infrastructure/auth"
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

	// DBが空なら自動シード
	if err := seed.AutoSeed(client); err != nil {
		// 空でない等の理由でスキップされる場合もあるため致命にはしない
		log.Printf("自動シード: %v", err)
	}

	// 明示的に投入（必要に応じて）
	if err := seed.CreateSeedData(client, auth.NewPasswordService()); err != nil {
		log.Fatalf("シード投入に失敗しました: %v", err)
	}

	fmt.Println("✅ Seed data created successfully!")
}
