-- queries.sql

-- name: CreateContractor :one
INSERT INTO contractors (
    company_name,
    company_type,
    gst,
    pan,
    category,
    person_id,
    associated_project,
    working_site,
    contract_start_date,
    contract_end_date,
    status,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING *;

-- name: GetContractorByID :one
SELECT * FROM contractors 
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetContractorByPersonID :one
SELECT * FROM contractors 
WHERE person_id = $1 AND deleted_at IS NULL;

-- name: GetContractorByPAN :one
SELECT * FROM contractors 
WHERE pan = $1 AND deleted_at IS NULL;

-- name: GetContractorByGST :one
SELECT * FROM contractors 
WHERE gst = $1 AND deleted_at IS NULL;

-- name: UpdateContractor :one
UPDATE contractors 
SET 
    company_name = COALESCE(sqlc.narg('company_name'), company_name),
    company_type = COALESCE(sqlc.narg('company_type'), company_type),
    gst = COALESCE(sqlc.narg('gst'), gst),
    pan = COALESCE(sqlc.narg('pan'), pan),
    category = COALESCE(sqlc.narg('category'), category),
    person_id = COALESCE(sqlc.narg('person_id'), person_id),
    associated_project = COALESCE(sqlc.narg('associated_project'), associated_project),
    working_site = COALESCE(sqlc.narg('working_site'), working_site),
    contract_start_date = COALESCE(sqlc.narg('contract_start_date'), contract_start_date),
    contract_end_date = COALESCE(sqlc.narg('contract_end_date'), contract_end_date),
    status = COALESCE(sqlc.narg('status'), status),
    metadata = COALESCE(sqlc.narg('metadata'), metadata)
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING *;

-- name: UpdateContractorPersonID :one
UPDATE contractors 
SET person_id = $2
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteContractor :exec
UPDATE contractors 
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListContractors :many
SELECT * FROM contractors
WHERE deleted_at IS NULL
    AND ($1::text IS NULL OR company_name ILIKE '%' || $1 || '%')
    AND ($2::text IS NULL OR category = $2)
    AND ($3::text IS NULL OR status = $3)
    AND ($4::text[] IS NULL OR associated_project && $4)
    AND ($5::text[] IS NULL OR working_site && $5)
ORDER BY created_at DESC
LIMIT $6 OFFSET $7;

-- name: CountContractors :one
SELECT COUNT(*) FROM contractors
WHERE deleted_at IS NULL
    AND ($1::text IS NULL OR company_name ILIKE '%' || $1 || '%')
    AND ($2::text IS NULL OR category = $2)
    AND ($3::text IS NULL OR status = $3)
    AND ($4::text[] IS NULL OR associated_project && $4)
    AND ($5::text[] IS NULL OR working_site && $5);

-- name: ListContractorsByProject :many
SELECT * FROM contractors
WHERE deleted_at IS NULL
    AND $1 = ANY(associated_project)
ORDER BY company_name ASC;

-- name: ListContractorsBySite :many
SELECT * FROM contractors
WHERE deleted_at IS NULL
    AND $1 = ANY(working_site)
ORDER BY company_name ASC;
