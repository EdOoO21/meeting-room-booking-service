package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/avito-internships/test-backend-1-EdOoO21/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type RoomRepository struct {
	db *DB
}

func NewRoomRepository(db *DB) *RoomRepository {
	return &RoomRepository{db: db}
}

func (r *RoomRepository) Create(ctx context.Context, room domain.Room) error {
	const query = "INSERT INTO rooms (id, name, description, capacity, created_at) VALUES (@id, @name, @description, @capacity, @created_at)"

	_, err := r.db.querier(ctx).Exec(ctx, query, pgx.NamedArgs{
		"id":          room.ID,
		"name":        room.Name,
		"description": room.Description,
		"capacity":    room.Capacity,
		"created_at":  room.CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("create room: %w", err)
	}

	return nil
}

func (r *RoomRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Room, bool, error) {
	const query = "SELECT id, name, description, capacity, created_at FROM rooms WHERE id = @id"
	row := r.db.querier(ctx).QueryRow(ctx, query, pgx.NamedArgs{"id": id})

	return scanRoom(row)
}

func (r *RoomRepository) List(ctx context.Context) ([]domain.Room, error) {
	const query = "SELECT id, name, description, capacity, created_at FROM rooms ORDER BY created_at ASC, name ASC"

	rows, err := r.db.querier(ctx).Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list rooms: %w", err)
	}
	defer rows.Close()

	rooms := make([]domain.Room, 0)
	for rows.Next() {
		room, err := scanRoomRow(rows)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rooms: %w", err)
	}

	return rooms, nil
}

func scanRoom(row pgx.Row) (domain.Room, bool, error) {
	room, err := scanRoomRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Room{}, false, nil
		}
		return domain.Room{}, false, err
	}

	return room, true, nil
}

func scanRoomRow(row interface{ Scan(dest ...any) error }) (domain.Room, error) {
	var (
		id          uuid.UUID
		name        string
		description string
		capacity    *int
		createdAt   time.Time
	)

	if err := row.Scan(&id, &name, &description, &capacity, &createdAt); err != nil {
		return domain.Room{}, wrapScanError("room", err)
	}

	room, err := domain.NewRoom(id, name, description, capacity, normalizeScannedTimestamp(createdAt))
	if err != nil {
		return domain.Room{}, fmt.Errorf("build room: %w", err)
	}

	return room, nil
}
