package jwt

import (
	"context"
	"fmt"
	"time"

	appports "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/ports"
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
