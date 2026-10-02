package main

import (
	"context"
	"fmt"
	"log"

	"github.com/newt239/chat/ent/migrate"
	"github.com/newt239/chat/internal/infrastructure/config"
	"github.com/newt239/chat/internal/infrastructure/database"
	"github.com/newt239/chat/internal/infrastructure/database/datamigration"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	client, db, err := database.InitDB(cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("failed to close database connection: %v", err)
		}
	}()

	ctx := context.Background()

	// 自動マイグレーション（既存のテーブルは保持）のあと、既存データの書き換えを行う
	if err := database.WithMigrationLock(ctx, db, func(ctx context.Context) error {
		if err := client.Schema.Create(
			ctx,
			migrate.WithGlobalUniqueID(true),
			migrate.WithForeignKeys(true),
		); err != nil {
			return err
		}
		return datamigration.Run(ctx, db)
	}); err != nil {
		log.Fatalf("failed to migrate database schema: %v", err)
	}

	fmt.Println("✅ Database migration completed successfully!")
}
