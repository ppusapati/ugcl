-- =============================================================================
-- Import Mappings Queries - DataBridge Module
-- =============================================================================

-- name: GetMappingsByTable :many
SELECT
    m.id,
    m.mapping_name,
    m.description,
    m.table_id,
    m.csv_headers,
    m.field_mappings,
    m.transformation_rules,
    m.validation_rules,
    m.is_active,
    m.created_at,
    m.updated_at,
    m.created_by,
    m.updated_by,
    m.last_used_at,
    m.usage_count
FROM import_mappings m
WHERE m.table_id = $1 AND m.is_active = true
ORDER BY m.last_used_at DESC NULLS LAST, m.created_at DESC;

-- name: GetMappingByID :one
SELECT
    m.id,
    m.mapping_name,
    m.description,
    m.table_id,
    m.csv_headers,
    m.field_mappings,
    m.transformation_rules,
    m.validation_rules,
    m.is_active,
    m.created_at,
    m.updated_at,
    m.created_by,
    m.updated_by,
    m.last_used_at,
    m.usage_count
FROM import_mappings m
WHERE m.id = $1 AND m.is_active = true;

-- name: GetMappingByTableAndName :one
SELECT
    m.id,
    m.mapping_name,
    m.description,
    m.table_id,
    m.csv_headers,
    m.field_mappings,
    m.transformation_rules,
    m.validation_rules,
    m.is_active,
    m.created_at,
    m.updated_at,
    m.created_by,
    m.updated_by,
    m.last_used_at,
    m.usage_count
FROM import_mappings m
WHERE m.table_id = $1 AND m.mapping_name = $2 AND m.is_active = true;

-- name: GetRecentMappingsByUser :many
SELECT
    m.id,
    m.mapping_name,
    m.description,
    m.table_id,
    m.csv_headers,
    m.field_mappings,
    m.transformation_rules,
    m.validation_rules,
    m.is_active,
    m.created_at,
    m.updated_at,
    m.created_by,
    m.updated_by,
    m.last_used_at,
    m.usage_count
FROM import_mappings m
WHERE m.created_by = $1 AND m.is_active = true
ORDER BY m.last_used_at DESC NULLS LAST, m.created_at DESC
LIMIT $2;

-- name: CreateMapping :one
INSERT INTO import_mappings (
    mapping_name,
    description,
    table_id,
    csv_headers,
    field_mappings,
    transformation_rules,
    validation_rules,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING
    id,
    mapping_name,
    description,
    table_id,
    csv_headers,
    field_mappings,
    transformation_rules,
    validation_rules,
    is_active,
    created_at,
    updated_at,
    created_by,
    updated_by,
    last_used_at,
    usage_count;

-- name: UpdateMapping :one
UPDATE import_mappings
SET
    mapping_name = $2,
    description = $3,
    csv_headers = $4,
    field_mappings = $5,
    transformation_rules = $6,
    validation_rules = $7,
    updated_at = NOW(),
    updated_by = $8
WHERE id = $1 AND is_active = true
RETURNING
    id,
    mapping_name,
    description,
    table_id,
    csv_headers,
    field_mappings,
    transformation_rules,
    validation_rules,
    is_active,
    created_at,
    updated_at,
    created_by,
    updated_by,
    last_used_at,
    usage_count;

-- name: UpdateMappingUsage :exec
UPDATE import_mappings
SET
    last_used_at = NOW(),
    usage_count = usage_count + 1
WHERE id = $1 AND is_active = true;

-- name: DeactivateMapping :exec
UPDATE import_mappings
SET
    is_active = false,
    updated_at = NOW(),
    updated_by = $2
WHERE id = $1;

-- name: SearchMappings :many
SELECT
    m.id,
    m.mapping_name,
    m.description,
    m.table_id,
    m.csv_headers,
    m.field_mappings,
    m.transformation_rules,
    m.validation_rules,
    m.is_active,
    m.created_at,
    m.updated_at,
    m.created_by,
    m.updated_by,
    m.last_used_at,
    m.usage_count
FROM import_mappings m
WHERE m.is_active = true
  AND (
    m.mapping_name ILIKE '%' || $1 || '%' OR
    m.description ILIKE '%' || $1 || '%'
  )
ORDER BY m.last_used_at DESC NULLS LAST, m.created_at DESC;