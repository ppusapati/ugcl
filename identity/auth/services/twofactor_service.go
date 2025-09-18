package services

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"

	"p9e.in/ugcl/identity/auth"
	"p9e.in/ugcl/identity/auth/models"
	"p9e.in/ugcl/identity/auth/services/interfaces"
	"p9e.in/ugcl/identity/auth/uow"
)

type twoFactorService struct {
	authModule *auth.Module
}

func NewTwoFactorService(authModule *auth.Module) interfaces.TwoFactorService {
	return &twoFactorService{
		authModule: authModule,
	}
}

func (s *twoFactorService) EnableTwoFactor(ctx context.Context, userID string, method string) (*interfaces.EnableTwoFactorResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	var response *interfaces.EnableTwoFactorResponse

	err = s.authModule.GetUnitOfWork().ExecuteInTransaction(ctx, func(ctx context.Context, uow uow.UnitOfWork) error {
		// Get user
		user, err := uow.Users().GetByID(ctx, userUUID)
		if err != nil {
			return fmt.Errorf("user not found: %w", err)
		}

		if user == nil {
			return fmt.Errorf("user not found")
		}

		// Check if 2FA is already enabled
		if !user.TwoFactorEnabled {
			return fmt.Errorf("two-factor authentication is already enabled")
		}

		switch method {
		case "totp":
			response, err = s.enableTOTP(ctx, uow, user)
		case "sms":
			response, err = s.enableSMS(ctx, uow, user)
		case "email":
			response, err = s.enableEmail(ctx, uow, user)
		default:
			return fmt.Errorf("unsupported two-factor method: %s", method)
		}

		if err != nil {
			return err
		}

		// Log security event
		s.logSecurityEvent(ctx, uow, userUUID, "two_factor_enabled", map[string]interface{}{
			"method": method,
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *twoFactorService) DisableTwoFactor(ctx context.Context, userID, verificationCode string) error {
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

		// Check if 2FA is enabled
		if !user.TwoFactorEnabled {
			return fmt.Errorf("two-factor authentication is not enabled")
		}

		// Verify the provided code
		valid, err := s.verifyTOTPCode(user.TwoFactorSecret, verificationCode)
		if err != nil {
			return fmt.Errorf("failed to verify code: %w", err)
		}

		if !valid {
			// Try backup code as alternative
			backupCode, err := uow.TwoFactor().UseBackupCode(ctx, userUUID, s.hashBackupCode(verificationCode))
			if err != nil || backupCode == nil {
				s.logSecurityEvent(ctx, uow, userUUID, "two_factor_disable_failed", map[string]interface{}{
					"reason": "invalid_verification_code",
				})
				return fmt.Errorf("invalid verification code")
			}
		}

		// Disable 2FA
		err = uow.TwoFactor().DisableTwoFactor(ctx, userUUID)
		if err != nil {
			return fmt.Errorf("failed to disable two-factor: %w", err)
		}

		// Delete all backup codes
		err = uow.TwoFactor().DeleteAllBackupCodes(ctx, userUUID)
		if err != nil {
			log.Printf("Warning: failed to delete backup codes: %v", err)
		}

		// Log security event
		s.logSecurityEvent(ctx, uow, userUUID, "two_factor_disabled", map[string]interface{}{
			"verification_method": "totp_or_backup",
		})

		return nil
	})
}

func (s *twoFactorService) VerifyTwoFactor(ctx context.Context, userID, code, method string) error {
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

		// Check if 2FA is enabled
		if !user.TwoFactorEnabled {
			return fmt.Errorf("two-factor authentication is not enabled")
		}

		var valid bool
		switch method {
		case "totp":
			valid, err = s.verifyTOTPCode(user.TwoFactorSecret, code)
		case "backup_code":
			backupCode, verifyErr := uow.TwoFactor().UseBackupCode(ctx, userUUID, s.hashBackupCode(code))
			valid = verifyErr == nil && backupCode != nil
			err = verifyErr
		case "sms", "email":
			// These would typically involve checking tokens stored in the database
			return fmt.Errorf("verification method %s not yet implemented", method)
		default:
			return fmt.Errorf("unsupported verification method: %s", method)
		}

		if err != nil {
			return fmt.Errorf("failed to verify code: %w", err)
		}

		if !valid {
			s.logSecurityEvent(ctx, uow, userUUID, "two_factor_verification_failed", map[string]interface{}{
				"method": method,
			})
			return fmt.Errorf("invalid verification code")
		}

		// Log successful verification
		s.logSecurityEvent(ctx, uow, userUUID, "two_factor_verified", map[string]interface{}{
			"method": method,
		})

		return nil
	})
}

