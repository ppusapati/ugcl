-- Auth Analytics and Security Queries

-- name: GetUserLoginStats :one
SELECT
    u.*,
    (SELECT COUNT(*) FROM auth_login_attempts ala WHERE ala.user_id = u.id AND ala.success = true AND ala.attempted_at >= CURRENT_DATE - INTERVAL '30 days') as successful_logins_30d,
    (SELECT COUNT(*) FROM auth_login_attempts ala WHERE ala.user_id = u.id AND ala.success = false AND ala.attempted_at >= CURRENT_DATE - INTERVAL '30 days') as failed_logins_30d,
    (SELECT COUNT(*) FROM auth_sessions s WHERE s.user_id = u.id AND s.is_active = true) as active_sessions
FROM auth_users u
WHERE u.id = $1;

-- name: GetSuspiciousLoginActivity :many
SELECT
    identifier,
    ip_address,
    COUNT(*) as failed_attempts,
    MAX(attempted_at) as last_attempt,
    MIN(attempted_at) as first_attempt
FROM auth_login_attempts
WHERE success = false AND attempted_at >= $1
GROUP BY identifier, ip_address
HAVING COUNT(*) >= $2
ORDER BY failed_attempts DESC, last_attempt DESC;

-- name: GetActiveSessionsStats :many
SELECT
    DATE_TRUNC('hour', created_at) as hour,
    COUNT(*) as sessions_created,
    COUNT(DISTINCT user_id) as unique_users,
    COUNT(DISTINCT tenant_id) as unique_tenants
FROM auth_sessions
WHERE created_at >= $1
GROUP BY DATE_TRUNC('hour', created_at)
ORDER BY hour DESC;

-- name: GetUsersByTenantCount :many
SELECT
    u.id,
    u.username,
    u.email,
    COUNT(aut.tenant_id) as tenant_count,
    array_agg(aut.tenant_id) as tenant_ids
FROM auth_users u
LEFT JOIN auth_user_tenants aut ON u.id = aut.user_id AND aut.is_active = true
WHERE u.is_active = true
GROUP BY u.id, u.username, u.email
HAVING COUNT(aut.tenant_id) >= $1
ORDER BY tenant_count DESC;

-- name: GetExpiredSessionsToCleanup :many
SELECT * FROM auth_sessions
WHERE expires_at <= CURRENT_TIMESTAMP AND is_active = true
LIMIT $1;