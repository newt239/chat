package transaction

import (
	"context"
	"errors"

	"github.com/newt239/chat/ent"
)

type txContextKey struct{}

// ResolveClient はトランザクション中ならそのクライアントを返します
func ResolveClient(ctx context.Context, client *ent.Client) *ent.Client {
	if tx, ok := ctx.Value(txContextKey{}).(*ent.Tx); ok {
		return tx.Client()
	}
	return client
}

// Manager は既にトランザクション中なら入れ子にせず、そのトランザクションで fn を実行します
type Manager struct {
	client *ent.Client
}

func NewManager(client *ent.Client) *Manager {
	return &Manager{client: client}
}

func (m *Manager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := ctx.Value(txContextKey{}).(*ent.Tx); ok {
		return fn(ctx)
	}
	tx, err := m.client.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if v := recover(); v != nil {
			_ = tx.Rollback()
			panic(v)
		}
	}()
	if err := fn(context.WithValue(ctx, txContextKey{}, tx)); err != nil {
		return errors.Join(err, tx.Rollback())
	}
	return tx.Commit()
}

// WithTx は Manager.Do をリポジトリから使うためのもので、トランザクションのクライアントを fn に渡します
func WithTx(ctx context.Context, client *ent.Client, fn func(*ent.Client) error) error {
	return NewManager(client).Do(ctx, func(ctx context.Context) error {
		return fn(ResolveClient(ctx, client))
	})
}
