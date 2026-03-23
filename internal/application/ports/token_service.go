package ports

import (
	"context"

	"github.com/avito-internships/test-backend-1-EdOoO21/internal/domain"
	"github.com/google/uuid"
)

type TokenClaims struct {
	UserID uuid.UUID
	Role   domain.Role
}

type TokenService interface {
	IssueToken(ctx context.Context, claims TokenClaims) (string, error)
}
