package password

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type Service struct{}

func New() Service {
	return Service{}
}

func (Service) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	return string(hash), nil
}

func (Service) Compare(hash, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return fmt.Errorf("compare password hash: %w", err)
	}

	return nil
}
