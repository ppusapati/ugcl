-- =============================================================================
-- Report Results Queries - InsightViewer Service
-- =============================================================================

-- name: CreateReportResult :one
INSERT INTO report_results (
    run_id,
    result_type,
    result_json,
    result_text,
    cache_ref,
    file_path,
    column_info,
    row_count,
    size_bytes,
    compression,
    expires_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING
    id,
    run_id,
    result_type,
    result_json,
    result_text,
    cache_ref,
    file_path,
    column_info,
    row_count,
    size_bytes,
    compression,
    created_at,
    expires_at,
    access_count,
    last_accessed_at;

-- name: GetReportResultByRunID :one
SELECT
    id,
    run_id,
    result_type,
    result_json,
    result_text,
    cache_ref,
    file_path,
    column_info,
    row_count,
    size_bytes,
    compression,
    created_at,
    expires_at,
    access_count,
    last_accessed_at
FROM report_results
WHERE run_id = $1 AND result_type = 'data';

-- name: GetReportResultByID :one
SELECT
    id,
    run_id,
    result_type,
    result_json,
    result_text,
    cache_ref,
    file_path,
    column_info,
    row_count,
    size_bytes,
    compression,
    created_at,
    expires_at,
    access_count,
    last_accessed_at
FROM report_results
WHERE id = $1;

-- name: UpdateResultAccess :exec
UPDATE report_results
SET
    access_count = access_count + 1,
    last_accessed_at = NOW()
WHERE id = $1;

-- name: GetReportResultsByCacheRef :one
SELECT
    id,
    run_id,
    result_type,
    result_json,
    result_text,
    cache_ref,
    file_path,
    column_info,
    row_count,
    size_bytes,
    compression,
    created_at,
    expires_at,
    access_count,
    last_accessed_at
FROM report_results
WHERE cache_ref = $1 AND (expires_at IS NULL OR expires_at > NOW());

-- name: DeleteExpiredResults :exec
DELETE FROM report_results
WHERE expires_at IS NOT NULL AND expires_at <= NOW();

-- name: DeleteReportResultsByRunID :exec
DELETE FROM report_results
WHERE run_id = $1;

-- name: GetResultStorageStatistics :one
SELECT
    COUNT(*) as total_results,
    COUNT(CASE WHEN result_json IS NOT NULL THEN 1 END) as json_results,
    COUNT(CASE WHEN result_text IS NOT NULL THEN 1 END) as text_results,
    COUNT(CASE WHEN cache_ref IS NOT NULL THEN 1 END) as cached_results,
    COUNT(CASE WHEN file_path IS NOT NULL THEN 1 END) as file_results,
    SUM(size_bytes) as total_size_bytes,
    AVG(size_bytes) as avg_size_bytes
FROM report_results
WHERE created_at >= $1;

-- name: GetLargestResults :many
SELECT
    id,
    run_id,
    result_type,
    size_bytes,
    compression,
    created_at,
    access_count,
    last_accessed_at
FROM report_results
ORDER BY size_bytes DESC
LIMIT $1;

-- name: GetMostAccessedResults :many
SELECT
    id,
    run_id,
    result_type,
    size_bytes,
    access_count,
    created_at,
    last_accessed_at
FROM report_results
WHERE created_at >= $1
ORDER BY access_count DESC
LIMIT $2;