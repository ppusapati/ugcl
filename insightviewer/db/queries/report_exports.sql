-- =============================================================================
-- Report Exports Queries - InsightViewer Service
-- =============================================================================

-- name: CreateReportExport :one
INSERT INTO report_exports (
    run_id,
    format,
    file_name,
    file_path,
    file_size,
    mime_type,
    generated_by,
    expires_at,
    export_options
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING
    id,
    run_id,
    format,
    file_name,
    file_path,
    file_size,
    mime_type,
    generated_at,
    generated_by,
    expires_at,
    download_count,
    last_downloaded_at,
    export_options,
    status,
    error_message;

-- name: GetReportExportByID :one
SELECT
    id,
    run_id,
    format,
    file_name,
    file_path,
    file_size,
    mime_type,
    generated_at,
    generated_by,
    expires_at,
    download_count,
    last_downloaded_at,
    export_options,
    status,
    error_message
FROM report_exports
WHERE id = $1;

-- name: GetReportExportsByRunID :many
SELECT
    id,
    run_id,
    format,
    file_name,
    file_path,
    file_size,
    mime_type,
    generated_at,
    generated_by,
    expires_at,
    download_count,
    last_downloaded_at,
    export_options,
    status,
    error_message
FROM report_exports
WHERE run_id = $1
ORDER BY generated_at DESC;

-- name: GetReportExportsByUser :many
SELECT
    id,
    run_id,
    format,
    file_name,
    file_path,
    file_size,
    mime_type,
    generated_at,
    generated_by,
    expires_at,
    download_count,
    last_downloaded_at,
    export_options,
    status,
    error_message
FROM report_exports
WHERE generated_by = $1
ORDER BY generated_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateExportStatus :one
UPDATE report_exports
SET
    status = $2,
    error_message = $3,
    file_size = COALESCE($4, file_size)
WHERE id = $1
RETURNING
    id,
    run_id,
    format,
    file_name,
    file_path,
    file_size,
    mime_type,
    generated_at,
    generated_by,
    expires_at,
    download_count,
    last_downloaded_at,
    export_options,
    status,
    error_message;

-- name: IncrementDownloadCount :exec
UPDATE report_exports
SET
    download_count = download_count + 1,
    last_downloaded_at = NOW()
WHERE id = $1;

-- name: DeleteExpiredExports :exec
DELETE FROM report_exports
WHERE expires_at IS NOT NULL AND expires_at <= NOW();

-- name: GetExportsByStatus :many
SELECT
    id,
    run_id,
    format,
    file_name,
    file_path,
    file_size,
    mime_type,
    generated_at,
    generated_by,
    expires_at,
    download_count,
    last_downloaded_at,
    export_options,
    status,
    error_message
FROM report_exports
WHERE status = $1
ORDER BY generated_at ASC;

-- name: GetExportStatistics :one
SELECT
    COUNT(*) as total_exports,
    COUNT(CASE WHEN status = 'ready' THEN 1 END) as ready_exports,
    COUNT(CASE WHEN status = 'generating' THEN 1 END) as generating_exports,
    COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_exports,
    SUM(file_size) as total_file_size,
    SUM(download_count) as total_downloads,
    COUNT(CASE WHEN format = 'csv' THEN 1 END) as csv_exports,
    COUNT(CASE WHEN format = 'excel' THEN 1 END) as excel_exports,
    COUNT(CASE WHEN format = 'pdf' THEN 1 END) as pdf_exports,
    COUNT(CASE WHEN format = 'json' THEN 1 END) as json_exports
FROM report_exports
WHERE generated_at >= $1;