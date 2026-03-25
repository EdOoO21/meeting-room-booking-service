package rooms

import (
	"context"
	"fmt"
	"strings"

	appports "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/ports"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/application/shared"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/domain"
)

type Service struct {
	rooms appports.RoomRepository
	ids   appports.IDGenerator
	clock appports.Clock
}

type CreateInput struct {
	Actor       shared.Actor
	Name        string
	Description string
	Capacity    *int
}

type CreateOutput struct {
	Room domain.Room
}

type ListInput struct {
	Actor shared.Actor
}

type ListOutput struct {
	Rooms []domain.Room
}

func NewService(rooms appports.RoomRepository, ids appports.IDGenerator, clock appports.Clock) *Service {
	return &Service{rooms: rooms, ids: ids, clock: clock}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (CreateOutput, error) {
	if err := input.Actor.RequireRole(domain.RoleAdmin); err != nil {
		return CreateOutput{}, fmt.Errorf("authorize room creation: %w", err)
	}

	room, err := domain.NewRoom(
		s.ids.NewUUID(),
		strings.TrimSpace(input.Name),
		strings.TrimSpace(input.Description),
		input.Capacity,
		s.clock.NowUTC(),
	)
	if err != nil {
		return CreateOutput{}, fmt.Errorf("build room: %w", err)
	}

	if createErr := s.rooms.Create(ctx, room); createErr != nil {
		return CreateOutput{}, fmt.Errorf("create room: %w", createErr)
	}

	return CreateOutput{Room: room}, nil
}

func (s *Service) List(ctx context.Context, input ListInput) (ListOutput, error) {
	if err := input.Actor.RequireRole(domain.RoleAdmin, domain.RoleUser); err != nil {
		return ListOutput{}, fmt.Errorf("authorize rooms list: %w", err)
	}

	rooms, err := s.rooms.List(ctx)
	if err != nil {
		return ListOutput{}, fmt.Errorf("list rooms: %w", err)
	}

	return ListOutput{Rooms: rooms}, nil
}
