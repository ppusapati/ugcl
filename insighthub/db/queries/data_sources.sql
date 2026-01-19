-- =============================================================================
-- Data Sources Queries - InsightHub Service
-- =============================================================================

-- name: CreateDataSource :one
INSERT INTO data_sources (
    name,
    description,
    connection_string,
    schema_id,
    created_by
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING
    id,
    name,
    description,
    connection_string,
    schema_id,
    is_active,
    created_by,
    created_at,
    updated_by,
    updated_at;

-- name: GetDataSourceByID :one
SELECT
    id,
    name,
    description,
    connection_string,
    schema_id,
    is_active,
    created_by,
    created_at,
    updated_by,
    updated_at
FROM data_sources
WHERE id = $1 AND is_active = true;

-- name: GetDataSourceByName :one
SELECT
    id,
    name,
    description,
    connection_string,
    schema_id,
    is_active,
    created_by,
    created_at,
    updated_by,
    updated_at
FROM data_sources
WHERE name = $1 AND is_active = true;

-- name: GetActiveDataSources :many
SELECT
    id,
    name,
    description,
    connection_string,
    schema_id,
    is_active,
    created_by,
    created_at,
    updated_by,
    updated_at
FROM data_sources
WHERE is_active = true
ORDER BY name;

-- name: GetDataSourcesBySchema :many
SELECT
    id,
    name,
    description,
    connection_string,
    schema_id,
    is_active,
    created_by,
    created_at,
    updated_by,
    updated_at
FROM data_sources
WHERE schema_id = $1 AND is_active = true
ORDER BY name;

-- name: UpdateDataSource :one
UPDATE data_sources
SET
    name = $2,
    description = $3,
    connection_string = $4,
    updated_by = $5,
    updated_at = NOW()
WHERE id = $1 AND is_active = true
RETURNING
    id,
    name,
    description,
    connection_string,
    schema_id,
    is_active,
    created_by,
    created_at,
    updated_by,
    updated_at;

-- name: DeleteDataSource :exec
UPDATE data_sources
SET
    is_active = false,
    updated_by = $2,
    updated_at = NOW()
WHERE id = $1;