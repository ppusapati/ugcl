-- Tenant Metadata Queries

-- name: CreateTenantMetadata :one
INSERT INTO tenant_metadata (
    tenant_id, data, schema_version
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetTenantMetadata :one
SELECT * FROM tenant_metadata
WHERE tenant_id = $1;

-- name: UpdateTenantMetadata :one
UPDATE tenant_metadata
SET data = $2, schema_version = $3
WHERE tenant_id = $1
RETURNING *;

-- name: UpsertTenantMetadata :one
INSERT INTO tenant_metadata (tenant_id, data, schema_version)
VALUES ($1, $2, $3)
ON CONFLICT (tenant_id)
DO UPDATE SET data = $2, schema_version = $3
RETURNING *;

-- name: DeleteTenantMetadata :exec
DELETE FROM tenant_metadata
WHERE tenant_id = $1;