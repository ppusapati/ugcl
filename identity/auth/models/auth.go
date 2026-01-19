package models

import (
	"time"

	"github.com/google/uuid"
)

// Domain models for auth service

type User struct {
	ID                  uuid.UUID  `json:"id"`
	UserID              string     `json:"user_id"`
	Username            string     `json:"username"`
	Email               string     `json:"email"`
	Phone               *string    `json:"phone,omitempty"`
	PasswordHash        string     `json:"-"` // Never expose in JSON
	PasswordSalt        string     `json:"-"` // Never expose in JSON
	PasswordChangedAt   *time.Time `json:"password_changed_at,omitempty"`
	PasswordExpiresAt   *time.Time `json:"password_expires_at,omitempty"`
	FailedLoginAttempts int32      `json:"failed_login_attempts"`
	AccountLockedUntil  *time.Time `json:"account_locked_until,omitempty"`
	IsActive            bool       `json:"is_active"`
	EmailVerified       bool       `json:"email_verified"`
	PhoneVerified       bool       `json:"phone_verified"`
	TwoFactorEnabled    bool       `json:"two_factor_enabled"`
	TwoFactorSecret     *string    `json:"-"` // Never expose in JSON
	BackupCodes         []string   `json:"-"` // Never expose in JSON
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	LastLoginAt         *time.Time `json:"last_login_at,omitempty"`
	LastPasswordChange  *time.Time `json:"last_password_change,omitempty"`
}

type Session struct {
	ID               uuid.UUID `json:"id"`
	SessionID        string    `json:"session_id"`
	UserID           uuid.UUID `json:"user_id"`
	TenantID         *string   `json:"tenant_id,omitempty"`
	RefreshTokenHash string    `json:"-"` // Never expose in JSON
	DeviceInfo       map[string]interface{} `json:"device_info,omitempty"`
	IPAddress        *string   `json:"ip_address,omitempty"`
	UserAgent        *string   `json:"user_agent,omitempty"`
	IsActive         bool      `json:"is_active"`
	ExpiresAt        time.Time `json:"expires_at"`
	CreatedAt        time.Time `json:"created_at"`
	LastAccessedAt   *time.Time `json:"last_accessed_at,omitempty"`
	RevokedAt        *time.Time `json:"revoked_at,omitempty"`
	RevokedReason    *string   `json:"revoked_reason,omitempty"`
}

type ApiKey struct {
	ID         uuid.UUID `json:"id"`
	KeyID      string    `json:"key_id"`
	KeyHash    string    `json:"-"` // Never expose in JSON
	Name       string    `json:"name"`
	TenantID   *string   `json:"tenant_id,omitempty"`
	Scopes     []string  `json:"scopes"`
	IsActive   bool      `json:"is_active"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	CreatedBy  uuid.UUID `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type UserTenant struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	TenantID  string    `json:"tenant_id"`
	IsDefault bool      `json:"is_default"`
	IsActive  bool      `json:"is_active"`
	Roles     []string  `json:"roles"`
	JoinedAt  *time.Time `json:"joined_at,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PasswordResetToken struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	TokenHash string    `json:"-"` // Never expose in JSON
	ExpiresAt time.Time `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	IPAddress *string   `json:"ip_address,omitempty"`
	UserAgent *string   `json:"user_agent,omitempty"`
}

type EmailVerificationToken struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	Email      string    `json:"email"`
	TokenHash  string    `json:"-"` // Never expose in JSON
	ExpiresAt  time.Time `json:"expires_at"`
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type PhoneVerificationToken struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	Phone      string    `json:"phone"`
	OtpCode    string    `json:"-"` // Never expose in JSON
	ExpiresAt  time.Time `json:"expires_at"`
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
	Attempts   int32     `json:"attempts"`
	CreatedAt  time.Time `json:"created_at"`
}

type TwoFactorBackupCode struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	CodeHash  string    `json:"-"` // Never expose in JSON
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Token struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Token     string    `json:"token"`
	Type      string    `json:"type"`
	IsUsed    bool      `json:"is_used"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
}

type RevokedToken struct {
	ID        uuid.UUID `json:"id"`
	TokenJti  string    `json:"token_jti"`
	TokenType string    `json:"token_type"`
	UserID    uuid.UUID `json:"user_id"`
	RevokedAt time.Time `json:"revoked_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Reason    *string   `json:"reason,omitempty"`
}

