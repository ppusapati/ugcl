package interfaces

import (
	"context"
	"time"

	"p9e.in/ugcl/identity/auth/models"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	// Authentication methods
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	LoginWithOTP(ctx context.Context, req *LoginOTPRequest) (*LoginResponse, error)
	Logout(ctx context.Context, sessionID string, deviceInfo string) error
	LogoutAll(ctx context.Context, userID string) error

	// Token management
	RefreshToken(ctx context.Context, refreshToken string, deviceInfo string) (*RefreshTokenResponse, error)
	ValidateToken(ctx context.Context, token string, requiredPermission, resource string) (*ValidateTokenResponse, error)
	RevokeToken(ctx context.Context, token string, tokenType string) error
}

// UserService defines the interface for user management operations
type UserService interface {
	// Password management
	ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error
	ForgotPassword(ctx context.Context, identifier, identifierType, tenantID string) error
	ResetPassword(ctx context.Context, resetToken, newPassword string) error

	// User profile
	GetUserProfile(ctx context.Context, userID string) (*models.User, error)
	UpdateUserProfile(ctx context.Context, userID string, updates *UserProfileUpdate) error
	DeactivateUser(ctx context.Context, userID string) error
	ActivateUser(ctx context.Context, userID string) error
}

// TwoFactorService defines the interface for two-factor authentication
type TwoFactorService interface {
	// Two-factor authentication
	EnableTwoFactor(ctx context.Context, userID string, method string) (*EnableTwoFactorResponse, error)
	DisableTwoFactor(ctx context.Context, userID, verificationCode string) error
	VerifyTwoFactor(ctx context.Context, userID, code, method string) error
	GenerateBackupCodes(ctx context.Context, userID string) ([]string, error)
	UseBackupCode(ctx context.Context, userID, code string) error
}

// SessionService defines the interface for session management
type SessionService interface {
	// Session management
	GetActiveSessions(ctx context.Context, userID string) ([]*models.Session, error)
	RevokeSession(ctx context.Context, sessionID, userID string) error
	RevokeAllSessions(ctx context.Context, userID string) error
	GetSessionByID(ctx context.Context, sessionID string) (*models.Session, error)
	ValidateSession(ctx context.Context, sessionID string) (*models.Session, error)
}

// TenantService defines the interface for multi-tenant operations
type TenantService interface {
	// Multi-tenant operations
	SwitchTenant(ctx context.Context, userID, targetTenantID, currentSessionID string) (*SwitchTenantResponse, error)
	GetUserTenants(ctx context.Context, userID string) ([]*models.UserTenant, error)
	ValidateUserTenantAccess(ctx context.Context, userID, tenantID string) (bool, error)
	GetUserRolesInTenant(ctx context.Context, userID, tenantID string) ([]string, error)
}

// JWTService defines the interface for JWT token operations
type JWTService interface {
	// Token generation and validation
	GenerateTokens(user *models.User, session *models.Session, tenantID string, roles []string) (*TokenPair, error)
	ValidateToken(tokenString string) (*TokenClaims, error)
	RefreshToken(refreshTokenString string) (*TokenPair, error)
	ExtractJTI(tokenString string) (string, error)
}

// OTPService defines the interface for OTP operations
type OTPService interface {
	// OTP generation and verification
	SendPhoneOTP(ctx context.Context, phone, purpose string) error
	SendEmailOTP(ctx context.Context, email, purpose string) error
	VerifyPhoneOTP(ctx context.Context, phone, code, purpose string) (bool, error)
	VerifyEmailOTP(ctx context.Context, email, code, purpose string) (bool, error)
	CleanupExpiredOTPs(ctx context.Context) error
}

// EmailService defines the interface for email operations
type EmailService interface {
	// Email sending operations
	SendOTP(ctx context.Context, email, code, purpose string) error
	SendPasswordReset(ctx context.Context, email, resetLink string) error
	SendWelcome(ctx context.Context, email, username string) error
}

// SMSService defines the interface for SMS operations
type SMSService interface {
	// SMS sending operations
	SendOTP(ctx context.Context, phone, code, purpose string) error
	SendVerification(ctx context.Context, phone, verificationLink string) error
}

// AuditService defines the interface for audit and security operations
type AuditService interface {
	// Security and audit
	GetSecurityEvents(ctx context.Context, userID string, pagination *models.Pagination) (*models.PaginatedResponse[*models.SecurityEvent], error)
	GetLoginAttempts(ctx context.Context, userID string, pagination *models.Pagination) (*models.PaginatedResponse[*models.LoginAttempt], error)
	GetSuspiciousActivities(ctx context.Context, since time.Time, pagination *models.Pagination) (*models.PaginatedResponse[*models.SecurityEvent], error)
	LogSecurityEvent(ctx context.Context, event *models.SecurityEvent) error
}

// Request/Response DTOs
type LoginRequest struct {
	IdentifierType string
	Identifier     string
	Password       string
	TenantID       string
	DeviceInfo     string
	IPAddress      string
	RememberMe     bool
}

type LoginOTPRequest struct {
	IdentifierType string
	Identifier     string
	OTPCode        string
	TenantID       string
	DeviceInfo     string
	IPAddress      string
}

type LoginResponse struct {
	AccessToken       string
	RefreshToken      string
	ExpiresAt         time.Time
	User              *models.User
	Session           *models.Session
	UserTenants       []*models.UserTenant
	RequiresTwoFactor bool
	SessionID         string
}

type RefreshTokenResponse struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

type ValidateTokenResponse struct {
	Valid       bool
	User        *models.User
	Session     *models.Session
	ExpiresAt   time.Time
	Permissions []string
	TenantID    string
}

type EnableTwoFactorResponse struct {
	Secret      string
	QRCode      string
	BackupCodes []string
}

type SwitchTenantResponse struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	User         *models.User
	Session      *models.Session
	SessionID    string
}

type UserProfileUpdate struct {
	FullName    *string
	Email       *string
	Phone       *string
	Username    *string
	IsActive    *bool
	Preferences map[string]interface{}
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	TokenType    string
}

type TokenClaims struct {
	UserID      string
	Username    string
	Email       string
	TenantID    string
	SessionID   string
	Roles       []string
	Permissions []string
	TokenType   string
	JTI         string
	ExpiresAt   time.Time
	IssuedAt    time.Time
}