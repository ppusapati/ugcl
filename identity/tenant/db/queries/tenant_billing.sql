-- Tenant Billing Queries

-- name: CreateTenantBilling :one
INSERT INTO tenant_billing (
    tenant_id, billing_email, payment_method_id, subscription_id,
    billing_address, tax_id, billing_currency
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetTenantBilling :one
SELECT * FROM tenant_billing
WHERE tenant_id = $1;

-- name: UpdateTenantBilling :one
UPDATE tenant_billing
SET billing_email = $2, payment_method_id = $3, subscription_id = $4,
    billing_address = $5, tax_id = $6, billing_currency = $7
WHERE tenant_id = $1
RETURNING *;

-- name: UpdateTenantPaymentMethod :exec
UPDATE tenant_billing
SET payment_method_id = $2
WHERE tenant_id = $1;

-- name: UpdateTenantSubscription :exec
UPDATE tenant_billing
SET subscription_id = $2
WHERE tenant_id = $1;

-- name: GetTenantsByPaymentMethod :many
SELECT tb.*, t.name as tenant_name, t.display_name
FROM tenant_billing tb
JOIN tenants t ON tb.tenant_id = t.id
WHERE tb.payment_method_id = $1 AND t.is_active = true;

-- name: GetTenantsBySubscription :many
SELECT tb.*, t.name as tenant_name, t.display_name
FROM tenant_billing tb
JOIN tenants t ON tb.tenant_id = t.id
WHERE tb.subscription_id = $1 AND t.is_active = true;

-- name: DeleteTenantBilling :exec
DELETE FROM tenant_billing
WHERE tenant_id = $1;