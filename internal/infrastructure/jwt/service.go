package jwt

import (
	"context"
	"fmt"
	"time"

	appports "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/ports"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/domain"
	"github.com/google/uuid"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

type Service struct {
	secret []byte
	ttl    time.Duration
}

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwtv5.RegisteredClaims
}

func New(secret string, ttl time.Duration) *Service {
	return &Service{secret: []byte(secret), ttl: ttl}
}

func (s *Service) IssueToken(ctx context.Context, claims appports.TokenClaims) (string, error) {
	_ = ctx

	now := time.Now().UTC()
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, Claims{
		UserID: claims.UserID.String(),
		Role:   string(claims.Role),
		RegisteredClaims: jwtv5.RegisteredClaims{
			IssuedAt:  jwtv5.NewNumericDate(now),
			ExpiresAt: jwtv5.NewNumericDate(now.Add(s.ttl)),
		},
	})

	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("sign jwt: %w", err)
	}

	return signed, nil
}

func (s *Service) ParseToken(token string) (appports.TokenClaims, error) {
	parsed, err := jwtv5.ParseWithClaims(token, &Claims{}, func(t *jwtv5.Token) (any, error) {
		if t.Method != jwtv5.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Method.Alg())
		}
		return s.secret, nil
	})
	if err != nil {
		return appports.TokenClaims{}, fmt.Errorf("parse jwt: %w", err)
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return appports.TokenClaims{}, fmt.Errorf("invalid jwt claims")
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return appports.TokenClaims{}, fmt.Errorf("parse jwt user id: %w", err)
	}

	role := domain.Role(claims.Role)
	if !role.IsValid() {
		return appports.TokenClaims{}, fmt.Errorf("invalid jwt role")
	}

	return appports.TokenClaims{UserID: userID, Role: role}, nil
}
