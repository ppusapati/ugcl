package models

import "time"

type UserPermission struct {
	UserID       string    `json:"user_id"`
	PermissionID string    `json:"permission_id"`
	Granted      bool      `json:"granted"`
	CreatedAt    time.Time `json:"created_at"`
}
