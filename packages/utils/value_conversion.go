package utils

import (
	"database/sql"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func StringOrNil(s sql.NullString) *wrapperspb.StringValue {
	if s.Valid {
		return &wrapperspb.StringValue{Value: *&s.String}
	}
	return nil
}

// Helper function to convert wrapperspb.StringValue to sql.NullString
func ToNullString(sv *wrapperspb.StringValue) sql.NullString {
	if sv != nil {
		return sql.NullString{String: sv.Value, Valid: true}
	}
	return sql.NullString{Valid: false}
}

func IntOrDefault(value *int32) int32 {
	if value != nil {
		return *value
	}
	return 0
}

func TimeToTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t != nil {
		return timestamppb.New(*t)
	}
	return nil
}

func StrSliceOrDefault(s *[]string) []string {
	if s != nil {
		return *s
	}
	return []string{}
}

func ByteSliceOrNil(b *[]byte) []byte {
	if b != nil {
		return *b
	}
	return nil
}

func BoolOrNil(value *bool) bool {
	if value != nil {
		return *value
	}
	return false // Change to the default value for bool fields
}

func StringPtrToNullString(s *string) sql.NullString {
	if s != nil {
		return sql.NullString{
			String: *s,
			Valid:  true,
		}
	}
	return sql.NullString{
		Valid: false,
	}
}

// Convert sql.NullString to *string
func NullStringToStringPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}
