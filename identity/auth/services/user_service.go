package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"p9e.in/ugcl/identity/auth"
	"p9e.in/ugcl/identity/auth/models"
	"p9e.in/ugcl/identity/auth/services/interfaces"
	"p9e.in/ugcl/identity/auth/uow"
)

type userService struct {
	authModule *auth.Module
}

func NewUserService(authModule *auth.Module) interfaces.UserService {
	return &userService{
		authModule: authModule,
	}
}

func (s *userService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	return s.authModule.GetUnitOfWork().ExecuteInTransaction(ctx, func(ctx context.Context, uow uow.UnitOfWork) error {
		// Get user
		user, err := uow.Users().GetByID(ctx, userUUID)
		if err != nil {
			return fmt.Errorf("user not found: %w", err)
		}

		if user == nil {
			return fmt.Errorf("user not found")
		}

		// Verify current password
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
			// Log failed password change attempt
			s.logSecurityEvent(ctx, uow, userUUID, "password_change_failed", map[string]interface{}{
				"reason": "invalid_current_password",
			})
			return fmt.Errorf("current password is incorrect")
		}

		// Validate new password strength
		if err := s.validatePasswordStrength(newPassword); err != nil {
			return err
		}

		// Hash new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		// Update password
		user.PasswordHash = string(hashedPassword)
		user.PasswordChangedAt = func() *time.Time { t := time.Now(); return &t }()

		_, err = uow.Users().Update(ctx, user)
		if err != nil {
			return fmt.Errorf("failed to update password: %w", err)
		}

		// Log successful password change
		s.logSecurityEvent(ctx, uow, userUUID, "password_changed", map[string]interface{}{
			"user_id": userID,
		})

		return nil
	})
}

func (s *userService) ForgotPassword(ctx context.Context, identifier, identifierType, tenantID string) error {
	return s.authModule.GetUnitOfWork().ExecuteInTransaction(ctx, func(ctx context.Context, uow uow.UnitOfWork) error {
		// Get user based on identifier type
		var user *models.User
		var err error

		switch identifierType {
		case "email":
			user, err = uow.Users().GetByEmail(ctx, identifier)
		case "username":
			user, err = uow.Users().GetByUsername(ctx, identifier)
		case "phone":
			user, err = uow.Users().GetByPhone(ctx, identifier)
		default:
			return fmt.Errorf("unsupported identifier type: %s", identifierType)
		}

		if err != nil || user == nil {
			// Don't reveal if user exists or not for security reasons
			log.Printf("Password reset requested for non-existent user: %s", identifier)
			return nil // Return success to prevent user enumeration
		}

		// Check if user is active
		if !user.IsActive {
			log.Printf("Password reset requested for inactive user: %s", identifier)
			return nil // Return success to prevent user enumeration
		}

		// Generate reset token
		resetToken := s.generateResetToken()
		hashedToken, err := bcrypt.GenerateFromPassword([]byte(resetToken), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash reset token: %w", err)
		}

		// Create password reset token
		passwordResetToken := &models.PasswordResetToken{
			UserID:    user.ID,
			TokenHash: string(hashedToken),
			ExpiresAt: time.Now().Add(1 * time.Hour), // Token valid for 1 hour
			// IsUsed:    false,
		}

		_, err = uow.Tokens().CreatePasswordResetToken(ctx, passwordResetToken)
		if err != nil {
			return fmt.Errorf("failed to create reset token: %w", err)
		}

		// TODO: Send password reset email/SMS
		// This would typically integrate with an email/SMS service
		log.Printf("Password reset token generated for user %s: %s", user.UserID, resetToken)

		// Log security event
		s.logSecurityEvent(ctx, uow, user.ID, "password_reset_requested", map[string]interface{}{
			"identifier_type": identifierType,
			"identifier":      identifier,
		})

		return nil
	})
}

