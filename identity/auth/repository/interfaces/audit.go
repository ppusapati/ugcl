package interfaces

import (
	"context"
	"time"

	"github.com/google/uuid"

	"p9e.in/ugcl/identity/auth/models"
)

type AuditRepository interface {
	// Security events
	CreateSecurityEvent(ctx context.Context, event *models.SecurityEvent) (*models.SecurityEvent, error)
	GetSecurityEventsByUser(ctx context.Context, userID uuid.UUID, pagination *models.Pagination) (*models.PaginatedResponse[*models.SecurityEvent], error)
	GetSecurityEventsByType(ctx context.Context, eventType string, pagination *models.Pagination) (*models.PaginatedResponse[*models.SecurityEvent], error)
	GetSecurityEventsByTenant(ctx context.Context, tenantID string, pagination *models.Pagination) (*models.PaginatedResponse[*models.SecurityEvent], error)
	CleanupOldSecurityEvents(ctx context.Context, olderThan time.Time) error

	// Login attempts
	CreateLoginAttempt(ctx context.Context, attempt *models.LoginAttempt) (*models.LoginAttempt, error)
	GetLoginAttemptsByUser(ctx context.Context, userID uuid.UUID, pagination *models.Pagination) (*models.PaginatedResponse[*models.LoginAttempt], error)
	GetLoginAttemptsByIdentifier(ctx context.Context, identifier string, pagination *models.Pagination) (*models.PaginatedResponse[*models.LoginAttempt], error)
	GetLoginAttemptsByIP(ctx context.Context, ipAddress string, pagination *models.Pagination) (*models.PaginatedResponse[*models.LoginAttempt], error)
	CleanupOldLoginAttempts(ctx context.Context, olderThan time.Time) error

	// Security analytics
	GetFailedLoginCount(ctx context.Context, identifier string, since time.Time) (int64, error)
	GetFailedLoginCountByIP(ctx context.Context, ipAddress string, since time.Time) (int64, error)
	GetSuspiciousActivities(ctx context.Context, since time.Time, pagination *models.Pagination) (*models.PaginatedResponse[*models.SecurityEvent], error)

	// User tenant management
	CreateUserTenant(ctx context.Context, userTenant *models.UserTenant) (*models.UserTenant, error)
	GetUserTenants(ctx context.Context, userID uuid.UUID) ([]*models.UserTenant, error)
	GetTenantUsers(ctx context.Context, tenantID string, pagination *models.Pagination) (*models.PaginatedResponse[*models.UserTenant], error)
	UpdateUserTenantRoles(ctx context.Context, userID uuid.UUID, tenantID string, roles []string) error
	SetDefaultTenant(ctx context.Context, userID uuid.UUID, tenantID string) error
	DeactivateUserTenant(ctx context.Context, userID uuid.UUID, tenantID string) error
}