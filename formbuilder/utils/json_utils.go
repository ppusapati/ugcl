// =============================================================================
// internal/helpers/utils/json_utils.go
// =============================================================================
package utils

import (
	"encoding/json"
	"fmt"
)

// PrettyPrintJSON formats JSON for human-readable output
func PrettyPrintJSON(v interface{}) (string, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(b), nil
}

// CompactJSON compacts JSON by removing whitespace
func CompactJSON(jsonStr string) (string, error) {
	var obj interface{}
	if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}

	b, err := json.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(b), nil
}

// IsValidJSON checks if a string is valid JSON
func IsValidJSON(jsonStr string) bool {
	var obj interface{}
	return json.Unmarshal([]byte(jsonStr), &obj) == nil
}

// MarshalToRawMessage converts interface{} to json.RawMessage
func MarshalToRawMessage(v interface{}) (json.RawMessage, error) {
	if v == nil {
		return json.RawMessage("null"), nil
	}
	return json.Marshal(v)
}

// UnmarshalFromRawMessage converts json.RawMessage to interface{}
func UnmarshalFromRawMessage(raw json.RawMessage, v interface{}) error {
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, v)
}
