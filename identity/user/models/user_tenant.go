package models

import "time"

// UserTenantAssociation represents the relationship between a user and tenant
type UserTenantAssociation struct {
	UserID      string    `json:"user_id" db:"user_id"`
	TenantID    string    `json:"tenant_id" db:"tenant_id"`
	Roles       []string  `json:"roles" db:"roles"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	IsPrimary   bool      `json:"is_primary" db:"is_primary"`
	AssignedAt  time.Time `json:"assigned_at" db:"assigned_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	ValidFrom   *time.Time `json:"valid_from,omitempty" db:"valid_from"`
	ValidUntil  *time.Time `json:"valid_until,omitempty" db:"valid_until"`
	AssignedBy  string    `json:"assigned_by" db:"assigned_by"`
	Notes       *string   `json:"notes,omitempty" db:"notes"`
}

// UserTenantHistoryEntry represents a historical record of user-tenant relationship changes
type UserTenantHistoryEntry struct {
	ID           string                 `json:"id" db:"id"`
	UserID       string                 `json:"user_id" db:"user_id"`
	TenantID     string                 `json:"tenant_id" db:"tenant_id"`
	Action       UserTenantAction       `json:"action" db:"action"`
	OldRoles     []string               `json:"old_roles" db:"old_roles"`
	NewRoles     []string               `json:"new_roles" db:"new_roles"`
	Timestamp    time.Time              `json:"timestamp" db:"timestamp"`
	PerformedBy  string                 `json:"performed_by" db:"performed_by"`
	Reason       *string                `json:"reason,omitempty" db:"reason"`
	Notes        *string                `json:"notes,omitempty" db:"notes"`
}

// UserTenantAction represents the type of action performed on user-tenant relationship
type UserTenantAction string

const (
	UserTenantActionAssigned      UserTenantAction = "assigned"
	UserTenantActionRemoved       UserTenantAction = "removed"
	UserTenantActionRolesUpdated  UserTenantAction = "roles_updated"
	UserTenantActionActivated     UserTenantAction = "activated"
	UserTenantActionDeactivated   UserTenantAction = "deactivated"
	UserTenantActionTransferred   UserTenantAction = "transferred"
	UserTenantActionPrimaryChanged UserTenantAction = "primary_changed"
)