-- Auth User Tenants Queries

-- name: CreateAuthUserTenant :one
INSERT INTO auth_user_tenants (
    user_id, tenant_id, is_default, is_active, roles
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetUserTenants :many
SELECT * FROM auth_user_tenants
WHERE user_id = $1 AND is_active = true
ORDER BY is_default DESC, joined_at ASC;

-- name: GetUserDefaultTenant :one
SELECT * FROM auth_user_tenants
WHERE user_id = $1 AND is_default = true AND is_active = true;

-- name: GetUserTenantByID :one
SELECT * FROM auth_user_tenants
WHERE user_id = $1 AND tenant_id = $2 AND is_active = true;

-- name: UpdateUserTenantRoles :exec
UPDATE auth_user_tenants
SET roles = $3
WHERE user_id = $1 AND tenant_id = $2;

-- name: SetUserDefaultTenant :exec
UPDATE auth_user_tenants
SET is_default = CASE WHEN tenant_id = $2 THEN true ELSE false END
WHERE user_id = $1;

-- name: DeactivateUserTenant :exec
UPDATE auth_user_tenants
SET is_active = false
WHERE user_id = $1 AND tenant_id = $2;

-- name: GetTenantUsers :many
SELECT * FROM auth_user_tenants
WHERE tenant_id = $1 AND is_active = true
ORDER BY joined_at ASC;