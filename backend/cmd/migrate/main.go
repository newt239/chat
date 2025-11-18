package main

import (
	"context"
	"fmt"
	"log"

	"github.com/newt239/chat/ent/migrate"
	"github.com/newt239/chat/internal/infrastructure/config"
	"github.com/newt239/chat/internal/infrastructure/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	client, err := database.NewConnection(cfg.Database.URL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("failed to close database connection: %v", err)
		}
	}()

	ctx := context.Background()

	// 自動マイグレーション（既存のテーブルは保持）
	if err := client.Schema.Create(
		ctx,
		migrate.WithGlobalUniqueID(true),
		migrate.WithForeignKeys(true),
	); err != nil {
		log.Fatalf("failed to migrate database schema: %v", err)
	}

	fmt.Println("✅ Database migration completed successfully!")
}
