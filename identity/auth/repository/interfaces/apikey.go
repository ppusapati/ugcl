package interfaces

import (
	"context"
	"time"

	"github.com/google/uuid"

	"p9e.in/ugcl/identity/auth/models"
)

type ApiKeyRepository interface {
	// Core CRUD operations
	Create(ctx context.Context, apiKey *models.ApiKey) (*models.ApiKey, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.ApiKey, error)
	GetByKeyID(ctx context.Context, keyID string) (*models.ApiKey, error)
	Update(ctx context.Context, apiKey *models.ApiKey) (*models.ApiKey, error)
	Delete(ctx context.Context, id uuid.UUID) error

	// API key management
	GetByUser(ctx context.Context, userID uuid.UUID) ([]*models.ApiKey, error)
	GetByTenant(ctx context.Context, tenantID string) ([]*models.ApiKey, error)
	Activate(ctx context.Context, id uuid.UUID) error
	Deactivate(ctx context.Context, id uuid.UUID) error
	UpdateLastUsed(ctx context.Context, keyID string) error

	// API key validation
	ValidateKey(ctx context.Context, keyID string, keyHash string) (*models.ApiKey, error)
	IsKeyActive(ctx context.Context, keyID string) (bool, error)

	// Search and filtering
	List(ctx context.Context, filter *models.ApiKeyFilter, pagination *models.Pagination) (*models.PaginatedResponse[*models.ApiKey], error)
	Count(ctx context.Context, filter *models.ApiKeyFilter) (int64, error)

	// Cleanup operations
	CleanupExpiredKeys(ctx context.Context) error

	// Analytics
	GetUsageStats(ctx context.Context, keyID string, since time.Time) (int64, error)
	GetMostUsedKeys(ctx context.Context, limit int32) ([]*models.ApiKey, error)
}