-- Auth Users Queries

-- name: CreateAuthUser :one
INSERT INTO auth_users (
    user_id, username, email, phone, password_hash, password_salt,
    two_factor_enabled, two_factor_secret, backup_codes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetAuthUserByID :one
SELECT * FROM auth_users WHERE id = $1 AND is_active = true;

-- name: GetAuthUserByUserID :one
SELECT * FROM auth_users WHERE user_id = $1 AND is_active = true;

-- name: GetAuthUserByUsername :one
SELECT * FROM auth_users WHERE username = $1 AND is_active = true;

-- name: GetAuthUserByEmail :one
SELECT * FROM auth_users WHERE email = $1 AND is_active = true;

-- name: GetAuthUserByPhone :one
SELECT * FROM auth_users WHERE phone = $1 AND is_active = true;

-- name: UpdateAuthUser :one
UPDATE auth_users
SET username = $2, email = $3, phone = $4,
    is_active = $5, email_verified = $6, phone_verified = $7,
    two_factor_enabled = $8, updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateAuthUserPassword :exec
UPDATE auth_users
SET password_hash = $2, password_salt = $3, password_changed_at = CURRENT_TIMESTAMP,
    last_password_change = CURRENT_TIMESTAMP, failed_login_attempts = 0, account_locked_until = NULL
WHERE id = $1;

-- name: UpdateAuthUserLoginAttempts :exec
UPDATE auth_users
SET failed_login_attempts = $2, account_locked_until = $3, last_login_at = $4
WHERE id = $1;

-- name: VerifyAuthUserEmail :exec
UPDATE auth_users SET email_verified = true WHERE id = $1;

-- name: VerifyAuthUserPhone :exec
UPDATE auth_users SET phone_verified = true WHERE id = $1;

-- name: EnableTwoFactor :exec
UPDATE auth_users
SET two_factor_enabled = true, two_factor_secret = $2, backup_codes = $3
WHERE id = $1;

-- name: DisableTwoFactor :exec
UPDATE auth_users
SET two_factor_enabled = false, two_factor_secret = NULL, backup_codes = NULL
WHERE id = $1;

-- name: DeactivateAuthUser :exec
UPDATE auth_users SET is_active = false WHERE id = $1;