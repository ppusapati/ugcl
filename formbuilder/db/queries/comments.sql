

-- =============================================================================
-- COMMENT QUERIES
-- =============================================================================

-- name: CreateComment :one
INSERT INTO comments (
    id, instance_id, user_id, text, created_at, internal
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetComments :many
SELECT * FROM comments 
WHERE instance_id = $1
ORDER BY created_at ASC;

-- name: GetPublicComments :many
SELECT * FROM comments 
WHERE instance_id = $1 AND internal = false
ORDER BY created_at ASC;

-- name: GetInternalComments :many
SELECT * FROM comments 
WHERE instance_id = $1 AND internal = true
ORDER BY created_at ASC;

-- name: UpdateComment :one
UPDATE comments SET
    text = $2
WHERE id = $1
RETURNING *;

-- name: DeleteComment :exec
DELETE FROM comments WHERE id = $1;

-- name: GetCommentsByUser :many
SELECT * FROM comments 
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
