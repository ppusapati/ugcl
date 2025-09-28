-- name: CreateBusinessTerm :one
INSERT INTO business_terms (
    term, definition, business_context, domain, owner_user_id, created_by
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetBusinessTerm :one
SELECT * FROM business_terms
WHERE id = $1 AND is_active = true;

-- name: GetBusinessTermByName :one
SELECT * FROM business_terms
WHERE term = $1 AND is_active = true;

-- name: ListBusinessTerms :many
SELECT * FROM business_terms
WHERE is_active = true
ORDER BY term;

-- name: ListBusinessTermsByDomain :many
SELECT * FROM business_terms
WHERE domain = $1 AND is_active = true
ORDER BY term;

-- name: SearchBusinessTerms :many
SELECT * FROM business_terms
WHERE is_active = true
  AND (term ILIKE '%' || $1 || '%' OR definition ILIKE '%' || $1 || '%' OR business_context ILIKE '%' || $1 || '%')
ORDER BY term;

-- name: UpdateBusinessTerm :one
UPDATE business_terms
SET definition = $2, business_context = $3, domain = $4, owner_user_id = $5, updated_by = $6, updated_at = NOW()
WHERE id = $1 AND is_active = true
RETURNING *;

-- name: DeleteBusinessTerm :exec
UPDATE business_terms
SET is_active = false, updated_by = $2, updated_at = NOW()
WHERE id = $1;

-- name: LinkColumnToBusinessTerm :one
INSERT INTO column_business_terms (
    column_id, business_term_id, created_by
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: UnlinkColumnFromBusinessTerm :exec
DELETE FROM column_business_terms
WHERE column_id = $1 AND business_term_id = $2;

-- name: GetBusinessTermsForColumn :many
SELECT bt.* FROM business_terms bt
JOIN column_business_terms cbt ON bt.id = cbt.business_term_id
WHERE cbt.column_id = $1 AND bt.is_active = true
ORDER BY bt.term;

-- name: GetColumnsForBusinessTerm :many
SELECT c.*, t.table_name, t.display_name as table_display_name,
       s.schema_name, s.display_name as schema_display_name
FROM columns_metadata c
JOIN column_business_terms cbt ON c.id = cbt.column_id
JOIN tables_metadata t ON c.table_id = t.id
JOIN schemas_metadata s ON t.schema_id = s.id
WHERE cbt.business_term_id = $1 AND c.is_active = true AND t.is_active = true AND s.is_active = true
ORDER BY s.display_name, t.display_name, c.column_order;