func (s *userService) ResetPassword(ctx context.Context, resetToken, newPassword string) error {
	return s.authModule.GetUnitOfWork().ExecuteInTransaction(ctx, func(ctx context.Context, uow uow.UnitOfWork) error {
		// Validate new password strength
		if err := s.validatePasswordStrength(newPassword); err != nil {
			return err
		}

		// Hash the provided token to compare with stored hash
		// Note: In a real implementation, you'd want to store the token hash differently
		// This is a simplified approach
		tokenHash := s.hashResetToken(resetToken)

		// Get and validate reset token
		storedToken, err := uow.Tokens().GetPasswordResetToken(ctx, tokenHash)
		if err != nil {
			return fmt.Errorf("invalid reset token")
		}

		if storedToken == nil {
			return fmt.Errorf("invalid reset token")
		}

		// Check if token is expired
		if storedToken.ExpiresAt.Before(time.Now()) {
			return fmt.Errorf("reset token has expired")
		}

		// // Check if token is already used
		// if storedToken.IsUsed {
		// 	return fmt.Errorf("reset token has already been used")
		// }

		// Get user
		user, err := uow.Users().GetByID(ctx, storedToken.UserID)
		if err != nil {
			return fmt.Errorf("user not found: %w", err)
		}

		if user == nil {
			return fmt.Errorf("user not found")
		}

		// Hash new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		// Update password
		user.PasswordHash = string(hashedPassword)
		user.PasswordChangedAt = func() *time.Time { t := time.Now(); return &t }()
		user.FailedLoginAttempts = 0  // Reset failed attempts
		user.AccountLockedUntil = nil // Unlock account if locked

		_, err = uow.Users().Update(ctx, user)
		if err != nil {
			return fmt.Errorf("failed to update password: %w", err)
		}

		// Mark token as used
		err = uow.Tokens().MarkPasswordResetTokenUsed(ctx, tokenHash)
		if err != nil {
			log.Printf("Warning: failed to mark reset token as used: %v", err)
		}

		// Log successful password reset
		s.logSecurityEvent(ctx, uow, user.ID, "password_reset_completed", map[string]interface{}{
			"user_id": user.UserID,
		})

		return nil
	})
}

func (s *userService) GetUserProfile(ctx context.Context, userID string) (*models.User, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	user, err := s.authModule.Users().GetByID(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

func (s *userService) UpdateUserProfile(ctx context.Context, userID string, updates *interfaces.UserProfileUpdate) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	return s.authModule.GetUnitOfWork().ExecuteInTransaction(ctx, func(ctx context.Context, uow uow.UnitOfWork) error {
		// Get current user
		user, err := uow.Users().GetByID(ctx, userUUID)
		if err != nil {
			return fmt.Errorf("user not found: %w", err)
		}

		if user == nil {
			return fmt.Errorf("user not found")
		}

		// Apply updates
		hasChanges := false
		if updates.FullName != nil && *updates.FullName != user.Username {
			user.Username = *updates.FullName
			hasChanges = true
		}

		if updates.Email != nil && *updates.Email != user.Email {
			// Check if email is already in use
			existingUser, err := uow.Users().GetByEmail(ctx, *updates.Email)
			if err == nil && existingUser != nil && existingUser.ID != user.ID {
				return fmt.Errorf("email already in use")
			}
			user.Email = *updates.Email
			user.EmailVerified = false // Reset email verification
			hasChanges = true
		}

		if updates.Phone != nil && (user.Phone == nil || *updates.Phone != *user.Phone) {
			// Check if phone is already in use
			if *updates.Phone != "" {
				existingUser, err := uow.Users().GetByPhone(ctx, *updates.Phone)
				if err == nil && existingUser != nil && existingUser.ID != user.ID {
					return fmt.Errorf("phone number already in use")
				}
			}
			user.Phone = updates.Phone
			user.PhoneVerified = false // Reset phone verification
			hasChanges = true
		}

		if updates.Username != nil && *updates.Username != user.Username {
			// Check if username is already in use
			existingUser, err := uow.Users().GetByUsername(ctx, *updates.Username)
			if err == nil && existingUser != nil && existingUser.ID != user.ID {
				return fmt.Errorf("username already in use")
			}
			user.Username = *updates.Username
			hasChanges = true
		}

		if updates.IsActive != nil && *updates.IsActive != user.IsActive {
			user.IsActive = *updates.IsActive
			hasChanges = true
		}

		// if updates.Preferences != nil {
		// 	user.Preferences = updates.Preferences
		// 	hasChanges = true
		// }

		if !hasChanges {
			return nil // No changes to save
		}

		// Update user
		_, err = uow.Users().Update(ctx, user)
		if err != nil {
			return fmt.Errorf("failed to update user profile: %w", err)
		}

		// Log profile update
		s.logSecurityEvent(ctx, uow, userUUID, "profile_updated", map[string]interface{}{
			"user_id": userID,
			"changes": s.getUpdateSummary(updates),
		})

		return nil
	})
}

func (s *userService) DeactivateUser(ctx context.Context, userID string) error {
	return s.updateUserStatus(ctx, userID, false, "user_deactivated")
}

func (s *userService) ActivateUser(ctx context.Context, userID string) error {
	return s.updateUserStatus(ctx, userID, true, "user_activated")
}

// Helper methods

func (s *userService) validatePasswordStrength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// Add more password complexity rules as needed
	// - Contains uppercase letter
	// - Contains lowercase letter
	// - Contains digit
	// - Contains special character

	return nil
}

