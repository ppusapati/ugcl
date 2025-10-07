package mappers

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UUID helpers
func StringToUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

func UUIDToString(u uuid.UUID) string {
	return u.String()
}

func StringPtrToUUID(s *string) (*uuid.UUID, error) {
	if s == nil {
		return nil, nil
	}
	id, err := uuid.Parse(*s)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func UUIDPtrToStringPtr(u *uuid.UUID) *string {
	if u == nil {
		return nil
	}
	s := u.String()
	return &s
}

// Time helpers
func TimeToTimestamp(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}

func TimePtrToTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

func TimestampToTimePtr(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}

// String helpers
func StringPtrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func StringToStringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// JSON helpers
func JSONRawMessageToMap(raw json.RawMessage) (map[string]string, error) {
	if raw == nil {
		return nil, nil
	}

	var result map[string]string
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func MapToJSONRawMessage(m map[string]string) (json.RawMessage, error) {
	if m == nil {
		return json.RawMessage("{}"), nil
	}

	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Array helpers
func StringArrayToPostgresArray(arr []string) []string {
	if arr == nil {
		return []string{}
	}
	return arr
}

func PostgresArrayToStringArray(arr []string) []string {
	if arr == nil {
		return []string{}
	}
	return arr
}
