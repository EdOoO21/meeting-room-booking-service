package auth

import (
	"context"
	"fmt"
	"strings"

	appports "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/ports"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/application/shared"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/domain"
	"github.com/google/uuid"
)

var (
	DummyAdminUserID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	DummyUserUserID  = uuid.MustParse("00000000-0000-0000-0000-000000000002")
)

type Service struct {
	users     appports.UserRepository
	ids       appports.IDGenerator
	clock     appports.Clock
	passwords appports.PasswordHasher
	tokens    appports.TokenService
}

type DummyLoginInput struct {
	Role domain.Role
}

type DummyLoginOutput struct {
	Token  string
	UserID uuid.UUID
	Role   domain.Role
}

type RegisterInput struct {
	Email    string
	Password string
	Role     domain.Role
}

type RegisterOutput struct {
	User domain.User
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	Token string
}

func NewService(
	users appports.UserRepository,
	ids appports.IDGenerator,
	clock appports.Clock,
	passwords appports.PasswordHasher,
	tokens appports.TokenService,
) *Service {
	return &Service{users: users, ids: ids, clock: clock, passwords: passwords, tokens: tokens}
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
		return DummyLoginOutput{}, fmt.Errorf("issue dummy login token: %w", err)
	}

	return DummyLoginOutput{Token: token, UserID: userID, Role: input.Role}, nil
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (RegisterOutput, error) {
	password := strings.TrimSpace(input.Password)
	if password == "" {
		return RegisterOutput{}, shared.ErrInvalidPassword
	}

	user, err := domain.NewUser(s.ids.NewUUID(), input.Email, input.Role, s.clock.NowUTC())
	if err != nil {
		return RegisterOutput{}, fmt.Errorf("create domain user: %w", err)
	}

	_, _, exists, err := s.users.GetByEmail(ctx, user.Email)
	if err != nil {
		return RegisterOutput{}, fmt.Errorf("get user by email: %w", err)
	}
	if exists {
		return RegisterOutput{}, shared.ErrEmailAlreadyExists
	}

	hash, err := s.passwords.Hash(password)
	if err != nil {
		return RegisterOutput{}, fmt.Errorf("hash user password: %w", err)
	}

	if createErr := s.users.Create(ctx, user, hash); createErr != nil {
		return RegisterOutput{}, fmt.Errorf("create user: %w", createErr)
	}

	return RegisterOutput{User: user}, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (LoginOutput, error) {
	user, passwordHash, exists, err := s.users.GetByEmail(ctx, input.Email)
	if err != nil {
		return LoginOutput{}, fmt.Errorf("get user by email: %w", err)
	}
	if !exists {
		return LoginOutput{}, shared.ErrInvalidCredentials
	}

	if compareErr := s.passwords.Compare(passwordHash, input.Password); compareErr != nil {
		return LoginOutput{}, shared.ErrInvalidCredentials
	}

	token, err := s.tokens.IssueToken(ctx, appports.TokenClaims{
		UserID: user.ID,
		Role:   user.Role,
	})
	if err != nil {
		return LoginOutput{}, fmt.Errorf("issue login token: %w", err)
	}

	return LoginOutput{Token: token}, nil
}

func dummyUserIDForRole(role domain.Role) uuid.UUID {
	if role == domain.RoleAdmin {
		return DummyAdminUserID
	}

	return DummyUserUserID
}
