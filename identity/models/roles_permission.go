package models

import "time"

type RolePermission struct {
	RoleID       string    `json:"role_id"`       // FK to roles.id
	PermissionID string    `json:"permission_id"` // FK to permissions.id (UUID)
	Granted      bool      `json:"granted"`       // true = allowed, false = denied
	CreatedAt    time.Time `json:"created_at"`
}
