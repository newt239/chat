package transaction

import (
	"context"
	"errors"

	"github.com/newt239/chat/ent"
)

type txContextKey struct{}

func contextWithTx(ctx context.Context, tx *ent.Tx) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}

func txFromContext(ctx context.Context) (*ent.Tx, bool) {
	tx, ok := ctx.Value(txContextKey{}).(*ent.Tx)
	return tx, ok && tx != nil
}

// ResolveClient returns the transaction client if in a transaction context, otherwise returns the regular client
func ResolveClient(ctx context.Context, client *ent.Client) *ent.Client {
	if tx, ok := txFromContext(ctx); ok {
		return tx.Client()
	}
	return client
}

// WithTx はトランザクション中ならそのまま、そうでなければ新しいトランザクションで fn を実行します
func WithTx(ctx context.Context, client *ent.Client, fn func(*ent.Client) error) error {
	if tx, ok := txFromContext(ctx); ok {
		return fn(tx.Client())
	}
	tx, err := client.Tx(ctx)
	if err != nil {
		return err
	}
	if err := fn(tx.Client()); err != nil {
		return errors.Join(err, tx.Rollback())
	}
	return tx.Commit()
}
