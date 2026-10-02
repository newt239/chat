package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/migrate"
	"github.com/newt239/chat/internal/infrastructure/config"
)

// migrationLockKey はスキーマ移行を 1 レプリカずつ行うための advisory lock のキー
const migrationLockKey = 0x63686174

// NewConnection creates a new ent client connection
func NewConnection(dsn string) (*ent.Client, error) {
	drv, err := entsql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	client := ent.NewClient(ent.Driver(drv))
	return client, nil
}

// InitDB は接続プールを設定し、DB が応答するまで待ってから接続を返します
func InitDB(cfg config.DatabaseConfig) (*ent.Client, *sql.DB, error) {
	drv, err := entsql.Open("postgres", cfg.URL)
	if err != nil {
		return nil, nil, err
	}
	db := drv.DB()
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	const maxRetries = 10
	for i := range maxRetries {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = db.PingContext(ctx)
		cancel()
		if err == nil {
			return ent.NewClient(ent.Driver(drv)), db, nil
		}
		if i < maxRetries-1 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}
	_ = db.Close()
	return nil, nil, fmt.Errorf("database is not reachable: %w", err)
}

// Migrate はスキーマを ent の定義に合わせます。複数のレプリカが同時に起動しても重ならないよう advisory lock を取ります
func Migrate(ctx context.Context, client *ent.Client, db *sql.DB) error {
	return withMigrationLock(ctx, db, func(ctx context.Context) error {
		return client.Schema.Create(ctx, migrate.WithGlobalUniqueID(true), migrate.WithForeignKeys(true))
	})
}

func withMigrationLock(ctx context.Context, db *sql.DB, fn func(context.Context) error) error {
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	if _, err := conn.ExecContext(ctx, "SELECT pg_advisory_lock($1)", migrationLockKey); err != nil {
		return fmt.Errorf("acquire migration lock: %w", err)
	}
	defer func() {
		_, _ = conn.ExecContext(context.Background(), "SELECT pg_advisory_unlock($1)", migrationLockKey)
	}()
	return fn(ctx)
}
