-- Tenant Audit Log Queries

-- name: CreateTenantAuditEntry :one
INSERT INTO tenant_audit_log (
    tenant_id, action, actor_id, actor_type, changes, ip_address, user_agent
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetTenantAuditLog :many
SELECT * FROM tenant_audit_log
WHERE tenant_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetTenantAuditLogByAction :many
SELECT * FROM tenant_audit_log
WHERE tenant_id = $1 AND action = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: GetTenantAuditLogByActor :many
SELECT * FROM tenant_audit_log
WHERE tenant_id = $1 AND actor_id = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: GetTenantAuditLogInPeriod :many
SELECT * FROM tenant_audit_log
WHERE tenant_id = $1 AND created_at BETWEEN $2 AND $3
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: GetAllTenantsAuditLog :many
SELECT tal.*, t.name as tenant_name, t.display_name
FROM tenant_audit_log tal
JOIN tenants t ON tal.tenant_id = t.id
WHERE tal.created_at BETWEEN $1 AND $2
ORDER BY tal.created_at DESC
LIMIT $3 OFFSET $4;

-- name: DeleteOldAuditEntries :exec
DELETE FROM tenant_audit_log
WHERE created_at <= $1;