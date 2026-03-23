package ports

import "context"

// TxManager выполняет use case'ы внутри транзакции.
type TxManager interface {
	// WithinTransaction запускает функцию в транзакционном контексте.
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
