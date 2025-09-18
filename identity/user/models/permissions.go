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
	Granted   bool      `db:"granted"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
