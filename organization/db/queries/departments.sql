-- name: CreateDepartment :one
INSERT INTO departments (
    tenant_id,
    division_id,
    code,
    name,
    description,
    department_type,
    head_user_id,
    parent_department_id,
    is_active,
    display_order,
    metadata,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING *;

-- name: GetDepartment :one
SELECT * FROM departments
WHERE id = $1 AND is_active = true
LIMIT 1;

-- name: GetDepartmentByCode :one
SELECT * FROM departments
WHERE tenant_id = $1 AND code = $2 AND is_active = true
LIMIT 1;

-- name: ListDepartments :many
SELECT * FROM departments
WHERE tenant_id = $1
  AND ($2::boolean IS NULL OR is_active = $2)
ORDER BY display_order ASC, name ASC
LIMIT $3 OFFSET $4;

-- name: ListDepartmentsByDivision :many
SELECT * FROM departments
WHERE division_id = $1
  AND ($2::boolean IS NULL OR is_active = $2)
ORDER BY display_order ASC, name ASC;

-- name: ListBusinessLevelDepartments :many
SELECT * FROM departments
WHERE tenant_id = $1
  AND division_id IS NULL
  AND ($2::boolean IS NULL OR is_active = $2)
ORDER BY display_order ASC, name ASC;

-- name: ListSubDepartments :many
SELECT * FROM departments
WHERE parent_department_id = $1
  AND ($2::boolean IS NULL OR is_active = $2)
ORDER BY display_order ASC, name ASC;

-- name: CountDepartments :one
SELECT COUNT(*) FROM departments
WHERE tenant_id = $1
  AND ($2::boolean IS NULL OR is_active = $2);

-- name: CountDepartmentsByDivision :one
SELECT COUNT(*) FROM departments
WHERE division_id = $1
  AND ($2::boolean IS NULL OR is_active = $2);

-- name: UpdateDepartment :one
UPDATE departments
SET
    code = COALESCE(sqlc.narg('code'), code),
    name = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    department_type = COALESCE(sqlc.narg('department_type'), department_type),
    head_user_id = COALESCE(sqlc.narg('head_user_id'), head_user_id),
    parent_department_id = COALESCE(sqlc.narg('parent_department_id'), parent_department_id),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    display_order = COALESCE(sqlc.narg('display_order'), display_order),
    metadata = COALESCE(sqlc.narg('metadata'), metadata),
    updated_by = sqlc.narg('updated_by')
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeleteDepartment :exec
UPDATE departments
SET is_active = false,
    updated_by = $2
WHERE id = $1;

-- name: HardDeleteDepartment :exec
DELETE FROM departments WHERE id = $1;

-- name: GetDepartmentHierarchy :many
SELECT * FROM department_hierarchy
WHERE tenant_id = $1
  AND ($2::boolean IS NULL OR is_active = $2)
ORDER BY level ASC, display_order ASC, name ASC;

-- name: GetOrganizationFullHierarchy :many
SELECT * FROM organization_full_hierarchy;
