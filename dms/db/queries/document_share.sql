-- name: CreateDocumentShare :one
INSERT INTO document_shares (
    tenant_id,
    document_id,
    shared_with_entity_id,
    shared_with_user_id,
    shared_with_email,
    division_id,
    branch_id,
    department_id,
    can_view,
    can_download,
    can_edit,
    can_delete,
    share_link,
    expires_at,
    created_by,
    updated_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16
) RETURNING *;

-- name: GetDocumentShare :one
SELECT * FROM document_shares
WHERE id = $1 AND tenant_id = $2;

-- name: ListDocumentShares :many
SELECT * FROM document_shares
WHERE document_id = $1 AND tenant_id = $2
ORDER BY created_at DESC;

-- name: GetShareByLink :one
SELECT * FROM document_shares
WHERE share_link = $1 AND is_expired = FALSE;

-- name: UpdateShareAccessCount :exec
UPDATE document_shares
SET access_count = access_count + 1,
    last_accessed_at = NOW()
WHERE id = $1;

-- name: DeleteDocumentShare :exec
DELETE FROM document_shares
WHERE id = $1 AND tenant_id = $2;

-- name: GetUserDocumentShares :many
SELECT ds.*, d.file_name, d.original_name, d.mime_type, d.size_bytes
FROM document_shares ds
JOIN documents d ON ds.document_id = d.id
WHERE ds.tenant_id = $1
  AND (ds.shared_with_user_id = $2 OR ds.shared_with_entity_id = $3)
  AND ds.is_expired = FALSE
ORDER BY ds.created_at DESC;
