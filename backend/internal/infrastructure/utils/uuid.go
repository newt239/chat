package utils

import (
	"fmt"

	"github.com/google/uuid"

	domerr "github.com/newt239/chat/internal/domain/errors"
)

// ParseUUID は文字列をUUIDに変換します。形式の誤りは入力の検証エラーとして返します
func ParseUUID(id string, label string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: invalid %s format", domerr.ErrValidation, label)
	}
	return parsed, nil
}

func ParseUUIDs(ids []string, label string) ([]uuid.UUID, error) {
	parsed := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		u, err := ParseUUID(id, label)
		if err != nil {
			return nil, err
		}
		parsed = append(parsed, u)
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
