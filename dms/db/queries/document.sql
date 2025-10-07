-- name: CreateDocument :one
INSERT INTO documents (
    tenant_id,
    owner_entity_type,
    owner_entity_id,
    division_id,
    branch_id,
    department_id,
    document_type,
    document_category,
    file_name,
    original_name,
    title,
    description,
    mime_type,
    file_extension,
    size_bytes,
    checksum,
    storage_path,
    uploaded_by,
    created_by,
    updated_by,
    expires_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21
) RETURNING *;

-- name: GetDocument :one
SELECT * FROM documents
WHERE id = $1 AND tenant_id = $2 AND is_deleted = FALSE;

-- name: ListDocuments :many
SELECT * FROM documents
WHERE tenant_id = $1
  AND is_deleted = FALSE
  AND ($2::varchar IS NULL OR owner_entity_type = $2)
  AND ($3::uuid IS NULL OR owner_entity_id = $3)
  AND ($4::uuid IS NULL OR division_id = $4)
  AND ($5::uuid IS NULL OR branch_id = $5)
  AND ($6::uuid IS NULL OR department_id = $6)
  AND ($7::varchar IS NULL OR document_type::text = $7)
  AND ($8::varchar IS NULL OR document_category::text = $8)
ORDER BY created_at DESC
LIMIT $9 OFFSET $10;

-- name: CountDocuments :one
SELECT COUNT(*) FROM documents
WHERE tenant_id = $1
  AND is_deleted = FALSE
  AND ($2::varchar IS NULL OR owner_entity_type = $2)
  AND ($3::uuid IS NULL OR owner_entity_id = $3)
  AND ($4::uuid IS NULL OR division_id = $4)
  AND ($5::uuid IS NULL OR branch_id = $5)
  AND ($6::uuid IS NULL OR department_id = $6)
  AND ($7::varchar IS NULL OR document_type::text = $7)
  AND ($8::varchar IS NULL OR document_category::text = $8);

-- name: UpdateDocument :one
UPDATE documents
SET
    title = COALESCE(sqlc.narg('title'), title),
    description = COALESCE(sqlc.narg('description'), description),
    document_type = COALESCE(sqlc.narg('document_type'), document_type),
    document_category = COALESCE(sqlc.narg('document_category'), document_category),
    division_id = COALESCE(sqlc.narg('division_id'), division_id),
    branch_id = COALESCE(sqlc.narg('branch_id'), branch_id),
    department_id = COALESCE(sqlc.narg('department_id'), department_id),
    expires_at = COALESCE(sqlc.narg('expires_at'), expires_at),
    updated_at = NOW(),
    updated_by = $1
WHERE id = $2 AND tenant_id = $3
RETURNING *;

-- name: SoftDeleteDocument :exec
UPDATE documents
SET is_deleted = TRUE,
    deleted_at = NOW(),
    deleted_by = $1,
    updated_at = NOW(),
    updated_by = $1
WHERE id = $2 AND tenant_id = $3;

-- name: GetExpiredDocuments :many
SELECT * FROM documents
WHERE tenant_id = $1
  AND is_expired = TRUE
  AND is_deleted = FALSE
ORDER BY expires_at;

-- name: SearchDocumentsByTags :many
SELECT * FROM documents
WHERE tenant_id = $1
  AND is_deleted = FALSE
  AND tags && $2::text[]
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;
