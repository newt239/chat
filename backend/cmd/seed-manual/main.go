package main

import (
	"fmt"
	"log"

	"github.com/example/chat/internal/adapter/gateway/persistence"
	"github.com/example/chat/internal/infrastructure/auth"
	"github.com/example/chat/internal/infrastructure/config"
	infradb "github.com/example/chat/internal/infrastructure/db"
	"github.com/example/chat/internal/infrastructure/seed"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Initialize database
	db, err := infradb.InitDB(cfg.Database.URL)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	// Force seed data creation (ignoring existing data)
	fmt.Println("🌱 Manually seeding database...")

	// Initialize repositories
	userRepo := persistence.NewUserRepository(db)
	workspaceRepo := persistence.NewWorkspaceRepository(db)
	channelRepo := persistence.NewChannelRepository(db)
	messageRepo := persistence.NewMessageRepository(db)

	// Initialize password service
	passwordService := auth.NewPasswordService()

	// Create seed data using the seed package
	if err := seed.CreateSeedData(db, userRepo, workspaceRepo, channelRepo, messageRepo, passwordService); err != nil {
		log.Fatalf("failed to create seed data: %v", err)
	}

	fmt.Println("✅ Manual seed completed successfully!")
	fmt.Println("   Test accounts created:")
	fmt.Println("   - alice@example.com (password: password123)")
	fmt.Println("   - bob@example.com (password: password123)")
	fmt.Println("   - charlie@example.com (password: password123)")
	fmt.Println("   - diana@example.com (password: password123)")
}
