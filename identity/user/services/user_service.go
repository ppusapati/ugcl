package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"p9e.in/ugcl/packages/utils"

	"p9e.in/ugcl/identity/user/models"
	"p9e.in/ugcl/identity/user/uow"
)

var _ IUserService = (*UserService)(nil)

// IUserService defines the interface for user-related operations,
// including user registration, role assignments, and retrieval of
// user information. It provides methods to manage users, such as
// creating, updating, deleting, and resetting passwords. Additionally,
// it supports operations for fetching user roles and assigning or
// revoking roles from users.
type IUserService interface {
	// User CRUD Operations
	RegisterUser(ctx context.Context, user *models.User) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id int32) error

	GetUserByID(ctx context.Context, id int32) (*models.User, error)
	GetUserByUUID(ctx context.Context, uuid string) (*models.User, error)
	GetByIdentifier(ctx context.Context, ident *models.UserIdentifier) (*models.User, error)
	ListUsers(ctx context.Context, filter *models.UserSearchCriteria) ([]*models.User, error)

	// Role Management
	AssignRoles(ctx context.Context, userID string, roleIDs []string) error
	GetUserRoles(ctx context.Context, userID string) ([]*models.Role, error)
	AssignUserToRole(ctx context.Context, userID, roleID string) error
	RevokeUserFromRole(ctx context.Context, userID, roleId string) error

	// Password Management
	ChangePassword(ctx context.Context, userID, currentPassword, newPassword, confirmPassword string) error
	InitiatePasswordReset(ctx context.Context, identifier, identifierType string, tenantID *string) (*PasswordResetResult, error)
	ResetPassword(ctx context.Context, resetToken, newPassword, confirmPassword string) error
	ValidatePasswordStrength(ctx context.Context, password string, username, email *string) (*PasswordStrengthResult, error)

	// Email/Phone Verification
	SendVerificationEmail(ctx context.Context, userID string) error
	VerifyEmail(ctx context.Context, userID, verificationCode string) error
	SendVerificationSMS(ctx context.Context, userID string) error
	VerifyPhone(ctx context.Context, userID, verificationCode string) error

	// User Profile Management
	UpdateProfile(ctx context.Context, userID string, profile *UserProfileUpdate) error
	UploadAvatar(ctx context.Context, userID string, avatarData []byte, contentType, filename string) (*AvatarUploadResult, error)
	DeleteAvatar(ctx context.Context, userID string) error

	// User Preferences
	GetUserPreferences(ctx context.Context, userID string) (*models.UserPreferences, error)
	UpdateUserPreferences(ctx context.Context, userID string, preferences *models.UserPreferences) error
}

// DTOs for new operations
type PasswordResetResult struct {
	ResetToken    string
	ExpiresAt     time.Time
	Method        string // "email" or "sms"
	MaskedContact string
}

type PasswordStrengthResult struct {
	IsValid            bool
	StrengthScore      int32
	StrengthLevel      string
	RequirementsMet    []string
	RequirementsFailed []string
	Suggestions        []string
}

type UserProfileUpdate struct {
	FullName    *string
	Phone       *string
	Gender      *string
	Bio         *string
	Location    *string
	Website     *string
	DateOfBirth *time.Time
}

type AvatarUploadResult struct {
	AvatarURL string
	AvatarID  string
}

type UserService struct {
	factory uow.UnitOfWorkFactory
}

func NewUserService(factory uow.UnitOfWorkFactory) IUserService {
	return &UserService{factory: factory}
}

func (s *UserService) RegisterUser(ctx context.Context, user *models.User) (*models.User, error) {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer uow.Rollback(ctx)

	exists, err := uow.UserRepo().Exists(ctx, &models.UserSearchCriteria{
		EmailSearch:    user.Email,
		UsernameSearch: user.Username,
		PhoneSearch:    user.Phone,
	})
	if err != nil {
		return nil, fmt.Errorf("check existing user failed: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("user already exists with given email/username/phone")
	}

	hash, salt, err := utils.EncryptPassword(user.PasswordHash, nil)
	if err != nil {
		return nil, fmt.Errorf("password hashing failed: %w", err)
	}
	user.PasswordHash = hash
	user.Salt = salt
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	res, err := uow.UserRepo().Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("create user failed: %w", err)
	}

	return res, uow.Commit(ctx)
}

