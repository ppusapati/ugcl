-- Tenant Management Queries

-- name: CreateTenant :one
INSERT INTO tenants (
    name, display_name, region, logo, tenant_db, created_by,
    subscription_plan, max_users, max_storage_gb
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetTenantByID :one
SELECT * FROM tenants WHERE id = $1 AND is_active = true;

-- name: GetTenantByName :one
SELECT * FROM tenants WHERE name = $1 AND is_active = true;

-- name: UpdateTenant :one
UPDATE tenants
SET display_name = $2, logo = $3, subscription_plan = $4,
    max_users = $5, max_storage_gb = $6
WHERE id = $1
RETURNING *;

-- name: DeactivateTenant :exec
UPDATE tenants SET is_active = false WHERE id = $1;

-- name: ListTenants :many
SELECT * FROM tenants
WHERE is_active = true
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListTenantsByRegion :many
SELECT * FROM tenants
WHERE region = $1 AND is_active = true
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: SearchTenantsByName :many
SELECT * FROM tenants
WHERE (name ILIKE $1 OR display_name ILIKE $1) AND is_active = true
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetTenantSummary :one
SELECT * FROM tenant_summary WHERE id = $1;