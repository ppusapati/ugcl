-- Security and Audit Queries

-- name: LogLoginAttempt :one
INSERT INTO auth_login_attempts (
    identifier, ip_address, user_agent, success, failure_reason, tenant_id, user_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetRecentLoginAttempts :many
SELECT * FROM auth_login_attempts
WHERE identifier = $1 AND attempted_at >= $2
ORDER BY attempted_at DESC;

-- name: GetFailedLoginAttemptsByIP :many
SELECT * FROM auth_login_attempts
WHERE ip_address = $1 AND success = false AND attempted_at >= $2
ORDER BY attempted_at DESC;

-- name: LogSecurityEvent :one
INSERT INTO auth_security_events (
    user_id, event_type, event_data, ip_address, user_agent, tenant_id
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetUserSecurityEvents :many
SELECT * FROM auth_security_events
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetSecurityEventsByType :many
SELECT * FROM auth_security_events
WHERE event_type = $1 AND created_at >= $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: GetTenantSecurityEvents :many
SELECT * FROM auth_security_events
WHERE tenant_id = $1 AND created_at >= $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;