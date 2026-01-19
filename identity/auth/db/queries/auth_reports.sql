-- Auth Reporting and Dashboard Queries

-- name: GetAuthDashboardStats :one
SELECT
    (SELECT COUNT(*) FROM auth_users WHERE is_active = true) as total_active_users,
    (SELECT COUNT(*) FROM auth_users WHERE two_factor_enabled = true) as users_with_2fa,
    (SELECT COUNT(*) FROM auth_sessions WHERE is_active = true) as active_sessions,
    (SELECT COUNT(*) FROM auth_login_attempts WHERE success = true AND attempted_at >= CURRENT_DATE) as todays_successful_logins,
    (SELECT COUNT(*) FROM auth_login_attempts WHERE success = false AND attempted_at >= CURRENT_DATE) as todays_failed_logins,
    (SELECT COUNT(DISTINCT tenant_id) FROM auth_user_tenants WHERE is_active = true) as active_tenants_with_users;

-- name: GetUserGrowthStats :many
SELECT
    DATE_TRUNC($1, created_at) as period,
    COUNT(*) as new_users,
    COUNT(*) FILTER (WHERE email_verified = true) as verified_users
FROM auth_users
WHERE created_at >= $2
GROUP BY DATE_TRUNC($1, created_at)
ORDER BY period ASC;

-- name: GetLoginTrends :many
SELECT
    DATE_TRUNC('day', attempted_at) as login_date,
    COUNT(*) FILTER (WHERE success = true) as successful_logins,
    COUNT(*) FILTER (WHERE success = false) as failed_logins,
    COUNT(DISTINCT user_id) FILTER (WHERE success = true) as unique_users_logged_in
FROM auth_login_attempts
WHERE attempted_at >= $1
GROUP BY DATE_TRUNC('day', attempted_at)
ORDER BY login_date ASC;

-- name: GetSecurityEventsSummary :many
SELECT
    event_type,
    COUNT(*) as event_count,
    COUNT(DISTINCT user_id) as unique_users,
    MAX(created_at) as last_occurrence
FROM auth_security_events
WHERE created_at >= $1
GROUP BY event_type
ORDER BY event_count DESC;

-- name: GetTenantUserDistribution :many
SELECT
    aut.tenant_id,
    COUNT(DISTINCT aut.user_id) as user_count,
    COUNT(*) FILTER (WHERE au.two_factor_enabled = true) as users_with_2fa,
    COUNT(*) FILTER (WHERE au.last_login_at >= CURRENT_DATE - INTERVAL '30 days') as active_last_30d
FROM auth_user_tenants aut
JOIN auth_users au ON aut.user_id = au.id
WHERE aut.is_active = true
GROUP BY aut.tenant_id
ORDER BY user_count DESC;