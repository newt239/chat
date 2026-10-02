package datamigration

import (
	"context"
	"os"
	"testing"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/channelmember"
	"github.com/newt239/chat/ent/migrate"
	"github.com/newt239/chat/ent/user"
)

// TEST_DATABASE_URL を指定したときだけ実行する
func TestConvertWebhooksToApps(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL が未設定のためスキップします")
	}
	drv, err := entsql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	client := ent.NewClient(ent.Driver(drv))
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()
	if err := client.Schema.Create(ctx, migrate.WithGlobalUniqueID(true), migrate.WithForeignKeys(true)); err != nil {
		t.Fatal(err)
	}

	suffix := uuid.NewString()[:8]
	creator := client.User.Create().SetEmail("creator-" + suffix + "@example.com").SetPasswordHash("x").SetDisplayName("creator").SaveX(ctx)
	bot := client.User.Create().SetEmail("bot-" + suffix + "@example.com").SetPasswordHash("!").SetDisplayName("CI").SetIsBot(true).SaveX(ctx)
	workspaceID := "ws-" + suffix
	client.Workspace.Create().SetID(workspaceID).SetName("ws").SetCreatedBy(creator).SaveX(ctx)
	ch := client.Channel.Create().SetName("general").SetWorkspaceID(workspaceID).SetCreatedBy(creator).SaveX(ctx)

	db := drv.DB()
	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS webhook (
		id uuid PRIMARY KEY, name text NOT NULL, avatar_url text, token_hash text NOT NULL, last_used_at timestamptz,
		created_at timestamptz NOT NULL, updated_at timestamptz NOT NULL, webhook_channel uuid NOT NULL, webhook_created_by uuid NOT NULL, webhook_bot_user uuid NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	webhookID := uuid.New()
	if _, err := db.ExecContext(ctx, `INSERT INTO webhook VALUES ($1, 'CI', NULL, 'hash', NULL, now(), now(), $2, $3, $4)`, webhookID, ch.ID, creator.ID, bot.ID); err != nil {
		t.Fatal(err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := convertWebhooksToApps(ctx, tx); err != nil {
		_ = tx.Rollback()
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	migrated := client.App.GetX(ctx, webhookID)
	if migrated.WorkspaceID != workspaceID || *migrated.TokenHash != "hash" || len(migrated.Permissions) != 1 || migrated.Permissions[0] != "post:joined_channels" {
		t.Fatalf("ID・トークン・権限が引き継がれていません: %+v", migrated)
	}
	if client.App.QueryDefaultChannel(migrated).OnlyIDX(ctx) != ch.ID {
		t.Fatal("元のチャンネルが既定の投稿先になっていません")
	}
	if !client.Channel.QueryMembers(ch).Where(channelmember.HasUserWith(user.ID(bot.ID))).ExistX(ctx) {
		t.Fatal("ボットユーザーがチャンネルに参加していません")
	}
	var exists bool
	if err := db.QueryRowContext(ctx, `SELECT to_regclass('webhook') IS NOT NULL`).Scan(&exists); err != nil || exists {
		t.Fatalf("webhook テーブルが残っています: %v", err)
	}
}
