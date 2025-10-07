package mappers

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UUID helpers
func stringToUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

func uuidToString(u uuid.UUID) string {
	return u.String()
}

func stringPtrToUUID(s *string) (*uuid.UUID, error) {
	if s == nil {
		return nil, nil
	}
	id, err := uuid.Parse(*s)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func uuidPtrToStringPtr(u *uuid.UUID) *string {
	if u == nil {
		return nil
	}
	s := u.String()
	return &s
}

// Time helpers
func timeToTimestamp(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}

func timePtrToTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

func timestampToTimePtr(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}

// String helpers
func stringPtrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// JSON helpers
func jsonRawMessageToMap(raw json.RawMessage) (map[string]string, error) {
	if raw == nil {
		return nil, nil
	}

	var result map[string]string
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func mapToJSONRawMessage(m map[string]string) (json.RawMessage, error) {
	if m == nil {
		return json.RawMessage("{}"), nil
	}

	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return data, nil
}
