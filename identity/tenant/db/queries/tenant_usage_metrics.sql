-- Tenant Usage Metrics Queries

-- name: RecordTenantUsageMetric :one
INSERT INTO tenant_usage_metrics (
    tenant_id, metric_name, metric_value, metric_unit, period_start, period_end
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetTenantUsageMetrics :many
SELECT * FROM tenant_usage_metrics
WHERE tenant_id = $1 AND metric_name = $2
ORDER BY recorded_at DESC
LIMIT $3 OFFSET $4;

-- name: GetTenantUsageMetricsInPeriod :many
SELECT * FROM tenant_usage_metrics
WHERE tenant_id = $1 AND recorded_at BETWEEN $2 AND $3
ORDER BY recorded_at DESC;

-- name: GetLatestTenantUsageMetric :one
SELECT * FROM tenant_usage_metrics
WHERE tenant_id = $1 AND metric_name = $2
ORDER BY recorded_at DESC
LIMIT 1;

-- name: GetTenantUsageMetricsSummary :many
SELECT
    metric_name,
    metric_unit,
    SUM(metric_value) as total_value,
    AVG(metric_value) as avg_value,
    MAX(metric_value) as max_value,
    MIN(metric_value) as min_value,
    COUNT(*) as record_count
FROM tenant_usage_metrics
WHERE tenant_id = $1 AND recorded_at BETWEEN $2 AND $3
GROUP BY metric_name, metric_unit
ORDER BY metric_name;

-- name: DeleteOldUsageMetrics :exec
DELETE FROM tenant_usage_metrics
WHERE recorded_at <= $1;