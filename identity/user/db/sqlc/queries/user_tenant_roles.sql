-- name: CreateUserTenantRole :exec
INSERT INTO user_tenant_roles (
    user_id, tenant_id, roles, is_active
) VALUES (
    $1, $2, $3, $4
);

-- name: GetUserTenantRoles :many
SELECT * FROM user_tenant_roles
WHERE user_id = $1 AND is_active = true;

-- name: GetUserRolesByTenant :one
SELECT * FROM user_tenant_roles
WHERE user_id = $1 AND tenant_id = $2;

-- name: UpdateUserTenantRoles :exec
UPDATE user_tenant_roles
SET roles = $3, updated_at = now()
WHERE user_id = $1 AND tenant_id = $2;

-- name: DeactivateUserTenantRole :exec
UPDATE user_tenant_roles
SET is_active = false, updated_at = now()
WHERE user_id = $1 AND tenant_id = $2;

-- name: GetTenantUsers :many
SELECT utr.*, u.username, u.email, u.fullname
FROM user_tenant_roles utr
JOIN users u ON u.id = utr.user_id
WHERE utr.tenant_id = $1 AND utr.is_active = true;

-- name: DeleteUserTenantRole :exec
DELETE FROM user_tenant_roles
WHERE user_id = $1 AND tenant_id = $2;