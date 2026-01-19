-- Tenant Security and Analytics Queries

-- name: GetTenantsWithExpiredSubscriptions :many
SELECT * FROM tenants
WHERE subscription_expires_at <= CURRENT_TIMESTAMP AND is_active = true
ORDER BY subscription_expires_at ASC;

-- name: GetTenantsExceedingLimits :many
SELECT t.*
FROM tenants t
WHERE t.is_active = true
ORDER BY t.created_at DESC;

-- name: GetTenantStats :one
SELECT
    t.*,
    (SELECT COUNT(*) FROM tenant_domains td WHERE td.tenant_id = t.id AND td.is_verified = true) as verified_domains,
    (SELECT COUNT(*) FROM tenant_features tf WHERE tf.tenant_id = t.id::text) as feature_count,
    (SELECT COUNT(*) FROM tenant_audit_log tal WHERE tal.tenant_id = t.id::text AND tal.created_at >= CURRENT_DATE - INTERVAL '30 days') as recent_activity
FROM tenants t
WHERE t.id = $1;

-- name: GetTenantsByRegionStats :many
SELECT
    region,
    COUNT(*) as tenant_count,
    COUNT(CASE WHEN is_active = true THEN 1 END) as active_count,
    COUNT(CASE WHEN tenant_db = 1 THEN 1 END) as separate_db_count,
    COUNT(CASE WHEN tenant_db = 2 THEN 1 END) as separate_schema_count
FROM tenants
GROUP BY region
ORDER BY tenant_count DESC;

-- name: GetTenantSubscriptionStats :many
SELECT
    subscription_plan,
    COUNT(*) as tenant_count,
    COUNT(CASE WHEN is_active = true THEN 1 END) as active_count,
    COUNT(CASE WHEN subscription_expires_at <= CURRENT_TIMESTAMP THEN 1 END) as expired_count
FROM tenants
WHERE subscription_plan IS NOT NULL
GROUP BY subscription_plan
ORDER BY tenant_count DESC;