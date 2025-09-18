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

	"p9e.in/ugcl/identity/auth/models"
	"p9e.in/ugcl/identity/auth/services/interfaces"
	"p9e.in/ugcl/identity/auth/uow"
)

type authService struct {
	uowFactory   uow.UnitOfWorkFactory
	jwtService   interfaces.JWTService
	eventService interfaces.EventService
}

func NewAuthService(uowFactory uow.UnitOfWorkFactory, jwtService interfaces.JWTService, eventService interfaces.EventService) interfaces.AuthService {
	return &authService{
		uowFactory:   uowFactory,
		jwtService:   jwtService,
		eventService: eventService,
	}
}

func (s *authService) Login(ctx context.Context, req *interfaces.LoginRequest) (*interfaces.LoginResponse, error) {
	var loginResponse *interfaces.LoginResponse

	err := s.uowFactory.Create().ExecuteInTransaction(ctx, func(ctx context.Context, uow uow.UnitOfWork) error {
		// Get user based on identifier type
		var user *models.User
		var err error

		switch req.IdentifierType {
		case "email":
			user, err = uow.Users().GetByEmail(ctx, req.Identifier)
		case "username":
			user, err = uow.Users().GetByUsername(ctx, req.Identifier)
		case "phone":
			// Phone lookup not implemented yet
			return fmt.Errorf("phone authentication not yet supported")
		default:
			return fmt.Errorf("unsupported identifier type: %s", req.IdentifierType)
		}

		if err != nil {
			// Log failed login attempt
			s.logFailedLoginAttempt(ctx, uow, req.Identifier, "user_not_found", nil, req.IPAddress)

			// Publish failed login event
			if s.eventService != nil {
				if err := s.eventService.PublishFailedLoginEvent(ctx, req.Identifier, "user_not_found", req.IPAddress); err != nil {
					log.Printf("Warning: failed to publish failed login event: %v", err)
				}
			}

			return fmt.Errorf("authentication failed")
		}

		if user == nil {
			s.logFailedLoginAttempt(ctx, uow, req.Identifier, "user_not_found", nil, req.IPAddress)
			return fmt.Errorf("authentication failed")
		}

		// Validate user account status
		if err := s.validateUserAccount(ctx, uow, user, req); err != nil {
			return err
		}

		// Verify password
		if err := s.verifyPassword(user.PasswordHash, req.Password); err != nil {
			return s.handleFailedPasswordAttempt(ctx, uow, user, req)
		}

		// Reset failed login attempts on successful authentication
		if err := s.resetFailedAttempts(ctx, uow, user); err != nil {
			return fmt.Errorf("failed to reset failed attempts: %w", err)
		}

		// Update last login
		if err := uow.Users().UpdateLastLogin(ctx, user.ID); err != nil {
			log.Printf("Warning: failed to update last login: %v", err)
		}

		// Check if two-factor is required
		requiresTwoFactor := user.TwoFactorEnabled

		// Create session
		session, err := s.createSession(ctx, uow, user, req)
		if err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}

		// Get user tenants if tenant context is provided
		var tenants []*models.UserTenant
		if req.TenantID != "" {
			tenants, err = uow.Audit().GetUserTenants(ctx, user.ID)
			if err != nil {
				log.Printf("Warning: failed to get user tenants: %v", err)
				tenants = []*models.UserTenant{}
			}
		}

		// Generate tokens
		accessToken, refreshToken, expiresAt, err := s.generateTokens(user, session, req.RememberMe)
		if err != nil {
			return fmt.Errorf("failed to generate tokens: %w", err)
		}

		// Log successful login
		s.logSuccessfulLoginAttempt(ctx, uow, user, req)
		s.logSecurityEvent(ctx, uow, user.ID.String(), "user_login", map[string]interface{}{
			"username":    user.Username,
			"email":       user.Email,
			"device_info": req.DeviceInfo,
			"ip_address":  req.IPAddress,
			"tenant_id":   req.TenantID,
		})

		// Publish user login event
		if s.eventService != nil {
			if err := s.eventService.PublishUserLoginEvent(ctx, user, session.ID.String(), req.TenantID, req.IPAddress); err != nil {
				log.Printf("Warning: failed to publish user login event: %v", err)
			}
		}

		loginResponse = &interfaces.LoginResponse{
			AccessToken:       accessToken,
			RefreshToken:      refreshToken,
			ExpiresAt:         expiresAt,
			User:              user,
			Session:           session,
			UserTenants:       tenants,
			RequiresTwoFactor: requiresTwoFactor,
			SessionID:         session.SessionID,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return loginResponse, nil
}

func (s *authService) LoginWithOTP(ctx context.Context, req *interfaces.LoginOTPRequest) (*interfaces.LoginResponse, error) {
	var loginResponse *interfaces.LoginResponse

	err := s.uowFactory.Create().ExecuteInTransaction(ctx, func(ctx context.Context, uow uow.UnitOfWork) error {
		// Get user based on identifier type
		var user *models.User
		var err error

		switch req.IdentifierType {
		case "email":
			user, err = uow.Users().GetByEmail(ctx, req.Identifier)
		case "username":
			user, err = uow.Users().GetByUsername(ctx, req.Identifier)
		case "phone":
			// Phone lookup not implemented yet
			return fmt.Errorf("phone authentication not yet supported")
		default:
			return fmt.Errorf("unsupported identifier type: %s", req.IdentifierType)
		}

		if err != nil || user == nil {
			s.logFailedLoginAttempt(ctx, uow, req.Identifier, "user_not_found", nil, req.IPAddress)
			return fmt.Errorf("authentication failed")
		}

		// Validate OTP based on identifier type
		var otpValid bool
		switch req.IdentifierType {
		case "phone":
			otpValid, err = s.validatePhoneOTP(ctx, uow, req.Identifier, req.OTPCode)
		case "email":
			otpValid, err = s.validateEmailOTP(ctx, uow, req.Identifier, req.OTPCode)
		default:
			return fmt.Errorf("OTP login not supported for identifier type: %s", req.IdentifierType)
		}

		if err != nil || !otpValid {
			s.logFailedLoginAttempt(ctx, uow, req.Identifier, "invalid_otp", &user.ID, req.IPAddress)
			return fmt.Errorf("invalid OTP code")
		}

		// Validate user account status
		if !user.IsActive {
			return fmt.Errorf("account is disabled")
		}

		// Create session
		session := &models.Session{
			SessionID:        generateSessionID(),
			UserID:           user.ID,
			RefreshTokenHash: generateRefreshTokenHash(),
			IsActive:         true,
			ExpiresAt:        time.Now().Add(24 * time.Hour),
			DeviceInfo:       map[string]interface{}{"user_agent": req.DeviceInfo},
		}

		if req.IPAddress != "" {
			session.IPAddress = &req.IPAddress
		}
		if req.TenantID != "" {
			session.TenantID = &req.TenantID
		}

		createdSession, err := uow.Sessions().Create(ctx, session)
		if err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}

		// Generate tokens
		accessToken, refreshToken, expiresAt, err := s.generateTokens(user, createdSession, false)
		if err != nil {
			return fmt.Errorf("failed to generate tokens: %w", err)
		}

		// Get user tenants
		var tenants []*models.UserTenant
		if req.TenantID != "" {
			tenants, err = uow.Audit().GetUserTenants(ctx, user.ID)
			if err != nil {
				log.Printf("Warning: failed to get user tenants: %v", err)
			}
		}

		// Log successful login
		s.logSuccessfulLoginAttempt(ctx, uow, user, &interfaces.LoginRequest{
			IdentifierType: req.IdentifierType,
			Identifier:     req.Identifier,
			DeviceInfo:     req.DeviceInfo,
			IPAddress:      req.IPAddress,
			TenantID:       req.TenantID,
		})

		// Publish user login event
		if s.eventService != nil {
			if err := s.eventService.PublishUserLoginEvent(ctx, user, createdSession.ID.String(), req.TenantID, req.IPAddress); err != nil {
				log.Printf("Warning: failed to publish user login event: %v", err)
			}
		}

		loginResponse = &interfaces.LoginResponse{
			AccessToken:       accessToken,
			RefreshToken:      refreshToken,
			ExpiresAt:         expiresAt,
			User:              user,
			Session:           createdSession,
			UserTenants:       tenants,
			RequiresTwoFactor: false, // OTP already verified
			SessionID:         createdSession.SessionID,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return loginResponse, nil
}

func (s *authService) Logout(ctx context.Context, sessionID string, deviceInfo string) error {
	return s.uowFactory.Create().ExecuteInTransaction(ctx, func(ctx context.Context, uow uow.UnitOfWork) error {
		// Get session to validate it exists and get user info
		session, err := uow.Sessions().GetBySessionID(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("session not found: %w", err)
		}

		// Use update with revoked status instead of Deactivate
		sessionUUID, err := uuid.Parse(sessionID)
		if err != nil {
			return fmt.Errorf("invalid session ID: %w", err)
		}
		session, err = uow.Sessions().GetByID(ctx, sessionUUID)
		if err != nil {
			return fmt.Errorf("session not found: %w", err)
		}
		now := time.Now()
		session.IsActive = false
		session.RevokedAt = &now
		reason := "user logout"
		session.RevokedReason = &reason
		_, err = uow.Sessions().Update(ctx, session)
		if err != nil {
			return fmt.Errorf("failed to deactivate session: %w", err)
		}

		// Log security event
		s.logSecurityEvent(ctx, uow, session.UserID.String(), "user_logout", map[string]interface{}{
			"session_id":  sessionID,
			"device_info": deviceInfo,
		})

		// Publish user logout event
		if s.eventService != nil {
			if err := s.eventService.PublishUserLogoutEvent(ctx, session.UserID.String(), sessionID, "user_requested"); err != nil {
				log.Printf("Warning: failed to publish user logout event: %v", err)
			}
		}

		return nil
	})
}

func (s *authService) LogoutAll(ctx context.Context, userID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	return s.uowFactory.Create().ExecuteInTransaction(ctx, func(ctx context.Context, uow uow.UnitOfWork) error {
		// Revoke all user sessions
		reason := "user logout all"
		err := uow.Sessions().RevokeAllUserSessions(ctx, userUUID, &reason)
		if err != nil {
			return fmt.Errorf("failed to revoke all sessions: %w", err)
		}

		// Log security event
		s.logSecurityEvent(ctx, uow, userID, "user_logout_all", map[string]interface{}{
			"user_id": userID,
		})

		// Publish user logout event
		if s.eventService != nil {
			if err := s.eventService.PublishUserLogoutEvent(ctx, userID, "all_sessions", "logout_all_requested"); err != nil {
				log.Printf("Warning: failed to publish user logout all event: %v", err)
			}
		}

		return nil
	})
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string, deviceInfo string) (*interfaces.RefreshTokenResponse, error) {
	// Get JWT service from module
	jwtService := s.jwtService

	// Refresh the token
	tokenPair, err := jwtService.RefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// TODO: Validate session is still active and update last activity

	return &interfaces.RefreshTokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
	}, nil
}

func (s *authService) ValidateToken(ctx context.Context, token string, requiredPermission, resource string) (*interfaces.ValidateTokenResponse, error) {
	// Get JWT service from module
	jwtService := s.jwtService

	// Validate the token
	claims, err := jwtService.ValidateToken(token)
	if err != nil {
		return &interfaces.ValidateTokenResponse{
			Valid: false,
		}, nil
	}

	// Check if token has expired
	if time.Now().After(claims.ExpiresAt) {
		return &interfaces.ValidateTokenResponse{
			Valid: false,
		}, nil
	}

	// TODO: Get user and session from database to ensure they're still valid
	// TODO: Check if token is revoked
	// TODO: Validate permissions against required permission and resource

	return &interfaces.ValidateTokenResponse{
		Valid:       true,
		ExpiresAt:   claims.ExpiresAt,
		Permissions: claims.Permissions,
		TenantID:    claims.TenantID,
	}, nil
}

func (s *authService) RevokeToken(ctx context.Context, token string, tokenType string) error {
	return s.uowFactory.Create().ExecuteInTransaction(ctx, func(ctx context.Context, uow uow.UnitOfWork) error {
		// Get JWT service from module
		jwtService := s.jwtService

		// Extract JTI from token
		jti, err := jwtService.ExtractJTI(token)
		if err != nil {
			return fmt.Errorf("failed to extract JTI from token: %w", err)
		}

		// Validate token to get expiry
		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			return fmt.Errorf("invalid token: %w", err)
		}

		// Create revoked token entry
		revokedToken := &models.RevokedToken{
			TokenJti:   jti,
			TokenType:  tokenType,
			RevokedAt:  time.Now(),
			ExpiresAt:  claims.ExpiresAt,
		}

		return uow.Tokens().RevokeToken(ctx, revokedToken)
	})
}

// Helper methods

func (s *authService) validateUserAccount(ctx context.Context, uow uow.UnitOfWork, user *models.User, req *interfaces.LoginRequest) error {
	if !user.IsActive {
		s.logFailedLoginAttempt(ctx, uow, req.Identifier, "account_disabled", &user.ID, req.IPAddress)
		return fmt.Errorf("account is disabled")
	}

	if user.AccountLockedUntil != nil && user.AccountLockedUntil.After(time.Now()) {
		s.logFailedLoginAttempt(ctx, uow, req.Identifier, "account_locked", &user.ID, req.IPAddress)
		return fmt.Errorf("account is locked")
	}

	return nil
}

func (s *authService) verifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func (s *authService) handleFailedPasswordAttempt(ctx context.Context, uow uow.UnitOfWork, user *models.User, req *interfaces.LoginRequest) error {
	// Increment failed login attempts
	user.FailedLoginAttempts++

	// Lock account after 5 failed attempts
	if user.FailedLoginAttempts >= 5 {
		lockUntil := time.Now().Add(30 * time.Minute)
		user.AccountLockedUntil = &lockUntil
	}

	_, err := uow.Users().Update(ctx, user)
	if err != nil {
		log.Printf("Failed to update user after failed password attempt: %v", err)
	}

	s.logFailedLoginAttempt(ctx, uow, req.Identifier, "invalid_password", &user.ID, req.IPAddress)
	return fmt.Errorf("authentication failed")
}

func (s *authService) resetFailedAttempts(ctx context.Context, uow uow.UnitOfWork, user *models.User) error {
	user.FailedLoginAttempts = 0
	user.AccountLockedUntil = nil
	_, err := uow.Users().Update(ctx, user)
	return err
}

func (s *authService) createSession(ctx context.Context, uow uow.UnitOfWork, user *models.User, req *interfaces.LoginRequest) (*models.Session, error) {
	session := &models.Session{
		SessionID:        generateSessionID(),
		UserID:           user.ID,
		RefreshTokenHash: generateRefreshTokenHash(),
		IsActive:         true,
		ExpiresAt:        time.Now().Add(24 * time.Hour),
		DeviceInfo:       map[string]interface{}{"user_agent": req.DeviceInfo},
	}

	if req.IPAddress != "" {
		session.IPAddress = &req.IPAddress
	}
	if req.TenantID != "" {
		session.TenantID = &req.TenantID
	}

	return uow.Sessions().Create(ctx, session)
}

func (s *authService) generateTokens(user *models.User, session *models.Session, rememberMe bool) (string, string, time.Time, error) {
	// Get JWT service from module
	jwtService := s.jwtService

	// Extract tenant ID and roles from session context
	tenantID := ""
	if session.TenantID != nil {
		tenantID = *session.TenantID
	}

	// TODO: Get user roles from database based on user and tenant
	var roles []string

	// Generate tokens
	tokenPair, err := jwtService.GenerateTokens(user, session, tenantID, roles)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return tokenPair.AccessToken, tokenPair.RefreshToken, tokenPair.ExpiresAt, nil
}

func (s *authService) validatePhoneOTP(ctx context.Context, uow uow.UnitOfWork, phone, otpCode string) (bool, error) {
	// Get phone verification token
	token, err := uow.Tokens().GetPhoneVerificationToken(ctx, phone, otpCode)
	if err != nil || token == nil {
		return false, err
	}

	// Check if token is valid and not expired
	if token.ExpiresAt.Before(time.Now()) || token.VerifiedAt != nil {
		return false, nil
	}

	// Mark token as used
	err = uow.Tokens().MarkPhoneVerificationTokenUsed(ctx, token.ID)
	return err == nil, err
}

func (s *authService) validateEmailOTP(ctx context.Context, uow uow.UnitOfWork, email, otpCode string) (bool, error) {
	// TODO: Implement email OTP validation
	// This would typically involve checking against email verification tokens
	return false, fmt.Errorf("email OTP validation not implemented")
}

func (s *authService) logFailedLoginAttempt(ctx context.Context, uow uow.UnitOfWork, identifier, reason string, userID *uuid.UUID, ipAddress string) {
	attempt := &models.LoginAttempt{
		Identifier:    identifier,
		Success:       false,
		FailureReason: &reason,
		AttemptedAt:   time.Now(),
	}

	if ipAddress != "" {
		attempt.IPAddress = ipAddress
	}
	if userID != nil {
		attempt.UserID = userID
	}

	_, err := uow.Audit().CreateLoginAttempt(ctx, attempt)
	if err != nil {
		log.Printf("Failed to log login attempt: %v", err)
	}
}

func (s *authService) logSuccessfulLoginAttempt(ctx context.Context, uow uow.UnitOfWork, user *models.User, req *interfaces.LoginRequest) {
	attempt := &models.LoginAttempt{
		Identifier:  req.Identifier,
		Success:     true,
		UserID:      &user.ID,
		AttemptedAt: time.Now(),
	}

	if req.IPAddress != "" {
		attempt.IPAddress = req.IPAddress
	}

	_, err := uow.Audit().CreateLoginAttempt(ctx, attempt)
	if err != nil {
		log.Printf("Failed to log successful login attempt: %v", err)
	}
}

func (s *authService) logSecurityEvent(ctx context.Context, uow uow.UnitOfWork, userID, eventType string, eventData map[string]interface{}) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		log.Printf("Invalid user ID for security event: %v", err)
		return
	}
	event := &models.SecurityEvent{
		UserID:    userUUID,
		EventType: eventType,
		EventData: eventData,
		CreatedAt: time.Now(),
	}

	_, err = uow.Audit().CreateSecurityEvent(ctx, event)
	if err != nil {
		log.Printf("Failed to log security event: %v", err)
	}
}

// Utility functions

func generateSessionID() string {
	return fmt.Sprintf("session_%d_%s", time.Now().UnixNano(), generateRandomString(8))
}

func generateRefreshTokenHash() string {
	return fmt.Sprintf("refresh_%s", generateRandomString(32))
}

func generateRandomString(length int) string {
	bytes := make([]byte, length/2)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

