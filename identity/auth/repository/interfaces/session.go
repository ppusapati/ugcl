package interfaces

import (
	"context"
	"time"

	"github.com/google/uuid"

	"p9e.in/ugcl/identity/auth/models"
)

type SessionRepository interface {
	// Core CRUD operations
	Create(ctx context.Context, session *models.Session) (*models.Session, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Session, error)
	GetBySessionID(ctx context.Context, sessionID string) (*models.Session, error)
	Update(ctx context.Context, session *models.Session) (*models.Session, error)
	Delete(ctx context.Context, id uuid.UUID) error

	// Session management
	GetActiveSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Session, error)
	GetActiveSessionsByUserIDAndTenant(ctx context.Context, userID uuid.UUID, tenantID string) ([]*models.Session, error)
	UpdateLastAccessed(ctx context.Context, sessionID string) error
	RevokeSession(ctx context.Context, sessionID string, reason *string) error
	RevokeAllUserSessions(ctx context.Context, userID uuid.UUID, reason *string) error
	RevokeAllUserSessionsExcept(ctx context.Context, userID uuid.UUID, exceptSessionID string, reason *string) error

	// Session validation
	IsSessionActive(ctx context.Context, sessionID string) (bool, error)
	ValidateRefreshToken(ctx context.Context, sessionID string, refreshTokenHash string) (bool, error)

	// Search and filtering
	List(ctx context.Context, filter *models.SessionFilter, pagination *models.Pagination) (*models.PaginatedResponse[*models.Session], error)
	Count(ctx context.Context, filter *models.SessionFilter) (int64, error)

	// Cleanup operations
	CleanupExpiredSessions(ctx context.Context) error
	CleanupRevokedSessions(ctx context.Context, olderThan time.Time) error

	// Analytics
	GetActiveSessionsCount(ctx context.Context) (int64, error)
	GetSessionsCountByTenant(ctx context.Context, tenantID string) (int64, error)
	GetUserActiveSessionsCount(ctx context.Context, userID uuid.UUID) (int64, error)
	GetActiveSessionsByUser(ctx context.Context, userUUID uuid.UUID) ([]*models.Session, error)
	Deactivate(ctx context.Context, sessionID string) error
	DeactivateAll(ctx context.Context, userID uuid.UUID) error
	DeactivateAllExcept(ctx context.Context, userID uuid.UUID, sessionID string) error
}
