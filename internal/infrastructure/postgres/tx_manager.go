package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type txContextKey struct{}

type TxManager struct {
	db *DB
}

func NewTxManager(db *DB) *TxManager {
	return &TxManager{db: db}
}

func (m *TxManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.db.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	txCtx := context.WithValue(ctx, txContextKey{}, tx)

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if fnErr := fn(txCtx); fnErr != nil {
		return fnErr
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return fmt.Errorf("commit tx: %w", commitErr)
	}

	return nil
}
