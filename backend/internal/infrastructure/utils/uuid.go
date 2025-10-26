package utils

import (
	"fmt"

	"github.com/google/uuid"
)

// ParseUUID は文字列をUUIDに変換します
func ParseUUID(id string, label string) (uuid.UUID, error) {
	if id == "" {
		return uuid.Nil, fmt.Errorf("%s cannot be empty", label)
	}
	parsed, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid %s format: %w", label, err)
	}
	return parsed, nil
}

// ParseUUIDPtr は文字列ポインタをUUIDポインタに変換します
func ParseUUIDPtr(id *string) *uuid.UUID {
	if id == nil {
		return nil
	}
	parsed, err := uuid.Parse(*id)
	if err != nil {
		return nil
	}
	return &parsed
}

// UUIDToString はUUIDを文字列に変換します
func UUIDToString(id uuid.UUID) string {
	if id == uuid.Nil {
		return ""
	}
	return id.String()
}

// UUIDPtrToStringPtr はUUIDポインタを文字列ポインタに変換します
func UUIDPtrToStringPtr(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	value := UUIDToString(*id)
	if value == "" {
		return nil
	}
	return &value
}

// NewUUID は新しいUUIDを生成します
func NewUUID() string {
	return uuid.NewString()
}
