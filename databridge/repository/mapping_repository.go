// =============================================================================
// repository/mapping_repository.go - Import mapping repository implementation
// =============================================================================
package repository

import (
	"context"

	db "p9e.in/ugcl/databridge/db/generated"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// MappingRepository implements IMappingRepository
type MappingRepository struct {
	queries *db.Queries
	db      *pgx.Conn
}

// NewMappingRepository creates a new mapping repository instance
func NewMappingRepository(database *pgx.Conn) IMappingRepository {
	return &MappingRepository{
		queries: db.New(database),
		db:      database,
	}
}

// GetMappingsByTable retrieves all mappings for a given table
func (r *MappingRepository) GetMappingsByTable(ctx context.Context, tableID uuid.UUID) ([]*db.ImportMapping, error) {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(tableID.String()); err != nil {
		return nil, err
	}

	mappings, err := r.queries.GetMappingsByTable(ctx, pgUUID)
	if err != nil {
		return nil, err
	}

	result := make([]*db.ImportMapping, len(mappings))
	for i, mapping := range mappings {
		result[i] = &mapping
	}

	return result, nil
}

// GetMappingByID retrieves a mapping by its ID
func (r *MappingRepository) GetMappingByID(ctx context.Context, mappingID uuid.UUID) (*db.GetMappingByIDRow, error) {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(mappingID.String()); err != nil {
		return nil, err
	}

	mapping, err := r.queries.GetMappingByID(ctx, pgUUID)
	if err != nil {
		return nil, err
	}

	return &mapping, nil
}

// GetMappingByTableAndName retrieves a mapping by table ID and mapping name
func (r *MappingRepository) GetMappingByTableAndName(ctx context.Context, tableID uuid.UUID, mappingName string) (*db.ImportMapping, error) {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(tableID.String()); err != nil {
		return nil, err
	}

	mapping, err := r.queries.GetMappingByTableAndName(ctx, db.GetMappingByTableAndNameParams{
		TableID:     pgUUID,
		MappingName: mappingName,
	})
	if err != nil {
		return nil, err
	}

	return &mapping, nil
}

// GetRecentMappingsByUser retrieves recent mappings for a user
func (r *MappingRepository) GetRecentMappingsByUser(ctx context.Context, userID string, limit int32) ([]*db.GetRecentMappingsByUserRow, error) {
	mappings, err := r.queries.GetRecentMappingsByUser(ctx, db.GetRecentMappingsByUserParams{
		CreatedBy: userID,
		Limit:     limit,
	})
	if err != nil {
		return nil, err
	}

	result := make([]*db.GetRecentMappingsByUserRow, len(mappings))
	for i, mapping := range mappings {
		result[i] = &mapping
	}

	return result, nil
}

// CreateMapping creates a new mapping
func (r *MappingRepository) CreateMapping(ctx context.Context, params *CreateMappingParams) (*db.ImportMapping, error) {
	var tableUUID pgtype.UUID
	if err := tableUUID.Scan(params.TableID.String()); err != nil {
		return nil, err
	}

	createParams := db.CreateMappingParams{
		MappingName:   params.MappingName,
		TableID:       tableUUID,
		CsvHeaders:    params.CSVHeaders,
		FieldMappings: params.FieldMappings,
		CreatedBy:     params.CreatedBy,
	}

	// Handle optional fields
	if params.Description != nil {
		var pgText pgtype.Text
		if err := pgText.Scan(*params.Description); err != nil {
			return nil, err
		}
		createParams.Description = pgText
	}

	if params.TransformationRules != nil {
		createParams.TransformationRules = params.TransformationRules
	}

	if params.ValidationRules != nil {
		createParams.ValidationRules = params.ValidationRules
	}

	mapping, err := r.queries.CreateMapping(ctx, createParams)
	if err != nil {
		return nil, err
	}

	return &mapping, nil
}

// UpdateMapping updates an existing mapping
func (r *MappingRepository) UpdateMapping(ctx context.Context, params *UpdateMappingParams) (*db.ImportMapping, error) {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(params.ID.String()); err != nil {
		return nil, err
	}

	updateParams := db.UpdateMappingParams{
		ID:            pgUUID,
		MappingName:   params.MappingName,
		CsvHeaders:    params.CSVHeaders,
		FieldMappings: params.FieldMappings,
		UpdatedBy: pgtype.Text{
			String: params.UpdatedBy,
			Valid:  true,
		},
	}

	// Handle optional fields
	if params.Description != nil {
		var pgText pgtype.Text
		if err := pgText.Scan(*params.Description); err != nil {
			return nil, err
		}
		updateParams.Description = pgText
	}

	if params.TransformationRules != nil {
		updateParams.TransformationRules = params.TransformationRules
	}

	if params.ValidationRules != nil {
		updateParams.ValidationRules = params.ValidationRules
	}

	mapping, err := r.queries.UpdateMapping(ctx, updateParams)
	if err != nil {
		return nil, err
	}

	return &mapping, nil
}

// UpdateMappingUsage updates the usage statistics of a mapping
func (r *MappingRepository) UpdateMappingUsage(ctx context.Context, mappingID uuid.UUID) error {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(mappingID.String()); err != nil {
		return err
	}

	return r.queries.UpdateMappingUsage(ctx, pgUUID)
}

// DeactivateMapping deactivates a mapping
func (r *MappingRepository) DeactivateMapping(ctx context.Context, mappingID uuid.UUID, updatedBy string) error {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(mappingID.String()); err != nil {
		return err
	}

	return r.queries.DeactivateMapping(ctx, db.DeactivateMappingParams{
		ID: pgUUID,
		UpdatedBy: pgtype.Text{
			String: updatedBy,
			Valid:  true,
		},
	})
}

// SearchMappings searches mappings by query
func (r *MappingRepository) SearchMappings(ctx context.Context, query string) ([]*db.SearchMappingsRow, error) {
	mappings, err := r.queries.SearchMappings(ctx, query)
	if err != nil {
		return nil, err
	}

	result := make([]*db.SearchMappingsRow, len(mappings))
	for i, mapping := range mappings {
		result[i] = &mapping
	}

	return result, nil
}