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

// TokenService выпускает токены доступа для API.
type TokenService interface {
	// IssueToken создает токен по переданным claims.
	IssueToken(ctx context.Context, claims TokenClaims) (string, error)
}
