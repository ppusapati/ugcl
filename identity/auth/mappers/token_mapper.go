package mappers

import (
	"net/netip"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	sqlcgen "p9e.in/ugcl/identity/auth/db/generated"
	"p9e.in/ugcl/identity/auth/models"
)

type TokenMapper struct{}

func NewTokenMapper() *TokenMapper {
	return &TokenMapper{}
}

// PasswordResetToken mapping functions

func (m *TokenMapper) FromSQLCPasswordResetToken(sqlcToken sqlcgen.AuthPasswordResetToken) *models.PasswordResetToken {
	token := &models.PasswordResetToken{
		ID:        sqlcToken.ID,
		UserID:    uuid.UUID(sqlcToken.UserID.Bytes),
		TokenHash: sqlcToken.TokenHash,
		ExpiresAt: sqlcToken.ExpiresAt.Time,
		CreatedAt: sqlcToken.CreatedAt,
	}

	if sqlcToken.UsedAt.Valid {
		token.UsedAt = &sqlcToken.UsedAt.Time
	}
	if sqlcToken.IpAddress != nil {
		ipStr := sqlcToken.IpAddress.String()
		token.IPAddress = &ipStr
	}
	if sqlcToken.UserAgent != nil {
		token.UserAgent = sqlcToken.UserAgent
	}

	return token
}

func (m *TokenMapper) ToCreatePasswordResetTokenParams(token *models.PasswordResetToken) sqlcgen.CreatePasswordResetTokenParams {
	params := sqlcgen.CreatePasswordResetTokenParams{
		UserID:    pgtype.UUID{Bytes: token.UserID, Valid: true},
		TokenHash: token.TokenHash,
		ExpiresAt: pgtype.Timestamp{Time: token.ExpiresAt, Valid: true},
	}

	if token.IPAddress != nil {
		if addr, err := netip.ParseAddr(*token.IPAddress); err == nil {
			params.IpAddress = &addr
		}
	}
	if token.UserAgent != nil {
		params.UserAgent = token.UserAgent
	}

	return params
}

// EmailVerificationToken mapping functions

func (m *TokenMapper) FromSQLCEmailVerificationToken(sqlcToken sqlcgen.AuthEmailVerificationToken) *models.EmailVerificationToken {
	token := &models.EmailVerificationToken{
		ID:        sqlcToken.ID,
		UserID:    uuid.UUID(sqlcToken.UserID.Bytes),
		Email:     sqlcToken.Email,
		TokenHash: sqlcToken.TokenHash,
		ExpiresAt: sqlcToken.ExpiresAt.Time,
		CreatedAt: sqlcToken.CreatedAt,
	}

	if sqlcToken.VerifiedAt.Valid {
		token.VerifiedAt = &sqlcToken.VerifiedAt.Time
	}

	return token
}

func (m *TokenMapper) ToCreateEmailVerificationTokenParams(token *models.EmailVerificationToken) sqlcgen.CreateEmailVerificationTokenParams {
	return sqlcgen.CreateEmailVerificationTokenParams{
		UserID:    pgtype.UUID{Bytes: token.UserID, Valid: true},
		Email:     token.Email,
		TokenHash: token.TokenHash,
		ExpiresAt: pgtype.Timestamp{Time: token.ExpiresAt, Valid: true},
	}
}

// PhoneVerificationToken mapping functions

func (m *TokenMapper) FromSQLCPhoneVerificationToken(sqlcToken sqlcgen.AuthPhoneVerificationToken) *models.PhoneVerificationToken {
	token := &models.PhoneVerificationToken{
		ID:        sqlcToken.ID,
		UserID:    uuid.UUID(sqlcToken.UserID.Bytes),
		Phone:     sqlcToken.Phone,
		OtpCode:   sqlcToken.OtpCode,
		ExpiresAt: sqlcToken.ExpiresAt.Time,
		Attempts:  func() int32 { if sqlcToken.Attempts != nil { return *sqlcToken.Attempts }; return 0 }(),
		CreatedAt: sqlcToken.CreatedAt,
	}

	if sqlcToken.VerifiedAt.Valid {
		token.VerifiedAt = &sqlcToken.VerifiedAt.Time
	}

	return token
}

func (m *TokenMapper) ToCreatePhoneVerificationTokenParams(token *models.PhoneVerificationToken) sqlcgen.CreatePhoneVerificationTokenParams {
	return sqlcgen.CreatePhoneVerificationTokenParams{
		UserID:    pgtype.UUID{Bytes: token.UserID, Valid: true},
		Phone:     token.Phone,
		OtpCode:   token.OtpCode,
		ExpiresAt: pgtype.Timestamp{Time: token.ExpiresAt, Valid: true},
		// Attempts:  &token.Attempts, // Field not available in CreatePhoneVerificationTokenParams
	}
}

// RevokedToken mapping functions

func (m *TokenMapper) FromSQLCRevokedToken(sqlcToken sqlcgen.AuthRevokedToken) *models.RevokedToken {
	token := &models.RevokedToken{
		ID:        sqlcToken.ID,
		TokenJti:  sqlcToken.TokenJti,
		TokenType: sqlcToken.TokenType,
		UserID:    uuid.UUID(sqlcToken.UserID.Bytes),
		RevokedAt: sqlcToken.RevokedAt.Time,
		ExpiresAt: sqlcToken.ExpiresAt.Time,
	}

	if sqlcToken.Reason != nil {
		token.Reason = sqlcToken.Reason
	}

	return token
}

// TODO: Implement when CreateRevokedTokenParams is added to SQLC queries
// func (m *TokenMapper) ToCreateRevokedTokenParams(token *models.RevokedToken) sqlcgen.CreateRevokedTokenParams {
// 	params := sqlcgen.CreateRevokedTokenParams{
// 		TokenJti:  token.TokenJti,
// 		TokenType: token.TokenType,
// 		UserID:    pgtype.UUID{Bytes: token.UserID, Valid: true},
// 		RevokedAt: pgtype.Timestamp{Time: token.RevokedAt, Valid: true},
// 		ExpiresAt: pgtype.Timestamp{Time: token.ExpiresAt, Valid: true},
// 	}

// 	if token.Reason != nil {
// 		params.Reason = token.Reason
// 	}

// 	return params
// }

// TwoFactorBackupCode mapping functions

func (m *TokenMapper) FromSQLCTwoFactorBackupCode(sqlcCode sqlcgen.AuthTwoFactorBackupCode) *models.TwoFactorBackupCode {
	code := &models.TwoFactorBackupCode{
		ID:        sqlcCode.ID,
		UserID:    uuid.UUID(sqlcCode.UserID.Bytes),
		CodeHash:  sqlcCode.CodeHash,
		CreatedAt: sqlcCode.CreatedAt,
	}

	if sqlcCode.UsedAt.Valid {
		code.UsedAt = &sqlcCode.UsedAt.Time
	}

	return code
}

func (m *TokenMapper) ToCreateTwoFactorBackupCodeParams(code *models.TwoFactorBackupCode) sqlcgen.CreateTwoFactorBackupCodeParams {
	return sqlcgen.CreateTwoFactorBackupCodeParams{
		UserID:   pgtype.UUID{Bytes: code.UserID, Valid: true},
		CodeHash: code.CodeHash,
	}
}