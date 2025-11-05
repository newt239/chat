package main

import (
	"context"
	"fmt"
	"log"

	"github.com/newt239/chat/ent/migrate"
	"github.com/newt239/chat/internal/infrastructure/auth"
	"github.com/newt239/chat/internal/infrastructure/config"
	"github.com/newt239/chat/internal/infrastructure/database"
	"github.com/newt239/chat/internal/infrastructure/seed"
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
	defer client.Close()

	ctx := context.Background()

	// Reset database schema (drop indexes and columns)
	log.Println("Resetting database schema...")
	if err := client.Schema.Create(
		ctx,
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
		migrate.WithGlobalUniqueID(true),
		migrate.WithForeignKeys(true),
	); err != nil {
		log.Fatalf("failed to reset database schema: %v", err)
	}
	log.Println("✅ Database schema reset successfully!")

	// Seed database with initial data
	log.Println("Seeding database with initial data...")
	passwordService := auth.NewPasswordService()
	if err := seed.CreateSeedData(client, passwordService); err != nil {
		log.Fatalf("failed to seed database: %v", err)
	}
	log.Println("✅ Database seed completed successfully!")

	fmt.Println("✅ Database reset and seed completed successfully!")
}
