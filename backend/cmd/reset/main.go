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
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Initialize database connection
	client, err := database.NewConnection(cfg.Database.URL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Drop all tables and recreate schema
	if err := client.Schema.Create(
		ctx,
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
		migrate.WithGlobalUniqueID(true),
		migrate.WithForeignKeys(true),
	); err != nil {
		log.Fatalf("failed to reset database schema: %v", err)
	}

	fmt.Println("✅ Database reset successfully!")
	fmt.Println("   All data has been cleared.")
	fmt.Println("   Tables have been dropped and recreated.")
}
