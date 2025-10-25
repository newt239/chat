package main

import (
	"fmt"
	"log"

	"github.com/newt239/chat/internal/adapter/gateway/persistence"
	"github.com/newt239/chat/internal/infrastructure/auth"
	"github.com/newt239/chat/internal/infrastructure/config"
	"github.com/newt239/chat/internal/infrastructure/db"
	"github.com/newt239/chat/internal/infrastructure/seed"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Initialize database
	database, err := db.InitDB(cfg.Database.URL)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	// Force seed data creation (ignoring existing data)
	fmt.Println("🌱 Manually seeding database...")

	// Initialize repositories
	userRepo := persistence.NewUserRepository(database)
	workspaceRepo := persistence.NewWorkspaceRepository(database)
	channelRepo := persistence.NewChannelRepository(database)
	channelMemberRepo := persistence.NewChannelMemberRepository(database)
	messageRepo := persistence.NewMessageRepository(database)

	// Initialize password service
	passwordService := auth.NewPasswordService()

	// Create seed data using the seed package
	if err := seed.CreateSeedData(database, userRepo, workspaceRepo, channelRepo, channelMemberRepo, messageRepo, passwordService); err != nil {
		log.Fatalf("failed to create seed data: %v", err)
	}

	fmt.Println("✅ Manual seed completed successfully!")
	fmt.Println("   Test accounts created:")
	fmt.Println("   - alice@example.com (password: password123)")
	fmt.Println("   - bob@example.com (password: password123)")
	fmt.Println("   - charlie@example.com (password: password123)")
	fmt.Println("   - diana@example.com (password: password123)")
}
