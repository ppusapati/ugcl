-- Tenant Cleanup and Maintenance Queries

-- name: CleanupUnverifiedDomains :exec
DELETE FROM tenant_domains
WHERE is_verified = false AND created_at <= $1;

-- name: CleanupInactiveTenantsData :exec
DELETE FROM tenant_features
WHERE tenant_id IN (
    SELECT id::text FROM tenants WHERE is_active = false AND tenants.updated_at <= $1
);

-- name: CleanupOldTenantAuditLogs :exec
DELETE FROM tenant_audit_log
WHERE created_at <= $1;

-- name: CleanupOldUsageMetrics :exec
DELETE FROM tenant_usage_metrics
WHERE recorded_at <= $1;

-- name: ArchiveInactiveTenants :many
SELECT * FROM tenants
WHERE is_active = false AND updated_at <= $1
ORDER BY updated_at ASC;

-- name: GetOrphanedTenantData :many
SELECT 'tenant_features' as table_name, tf.tenant_id
FROM tenant_features tf
LEFT JOIN tenants t ON tf.tenant_id = t.id::text
WHERE t.id IS NULL
UNION ALL
SELECT 'tenant_connection_strings' as table_name, tcs.tenant_id
FROM tenant_connection_strings tcs
LEFT JOIN tenants t ON tcs.tenant_id = t.id::text
WHERE t.id IS NULL
UNION ALL
SELECT 'tenant_domains' as table_name, td.tenant_id
FROM tenant_domains td
LEFT JOIN tenants t ON td.tenant_id = t.id::text
WHERE t.id IS NULL;

-- name: GetTenantDataSizes :many
SELECT
    t.id,
    t.name,
    (SELECT COUNT(*) FROM tenant_features tf WHERE tf.tenant_id = t.id::text) as features_count,
    (SELECT COUNT(*) FROM tenant_connection_strings tcs WHERE tcs.tenant_id = t.id::text) as connection_strings_count,
    (SELECT COUNT(*) FROM tenant_domains td WHERE td.tenant_id = t.id::text) as domains_count,
    (SELECT COUNT(*) FROM tenant_usage_metrics tum WHERE tum.tenant_id = t.id::text) as usage_metrics_count,
    (SELECT COUNT(*) FROM tenant_audit_log tal WHERE tal.tenant_id = t.id::text) as audit_entries_count
FROM tenants t
WHERE t.is_active = true
ORDER BY t.name;