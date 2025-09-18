package models

import "time"

type RolePermissionDef struct {
	RoleID            string    `db:"role_id" bun:"role_id"`
	PermissionDefName string    `db:"permission_def_name" bun:"permission_def_name"`
	CreatedAt         time.Time `db:"created_at" bun:",nullzero,default:current_timestamp"`
}
