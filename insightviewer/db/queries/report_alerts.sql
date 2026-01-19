-- =============================================================================
-- Report Alerts Queries - InsightViewer Service
-- =============================================================================

-- name: CreateReportAlert :one
INSERT INTO report_alerts (
    report_id,
    name,
    description,
    field_id,
    operator,
    threshold_value,
    condition_type,
    severity,
    notification_channels,
    notification_recipients,
    notification_message,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING
    id,
    report_id,
    name,
    description,
    field_id,
    operator,
    threshold_value,
    condition_type,
    is_active,
    severity,
    notification_channels,
    notification_recipients,
    notification_message,
    last_triggered_at,
    trigger_count,
    created_by,
    created_at,
    updated_by,
    updated_at;

-- name: GetReportAlertByID :one
SELECT
    id,
    report_id,
    name,
    description,
    field_id,
    operator,
    threshold_value,
    condition_type,
    is_active,
    severity,
    notification_channels,
    notification_recipients,
    notification_message,
    last_triggered_at,
    trigger_count,
    created_by,
    created_at,
    updated_by,
    updated_at
FROM report_alerts
WHERE id = $1;

-- name: GetReportAlertsByReportID :many
SELECT
    id,
    report_id,
    name,
    description,
    field_id,
    operator,
    threshold_value,
    condition_type,
    is_active,
    severity,
    notification_channels,
    notification_recipients,
    notification_message,
    last_triggered_at,
    trigger_count,
    created_by,
    created_at,
    updated_by,
    updated_at
FROM report_alerts
WHERE report_id = $1
ORDER BY created_at DESC;

-- name: GetActiveAlerts :many
SELECT
    id,
    report_id,
    name,
    description,
    field_id,
    operator,
    threshold_value,
    condition_type,
    is_active,
    severity,
    notification_channels,
    notification_recipients,
    notification_message,
    last_triggered_at,
    trigger_count,
    created_by,
    created_at,
    updated_by,
    updated_at
FROM report_alerts
WHERE is_active = true
ORDER BY severity DESC, created_at DESC;

-- name: GetAlertsBySeverity :many
SELECT
    id,
    report_id,
    name,
    description,
    field_id,
    operator,
    threshold_value,
    condition_type,
    is_active,
    severity,
    notification_channels,
    notification_recipients,
    notification_message,
    last_triggered_at,
    trigger_count,
    created_by,
    created_at,
    updated_by,
    updated_at
FROM report_alerts
WHERE severity = $1 AND is_active = true
ORDER BY created_at DESC;

-- name: UpdateReportAlert :one
UPDATE report_alerts
SET
    name = $2,
    description = $3,
    field_id = $4,
    operator = $5,
    threshold_value = $6,
    condition_type = $7,
    severity = $8,
    notification_channels = $9,
    notification_recipients = $10,
    notification_message = $11,
    updated_by = $12,
    updated_at = NOW()
WHERE id = $1
RETURNING
    id,
    report_id,
    name,
    description,
    field_id,
    operator,
    threshold_value,
    condition_type,
    is_active,
    severity,
    notification_channels,
    notification_recipients,
    notification_message,
    last_triggered_at,
    trigger_count,
    created_by,
    created_at,
    updated_by,
    updated_at;

-- name: UpdateAlertStatus :one
UPDATE report_alerts
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
    field_id,
    operator,
    threshold_value,
    condition_type,
    is_active,
    severity,
    notification_channels,
    notification_recipients,
    notification_message,
    last_triggered_at,
    trigger_count,
    created_by,
    created_at,
    updated_by,
    updated_at;

-- name: TriggerAlert :exec
UPDATE report_alerts
SET
    last_triggered_at = NOW(),
    trigger_count = trigger_count + 1
WHERE id = $1;

-- name: DeleteReportAlert :exec
DELETE FROM report_alerts
WHERE id = $1;

-- name: GetAlertsByUser :many
SELECT
    id,
    report_id,
    name,
    description,
    field_id,
    operator,
    threshold_value,
    condition_type,
    is_active,
    severity,
    notification_channels,
    notification_recipients,
    notification_message,
    last_triggered_at,
    trigger_count,
    created_by,
    created_at,
    updated_by,
    updated_at
FROM report_alerts
WHERE created_by = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetRecentlyTriggeredAlerts :many
SELECT
    id,
    report_id,
    name,
    description,
    field_id,
    operator,
    threshold_value,
    condition_type,
    is_active,
    severity,
    notification_channels,
    notification_recipients,
    notification_message,
    last_triggered_at,
    trigger_count,
    created_by,
    created_at,
    updated_by,
    updated_at
FROM report_alerts
WHERE last_triggered_at >= $1
ORDER BY last_triggered_at DESC;

-- name: GetAlertStatistics :one
SELECT
    COUNT(*) as total_alerts,
    COUNT(CASE WHEN is_active = true THEN 1 END) as active_alerts,
    COUNT(CASE WHEN severity = 'critical' THEN 1 END) as critical_alerts,
    COUNT(CASE WHEN severity = 'high' THEN 1 END) as high_alerts,
    COUNT(CASE WHEN severity = 'medium' THEN 1 END) as medium_alerts,
    COUNT(CASE WHEN severity = 'low' THEN 1 END) as low_alerts,
    COUNT(CASE WHEN last_triggered_at >= $1 THEN 1 END) as recently_triggered,
    SUM(trigger_count) as total_triggers
FROM report_alerts;