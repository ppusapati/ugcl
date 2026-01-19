-- =============================================================================
-- Job Management Queries - Scheduler Service
-- =============================================================================

-- name: CreateJob :one
INSERT INTO jobs (
    name,
    description,
    cron_expression,
    is_active,
    target_service,
    target_method,
    target_payload,
    timeout_seconds,
    max_retries,
    retry_interval,
    next_run_at,
    notify_on_success,
    notify_on_failure,
    notification_channels,
    notification_recipients,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
) RETURNING *;

-- name: GetJobByID :one
SELECT * FROM jobs WHERE id = $1;

-- name: UpdateJob :one
UPDATE jobs SET
    name = $2,
    description = $3,
    cron_expression = $4,
    is_active = $5,
    target_service = $6,
    target_method = $7,
    target_payload = $8,
    timeout_seconds = $9,
    max_retries = $10,
    retry_interval = $11,
    next_run_at = $12,
    notify_on_success = $13,
    notify_on_failure = $14,
    notification_channels = $15,
    notification_recipients = $16,
    updated_by = $17,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteJob :exec
DELETE FROM jobs WHERE id = $1;

-- name: ListJobs :many
SELECT * FROM jobs
WHERE
    ($1::TEXT IS NULL OR target_service = $1) AND
    ($2::BOOLEAN IS NULL OR ($2 = true AND is_active = true) OR ($2 = false))
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountJobs :one
SELECT COUNT(*) FROM jobs
WHERE
    ($1::TEXT IS NULL OR target_service = $1) AND
    ($2::BOOLEAN IS NULL OR ($2 = true AND is_active = true) OR ($2 = false));

-- name: GetActiveJobs :many
SELECT * FROM jobs WHERE is_active = true ORDER BY next_run_at ASC;

-- name: GetJobsDueForExecution :many
SELECT * FROM jobs
WHERE is_active = true
  AND next_run_at IS NOT NULL
  AND next_run_at <= NOW()
ORDER BY next_run_at ASC;

-- name: UpdateJobStatus :exec
UPDATE jobs SET
    is_active = $2,
    updated_by = $3,
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateJobNextRun :exec
UPDATE jobs SET
    next_run_at = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateJobStats :exec
UPDATE jobs SET
    last_run_at = NOW(),
    last_run_status = $2,
    run_count = run_count + 1,
    failure_count = CASE WHEN $2 IN ('failed', 'timeout') THEN failure_count + 1 ELSE failure_count END,
    updated_at = NOW()
WHERE id = $1;

-- =============================================================================
-- Job Execution Queries
-- =============================================================================

-- name: CreateJobExecution :one
INSERT INTO job_executions (
    job_id,
    status,
    started_at,
    request_payload,
    triggered_by,
    host_name,
    retry_attempt
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetJobExecutionByID :one
SELECT * FROM job_executions WHERE id = $1;

-- name: UpdateJobExecution :one
UPDATE job_executions SET
    status = $2,
    completed_at = $3,
    duration_ms = $4,
    response_data = $5,
    error_message = $6
WHERE id = $1
RETURNING *;

-- name: GetJobExecutions :many
SELECT * FROM job_executions
WHERE job_id = $1
ORDER BY started_at DESC
LIMIT $2 OFFSET $3;

-- name: CountJobExecutions :one
SELECT COUNT(*) FROM job_executions WHERE job_id = $1;

-- name: CancelJobExecution :exec
UPDATE job_executions SET
    status = 'cancelled',
    completed_at = NOW(),
    error_message = 'Execution cancelled by user'
WHERE id = $1 AND status = 'running';

-- name: GetRunningExecutions :many
SELECT * FROM job_executions
WHERE status = 'running'
  AND started_at < NOW() - INTERVAL '1 hour' -- Find stuck executions
ORDER BY started_at ASC;

-- =============================================================================
-- Statistics and Monitoring Queries
-- =============================================================================

-- name: GetJobStatistics :one
SELECT
    COUNT(*) as total_executions,
    COUNT(*) FILTER (WHERE status = 'success') as successful_executions,
    COUNT(*) FILTER (WHERE status IN ('failed', 'timeout')) as failed_executions,
    CASE
        WHEN COUNT(*) > 0
        THEN ROUND((COUNT(*) FILTER (WHERE status = 'success')::numeric / COUNT(*)::numeric) * 100, 2)
        ELSE 0
    END as success_rate,
    COALESCE(ROUND(AVG(duration_ms)), 0) as average_duration_ms,
    MAX(completed_at) as last_execution
FROM job_executions
WHERE
    ($1::UUID IS NULL OR job_id = $1) AND
    ($2::TIMESTAMPTZ IS NULL OR started_at >= $2) AND
    completed_at IS NOT NULL;

-- name: GetSystemHealth :one
SELECT
    COUNT(*) as total_jobs,
    COUNT(*) FILTER (WHERE is_active = true) as active_jobs,
    COUNT(*) FILTER (WHERE is_active = true AND next_run_at <= NOW()) as overdue_jobs
FROM jobs;

-- name: GetRecentExecutions :many
SELECT
    je.*,
    j.name as job_name,
    j.target_service,
    j.target_method
FROM job_executions je
JOIN jobs j ON je.job_id = j.id
ORDER BY je.started_at DESC
LIMIT $1;

-- =============================================================================
-- Cleanup Queries
-- =============================================================================

-- name: CleanupOldExecutions :exec
DELETE FROM job_executions
WHERE completed_at < NOW() - INTERVAL '30 days';

-- name: CleanupOldLogs :exec
DELETE FROM job_execution_logs
WHERE timestamp < NOW() - INTERVAL '7 days';