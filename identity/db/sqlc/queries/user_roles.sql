-- name: AssignRoleToUser :exec
INSERT INTO user_roles (user_uuid, role_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveRoleFromUser :exec
DELETE FROM user_roles
WHERE user_uuid = $1 AND role_id = $2;
