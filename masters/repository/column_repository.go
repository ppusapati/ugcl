// =============================================================================
// repository/column_repository.go - Column metadata repository implementation
// =============================================================================
package repository

import (
	"context"

	db "masters/db/generated"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// ColumnRepository implements IColumnRepository
type ColumnRepository struct {
	queries *db.Queries
	db      *pgx.Conn
}

// NewColumnRepository creates a new column repository instance
func NewColumnRepository(database *pgx.Conn) IColumnRepository {
	return &ColumnRepository{
		queries: db.New(database),
		db:      database,
	}
}

// GetColumnsByTable retrieves all columns for a given table
func (r *ColumnRepository) GetColumnsByTable(ctx context.Context, tableID uuid.UUID) ([]*db.ColumnsMetadatum, error) {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(tableID.String()); err != nil {
		return nil, err
	}

	columns, err := r.queries.GetColumnsByTable(ctx, pgUUID)
	if err != nil {
		return nil, err
	}

	result := make([]*db.ColumnsMetadatum, len(columns))
	for i, column := range columns {
		result[i] = &column
	}

	return result, nil
}

// GetColumnByID retrieves a column by its ID
func (r *ColumnRepository) GetColumnByID(ctx context.Context, columnID uuid.UUID) (*db.ColumnsMetadatum, error) {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(columnID.String()); err != nil {
		return nil, err
	}

	column, err := r.queries.GetColumnByID(ctx, pgUUID)
	if err != nil {
		return nil, err
	}

	return &column, nil
}

// GetColumnByTableAndName retrieves a column by table ID and column name
func (r *ColumnRepository) GetColumnByTableAndName(ctx context.Context, tableID uuid.UUID, columnName string) (*db.ColumnsMetadatum, error) {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(tableID.String()); err != nil {
		return nil, err
	}

	column, err := r.queries.GetColumnByTableAndName(ctx, db.GetColumnByTableAndNameParams{
		TableID:    pgUUID,
		ColumnName: columnName,
	})
	if err != nil {
		return nil, err
	}

	return &column, nil
}

// GetImportableColumnsByTable retrieves all importable columns for a given table
func (r *ColumnRepository) GetImportableColumnsByTable(ctx context.Context, tableID uuid.UUID) ([]*db.ColumnsMetadatum, error) {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(tableID.String()); err != nil {
		return nil, err
	}

	columns, err := r.queries.GetImportableColumnsByTable(ctx, pgUUID)
	if err != nil {
		return nil, err
	}

	result := make([]*db.ColumnsMetadatum, len(columns))
	for i, column := range columns {
		result[i] = &column
	}

	return result, nil
}

// GetRequiredColumnsByTable retrieves all required columns for a given table
func (r *ColumnRepository) GetRequiredColumnsByTable(ctx context.Context, tableID uuid.UUID) ([]*db.ColumnsMetadatum, error) {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(tableID.String()); err != nil {
		return nil, err
	}

	columns, err := r.queries.GetRequiredColumnsByTable(ctx, pgUUID)
	if err != nil {
		return nil, err
	}

	result := make([]*db.ColumnsMetadatum, len(columns))
	for i, column := range columns {
		result[i] = &column
	}

	return result, nil
}

// CreateColumn creates a new column
func (r *ColumnRepository) CreateColumn(ctx context.Context, params *CreateColumnParams) (*db.ColumnsMetadatum, error) {
	var tableUUID pgtype.UUID
	if err := tableUUID.Scan(params.TableID.String()); err != nil {
		return nil, err
	}

	createParams := db.CreateColumnParams{
		TableID:        tableUUID,
		ColumnName:     params.ColumnName,
		DisplayName:    params.DisplayName,
		DataType:       params.DataType,
		IsRequired:     params.IsRequired,
		IsPrimaryKey:   params.IsPrimaryKey,
		IsUnique:       params.IsUnique,
		SupportsImport: params.SupportsImport,
		ColumnOrder:    params.ColumnOrder,
		CreatedBy:      params.CreatedBy,
	}

	// Handle optional fields
	if params.Description != nil {
		var pgText pgtype.Text
		if err := pgText.Scan(*params.Description); err != nil {
			return nil, err
		}
		createParams.Description = pgText
	}

	if params.MaxLength != nil {
		var pgInt4 pgtype.Int4
		if err := pgInt4.Scan(*params.MaxLength); err != nil {
			return nil, err
		}
		createParams.MaxLength = pgInt4
	}

	if params.DefaultValue != nil {
		var pgText pgtype.Text
		if err := pgText.Scan(*params.DefaultValue); err != nil {
			return nil, err
		}
		createParams.DefaultValue = pgText
	}

	if params.ValidationRules != nil {
		createParams.ValidationRules = params.ValidationRules
	}

	if len(params.ExampleValues) > 0 {
		createParams.ExampleValues = params.ExampleValues
	}

	column, err := r.queries.CreateColumn(ctx, createParams)
	if err != nil {
		return nil, err
	}

	return &column, nil
}

// UpdateColumn updates an existing column
func (r *ColumnRepository) UpdateColumn(ctx context.Context, params *UpdateColumnParams) (*db.ColumnsMetadatum, error) {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(params.ID.String()); err != nil {
		return nil, err
	}

	updateParams := db.UpdateColumnParams{
		ID:             pgUUID,
		DisplayName:    params.DisplayName,
		DataType:       params.DataType,
		IsRequired:     params.IsRequired,
		IsUnique:       params.IsUnique,
		SupportsImport: params.SupportsImport,
		ColumnOrder:    params.ColumnOrder,
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

	if params.MaxLength != nil {
		var pgInt4 pgtype.Int4
		if err := pgInt4.Scan(*params.MaxLength); err != nil {
			return nil, err
		}
		updateParams.MaxLength = pgInt4
	}

	if params.DefaultValue != nil {
		var pgText pgtype.Text
		if err := pgText.Scan(*params.DefaultValue); err != nil {
			return nil, err
		}
		updateParams.DefaultValue = pgText
	}

	if params.ValidationRules != nil {
		updateParams.ValidationRules = params.ValidationRules
	}

	if len(params.ExampleValues) > 0 {
		updateParams.ExampleValues = params.ExampleValues
	}

	column, err := r.queries.UpdateColumn(ctx, updateParams)
	if err != nil {
		return nil, err
	}

	return &column, nil
}

// DeactivateColumn deactivates a column
func (r *ColumnRepository) DeactivateColumn(ctx context.Context, columnID uuid.UUID, updatedBy string) error {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(columnID.String()); err != nil {
		return err
	}

	return r.queries.DeactivateColumn(ctx, db.DeactivateColumnParams{
		ID: pgUUID,
		UpdatedBy: pgtype.Text{
			String: updatedBy,
			Valid:  true,
		},
	})
}

// ReorderColumns updates the column order
func (r *ColumnRepository) ReorderColumns(ctx context.Context, columnID uuid.UUID, newOrder int32, updatedBy string) error {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(columnID.String()); err != nil {
		return err
	}

	return r.queries.ReorderColumns(ctx, db.ReorderColumnsParams{
		ID:          pgUUID,
		ColumnOrder: newOrder,
		UpdatedBy: pgtype.Text{
			String: updatedBy,
			Valid:  true,
		},
	})
}