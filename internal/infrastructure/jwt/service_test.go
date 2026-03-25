package jwt

import (
	"context"
	"testing"
	"time"

	appports "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/ports"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/domain"
	"github.com/google/uuid"
)

func TestService_IssueAndParseToken(t *testing.T) {
	t.Parallel()

	svc := New("test-secret", time.Hour)
	claims := appports.TokenClaims{
		UserID: uuid.New(),
		Role:   domain.RoleUser,
	}

	token, err := svc.IssueToken(context.Background(), claims)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}
	if token == "" {
		t.Fatal("IssueToken() returned empty token")
	}

	parsed, err := svc.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if parsed.UserID != claims.UserID {
		t.Fatalf("parsed.UserID = %v, want %v", parsed.UserID, claims.UserID)
	}
	if parsed.Role != claims.Role {
		t.Fatalf("parsed.Role = %q, want %q", parsed.Role, claims.Role)
	}
}

func TestService_ParseToken_ReturnsErrorForInvalidToken(t *testing.T) {
	t.Parallel()

	svc := New("test-secret", time.Hour)

	if _, err := svc.ParseToken("not-a-jwt"); err == nil {
		t.Fatal("ParseToken() error = nil, want error for malformed token")
	}
}
