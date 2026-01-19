// =============================================================================
// internal/helpers/utils/string_utils.go
// =============================================================================
package utils

import (
	"strings"
)

// StringPtr returns a pointer to the given string, nil for empty string
func StringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// StringValue returns the string value from pointer, empty string if nil
func StringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// BoolPtr returns a pointer to the given bool
func BoolPtr(b bool) *bool {
	return &b
}

// BoolValue returns the bool value from pointer, false if nil
func BoolValue(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// Int32Ptr returns a pointer to the given int32, nil for zero
func Int32Ptr(i int32) *int32 {
	if i == 0 {
		return nil
	}
	return &i
}

// Int32Value returns the int32 value from pointer, zero if nil
func Int32Value(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}

// Int64Ptr returns a pointer to the given int64, nil for zero
func Int64Ptr(i int64) *int64 {
	if i == 0 {
		return nil
	}
	return &i
}

// Int64Value returns the int64 value from pointer, zero if nil
func Int64Value(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

// TrimSpace trims whitespace and returns pointer, nil for empty result
func TrimSpace(s string) *string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// IsEmpty checks if string pointer is nil or empty
func IsEmpty(s *string) bool {
	return s == nil || *s == ""
}

// Coalesce returns the first non-empty string
func Coalesce(strings ...*string) string {
	for _, s := range strings {
		if !IsEmpty(s) {
			return *s
		}
	}
	return ""
}
