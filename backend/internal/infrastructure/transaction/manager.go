package transaction

import (
	"context"
	"errors"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/internal/domain/transaction"
)

type transactionManager struct {
	client *ent.Client
}

func NewTransactionManager(client *ent.Client) transaction.Manager {
	return &transactionManager{client: client}
}

// Do は既にトランザクション中なら入れ子にせずそのまま fn を実行します
func (m *transactionManager) Do(ctx context.Context, fn func(context.Context) error) error {
	if _, ok := txFromContext(ctx); ok {
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

	if err := fn(contextWithTx(ctx, tx)); err != nil {
		// ロールバックに失敗しても元のエラーを残す
		return errors.Join(err, tx.Rollback())
	}
	return tx.Commit()
}
