// =============================================================================
// services/masters_client.go - Client interface for calling Masters module
// =============================================================================
package services

import (
	"context"

	"github.com/google/uuid"
)

// MasterData represents basic master data structures that DataBridge needs
type MasterData struct {
	Schema *Schema `json:"schema,omitempty"`
	Table  *Table  `json:"table,omitempty"`
	Column *Column `json:"column,omitempty"`
}

// Schema represents schema metadata
type Schema struct {
	ID          string `json:"id"`
	SchemaName  string `json:"schema_name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description,omitempty"`
	IsActive    bool   `json:"is_active"`
}

// Table represents table metadata
type Table struct {
	ID             string `json:"id"`
	SchemaID       string `json:"schema_id"`
	TableName      string `json:"table_name"`
	DisplayName    string `json:"display_name"`
	Description    string `json:"description,omitempty"`
	IsActive       bool   `json:"is_active"`
	SupportsImport bool   `json:"supports_import"`
	SchemaName     string `json:"schema_name,omitempty"`
}

// Column represents column metadata
type Column struct {
	ID              string   `json:"id"`
	TableID         string   `json:"table_id"`
	ColumnName      string   `json:"column_name"`
	DisplayName     string   `json:"display_name"`
	Description     string   `json:"description,omitempty"`
	DataType        string   `json:"data_type"`
	MaxLength       *int32   `json:"max_length,omitempty"`
	IsRequired      bool     `json:"is_required"`
	IsPrimaryKey    bool     `json:"is_primary_key"`
	IsUnique        bool     `json:"is_unique"`
	DefaultValue    string   `json:"default_value,omitempty"`
	ValidationRules string   `json:"validation_rules,omitempty"`
	ExampleValues   []string `json:"example_values,omitempty"`
	IsActive        bool     `json:"is_active"`
	SupportsImport  bool     `json:"supports_import"`
	ColumnOrder     int32    `json:"column_order"`
}

// IMastersClient defines the interface for calling the Masters module
type IMastersClient interface {
	// Schema operations
	GetSchemas(ctx context.Context) ([]*Schema, error)
	GetSchemaByID(ctx context.Context, schemaID uuid.UUID) (*Schema, error)

	// Table operations
	GetTables(ctx context.Context, schemaID *uuid.UUID, importableOnly bool) ([]*Table, error)
	GetTableByID(ctx context.Context, tableID uuid.UUID) (*Table, error)

	// Column operations
	GetColumns(ctx context.Context, tableID uuid.UUID, importableOnly bool) ([]*Column, error)
	GetColumnByID(ctx context.Context, columnID uuid.UUID) (*Column, error)
}

// HTTPMastersClient implements IMastersClient using HTTP calls to Masters service
type HTTPMastersClient struct {
	baseURL string
}

// NewHTTPMastersClient creates a new HTTP-based masters client
func NewHTTPMastersClient(baseURL string) IMastersClient {
	return &HTTPMastersClient{
		baseURL: baseURL,
	}
}

// GetSchemas fetches all schemas from the Masters service
func (c *HTTPMastersClient) GetSchemas(ctx context.Context) ([]*Schema, error) {
	// TODO: Implement HTTP call to masters service
	// For now, return empty to avoid compilation errors
	return []*Schema{}, nil
}

// GetSchemaByID fetches a schema by ID from the Masters service
func (c *HTTPMastersClient) GetSchemaByID(ctx context.Context, schemaID uuid.UUID) (*Schema, error) {
	// TODO: Implement HTTP call to masters service
	return nil, nil
}

// GetTables fetches tables from the Masters service
func (c *HTTPMastersClient) GetTables(ctx context.Context, schemaID *uuid.UUID, importableOnly bool) ([]*Table, error) {
	// TODO: Implement HTTP call to masters service
	return []*Table{}, nil
}

// GetTableByID fetches a table by ID from the Masters service
func (c *HTTPMastersClient) GetTableByID(ctx context.Context, tableID uuid.UUID) (*Table, error) {
	// TODO: Implement HTTP call to masters service
	return nil, nil
}

// GetColumns fetches columns from the Masters service
func (c *HTTPMastersClient) GetColumns(ctx context.Context, tableID uuid.UUID, importableOnly bool) ([]*Column, error) {
	// TODO: Implement HTTP call to masters service
	return []*Column{}, nil
}

// GetColumnByID fetches a column by ID from the Masters service
func (c *HTTPMastersClient) GetColumnByID(ctx context.Context, columnID uuid.UUID) (*Column, error) {
	// TODO: Implement HTTP call to masters service
	return nil, nil
}