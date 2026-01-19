package mappers

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	sqlcgen "p9e.in/ugcl/identity/auth/db/generated"
	"p9e.in/ugcl/identity/auth/models"
)

type UserMapper struct{}

func NewUserMapper() *UserMapper {
	return &UserMapper{}
}

// User mapping functions

func (m *UserMapper) FromSQLCUser(sqlcUser sqlcgen.AuthUser) *models.User {
	user := &models.User{
		ID:                  sqlcUser.ID,
		UserID:              sqlcUser.UserID,
		Username:            sqlcUser.Username,
		Email:               sqlcUser.Email,
		PasswordHash:        sqlcUser.PasswordHash,
		PasswordSalt:        sqlcUser.PasswordSalt,
		FailedLoginAttempts: func() int32 { if sqlcUser.FailedLoginAttempts != nil { return *sqlcUser.FailedLoginAttempts }; return 0 }(),
		IsActive:            func() bool { if sqlcUser.IsActive != nil { return *sqlcUser.IsActive }; return false }(),
		EmailVerified:       func() bool { if sqlcUser.EmailVerified != nil { return *sqlcUser.EmailVerified }; return false }(),
		PhoneVerified:       func() bool { if sqlcUser.PhoneVerified != nil { return *sqlcUser.PhoneVerified }; return false }(),
		TwoFactorEnabled:    func() bool { if sqlcUser.TwoFactorEnabled != nil { return *sqlcUser.TwoFactorEnabled }; return false }(),
		BackupCodes:         sqlcUser.BackupCodes,
		CreatedAt:           sqlcUser.CreatedAt,
		UpdatedAt:           sqlcUser.UpdatedAt,
	}

	if sqlcUser.Phone != nil {
		user.Phone = sqlcUser.Phone
	}
	if sqlcUser.PasswordChangedAt.Valid {
		user.PasswordChangedAt = &sqlcUser.PasswordChangedAt.Time
	}
	if sqlcUser.PasswordExpiresAt.Valid {
		user.PasswordExpiresAt = &sqlcUser.PasswordExpiresAt.Time
	}
	if sqlcUser.AccountLockedUntil.Valid {
		user.AccountLockedUntil = &sqlcUser.AccountLockedUntil.Time
	}
	if sqlcUser.TwoFactorSecret != nil {
		user.TwoFactorSecret = sqlcUser.TwoFactorSecret
	}
	if sqlcUser.LastLoginAt.Valid {
		user.LastLoginAt = &sqlcUser.LastLoginAt.Time
	}
	if sqlcUser.LastPasswordChange.Valid {
		user.LastPasswordChange = &sqlcUser.LastPasswordChange.Time
	}

	return user
}

func (m *UserMapper) ToCreateUserParams(user *models.User) sqlcgen.CreateAuthUserParams {
	params := sqlcgen.CreateAuthUserParams{
		UserID:       user.UserID,
		Username:     user.Username,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		PasswordSalt: user.PasswordSalt,
		// IsActive:     &user.IsActive, // Field not available in CreateAuthUserParams
		BackupCodes:  user.BackupCodes,
	}

	if user.Phone != nil {
		params.Phone = user.Phone
	}
	if user.TwoFactorSecret != nil {
		params.TwoFactorSecret = user.TwoFactorSecret
	}

	return params
}

func (m *UserMapper) ToUpdateUserParams(user *models.User) sqlcgen.UpdateAuthUserParams {
	params := sqlcgen.UpdateAuthUserParams{
		ID:               user.ID,
		Username:         user.Username,
		Email:            user.Email,
		IsActive:         &user.IsActive,
		EmailVerified:    &user.EmailVerified,
		PhoneVerified:    &user.PhoneVerified,
		TwoFactorEnabled: &user.TwoFactorEnabled,
	}

	if user.Phone != nil {
		params.Phone = user.Phone
	}

	return params
}

// UserTenant mapping functions

func (m *UserMapper) FromSQLCUserTenant(sqlcUserTenant sqlcgen.AuthUserTenant) *models.UserTenant {
	userTenant := &models.UserTenant{
		ID:        sqlcUserTenant.ID,
		UserID:    uuid.UUID(sqlcUserTenant.UserID.Bytes),
		TenantID:  sqlcUserTenant.TenantID,
		IsDefault: func() bool { if sqlcUserTenant.IsDefault != nil { return *sqlcUserTenant.IsDefault }; return false }(),
		IsActive:  func() bool { if sqlcUserTenant.IsActive != nil { return *sqlcUserTenant.IsActive }; return false }(),
		Roles:     sqlcUserTenant.Roles,
		CreatedAt: sqlcUserTenant.CreatedAt,
		UpdatedAt: sqlcUserTenant.UpdatedAt,
	}

	if sqlcUserTenant.JoinedAt.Valid {
		userTenant.JoinedAt = &sqlcUserTenant.JoinedAt.Time
	}

	return userTenant
}

func (m *UserMapper) ToCreateUserTenantParams(userTenant *models.UserTenant) sqlcgen.CreateAuthUserTenantParams {
	params := sqlcgen.CreateAuthUserTenantParams{
		UserID:    pgtype.UUID{Bytes: userTenant.UserID, Valid: true},
		TenantID:  userTenant.TenantID,
		IsDefault: &userTenant.IsDefault,
		IsActive:  &userTenant.IsActive,
		Roles:     userTenant.Roles,
	}

	// if userTenant.JoinedAt != nil {
	// 	params.JoinedAt = pgtype.Timestamp{Time: *userTenant.JoinedAt, Valid: true}
	// } // Field not available in CreateAuthUserTenantParams

	return params
}

// TODO: Implement when UpdateAuthUserTenantParams is added to SQLC queries
// func (m *UserMapper) ToUpdateUserTenantParams(userTenant *models.UserTenant) sqlcgen.UpdateAuthUserTenantParams {
// 	params := sqlcgen.UpdateAuthUserTenantParams{
// 		UserID:    pgtype.UUID{Bytes: userTenant.UserID, Valid: true},
// 		TenantID:  userTenant.TenantID,
// 		IsDefault: &userTenant.IsDefault,
// 		IsActive:  &userTenant.IsActive,
// 		Roles:     userTenant.Roles,
// 	}

// 	return params
// }