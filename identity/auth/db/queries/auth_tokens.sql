-- Auth Token Management Queries

-- name: CreatePasswordResetToken :one
INSERT INTO auth_password_reset_tokens (
    user_id, token_hash, expires_at, ip_address, user_agent
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetPasswordResetToken :one
SELECT * FROM auth_password_reset_tokens
WHERE token_hash = $1 AND expires_at > CURRENT_TIMESTAMP AND used_at IS NULL;

-- name: MarkPasswordResetTokenUsed :exec
UPDATE auth_password_reset_tokens
SET used_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: CreateEmailVerificationToken :one
INSERT INTO auth_email_verification_tokens (
    user_id, email, token_hash, expires_at
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetEmailVerificationToken :one
SELECT * FROM auth_email_verification_tokens
WHERE token_hash = $1 AND expires_at > CURRENT_TIMESTAMP AND verified_at IS NULL;

-- name: MarkEmailVerified :exec
UPDATE auth_email_verification_tokens
SET verified_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: CreatePhoneVerificationToken :one
INSERT INTO auth_phone_verification_tokens (
    user_id, phone, otp_code, expires_at
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetPhoneVerificationToken :one
SELECT * FROM auth_phone_verification_tokens
WHERE user_id = $1 AND phone = $2 AND otp_code = $3
AND expires_at > CURRENT_TIMESTAMP AND verified_at IS NULL;

-- name: IncrementPhoneVerificationAttempts :exec
UPDATE auth_phone_verification_tokens
SET attempts = attempts + 1
WHERE id = $1;

-- name: MarkPhoneVerified :exec
UPDATE auth_phone_verification_tokens
SET verified_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: RevokeToken :one
INSERT INTO auth_revoked_tokens (
    token_jti, token_type, user_id, expires_at, reason
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: IsTokenRevoked :one
SELECT EXISTS(
    SELECT 1 FROM auth_revoked_tokens
    WHERE token_jti = $1 AND expires_at > CURRENT_TIMESTAMP
);

-- name: CleanupExpiredTokens :exec
DELETE FROM auth_revoked_tokens WHERE expires_at <= CURRENT_TIMESTAMP;