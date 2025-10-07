-- =============================================================================
-- approval.sql - Queries for approval workflow functionality
-- =============================================================================

-- name: CreateApprovalAction :one
INSERT INTO approval_actions (
    form_instance_id, approver_id, action, comments, acted_at,
    ip_address, user_agent, delegated_from, attachment_urls, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: GetApprovalActionsByInstanceID :many
SELECT * FROM approval_actions
WHERE form_instance_id = $1
ORDER BY acted_at ASC;

-- name: GetApprovalActionsByApprover :many
SELECT * FROM approval_actions
WHERE approver_id = $1
ORDER BY acted_at DESC
LIMIT $2 OFFSET $3;

-- name: GetApprovalActionByID :one
SELECT * FROM approval_actions
WHERE id = $1;

-- name: CreateApprovalDelegate :one
INSERT INTO approval_delegates (
    delegator_id, delegate_id, entity_types, start_date, end_date,
    is_active, reason, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetApprovalDelegatesByDelegator :many
SELECT * FROM approval_delegates
WHERE delegator_id = $1 AND is_active = true
ORDER BY created_at DESC;

-- name: GetApprovalDelegatesByDelegate :many
SELECT * FROM approval_delegates
WHERE delegate_id = $1 AND is_active = true
ORDER BY created_at DESC;

-- name: GetActiveDelegation :one
SELECT * FROM approval_delegates
WHERE delegator_id = $1
  AND is_active = true
  AND (entity_types = '[]'::jsonb OR entity_types ? $2)
  AND start_date <= NOW()
  AND (end_date IS NULL OR end_date > NOW())
ORDER BY created_at DESC
LIMIT 1;

-- name: UpdateApprovalDelegate :one
UPDATE approval_delegates
SET delegate_id = $2, entity_types = $3, end_date = $4,
    is_active = $5, reason = $6, metadata = $7, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteApprovalDelegate :exec
UPDATE approval_delegates
SET is_active = false, updated_at = NOW()
WHERE id = $1;

-- name: CreateApprovalReport :one
INSERT INTO approval_reports (
    workflow_id, entity_type, report_date, total_requests, approved_count,
    rejected_count, pending_count, escalated_count, avg_processing_time,
    sla_breach_count, sla_compliance_rate, bottleneck_steps, top_approvers, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
) RETURNING *;

-- name: GetApprovalReportsByWorkflow :many
SELECT * FROM approval_reports
WHERE workflow_id = $1
  AND ($2::DATE IS NULL OR report_date >= $2)
  AND ($3::DATE IS NULL OR report_date <= $3)
ORDER BY report_date DESC;

-- name: GetApprovalReportsByEntityType :many
SELECT * FROM approval_reports
WHERE entity_type = $1
  AND ($2::DATE IS NULL OR report_date >= $2)
  AND ($3::DATE IS NULL OR report_date <= $3)
ORDER BY report_date DESC;

-- name: GetPendingApprovalsForUser :many
SELECT fi.*, f.title as form_title, f.description as form_description
FROM form_instances fi
JOIN forms f ON fi.form_id = f.id
LEFT JOIN approval_delegates ad ON ad.delegator_id = fi.assigned_to
  AND ad.delegate_id = $1
  AND ad.is_active = true
  AND ad.start_date <= NOW()
  AND (ad.end_date IS NULL OR ad.end_date > NOW())
WHERE (fi.assigned_to = $1 OR ad.id IS NOT NULL)
  AND fi.current_state IN ('pending_approval', 'in_review', 'escalated')
  AND fi.deleted_at IS NULL
ORDER BY fi.created_at ASC
LIMIT $2 OFFSET $3;

-- name: GetApprovalMetrics :one
SELECT
    COUNT(*) as total_requests,
    COUNT(*) FILTER (WHERE fi.current_state = 'approved') as approved_count,
    COUNT(*) FILTER (WHERE fi.current_state = 'rejected') as rejected_count,
    COUNT(*) FILTER (WHERE fi.current_state IN ('pending_approval', 'in_review')) as pending_count,
    COUNT(*) FILTER (WHERE fi.current_state = 'escalated') as escalated_count,
    AVG(EXTRACT(EPOCH FROM (fi.updated_at - fi.created_at))/3600) as avg_processing_time_hours
FROM form_instances fi
JOIN forms f ON fi.form_id = f.id
WHERE ($1::TEXT IS NULL OR f.module = $1)
  AND ($2::UUID IS NULL OR f.workflow_id = $2)
  AND fi.created_at >= $3
  AND fi.created_at <= $4
  AND fi.deleted_at IS NULL;

-- name: GetEscalatedInstances :many
SELECT fi.*, f.title as form_title
FROM form_instances fi
JOIN forms f ON fi.form_id = f.id
WHERE fi.current_state = 'escalated'
  AND fi.deleted_at IS NULL
ORDER BY fi.updated_at DESC
LIMIT $1 OFFSET $2;

-- name: GetApprovalHistory :many
SELECT
    aa.id, aa.approver_id, aa.action, aa.comments, aa.acted_at,
    aa.ip_address, aa.user_agent, aa.delegated_from, aa.attachment_urls,
    al.action as audit_action, al.changes, al.timestamp as audit_timestamp
FROM approval_actions aa
LEFT JOIN audit_logs al ON al.instance_id = aa.form_instance_id
  AND al.timestamp >= aa.acted_at - INTERVAL '1 minute'
  AND al.timestamp <= aa.acted_at + INTERVAL '1 minute'
WHERE aa.form_instance_id = $1
ORDER BY aa.acted_at ASC;