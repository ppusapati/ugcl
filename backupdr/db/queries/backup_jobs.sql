-- name: CreateBackupJob :one
INSERT INTO backup_jobs (
    policy_id, job_type, status, priority, scheduled_at, total_size,
    processed_size, compressed_size, progress_percent, estimated_finish,
    total_files, processed_files, skipped_files, errored_files,
    backup_path, backup_size, compression_ratio, checksum,
    throughput_mbps, average_speed, network_usage, cpu_usage,
    error_message, retry_count, max_retries, parent_job_id,
    baseline_job_id, executed_by, job_metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
    $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29
) RETURNING *;

-- name: GetBackupJobByID :one
SELECT * FROM backup_jobs WHERE id = $1;

-- name: GetJobsByPolicy :many
SELECT * FROM backup_jobs WHERE policy_id = $1 ORDER BY scheduled_at DESC;

-- name: UpdateBackupJob :exec
UPDATE backup_jobs SET
    status = $2, started_at = $3, completed_at = $4, duration = $5,
    total_size = $6, processed_size = $7, compressed_size = $8,
    progress_percent = $9, estimated_finish = $10, total_files = $11,
    processed_files = $12, skipped_files = $13, errored_files = $14,
    backup_path = $15, backup_size = $16, compression_ratio = $17,
    checksum = $18, throughput_mbps = $19, average_speed = $20,
    network_usage = $21, cpu_usage = $22, error_message = $23,
    retry_count = $24, job_metadata = $25
WHERE id = $1;

-- name: ListBackupJobs :many
SELECT * FROM backup_jobs
WHERE ($1::uuid IS NULL OR policy_id = $1::uuid)
  AND ($2::text IS NULL OR job_type = $2::text)
  AND ($3::text IS NULL OR status = $3::text)
  AND ($4::text IS NULL OR priority = $4::text)
  AND ($5::timestamptz IS NULL OR scheduled_at >= $5::timestamptz)
  AND ($6::timestamptz IS NULL OR scheduled_at <= $6::timestamptz)
  AND ($7::timestamptz IS NULL OR started_at >= $7::timestamptz)
  AND ($8::timestamptz IS NULL OR started_at <= $8::timestamptz)
ORDER BY scheduled_at DESC;

-- name: GetActiveJobs :many
SELECT * FROM backup_jobs
WHERE status IN ('RUNNING', 'QUEUED', 'SCHEDULED')
ORDER BY scheduled_at ASC;

-- name: GetJobsForScheduling :many
SELECT * FROM backup_jobs
WHERE status = 'SCHEDULED' AND scheduled_at <= $1
ORDER BY priority DESC, scheduled_at ASC;

-- name: GetJobsByDateRange :many
SELECT * FROM backup_jobs
WHERE policy_id = $1
  AND scheduled_at >= $2
  AND scheduled_at <= $3
ORDER BY scheduled_at DESC;

-- name: GetRecentJobs :many
SELECT * FROM backup_jobs
WHERE scheduled_at >= $1
ORDER BY scheduled_at DESC
LIMIT $2;

-- name: UpdateJobStatus :exec
UPDATE backup_jobs
SET status = $2, updated_at = NOW()
WHERE id = $1;