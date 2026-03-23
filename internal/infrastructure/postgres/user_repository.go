package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/avito-internships/test-backend-1-EdOoO21/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserRepository struct {
	db *DB
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) error {
	const query = "INSERT INTO users (id, email, password_hash, role, created_at) VALUES (@id, @email, NULL, @role, @created_at)"

	_, err := r.db.querier(ctx).Exec(ctx, query, pgx.NamedArgs{
		"id":         user.ID,
		"email":      strings.TrimSpace(strings.ToLower(user.Email)),
		"role":       string(user.Role),
		"created_at": user.CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.User, bool, error) {
	const query = "SELECT id, email, role, created_at FROM users WHERE id = @id"
	row := r.db.querier(ctx).QueryRow(ctx, query, pgx.NamedArgs{"id": id})

	return scanUser(row)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (domain.User, bool, error) {
	const query = "SELECT id, email, role, created_at FROM users WHERE email = @email"
	row := r.db.querier(ctx).QueryRow(ctx, query, pgx.NamedArgs{"email": strings.TrimSpace(strings.ToLower(email))})

	return scanUser(row)
}

func scanUser(row pgx.Row) (domain.User, bool, error) {
	var (
		id        uuid.UUID
		email     string
		role      string
		createdAt time.Time
	)

	if err := row.Scan(&id, &email, &role, &createdAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, false, nil
		}
		return domain.User{}, false, wrapScanError("user", err)
	}

	user, err := domain.NewUser(id, email, domain.Role(role), normalizeScannedTimestamp(createdAt))
	if err != nil {
		return domain.User{}, false, fmt.Errorf("build user: %w", err)
	}

	return user, true, nil
}
