package services

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"p9e.in/ugcl/identity/auth"
	"p9e.in/ugcl/identity/auth/models"
	"p9e.in/ugcl/identity/auth/services/interfaces"
	"p9e.in/ugcl/identity/auth/uow"
)

type otpService struct {
	authModule   *auth.Module
	emailService interfaces.EmailService
	smsService   interfaces.SMSService
}

func NewOTPService(authModule *auth.Module, emailService interfaces.EmailService, smsService interfaces.SMSService) interfaces.OTPService {
	return &otpService{
		authModule:   authModule,
		emailService: emailService,
		smsService:   smsService,
	}
}

// SendPhoneOTP sends an OTP code via SMS to the specified phone number
func (s *otpService) SendPhoneOTP(ctx context.Context, phone, purpose string) error {
	return s.authModule.GetUnitOfWork().ExecuteInTransaction(ctx, func(ctx context.Context, uow uow.UnitOfWork) error {
		// Generate OTP code
		otpCode := s.generateOTPCode()

		// Create OTP verification token
		// TODO: Implement user lookup by phone to get UserID
		// For now, create a zero UUID placeholder
		token := &models.PhoneVerificationToken{
			UserID:    uuid.Nil,
			Phone:     phone,
			OtpCode:   otpCode,
			ExpiresAt: time.Now().Add(5 * time.Minute), // 5 minutes expiry
			CreatedAt: time.Now(),
		}

		_, err := uow.Tokens().CreatePhoneVerificationToken(ctx, token)
		if err != nil {
			return fmt.Errorf("failed to store phone verification token: %w", err)
		}

		// Send SMS
		err = s.smsService.SendOTP(ctx, phone, otpCode, purpose)
		if err != nil {
			return fmt.Errorf("failed to send SMS OTP: %w", err)
		}

		log.Printf("Phone OTP sent to %s for purpose: %s", phone, purpose)
		return nil
	})
}

// SendEmailOTP sends an OTP code via email to the specified email address
func (s *otpService) SendEmailOTP(ctx context.Context, email, purpose string) error {
	return s.authModule.GetUnitOfWork().ExecuteInTransaction(ctx, func(ctx context.Context, uow uow.UnitOfWork) error {
		// Generate OTP code
		otpCode := s.generateOTPCode()

		// Create email verification token
		// TODO: Implement user lookup by email to get UserID
		// For now, create a zero UUID placeholder
		token := &models.EmailVerificationToken{
			UserID:    uuid.Nil,
			Email:     email,
			TokenHash: otpCode,                          // Store OTP as token hash
			ExpiresAt: time.Now().Add(10 * time.Minute), // 10 minutes expiry
			CreatedAt: time.Now(),
		}

		_, err := uow.Tokens().CreateEmailVerificationToken(ctx, token)
		if err != nil {
			return fmt.Errorf("failed to store email verification token: %w", err)
		}

		// Send email
		err = s.emailService.SendOTP(ctx, email, otpCode, purpose)
		if err != nil {
			return fmt.Errorf("failed to send email OTP: %w", err)
		}

		log.Printf("Email OTP sent to %s for purpose: %s", email, purpose)
		return nil
	})
}

// VerifyPhoneOTP verifies an OTP code for a phone number
func (s *otpService) VerifyPhoneOTP(ctx context.Context, phone, code, purpose string) (bool, error) {
	var isValid bool

	err := s.authModule.GetUnitOfWork().ExecuteInTransaction(ctx, func(ctx context.Context, uow uow.UnitOfWork) error {
		// Get verification token
		token, err := uow.Tokens().GetPhoneVerificationToken(ctx, phone, code)
		if err != nil {
			return err
		}

		if token == nil {
			isValid = false
			return nil
		}

		// Check if token is valid
		if token.VerifiedAt != nil || token.ExpiresAt.Before(time.Now()) {
			isValid = false
			return nil
		}

		// Mark token as used
		err = uow.Tokens().MarkPhoneVerificationTokenUsed(ctx, token.ID)
		if err != nil {
			return fmt.Errorf("failed to mark phone verification token as used: %w", err)
		}

		isValid = true
		return nil
	})

	if err != nil {
		return false, err
	}

	return isValid, nil
}