func (s *UserService) AssignRoles(ctx context.Context, userID string, roleIDs []string) error {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return err
	}
	defer uow.Rollback(ctx)

	for _, roleID := range roleIDs {
		if err := uow.UserRepo().AssignToUser(ctx, userID, roleID); err != nil {
			return fmt.Errorf("assign role %s failed: %w", roleID, err)
		}
	}

	return uow.Commit(ctx)
}

func (s *UserService) GetUserByID(ctx context.Context, id int32) (*models.User, error) {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return uow.UserRepo().GetByID(ctx, id)
}

func (s *UserService) GetUserByUUID(ctx context.Context, uuid string) (*models.User, error) {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return uow.UserRepo().GetByUUID(ctx, uuid)
}

func (s *UserService) UpdateUser(ctx context.Context, user *models.User) error {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return err
	}
	defer uow.Rollback(ctx)

	user.UpdatedAt = time.Now()
	if err := uow.UserRepo().Update(ctx, user); err != nil {
		return fmt.Errorf("update user failed: %w", err)
	}
	return uow.Commit(ctx)
}

func (s *UserService) DeleteUser(ctx context.Context, id int32) error {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return err
	}
	defer uow.Rollback(ctx)

	if err := uow.UserRepo().Delete(ctx, id); err != nil {
		return fmt.Errorf("delete user failed: %w", err)
	}
	return uow.Commit(ctx)
}

func (s *UserService) ListUsers(ctx context.Context, filter *models.UserSearchCriteria) ([]*models.User, error) {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Println(filter)
	return uow.UserRepo().List(ctx, filter)
}

func (s *UserService) GetUserRoles(ctx context.Context, userID string) ([]*models.Role, error) {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return uow.UserRepo().ListUserRoles(ctx, userID)
}

func (s *UserService) ResetPassword(ctx context.Context, userID string, newHash string, testing string) error {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return err
	}
	defer uow.Rollback(ctx)

	user, err := uow.UserRepo().GetByUUID(ctx, userID)
	if err != nil {
		return err
	}

	if user == nil {
		return errors.New("user not found")
	}

	hash, salt, err := utils.EncryptPassword(newHash, nil)
	if err != nil {
		return fmt.Errorf("hashing password failed: %w", err)
	}
	user.PasswordHash = hash
	user.Salt = salt
	user.UpdatedAt = time.Now()

	if err := uow.UserRepo().Update(ctx, user); err != nil {
		return fmt.Errorf("reset password failed: %w", err)
	}

	return uow.Commit(ctx)
}

func (s *UserService) GetByIdentifier(ctx context.Context, ident *models.UserIdentifier) (*models.User, error) {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer uow.Rollback(ctx)
	switch {
	case ident.Id != 0:
		return uow.UserRepo().GetByID(ctx, ident.Id)
	case ident.Uuid != "":
		return uow.UserRepo().GetByUUID(ctx, ident.Uuid)
	default:
		return nil, errors.New("must specify id or uuid")
	}
}

func (s *UserService) AssignUserToRole(ctx context.Context, userID, roleID string) error {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return err
	}
	defer uow.Rollback(ctx)
	if err := uow.UserRepo().AssignToUser(ctx, userID, roleID); err != nil {
		return fmt.Errorf("failed to assign role to user: %w", err)
	}
	return uow.Commit(ctx)
}

func (s *UserService) RevokeUserFromRole(ctx context.Context, userID, roleID string) error {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return err
	}
	defer uow.Rollback(ctx)
	if err := uow.UserRepo().RevokeFromUser(ctx, userID, roleID); err != nil {
		return fmt.Errorf("failed to revoke role to user: %w", err)
	}
	return uow.Commit(ctx)
}

