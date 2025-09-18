package sqlc

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	sqlcgen "p9e.in/ugcl/identity/auth/db/generated"
	"p9e.in/ugcl/identity/auth/mappers"
	"p9e.in/ugcl/identity/auth/models"
	"p9e.in/ugcl/identity/auth/repository/interfaces"
)

type apiKeyRepository struct {
	db     *pgxpool.Pool
	tx     pgx.Tx
	mapper *mappers.ApiKeyMapper
}

func NewApiKeyRepository(db *pgxpool.Pool, tx pgx.Tx) interfaces.ApiKeyRepository {
	return &apiKeyRepository{
		db:     db,
		tx:     tx,
		mapper: mappers.NewApiKeyMapper(),
	}
}

func (r *apiKeyRepository) getQueries() *sqlcgen.Queries {
	if r.tx != nil {
		return sqlcgen.New(r.tx)
	}
	return sqlcgen.New(r.db)
}

func (r *apiKeyRepository) Create(ctx context.Context, apiKey *models.ApiKey) (*models.ApiKey, error) {
	queries := r.getQueries()
	params := r.mapper.ToCreateApiKeyParams(apiKey)

	sqlcApiKey, err := queries.CreateAPIKey(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	return r.mapper.FromSQLCApiKey(*sqlcApiKey), nil
}

func (r *apiKeyRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.ApiKey, error) {
	// Note: This method expects KeyID string, but interface provides UUID
	// This is a design mismatch that needs to be resolved
	return nil, fmt.Errorf("GetByID not implemented - design mismatch between UUID and string KeyID")
}

func (r *apiKeyRepository) GetByKeyID(ctx context.Context, keyID string) (*models.ApiKey, error) {
	queries := r.getQueries()

	sqlcApiKey, err := queries.GetAPIKeyByID(ctx, sqlcgen.GetAPIKeyByIDParams{
		KeyID: keyID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get API key by key ID: %w", err)
	}

	return r.mapper.FromSQLCApiKey(*sqlcApiKey), nil
}

func (r *apiKeyRepository) GetByUser(ctx context.Context, userID uuid.UUID) ([]*models.ApiKey, error) {
	queries := r.getQueries()

	// Convert UUID to pgtype.UUID
	var pgUUID pgtype.UUID
	pgUUID.Scan(userID)

	sqlcApiKeys, err := queries.GetAPIKeysByUser(ctx, sqlcgen.GetAPIKeysByUserParams{
		CreatedBy: pgUUID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get API keys by user ID: %w", err)
	}

	apiKeys := make([]*models.ApiKey, len(sqlcApiKeys))
	for i, sqlcApiKey := range sqlcApiKeys {
		apiKeys[i] = r.mapper.FromSQLCApiKey(*sqlcApiKey)
	}

	return apiKeys, nil
}

func (r *apiKeyRepository) Update(ctx context.Context, apiKey *models.ApiKey) (*models.ApiKey, error) {
	// TODO: Implement when UpdateAPIKey query is available
	return nil, fmt.Errorf("Update not implemented")
}

func (r *apiKeyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// TODO: Implement when DeleteAPIKey query is available
	return fmt.Errorf("Delete not implemented")
}

func (r *apiKeyRepository) GetByTenant(ctx context.Context, tenantID string) ([]*models.ApiKey, error) {
	queries := r.getQueries()

	sqlcApiKeys, err := queries.GetAPIKeysByTenant(ctx, sqlcgen.GetAPIKeysByTenantParams{
		TenantID: &tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get API keys by tenant: %w", err)
	}

	apiKeys := make([]*models.ApiKey, len(sqlcApiKeys))
	for i, sqlcApiKey := range sqlcApiKeys {
		apiKeys[i] = r.mapper.FromSQLCApiKey(*sqlcApiKey)
	}

	return apiKeys, nil
}

// Stub implementations for interface methods not yet supported by SQLC
func (r *apiKeyRepository) Activate(ctx context.Context, id uuid.UUID) error {
	return fmt.Errorf("Activate not implemented")
}

func (r *apiKeyRepository) Deactivate(ctx context.Context, id uuid.UUID) error {
	// This method currently can't be implemented because DeactivateAPIKey
	// expects a KeyID string, but we only have the UUID ID
	// TODO: Add query to deactivate by UUID ID or implement lookup first
	return fmt.Errorf("Deactivate by UUID not implemented - design mismatch")
}

func (r *apiKeyRepository) UpdateLastUsed(ctx context.Context, keyID string) error {
	queries := r.getQueries()
	err := queries.UpdateAPIKeyLastUsed(ctx, sqlcgen.UpdateAPIKeyLastUsedParams{
		KeyID: keyID,
	})
	if err != nil {
		return fmt.Errorf("failed to update API key last used: %w", err)
	}
	return nil
}

func (r *apiKeyRepository) ValidateKey(ctx context.Context, keyID string, keyHash string) (*models.ApiKey, error) {
	return nil, fmt.Errorf("ValidateKey not implemented")
}

func (r *apiKeyRepository) IsKeyActive(ctx context.Context, keyID string) (bool, error) {
	return false, fmt.Errorf("IsKeyActive not implemented")
}

func (r *apiKeyRepository) List(ctx context.Context, filter *models.ApiKeyFilter, pagination *models.Pagination) (*models.PaginatedResponse[*models.ApiKey], error) {
	return nil, fmt.Errorf("List not implemented")
}

func (r *apiKeyRepository) Count(ctx context.Context, filter *models.ApiKeyFilter) (int64, error) {
	return 0, fmt.Errorf("Count not implemented")
}

func (r *apiKeyRepository) CleanupExpiredKeys(ctx context.Context) error {
	queries := r.getQueries()
	err := queries.CleanupExpiredAPIKeys(ctx)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired API keys: %w", err)
	}
	return nil
}

func (r *apiKeyRepository) GetUsageStats(ctx context.Context, keyID string, since time.Time) (int64, error) {
	return 0, fmt.Errorf("GetUsageStats not implemented")
}

func (r *apiKeyRepository) GetMostUsedKeys(ctx context.Context, limit int32) ([]*models.ApiKey, error) {
	return nil, fmt.Errorf("GetMostUsedKeys not implemented")
}

