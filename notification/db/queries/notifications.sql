-- Notification CRUD Operations

-- name: CreateNotification :one
INSERT INTO notifications (
    type, channel, priority, recipient_id, recipient_type, 
    subject, message, template_id, template_data, status,
    source_id, source_type, scheduled_at, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
) RETURNING *;

-- name: GetNotification :one
SELECT * FROM notifications WHERE id = $1;

-- name: GetNotificationsByRecipient :many
SELECT * FROM notifications 
WHERE recipient_id = $1 AND recipient_type = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListNotifications :many
SELECT * FROM notifications
WHERE ($1::text IS NULL OR recipient_id = $1)
  AND ($2::notification_status IS NULL OR status = $2)
  AND ($3::notification_type IS NULL OR type = $3)
  AND ($4::timestamptz IS NULL OR created_at >= $4)
  AND ($5::timestamptz IS NULL OR created_at <= $5)
ORDER BY 
  CASE WHEN $6 = 'created_at' AND $7 = true THEN created_at END DESC,
  CASE WHEN $6 = 'created_at' AND $7 = false THEN created_at END ASC,
  CASE WHEN $6 = 'priority' AND $7 = true THEN priority END DESC,
  CASE WHEN $6 = 'priority' AND $7 = false THEN priority END ASC,
  created_at DESC
LIMIT $8 OFFSET $9;

-- name: CountNotifications :one
SELECT COUNT(*) FROM notifications
WHERE ($1::text IS NULL OR recipient_id = $1)
  AND ($2::notification_status IS NULL OR status = $2)
  AND ($3::notification_type IS NULL OR type = $3)
  AND ($4::timestamptz IS NULL OR created_at >= $4)
  AND ($5::timestamptz IS NULL OR created_at <= $5);

-- name: UpdateNotificationStatus :one
UPDATE notifications 
SET status = $2, error_message = $3, sent_at = $4, delivered_at = $5, updated_at = NOW()
WHERE id = $1 
RETURNING *;

-- name: MarkNotificationRead :one
UPDATE notifications 
SET read_at = NOW(), updated_at = NOW()
WHERE id = $1 AND recipient_id = $2
RETURNING *;

-- name: MarkAllNotificationsRead :exec
UPDATE notifications 
SET read_at = NOW(), updated_at = NOW()
WHERE recipient_id = $1 
  AND read_at IS NULL
  AND ($2::notification_type IS NULL OR type = $2);

-- name: UpdateNotificationRetry :one
UPDATE notifications 
SET retry_count = retry_count + 1, next_retry_at = $2, error_message = $3, updated_at = NOW()
WHERE id = $1 
RETURNING *;

-- name: GetPendingNotifications :many
SELECT * FROM notifications 
WHERE status = 'PENDING' 
   OR (status = 'FAILED' AND retry_count < 5 AND (next_retry_at IS NULL OR next_retry_at <= NOW()))
ORDER BY priority DESC, created_at ASC
LIMIT $1;

-- name: GetScheduledNotifications :many
SELECT * FROM notifications 
WHERE status = 'SCHEDULED' AND scheduled_at <= NOW()
ORDER BY priority DESC, scheduled_at ASC
LIMIT $1;

-- name: GetUnreadNotifications :many
SELECT * FROM notifications 
WHERE recipient_id = $1 AND read_at IS NULL AND status IN ('SENT', 'DELIVERED')
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountUnreadNotifications :one
SELECT COUNT(*) FROM notifications 
WHERE recipient_id = $1 AND read_at IS NULL AND status IN ('SENT', 'DELIVERED');

-- name: DeleteNotification :exec
DELETE FROM notifications WHERE id = $1;

-- name: DeleteOldNotifications :exec
DELETE FROM notifications 
WHERE created_at < $1 AND status IN ('READ', 'DELIVERED', 'CANCELLED');

-- Notification Templates CRUD Operations

