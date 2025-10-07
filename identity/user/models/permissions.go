package models

import (
	"time"
)

type Effect int32

const (
	EffectUnknown   Effect = 0
	EffectGrant     Effect = 1
	EffectForbidden Effect = 2
)

type Permission struct {
	Namespace string    `db:"namespace"`
	Resource  string    `db:"resource"`
	Action    string    `db:"action"`
	Subject   string    `db:"subject"`
	Effect    Effect    `db:"effect"`
	TenantID  string    `db:"tenant_id"`
	DefName   string    `db:"def_name"` // Link to permission definition
	Granted   bool      `db:"granted"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`

	// Enhanced fields for organizational scope
	DivisionID   *string `db:"division_id"`
	BranchID     *string `db:"branch_id"`
	DepartmentID *string `db:"department_id"`

	// Resource instance scoping
	ResourceID *string `db:"resource_id"`

	// Temporal constraints
	ValidFrom *time.Time `db:"valid_from"`
	ValidUntil *time.Time `db:"valid_until"`

	// Inheritance control
	AllowInheritance *bool `db:"allow_inheritance"`
}
