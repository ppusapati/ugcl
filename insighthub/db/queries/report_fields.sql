-- =============================================================================
-- Report Fields Queries - InsightHub Service
-- =============================================================================

-- name: GetReportFields :many
SELECT
    rf.id,
    rf.report_id,
    rf.column_id,
    rf.report_table_id,
    rf.field_alias,
    rf.aggregate_function,
    rf.order_index,
    rf.is_visible,
    rf.formatting_rules,
    rf.created_at
FROM report_fields rf
WHERE rf.report_id = $1
ORDER BY rf.order_index;

-- name: GetReportFieldByID :one
SELECT
    rf.id,
    rf.report_id,
    rf.column_id,
    rf.report_table_id,
    rf.field_alias,
    rf.aggregate_function,
    rf.order_index,
    rf.is_visible,
    rf.formatting_rules,
    rf.created_at
FROM report_fields rf
WHERE rf.id = $1;

-- name: CreateReportField :one
INSERT INTO report_fields (
    report_id,
    column_id,
    report_table_id,
    field_alias,
    aggregate_function,
    order_index,
    is_visible,
    formatting_rules
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING
    id,
    report_id,
    column_id,
    report_table_id,
    field_alias,
    aggregate_function,
    order_index,
    is_visible,
    formatting_rules,
    created_at;

-- name: UpdateReportField :one
UPDATE report_fields
SET
    field_alias = $2,
    aggregate_function = $3,
    order_index = $4,
    is_visible = $5,
    formatting_rules = $6
WHERE id = $1
RETURNING
    id,
    report_id,
    column_id,
    report_table_id,
    field_alias,
    aggregate_function,
    order_index,
    is_visible,
    formatting_rules,
    created_at;

-- name: DeleteReportField :exec
DELETE FROM report_fields
WHERE id = $1;

-- name: DeleteReportFieldsByReport :exec
DELETE FROM report_fields
WHERE report_id = $1;

-- name: UpdateReportFieldsOrder :exec
UPDATE report_fields
SET order_index = $2
WHERE id = $1;

-- name: GetMaxOrderIndex :one
SELECT COALESCE(MAX(order_index), 0) as max_order
FROM report_fields
WHERE report_id = $1;