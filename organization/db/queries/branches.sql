-- name: CreateBranch :one
INSERT INTO branches (
    division_id,
    tenant_id,
    code,
    name,
    branch_type,
    address_line1,
    address_line2,
    city,
    state,
    country,
    postal_code,
    latitude,
    longitude,
    phone,
    email,
    branch_manager_user_id,
    is_active,
    display_order,
    metadata,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
) RETURNING *;

-- name: GetBranch :one
SELECT * FROM branches
WHERE id = $1 AND is_active = true
LIMIT 1;

-- name: GetBranchByCode :one
SELECT * FROM branches
WHERE tenant_id = $1 AND code = $2 AND is_active = true
LIMIT 1;

-- name: ListBranches :many
SELECT * FROM branches
WHERE tenant_id = $1
  AND ($2::boolean IS NULL OR is_active = $2)
ORDER BY display_order ASC, name ASC
LIMIT $3 OFFSET $4;

-- name: ListBranchesByDivision :many
SELECT * FROM branches
WHERE division_id = $1
  AND ($2::boolean IS NULL OR is_active = $2)
ORDER BY display_order ASC, name ASC;

-- name: ListBranchesByCity :many
SELECT * FROM branches
WHERE tenant_id = $1
  AND city = $2
  AND is_active = true
ORDER BY name ASC;

-- name: ListBranchesByState :many
SELECT * FROM branches
WHERE tenant_id = $1
  AND state = $2
  AND is_active = true
ORDER BY city ASC, name ASC;

-- name: CountBranches :one
SELECT COUNT(*) FROM branches
WHERE tenant_id = $1
  AND ($2::boolean IS NULL OR is_active = $2);

-- name: CountBranchesByDivision :one
SELECT COUNT(*) FROM branches
WHERE division_id = $1
  AND ($2::boolean IS NULL OR is_active = $2);

-- name: UpdateBranch :one
UPDATE branches
SET
    code = COALESCE(sqlc.narg('code'), code),
    name = COALESCE(sqlc.narg('name'), name),
    branch_type = COALESCE(sqlc.narg('branch_type'), branch_type),
    address_line1 = COALESCE(sqlc.narg('address_line1'), address_line1),
    address_line2 = COALESCE(sqlc.narg('address_line2'), address_line2),
    city = COALESCE(sqlc.narg('city'), city),
    state = COALESCE(sqlc.narg('state'), state),
    country = COALESCE(sqlc.narg('country'), country),
    postal_code = COALESCE(sqlc.narg('postal_code'), postal_code),
    latitude = COALESCE(sqlc.narg('latitude'), latitude),
    longitude = COALESCE(sqlc.narg('longitude'), longitude),
    phone = COALESCE(sqlc.narg('phone'), phone),
    email = COALESCE(sqlc.narg('email'), email),
    branch_manager_user_id = COALESCE(sqlc.narg('branch_manager_user_id'), branch_manager_user_id),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    display_order = COALESCE(sqlc.narg('display_order'), display_order),
    metadata = COALESCE(sqlc.narg('metadata'), metadata),
    updated_by = sqlc.narg('updated_by')
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeleteBranch :exec
UPDATE branches
SET is_active = false,
    updated_by = $2
WHERE id = $1;

-- name: HardDeleteBranch :exec
DELETE FROM branches WHERE id = $1;

-- name: GetBranchSummary :many
SELECT * FROM branch_summary
WHERE tenant_id = $1
  AND ($2::boolean IS NULL OR is_active = $2)
ORDER BY display_order ASC, name ASC;

-- name: GetBranchSummaryByDivision :many
SELECT * FROM branch_summary
WHERE division_id = $1
  AND ($2::boolean IS NULL OR is_active = $2)
ORDER BY display_order ASC, name ASC;
