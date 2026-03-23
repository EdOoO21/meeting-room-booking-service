package shared

import (
	"errors"
	"testing"

	"github.com/avito-internships/test-backend-1-EdOoO21/internal/domain"
	"github.com/google/uuid"
)

func TestActorAuthenticationAndRoles(t *testing.T) {
	t.Parallel()

	actor := Actor{UserID: uuid.New(), Role: domain.RoleAdmin}
	if !actor.IsAuthenticated() {
		t.Fatal("actor should be authenticated")
	}

	if err := actor.RequireAuthenticated(); err != nil {
		t.Fatalf("RequireAuthenticated() error = %v", err)
	}

	if err := actor.RequireRole(domain.RoleAdmin, domain.RoleUser); err != nil {
		t.Fatalf("RequireRole() error = %v", err)
	}

	if err := actor.RequireRole(domain.RoleUser); !errors.Is(err, ErrForbidden) {
		t.Fatalf("RequireRole() error = %v, want %v", err, ErrForbidden)
	}
}

func TestActorRequireAuthenticated_ReturnsUnauthorized(t *testing.T) {
	t.Parallel()

	actor := Actor{}
	if actor.IsAuthenticated() {
		t.Fatal("zero actor should not be authenticated")
	}

	if err := actor.RequireAuthenticated(); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("RequireAuthenticated() error = %v, want %v", err, ErrUnauthorized)
	}

	if err := actor.RequireRole(domain.RoleAdmin); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("RequireRole() error = %v, want %v", err, ErrUnauthorized)
	}
}
