-- name: CreatePermissionDef :one
INSERT INTO permission_defs (
    name,
    description,
    side,
    metadata,
    namespace,
    resource,
    action,
    scope,
    version
) VALUES (
    @name,                         -- TEXT
    @description,                  -- NULLABLE TEXT
    @side,                         -- NULLABLE INT
    @metadata,                     -- NULLABLE JSONB
    @namespace,                    -- TEXT
    @resource,                     -- TEXT
    @action,                       -- TEXT
    COALESCE(@scope,  '*'),        -- default to "*"
    COALESCE(@version, 1)          -- default to 1
)
RETURNING *;

-- name: GetPermissionDefByName :one
SELECT * FROM permission_defs WHERE name = $1;

-- name: GetPermissionDefByID :one
SELECT * FROM permission_defs WHERE id = $1;

-- name: ListAllPermissionDefs :many
SELECT * FROM permission_defs;

-- name: ListPermissionDefGroups :many
SELECT DISTINCT ON (namespace) *
FROM permission_defs
ORDER BY namespace, name;

-- name: DeletePermissionDef :exec
DELETE FROM permission_defs WHERE name = $1;

-- name: ListDefsByGroup :many
SELECT pd.*
FROM permission_defs pd
JOIN permission_def_groups map ON map.permission_def_name = pd.name
WHERE map.name = $1;

-- name: UpdatePermissionDef :exec
UPDATE permission_defs
SET name = $2,
    side         = $3,
    metadata     = $4,
    namespace    = $5,
    resource     = $6,
    action       = $7,
    scope        = COALESCE($8, scope),
    version      = version + 1,        -- bump version on every update
    updated_at   = now()
WHERE name = $1;

-- name: ListPermissionDefsByRole :many
SELECT pd.*
FROM permission_defs pd
JOIN role_permission_defs rpd ON pd.name = rpd.permission_def_name
WHERE rpd.role_id = $1;
