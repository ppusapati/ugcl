-- =============================================================================
-- Import Jobs Queries - DataBridge Module
-- =============================================================================

-- name: GetJobsByUser :many
SELECT
    j.id,
    j.job_name,
    j.mapping_id,
    j.file_name,
    j.file_size,
    j.file_hash,
    j.status,
    j.total_rows,
    j.processed_rows,
    j.successful_rows,
    j.failed_rows,
    j.error_details,
    j.started_at,
    j.completed_at,
    j.created_at,
    j.created_by,
    m.mapping_name
FROM import_jobs j
JOIN import_mappings m ON j.mapping_id = m.id
WHERE j.created_by = $1
ORDER BY j.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetJobByID :one
SELECT
    j.id,
    j.job_name,
    j.mapping_id,
    j.file_name,
    j.file_size,
    j.file_hash,
    j.status,
    j.total_rows,
    j.processed_rows,
    j.successful_rows,
    j.failed_rows,
    j.error_details,
    j.started_at,
    j.completed_at,
    j.created_at,
    j.created_by,
    m.mapping_name,
    m.field_mappings,
    m.table_id
FROM import_jobs j
JOIN import_mappings m ON j.mapping_id = m.id
WHERE j.id = $1;

-- name: GetActiveJobs :many
SELECT
    j.id,
    j.job_name,
    j.mapping_id,
    j.file_name,
    j.file_size,
    j.file_hash,
    j.status,
    j.total_rows,
    j.processed_rows,
    j.successful_rows,
    j.failed_rows,
    j.error_details,
    j.started_at,
    j.completed_at,
    j.created_at,
    j.created_by
FROM import_jobs j
WHERE j.status IN ('pending', 'processing')
ORDER BY j.created_at;

-- name: GetJobsByStatus :many
SELECT
    j.id,
    j.job_name,
    j.mapping_id,
    j.file_name,
    j.file_size,
    j.file_hash,
    j.status,
    j.total_rows,
    j.processed_rows,
    j.successful_rows,
    j.failed_rows,
    j.error_details,
    j.started_at,
    j.completed_at,
    j.created_at,
    j.created_by
FROM import_jobs j
WHERE j.status = $1
ORDER BY j.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateJob :one
INSERT INTO import_jobs (
    job_name,
    mapping_id,
    file_name,
    file_size,
    file_hash,
    total_rows,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING
    id,
    job_name,
    mapping_id,
    file_name,
    file_size,
    file_hash,
    status,
    total_rows,
    processed_rows,
    successful_rows,
    failed_rows,
    error_details,
    started_at,
    completed_at,
    created_at,
    created_by;

-- name: UpdateJobStatus :exec
UPDATE import_jobs
SET
    status = $2,
    started_at = CASE WHEN $2 = 'processing' THEN NOW() ELSE started_at END,
    completed_at = CASE WHEN $2 IN ('completed', 'failed', 'cancelled') THEN NOW() ELSE completed_at END
WHERE id = $1;

-- name: UpdateJobProgress :exec
UPDATE import_jobs
SET
    processed_rows = $2,
    successful_rows = $3,
    failed_rows = $4,
    error_details = $5
WHERE id = $1;

-- name: UpdateJobError :exec
UPDATE import_jobs
SET
    status = 'failed',
    completed_at = NOW(),
    error_details = $2
WHERE id = $1;

-- name: CancelJob :exec
UPDATE import_jobs
SET
    status = 'cancelled',
    completed_at = NOW()
WHERE id = $1 AND status IN ('pending', 'processing');

-- name: GetJobStatistics :one
SELECT
    COUNT(*) as total_jobs,
    COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_jobs,
    COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_jobs,
    COUNT(CASE WHEN status = 'processing' THEN 1 END) as processing_jobs,
    COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_jobs,
    COALESCE(SUM(successful_rows), 0) as total_successful_rows,
    COALESCE(SUM(failed_rows), 0) as total_failed_rows
FROM import_jobs
WHERE created_by = $1
  AND created_at >= $2;

-- name: CleanupOldJobs :exec
DELETE FROM import_jobs
WHERE status IN ('completed', 'failed', 'cancelled')
  AND completed_at < $1;