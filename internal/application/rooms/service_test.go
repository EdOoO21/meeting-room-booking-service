package rooms

import (
	"context"
	"errors"
	"testing"
	"time"

	appports "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/ports"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/application/shared"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/domain"
	"github.com/google/uuid"
)

type fakeRoomRepository struct {
	createFn  func(ctx context.Context, room domain.Room) error
	getByIDFn func(ctx context.Context, id uuid.UUID) (domain.Room, bool, error)
	listFn    func(ctx context.Context) ([]domain.Room, error)
}

func (f fakeRoomRepository) Create(ctx context.Context, room domain.Room) error {
	if f.createFn != nil {
		return f.createFn(ctx, room)
	}
	return nil
}

func (f fakeRoomRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Room, bool, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}
	return domain.Room{}, false, nil
}

func (f fakeRoomRepository) List(ctx context.Context) ([]domain.Room, error) {
	if f.listFn != nil {
		return f.listFn(ctx)
	}
	return nil, nil
}

type fakeIDGenerator struct{ next uuid.UUID }

func (f fakeIDGenerator) NewUUID() uuid.UUID { return f.next }

type fakeClock struct{ now time.Time }

func (f fakeClock) NowUTC() time.Time { return f.now }

func TestService_Create_Success(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	now := time.Date(2026, time.March, 24, 12, 0, 0, 0, time.UTC)
	capacity := 4
	var created domain.Room

	service := NewService(
		fakeRoomRepository{createFn: func(ctx context.Context, room domain.Room) error {
			created = room
			return nil
		}},
		fakeIDGenerator{next: id},
		fakeClock{now: now},
	)

	out, err := service.Create(context.Background(), CreateInput{
		Actor:       shared.Actor{UserID: uuid.New(), Role: domain.RoleAdmin},
		Name:        "  Green room ",
		Description: "  Quiet and bright ",
		Capacity:    &capacity,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if out.Room.ID != id {
		t.Fatalf("out.Room.ID = %v, want %v", out.Room.ID, id)
	}

	if out.Room.Name != "Green room" {
		t.Fatalf("out.Room.Name = %q, want %q", out.Room.Name, "Green room")
	}

	if created.ID != id {
		t.Fatalf("created.ID = %v, want %v", created.ID, id)
	}
}

func TestService_Create_RequiresAdmin(t *testing.T) {
	t.Parallel()

	service := NewService(fakeRoomRepository{}, fakeIDGenerator{}, fakeClock{})
	_, err := service.Create(context.Background(), CreateInput{Actor: shared.Actor{UserID: uuid.New(), Role: domain.RoleUser}, Name: "Room"})
	if !errors.Is(err, shared.ErrForbidden) {
		t.Fatalf("Create() error = %v, want %v", err, shared.ErrForbidden)
	}
}

func TestService_List_ReturnsRoomsForAuthorizedActor(t *testing.T) {
	t.Parallel()

	wantRooms := []domain.Room{{ID: uuid.New(), Name: "A"}, {ID: uuid.New(), Name: "B"}}
	service := NewService(
		fakeRoomRepository{listFn: func(ctx context.Context) ([]domain.Room, error) {
			return wantRooms, nil
		}},
		fakeIDGenerator{},
		fakeClock{},
	)

	out, err := service.List(context.Background(), ListInput{Actor: shared.Actor{UserID: uuid.New(), Role: domain.RoleUser}})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(out.Rooms) != len(wantRooms) {
		t.Fatalf("len(out.Rooms) = %d, want %d", len(out.Rooms), len(wantRooms))
	}
}

func TestService_List_RequiresAuthentication(t *testing.T) {
	t.Parallel()

	service := NewService(fakeRoomRepository{}, fakeIDGenerator{}, fakeClock{})
	_, err := service.List(context.Background(), ListInput{Actor: shared.Actor{}})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("List() error = %v, want %v", err, shared.ErrUnauthorized)
	}
}

var _ appports.RoomRepository = fakeRoomRepository{}
