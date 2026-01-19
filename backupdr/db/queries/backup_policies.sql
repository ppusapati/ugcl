-- name: CreateBackupPolicy :one
INSERT INTO backup_policies (
    name, description, backup_type, target_type, target_name,
    include_rules, exclude_rules, schedule_type, cron_expression,
    interval_minutes, backup_window, retention_policy, storage_config,
    compression, encryption, max_parallel_jobs, bandwidth_limit,
    timeout_minutes, notify_on_success, notify_on_failure,
    notify_channels, is_active, created_by, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
    $16, $17, $18, $19, $20, $21, $22, $23, $24
) RETURNING *;

-- name: GetBackupPolicyByID :one
SELECT * FROM backup_policies WHERE id = $1;

-- name: GetPoliciesByTarget :many
SELECT * FROM backup_policies
WHERE target_type = $1 AND target_name = $2 AND is_active = true
ORDER BY created_at DESC;

-- name: UpdateBackupPolicy :exec
UPDATE backup_policies SET
    name = $2, description = $3, backup_type = $4, target_type = $5,
    target_name = $6, include_rules = $7, exclude_rules = $8,
    schedule_type = $9, cron_expression = $10, interval_minutes = $11,
    backup_window = $12, retention_policy = $13, storage_config = $14,
    compression = $15, encryption = $16, max_parallel_jobs = $17,
    bandwidth_limit = $18, timeout_minutes = $19, notify_on_success = $20,
    notify_on_failure = $21, notify_channels = $22, is_active = $23,
    updated_by = $24, updated_at = NOW(), metadata = $25
WHERE id = $1;

-- name: DeleteBackupPolicy :exec
DELETE FROM backup_policies WHERE id = $1;

-- name: ListBackupPolicies :many
SELECT * FROM backup_policies
WHERE ($1::text IS NULL OR target_type = $1::text)
  AND ($2::text IS NULL OR backup_type = $2::text)
  AND ($3::boolean IS NULL OR is_active = $3::boolean)
  AND ($4::text IS NULL OR schedule_type = $4::text)
ORDER BY created_at DESC;

-- name: GetActivePolicies :many
SELECT * FROM backup_policies WHERE is_active = true ORDER BY created_at DESC;

-- name: GetPoliciesForScheduling :many
SELECT * FROM backup_policies
WHERE is_active = true
  AND next_backup IS NOT NULL
  AND next_backup <= $1
ORDER BY next_backup ASC;

-- name: UpdatePolicyLastBackup :exec
UPDATE backup_policies
SET last_backup = $2, next_backup = $3, updated_at = NOW()
WHERE id = $1;