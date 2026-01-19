-- =============================================================================
-- Report Runs Queries - InsightViewer Service
-- =============================================================================

-- name: CreateReportRun :one
INSERT INTO report_runs (
    report_id,
    report_version,
    run_by,
    parameters,
    execution_plan,
    cache_key,
    triggered_by,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING
    id,
    report_id,
    report_version,
    run_by,
    run_at,
    completed_at,
    status,
    duration_ms,
    row_count,
    error_message,
    parameters,
    execution_plan,
    cache_key,
    triggered_by,
    metadata;

-- name: GetReportRunByID :one
SELECT
    id,
    report_id,
    report_version,
    run_by,
    run_at,
    completed_at,
    status,
    duration_ms,
    row_count,
    error_message,
    parameters,
    execution_plan,
    cache_key,
    triggered_by,
    metadata
FROM report_runs
WHERE id = $1;

-- name: GetReportRunsByReportID :many
SELECT
    id,
    report_id,
    report_version,
    run_by,
    run_at,
    completed_at,
    status,
    duration_ms,
    row_count,
    error_message,
    parameters,
    execution_plan,
    cache_key,
    triggered_by,
    metadata
FROM report_runs
WHERE report_id = $1
ORDER BY run_at DESC
LIMIT $2 OFFSET $3;

-- name: GetReportRunsByUser :many
SELECT
    id,
    report_id,
    report_version,
    run_by,
    run_at,
    completed_at,
    status,
    duration_ms,
    row_count,
    error_message,
    parameters,
    execution_plan,
    cache_key,
    triggered_by,
    metadata
FROM report_runs
WHERE run_by = $1
ORDER BY run_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateReportRunStatus :one
UPDATE report_runs
SET
    status = $2,
    completed_at = CASE WHEN $2 IN ('completed', 'failed', 'cancelled', 'timeout') THEN NOW() ELSE completed_at END,
    duration_ms = CASE WHEN $2 IN ('completed', 'failed', 'cancelled', 'timeout') THEN EXTRACT(EPOCH FROM (NOW() - run_at)) * 1000 ELSE duration_ms END,
    row_count = COALESCE($3, row_count),
    error_message = COALESCE($4, error_message)
WHERE id = $1
RETURNING
    id,
    report_id,
    report_version,
    run_by,
    run_at,
    completed_at,
    status,
    duration_ms,
    row_count,
    error_message,
    parameters,
    execution_plan,
    cache_key,
    triggered_by,
    metadata;

-- name: GetRunningReports :many
SELECT
    id,
    report_id,
    report_version,
    run_by,
    run_at,
    completed_at,
    status,
    duration_ms,
    row_count,
    error_message,
    parameters,
    execution_plan,
    cache_key,
    triggered_by,
    metadata
FROM report_runs
WHERE status = 'running'
ORDER BY run_at ASC;

-- name: GetReportRunsWithStatus :many
SELECT
    id,
    report_id,
    report_version,
    run_by,
    run_at,
    completed_at,
    status,
    duration_ms,
    row_count,
    error_message,
    parameters,
    execution_plan,
    cache_key,
    triggered_by,
    metadata
FROM report_runs
WHERE status = $1
ORDER BY run_at DESC
LIMIT $2 OFFSET $3;

-- name: CleanupOldReportRuns :exec
DELETE FROM report_runs
WHERE run_at < $1 AND status IN ('completed', 'failed', 'cancelled', 'timeout');

-- name: GetReportRunStatistics :one
SELECT
    COUNT(*) as total_runs,
    COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_runs,
    COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_runs,
    COUNT(CASE WHEN status = 'running' THEN 1 END) as running_runs,
    AVG(duration_ms) as avg_duration_ms,
    MAX(duration_ms) as max_duration_ms,
    MIN(duration_ms) as min_duration_ms
FROM report_runs
WHERE report_id = $1 AND run_at >= $2;