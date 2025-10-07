-- queries.sql

-- name: CreateVendor :one
INSERT INTO vendors (
    company_name,
    company_type,
    gst,
    pan,
    vendor_category,
    person_id,
    payment_terms,
    credit_period_days,
    bank_name,
    account_number,
    ifsc,
    rating,
    is_blacklisted,
    contracts,
    purchase_orders,
    status,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
) RETURNING *;

-- name: GetVendorByID :one
SELECT * FROM vendors
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetVendorByUUID :one
SELECT * FROM vendors
WHERE uuid = $1 AND deleted_at IS NULL;

-- name: GetVendorByPersonID :one
SELECT * FROM vendors
WHERE person_id = $1 AND deleted_at IS NULL;

-- name: GetVendorByPAN :one
SELECT * FROM vendors
WHERE pan = $1 AND deleted_at IS NULL;

-- name: GetVendorByGST :one
SELECT * FROM vendors
WHERE gst = $1 AND deleted_at IS NULL;

-- name: UpdateVendor :one
UPDATE vendors
SET
    company_name = COALESCE(sqlc.narg('company_name'), company_name),
    company_type = COALESCE(sqlc.narg('company_type'), company_type),
    gst = COALESCE(sqlc.narg('gst'), gst),
    pan = COALESCE(sqlc.narg('pan'), pan),
    vendor_category = COALESCE(sqlc.narg('vendor_category'), vendor_category),
    person_id = COALESCE(sqlc.narg('person_id'), person_id),
    payment_terms = COALESCE(sqlc.narg('payment_terms'), payment_terms),
    credit_period_days = COALESCE(sqlc.narg('credit_period_days'), credit_period_days),
    bank_name = COALESCE(sqlc.narg('bank_name'), bank_name),
    account_number = COALESCE(sqlc.narg('account_number'), account_number),
    ifsc = COALESCE(sqlc.narg('ifsc'), ifsc),
    rating = COALESCE(sqlc.narg('rating'), rating),
    is_blacklisted = COALESCE(sqlc.narg('is_blacklisted'), is_blacklisted),
    contracts = COALESCE(sqlc.narg('contracts'), contracts),
    purchase_orders = COALESCE(sqlc.narg('purchase_orders'), purchase_orders),
    status = COALESCE(sqlc.narg('status'), status),
    metadata = COALESCE(sqlc.narg('metadata'), metadata)
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING *;

-- name: UpdateVendorPersonID :one
UPDATE vendors
SET person_id = $2
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteVendor :exec
UPDATE vendors
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListVendors :many
SELECT * FROM vendors
WHERE deleted_at IS NULL
    AND ($1::text IS NULL OR company_name ILIKE '%' || $1 || '%')
    AND ($2::text IS NULL OR vendor_category = $2)
    AND ($3::text IS NULL OR status = $3)
    AND ($4::int IS NULL OR rating >= $4)
    AND ($5::int IS NULL OR rating <= $5)
    AND ($6::bool IS NULL OR is_blacklisted = $6)
ORDER BY created_at DESC
LIMIT $7 OFFSET $8;

-- name: CountVendors :one
SELECT COUNT(*) FROM vendors
WHERE deleted_at IS NULL
    AND ($1::text IS NULL OR company_name ILIKE '%' || $1 || '%')
    AND ($2::text IS NULL OR vendor_category = $2)
    AND ($3::text IS NULL OR status = $3)
    AND ($4::int IS NULL OR rating >= $4)
    AND ($5::int IS NULL OR rating <= $5)
    AND ($6::bool IS NULL OR is_blacklisted = $6);

-- name: ListVendorsByCategory :many
SELECT * FROM vendors
WHERE deleted_at IS NULL
    AND vendor_category = $1
ORDER BY company_name ASC;

-- name: ListVendorsByContract :many
SELECT * FROM vendors
WHERE deleted_at IS NULL
    AND $1 = ANY(contracts)
ORDER BY company_name ASC;

-- name: ListVendorsByPurchaseOrder :many
SELECT * FROM vendors
WHERE deleted_at IS NULL
    AND $1 = ANY(purchase_orders)
ORDER BY company_name ASC;

-- name: ListBlacklistedVendors :many
SELECT * FROM vendors
WHERE deleted_at IS NULL
    AND is_blacklisted = TRUE
ORDER BY company_name ASC;
