package auth

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

type fakeUserRepository struct {
	createFn     func(ctx context.Context, user domain.User, passwordHash string) error
	getByIDFn    func(ctx context.Context, id uuid.UUID) (domain.User, bool, error)
	getByEmailFn func(ctx context.Context, email string) (domain.User, string, bool, error)
}

func (f fakeUserRepository) Create(ctx context.Context, user domain.User, passwordHash string) error {
	if f.createFn != nil {
		return f.createFn(ctx, user, passwordHash)
	}
	return nil
}

func (f fakeUserRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.User, bool, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}
	return domain.User{}, false, nil
}

func (f fakeUserRepository) GetByEmail(ctx context.Context, email string) (domain.User, string, bool, error) {
	if f.getByEmailFn != nil {
		return f.getByEmailFn(ctx, email)
	}
	return domain.User{}, "", false, nil
}

type fakeIDGenerator struct{ next uuid.UUID }

func (f fakeIDGenerator) NewUUID() uuid.UUID { return f.next }

type fakeClock struct{ now time.Time }

func (f fakeClock) NowUTC() time.Time { return f.now }

type fakePasswordHasher struct {
	hashFn    func(password string) (string, error)
	compareFn func(hash, password string) error
}

func (f fakePasswordHasher) Hash(password string) (string, error) {
	if f.hashFn != nil {
		return f.hashFn(password)
	}
	return "", nil
}

func (f fakePasswordHasher) Compare(hash, password string) error {
	if f.compareFn != nil {
		return f.compareFn(hash, password)
	}
	return nil
}

type fakeTokenService struct {
	issueFn func(ctx context.Context, claims appports.TokenClaims) (string, error)
}

func (f fakeTokenService) IssueToken(ctx context.Context, claims appports.TokenClaims) (string, error) {
	if f.issueFn != nil {
		return f.issueFn(ctx, claims)
	}
	return "", nil
}

func TestService_DummyLogin_ReturnsFixedUserForRole(t *testing.T) {
	t.Parallel()

	service := NewService(
		fakeUserRepository{},
		fakeIDGenerator{},
		fakeClock{},
		fakePasswordHasher{},
		fakeTokenService{issueFn: func(ctx context.Context, claims appports.TokenClaims) (string, error) {
			if claims.UserID != DummyAdminUserID {
				t.Fatalf("claims.UserID = %v, want %v", claims.UserID, DummyAdminUserID)
			}
			if claims.Role != domain.RoleAdmin {
				t.Fatalf("claims.Role = %q, want %q", claims.Role, domain.RoleAdmin)
			}
			return "token", nil
		}},
	)

	out, err := service.DummyLogin(context.Background(), DummyLoginInput{Role: domain.RoleAdmin})
	if err != nil {
		t.Fatalf("DummyLogin() error = %v", err)
	}

	if out.Token != "token" {
		t.Fatalf("out.Token = %q, want %q", out.Token, "token")
	}

	if out.UserID != DummyAdminUserID {
		t.Fatalf("out.UserID = %v, want %v", out.UserID, DummyAdminUserID)
	}
}

func TestService_DummyLogin_InvalidRole(t *testing.T) {
	t.Parallel()

	service := NewService(fakeUserRepository{}, fakeIDGenerator{}, fakeClock{}, fakePasswordHasher{}, fakeTokenService{})
	_, err := service.DummyLogin(context.Background(), DummyLoginInput{Role: domain.Role("manager")})
	if !errors.Is(err, domain.ErrInvalidRole) {
		t.Fatalf("DummyLogin() error = %v, want %v", err, domain.ErrInvalidRole)
	}
}

func TestService_Register_Success(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	now := time.Date(2026, time.March, 24, 12, 0, 0, 0, time.UTC)
	var createdUser domain.User
	var createdHash string

	service := NewService(
		fakeUserRepository{
			getByEmailFn: func(ctx context.Context, email string) (domain.User, string, bool, error) {
				if email != "user@example.com" {
					t.Fatalf("email lookup = %q, want %q", email, "user@example.com")
				}
				return domain.User{}, "", false, nil
			},
			createFn: func(ctx context.Context, user domain.User, passwordHash string) error {
				createdUser = user
				createdHash = passwordHash
				return nil
			},
		},
		fakeIDGenerator{next: id},
		fakeClock{now: now},
		fakePasswordHasher{hashFn: func(password string) (string, error) {
			if password != "secret" {
				t.Fatalf("password passed to hash = %q, want %q", password, "secret")
			}
			return "hashed-secret", nil
		}},
		fakeTokenService{},
	)

	out, err := service.Register(context.Background(), RegisterInput{Email: "  USER@Example.com ", Password: "  secret  ", Role: domain.RoleUser})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if out.User.ID != id {
		t.Fatalf("out.User.ID = %v, want %v", out.User.ID, id)
	}

	if out.User.Email != "user@example.com" {
		t.Fatalf("out.User.Email = %q, want %q", out.User.Email, "user@example.com")
	}

	if createdUser.ID != id {
		t.Fatalf("createdUser.ID = %v, want %v", createdUser.ID, id)
	}

	if createdHash != "hashed-secret" {
		t.Fatalf("createdHash = %q, want %q", createdHash, "hashed-secret")
	}
}

