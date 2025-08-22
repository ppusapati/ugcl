-- name: CreateUser :exec
INSERT INTO users (
    uuid, username, normalized_username, fullname,
    phone, phone_confirmed, email, normalized_email,
    email_confirmed, password, password_hash, gender,
    roles_cache, avatar, two_factor_enabled, salt,
    two_factor_secret, is_active
) VALUES (
    $1, $2, $3, $4,
    $5, $6, $7, $8,
    $9, $10, $11, $12,
    $13, $14, $15, $16,
    $17, $18
);

-- name: ListUsers :many
SELECT * FROM users;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByUUID :one
SELECT * FROM users WHERE uuid = $1;

-- name: UpdateUser :exec
UPDATE users
SET username           = $2,
    normalized_username= $3,
    fullname           = $4,
    phone              = $5,
    phone_confirmed    = $6,
    email              = $7,
    normalized_email   = $8,
    email_confirmed    = $9,
    password           = $10,
    password_hash      = $11,
    gender             = $12,
    roles_cache        = $13,
    avatar             = $14,
    two_factor_enabled = $15,
    salt               = $16,
    two_factor_secret  = $17,
    is_active          = $18,
    updated_at         = now()
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: UserExists :one
SELECT EXISTS (
  SELECT 1 FROM users
  WHERE
    (@username::text IS NULL OR username ILIKE '%' || @username || '%') AND
    (@email::text    IS NULL OR email    ILIKE '%' || @email    || '%') AND
    (@phone::text    IS NULL OR phone    ILIKE '%' || @phone    || '%') AND
    (@fullname::text IS NULL OR fullname ILIKE '%' || @fullname || '%')
) AS exists;

-- name: GetUserPermissions :many
SELECT p.*
FROM permissions p
JOIN user_permissions up ON up.permission_id = p.id
WHERE up.user_uuid = $1;
