-- =============================================================================
-- SLA TRACKING QUERIES
-- =============================================================================

-- =============================================================================
-- SLA INSTANCES QUERIES
-- =============================================================================

-- name: CreateSLAInstance :one
INSERT INTO sla_instances (
    id, instance_id, sla_rule_id, state, start_time, due_time,
    status, assigned_to, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: GetSLAInstance :one
SELECT * FROM sla_instances WHERE id = $1;

-- name: GetSLAInstancesByFormInstance :many
SELECT * FROM sla_instances 
WHERE instance_id = $1
ORDER BY created_at DESC;

-- name: GetActiveSLAInstances :many
SELECT * FROM sla_instances 
WHERE status = 'active'
ORDER BY due_time ASC;

-- name: GetOverdueSLAInstances :many
SELECT * FROM sla_instances 
WHERE status = 'active' AND due_time < NOW()
ORDER BY due_time ASC;

-- name: GetSLAInstancesByState :many
SELECT * FROM sla_instances 
WHERE state = $1 AND status = 'active'
ORDER BY due_time ASC;

-- name: GetSLAInstancesByAssignee :many
SELECT * FROM sla_instances 
WHERE assigned_to = $1 AND status = 'active'
ORDER BY due_time ASC;

-- name: UpdateSLAInstanceStatus :one
UPDATE sla_instances SET
    status = $2,
    completion_time = $3,
    breach_time = $4,
    breach_duration = $5,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CompleteSLAInstance :one
UPDATE sla_instances SET
    status = 'completed',
    completion_time = NOW(),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: BreachSLAInstance :one
UPDATE sla_instances SET
    status = 'breached',
    breach_time = NOW(),
    breach_duration = NOW() - due_time,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: PauseSLAInstance :one
UPDATE sla_instances SET
    status = 'paused',
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ResumeSLAInstance :one
UPDATE sla_instances SET
    status = 'active',
    due_time = due_time + $2, -- Add pause duration to due time
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ReassignSLAInstance :one
UPDATE sla_instances SET
    assigned_to = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- =============================================================================
-- SLA VIOLATIONS QUERIES
-- =============================================================================

-- name: CreateSLAViolation :one
INSERT INTO sla_violations (
    id, sla_instance_id, instance_id, sla_rule_id, violation_time,
    breach_duration, severity, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetSLAViolation :one
SELECT * FROM sla_violations WHERE id = $1;

-- name: GetSLAViolationsByInstance :many
SELECT * FROM sla_violations 
WHERE instance_id = $1
ORDER BY violation_time DESC;

-- name: GetUnresolvedSLAViolations :many
SELECT * FROM sla_violations 
WHERE resolved = false
ORDER BY violation_time DESC;

-- name: GetSLAViolationsBySeverity :many
SELECT * FROM sla_violations 
WHERE severity = $1 AND resolved = false
ORDER BY violation_time DESC;

-- name: GetSLAViolationsByDateRange :many
SELECT * FROM sla_violations 
WHERE violation_time BETWEEN $1 AND $2
ORDER BY violation_time DESC;

-- name: ResolveSLAViolation :one
UPDATE sla_violations SET
    resolved = true,
    resolved_at = NOW(),
    resolved_by = $2,
    resolution_notes = $3
WHERE id = $1
RETURNING *;

-- name: UpdateSLAViolationNotificationStatus :exec
UPDATE sla_violations SET
    notification_sent = $2,
    escalation_triggered = $3
WHERE id = $1;

-- =============================================================================
-- SLA ESCALATION INSTANCES QUERIES
-- =============================================================================

-- name: CreateSLAEscalationInstance :one
INSERT INTO sla_escalation_instances (
    id, sla_instance_id, escalation_level, triggered_at,
    actions_executed, status, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetSLAEscalationInstance :one
SELECT * FROM sla_escalation_instances WHERE id = $1;

-- name: GetSLAEscalationInstancesBySLAInstance :many
SELECT * FROM sla_escalation_instances 
WHERE sla_instance_id = $1
ORDER BY escalation_level ASC;

-- name: GetPendingEscalations :many
SELECT * FROM sla_escalation_instances 
WHERE status = 'pending'
ORDER BY triggered_at ASC;

-- name: GetFailedEscalations :many
SELECT * FROM sla_escalation_instances 
WHERE status = 'failed' AND (next_retry_at IS NULL OR next_retry_at <= NOW())
ORDER BY triggered_at ASC;

-- name: UpdateEscalationStatus :one
UPDATE sla_escalation_instances SET
    status = $2,
    error_message = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CompleteEscalation :one
UPDATE sla_escalation_instances SET
    status = 'completed',
    actions_executed = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: FailEscalation :one
UPDATE sla_escalation_instances SET
    status = 'failed',
    error_message = $2,
    retry_count = retry_count + 1,
    next_retry_at = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- =============================================================================
-- SLA NOTIFICATIONS QUERIES
-- =============================================================================

-- name: CreateSLANotification :one
INSERT INTO sla_notifications (
    id, sla_instance_id, sla_violation_id, escalation_instance_id,
    notification_type, recipient, channel, subject, message,
    template_used, status, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING *;

-- name: GetSLANotification :one
SELECT * FROM sla_notifications WHERE id = $1;

-- name: GetPendingNotifications :many
SELECT * FROM sla_notifications 
WHERE status = 'pending'
ORDER BY created_at ASC;

-- name: GetNotificationsByRecipient :many
SELECT * FROM sla_notifications 
WHERE recipient = $1
ORDER BY created_at DESC;

-- name: GetNotificationsByType :many
SELECT * FROM sla_notifications 
WHERE notification_type = $1
ORDER BY created_at DESC;

-- name: UpdateNotificationStatus :one
UPDATE sla_notifications SET
    status = $2,
    sent_at = $3,
    delivered_at = $4,
    error_message = $5
WHERE id = $1
RETURNING *;

-- name: MarkNotificationSent :one
UPDATE sla_notifications SET
    status = 'sent',
    sent_at = NOW()
WHERE id = $1
RETURNING *;

-- name: MarkNotificationDelivered :one
UPDATE sla_notifications SET
    status = 'delivered',
    delivered_at = NOW()
WHERE id = $1
RETURNING *;

-- name: FailNotification :one
UPDATE sla_notifications SET
    status = 'failed',
    error_message = $2,
    retry_count = retry_count + 1
WHERE id = $1
RETURNING *;

-- =============================================================================
-- SLA METRICS QUERIES
-- =============================================================================

-- name: CreateSLAMetric :one
INSERT INTO sla_metrics (
    id, metric_date, sla_rule_id, state, assigned_role,
    total_instances, completed_on_time, breached_instances,
    avg_completion_time, avg_breach_duration, compliance_percentage,
    created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
) RETURNING *;

-- name: GetSLAMetric :one
SELECT * FROM sla_metrics WHERE id = $1;

-- name: GetSLAMetricsByDate :many
SELECT * FROM sla_metrics 
WHERE metric_date = $1
ORDER BY sla_rule_id;

-- name: GetSLAMetricsByRule :many
SELECT * FROM sla_metrics 
WHERE sla_rule_id = $1
ORDER BY metric_date DESC;

-- name: GetSLAMetricsByDateRange :many
SELECT * FROM sla_metrics 
WHERE metric_date BETWEEN $1 AND $2
ORDER BY metric_date DESC, sla_rule_id;

-- name: UpdateSLAMetric :one
UPDATE sla_metrics SET
    total_instances = $2,
    completed_on_time = $3,
    breached_instances = $4,
    avg_completion_time = $5,
    avg_breach_duration = $6,
    compliance_percentage = $7,
    updated_at = NOW()
WHERE metric_date = $1 AND sla_rule_id = $8 AND state = $9 AND assigned_role = $10
RETURNING *;

-- name: UpsertSLAMetric :one
INSERT INTO sla_metrics (
    id, metric_date, sla_rule_id, state, assigned_role,
    total_instances, completed_on_time, breached_instances,
    avg_completion_time, avg_breach_duration, compliance_percentage,
    created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
ON CONFLICT (metric_date, sla_rule_id, state, assigned_role)
DO UPDATE SET
    total_instances = EXCLUDED.total_instances,
    completed_on_time = EXCLUDED.completed_on_time,
    breached_instances = EXCLUDED.breached_instances,
    avg_completion_time = EXCLUDED.avg_completion_time,
    avg_breach_duration = EXCLUDED.avg_breach_duration,
    compliance_percentage = EXCLUDED.compliance_percentage,
    updated_at = NOW()
RETURNING *;

-- =============================================================================
-- SLA PAUSE LOGS QUERIES
-- =============================================================================

-- name: CreateSLAPauseLog :one
INSERT INTO sla_pause_logs (
    id, sla_instance_id, action, reason, paused_by,
    paused_at, resumed_at, pause_duration, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetSLAPauseLogsBySLAInstance :many
SELECT * FROM sla_pause_logs 
WHERE sla_instance_id = $1
ORDER BY paused_at DESC;

-- name: GetActivePauses :many
SELECT * FROM sla_pause_logs 
WHERE action = 'pause' AND resumed_at IS NULL
ORDER BY paused_at DESC;

-- name: UpdatePauseLogResume :one
UPDATE sla_pause_logs SET
    resumed_at = $2,
    pause_duration = $2 - paused_at
WHERE id = $1
RETURNING *;

-- =============================================================================
-- COMPLEX SLA QUERIES
-- =============================================================================

-- name: GetSLADashboardData :many
SELECT 
    sr.id as sla_rule_id,
    sr.name as sla_rule_name,
    sr.state,
    COUNT(si.id) as total_instances,
    COUNT(CASE WHEN si.status = 'active' THEN 1 END) as active_instances,
    COUNT(CASE WHEN si.status = 'completed' THEN 1 END) as completed_instances,
    COUNT(CASE WHEN si.status = 'breached' THEN 1 END) as breached_instances,
    COUNT(CASE WHEN si.status = 'active' AND si.due_time < NOW() THEN 1 END) as overdue_instances,
    ROUND(
        (COUNT(CASE WHEN si.status = 'completed' THEN 1 END) * 100.0) / 
        NULLIF(COUNT(si.id), 0), 2
    ) as compliance_percentage
FROM sla_rules sr
LEFT JOIN sla_instances si ON sr.id = si.sla_rule_id
WHERE sr.active = true AND sr.deleted_at IS NULL
GROUP BY sr.id, sr.name, sr.state
ORDER BY sr.name;

-- name: GetSLAInstancesWithDetails :many
SELECT 
    si.*,
    fi.form_id,
    fi.created_by as instance_created_by,
    sr.name as sla_rule_name,
    sr.duration as sla_duration,
    EXTRACT(EPOCH FROM (si.due_time - NOW())) / 3600 as hours_remaining,
    CASE WHEN NOW() > si.due_time THEN true ELSE false END as is_overdue
FROM sla_instances si
JOIN form_instances fi ON si.instance_id = fi.id
JOIN sla_rules sr ON si.sla_rule_id = sr.id
WHERE si.status = $1 AND fi.deleted_at IS NULL
ORDER BY si.due_time ASC;

-- name: GetSLAViolationsWithDetails :many
SELECT 
    sv.*,
    si.instance_id,
    si.state,
    si.assigned_to,
    sr.name as sla_rule_name,
    fi.form_id,
    fi.created_by as instance_created_by
FROM sla_violations sv
JOIN sla_instances si ON sv.sla_instance_id = si.id
JOIN sla_rules sr ON sv.sla_rule_id = sr.id
JOIN form_instances fi ON sv.instance_id = fi.id
WHERE sv.created_at >= $1
ORDER BY sv.violation_time DESC;

-- name: GetSLAPerformanceReport :many
SELECT 
    DATE_TRUNC('day', si.created_at) as report_date,
    sr.name as sla_rule_name,
    sr.state,
    COUNT(si.id) as total_instances,
    COUNT(CASE WHEN si.status = 'completed' AND si.completion_time <= si.due_time THEN 1 END) as on_time_completions,
    COUNT(CASE WHEN si.status = 'breached' OR (si.status = 'completed' AND si.completion_time > si.due_time) THEN 1 END) as breached_instances,
    AVG(CASE WHEN si.status = 'completed' THEN EXTRACT(EPOCH FROM (si.completion_time - si.start_time)) / 3600 END) as avg_completion_hours,
    AVG(CASE WHEN si.breach_duration IS NOT NULL THEN EXTRACT(EPOCH FROM si.breach_duration) / 3600 END) as avg_breach_hours
FROM sla_instances si
JOIN sla_rules sr ON si.sla_rule_id = sr.id
WHERE si.created_at >= $1 AND si.created_at <= $2
GROUP BY DATE_TRUNC('day', si.created_at), sr.name, sr.state
ORDER BY report_date DESC, sr.name;
