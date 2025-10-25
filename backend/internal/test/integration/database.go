package integration

import (
	"testing"

	"github.com/newt239/chat/internal/infrastructure/config"
	"github.com/newt239/chat/internal/infrastructure/db"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestDB は統合テスト用のデータベース接続を管理します
type TestDB struct {
	DB *gorm.DB
}

// NewTestDB は統合テスト用のデータベース接続を作成します
func NewTestDB(t *testing.T) *TestDB {
	// テスト用の設定を読み込み
	cfg, err := config.Load()
	require.NoError(t, err)

	// テスト用のデータベースURLを使用（環境変数で設定可能）
	testDBURL := cfg.Database.URL
	if testDBURL == "" {
		testDBURL = "postgres://postgres:password@localhost:5432/chat_test?sslmode=disable"
	}

	// データベース接続
	database, err := db.InitDB(testDBURL)
	require.NoError(t, err)

	// テスト用のテーブルをクリーンアップ
	cleanupTestDB(t, database)

	return &TestDB{DB: database}
}

// Cleanup はテスト後にデータベースをクリーンアップします
func (tdb *TestDB) Cleanup(t *testing.T) {
	cleanupTestDB(t, tdb.DB)
}

// cleanupTestDB はテスト用のデータベースをクリーンアップします
func cleanupTestDB(t *testing.T, db *gorm.DB) {
	// テスト用のテーブルを削除
	tables := []string{
		"message_reactions",
		"message_user_mentions",
		"message_group_mentions",
		"message_links",
		"messages",
		"read_states",
		"channels",
		"workspace_members",
		"workspaces",
		"user_group_members",
		"user_groups",
		"sessions",
		"users",
	}

	for _, table := range tables {
		if err := db.Exec("DELETE FROM " + table).Error; err != nil {
			t.Logf("Warning: failed to cleanup table %s: %v", table, err)
		}
	}
}

// SetupTestData はテスト用のデータをセットアップします
