package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

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
	fmt.Println("   Tables have been dropped.")

	// Reset Atlas migration state and apply migrations to recreate tables
	fmt.Println("🔄 Resetting Atlas migration state...")
	if err := resetAtlasState(); err != nil {
		log.Fatalf("failed to reset Atlas state: %v", err)
	}

	fmt.Println("🔄 Applying migrations to recreate tables...")
	if err := applyMigrations(); err != nil {
		log.Fatalf("failed to apply migrations: %v", err)
	}

	fmt.Println("✅ Migrations applied successfully!")
	fmt.Println("   Tables have been recreated and are ready for use.")
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
		// Note: atlas_schema_revisions is intentionally not dropped to preserve migration history
	}

	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)).Error; err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
		fmt.Printf("✅ Dropped table: %s\n", table)
	}

	return nil
}

func resetAtlasState() error {
	// Reset Atlas migration state to force re-application of all migrations
	cmd := exec.Command("atlas", "migrate", "set", "--env", "docker", "0")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func applyMigrations() error {
	cmd := exec.Command("atlas", "migrate", "apply", "--env", "docker")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
