-- Two-Factor Authentication Queries

-- name: CreateTwoFactorBackupCode :one
INSERT INTO auth_two_factor_backup_codes (
    user_id, code_hash
) VALUES (
    $1, $2
) RETURNING *;

-- name: GetTwoFactorBackupCodes :many
SELECT * FROM auth_two_factor_backup_codes
WHERE user_id = $1 AND used_at IS NULL
ORDER BY created_at ASC;

-- name: UseTwoFactorBackupCode :exec
UPDATE auth_two_factor_backup_codes
SET used_at = CURRENT_TIMESTAMP
WHERE user_id = $1 AND code_hash = $2 AND used_at IS NULL;

-- name: DeleteTwoFactorBackupCodes :exec
DELETE FROM auth_two_factor_backup_codes
WHERE user_id = $1;

-- name: CountUnusedBackupCodes :one
SELECT COUNT(*) FROM auth_two_factor_backup_codes
WHERE user_id = $1 AND used_at IS NULL;