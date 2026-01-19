// =============================================================================
// repository/table_repository.go - Table metadata repository implementation
// =============================================================================
package repository

import (
	"context"

	db "masters/db/generated"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// TableRepository implements ITableRepository
type TableRepository struct {
	queries *db.Queries
	db      *pgx.Conn
}

// NewTableRepository creates a new table repository instance
func NewTableRepository(database *pgx.Conn) ITableRepository {
	return &TableRepository{
		queries: db.New(database),
		db:      database,
	}
}

// GetTablesBySchema retrieves all tables for a given schema
func (r *TableRepository) GetTablesBySchema(ctx context.Context, schemaID uuid.UUID) ([]*db.GetTablesBySchemaRow, error) {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(schemaID.String()); err != nil {
		return nil, err
	}

	tables, err := r.queries.GetTablesBySchema(ctx, pgUUID)
	if err != nil {
		return nil, err
	}

	result := make([]*db.GetTablesBySchemaRow, len(tables))
	for i, table := range tables {
		result[i] = &table
	}

	return result, nil
}

// GetTableByID retrieves a table by its ID
func (r *TableRepository) GetTableByID(ctx context.Context, tableID uuid.UUID) (*db.GetTableByIDRow, error) {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(tableID.String()); err != nil {
		return nil, err
	}

	table, err := r.queries.GetTableByID(ctx, pgUUID)
	if err != nil {
		return nil, err
	}

	return &table, nil
}

// GetTableBySchemaAndName retrieves a table by schema ID and table name
func (r *TableRepository) GetTableBySchemaAndName(ctx context.Context, schemaID uuid.UUID, tableName string) (*db.GetTableBySchemaAndNameRow, error) {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(schemaID.String()); err != nil {
		return nil, err
	}

	table, err := r.queries.GetTableBySchemaAndName(ctx, db.GetTableBySchemaAndNameParams{
		SchemaID:  pgUUID,
		TableName: tableName,
	})
	if err != nil {
		return nil, err
	}

	return &table, nil
}

// GetAllImportableTables retrieves all importable tables across all schemas
func (r *TableRepository) GetAllImportableTables(ctx context.Context) ([]*db.GetAllImportableTablesRow, error) {
	tables, err := r.queries.GetAllImportableTables(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*db.GetAllImportableTablesRow, len(tables))
	for i, table := range tables {
		result[i] = &table
	}

	return result, nil
}

// CreateTable creates a new table
func (r *TableRepository) CreateTable(ctx context.Context, params *CreateTableParams) (*db.TablesMetadatum, error) {
	var schemaUUID pgtype.UUID
	if err := schemaUUID.Scan(params.SchemaID.String()); err != nil {
		return nil, err
	}

	createParams := db.CreateTableParams{
		SchemaID:       schemaUUID,
		TableName:      params.TableName,
		DisplayName:    params.DisplayName,
		SupportsImport: params.SupportsImport,
		CreatedBy:      params.CreatedBy,
	}

	if params.Description != nil {
		var pgText pgtype.Text
		if err := pgText.Scan(*params.Description); err != nil {
			return nil, err
		}
		createParams.Description = pgText
	}

	table, err := r.queries.CreateTable(ctx, createParams)
	if err != nil {
		return nil, err
	}

	return &table, nil
}

// UpdateTable updates an existing table
func (r *TableRepository) UpdateTable(ctx context.Context, params *UpdateTableParams) (*db.TablesMetadatum, error) {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(params.ID.String()); err != nil {
		return nil, err
	}

	updateParams := db.UpdateTableParams{
		ID:             pgUUID,
		DisplayName:    params.DisplayName,
		SupportsImport: params.SupportsImport,
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

	table, err := r.queries.UpdateTable(ctx, updateParams)
	if err != nil {
		return nil, err
	}

	return &table, nil
}

// DeactivateTable deactivates a table
func (r *TableRepository) DeactivateTable(ctx context.Context, tableID uuid.UUID, updatedBy string) error {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(tableID.String()); err != nil {
		return err
	}

	return r.queries.DeactivateTable(ctx, db.DeactivateTableParams{
		ID: pgUUID,
		UpdatedBy: pgtype.Text{
			String: updatedBy,
			Valid:  true,
		},
	})
}

// SearchTables searches tables by query
func (r *TableRepository) SearchTables(ctx context.Context, query string) ([]*db.SearchTablesRow, error) {
	tables, err := r.queries.SearchTables(ctx, query)
	if err != nil {
		return nil, err
	}

	result := make([]*db.SearchTablesRow, len(tables))
	for i, table := range tables {
		result[i] = &table
	}

	return result, nil
}