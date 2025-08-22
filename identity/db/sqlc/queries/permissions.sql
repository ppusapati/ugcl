-- name: GetPermissionsBySubject :many
SELECT * FROM permissions WHERE subject = $1;

-- name: CheckPermission :one
SELECT * FROM permissions
WHERE namespace = $1 AND resource = $2 AND action = $3 AND subject = $4;

-- name: DeletePermission :exec
DELETE FROM permissions
WHERE namespace = $1 AND resource = $2 AND action = $3 AND subject = $4;

-- name: ListAllPermissions :many
SELECT * FROM permissions;

-- name: ListEffectivePermissionsForUser :many
SELECT p.*
FROM permissions p
WHERE p.effect = 'allow'
  AND p.id IN (
    SELECT up.permission_id
      FROM user_permissions up
     WHERE up.user_uuid = $1 AND up.granted = true
    UNION
    SELECT rp.permission_id
      FROM role_permissions rp
      JOIN user_roles ur ON ur.role_id = rp.role_id
     WHERE ur.user_uuid = $1 AND rp.granted = true
  );

-- name: GrantPermission :exec
INSERT INTO permissions (
    namespace, resource, action, subject, effect
) VALUES (
    @namespace,                          -- text
    @resource,                           -- text
    @action,                             -- text
    @subject,                            -- text
    @effect::permission_effect           -- enum cast
)
ON CONFLICT (namespace, resource, action, subject)
DO UPDATE SET effect = EXCLUDED.effect;

-- name: RevokePermission :exec
DELETE FROM permissions
WHERE namespace = $1 AND resource = $2 AND action = $3 AND subject = $4;
