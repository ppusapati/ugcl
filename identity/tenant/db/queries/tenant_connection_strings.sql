-- Tenant Connection Strings Queries

-- name: CreateTenantConnectionString :one
INSERT INTO tenant_connection_strings (
    tenant_id, key, value, is_encrypted
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetTenantConnectionString :one
SELECT * FROM tenant_connection_strings
WHERE tenant_id = $1 AND key = $2;

-- name: GetTenantConnectionStrings :many
SELECT * FROM tenant_connection_strings
WHERE tenant_id = $1
ORDER BY key ASC;

-- name: UpdateTenantConnectionString :one
UPDATE tenant_connection_strings
SET value = $3, is_encrypted = $4
WHERE tenant_id = $1 AND key = $2
RETURNING *;

-- name: DeleteTenantConnectionString :exec
DELETE FROM tenant_connection_strings
WHERE tenant_id = $1 AND key = $2;

-- name: UpsertTenantConnectionString :one
INSERT INTO tenant_connection_strings (tenant_id, key, value, is_encrypted)
VALUES ($1, $2, $3, $4)
ON CONFLICT (tenant_id, key)
DO UPDATE SET value = $3, is_encrypted = $4
RETURNING *;