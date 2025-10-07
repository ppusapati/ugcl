package mappers

import (
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"
)

// String pointer helpers
func stringPtrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func stringToStringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Bool pointer helpers
func boolPtrToBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func boolToBoolPtr(b bool) *bool {
	return &b
}

// Int32 pointer helpers
func int32PtrToInt32(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}

func int32ToInt32Ptr(i int32) *int32 {
	return &i
}

// Float64 helpers
func float64PtrToFloat64(f *float64) float64 {
	if f == nil {
		return 0.0
	}
	return *f
}

func float64ToFloat64Ptr(f float64) *float64 {
	return &f
}

func sqlNullFloat64ToFloat64(nf sql.NullFloat64) float64 {
	if !nf.Valid {
		return 0.0
	}
	return nf.Float64
}

func float64ToSqlNullFloat64(f float64) sql.NullFloat64 {
	if f == 0.0 {
		return sql.NullFloat64{Valid: false}
	}
	return sql.NullFloat64{Float64: f, Valid: true}
}

// UUID helpers
func uuidPtrToString(u *uuid.UUID) string {
	if u == nil {
		return ""
	}
	return u.String()
}

func uuidNullToString(u uuid.NullUUID) string {
	if !u.Valid {
		return ""
	}
	return u.UUID.String()
}

func stringToUUIDNull(s string) uuid.NullUUID {
	if s == "" {
		return uuid.NullUUID{Valid: false}
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.NullUUID{Valid: false}
	}
	return uuid.NullUUID{UUID: id, Valid: true}
}

// JSON helpers
func jsonRawMessageToMap(raw json.RawMessage) map[string]interface{} {
	if raw == nil {
		return nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	return m
}

func structToJSONRawMessage(s *structpb.Struct) json.RawMessage {
	if s == nil {
		return nil
	}
	data, err := json.Marshal(s.AsMap())
	if err != nil {
		return nil
	}
	return data
}

// SQL null string helpers
func sqlNullStringToString(ns sql.NullString) string {
	if !ns.Valid {
		return ""
	}
	return ns.String
}

func stringToSqlNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
