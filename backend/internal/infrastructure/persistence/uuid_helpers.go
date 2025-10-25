package persistence

import (
	"fmt"

	"github.com/google/uuid"
)

func parseUUID(id string, label string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid %s format", label)
	}
	return parsed, nil
}

func parseUUIDPtr(id *string, label string) (*uuid.UUID, error) {
	if id == nil {
		return nil, nil
	}

	parsed, err := uuid.Parse(*id)
	if err != nil {
		return nil, fmt.Errorf("invalid %s format", label)
	}

	value := parsed
	return &value, nil
}
