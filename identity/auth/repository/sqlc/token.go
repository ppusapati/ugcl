package sqlc

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	sqlcgen "p9e.in/ugcl/identity/auth/db/generated"
	"p9e.in/ugcl/identity/auth/mappers"
	"p9e.in/ugcl/identity/auth/models"
	"p9e.in/ugcl/identity/auth/repository/interfaces"
)

type tokenRepository struct {
	db     *pgxpool.Pool
	tx     pgx.Tx
	mapper *mappers.TokenMapper
}

func NewTokenRepository(db *pgxpool.Pool, tx pgx.Tx) interfaces.TokenRepository {
	return &tokenRepository{
		db:     db,
		tx:     tx,
		mapper: mappers.NewTokenMapper(),
	}
}

func (r *tokenRepository) getQueries() *sqlcgen.Queries {
	if r.tx != nil {
		return sqlcgen.New(r.tx)
	}
	return sqlcgen.New(r.db)
}

// Password reset tokens
func (r *tokenRepository) CreatePasswordResetToken(ctx context.Context, token *models.PasswordResetToken) (*models.PasswordResetToken, error) {
	return nil, fmt.Errorf("CreatePasswordResetToken not implemented")
}

func (r *tokenRepository) GetPasswordResetToken(ctx context.Context, tokenHash string) (*models.PasswordResetToken, error) {
	return nil, fmt.Errorf("GetPasswordResetToken not implemented")
}

func (r *tokenRepository) MarkPasswordResetTokenUsed(ctx context.Context, tokenHash string) error {
	return fmt.Errorf("MarkPasswordResetTokenUsed not implemented")
}

func (r *tokenRepository) CleanupExpiredPasswordResetTokens(ctx context.Context) error {
	return fmt.Errorf("CleanupExpiredPasswordResetTokens not implemented")
}

func (r *tokenRepository) CleanupUsedPasswordResetTokens(ctx context.Context, olderThan time.Time) error {
	return fmt.Errorf("CleanupUsedPasswordResetTokens not implemented")
}

// Email verification tokens
func (r *tokenRepository) CreateEmailVerificationToken(ctx context.Context, token *models.EmailVerificationToken) (*models.EmailVerificationToken, error) {
	return nil, fmt.Errorf("CreateEmailVerificationToken not implemented")
}

func (r *tokenRepository) GetEmailVerificationToken(ctx context.Context, tokenHash string) (*models.EmailVerificationToken, error) {
	return nil, fmt.Errorf("GetEmailVerificationToken not implemented")
}

func (r *tokenRepository) GetEmailVerificationTokenByEmailAndCode(ctx context.Context, email, code string) (*models.EmailVerificationToken, error) {
	return nil, fmt.Errorf("GetEmailVerificationTokenByEmailAndCode not implemented")
}

func (r *tokenRepository) MarkEmailVerificationTokenUsed(ctx context.Context, tokenHash string) error {
	return fmt.Errorf("MarkEmailVerificationTokenUsed not implemented")
}

func (r *tokenRepository) CleanupExpiredEmailVerificationTokens(ctx context.Context) error {
	return fmt.Errorf("CleanupExpiredEmailVerificationTokens not implemented")
}

func (r *tokenRepository) CleanupVerifiedEmailTokens(ctx context.Context, olderThan time.Time) error {
	return fmt.Errorf("CleanupVerifiedEmailTokens not implemented")
}

// Phone verification tokens
func (r *tokenRepository) CreatePhoneVerificationToken(ctx context.Context, token *models.PhoneVerificationToken) (*models.PhoneVerificationToken, error) {
	return nil, fmt.Errorf("CreatePhoneVerificationToken not implemented")
}

func (r *tokenRepository) GetPhoneVerificationToken(ctx context.Context, phone string, otpCode string) (*models.PhoneVerificationToken, error) {
	return nil, fmt.Errorf("GetPhoneVerificationToken not implemented")
}

func (r *tokenRepository) IncrementPhoneVerificationAttempts(ctx context.Context, tokenID uuid.UUID) error {
	return fmt.Errorf("IncrementPhoneVerificationAttempts not implemented")
}

func (r *tokenRepository) MarkPhoneVerificationTokenUsed(ctx context.Context, tokenID uuid.UUID) error {
	return fmt.Errorf("MarkPhoneVerificationTokenUsed not implemented")
}

func (r *tokenRepository) CleanupExpiredPhoneVerificationTokens(ctx context.Context) error {
	return fmt.Errorf("CleanupExpiredPhoneVerificationTokens not implemented")
}

func (r *tokenRepository) DeleteExpiredPhoneVerificationTokens(ctx context.Context) any {
	return r.CleanupExpiredPhoneVerificationTokens(ctx)
}

func (r *tokenRepository) DeleteExpiredEmailVerificationTokens(ctx context.Context) any {
	return r.CleanupExpiredEmailVerificationTokens(ctx)
}

func (r *tokenRepository) CleanupVerifiedPhoneTokens(ctx context.Context, olderThan time.Time) error {
	return fmt.Errorf("CleanupVerifiedPhoneTokens not implemented")
}

// Revoked tokens
func (r *tokenRepository) RevokeToken(ctx context.Context, token *models.RevokedToken) error {
	return fmt.Errorf("RevokeToken not implemented")
}

func (r *tokenRepository) IsTokenRevoked(ctx context.Context, tokenJti string) (bool, error) {
	return false, fmt.Errorf("IsTokenRevoked not implemented")
}

func (r *tokenRepository) CleanupExpiredRevokedTokens(ctx context.Context) error {
	return fmt.Errorf("CleanupExpiredRevokedTokens not implemented")
}

// Token validation helpers
func (r *tokenRepository) IsValidPasswordResetToken(ctx context.Context, tokenHash string) (bool, error) {
	return false, fmt.Errorf("IsValidPasswordResetToken not implemented")
}

func (r *tokenRepository) IsValidEmailVerificationToken(ctx context.Context, tokenHash string) (bool, error) {
	return false, fmt.Errorf("IsValidEmailVerificationToken not implemented")
}

func (r *tokenRepository) IsValidPhoneVerificationToken(ctx context.Context, phone string, otpCode string) (bool, error) {
	return false, fmt.Errorf("IsValidPhoneVerificationToken not implemented")
}