package util

import (
	"fmt"

	"github.com/google/uuid"
)

type (
	UUID interface {
		fmt.Stringer
	}
	uuidEntity struct {
		v uuid.UUID
	}
)

func NewUUID() UUID {
	return &uuidEntity{
		v: uuid.New(),
	}
}

func (s *uuidEntity) String() string {
	if s == nil {
		return ""
	}
	return s.v.String()
}
