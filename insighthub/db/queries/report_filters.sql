-- =============================================================================
-- Report Filters Queries - InsightHub Service
-- =============================================================================

-- name: GetReportFilters :many
SELECT
    rf.id,
    rf.report_id,
    rf.column_id,
    rf.report_table_id,
    rf.operator,
    rf.value,
    rf.value_type,
    rf.logical_operator,
    rf.group_index,
    rf.order_index,
    rf.is_active,
    rf.created_at
FROM report_filters rf
WHERE rf.report_id = $1 AND rf.is_active = true
ORDER BY rf.group_index, rf.order_index;

-- name: CreateReportFilter :one
INSERT INTO report_filters (
    report_id,
    column_id,
    report_table_id,
    operator,
    value,
    value_type,
    logical_operator,
    group_index,
    order_index
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING
    id,
    report_id,
    column_id,
    report_table_id,
    operator,
    value,
    value_type,
    logical_operator,
    group_index,
    order_index,
    is_active,
    created_at;

-- name: UpdateReportFilter :one
UPDATE report_filters
SET
    operator = $2,
    value = $3,
    value_type = $4,
    logical_operator = $5,
    group_index = $6,
    order_index = $7
WHERE id = $1 AND is_active = true
RETURNING
    id,
    report_id,
    column_id,
    report_table_id,
    operator,
    value,
    value_type,
    logical_operator,
    group_index,
    order_index,
    is_active,
    created_at;

-- name: DeleteReportFilter :exec
UPDATE report_filters
SET is_active = false
WHERE id = $1;

-- name: DeleteReportFiltersByReport :exec
UPDATE report_filters
SET is_active = false
WHERE report_id = $1;