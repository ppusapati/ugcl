-- name: CreateEntity :one
INSERT INTO entities (
    tenant_id,
    entity_type,
    user_id,
    reference_id,
    reference_source,
    status,
    division_id,
    branch_id,
    department_id,
    metadata,
    created_by,
    updated_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING *;

-- name: UpdateEntity :one
UPDATE entities
SET
    status = COALESCE(sqlc.narg('status'), status),
    division_id = COALESCE(sqlc.narg('division_id'), division_id),
    branch_id = COALESCE(sqlc.narg('branch_id'), branch_id),
    department_id = COALESCE(sqlc.narg('department_id'), department_id),
    metadata = COALESCE(sqlc.narg('metadata'), metadata),
    updated_at = NOW(),
    updated_by = $1
WHERE id = $2 AND tenant_id = $3
RETURNING *;

-- name: GetEntity :one
SELECT * FROM entities
WHERE id = $1 AND tenant_id = $2;

-- name: GetEntityByUserId :one
SELECT * FROM entities
WHERE user_id = $1 AND tenant_id = $2;

-- name: GetEntityByReference :one
SELECT * FROM entities
WHERE reference_id = $1
  AND reference_source = $2
  AND tenant_id = $3;

-- name: ListEntities :many
SELECT * FROM entities
WHERE tenant_id = sqlc.arg('tenant_id')
  AND (sqlc.arg('entity_type')::entity_type IS NULL OR entity_type = sqlc.arg('entity_type'))
  AND (sqlc.arg('status')::entity_status IS NULL OR status = sqlc.arg('status'))
  AND (sqlc.arg('division_id')::uuid IS NULL OR division_id = sqlc.arg('division_id'))
  AND (sqlc.arg('branch_id')::uuid IS NULL OR branch_id = sqlc.arg('branch_id'))
  AND (sqlc.arg('department_id')::uuid IS NULL OR department_id = sqlc.arg('department_id'))
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountEntities :one
SELECT COUNT(*) FROM entities
WHERE tenant_id = $1
  AND (sqlc.arg('entity_type')::entity_type IS NULL OR entity_type = sqlc.arg('entity_type'))
  AND (sqlc.arg('status')::entity_status IS NULL OR status = sqlc.arg('status'))
  AND (sqlc.arg('division_id')::uuid IS NULL OR division_id = sqlc.arg('division_id'))
  AND (sqlc.arg('branch_id')::uuid IS NULL OR branch_id = sqlc.arg('branch_id'))
  AND (sqlc.arg('department_id')::uuid IS NULL OR department_id = sqlc.arg('department_id'));

-- name: DeleteEntity :exec
DELETE FROM entities
WHERE id = $1 AND tenant_id = $2;

-- name: CreateEntityRoleBinding :one
INSERT INTO entity_role_bindings (
    tenant_id,
    entity_id,
    role_id,
    division_id,
    branch_id,
    department_id,
    valid_from,
    valid_until,
    created_by,
    updated_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: GetEntityRoleBindings :many
SELECT * FROM entity_role_bindings
WHERE entity_id = $1 AND tenant_id = $2
ORDER BY created_at DESC;

-- name: GetActiveEntityRoleBindings :many
SELECT * FROM entity_role_bindings
WHERE entity_id = $1
  AND tenant_id = $2
  AND (valid_from IS NULL OR valid_from <= NOW())
  AND (valid_until IS NULL OR valid_until >= NOW())
ORDER BY created_at DESC;

-- name: DeleteEntityRoleBinding :exec
DELETE FROM entity_role_bindings
WHERE id = $1 AND tenant_id = $2;

-- name: DeleteAllEntityRoleBindings :exec
DELETE FROM entity_role_bindings
WHERE entity_id = $1 AND tenant_id = $2;
