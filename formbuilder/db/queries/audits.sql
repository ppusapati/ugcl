
-- =============================================================================
-- AUDIT LOG QUERIES
-- =============================================================================

-- name: CreateAuditLog :one
INSERT INTO audit_logs (
    id, instance_id, user_id, action, from_state, to_state,
    changes, timestamp, ip_address, user_agent
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: GetAuditLogs :many
SELECT * FROM audit_logs 
WHERE instance_id = $1
ORDER BY timestamp DESC;

-- name: GetAuditLogsByUser :many
SELECT * FROM audit_logs 
WHERE user_id = $1
ORDER BY timestamp DESC
LIMIT $2 OFFSET $3;

-- name: GetAuditLogsByAction :many
SELECT * FROM audit_logs 
WHERE instance_id = $1 AND action = $2
ORDER BY timestamp DESC;

-- name: GetRecentAuditLogs :many
SELECT * FROM audit_logs 
WHERE timestamp >= $1
ORDER BY timestamp DESC
LIMIT $2 OFFSET $3;
