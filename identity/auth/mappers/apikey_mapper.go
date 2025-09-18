package mappers

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	sqlcgen "p9e.in/ugcl/identity/auth/db/generated"
	"p9e.in/ugcl/identity/auth/models"
)

type ApiKeyMapper struct{}

func NewApiKeyMapper() *ApiKeyMapper {
	return &ApiKeyMapper{}
}

// ApiKey mapping functions

func (m *ApiKeyMapper) FromSQLCApiKey(sqlcApiKey sqlcgen.AuthApiKey) *models.ApiKey {
	apiKey := &models.ApiKey{
		ID:        sqlcApiKey.ID,
		KeyID:     sqlcApiKey.KeyID,
		KeyHash:   sqlcApiKey.KeyHash,
		Name:      sqlcApiKey.Name,
		Scopes:    sqlcApiKey.Scopes,
		IsActive:  func() bool { if sqlcApiKey.IsActive != nil { return *sqlcApiKey.IsActive }; return false }(),
		CreatedBy: uuid.UUID(sqlcApiKey.CreatedBy.Bytes),
		CreatedAt: sqlcApiKey.CreatedAt,
		UpdatedAt: sqlcApiKey.UpdatedAt,
	}

	if sqlcApiKey.TenantID != nil {
		apiKey.TenantID = sqlcApiKey.TenantID
	}
	if sqlcApiKey.LastUsedAt.Valid {
		apiKey.LastUsedAt = &sqlcApiKey.LastUsedAt.Time
	}
	if sqlcApiKey.ExpiresAt.Valid {
		apiKey.ExpiresAt = &sqlcApiKey.ExpiresAt.Time
	}

	return apiKey
}

func (m *ApiKeyMapper) ToCreateApiKeyParams(apiKey *models.ApiKey) sqlcgen.CreateAPIKeyParams {
	params := sqlcgen.CreateAPIKeyParams{
		KeyID:     apiKey.KeyID,
		KeyHash:   apiKey.KeyHash,
		Name:      apiKey.Name,
		Scopes:    apiKey.Scopes,
		CreatedBy: pgtype.UUID{Bytes: apiKey.CreatedBy, Valid: true},
	}

	if apiKey.TenantID != nil {
		params.TenantID = apiKey.TenantID
	}
	if apiKey.ExpiresAt != nil {
		params.ExpiresAt = pgtype.Timestamp{Time: *apiKey.ExpiresAt, Valid: true}
	}

	return params
}

// TODO: Implement when UpdateAPIKeyParams is added to SQLC queries
// func (m *ApiKeyMapper) ToUpdateApiKeyParams(apiKey *models.ApiKey) sqlcgen.UpdateAPIKeyParams {
// 	params := sqlcgen.UpdateAPIKeyParams{
// 		ID:       apiKey.ID,
// 		Name:     apiKey.Name,
// 		Scopes:   apiKey.Scopes,
// 		IsActive: &apiKey.IsActive,
// 	}

// 	if apiKey.TenantID != nil {
// 		params.TenantID = apiKey.TenantID
// 	}
// 	if apiKey.LastUsedAt != nil {
// 		params.LastUsedAt = pgtype.Timestamp{Time: *apiKey.LastUsedAt, Valid: true}
// 	}
// 	if apiKey.ExpiresAt != nil {
// 		params.ExpiresAt = pgtype.Timestamp{Time: *apiKey.ExpiresAt, Valid: true}
// 	}

// 	return params
// }