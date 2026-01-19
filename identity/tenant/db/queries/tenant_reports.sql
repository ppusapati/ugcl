-- Tenant Reporting and Dashboard Queries

-- name: GetTenantDashboardStats :one
SELECT
    (SELECT COUNT(*) FROM tenants WHERE is_active = true) as total_active_tenants,
    (SELECT COUNT(*) FROM tenants WHERE tenant_db = 0) as shared_db_tenants,
    (SELECT COUNT(*) FROM tenants WHERE tenant_db = 1) as separate_db_tenants,
    (SELECT COUNT(*) FROM tenants WHERE tenant_db = 2) as separate_schema_tenants,
    (SELECT COUNT(*) FROM tenant_domains WHERE is_verified = true) as verified_domains,
    (SELECT COUNT(DISTINCT tenant_id) FROM tenant_features) as tenants_with_features;

-- name: GetTenantGrowthStats :many
SELECT
    DATE_TRUNC($1, created_at) as period,
    COUNT(*) as new_tenants,
    COUNT(*) FILTER (WHERE tenant_db = 0) as shared_db,
    COUNT(*) FILTER (WHERE tenant_db = 1) as separate_db,
    COUNT(*) FILTER (WHERE tenant_db = 2) as separate_schema
FROM tenants
WHERE created_at >= $2 AND is_active = true
GROUP BY DATE_TRUNC($1, created_at)
ORDER BY period ASC;

-- name: GetTenantFeatureUsage :many
SELECT
    key as feature_name,
    COUNT(*) as tenant_count,
    COUNT(*) FILTER (WHERE value_type = 'boolean' AND value = 'true') as enabled_count,
    array_agg(DISTINCT value) as unique_values
FROM tenant_features
GROUP BY key
ORDER BY tenant_count DESC;

-- name: GetTenantsBySubscriptionPlan :many
SELECT
    subscription_plan,
    COUNT(*) as tenant_count,
    COUNT(*) FILTER (WHERE is_active = true) as active_count,
    COUNT(*) FILTER (WHERE subscription_expires_at > CURRENT_TIMESTAMP) as valid_subscriptions,
    AVG(max_users) as avg_max_users,
    AVG(max_storage_gb) as avg_max_storage
FROM tenants
WHERE subscription_plan IS NOT NULL
GROUP BY subscription_plan
ORDER BY tenant_count DESC;

-- name: GetTenantActivitySummary :many
SELECT
    t.id,
    t.name,
    t.display_name,
    t.created_at,
    (SELECT COUNT(*) FROM tenant_audit_log tal WHERE tal.tenant_id = t.id::text AND tal.created_at >= CURRENT_DATE - INTERVAL '30 days') as recent_activity,
    (SELECT MAX(tal.created_at) FROM tenant_audit_log tal WHERE tal.tenant_id = t.id::text) as last_activity
FROM tenants t
WHERE t.is_active = true
ORDER BY recent_activity DESC, last_activity DESC
LIMIT $1 OFFSET $2;