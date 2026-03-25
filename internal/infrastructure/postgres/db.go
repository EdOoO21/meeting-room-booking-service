package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, dsn string) (*DB, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("create pg pool: %w", err)
	}

	if pingErr := pool.Ping(ctx); pingErr != nil {
		pool.Close()
		return nil, fmt.Errorf("ping pg: %w", pingErr)
	}

	return &DB{pool: pool}, nil
}

func (db *DB) Close() {
	if db == nil || db.pool == nil {
		return
	}

	db.pool.Close()
}
