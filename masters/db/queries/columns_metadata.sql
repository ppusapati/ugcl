-- =============================================================================
-- Columns Metadata Queries
-- =============================================================================

-- name: GetColumnsByTable :many
SELECT
    c.id,
    c.table_id,
    c.column_name,
    c.display_name,
    c.description,
    c.data_type,
    c.max_length,
    c.is_required,
    c.is_primary_key,
    c.is_unique,
    c.default_value,
    c.validation_rules,
    c.example_values,
    c.is_active,
    c.supports_import,
    c.column_order,
    c.created_at,
    c.updated_at,
    c.created_by,
    c.updated_by
FROM columns_metadata c
WHERE c.table_id = $1 AND c.is_active = true
ORDER BY c.column_order, c.display_name;

-- name: GetColumnByID :one
SELECT
    c.id,
    c.table_id,
    c.column_name,
    c.display_name,
    c.description,
    c.data_type,
    c.max_length,
    c.is_required,
    c.is_primary_key,
    c.is_unique,
    c.default_value,
    c.validation_rules,
    c.example_values,
    c.is_active,
    c.supports_import,
    c.column_order,
    c.created_at,
    c.updated_at,
    c.created_by,
    c.updated_by
FROM columns_metadata c
WHERE c.id = $1 AND c.is_active = true;

-- name: GetColumnByTableAndName :one
SELECT
    c.id,
    c.table_id,
    c.column_name,
    c.display_name,
    c.description,
    c.data_type,
    c.max_length,
    c.is_required,
    c.is_primary_key,
    c.is_unique,
    c.default_value,
    c.validation_rules,
    c.example_values,
    c.is_active,
    c.supports_import,
    c.column_order,
    c.created_at,
    c.updated_at,
    c.created_by,
    c.updated_by
FROM columns_metadata c
WHERE c.table_id = $1 AND c.column_name = $2 AND c.is_active = true;

-- name: GetImportableColumnsByTable :many
SELECT
    c.id,
    c.table_id,
    c.column_name,
    c.display_name,
    c.description,
    c.data_type,
    c.max_length,
    c.is_required,
    c.is_primary_key,
    c.is_unique,
    c.default_value,
    c.validation_rules,
    c.example_values,
    c.is_active,
    c.supports_import,
    c.column_order,
    c.created_at,
    c.updated_at,
    c.created_by,
    c.updated_by
FROM columns_metadata c
WHERE c.table_id = $1 AND c.is_active = true AND c.supports_import = true
ORDER BY c.column_order, c.display_name;

-- name: GetRequiredColumnsByTable :many
SELECT
    c.id,
    c.table_id,
    c.column_name,
    c.display_name,
    c.description,
    c.data_type,
    c.max_length,
    c.is_required,
    c.is_primary_key,
    c.is_unique,
    c.default_value,
    c.validation_rules,
    c.example_values,
    c.is_active,
    c.supports_import,
    c.column_order,
    c.created_at,
    c.updated_at,
    c.created_by,
    c.updated_by
FROM columns_metadata c
WHERE c.table_id = $1 AND c.is_active = true AND c.is_required = true AND c.supports_import = true
ORDER BY c.column_order, c.display_name;

-- name: CreateColumn :one
INSERT INTO columns_metadata (
    table_id,
    column_name,
    display_name,
    description,
    data_type,
    max_length,
    is_required,
    is_primary_key,
    is_unique,
    default_value,
    validation_rules,
    example_values,
    supports_import,
    column_order,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
) RETURNING
    id,
    table_id,
    column_name,
    display_name,
    description,
    data_type,
    max_length,
    is_required,
    is_primary_key,
    is_unique,
    default_value,
    validation_rules,
    example_values,
    is_active,
    supports_import,
    column_order,
    created_at,
    updated_at,
    created_by,
    updated_by;

-- name: UpdateColumn :one
UPDATE columns_metadata
SET
    display_name = $2,
    description = $3,
    data_type = $4,
    max_length = $5,
    is_required = $6,
    is_unique = $7,
    default_value = $8,
    validation_rules = $9,
    example_values = $10,
    supports_import = $11,
    column_order = $12,
    updated_at = NOW(),
    updated_by = $13
WHERE id = $1 AND is_active = true
RETURNING
    id,
    table_id,
    column_name,
    display_name,
    description,
    data_type,
    max_length,
    is_required,
    is_primary_key,
    is_unique,
    default_value,
    validation_rules,
    example_values,
    is_active,
    supports_import,
    column_order,
    created_at,
    updated_at,
    created_by,
    updated_by;

-- name: DeactivateColumn :exec
UPDATE columns_metadata
SET
    is_active = false,
    updated_at = NOW(),
    updated_by = $2
WHERE id = $1;

-- name: ReorderColumns :exec
UPDATE columns_metadata
SET
    column_order = $2,
    updated_at = NOW(),
    updated_by = $3
WHERE id = $1 AND is_active = true;