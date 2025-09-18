package models

import "time"

type UserRole struct {
	UserID    string    `db:"user_id" bun:"user_id"`
	RoleID    string    `db:"role_id" bun:"role_id"`
	CreatedAt time.Time `db:"created_at" bun:",nullzero,default:current_timestamp"`
}
