-- Tenant Features Queries

-- name: CreateTenantFeature :one
INSERT INTO tenant_features (
    tenant_id, key, value, value_type, description
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetTenantFeature :one
SELECT * FROM tenant_features
WHERE tenant_id = $1 AND key = $2;

-- name: GetTenantFeatures :many
SELECT * FROM tenant_features
WHERE tenant_id = $1
ORDER BY key ASC;

-- name: UpdateTenantFeature :one
UPDATE tenant_features
SET value = $3, value_type = $4, description = $5
WHERE tenant_id = $1 AND key = $2
RETURNING *;

-- name: DeleteTenantFeature :exec
DELETE FROM tenant_features
WHERE tenant_id = $1 AND key = $2;

-- name: UpsertTenantFeature :one
INSERT INTO tenant_features (tenant_id, key, value, value_type, description)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (tenant_id, key)
DO UPDATE SET value = $3, value_type = $4, description = $5
RETURNING *;