func (s *userService) generateResetToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (s *userService) hashResetToken(token string) string {
	// In a real implementation, you'd use a proper cryptographic hash
	// This is simplified for demonstration
	hashedToken, _ := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	return string(hashedToken)
}

func (s *userService) updateUserStatus(ctx context.Context, userID string, isActive bool, eventType string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	return s.authModule.GetUnitOfWork().ExecuteInTransaction(ctx, func(ctx context.Context, uow uow.UnitOfWork) error {
		// Get user
		user, err := uow.Users().GetByID(ctx, userUUID)
		if err != nil {
			return fmt.Errorf("user not found: %w", err)
		}

		if user == nil {
			return fmt.Errorf("user not found")
		}

		// Update status
		user.IsActive = isActive
		_, err = uow.Users().Update(ctx, user)
		if err != nil {
			return fmt.Errorf("failed to update user status: %w", err)
		}

		// If deactivating, revoke all sessions
		if !isActive {
			err = uow.Sessions().DeactivateAll(ctx, userUUID)
			if err != nil {
				log.Printf("Warning: failed to deactivate user sessions: %v", err)
			}
		}

		// Log status change
		s.logSecurityEvent(ctx, uow, userUUID, eventType, map[string]interface{}{
			"user_id":   userUUID,
			"is_active": isActive,
		})

		return nil
	})
}

func (s *userService) getUpdateSummary(updates *interfaces.UserProfileUpdate) []string {
	var changes []string
	if updates.FullName != nil {
		changes = append(changes, "full_name")
	}
	if updates.Email != nil {
		changes = append(changes, "email")
	}
	if updates.Phone != nil {
		changes = append(changes, "phone")
	}
	if updates.Username != nil {
		changes = append(changes, "username")
	}
	if updates.IsActive != nil {
		changes = append(changes, "is_active")
	}
	if updates.Preferences != nil {
		changes = append(changes, "preferences")
	}
	return changes
}

func (s *userService) logSecurityEvent(ctx context.Context, uow uow.UnitOfWork, userID uuid.UUID, eventType string, eventData map[string]interface{}) {
	event := &models.SecurityEvent{
		UserID:    userID,
		EventType: eventType,
		EventData: eventData,
		CreatedAt: time.Now(),
	}

	_, err := uow.Audit().CreateSecurityEvent(ctx, event)
	if err != nil {
		log.Printf("Failed to log security event: %v", err)
	}
}
