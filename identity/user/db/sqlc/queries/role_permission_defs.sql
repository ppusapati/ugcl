-- name: AssignRolePermissionDef :exec
INSERT INTO role_permission_defs (role_id, permission_def_name)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveRolePermissionDef :exec
DELETE FROM role_permission_defs
WHERE role_id = $1 AND permission_def_name = $2;

-- name: GetPermissionDefsByRoleID :many
SELECT pd.*
FROM permission_defs pd
JOIN role_permission_defs rpd ON rpd.permission_def_name = pd.name
WHERE rpd.role_id = $1;

-- name: AssignPermissionToRole :exec
-- (alias kept for backward-compat)
INSERT INTO role_permission_defs (role_id, permission_def_name)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: ListPermissionsForRole :many
SELECT p.*
FROM permissions p
JOIN role_permissions rp ON rp.permission_id = p.id
WHERE rp.role_id = $1;
