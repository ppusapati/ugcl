
-- =============================================================================
-- ATTACHMENT QUERIES
-- =============================================================================

-- name: CreateAttachment :one
INSERT INTO attachments (
    id, instance_id, field_id, filename, mime_type, size,
    storage_path, uploaded_at, uploaded_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetAttachments :many
SELECT * FROM attachments 
WHERE instance_id = $1
ORDER BY uploaded_at DESC;

-- name: GetAttachmentsByField :many
SELECT * FROM attachments 
WHERE instance_id = $1 AND field_id = $2
ORDER BY uploaded_at DESC;

-- name: GetAttachment :one
SELECT * FROM attachments 
WHERE id = $1;

-- name: DeleteAttachment :exec
DELETE FROM attachments WHERE id = $1;

-- name: GetAttachmentsByUser :many
SELECT * FROM attachments 
WHERE uploaded_by = $1
ORDER BY uploaded_at DESC
LIMIT $2 OFFSET $3;
