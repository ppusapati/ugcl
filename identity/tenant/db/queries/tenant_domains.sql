-- Tenant Domains Queries

-- name: CreateTenantDomain :one
INSERT INTO tenant_domains (
    tenant_id, domain, is_primary, verification_token
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetTenantDomain :one
SELECT * FROM tenant_domains
WHERE tenant_id = $1 AND domain = $2;

-- name: GetTenantDomains :many
SELECT * FROM tenant_domains
WHERE tenant_id = $1
ORDER BY is_primary DESC, created_at ASC;

-- name: GetTenantByDomain :one
SELECT td.*, t.name as tenant_name, t.display_name
FROM tenant_domains td
JOIN tenants t ON td.tenant_id = t.id
WHERE td.domain = $1 AND td.is_verified = true AND t.is_active = true;

-- name: GetPrimaryDomain :one
SELECT * FROM tenant_domains
WHERE tenant_id = $1 AND is_primary = true;

-- name: SetPrimaryDomain :exec
UPDATE tenant_domains
SET is_primary = CASE WHEN domain = $2 THEN true ELSE false END
WHERE tenant_id = $1;

-- name: VerifyDomain :exec
UPDATE tenant_domains
SET is_verified = true, verified_at = CURRENT_TIMESTAMP
WHERE tenant_id = $1 AND domain = $2;

-- name: UpdateDomainSSL :exec
UPDATE tenant_domains
SET ssl_certificate = $3, ssl_private_key = $4
WHERE tenant_id = $1 AND domain = $2;

-- name: DeleteTenantDomain :exec
DELETE FROM tenant_domains
WHERE tenant_id = $1 AND domain = $2;