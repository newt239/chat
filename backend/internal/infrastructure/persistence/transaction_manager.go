package persistence

import (
	"context"

	"gorm.io/gorm"

	"github.com/newt239/chat/internal/domain/transaction"
)

type transactionManager struct {
	db *gorm.DB
}

func NewTransactionManager(db *gorm.DB) transaction.Manager {
	return &transactionManager{db: db}
}

func (m *transactionManager) Do(ctx context.Context, fn func(context.Context) error) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctxWithTx := contextWithTx(ctx, tx)
		return fn(ctxWithTx)
	})
}
