package main

import (
	"fmt"
	"log"

	"github.com/newt239/chat/internal/infrastructure/config"
	"github.com/newt239/chat/internal/infrastructure/database"
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
	database, err := database.NewConnection(cfg.Database.URL, logger.Info)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Drop all tables
	if err := dropAllTables(database); err != nil {
		log.Fatalf("failed to drop tables: %v", err)
	}

	fmt.Println("✅ Database reset successfully!")
	fmt.Println("   All data has been cleared.")
	fmt.Println("   Run your migration tool (e.g. `atlas migrate apply`) to recreate tables.")
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
