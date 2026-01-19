-- Tenant Database Schemas Queries

-- name: CreateTenantDatabaseSchema :one
INSERT INTO tenant_database_schemas (
    tenant_id, schema_name, database_name, connection_string, migration_version
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetTenantDatabaseSchema :one
SELECT * FROM tenant_database_schemas
WHERE tenant_id = $1 AND is_active = true;

-- name: UpdateTenantDatabaseSchema :one
UPDATE tenant_database_schemas
SET schema_name = $2, database_name = $3, connection_string = $4
WHERE tenant_id = $1
RETURNING *;

-- name: UpdateTenantMigrationVersion :exec
UPDATE tenant_database_schemas
SET migration_version = $2, last_migration_at = CURRENT_TIMESTAMP
WHERE tenant_id = $1;

-- name: DeactivateTenantDatabaseSchema :exec
UPDATE tenant_database_schemas
SET is_active = false
WHERE tenant_id = $1;

-- name: GetTenantsWithSeparateSchemas :many
SELECT tds.*, t.name as tenant_name, t.display_name
FROM tenant_database_schemas tds
JOIN tenants t ON tds.tenant_id = t.id
WHERE tds.is_active = true AND t.is_active = true
ORDER BY tds.created_at ASC;