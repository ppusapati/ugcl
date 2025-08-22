package models

import (
	"time"
)

type PermissionSide int32

const (
	SideBoth       PermissionSide = 0
	SideHostOnly   PermissionSide = 1
	SideTenantOnly PermissionSide = 2
)

type PermissionDef struct {
	// Primary/business keys
	Name      string `bun:"name,pk"          json:"name"`      // proto: name
	Namespace string `bun:"namespace"        json:"namespace"` // proto: namespace
	Resource  string `bun:"resource"         json:"resource"`  // proto: resource
	Action    string `bun:"action"           json:"action"`    // proto: action

	// Scope & versioning
	Scope   string `bun:"scope"            json:"scope"`   // proto: scope  ("*" or pattern)
	Version int32  `bun:"version"          json:"version"` // proto: version

	// Optional / future-multi-tenant
	TenantID    *string `bun:"tenant_id,nullzero"   json:"tenant_id,omitempty"` // proto: tenant_id
	Description *string `bun:"description,nullzero" json:"description,omitempty"`

	// Host / tenant semantics
	Side PermissionSide `bun:"side"            json:"side"` // proto: side (enum)

	// House-keeping
	CreatedAt time.Time `bun:"created_at,default:now()" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,default:now()" json:"updated_at"`
}

type PermissionDefGroup struct {
	Name        string           `db:"name"`
	DisplayName string           `db:"display_name"`
	Side        PermissionSide   `db:"side"`
	Priority    int32            `db:"priority"`
	Metadata    map[string]any   `db:"metadata"` // JSONB
	Defs        []*PermissionDef // not persisted as a whole, joined
	CreatedAt   time.Time        `db:"created_at"`
	UpdatedAt   time.Time        `db:"updated_at"`
}
