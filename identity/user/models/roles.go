package models

import (
	"time"
)

type Role struct {
	ID          string         `db:"id"`
	Name        string         `db:"name"`
	TenantID    string         `db:"tenant_id"`
	ParentID    string         `db:"parent_id"`
	Metadata    map[string]any `db:"metadata" json:"metadata"`
	IsPreserved bool
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
