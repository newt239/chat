package transaction

import (
	"context"

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
