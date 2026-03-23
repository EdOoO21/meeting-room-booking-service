package shared

import (
	"slices"

	"github.com/avito-internships/test-backend-1-EdOoO21/internal/domain"
	"github.com/google/uuid"
)

type Actor struct {
	UserID uuid.UUID
	Role   domain.Role
}

func (a Actor) IsAuthenticated() bool {
	return a.UserID != uuid.Nil && a.Role.IsValid()
}

func (a Actor) RequireAuthenticated() error {
	if !a.IsAuthenticated() {
		return ErrUnauthorized
	}

	return nil
}

func (a Actor) RequireRole(allowed ...domain.Role) error {
	if err := a.RequireAuthenticated(); err != nil {
		return err
	}

	if slices.Contains(allowed, a.Role) {
		return nil
	}

	return ErrForbidden
}
