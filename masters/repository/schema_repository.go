// =============================================================================
// repository/schema_repository.go - Schema metadata repository implementation
// =============================================================================
package repository

import (
	"context"

	db "masters/db/generated"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// SchemaRepository implements ISchemaRepository
type SchemaRepository struct {
	queries *db.Queries
	db      *pgx.Conn
}

// NewSchemaRepository creates a new schema repository instance
func NewSchemaRepository(database *pgx.Conn) ISchemaRepository {
	return &SchemaRepository{
		queries: db.New(database),
		db:      database,
	}
}

// GetAllSchemas retrieves all active schemas
func (r *SchemaRepository) GetAllSchemas(ctx context.Context) ([]*db.SchemasMetadatum, error) {
	schemas, err := r.queries.GetAllSchemas(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*db.SchemasMetadatum, len(schemas))
	for i, schema := range schemas {
		result[i] = &schema
	}

	return result, nil
}

// GetSchemaByID retrieves a schema by its ID
func (r *SchemaRepository) GetSchemaByID(ctx context.Context, schemaID uuid.UUID) (*db.SchemasMetadatum, error) {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(schemaID.String()); err != nil {
		return nil, err
	}

	schema, err := r.queries.GetSchemaByID(ctx, pgUUID)
	if err != nil {
		return nil, err
	}

	return &schema, nil
}

// GetSchemaByName retrieves a schema by its name
func (r *SchemaRepository) GetSchemaByName(ctx context.Context, schemaName string) (*db.SchemasMetadatum, error) {
	schema, err := r.queries.GetSchemaByName(ctx, schemaName)
	if err != nil {
		return nil, err
	}

	return &schema, nil
}

// CreateSchema creates a new schema
func (r *SchemaRepository) CreateSchema(ctx context.Context, params *CreateSchemaParams) (*db.SchemasMetadatum, error) {
	createParams := db.CreateSchemaParams{
		SchemaName:  params.SchemaName,
		DisplayName: params.DisplayName,
		CreatedBy:   params.CreatedBy,
	}

	if params.Description != nil {
		var pgText pgtype.Text
		if err := pgText.Scan(*params.Description); err != nil {
			return nil, err
		}
		createParams.Description = pgText
	}

	schema, err := r.queries.CreateSchema(ctx, createParams)
	if err != nil {
		return nil, err
	}

	return &schema, nil
}

// UpdateSchema updates an existing schema
func (r *SchemaRepository) UpdateSchema(ctx context.Context, params *UpdateSchemaParams) (*db.SchemasMetadatum, error) {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(params.ID.String()); err != nil {
		return nil, err
	}

	updateParams := db.UpdateSchemaParams{
		ID:          pgUUID,
		DisplayName: params.DisplayName,
		UpdatedBy: pgtype.Text{
			String: params.UpdatedBy,
			Valid:  true,
		},
	}

	if params.Description != nil {
		var pgText pgtype.Text
		if err := pgText.Scan(*params.Description); err != nil {
			return nil, err
		}
		updateParams.Description = pgText
	}

	schema, err := r.queries.UpdateSchema(ctx, updateParams)
	if err != nil {
		return nil, err
	}

	return &schema, nil
}

// DeactivateSchema deactivates a schema
func (r *SchemaRepository) DeactivateSchema(ctx context.Context, schemaID uuid.UUID, updatedBy string) error {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(schemaID.String()); err != nil {
		return err
	}

	return r.queries.DeactivateSchema(ctx, db.DeactivateSchemaParams{
		ID: pgUUID,
		UpdatedBy: pgtype.Text{
			String: updatedBy,
			Valid:  true,
		},
	})
}

// SearchSchemas searches schemas by query
func (r *SchemaRepository) SearchSchemas(ctx context.Context, query string) ([]*db.SchemasMetadatum, error) {
	schemas, err := r.queries.SearchSchemas(ctx, query)
	if err != nil {
		return nil, err
	}

	result := make([]*db.SchemasMetadatum, len(schemas))
	for i, schema := range schemas {
		result[i] = &schema
	}

	return result, nil
}