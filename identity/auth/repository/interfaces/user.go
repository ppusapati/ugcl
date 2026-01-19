package interfaces

import (
	"context"
	"time"

	"github.com/google/uuid"

	"p9e.in/ugcl/identity/auth/models"
)

type UserRepository interface {
	// Core CRUD operations
	Create(ctx context.Context, user *models.User) (*models.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByUserID(ctx context.Context, userID string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByPhone(ctx context.Context, identifier string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByEmailOrUsername(ctx context.Context, identifier string) (*models.User, error)
	Update(ctx context.Context, user *models.User) (*models.User, error)
	Delete(ctx context.Context, id uuid.UUID) error

	// User status operations
	Activate(ctx context.Context, id uuid.UUID) error
	Deactivate(ctx context.Context, id uuid.UUID) error

	// Password operations
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash, passwordSalt string) error
	UpdateFailedLoginAttempts(ctx context.Context, id uuid.UUID, attempts int32) error
	LockAccount(ctx context.Context, id uuid.UUID, lockedUntil *time.Time) error
	UnlockAccount(ctx context.Context, id uuid.UUID) error

	// Email/Phone verification
	MarkEmailVerified(ctx context.Context, id uuid.UUID) error
	MarkPhoneVerified(ctx context.Context, id uuid.UUID) error
	UpdateEmail(ctx context.Context, id uuid.UUID, email string) error
	UpdatePhone(ctx context.Context, id uuid.UUID, phone *string) error

	// Two-factor authentication
	EnableTwoFactor(ctx context.Context, id uuid.UUID, secret string) error
	DisableTwoFactor(ctx context.Context, id uuid.UUID) error
	UpdateBackupCodes(ctx context.Context, id uuid.UUID, codes []string) error

	// Login tracking
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
	UpdateLastPasswordChange(ctx context.Context, id uuid.UUID) error

	// Search and filtering
	List(ctx context.Context, filter *models.UserFilter, pagination *models.Pagination) (*models.PaginatedResponse[*models.User], error)
	Count(ctx context.Context, filter *models.UserFilter) (int64, error)
	Search(ctx context.Context, query string, pagination *models.Pagination) (*models.PaginatedResponse[*models.User], error)

	// User tenant operations
	GetUserTenants(ctx context.Context, userID uuid.UUID) ([]*models.UserTenant, error)
	GetDefaultTenant(ctx context.Context, userID uuid.UUID) (*models.UserTenant, error)

	// Validation
	EmailExists(ctx context.Context, email string) (bool, error)
	UsernameExists(ctx context.Context, username string) (bool, error)
	UserIDExists(ctx context.Context, userID string) (bool, error)
}