// VerifyEmailOTP verifies an OTP code for an email address
func (s *otpService) VerifyEmailOTP(ctx context.Context, email, code, purpose string) (bool, error) {
	var isValid bool

	err := s.authModule.GetUnitOfWork().ExecuteInTransaction(ctx, func(ctx context.Context, uow uow.UnitOfWork) error {
		// Get verification token
		token, err := uow.Tokens().GetEmailVerificationTokenByEmailAndCode(ctx, email, code)
		if err != nil {
			return err
		}

		if token == nil {
			isValid = false
			return nil
		}

		// Check if token is valid
		if token.VerifiedAt != nil || token.ExpiresAt.Before(time.Now()) {
			isValid = false
			return nil
		}

		// Mark token as used
		err = uow.Tokens().MarkEmailVerificationTokenUsed(ctx, token.TokenHash)
		if err != nil {
			return fmt.Errorf("failed to mark email verification token as used: %w", err)
		}

		isValid = true
		return nil
	})

	if err != nil {
		return false, err
	}

	return isValid, nil
}

// CleanupExpiredOTPs removes expired OTP tokens from the database
func (s *otpService) CleanupExpiredOTPs(ctx context.Context) error {
	return s.authModule.GetUnitOfWork().ExecuteInTransaction(ctx, func(ctx context.Context, uow uow.UnitOfWork) error {
		// Clean up expired phone verification tokens
		err := uow.Tokens().DeleteExpiredPhoneVerificationTokens(ctx)
		if err != nil {
			log.Printf("Failed to cleanup expired phone verification tokens: %v", err)
		}

		// Clean up expired email verification tokens
		err = uow.Tokens().DeleteExpiredEmailVerificationTokens(ctx)
		if err != nil {
			log.Printf("Failed to cleanup expired email verification tokens: %v", err)
		}

		return nil
	})
}

// Helper methods

func (s *otpService) generateOTPCode() string {
	// Generate a 6-digit OTP code
	code := make([]byte, 3)
	rand.Read(code)
	return fmt.Sprintf("%06d", int(code[0])<<16|int(code[1])<<8|int(code[2]))[:6]
}

// Mock email service for demonstration
type mockEmailService struct{}

func NewMockEmailService() interfaces.EmailService {
	return &mockEmailService{}
}

func (m *mockEmailService) SendOTP(ctx context.Context, email, code, purpose string) error {
	// In a real implementation, this would integrate with an email service like SendGrid, AWS SES, etc.
	log.Printf("MOCK EMAIL: Sending OTP %s to %s for %s", code, email, purpose)
	return nil
}

func (m *mockEmailService) SendPasswordReset(ctx context.Context, email, resetLink string) error {
	log.Printf("MOCK EMAIL: Sending password reset link to %s: %s", email, resetLink)
	return nil
}

func (m *mockEmailService) SendWelcome(ctx context.Context, email, username string) error {
	log.Printf("MOCK EMAIL: Sending welcome email to %s (username: %s)", email, username)
	return nil
}

// Mock SMS service for demonstration
type mockSMSService struct{}

func NewMockSMSService() interfaces.SMSService {
	return &mockSMSService{}
}

func (m *mockSMSService) SendOTP(ctx context.Context, phone, code, purpose string) error {
	// In a real implementation, this would integrate with an SMS service like Twilio, AWS SNS, etc.
	log.Printf("MOCK SMS: Sending OTP %s to %s for %s", code, phone, purpose)
	return nil
}

func (m *mockSMSService) SendVerification(ctx context.Context, phone, verificationLink string) error {
	log.Printf("MOCK SMS: Sending verification link to %s: %s", phone, verificationLink)
	return nil
}
