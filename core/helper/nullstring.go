package helper

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Null helpers
func nullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

func nullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func nullUUID(u *uuid.UUID) uuid.NullUUID {
	if u == nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{UUID: *u, Valid: true}
}

func nullFloat64(f *float64) sql.NullFloat64 {
	if f == nil {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: *f, Valid: true}
}

// Pointer helpers
func stringPtr(s sql.NullString) *string {
	if s.Valid {
		return &s.String
	}
	return nil
}

func timePtr(t sql.NullTime) *time.Time {
	if t.Valid {
		return &t.Time
	}
	return nil
}

func uuidPtr(u uuid.NullUUID) *uuid.UUID {
	if u.Valid {
		return &u.UUID
	}
	return nil
}
