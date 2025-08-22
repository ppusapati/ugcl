-- name: CreateRole :exec
INSERT INTO roles (
    id,
    parent_id,
    name,
    is_preserved,
    metadata
) VALUES (
    @id,                    -- TEXT
    @parent_id,             -- NULLABLE TEXT
    @name,                  -- TEXT
    COALESCE(@is_preserved, FALSE),  -- BOOL with default
    @metadata               -- NULLABLE JSONB
);

-- name: GetRoleByID :one
SELECT * FROM roles WHERE id = $1;

-- name: UpdateRole :exec
UPDATE roles
SET
    parent_id    = $2,
    name         = $3,
    is_preserved = $4,
    metadata     = $5,
    updated_at   = now()
WHERE id = $1;

-- name: DeleteRole :exec
DELETE FROM roles WHERE id = $1;

-- name: GetRolesByUserUUID :many
SELECT r.*
FROM roles r
JOIN user_roles ur ON ur.role_id = r.id
WHERE ur.user_uuid = $1;

-- name: ListRoles :many
SELECT * FROM roles
ORDER BY name;
