// =============================================================================
// repository/irepository.go - Repository interfaces for Masters module
// =============================================================================
package repository

import (
	"context"

	db "masters/db/generated"

	"github.com/google/uuid"
)

// ISchemaRepository defines schema metadata repository interface
type ISchemaRepository interface {
	GetAllSchemas(ctx context.Context) ([]*db.SchemasMetadatum, error)
	GetSchemaByID(ctx context.Context, schemaID uuid.UUID) (*db.SchemasMetadatum, error)
	GetSchemaByName(ctx context.Context, schemaName string) (*db.SchemasMetadatum, error)
	CreateSchema(ctx context.Context, params *CreateSchemaParams) (*db.SchemasMetadatum, error)
	UpdateSchema(ctx context.Context, params *UpdateSchemaParams) (*db.SchemasMetadatum, error)
	DeactivateSchema(ctx context.Context, schemaID uuid.UUID, updatedBy string) error
	SearchSchemas(ctx context.Context, query string) ([]*db.SchemasMetadatum, error)
}

// ITableRepository defines table metadata repository interface
type ITableRepository interface {
	GetTablesBySchema(ctx context.Context, schemaID uuid.UUID) ([]*db.GetTablesBySchemaRow, error)
	GetTableByID(ctx context.Context, tableID uuid.UUID) (*db.GetTableByIDRow, error)
	GetTableBySchemaAndName(ctx context.Context, schemaID uuid.UUID, tableName string) (*db.GetTableBySchemaAndNameRow, error)
	GetAllImportableTables(ctx context.Context) ([]*db.GetAllImportableTablesRow, error)
	CreateTable(ctx context.Context, params *CreateTableParams) (*db.TablesMetadatum, error)
	UpdateTable(ctx context.Context, params *UpdateTableParams) (*db.TablesMetadatum, error)
	DeactivateTable(ctx context.Context, tableID uuid.UUID, updatedBy string) error
	SearchTables(ctx context.Context, query string) ([]*db.SearchTablesRow, error)
}

// IColumnRepository defines column metadata repository interface
type IColumnRepository interface {
	GetColumnsByTable(ctx context.Context, tableID uuid.UUID) ([]*db.ColumnsMetadatum, error)
	GetColumnByID(ctx context.Context, columnID uuid.UUID) (*db.ColumnsMetadatum, error)
	GetColumnByTableAndName(ctx context.Context, tableID uuid.UUID, columnName string) (*db.ColumnsMetadatum, error)
	GetImportableColumnsByTable(ctx context.Context, tableID uuid.UUID) ([]*db.ColumnsMetadatum, error)
	GetRequiredColumnsByTable(ctx context.Context, tableID uuid.UUID) ([]*db.ColumnsMetadatum, error)
	CreateColumn(ctx context.Context, params *CreateColumnParams) (*db.ColumnsMetadatum, error)
	UpdateColumn(ctx context.Context, params *UpdateColumnParams) (*db.ColumnsMetadatum, error)
	DeactivateColumn(ctx context.Context, columnID uuid.UUID, updatedBy string) error
	ReorderColumns(ctx context.Context, columnID uuid.UUID, newOrder int32, updatedBy string) error
}

// Parameter structs for repository methods
type CreateSchemaParams struct {
	SchemaName  string
	DisplayName string
	Description *string
	CreatedBy   string
}

type UpdateSchemaParams struct {
	ID          uuid.UUID
	DisplayName string
	Description *string
	UpdatedBy   string
}

type CreateTableParams struct {
	SchemaID       uuid.UUID
	TableName      string
	DisplayName    string
	Description    *string
	SupportsImport bool
	CreatedBy      string
}

type UpdateTableParams struct {
	ID             uuid.UUID
	DisplayName    string
	Description    *string
	SupportsImport bool
	UpdatedBy      string
}

type CreateColumnParams struct {
	TableID         uuid.UUID
	ColumnName      string
	DisplayName     string
	Description     *string
	DataType        string
	MaxLength       *int32
	IsRequired      bool
	IsPrimaryKey    bool
	IsUnique        bool
	DefaultValue    *string
	ValidationRules []byte
	ExampleValues   []string
	SupportsImport  bool
	ColumnOrder     int32
	CreatedBy       string
}

type UpdateColumnParams struct {
	ID              uuid.UUID
	DisplayName     string
	Description     *string
	DataType        string
	MaxLength       *int32
	IsRequired      bool
	IsUnique        bool
	DefaultValue    *string
	ValidationRules []byte
	ExampleValues   []string
	SupportsImport  bool
	ColumnOrder     int32
	UpdatedBy       string
}