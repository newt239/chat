package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/newt239/chat/ent/migrate"
	"github.com/newt239/chat/internal/infrastructure/auth"
	"github.com/newt239/chat/internal/infrastructure/config"
	"github.com/newt239/chat/internal/infrastructure/database"
	"github.com/newt239/chat/internal/infrastructure/seed"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("設定の読み込みに失敗しました: %v", err)
	}

	client, err := database.NewConnection(cfg.Database.URL)
	if err != nil {
		log.Fatalf("データベースへの接続に失敗しました: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("データベース接続のクローズに失敗しました: %v", err)
		}
	}()

	ctx := context.Background()

	// Delete all existing tables
	log.Println("既存のテーブルを削除しています...")
	if err := deleteAllTables(cfg.Database.URL); err != nil {
		log.Fatalf("テーブルの削除に失敗しました: %v", err)
	}
	log.Println("✅ すべてのテーブルを削除しました!")

	// Reset database schema (drop indexes and columns)
	log.Println("データベーススキーマをリセットしています...")
	if err := client.Schema.Create(
		ctx,
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
		migrate.WithGlobalUniqueID(true),
		migrate.WithForeignKeys(true),
	); err != nil {
		log.Fatalf("データベーススキーマのリセットに失敗しました: %v", err)
	}
	log.Println("✅ データベーススキーマのリセットが完了しました!")

	// Seed database with initial data
	log.Println("初期データを投入しています...")
	passwordService := auth.NewPasswordService()
	if err := seed.CreateSeedData(client, passwordService); err != nil {
		log.Fatalf("データベースへのシードデータ投入に失敗しました: %v", err)
	}
	log.Println("✅ データベースへのシードデータ投入が完了しました!")

	fmt.Println("✅ データベースのリセットとシードが正常に完了しました!")
}

func deleteAllTables(dsn string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("データベース接続を開けませんでした: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("データベース接続のクローズに失敗しました: %v", err)
		}
	}()

	// Get all table names from the current schema
	rows, err := db.Query(`
		SELECT tablename 
		FROM pg_tables 
		WHERE schemaname = 'public'
	`)
	if err != nil {
		return fmt.Errorf("テーブル名の取得に失敗しました: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("rowsのクローズに失敗しました: %v", err)
		}
	}()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("テーブル名のスキャンに失敗しました: %w", err)
		}
		tables = append(tables, tableName)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("テーブル名の取得中にエラーが発生しました: %w", err)
	}

	if len(tables) == 0 {
		return nil
	}

	// Build DROP TABLE statement for all tables
	dropSQL := "DROP TABLE IF EXISTS "
	for i, table := range tables {
		if i > 0 {
			dropSQL += ", "
		}
		dropSQL += fmt.Sprintf(`"%s"`, table)
	}
	dropSQL += " CASCADE"

	if _, err := db.Exec(dropSQL); err != nil {
		return fmt.Errorf("テーブルの削除に失敗しました: %w", err)
	}

	return nil
}
