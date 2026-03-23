package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/avito-internships/test-backend-1-EdOoO21/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func (db *DB) querier(ctx context.Context) querier {
	if tx, ok := ctx.Value(txContextKey{}).(pgx.Tx); ok {
		return tx
	}

	return db.pool
}

func timeOfDayDBValue(value domain.TimeOfDay) time.Time {
	return time.Date(2000, 1, 1, value.Hour, value.Minute, 0, 0, time.UTC)
}

func scanTimeOfDay(value time.Time) (domain.TimeOfDay, error) {
	return domain.ParseTimeOfDay(value.UTC().Format("15:04"))
}

func scanDaysOfWeek(values []int16) ([]domain.DayOfWeek, error) {
	result := make([]domain.DayOfWeek, 0, len(values))
	for _, value := range values {
		result = append(result, domain.DayOfWeek(value))
	}
	return result, nil
}

func dbDaysOfWeek(values []domain.DayOfWeek) []int16 {
	result := make([]int16, 0, len(values))
	for _, value := range values {
		result = append(result, int16(value))
	}
	return result
}

func normalizeScannedTimestamp(value time.Time) time.Time {
	return value.UTC()
}

func startOfDayUTC(value time.Time) time.Time {
	value = value.UTC()
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

func wrapScanError(entity string, err error) error {
	return fmt.Errorf("scan %s: %w", entity, err)
}
