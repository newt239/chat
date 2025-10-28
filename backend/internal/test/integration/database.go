package integration

import (
	"context"
	"testing"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/internal/infrastructure/config"
	"github.com/newt239/chat/internal/infrastructure/database"
	"github.com/stretchr/testify/require"
)

// TestDB は統合テスト用のデータベース接続を管理します
type TestDB struct {
	Client *ent.Client
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
	client, err := database.InitDB(testDBURL)
	require.NoError(t, err)

	// テスト用のテーブルをクリーンアップ
	cleanupTestDB(t, client)

	return &TestDB{Client: client}
}

// Cleanup はテスト後にデータベースをクリーンアップします
func (tdb *TestDB) Cleanup(t *testing.T) {
	cleanupTestDB(t, tdb.Client)
}

// cleanupTestDB はテスト用のデータベースをクリーンアップします
func cleanupTestDB(t *testing.T, client *ent.Client) {
	ctx := context.Background()

	// テスト用のデータを削除
	if _, err := client.MessageReaction.Delete().Exec(ctx); err != nil {
		t.Logf("Warning: failed to cleanup message_reaction: %v", err)
	}
	if _, err := client.MessageUserMention.Delete().Exec(ctx); err != nil {
		t.Logf("Warning: failed to cleanup message_user_mention: %v", err)
	}
	if _, err := client.MessageGroupMention.Delete().Exec(ctx); err != nil {
		t.Logf("Warning: failed to cleanup message_group_mention: %v", err)
	}
	if _, err := client.MessageLink.Delete().Exec(ctx); err != nil {
		t.Logf("Warning: failed to cleanup message_link: %v", err)
	}
	if _, err := client.Message.Delete().Exec(ctx); err != nil {
		t.Logf("Warning: failed to cleanup message: %v", err)
	}
	if _, err := client.ChannelReadState.Delete().Exec(ctx); err != nil {
		t.Logf("Warning: failed to cleanup channel_read_state: %v", err)
	}
	if _, err := client.Channel.Delete().Exec(ctx); err != nil {
		t.Logf("Warning: failed to cleanup channel: %v", err)
	}
	if _, err := client.WorkspaceMember.Delete().Exec(ctx); err != nil {
		t.Logf("Warning: failed to cleanup workspace_member: %v", err)
	}
	if _, err := client.Workspace.Delete().Exec(ctx); err != nil {
		t.Logf("Warning: failed to cleanup workspace: %v", err)
	}
	if _, err := client.UserGroupMember.Delete().Exec(ctx); err != nil {
		t.Logf("Warning: failed to cleanup user_group_member: %v", err)
	}
	if _, err := client.UserGroup.Delete().Exec(ctx); err != nil {
		t.Logf("Warning: failed to cleanup user_group: %v", err)
	}
	if _, err := client.Session.Delete().Exec(ctx); err != nil {
		t.Logf("Warning: failed to cleanup session: %v", err)
	}
	if _, err := client.User.Delete().Exec(ctx); err != nil {
		t.Logf("Warning: failed to cleanup user: %v", err)
	}
}

// SetupTestData はテスト用のデータをセットアップします
