-- Table queries
-- name: GetAllDairySites :many
SELECT *
FROM dairy_sites
WHERE deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetDairySiteByID :one
SELECT *
FROM dairy_sites
WHERE id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: GetDairySitesByEmployeeID :many
SELECT *
FROM dairy_sites
WHERE employee_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: CreateDairySite :one
INSERT INTO dairy_sites (
    name_of_site,
    todays_work,
    employee_id,
    latitude,
    longitude,
    submitted_at
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: UpdateDairySite :one
UPDATE dairy_sites
SET
    name_of_site = $2,
    todays_work = $3,
    employee_id = $4,
    latitude = $5,
    longitude = $6,
    submitted_at = $7,
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteDairySite :exec
UPDATE dairy_sites
SET deleted_at = now()
WHERE id = $1 AND deleted_at IS NULL;

-- View queries
-- name: GetAllDairySitesWithUser :many
SELECT *
FROM dairy_sites_with_user
ORDER BY created_at DESC;

-- name: GetDairySiteWithUserByID :one
SELECT *
FROM dairy_sites_with_user
WHERE id = $1
LIMIT 1;

-- -- name: GetDairySitesWithUserByEmployeeID :many
-- SELECT *
-- FROM dairy_sites_with_user
-- WHERE employee_id = $1
-- ORDER BY created_at DESC;