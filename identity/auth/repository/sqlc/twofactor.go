package sqlc

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	sqlcgen "p9e.in/ugcl/identity/auth/db/generated"
	"p9e.in/ugcl/identity/auth/models"
	"p9e.in/ugcl/identity/auth/repository/interfaces"
)

type twoFactorRepository struct {
	db *pgxpool.Pool
	tx pgx.Tx
}

func NewTwoFactorRepository(db *pgxpool.Pool, tx pgx.Tx) interfaces.TwoFactorRepository {
	return &twoFactorRepository{
		db: db,
		tx: tx,
	}
}

func (r *twoFactorRepository) getQueries() *sqlcgen.Queries {
	if r.tx != nil {
		return sqlcgen.New(r.tx)
	}
	return sqlcgen.New(r.db)
}

func (r *twoFactorRepository) SetSecret(ctx context.Context, userID uuid.UUID, secret string) error {
	queries := r.getQueries()

	err := queries.EnableTwoFactor(ctx, sqlcgen.EnableTwoFactorParams{
		ID:              userID,
		TwoFactorSecret: &secret,
		BackupCodes:     []string{}, // Empty backup codes for now
	})
	if err != nil {
		return fmt.Errorf("failed to set two factor secret: %w", err)
	}

	return nil
}

func (r *twoFactorRepository) GetSecret(ctx context.Context, userID uuid.UUID) (string, error) {
	queries := r.getQueries()

	user, err := queries.GetAuthUserByID(ctx, sqlcgen.GetAuthUserByIDParams{
		ID: userID,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get user for two factor secret: %w", err)
	}

	if user.TwoFactorSecret == nil {
		return "", fmt.Errorf("two factor secret not set")
	}

	return *user.TwoFactorSecret, nil
}

func (r *twoFactorRepository) EnableTwoFactor(ctx context.Context, userID uuid.UUID) error {
	queries := r.getQueries()

	err := queries.EnableTwoFactor(ctx, sqlcgen.EnableTwoFactorParams{
		ID:              userID,
		TwoFactorSecret: nil, // Keep existing secret
		BackupCodes:     []string{}, // Empty backup codes for now
	})
	if err != nil {
		return fmt.Errorf("failed to enable two factor: %w", err)
	}

	return nil
}

func (r *twoFactorRepository) DisableTwoFactor(ctx context.Context, userID uuid.UUID) error {
	queries := r.getQueries()

	err := queries.DisableTwoFactor(ctx, sqlcgen.DisableTwoFactorParams{
		ID: userID,
	})
	if err != nil {
		return fmt.Errorf("failed to disable two factor: %w", err)
	}

	return nil
}

func (r *twoFactorRepository) IsTwoFactorEnabled(ctx context.Context, userID uuid.UUID) (bool, error) {
	queries := r.getQueries()

	user, err := queries.GetAuthUserByID(ctx, sqlcgen.GetAuthUserByIDParams{
		ID: userID,
	})
	if err != nil {
		return false, fmt.Errorf("failed to get user for two factor status: %w", err)
	}

	if user.TwoFactorEnabled == nil {
		return false, nil
	}

	return *user.TwoFactorEnabled, nil
}

// Backup code methods
func (r *twoFactorRepository) CreateBackupCodes(ctx context.Context, userID uuid.UUID, codeHashes []string) error {
	queries := r.getQueries()

	// Convert UUID to pgtype.UUID
	var pgUUID pgtype.UUID
	pgUUID.Scan(userID)

	for _, codeHash := range codeHashes {
		_, err := queries.CreateTwoFactorBackupCode(ctx, sqlcgen.CreateTwoFactorBackupCodeParams{
			UserID:   pgUUID,
			CodeHash: codeHash,
		})
		if err != nil {
			return fmt.Errorf("failed to create backup code: %w", err)
		}
	}

	return nil
}

func (r *twoFactorRepository) UseBackupCode(ctx context.Context, userID uuid.UUID, codeHash string) (*models.TwoFactorBackupCode, error) {
	queries := r.getQueries()

	// Convert UUID to pgtype.UUID
	var pgUUID pgtype.UUID
	pgUUID.Scan(userID)

	err := queries.UseTwoFactorBackupCode(ctx, sqlcgen.UseTwoFactorBackupCodeParams{
		UserID:   pgUUID,
		CodeHash: codeHash,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to use backup code: %w", err)
	}

	// Return a basic backup code model - this could be enhanced to return actual data
	return &models.TwoFactorBackupCode{
		UserID:   userID,
		CodeHash: codeHash,
	}, nil
}

func (r *twoFactorRepository) GetUnusedBackupCodesCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	queries := r.getQueries()

	// Convert UUID to pgtype.UUID
	var pgUUID pgtype.UUID
	pgUUID.Scan(userID)

	count, err := queries.CountUnusedBackupCodes(ctx, sqlcgen.CountUnusedBackupCodesParams{
		UserID: pgUUID,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to count unused backup codes: %w", err)
	}

	return count, nil
}

func (r *twoFactorRepository) DeleteAllBackupCodes(ctx context.Context, userID uuid.UUID) error {
	queries := r.getQueries()

	// Convert UUID to pgtype.UUID
	var pgUUID pgtype.UUID
	pgUUID.Scan(userID)

	err := queries.DeleteTwoFactorBackupCodes(ctx, sqlcgen.DeleteTwoFactorBackupCodesParams{
		UserID: pgUUID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete all backup codes: %w", err)
	}

	return nil
}