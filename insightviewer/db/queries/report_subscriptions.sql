-- =============================================================================
-- Report Subscriptions Queries - InsightViewer Service
-- =============================================================================

-- name: CreateReportSubscription :one
INSERT INTO report_subscriptions (
    report_id,
    user_id,
    notification_frequency,
    notification_channels,
    notify_on_changes,
    notify_on_errors,
    threshold_conditions
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING
    id,
    report_id,
    user_id,
    notification_frequency,
    notification_channels,
    notify_on_changes,
    notify_on_errors,
    threshold_conditions,
    is_active,
    last_notification_at,
    created_at,
    updated_at;

-- name: GetReportSubscriptionByID :one
SELECT
    id,
    report_id,
    user_id,
    notification_frequency,
    notification_channels,
    notify_on_changes,
    notify_on_errors,
    threshold_conditions,
    is_active,
    last_notification_at,
    created_at,
    updated_at
FROM report_subscriptions
WHERE id = $1;

-- name: GetSubscriptionByReportAndUser :one
SELECT
    id,
    report_id,
    user_id,
    notification_frequency,
    notification_channels,
    notify_on_changes,
    notify_on_errors,
    threshold_conditions,
    is_active,
    last_notification_at,
    created_at,
    updated_at
FROM report_subscriptions
WHERE report_id = $1 AND user_id = $2;

-- name: GetSubscriptionsByReportID :many
SELECT
    id,
    report_id,
    user_id,
    notification_frequency,
    notification_channels,
    notify_on_changes,
    notify_on_errors,
    threshold_conditions,
    is_active,
    last_notification_at,
    created_at,
    updated_at
FROM report_subscriptions
WHERE report_id = $1 AND is_active = true
ORDER BY created_at DESC;

-- name: GetSubscriptionsByUserID :many
SELECT
    id,
    report_id,
    user_id,
    notification_frequency,
    notification_channels,
    notify_on_changes,
    notify_on_errors,
    threshold_conditions,
    is_active,
    last_notification_at,
    created_at,
    updated_at
FROM report_subscriptions
WHERE user_id = $1 AND is_active = true
ORDER BY created_at DESC;

-- name: UpdateReportSubscription :one
UPDATE report_subscriptions
SET
    notification_frequency = $2,
    notification_channels = $3,
    notify_on_changes = $4,
    notify_on_errors = $5,
    threshold_conditions = $6,
    updated_at = NOW()
WHERE id = $1
RETURNING
    id,
    report_id,
    user_id,
    notification_frequency,
    notification_channels,
    notify_on_changes,
    notify_on_errors,
    threshold_conditions,
    is_active,
    last_notification_at,
    created_at,
    updated_at;

-- name: UpdateSubscriptionStatus :one
UPDATE report_subscriptions
SET
    is_active = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING
    id,
    report_id,
    user_id,
    notification_frequency,
    notification_channels,
    notify_on_changes,
    notify_on_errors,
    threshold_conditions,
    is_active,
    last_notification_at,
    created_at,
    updated_at;

-- name: UpdateLastNotification :exec
UPDATE report_subscriptions
SET last_notification_at = NOW()
WHERE id = $1;

-- name: DeleteReportSubscription :exec
DELETE FROM report_subscriptions
WHERE id = $1;

-- name: GetSubscriptionsForNotification :many
SELECT
    id,
    report_id,
    user_id,
    notification_frequency,
    notification_channels,
    notify_on_changes,
    notify_on_errors,
    threshold_conditions,
    is_active,
    last_notification_at,
    created_at,
    updated_at
FROM report_subscriptions
WHERE report_id = $1
  AND is_active = true
  AND (
    notification_frequency = 'immediate' OR
    (notification_frequency = 'daily' AND (last_notification_at IS NULL OR last_notification_at < NOW() - INTERVAL '1 day')) OR
    (notification_frequency = 'weekly' AND (last_notification_at IS NULL OR last_notification_at < NOW() - INTERVAL '1 week')) OR
    (notification_frequency = 'monthly' AND (last_notification_at IS NULL OR last_notification_at < NOW() - INTERVAL '1 month'))
  );

-- name: GetActiveSubscriptions :many
SELECT
    id,
    report_id,
    user_id,
    notification_frequency,
    notification_channels,
    notify_on_changes,
    notify_on_errors,
    threshold_conditions,
    is_active,
    last_notification_at,
    created_at,
    updated_at
FROM report_subscriptions
WHERE is_active = true
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetSubscriptionStatistics :one
SELECT
    COUNT(*) as total_subscriptions,
    COUNT(CASE WHEN is_active = true THEN 1 END) as active_subscriptions,
    COUNT(CASE WHEN notification_frequency = 'immediate' THEN 1 END) as immediate_subscriptions,
    COUNT(CASE WHEN notification_frequency = 'daily' THEN 1 END) as daily_subscriptions,
    COUNT(CASE WHEN notification_frequency = 'weekly' THEN 1 END) as weekly_subscriptions,
    COUNT(CASE WHEN notification_frequency = 'monthly' THEN 1 END) as monthly_subscriptions,
    COUNT(CASE WHEN 'email' = ANY(notification_channels) THEN 1 END) as email_subscriptions,
    COUNT(CASE WHEN 'in_app' = ANY(notification_channels) THEN 1 END) as in_app_subscriptions
FROM report_subscriptions
WHERE created_at >= $1;