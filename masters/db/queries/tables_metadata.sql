-- =============================================================================
-- Tables Metadata Queries
-- =============================================================================

-- name: GetTablesBySchema :many
SELECT
    t.id,
    t.schema_id,
    t.table_name,
    t.display_name,
    t.description,
    t.is_active,
    t.supports_import,
    t.created_at,
    t.updated_at,
    t.created_by,
    t.updated_by,
    s.schema_name,
    s.display_name as schema_display_name
FROM tables_metadata t
JOIN schemas_metadata s ON t.schema_id = s.id
WHERE t.schema_id = $1 AND t.is_active = true AND s.is_active = true
ORDER BY t.display_name;

-- name: GetTableByID :one
SELECT
    t.id,
    t.schema_id,
    t.table_name,
    t.display_name,
    t.description,
    t.is_active,
    t.supports_import,
    t.created_at,
    t.updated_at,
    t.created_by,
    t.updated_by,
    s.schema_name,
    s.display_name as schema_display_name
FROM tables_metadata t
JOIN schemas_metadata s ON t.schema_id = s.id
WHERE t.id = $1 AND t.is_active = true AND s.is_active = true;

-- name: GetTableBySchemaAndName :one
SELECT
    t.id,
    t.schema_id,
    t.table_name,
    t.display_name,
    t.description,
    t.is_active,
    t.supports_import,
    t.created_at,
    t.updated_at,
    t.created_by,
    t.updated_by,
    s.schema_name,
    s.display_name as schema_display_name
FROM tables_metadata t
JOIN schemas_metadata s ON t.schema_id = s.id
WHERE t.schema_id = $1 AND t.table_name = $2 AND t.is_active = true AND s.is_active = true;

-- name: GetAllImportableTables :many
SELECT
    t.id,
    t.schema_id,
    t.table_name,
    t.display_name,
    t.description,
    t.is_active,
    t.supports_import,
    t.created_at,
    t.updated_at,
    t.created_by,
    t.updated_by,
    s.schema_name,
    s.display_name as schema_display_name
FROM tables_metadata t
JOIN schemas_metadata s ON t.schema_id = s.id
WHERE t.is_active = true AND t.supports_import = true AND s.is_active = true
ORDER BY s.display_name, t.display_name;

-- name: CreateTable :one
INSERT INTO tables_metadata (
    schema_id,
    table_name,
    display_name,
    description,
    supports_import,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING
    id,
    schema_id,
    table_name,
    display_name,
    description,
    is_active,
    supports_import,
    created_at,
    updated_at,
    created_by,
    updated_by;

-- name: UpdateTable :one
UPDATE tables_metadata
SET
    display_name = $2,
    description = $3,
    supports_import = $4,
    updated_at = NOW(),
    updated_by = $5
WHERE id = $1 AND is_active = true
RETURNING
    id,
    schema_id,
    table_name,
    display_name,
    description,
    is_active,
    supports_import,
    created_at,
    updated_at,
    created_by,
    updated_by;

-- name: DeactivateTable :exec
UPDATE tables_metadata
SET
    is_active = false,
    updated_at = NOW(),
    updated_by = $2
WHERE id = $1;

-- name: SearchTables :many
SELECT
    t.id,
    t.schema_id,
    t.table_name,
    t.display_name,
    t.description,
    t.is_active,
    t.supports_import,
    t.created_at,
    t.updated_at,
    t.created_by,
    t.updated_by,
    s.schema_name,
    s.display_name as schema_display_name
FROM tables_metadata t
JOIN schemas_metadata s ON t.schema_id = s.id
WHERE t.is_active = true AND s.is_active = true
  AND (
    t.display_name ILIKE '%' || $1 || '%' OR
    t.description ILIKE '%' || $1 || '%' OR
    t.table_name ILIKE '%' || $1 || '%'
  )
ORDER BY s.display_name, t.display_name;