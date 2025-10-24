package integration

import (
	"testing"

	"github.com/example/chat/internal/infrastructure/config"
	infradb "github.com/example/chat/internal/infrastructure/db"
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
	db, err := infradb.InitDB(testDBURL)
	require.NoError(t, err)

	// テスト用のテーブルをクリーンアップ
	cleanupTestDB(t, db)

	return &TestDB{DB: db}
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
func (tdb *TestDB) SetupTestData(t *testing.T) {
	// テスト用のユーザーを作成
	users := []struct {
		ID          string
		Email       string
		DisplayName string
	}{
		{"11111111-1111-1111-1111-111111111111", "alice@example.com", "Alice Johnson"},
		{"22222222-2222-2222-2222-222222222222", "bob@example.com", "Bob Smith"},
		{"33333333-3333-3333-3333-333333333333", "charlie@example.com", "Charlie Brown"},
	}

	for _, user := range users {
		err := tdb.DB.Exec(`
			INSERT INTO users (id, email, password_hash, display_name, created_at, updated_at)
			VALUES (?, ?, ?, ?, NOW(), NOW())
		`, user.ID, user.Email, "hashed_password", user.DisplayName).Error
		require.NoError(t, err)
	}

	// テスト用のワークスペースを作成
	workspaceID := "workspace-11111111-1111-1111-1111-111111111111"
	err := tdb.DB.Exec(`
		INSERT INTO workspaces (id, name, description, created_at, updated_at)
		VALUES (?, ?, ?, NOW(), NOW())
	`, workspaceID, "Test Workspace", "Test Description").Error
	require.NoError(t, err)

	// ワークスペースメンバーを追加
	err = tdb.DB.Exec(`
		INSERT INTO workspace_members (workspace_id, user_id, role, created_at, updated_at)
		VALUES (?, ?, ?, NOW(), NOW())
	`, workspaceID, "11111111-1111-1111-1111-111111111111", "admin").Error
	require.NoError(t, err)

	// テスト用のチャンネルを作成
	channelID := "channel-11111111-1111-1111-1111-111111111111"
	err = tdb.DB.Exec(`
		INSERT INTO channels (id, workspace_id, name, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, NOW(), NOW())
	`, channelID, workspaceID, "general", "General channel").Error
	require.NoError(t, err)
}
