-- =============================================================================
-- Report Tables Queries - InsightHub Service
-- =============================================================================

-- name: GetReportTables :many
SELECT
    rt.id,
    rt.report_id,
    rt.table_id,
    rt.table_alias,
    rt.data_source_id,
    rt.is_primary,
    rt.created_at
FROM report_tables rt
WHERE rt.report_id = $1
ORDER BY rt.is_primary DESC, rt.created_at ASC;

-- name: GetReportTableByID :one
SELECT
    rt.id,
    rt.report_id,
    rt.table_id,
    rt.table_alias,
    rt.data_source_id,
    rt.is_primary,
    rt.created_at
FROM report_tables rt
WHERE rt.id = $1;

-- name: GetPrimaryReportTable :one
SELECT
    rt.id,
    rt.report_id,
    rt.table_id,
    rt.table_alias,
    rt.data_source_id,
    rt.is_primary,
    rt.created_at
FROM report_tables rt
WHERE rt.report_id = $1 AND rt.is_primary = true;

-- name: CreateReportTable :one
INSERT INTO report_tables (
    report_id,
    table_id,
    table_alias,
    data_source_id,
    is_primary
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING
    id,
    report_id,
    table_id,
    table_alias,
    data_source_id,
    is_primary,
    created_at;

-- name: UpdateReportTable :one
UPDATE report_tables
SET
    table_alias = $2,
    is_primary = $3
WHERE id = $1
RETURNING
    id,
    report_id,
    table_id,
    table_alias,
    data_source_id,
    is_primary,
    created_at;

-- name: DeleteReportTable :exec
DELETE FROM report_tables
WHERE id = $1;

-- name: DeleteReportTablesByReport :exec
DELETE FROM report_tables
WHERE report_id = $1;

-- name: SetPrimaryTable :exec
-- First clear existing primary table for the report, then set the new one
UPDATE report_tables
SET is_primary = CASE WHEN id = $2 THEN true ELSE false END
WHERE report_id = $1;

-- name: GetReportTablesByDataSource :many
SELECT
    rt.id,
    rt.report_id,
    rt.table_id,
    rt.table_alias,
    rt.data_source_id,
    rt.is_primary,
    rt.created_at
FROM report_tables rt
WHERE rt.data_source_id = $1
ORDER BY rt.created_at DESC;