package auth

import (
	"context"

	appports "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/ports"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/domain"
	"github.com/google/uuid"
)

var (
	DummyAdminUserID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	DummyUserUserID  = uuid.MustParse("00000000-0000-0000-0000-000000000002")
)

type Service struct {
	tokens appports.TokenService
}

type DummyLoginInput struct {
	Role domain.Role
}

type DummyLoginOutput struct {
	Token  string
	UserID uuid.UUID
	Role   domain.Role
}

func NewService(tokens appports.TokenService) *Service {
	return &Service{tokens: tokens}
}

func (s *Service) DummyLogin(ctx context.Context, input DummyLoginInput) (DummyLoginOutput, error) {
	if !input.Role.IsValid() {
		return DummyLoginOutput{}, domain.ErrInvalidRole
	}

	userID := dummyUserIDForRole(input.Role)
	token, err := s.tokens.IssueToken(ctx, appports.TokenClaims{
		UserID: userID,
		Role:   input.Role,
	})
	if err != nil {
		return DummyLoginOutput{}, err
	}

	return DummyLoginOutput{Token: token, UserID: userID, Role: input.Role}, nil
}

func dummyUserIDForRole(role domain.Role) uuid.UUID {
	if role == domain.RoleAdmin {
		return DummyAdminUserID
	}

	return DummyUserUserID
}
