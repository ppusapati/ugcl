-- =============================================================================
-- Schema Metadata Queries
-- =============================================================================

-- name: GetAllSchemas :many
SELECT
    id,
    schema_name,
    display_name,
    description,
    is_active,
    created_at,
    updated_at,
    created_by,
    updated_by
FROM schemas_metadata
WHERE is_active = true
ORDER BY display_name;

-- name: GetSchemaByID :one
SELECT
    id,
    schema_name,
    display_name,
    description,
    is_active,
    created_at,
    updated_at,
    created_by,
    updated_by
FROM schemas_metadata
WHERE id = $1 AND is_active = true;

-- name: GetSchemaByName :one
SELECT
    id,
    schema_name,
    display_name,
    description,
    is_active,
    created_at,
    updated_at,
    created_by,
    updated_by
FROM schemas_metadata
WHERE schema_name = $1 AND is_active = true;

-- name: CreateSchema :one
INSERT INTO schemas_metadata (
    schema_name,
    display_name,
    description,
    created_by
) VALUES (
    $1, $2, $3, $4
) RETURNING
    id,
    schema_name,
    display_name,
    description,
    is_active,
    created_at,
    updated_at,
    created_by,
    updated_by;

-- name: UpdateSchema :one
UPDATE schemas_metadata
SET
    display_name = $2,
    description = $3,
    updated_at = NOW(),
    updated_by = $4
WHERE id = $1 AND is_active = true
RETURNING
    id,
    schema_name,
    display_name,
    description,
    is_active,
    created_at,
    updated_at,
    created_by,
    updated_by;

-- name: DeactivateSchema :exec
UPDATE schemas_metadata
SET
    is_active = false,
    updated_at = NOW(),
    updated_by = $2
WHERE id = $1;

-- name: SearchSchemas :many
SELECT
    id,
    schema_name,
    display_name,
    description,
    is_active,
    created_at,
    updated_at,
    created_by,
    updated_by
FROM schemas_metadata
WHERE is_active = true
  AND (
    display_name ILIKE '%' || $1 || '%' OR
    description ILIKE '%' || $1 || '%' OR
    schema_name ILIKE '%' || $1 || '%'
  )
ORDER BY display_name;