-- =============================================================================
-- Report Schedules Queries - InsightViewer Service
-- =============================================================================

-- name: CreateReportSchedule :one
INSERT INTO report_schedules (
    report_id,
    name,
    description,
    cron_expression,
    timezone,
    parameters,
    export_formats,
    notify_on_success,
    notify_on_failure,
    notification_recipients,
    notification_channels,
    max_results_retained,
    auto_cleanup_days,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
) RETURNING
    id,
    report_id,
    name,
    description,
    cron_expression,
    timezone,
    is_active,
    next_run_at,
    last_run_at,
    last_run_status,
    last_run_id,
    run_count,
    failure_count,
    parameters,
    export_formats,
    notify_on_success,
    notify_on_failure,
    notification_recipients,
    notification_channels,
    max_results_retained,
    auto_cleanup_days,
    created_by,
    created_at,
    updated_by,
    updated_at;

-- name: GetReportScheduleByID :one
SELECT
    id,
    report_id,
    name,
    description,
    cron_expression,
    timezone,
    is_active,
    next_run_at,
    last_run_at,
    last_run_status,
    last_run_id,
    run_count,
    failure_count,
    parameters,
    export_formats,
    notify_on_success,
    notify_on_failure,
    notification_recipients,
    notification_channels,
    max_results_retained,
    auto_cleanup_days,
    created_by,
    created_at,
    updated_by,
    updated_at
FROM report_schedules
WHERE id = $1;

-- name: GetReportSchedulesByReportID :many
SELECT
    id,
    report_id,
    name,
    description,
    cron_expression,
    timezone,
    is_active,
    next_run_at,
    last_run_at,
    last_run_status,
    last_run_id,
    run_count,
    failure_count,
    parameters,
    export_formats,
    notify_on_success,
    notify_on_failure,
    notification_recipients,
    notification_channels,
    max_results_retained,
    auto_cleanup_days,
    created_by,
    created_at,
    updated_by,
    updated_at
FROM report_schedules
WHERE report_id = $1
ORDER BY created_at DESC;

-- name: GetActiveSchedules :many
SELECT
    id,
    report_id,
    name,
    description,
    cron_expression,
    timezone,
    is_active,
    next_run_at,
    last_run_at,
    last_run_status,
    last_run_id,
    run_count,
    failure_count,
    parameters,
    export_formats,
    notify_on_success,
    notify_on_failure,
    notification_recipients,
    notification_channels,
    max_results_retained,
    auto_cleanup_days,
    created_by,
    created_at,
    updated_by,
    updated_at
FROM report_schedules
WHERE is_active = true;

-- name: GetSchedulesDueForExecution :many
SELECT
    id,
    report_id,
    name,
    description,
    cron_expression,
    timezone,
    is_active,
    next_run_at,
    last_run_at,
    last_run_status,
    last_run_id,
    run_count,
    failure_count,
    parameters,
    export_formats,
    notify_on_success,
    notify_on_failure,
    notification_recipients,
    notification_channels,
    max_results_retained,
    auto_cleanup_days,
    created_by,
    created_at,
    updated_by,
    updated_at
FROM report_schedules
WHERE is_active = true
  AND next_run_at IS NOT NULL
  AND next_run_at <= NOW();

-- name: UpdateReportSchedule :one
UPDATE report_schedules
SET
    name = $2,
    description = $3,
    cron_expression = $4,
    timezone = $5,
    parameters = $6,
    export_formats = $7,
    notify_on_success = $8,
    notify_on_failure = $9,
    notification_recipients = $10,
    notification_channels = $11,
    max_results_retained = $12,
    auto_cleanup_days = $13,
    updated_by = $14,
    updated_at = NOW()
WHERE id = $1
RETURNING
    id,
    report_id,
    name,
    description,
    cron_expression,
    timezone,
    is_active,
    next_run_at,
    last_run_at,
    last_run_status,
    last_run_id,
    run_count,
    failure_count,
    parameters,
    export_formats,
    notify_on_success,
    notify_on_failure,
    notification_recipients,
    notification_channels,
    max_results_retained,
    auto_cleanup_days,
    created_by,
    created_at,
    updated_by,
    updated_at;

-- name: UpdateScheduleStatus :one
UPDATE report_schedules
SET
    is_active = $2,
    updated_by = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING
    id,
    report_id,
    name,
    description,
    cron_expression,
    timezone,
    is_active,
    next_run_at,
    last_run_at,
    last_run_status,
    last_run_id,
    run_count,
    failure_count,
    parameters,
    export_formats,
    notify_on_success,
    notify_on_failure,
    notification_recipients,
    notification_channels,
    max_results_retained,
    auto_cleanup_days,
    created_by,
    created_at,
    updated_by,
    updated_at;

-- name: UpdateScheduleExecution :exec
UPDATE report_schedules
SET
    next_run_at = $2,
    last_run_at = $3,
    last_run_status = $4,
    last_run_id = $5,
    run_count = run_count + 1,
    failure_count = CASE WHEN $4 = 'failed' THEN failure_count + 1 ELSE failure_count END
WHERE id = $1;

-- name: DeleteReportSchedule :exec
DELETE FROM report_schedules
WHERE id = $1;

-- name: GetSchedulesByUser :many
SELECT
    id,
    report_id,
    name,
    description,
    cron_expression,
    timezone,
    is_active,
    next_run_at,
    last_run_at,
    last_run_status,
    last_run_id,
    run_count,
    failure_count,
    parameters,
    export_formats,
    notify_on_success,
    notify_on_failure,
    notification_recipients,
    notification_channels,
    max_results_retained,
    auto_cleanup_days,
    created_by,
    created_at,
    updated_by,
    updated_at
FROM report_schedules
WHERE created_by = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;