-- name: CreateNotificationTemplate :one
INSERT INTO notification_templates (
    name, type, channel, subject_template, body_template, 
    language, variables, active, created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetNotificationTemplate :one
SELECT * FROM notification_templates WHERE id = $1;

-- name: GetNotificationTemplateByTypeChannel :one
SELECT * FROM notification_templates 
WHERE type = $1 AND channel = $2 AND language = $3 AND active = true
ORDER BY created_at DESC
LIMIT 1;

-- name: ListNotificationTemplates :many
SELECT * FROM notification_templates
WHERE ($1::notification_type IS NULL OR type = $1)
  AND ($2::notification_channel IS NULL OR channel = $2)
  AND ($3::boolean IS NULL OR active = $3)
ORDER BY name ASC
LIMIT $4 OFFSET $5;

-- name: CountNotificationTemplates :one
SELECT COUNT(*) FROM notification_templates
WHERE ($1::notification_type IS NULL OR type = $1)
  AND ($2::notification_channel IS NULL OR channel = $2)
  AND ($3::boolean IS NULL OR active = $3);

-- name: UpdateNotificationTemplate :one
UPDATE notification_templates 
SET name = $2, type = $3, channel = $4, subject_template = $5, 
    body_template = $6, language = $7, variables = $8, active = $9, updated_at = NOW()
WHERE id = $1 
RETURNING *;

-- name: DeleteNotificationTemplate :exec
DELETE FROM notification_templates WHERE id = $1;

-- Notification Preferences CRUD Operations

-- name: CreateNotificationPreference :one
INSERT INTO notification_preferences (
    user_id, type, enabled_channels, enabled, settings
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetNotificationPreference :one
SELECT * FROM notification_preferences WHERE user_id = $1 AND type = $2;

-- name: GetUserNotificationPreferences :many
SELECT * FROM notification_preferences WHERE user_id = $1 ORDER BY type ASC;

-- name: UpdateNotificationPreference :one
UPDATE notification_preferences 
SET enabled_channels = $3, enabled = $4, settings = $5, updated_at = NOW()
WHERE user_id = $1 AND type = $2 
RETURNING *;

-- name: UpsertNotificationPreference :one
INSERT INTO notification_preferences (user_id, type, enabled_channels, enabled, settings)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id, type) 
DO UPDATE SET 
    enabled_channels = EXCLUDED.enabled_channels,
    enabled = EXCLUDED.enabled,
    settings = EXCLUDED.settings,
    updated_at = NOW()
RETURNING *;

-- name: DeleteNotificationPreference :exec
DELETE FROM notification_preferences WHERE user_id = $1 AND type = $2;

-- Workflow State Notifications CRUD Operations

-- name: CreateWorkflowStateNotification :one
INSERT INTO workflow_state_notifications (
    notification_id, instance_id, form_id, previous_state, current_state,
    changed_by, assigned_to, assigned_role, form_data, transition_event, comment
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: GetWorkflowStateNotification :one
SELECT * FROM workflow_state_notifications WHERE notification_id = $1;

-- name: GetWorkflowStateNotificationsByInstance :many
SELECT wsn.*, n.subject, n.message, n.status, n.created_at, n.sent_at, n.read_at
FROM workflow_state_notifications wsn
JOIN notifications n ON wsn.notification_id = n.id
WHERE wsn.instance_id = $1
ORDER BY wsn.created_at DESC;

-- name: GetWorkflowStateNotificationsByForm :many
SELECT wsn.*, n.subject, n.message, n.status, n.created_at, n.sent_at, n.read_at
FROM workflow_state_notifications wsn
JOIN notifications n ON wsn.notification_id = n.id
WHERE wsn.form_id = $1
ORDER BY wsn.created_at DESC
LIMIT $2 OFFSET $3;

-- SLA Event Notifications CRUD Operations

-- name: CreateSLAEventNotification :one
INSERT INTO sla_event_notifications (
    notification_id, sla_instance_id, instance_id, sla_rule_id, sla_rule_name,
    event_type, state, due_time, breach_time, assigned_to, severity, 
    escalation_level, breach_duration, context
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
) RETURNING *;

-- name: GetSLAEventNotification :one
SELECT * FROM sla_event_notifications WHERE notification_id = $1;

-- name: GetSLAEventNotificationsBySLAInstance :many
SELECT sen.*, n.subject, n.message, n.status, n.created_at, n.sent_at, n.read_at
FROM sla_event_notifications sen
JOIN notifications n ON sen.notification_id = n.id
WHERE sen.sla_instance_id = $1
ORDER BY sen.created_at DESC;

-- name: GetSLAEventNotificationsByRule :many
SELECT sen.*, n.subject, n.message, n.status, n.created_at, n.sent_at, n.read_at
FROM sla_event_notifications sen
JOIN notifications n ON sen.notification_id = n.id
WHERE sen.sla_rule_id = $1
ORDER BY sen.created_at DESC
LIMIT $2 OFFSET $3;

-- Notification Delivery Log CRUD Operations

-- name: CreateNotificationDeliveryLog :one
INSERT INTO notification_delivery_log (
    notification_id, attempt_number, status, attempted_at, 
    completed_at, error_message, response_data
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetNotificationDeliveryLogs :many
SELECT * FROM notification_delivery_log 
WHERE notification_id = $1 
ORDER BY attempt_number ASC;

-- name: GetLatestDeliveryLog :one
SELECT * FROM notification_delivery_log 
WHERE notification_id = $1 
ORDER BY attempt_number DESC 
LIMIT 1;

-- Notification Statistics Operations

-- name: CreateNotificationStat :one
INSERT INTO notification_stats (date, type, channel, status, count)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (date, type, channel, status)
DO UPDATE SET count = notification_stats.count + EXCLUDED.count
RETURNING *;

-- name: GetNotificationStats :many
SELECT * FROM notification_stats
WHERE date BETWEEN $1 AND $2
  AND ($3::notification_type IS NULL OR type = $3)
  AND ($4::notification_channel IS NULL OR channel = $4)
ORDER BY date DESC, type ASC, channel ASC;

-- name: GetNotificationStatsSummary :one
SELECT 
    COUNT(*) as total_notifications,
    COUNT(CASE WHEN read_at IS NULL THEN 1 END) as unread_notifications,
    COUNT(CASE WHEN type IN ('WORKFLOW_STATE_CHANGE', 'WORKFLOW_ASSIGNMENT', 'WORKFLOW_APPROVAL_REQUEST', 'WORKFLOW_APPROVAL_RESPONSE') THEN 1 END) as workflow_notifications,
    COUNT(CASE WHEN type IN ('SLA_WARNING', 'SLA_BREACH', 'SLA_ESCALATION', 'SLA_COMPLETION') THEN 1 END) as sla_notifications
FROM notifications
WHERE recipient_id = $1
  AND ($2::timestamptz IS NULL OR created_at >= $2)
  AND ($3::timestamptz IS NULL OR created_at <= $3);

-- Advanced Query Operations

-- name: GetNotificationsBySource :many
SELECT * FROM notifications 
WHERE source_id = $1 AND source_type = $2
ORDER BY created_at DESC;

-- name: GetFailedNotifications :many
SELECT n.*, dl.error_message as last_error, dl.attempted_at as last_attempt
FROM notifications n
LEFT JOIN notification_delivery_log dl ON n.id = dl.notification_id
WHERE n.status = 'FAILED' 
  AND (n.next_retry_at IS NULL OR n.next_retry_at <= NOW())
  AND n.retry_count < $1
ORDER BY n.priority DESC, n.created_at ASC
LIMIT $2;

-- name: GetNotificationsByPriority :many
SELECT * FROM notifications 
WHERE priority = $1 AND status = $2
ORDER BY created_at ASC
LIMIT $3;

-- name: BulkUpdateNotificationStatus :exec
UPDATE notifications 
SET status = $2, updated_at = NOW()
WHERE id = ANY($1::uuid[]);

-- name: GetNotificationsByChannel :many
SELECT * FROM notifications 
WHERE channel = $1 
  AND ($2::notification_status IS NULL OR status = $2)
  AND ($3::timestamptz IS NULL OR created_at >= $3)
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: GetNotificationsByType :many
SELECT * FROM notifications 
WHERE type = $1 
  AND ($2::text IS NULL OR recipient_id = $2)
  AND ($3::timestamptz IS NULL OR created_at >= $3)
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- View-based Queries

-- name: GetUnreadNotificationsView :many
SELECT * FROM unread_notifications 
WHERE recipient_id = $1
  AND ($2::text IS NULL OR category = $2)
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: GetNotificationSummaryView :one
SELECT * FROM notification_summary 
WHERE recipient_id = $1 AND recipient_type = $2;

-- name: GetSLANotificationDetails :many
SELECT * FROM sla_notification_details 
WHERE recipient_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetWorkflowNotificationDetails :many
SELECT * FROM workflow_notification_details 
WHERE recipient_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- Cleanup Operations

-- name: CleanupOldNotifications :exec
DELETE FROM notifications 
WHERE created_at < $1 
  AND status IN ('DELIVERED', 'READ', 'CANCELLED', 'FAILED')
  AND retry_count >= 5;

-- name: CleanupOldDeliveryLogs :exec
DELETE FROM notification_delivery_log 
WHERE created_at < $1;

-- name: CleanupOldStats :exec
DELETE FROM notification_stats 
WHERE date < $1;
