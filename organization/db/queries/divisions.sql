-- name: CreateDivision :one
INSERT INTO divisions (
    tenant_id,
    code,
    name,
    description,
    head_user_id,
    is_active,
    display_order,
    metadata,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetDivision :one
SELECT * FROM divisions
WHERE id = $1 AND is_active = true
LIMIT 1;

-- name: GetDivisionByCode :one
SELECT * FROM divisions
WHERE tenant_id = $1 AND code = $2 AND is_active = true
LIMIT 1;

-- name: ListDivisions :many
SELECT * FROM divisions
WHERE tenant_id = $1
  AND ($2::boolean IS NULL OR is_active = $2)
ORDER BY display_order ASC, name ASC
LIMIT $3 OFFSET $4;

-- name: CountDivisions :one
SELECT COUNT(*) FROM divisions
WHERE tenant_id = $1
  AND ($2::boolean IS NULL OR is_active = $2);

-- name: UpdateDivision :one
UPDATE divisions
SET
    code = COALESCE(sqlc.narg('code'), code),
    name = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    head_user_id = COALESCE(sqlc.narg('head_user_id'), head_user_id),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    display_order = COALESCE(sqlc.narg('display_order'), display_order),
    metadata = COALESCE(sqlc.narg('metadata'), metadata),
    updated_by = sqlc.narg('updated_by')
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeleteDivision :exec
UPDATE divisions
SET is_active = false,
    updated_by = $2
WHERE id = $1;

-- name: HardDeleteDivision :exec
DELETE FROM divisions WHERE id = $1;

-- name: GetDivisionSummary :many
SELECT * FROM division_summary
WHERE tenant_id = $1
  AND ($2::boolean IS NULL OR is_active = $2)
ORDER BY display_order ASC, name ASC;
