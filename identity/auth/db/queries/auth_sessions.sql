-- Auth Sessions Queries

-- name: CreateAuthSession :one
INSERT INTO auth_sessions (
    session_id, user_id, tenant_id, refresh_token_hash, device_info,
    ip_address, user_agent, expires_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetAuthSessionByID :one
SELECT * FROM auth_sessions
WHERE session_id = $1 AND is_active = true AND expires_at > CURRENT_TIMESTAMP;

-- name: GetActiveSessionsByUserID :many
SELECT * FROM auth_sessions
WHERE user_id = $1 AND is_active = true AND expires_at > CURRENT_TIMESTAMP
ORDER BY last_accessed_at DESC;

-- name: GetActiveSessionsByUserIDAndTenant :many
SELECT * FROM auth_sessions
WHERE user_id = $1 AND tenant_id = $2 AND is_active = true AND expires_at > CURRENT_TIMESTAMP
ORDER BY last_accessed_at DESC;

-- name: UpdateSessionLastAccessed :exec
UPDATE auth_sessions
SET last_accessed_at = CURRENT_TIMESTAMP
WHERE session_id = $1;

-- name: RevokeSession :exec
UPDATE auth_sessions
SET is_active = false, revoked_at = CURRENT_TIMESTAMP, revoked_reason = $2
WHERE session_id = $1;

-- name: RevokeAllUserSessions :exec
UPDATE auth_sessions
SET is_active = false, revoked_at = CURRENT_TIMESTAMP, revoked_reason = 'logout_all'
WHERE user_id = $1 AND is_active = true;

-- name: CleanupExpiredSessions :exec
UPDATE auth_sessions
SET is_active = false, revoked_at = CURRENT_TIMESTAMP, revoked_reason = 'expired'
WHERE expires_at <= CURRENT_TIMESTAMP AND is_active = true;