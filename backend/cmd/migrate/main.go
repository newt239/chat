package main

import (
	"context"
	"fmt"
	"log"

	"github.com/newt239/chat/internal/infrastructure/config"
	"github.com/newt239/chat/internal/infrastructure/database"
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

	if err := database.Migrate(context.Background(), client, db); err != nil {
		log.Fatalf("failed to migrate database schema: %v", err)
	}

	fmt.Println("✅ Database migration completed successfully!")
}
