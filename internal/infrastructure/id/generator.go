package id

import (
	"github.com/google/uuid"
)

type Generator struct{}

func New() Generator {
	return Generator{}
}

func (Generator) NewUUID() uuid.UUID {
	return uuid.New()
}