func (s *twoFactorService) GenerateBackupCodes(ctx context.Context, userID string) ([]string, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	var codes []string

	err = s.authModule.GetUnitOfWork().ExecuteInTransaction(ctx, func(ctx context.Context, uow uow.UnitOfWork) error {
		// Generate new backup codes
		codes = s.generateBackupCodes(10) // Generate 10 backup codes
		hashedCodes := make([]string, len(codes))

		for i, code := range codes {
			hashedCodes[i] = s.hashBackupCode(code)
		}

		// Delete existing backup codes
		err := uow.TwoFactor().DeleteAllBackupCodes(ctx, userUUID)
		if err != nil {
			log.Printf("Warning: failed to delete existing backup codes: %v", err)
		}

		// Store new backup codes
		err = uow.TwoFactor().CreateBackupCodes(ctx, userUUID, hashedCodes)
		if err != nil {
			return fmt.Errorf("failed to store backup codes: %w", err)
		}

		// Log security event
		s.logSecurityEvent(ctx, uow, userUUID, "backup_codes_generated", map[string]interface{}{
			"count": len(codes),
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	return codes, nil
}

func (s *twoFactorService) UseBackupCode(ctx context.Context, userID, code string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	return s.authModule.GetUnitOfWork().ExecuteInTransaction(ctx, func(ctx context.Context, uow uow.UnitOfWork) error {
		// Try to use the backup code
		backupCode, err := uow.TwoFactor().UseBackupCode(ctx, userUUID, s.hashBackupCode(code))
		if err != nil {
			return fmt.Errorf("failed to use backup code: %w", err)
		}

		if backupCode == nil {
			s.logSecurityEvent(ctx, uow, userUUID, "backup_code_use_failed", map[string]interface{}{
				"reason": "invalid_code",
			})
			return fmt.Errorf("invalid backup code")
		}

		// Log successful backup code use
		s.logSecurityEvent(ctx, uow, userUUID, "backup_code_used", map[string]interface{}{
			"code_id": backupCode.ID.String(),
		})

		// Check remaining backup codes
		remainingCount, err := uow.TwoFactor().GetUnusedBackupCodesCount(ctx, userUUID)
		if err != nil {
			log.Printf("Warning: failed to get remaining backup codes count: %v", err)
		} else if remainingCount <= 2 {
			// Log warning when backup codes are running low
			s.logSecurityEvent(ctx, uow, userUUID, "backup_codes_low", map[string]interface{}{
				"remaining_count": remainingCount,
			})
		}

		return nil
	})
}

// Helper methods

func (s *twoFactorService) enableTOTP(ctx context.Context, uow uow.UnitOfWork, user *models.User) (*interfaces.EnableTwoFactorResponse, error) {
	// Generate TOTP secret
	secret := s.generateTOTPSecret()

	// Generate QR code (base64 encoded)
	qrCode := s.generateQRCode(user.Email, secret)

	// Generate backup codes
	backupCodes := s.generateBackupCodes(10)
	hashedBackupCodes := make([]string, len(backupCodes))
	for i, code := range backupCodes {
		hashedBackupCodes[i] = s.hashBackupCode(code)
	}

	// Enable 2FA for user
	err := uow.TwoFactor().EnableTwoFactor(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to enable two-factor: %w", err)
	}

	// Store backup codes
	err = uow.TwoFactor().CreateBackupCodes(ctx, user.ID, hashedBackupCodes)
	if err != nil {
		log.Printf("Warning: failed to store backup codes: %v", err)
	}

	return &interfaces.EnableTwoFactorResponse{
		Secret:      secret,
		QRCode:      qrCode,
		BackupCodes: backupCodes,
	}, nil
}

func (s *twoFactorService) enableSMS(ctx context.Context, uow uow.UnitOfWork, user *models.User) (*interfaces.EnableTwoFactorResponse, error) {
	// Check if user has a verified phone number
	if user.Phone == nil || !user.PhoneVerified {
		return nil, fmt.Errorf("user must have a verified phone number to enable SMS 2FA")
	}

	// TODO: Implement SMS-based 2FA setup
	// This would typically involve:
	// 1. Sending a verification SMS
	// 2. Storing SMS 2FA configuration
	// 3. Returning appropriate response

	return nil, fmt.Errorf("SMS two-factor authentication not yet implemented")
}

func (s *twoFactorService) enableEmail(ctx context.Context, uow uow.UnitOfWork, user *models.User) (*interfaces.EnableTwoFactorResponse, error) {
	// Check if user has a verified email
	if !user.EmailVerified {
		return nil, fmt.Errorf("user must have a verified email to enable email 2FA")
	}

	// TODO: Implement email-based 2FA setup
	// This would typically involve:
	// 1. Sending a verification email
	// 2. Storing email 2FA configuration
	// 3. Returning appropriate response

	return nil, fmt.Errorf("email two-factor authentication not yet implemented")
}

func (s *twoFactorService) generateTOTPSecret() string {
	// Generate a 32-byte secret for TOTP
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return base32.StdEncoding.EncodeToString(bytes)
}

func (s *twoFactorService) generateQRCode(email, secret string) string {
	// Generate TOTP key using the otp library
	key, err := otp.NewKeyFromURL(fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s",
		"UGCL Identity", email, secret, "UGCL Identity"))
	if err != nil {
		log.Printf("Failed to generate OTP key: %v", err)
		// Fallback to manual URL generation
		return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s",
			"UGCL Identity", email, secret, "UGCL Identity")
	}

	// Return the URL that can be used to generate QR code
	return key.URL()
}

func (s *twoFactorService) verifyTOTPCode(secret *string, code string) (bool, error) {
	if secret == nil {
		return false, fmt.Errorf("no TOTP secret configured")
	}

	// Basic validation
	if len(code) != 6 {
		return false, nil
	}

	// Remove any spaces from the code
	code = strings.ReplaceAll(code, " ", "")

	// Verify TOTP code using the otp library
	return totp.Validate(code, *secret), nil
}

func (s *twoFactorService) generateBackupCodes(count int) []string {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		codes[i] = s.generateBackupCode()
	}
	return codes
}

func (s *twoFactorService) generateBackupCode() string {
	// Generate 8-character backup code
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return fmt.Sprintf("%08x", bytes)
}

func (s *twoFactorService) hashBackupCode(code string) string {
	hashedCode, _ := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	return string(hashedCode)
}

func (s *twoFactorService) logSecurityEvent(ctx context.Context, uow uow.UnitOfWork, userID uuid.UUID, eventType string, eventData map[string]interface{}) {
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
