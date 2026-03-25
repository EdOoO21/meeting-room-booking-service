package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/avito-internships/test-backend-1-EdOoO21/internal/application/shared"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type BookingRepository struct {
	db *DB
}

func NewBookingRepository(db *DB) *BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) Create(ctx context.Context, booking domain.Booking) error {
	const query = "INSERT INTO bookings (id, slot_id, user_id, status, conference_link, created_at) VALUES (@id, @slot_id, @user_id, @status, @conference_link, @created_at)"

	_, err := r.db.querier(ctx).Exec(ctx, query, pgx.NamedArgs{
		"id":              booking.ID,
		"slot_id":         booking.SlotID,
		"user_id":         booking.UserID,
		"status":          string(booking.Status),
		"conference_link": booking.ConferenceLink,
		"created_at":      booking.CreatedAt,
	})
	if err != nil {
		if isActiveBookingConflict(err) {
			return shared.ErrSlotBooked
		}
		return fmt.Errorf("create booking: %w", err)
	}

	return nil
}

func (r *BookingRepository) Update(ctx context.Context, booking domain.Booking) error {
	const query = "UPDATE bookings SET status = @status, conference_link = @conference_link, created_at = @created_at WHERE id = @id"

	commandTag, err := r.db.querier(ctx).Exec(ctx, query, pgx.NamedArgs{
		"id":              booking.ID,
		"status":          string(booking.Status),
		"conference_link": booking.ConferenceLink,
		"created_at":      booking.CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("update booking: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return shared.ErrBookingNotFound
	}

	return nil
}

func (r *BookingRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Booking, bool, error) {
	const query = "SELECT id, slot_id, user_id, status, conference_link, created_at FROM bookings WHERE id = @id"
	row := r.db.querier(ctx).QueryRow(ctx, query, pgx.NamedArgs{"id": id})

	return scanBooking(row)
}

func (r *BookingRepository) HasActiveBySlotID(ctx context.Context, slotID uuid.UUID) (bool, error) {
	const query = "SELECT EXISTS (SELECT 1 FROM bookings WHERE slot_id = @slot_id AND status = 'active')"

	var exists bool
	if err := r.db.querier(ctx).QueryRow(ctx, query, pgx.NamedArgs{"slot_id": slotID}).Scan(&exists); err != nil {
		return false, fmt.Errorf("check active booking by slot: %w", err)
	}

	return exists, nil
}

func (r *BookingRepository) ListByUserFuture(ctx context.Context, userID uuid.UUID, now time.Time) ([]domain.Booking, error) {
	const query = "SELECT b.id, b.slot_id, b.user_id, b.status, b.conference_link, b.created_at FROM bookings b JOIN slots s ON s.id = b.slot_id WHERE b.user_id = @user_id AND s.start_at > @now ORDER BY s.start_at ASC, b.created_at ASC"

	rows, err := r.db.querier(ctx).Query(ctx, query, pgx.NamedArgs{
		"user_id": userID,
		"now":     now,
	})
	if err != nil {
		return nil, fmt.Errorf("list future bookings by user: %w", err)
	}
	defer rows.Close()

	bookings := make([]domain.Booking, 0)
	for rows.Next() {
		booking, scanErr := scanBookingRow(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		bookings = append(bookings, booking)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("iterate future bookings by user: %w", rowsErr)
	}

	return bookings, nil
}

func (r *BookingRepository) List(ctx context.Context, page, pageSize int) ([]domain.Booking, int, error) {
	const countQuery = "SELECT COUNT(*) FROM bookings"
	const listQuery = "SELECT id, slot_id, user_id, status, conference_link, created_at FROM bookings ORDER BY created_at DESC, id DESC LIMIT @limit OFFSET @offset"

	var total int
	if err := r.db.querier(ctx).QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count bookings: %w", err)
	}

	offset := (page - 1) * pageSize
	rows, err := r.db.querier(ctx).Query(ctx, listQuery, pgx.NamedArgs{
		"limit":  pageSize,
		"offset": offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list bookings: %w", err)
	}
	defer rows.Close()

	bookings := make([]domain.Booking, 0)
	for rows.Next() {
		booking, scanErr := scanBookingRow(rows)
		if scanErr != nil {
			return nil, 0, scanErr
		}
		bookings = append(bookings, booking)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, 0, fmt.Errorf("iterate bookings: %w", rowsErr)
	}

	return bookings, total, nil
}

func scanBooking(row pgx.Row) (domain.Booking, bool, error) {
	booking, err := scanBookingRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Booking{}, false, nil
		}
		return domain.Booking{}, false, err
	}

	return booking, true, nil
}

func scanBookingRow(row interface{ Scan(dest ...any) error }) (domain.Booking, error) {
	var (
		id             uuid.UUID
		slotID         uuid.UUID
		userID         uuid.UUID
		status         string
		conferenceLink *string
		createdAt      time.Time
	)

	if err := row.Scan(&id, &slotID, &userID, &status, &conferenceLink, &createdAt); err != nil {
		return domain.Booking{}, wrapScanError("booking", err)
	}

	booking, err := domain.NewBooking(id, slotID, userID, domain.BookingStatus(status), conferenceLink, normalizeScannedTimestamp(createdAt))
	if err != nil {
		return domain.Booking{}, fmt.Errorf("build booking: %w", err)
	}

	return booking, nil
}

func isActiveBookingConflict(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}

	return pgErr.Code == "23505" && strings.Contains(pgErr.ConstraintName, "bookings_active_slot")
}
