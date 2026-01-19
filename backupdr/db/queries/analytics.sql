-- name: GetBackupStatistics :one
SELECT
    COUNT(DISTINCT bp.id) as total_policies,
    COUNT(DISTINCT bp.id) FILTER (WHERE bp.is_active = true) as active_policies,
    COUNT(bj.id) as total_jobs,
    COUNT(bj.id) FILTER (WHERE bj.status = 'COMPLETED') as successful_jobs,
    COUNT(bj.id) FILTER (WHERE bj.status = 'FAILED') as failed_jobs,
    COALESCE(SUM(bj.backup_size) FILTER (WHERE bj.status = 'COMPLETED'), 0) as total_backup_size,
    AVG(EXTRACT(EPOCH FROM bj.duration)) FILTER (WHERE bj.status = 'COMPLETED' AND bj.duration IS NOT NULL) as average_backup_time_seconds,
    AVG(bj.compression_ratio) FILTER (WHERE bj.status = 'COMPLETED' AND bj.compression_ratio > 0) as compression_ratio,
    (COUNT(bj.id) FILTER (WHERE bj.status = 'COMPLETED')::float / NULLIF(COUNT(bj.id), 0) * 100) as success_rate
FROM backup_policies bp
LEFT JOIN backup_jobs bj ON bp.id = bj.policy_id
WHERE ($1::timestamptz IS NULL OR bj.scheduled_at >= $1::timestamptz)
  AND ($2::timestamptz IS NULL OR bj.scheduled_at <= $2::timestamptz)
  AND ($3::text IS NULL OR bp.target_type = $3::text)
  AND ($4::text IS NULL OR bp.backup_type = $4::text)
  AND ($5::uuid IS NULL OR bp.id = $5::uuid);

-- name: GetStorageUtilization :many
WITH policy_storage AS (
    SELECT
        bp.target_type,
        bp.target_name,
        COUNT(bj.id) FILTER (WHERE bj.status = 'COMPLETED') as backup_count,
        COALESCE(SUM(bj.backup_size) FILTER (WHERE bj.status = 'COMPLETED'), 0) as used_size,
        MIN(bj.scheduled_at) FILTER (WHERE bj.status = 'COMPLETED') as oldest_backup,
        MAX(bj.scheduled_at) FILTER (WHERE bj.status = 'COMPLETED') as newest_backup
    FROM backup_policies bp
    LEFT JOIN backup_jobs bj ON bp.id = bj.policy_id
    WHERE bp.target_type = $1::text
    GROUP BY bp.target_type, bp.target_name
)
SELECT
    target_type,
    target_name,
    1000000000::bigint as total_size, -- 1GB default
    used_size,
    (1000000000 - used_size) as available_size,
    (used_size::float / 1000000000 * 100) as utilization_pct,
    backup_count,
    oldest_backup,
    newest_backup,
    NOW() as last_updated
FROM policy_storage
ORDER BY target_name;

-- name: GetComplianceReport :one
WITH policy_metrics AS (
    SELECT
        bp.id as policy_id,
        bp.name as policy_name,
        COUNT(bj.id) as backup_count,
        COUNT(bj.id) FILTER (WHERE bj.status = 'FAILED') as failure_count,
        MAX(bj.completed_at) FILTER (WHERE bj.status = 'COMPLETED') as last_successful_backup,
        MIN(bj.scheduled_at) FILTER (WHERE bj.status = 'SCHEDULED') as next_scheduled_backup,
        (COUNT(bj.id) FILTER (WHERE bj.status = 'COMPLETED')::float / NULLIF(COUNT(bj.id), 0) * 100) as success_rate
    FROM backup_policies bp
    LEFT JOIN backup_jobs bj ON bp.id = bj.policy_id
    WHERE bp.id = $1::uuid
    GROUP BY bp.id, bp.name
)
SELECT
    policy_id,
    policy_name,
    CASE
        WHEN success_rate >= 99 THEN 'COMPLIANT'
        WHEN success_rate >= 95 THEN 'WARNING'
        ELSE 'NON_COMPLIANT'
    END as compliance_status,
    last_successful_backup,
    next_scheduled_backup,
    backup_count,
    failure_count,
    COALESCE(success_rate, 0) as success_rate,
    true as rto_compliance, -- Default
    true as rpo_compliance, -- Default
    NOW() as last_checked
FROM policy_metrics;

-- name: GetBackupTrends :many
SELECT
    DATE_TRUNC('day', bj.scheduled_at) as date,
    COUNT(bj.id) as total_jobs,
    COUNT(bj.id) FILTER (WHERE bj.status = 'COMPLETED') as successful_jobs,
    COUNT(bj.id) FILTER (WHERE bj.status = 'FAILED') as failed_jobs,
    COALESCE(SUM(bj.backup_size) FILTER (WHERE bj.status = 'COMPLETED'), 0) as total_size,
    AVG(EXTRACT(EPOCH FROM bj.duration)) FILTER (WHERE bj.status = 'COMPLETED' AND bj.duration IS NOT NULL) as avg_duration_seconds
FROM backup_jobs bj
WHERE bj.scheduled_at >= $1::timestamptz
  AND bj.scheduled_at <= $2::timestamptz
GROUP BY DATE_TRUNC('day', bj.scheduled_at)
ORDER BY date;

-- name: GetFailedJobs :many
SELECT
    bj.*,
    bp.name as policy_name,
    bp.target_name
FROM backup_jobs bj
JOIN backup_policies bp ON bj.policy_id = bp.id
WHERE bj.status = 'FAILED'
  AND bj.scheduled_at >= $1::timestamptz
ORDER BY bj.scheduled_at DESC
LIMIT $2;