func TestService_Register_ReturnsValidationAndConflictErrors(t *testing.T) {
	t.Parallel()

	t.Run("empty password", func(t *testing.T) {
		t.Parallel()

		service := NewService(fakeUserRepository{}, fakeIDGenerator{}, fakeClock{}, fakePasswordHasher{}, fakeTokenService{})
		_, err := service.Register(context.Background(), RegisterInput{Email: "user@example.com", Password: "   ", Role: domain.RoleUser})
		if !errors.Is(err, shared.ErrInvalidPassword) {
			t.Fatalf("Register() error = %v, want %v", err, shared.ErrInvalidPassword)
		}
	})

	t.Run("email exists", func(t *testing.T) {
		t.Parallel()

		service := NewService(
			fakeUserRepository{getByEmailFn: func(ctx context.Context, email string) (domain.User, string, bool, error) {
				return domain.User{ID: uuid.New()}, "hash", true, nil
			}},
			fakeIDGenerator{next: uuid.New()},
			fakeClock{now: time.Date(2026, time.March, 24, 12, 0, 0, 0, time.UTC)},
			fakePasswordHasher{},
			fakeTokenService{},
		)

		_, err := service.Register(context.Background(), RegisterInput{Email: "user@example.com", Password: "secret", Role: domain.RoleUser})
		if !errors.Is(err, shared.ErrEmailAlreadyExists) {
			t.Fatalf("Register() error = %v, want %v", err, shared.ErrEmailAlreadyExists)
		}
	})
}

func TestService_Login_Success(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	service := NewService(
		fakeUserRepository{getByEmailFn: func(ctx context.Context, email string) (domain.User, string, bool, error) {
			return domain.User{ID: userID, Role: domain.RoleUser}, "hashed", true, nil
		}},
		fakeIDGenerator{},
		fakeClock{},
		fakePasswordHasher{compareFn: func(hash, password string) error {
			if hash != "hashed" || password != "secret" {
				t.Fatalf("Compare() got (%q, %q), want (%q, %q)", hash, password, "hashed", "secret")
			}
			return nil
		}},
		fakeTokenService{issueFn: func(ctx context.Context, claims appports.TokenClaims) (string, error) {
			if claims.UserID != userID || claims.Role != domain.RoleUser {
				t.Fatalf("claims = %+v, want userID=%v role=%q", claims, userID, domain.RoleUser)
			}
			return "jwt", nil
		}},
	)

	out, err := service.Login(context.Background(), LoginInput{Email: "user@example.com", Password: "secret"})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if out.Token != "jwt" {
		t.Fatalf("out.Token = %q, want %q", out.Token, "jwt")
	}
}

func TestService_Login_ReturnsInvalidCredentials(t *testing.T) {
	t.Parallel()

	t.Run("missing user", func(t *testing.T) {
		t.Parallel()

		service := NewService(
			fakeUserRepository{getByEmailFn: func(ctx context.Context, email string) (domain.User, string, bool, error) {
				return domain.User{}, "", false, nil
			}},
			fakeIDGenerator{},
			fakeClock{},
			fakePasswordHasher{},
			fakeTokenService{},
		)

		_, err := service.Login(context.Background(), LoginInput{Email: "user@example.com", Password: "secret"})
		if !errors.Is(err, shared.ErrInvalidCredentials) {
			t.Fatalf("Login() error = %v, want %v", err, shared.ErrInvalidCredentials)
		}
	})

	t.Run("password mismatch", func(t *testing.T) {

		t.Parallel()

		service := NewService(
			fakeUserRepository{getByEmailFn: func(ctx context.Context, email string) (domain.User, string, bool, error) {
				return domain.User{ID: uuid.New(), Role: domain.RoleUser}, "hashed", true, nil
			}},
			fakeIDGenerator{},
			fakeClock{},
			fakePasswordHasher{compareFn: func(hash, password string) error {
				return errors.New("mismatch")
			}},
			fakeTokenService{},
		)

		_, err := service.Login(context.Background(), LoginInput{Email: "user@example.com", Password: "secret"})
		if !errors.Is(err, shared.ErrInvalidCredentials) {
			t.Fatalf("Login() error = %v, want %v", err, shared.ErrInvalidCredentials)
		}
	})
}
