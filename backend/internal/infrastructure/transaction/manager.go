package transaction

import (
	"context"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/internal/domain/transaction"
)

type transactionManager struct {
	client *ent.Client
}

func NewTransactionManager(client *ent.Client) transaction.Manager {
	return &transactionManager{client: client}
}

func (m *transactionManager) Do(ctx context.Context, fn func(context.Context) error) error {
	tx, err := m.client.Tx(ctx)
	if err != nil {
		return err
	}

	ctxWithTx := contextWithTx(ctx, tx)

	defer func() {
		if v := recover(); v != nil {
			if err := tx.Rollback(); err != nil {
				_ = err // ロールバックエラーは無視（panic中なので）
			}
			panic(v)
		}
	}()

	if err := fn(ctxWithTx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			return rerr
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
