-- Tenant Admin Users Queries

-- name: CreateTenantAdminUser :one
INSERT INTO tenant_admin_users (
    tenant_id, user_id, username, email, is_primary
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetTenantAdminUsers :many
SELECT * FROM tenant_admin_users
WHERE tenant_id = $1
ORDER BY is_primary DESC, created_at ASC;

-- name: GetTenantPrimaryAdmin :one
SELECT * FROM tenant_admin_users
WHERE tenant_id = $1 AND is_primary = true;

-- name: GetUserTenantAdminRoles :many
SELECT * FROM tenant_admin_users
WHERE user_id = $1;

-- name: UpdateTenantAdminUser :one
UPDATE tenant_admin_users
SET username = $3, email = $4, is_primary = $5
WHERE tenant_id = $1 AND user_id = $2
RETURNING *;

-- name: SetPrimaryAdmin :exec
UPDATE tenant_admin_users
SET is_primary = CASE WHEN user_id = $2 THEN true ELSE false END
WHERE tenant_id = $1;

-- name: DeleteTenantAdminUser :exec
DELETE FROM tenant_admin_users
WHERE tenant_id = $1 AND user_id = $2;