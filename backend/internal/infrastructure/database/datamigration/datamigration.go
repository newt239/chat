// Package datamigration は ent の自動マイグレーションでは扱えない、既存データの書き換えを 1 度だけ実行します
package datamigration

import (
	"context"
	"database/sql"
	"fmt"
)

type migration struct {
	name string
	run  func(ctx context.Context, tx *sql.Tx) error
}

// migrations は追加した順に実行する。名前は実行済みの記録に使うため変えない
var migrations = []migration{
	{name: "mention_id_syntax", run: convertLegacyMentions},
	{name: "webhook_to_app", run: convertWebhooksToApps},
}

// Run は未実行のデータ移行を 1 件ずつトランザクションで実行します。スキーマの移行と同じロックの中で呼びます
func Run(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS data_migration (name text PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
		return fmt.Errorf("create data_migration table: %w", err)
	}
	for _, m := range migrations {
		var applied bool
		if err := db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM data_migration WHERE name = $1)`, m.name).Scan(&applied); err != nil {
			return fmt.Errorf("check data migration %s: %w", m.name, err)
		}
		if applied {
			continue
		}
		if err := runInTx(ctx, db, m); err != nil {
			return fmt.Errorf("data migration %s: %w", m.name, err)
		}
	}
	return nil
}

func runInTx(ctx context.Context, db *sql.DB, m migration) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := m.run(ctx, tx); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO data_migration (name) VALUES ($1)`, m.name); err != nil {
		return err
	}
	return tx.Commit()
}