type AuditLog struct {
	ID        uuid.UUID              `json:"id"`
	UserID    *uuid.UUID             `json:"user_id,omitempty"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	Details   map[string]interface{} `json:"details,omitempty"`
	IPAddress *string                `json:"ip_address,omitempty"`
	UserAgent *string                `json:"user_agent,omitempty"`
	TenantID  *string                `json:"tenant_id,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

type SecurityEvent struct {
	ID        uuid.UUID              `json:"id"`
	UserID    uuid.UUID              `json:"user_id"`
	EventType string                 `json:"event_type"`
	EventData map[string]interface{} `json:"event_data,omitempty"`
	IPAddress *string                `json:"ip_address,omitempty"`
	UserAgent *string                `json:"user_agent,omitempty"`
	TenantID  *string                `json:"tenant_id,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

type LoginAttempt struct {
	ID            uuid.UUID `json:"id"`
	Identifier    string    `json:"identifier"`
	IPAddress     string    `json:"ip_address"`
	UserAgent     *string   `json:"user_agent,omitempty"`
	Success       bool      `json:"success"`
	FailureReason *string   `json:"failure_reason,omitempty"`
	TenantID      *string   `json:"tenant_id,omitempty"`
	UserID        *uuid.UUID `json:"user_id,omitempty"`
	AttemptedAt   time.Time `json:"attempted_at"`
}

// Request/Response models

type CreateUserRequest struct {
	UserID    string  `json:"user_id" validate:"required"`
	Username  string  `json:"username" validate:"required,min=3,max=50"`
	Email     string  `json:"email" validate:"required,email"`
	Phone     *string `json:"phone,omitempty"`
	Password  string  `json:"password" validate:"required,min=8"`
	TenantID  *string `json:"tenant_id,omitempty"`
}

type UpdateUserRequest struct {
	Username *string `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	Email    *string `json:"email,omitempty" validate:"omitempty,email"`
	Phone    *string `json:"phone,omitempty"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

type LoginRequest struct {
	Identifier string `json:"identifier" validate:"required"` // email or username
	Password   string `json:"password" validate:"required"`
	TenantID   *string `json:"tenant_id,omitempty"`
	DeviceInfo map[string]interface{} `json:"device_info,omitempty"`
	RememberMe bool   `json:"remember_me"`
}

type LoginResponse struct {
	User         *User   `json:"user"`
	AccessToken  string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
	ExpiresIn    int64   `json:"expires_in"`
	SessionID    string  `json:"session_id"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type CreateApiKeyRequest struct {
	Name      string   `json:"name" validate:"required,min=1,max=100"`
	TenantID  *string  `json:"tenant_id,omitempty"`
	Scopes    []string `json:"scopes" validate:"required,min=1"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type CreateApiKeyResponse struct {
	ApiKey *ApiKey `json:"api_key"`
	RawKey string  `json:"raw_key"` // Only returned once
}

// Analytics models

type UserStats struct {
	TotalUsers          int64     `json:"total_users"`
	ActiveUsers         int64     `json:"active_users"`
	NewUsersCount       int64     `json:"new_users_count"`
	VerifiedUsers       int64     `json:"verified_users"`
	TwoFactorUsers      int64     `json:"two_factor_users"`
	LastCalculatedAt    time.Time `json:"last_calculated_at"`
}

// Search and filter models

type UserFilter struct {
	Username      *string `json:"username,omitempty"`
	Email         *string `json:"email,omitempty"`
	IsActive      *bool   `json:"is_active,omitempty"`
	EmailVerified *bool   `json:"email_verified,omitempty"`
	TenantID      *string `json:"tenant_id,omitempty"`
}

type SessionFilter struct {
	UserID   *uuid.UUID `json:"user_id,omitempty"`
	TenantID *string    `json:"tenant_id,omitempty"`
	IsActive *bool      `json:"is_active,omitempty"`
}

type ApiKeyFilter struct {
	TenantID  *string `json:"tenant_id,omitempty"`
	CreatedBy *uuid.UUID `json:"created_by,omitempty"`
	IsActive  *bool   `json:"is_active,omitempty"`
}

// Pagination
type Pagination struct {
	Limit  int32 `json:"limit" validate:"min=1,max=100"`
	Offset int32 `json:"offset" validate:"min=0"`
}

type PaginatedResponse[T any] struct {
	Items      []T   `json:"items"`
	TotalCount int64 `json:"total_count"`
	Limit      int32 `json:"limit"`
	Offset     int32 `json:"offset"`
	HasMore    bool  `json:"has_more"`
}