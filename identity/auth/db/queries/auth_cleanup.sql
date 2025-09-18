-- Cleanup and Maintenance Queries

-- name: CleanupExpiredPasswordResetTokens :exec
DELETE FROM auth_password_reset_tokens
WHERE expires_at <= CURRENT_TIMESTAMP;

-- name: CleanupUsedPasswordResetTokens :exec
DELETE FROM auth_password_reset_tokens
WHERE used_at IS NOT NULL AND used_at <= $1;

-- name: CleanupExpiredEmailVerificationTokens :exec
DELETE FROM auth_email_verification_tokens
WHERE expires_at <= CURRENT_TIMESTAMP;

-- name: CleanupVerifiedEmailTokens :exec
DELETE FROM auth_email_verification_tokens
WHERE verified_at IS NOT NULL AND verified_at <= $1;

-- name: CleanupExpiredPhoneVerificationTokens :exec
DELETE FROM auth_phone_verification_tokens
WHERE expires_at <= CURRENT_TIMESTAMP;

-- name: CleanupVerifiedPhoneTokens :exec
DELETE FROM auth_phone_verification_tokens
WHERE verified_at IS NOT NULL AND verified_at <= $1;

-- name: CleanupOldLoginAttempts :exec
DELETE FROM auth_login_attempts
WHERE attempted_at <= $1;

-- name: CleanupOldSecurityEvents :exec
DELETE FROM auth_security_events
WHERE created_at <= $1;

-- name: CleanupRevokedSessions :exec
DELETE FROM auth_sessions
WHERE is_active = false AND revoked_at <= $1;