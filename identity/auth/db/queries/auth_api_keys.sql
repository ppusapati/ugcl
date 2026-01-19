-- API Keys Queries

-- name: CreateAPIKey :one
INSERT INTO auth_api_keys (
    key_id, key_hash, name, tenant_id, scopes, expires_at, created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetAPIKeyByID :one
SELECT * FROM auth_api_keys
WHERE key_id = $1 AND is_active = true;

-- name: GetAPIKeysByTenant :many
SELECT * FROM auth_api_keys
WHERE tenant_id = $1 AND is_active = true
ORDER BY created_at DESC;

-- name: GetAPIKeysByUser :many
SELECT * FROM auth_api_keys
WHERE created_by = $1 AND is_active = true
ORDER BY created_at DESC;

-- name: UpdateAPIKeyLastUsed :exec
UPDATE auth_api_keys
SET last_used_at = CURRENT_TIMESTAMP
WHERE key_id = $1;

-- name: DeactivateAPIKey :exec
UPDATE auth_api_keys
SET is_active = false
WHERE key_id = $1;

-- name: UpdateAPIKeyScopes :exec
UPDATE auth_api_keys
SET scopes = $2
WHERE key_id = $1;

-- name: CleanupExpiredAPIKeys :exec
UPDATE auth_api_keys
SET is_active = false
WHERE expires_at <= CURRENT_TIMESTAMP AND is_active = true;