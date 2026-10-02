package datamigration

import (
	"context"
	"database/sql"
)

// convertWebhooksToApps はチャンネル単位の着信 Webhook を、そのチャンネルにだけ投稿できるアプリに移します。
// URL を変えずに使い続けられるよう ID・トークン・ボットユーザーを引き継ぎ、ボットユーザーをチャンネルに参加させる
func convertWebhooksToApps(ctx context.Context, tx *sql.Tx) error {
	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT to_regclass('webhook') IS NOT NULL`).Scan(&exists); err != nil || !exists {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO app (id, workspace_id, name, avatar_url, token_hash, permissions, is_official, last_used_at, created_at, updated_at,
			app_created_by, app_bot_user, app_default_channel)
		SELECT w.id, c.channel_workspace, w.name, w.avatar_url, w.token_hash, '["post:joined_channels"]', false, w.last_used_at, w.created_at, w.updated_at,
			w.webhook_created_by, w.webhook_bot_user, w.webhook_channel
		FROM webhook w JOIN channels c ON c.id = w.webhook_channel
		ON CONFLICT (id) DO NOTHING`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO channel_members (id, role, joined_at, channel_member_channel, channel_member_user)
		SELECT gen_random_uuid(), 'member', w.created_at, w.webhook_channel, w.webhook_bot_user
		FROM webhook w
		WHERE NOT EXISTS (
			SELECT 1 FROM channel_members m WHERE m.channel_member_channel = w.webhook_channel AND m.channel_member_user = w.webhook_bot_user
		)`); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `DROP TABLE webhook`)
	return err
}