// Password Management
func (s *UserService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword, confirmPassword string) error {
	// TODO: Implement password change with current password verification
	if newPassword != confirmPassword {
		return errors.New("new password and confirm password do not match")
	}
	// TODO: Verify current password before changing
	// TODO: Apply password strength requirements
	return s.ResetPassword(ctx, userID, newPassword, confirmPassword)
}

func (s *UserService) InitiatePasswordReset(ctx context.Context, identifier, identifierType string, tenantID *string) (*PasswordResetResult, error) {
	// TODO: Implement password reset initiation
	// TODO: Generate reset token
	// TODO: Send email/SMS with reset link
	return &PasswordResetResult{
		ResetToken:    "dummy_reset_token",
		ExpiresAt:     time.Now().Add(24 * time.Hour),
		Method:        "email",
		MaskedContact: "j***@example.com",
	}, nil
}

func (s *UserService) ValidatePasswordStrength(ctx context.Context, password string, username, email *string) (*PasswordStrengthResult, error) {
	// TODO: Implement password strength validation
	// TODO: Check against common passwords, user info, etc.
	score := int32(85) // Mock score
	return &PasswordStrengthResult{
		IsValid:            score >= 60,
		StrengthScore:      score,
		StrengthLevel:      "strong",
		RequirementsMet:    []string{"length", "uppercase", "lowercase", "numbers"},
		RequirementsFailed: []string{},
		Suggestions:        []string{},
	}, nil
}

// Email/Phone Verification
func (s *UserService) SendVerificationEmail(ctx context.Context, userID string) error {
	// TODO: Implement email verification sending
	// TODO: Generate verification code
	// TODO: Send email with verification link
	return nil
}

func (s *UserService) VerifyEmail(ctx context.Context, userID, verificationCode string) error {
	// TODO: Implement email verification
	// TODO: Validate verification code
	// TODO: Update user email_verified status
	return nil
}

func (s *UserService) SendVerificationSMS(ctx context.Context, userID string) error {
	// TODO: Implement SMS verification sending
	// TODO: Generate verification code
	// TODO: Send SMS with verification code
	return nil
}

func (s *UserService) VerifyPhone(ctx context.Context, userID, verificationCode string) error {
	// TODO: Implement phone verification
	// TODO: Validate verification code
	// TODO: Update user phone_verified status
	return nil
}

// User Profile Management
func (s *UserService) UpdateProfile(ctx context.Context, userID string, profile *UserProfileUpdate) error {
	// TODO: Implement profile update
	// TODO: Update user profile fields
	return nil
}

func (s *UserService) UploadAvatar(ctx context.Context, userID string, avatarData []byte, contentType, filename string) (*AvatarUploadResult, error) {
	// TODO: Implement avatar upload
	// TODO: Store avatar file
	// TODO: Update user avatar field
	return &AvatarUploadResult{
		AvatarURL: "https://example.com/avatars/user_avatar.jpg",
		AvatarID:  "avatar_123",
	}, nil
}

func (s *UserService) DeleteAvatar(ctx context.Context, userID string) error {
	// TODO: Implement avatar deletion
	// TODO: Delete avatar file
	// TODO: Clear user avatar field
	return nil
}

// User Preferences
func (s *UserService) GetUserPreferences(ctx context.Context, userID string) (*models.UserPreferences, error) {
	// TODO: Implement get user preferences
	// TODO: Fetch from database or return defaults
	return &models.UserPreferences{
		Language:           "en",
		Timezone:           "UTC",
		DateFormat:         "YYYY-MM-DD",
		TimeFormat:         "24h",
		EmailNotifications: true,
		SmsNotifications:   false,
		PushNotifications:  true,
	}, nil
}

func (s *UserService) UpdateUserPreferences(ctx context.Context, userID string, preferences *models.UserPreferences) error {
	// TODO: Implement update user preferences
	// TODO: Validate preferences
	// TODO: Store in database
	return nil
}
