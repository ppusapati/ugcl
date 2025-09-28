-- =============================================================================
-- Reports Queries - InsightHub Service
-- =============================================================================

-- name: GetReportsByUser :many
SELECT
    r.id,
    r.name,
    r.description,
    r.category,
    r.tags,
    r.is_template,
    r.is_active,
    r.created_by,
    r.created_at,
    r.updated_by,
    r.updated_at,
    r.version,
    r.metadata
FROM reports r
WHERE r.created_by = $1 AND r.is_active = true
ORDER BY r.updated_at DESC
LIMIT $2 OFFSET $3;

-- name: GetReportByID :one
SELECT
    r.id,
    r.name,
    r.description,
    r.category,
    r.tags,
    r.is_template,
    r.is_active,
    r.created_by,
    r.created_at,
    r.updated_by,
    r.updated_at,
    r.version,
    r.metadata
FROM reports r
WHERE r.id = $1 AND r.is_active = true;

-- name: GetReportWithPermissions :one
SELECT
    r.id,
    r.name,
    r.description,
    r.category,
    r.tags,
    r.is_template,
    r.is_active,
    r.created_by,
    r.created_at,
    r.updated_by,
    r.updated_at,
    r.version,
    r.metadata
FROM reports r
LEFT JOIN report_permissions rp ON r.id = rp.report_id
WHERE r.id = $1 AND r.is_active = true
  AND (
    r.created_by = $2 OR
    (rp.principal_type = 'user' AND rp.principal_id = $2 AND rp.is_active = true) OR
    (rp.principal_type = 'role' AND rp.principal_id = ANY($3::UUID[]) AND rp.is_active = true)
  );

-- name: CreateReport :one
INSERT INTO reports (
    name,
    description,
    category,
    tags,
    is_template,
    created_by,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING
    id,
    name,
    description,
    category,
    tags,
    is_template,
    is_active,
    created_by,
    created_at,
    updated_by,
    updated_at,
    version,
    metadata;

-- name: UpdateReport :one
UPDATE reports
SET
    name = $2,
    description = $3,
    category = $4,
    tags = $5,
    metadata = $6,
    updated_by = $7,
    updated_at = NOW(),
    version = version + 1
WHERE id = $1 AND is_active = true
RETURNING
    id,
    name,
    description,
    category,
    tags,
    is_template,
    is_active,
    created_by,
    created_at,
    updated_by,
    updated_at,
    version,
    metadata;

-- name: DeleteReport :exec
UPDATE reports
SET
    is_active = false,
    updated_by = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: SearchReports :many
SELECT
    r.id,
    r.name,
    r.description,
    r.category,
    r.tags,
    r.is_template,
    r.is_active,
    r.created_by,
    r.created_at,
    r.updated_by,
    r.updated_at,
    r.version,
    r.metadata
FROM reports r
LEFT JOIN report_permissions rp ON r.id = rp.report_id
WHERE r.is_active = true
  AND (
    r.name ILIKE '%' || $1 || '%' OR
    r.description ILIKE '%' || $1 || '%' OR
    r.category ILIKE '%' || $1 || '%' OR
    $1 = ANY(r.tags)
  )
  AND (
    r.created_by = $2 OR
    (rp.principal_type = 'user' AND rp.principal_id = $2 AND rp.is_active = true) OR
    (rp.principal_type = 'role' AND rp.principal_id = ANY($3::UUID[]) AND rp.is_active = true)
  )
ORDER BY r.updated_at DESC
LIMIT $4 OFFSET $5;

-- name: GetReportsByCategory :many
SELECT
    r.id,
    r.name,
    r.description,
    r.category,
    r.tags,
    r.is_template,
    r.is_active,
    r.created_by,
    r.created_at,
    r.updated_by,
    r.updated_at,
    r.version,
    r.metadata
FROM reports r
WHERE r.category = $1 AND r.is_active = true
ORDER BY r.name
LIMIT $2 OFFSET $3;

-- name: GetReportTemplates :many
SELECT
    r.id,
    r.name,
    r.description,
    r.category,
    r.tags,
    r.is_template,
    r.is_active,
    r.created_by,
    r.created_at,
    r.updated_by,
    r.updated_at,
    r.version,
    r.metadata
FROM reports r
WHERE r.is_template = true AND r.is_active = true
ORDER BY r.category, r.name;

-- name: CloneReport :one
INSERT INTO reports (
    name,
    description,
    category,
    tags,
    is_template,
    created_by,
    metadata
)
SELECT
    $2 as name, -- New name
    r.description,
    r.category,
    r.tags,
    false as is_template, -- Clones are not templates
    $3 as created_by, -- New creator
    r.metadata
FROM reports r
WHERE r.id = $1 AND r.is_active = true
RETURNING
    id,
    name,
    description,
    category,
    tags,
    is_template,
    is_active,
    created_by,
    created_at,
    updated_by,
    updated_at,
    version,
    metadata;