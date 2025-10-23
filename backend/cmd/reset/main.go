package main

import (
	"fmt"
	"log"

	"github.com/example/chat/internal/infrastructure/config"
	infradb "github.com/example/chat/internal/infrastructure/db"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Initialize database connection
	db, err := infradb.NewConnection(cfg.Database.URL, logger.Info)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Drop all tables
	if err := dropAllTables(db); err != nil {
		log.Fatalf("failed to drop tables: %v", err)
	}

	// Recreate tables
	if err := infradb.AutoMigrate(db); err != nil {
		log.Fatalf("failed to recreate tables: %v", err)
	}

	fmt.Println("✅ Database reset successfully!")
	fmt.Println("   All data has been cleared and tables recreated.")
	fmt.Println("   Run 'go run cmd/seed/main.go' to populate with seed data.")
}

func dropAllTables(db *gorm.DB) error {
	tables := []string{
		"attachments",
		"channel_read_states",
		"message_reactions",
		"messages",
		"channel_members",
		"channels",
		"workspace_members",
		"workspaces",
		"sessions",
		"users",
	}

	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)).Error; err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
		fmt.Printf("✅ Dropped table: %s\n", table)
	}

	return nil
}
