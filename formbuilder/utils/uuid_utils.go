// =============================================================================
// internal/helpers/utils/uuid_utils.go
// =============================================================================
package utils

import (
	"github.com/google/uuid"
)

// GenerateUUID generates a new UUID
func GenerateUUID() uuid.UUID {
	return uuid.New()
}

// ParseUUID parses a UUID string and returns error if invalid
func ParseUUID(uuidStr string) (uuid.UUID, error) {
	return uuid.Parse(uuidStr)
}

// IsValidUUID checks if a string is a valid UUID
func IsValidUUID(uuidStr string) bool {
	_, err := uuid.Parse(uuidStr)
	return err == nil
}

// UUIDToString converts UUID to string, returns empty string for nil UUID
func UUIDToString(id *uuid.UUID) string {
	if id == nil {
		return ""
	}
	return id.String()
}

// StringToUUIDPtr converts string to UUID pointer, returns nil for empty string
func StringToUUIDPtr(s string) *uuid.UUID {
	if s == "" {
		return nil
	}
	if id, err := uuid.Parse(s); err == nil {
		return &id
	}
	return nil
}

func StringToUUID(s string) uuid.UUID {
	if s == "" {
		return uuid.Nil
	}
	if id, err := uuid.Parse(s); err == nil {
		return id
	}
	return uuid.Nil
}
