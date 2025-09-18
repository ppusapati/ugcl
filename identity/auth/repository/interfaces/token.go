package interfaces

import (
	"context"
	"time"

	"github.com/google/uuid"

	"p9e.in/ugcl/identity/auth/models"
)

type TokenRepository interface {
	// Password reset tokens
	CreatePasswordResetToken(ctx context.Context, token *models.PasswordResetToken) (*models.PasswordResetToken, error)
	GetPasswordResetToken(ctx context.Context, tokenHash string) (*models.PasswordResetToken, error)
	MarkPasswordResetTokenUsed(ctx context.Context, tokenHash string) error
	CleanupExpiredPasswordResetTokens(ctx context.Context) error
	CleanupUsedPasswordResetTokens(ctx context.Context, olderThan time.Time) error

	// Email verification tokens
	CreateEmailVerificationToken(ctx context.Context, token *models.EmailVerificationToken) (*models.EmailVerificationToken, error)
	GetEmailVerificationToken(ctx context.Context, tokenHash string) (*models.EmailVerificationToken, error)
	GetEmailVerificationTokenByEmailAndCode(ctx context.Context, email, code string) (*models.EmailVerificationToken, error)
	MarkEmailVerificationTokenUsed(ctx context.Context, tokenHash string) error
	CleanupExpiredEmailVerificationTokens(ctx context.Context) error
	CleanupVerifiedEmailTokens(ctx context.Context, olderThan time.Time) error

	// Phone verification tokens
	CreatePhoneVerificationToken(ctx context.Context, token *models.PhoneVerificationToken) (*models.PhoneVerificationToken, error)
	GetPhoneVerificationToken(ctx context.Context, phone string, otpCode string) (*models.PhoneVerificationToken, error)
	IncrementPhoneVerificationAttempts(ctx context.Context, tokenID uuid.UUID) error
	MarkPhoneVerificationTokenUsed(ctx context.Context, tokenID uuid.UUID) error
	CleanupExpiredPhoneVerificationTokens(ctx context.Context) error
	CleanupVerifiedPhoneTokens(ctx context.Context, olderThan time.Time) error

	// Revoked tokens
	RevokeToken(ctx context.Context, token *models.RevokedToken) error
	IsTokenRevoked(ctx context.Context, tokenJti string) (bool, error)
	CleanupExpiredRevokedTokens(ctx context.Context) error

	// Token validation helpers
	IsValidPasswordResetToken(ctx context.Context, tokenHash string) (bool, error)
	IsValidEmailVerificationToken(ctx context.Context, tokenHash string) (bool, error)
	IsValidPhoneVerificationToken(ctx context.Context, phone string, otpCode string) (bool, error)
	DeleteExpiredEmailVerificationTokens(ctx context.Context) any
	DeleteExpiredPhoneVerificationTokens(ctx context.Context) any
}

type TwoFactorRepository interface {
	// Backup codes
	CreateBackupCodes(ctx context.Context, userID uuid.UUID, codeHashes []string) error
	UseBackupCode(ctx context.Context, userID uuid.UUID, codeHash string) (*models.TwoFactorBackupCode, error)
	GetUnusedBackupCodesCount(ctx context.Context, userID uuid.UUID) (int64, error)
	DeleteAllBackupCodes(ctx context.Context, userID uuid.UUID) error

	// Two-factor management
	SetSecret(ctx context.Context, userID uuid.UUID, secret string) error
	GetSecret(ctx context.Context, userID uuid.UUID) (string, error)
	EnableTwoFactor(ctx context.Context, userID uuid.UUID) error
	DisableTwoFactor(ctx context.Context, userID uuid.UUID) error
	IsTwoFactorEnabled(ctx context.Context, userID uuid.UUID) (bool, error)
}
