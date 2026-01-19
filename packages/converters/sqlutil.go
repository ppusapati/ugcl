package converters

import (
	"database/sql"
	"encoding/json"

	"github.com/sqlc-dev/pqtype"
)

func NullRawMessageToMap(nrm pqtype.NullRawMessage) map[string]interface{} {
	if !nrm.Valid || len(nrm.RawMessage) == 0 {
		return nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal(nrm.RawMessage, &m); err != nil {
		return nil
	}
	return m
}

func MapToNullRawMessage(m map[string]interface{}) pqtype.NullRawMessage {
	if m == nil {
		return pqtype.NullRawMessage{Valid: false}
	}
	b, err := json.Marshal(m)
	if err != nil {
		return pqtype.NullRawMessage{Valid: false}
	}
	return pqtype.NullRawMessage{
		RawMessage: b,
		Valid:      true,
	}
}

func MapToNullString(m map[string]any) sql.NullString {
	if m == nil {
		return sql.NullString{Valid: false}
	}
	b, err := json.Marshal(m)
	if err != nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{
		String: string(b),
		Valid:  true,
	}
}

func NullStringToMap(ns sql.NullString) map[string]any {
	if !ns.Valid {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(ns.String), &m); err != nil {
		return nil
	}
	return m
}
