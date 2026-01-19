package services

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "p9e.in/ugcl/masters/api/proto"
	"p9e.in/ugcl/masters/db/generated"
	"p9e.in/ugcl/masters/mappers"
)

// MetadataService implements the IMetadataService interface
type MetadataService struct {
	queries *generated.Queries
	mapper  *mappers.MetadataMapper
}

// NewMetadataService creates a new metadata service instance
func NewMetadataService(db *sql.DB) *MetadataService {
	return &MetadataService{
		queries: generated.New(db),
		mapper:  mappers.NewMetadataMapper(),
	}
}

// =============================================================================
// Schema Operations
// =============================================================================

func (s *MetadataService) GetSchemas(ctx context.Context) ([]*pb.Schema, error) {
	schemas, err := s.queries.GetAllSchemas(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get schemas: %w", err)
	}

	var result []*pb.Schema
	for _, schema := range schemas {
		result = append(result, s.mapper.ToProtoSchema(schema))
	}

	return result, nil
}

func (s *MetadataService) GetSchemaByID(ctx context.Context, schemaID uuid.UUID) (*pb.Schema, error) {
	schema, err := s.queries.GetSchemaByID(ctx, schemaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema by ID: %w", err)
	}

	return s.mapper.ToProtoSchema(schema), nil
}

func (s *MetadataService) CreateSchema(ctx context.Context, req *pb.CreateSchemaRequest) (*pb.Schema, error) {
	schema, err := s.queries.CreateSchema(ctx, generated.CreateSchemaParams{
		SchemaName:  req.SchemaName,
		DisplayName: req.DisplayName,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		CreatedBy:   req.CreatedBy,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return s.mapper.ToProtoSchema(schema), nil
}

func (s *MetadataService) UpdateSchema(ctx context.Context, req *pb.UpdateSchemaRequest) (*pb.Schema, error) {
	schemaID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid schema ID: %w", err)
	}

	schema, err := s.queries.UpdateSchema(ctx, generated.UpdateSchemaParams{
		ID:          schemaID,
		DisplayName: req.DisplayName,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		UpdatedBy:   sql.NullString{String: req.UpdatedBy, Valid: req.UpdatedBy != ""},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update schema: %w", err)
	}

	return s.mapper.ToProtoSchema(schema), nil
}

func (s *MetadataService) DeleteSchema(ctx context.Context, schemaID uuid.UUID, userID string) error {
	err := s.queries.DeactivateSchema(ctx, generated.DeactivateSchemaParams{
		ID:        schemaID,
		UpdatedBy: sql.NullString{String: userID, Valid: userID != ""},
	})
	if err != nil {
		return fmt.Errorf("failed to delete schema: %w", err)
	}

	return nil
}

// =============================================================================
// Table Operations
// =============================================================================

func (s *MetadataService) GetTables(ctx context.Context, schemaID *uuid.UUID, importableOnly bool) ([]*pb.Table, error) {
	var tables []generated.GetTablesBySchemaRow
	var err error

	if schemaID != nil {
		tables, err = s.queries.GetTablesBySchema(ctx, *schemaID)
	} else {
		allTables, err := s.queries.GetAllTables(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get tables: %w", err)
		}

		// Convert to the expected type
		for _, table := range allTables {
			tables = append(tables, generated.GetTablesBySchemaRow{
				ID:                  table.ID,
				SchemaID:            table.SchemaID,
				TableName:           table.TableName,
				DisplayName:         table.DisplayName,
				Description:         table.Description,
				IsActive:            table.IsActive,
				SupportsImport:      table.SupportsImport,
				CreatedAt:           table.CreatedAt,
				UpdatedAt:           table.UpdatedAt,
				CreatedBy:           table.CreatedBy,
				UpdatedBy:           table.UpdatedBy,
				SchemaName:          sql.NullString{}, // Will be filled by join
				SchemaDisplayName:   sql.NullString{}, // Will be filled by join
			})
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}

	var result []*pb.Table
	for _, table := range tables {
		if !importableOnly || table.SupportsImport {
			result = append(result, s.mapper.ToProtoTable(table))
		}
	}

	return result, nil
}

func (s *MetadataService) GetTableByID(ctx context.Context, tableID uuid.UUID) (*pb.Table, error) {
	table, err := s.queries.GetTableByID(ctx, tableID)
	if err != nil {
		return nil, fmt.Errorf("failed to get table by ID: %w", err)
	}

	// Convert to the expected row type for mapper
	tableRow := generated.GetTablesBySchemaRow{
		ID:                table.ID,
		SchemaID:          table.SchemaID,
		TableName:         table.TableName,
		DisplayName:       table.DisplayName,
		Description:       table.Description,
		IsActive:          table.IsActive,
		SupportsImport:    table.SupportsImport,
		CreatedAt:         table.CreatedAt,
		UpdatedAt:         table.UpdatedAt,
		CreatedBy:         table.CreatedBy,
		UpdatedBy:         table.UpdatedBy,
		SchemaName:        sql.NullString{},
		SchemaDisplayName: sql.NullString{},
	}

	return s.mapper.ToProtoTable(tableRow), nil
}

func (s *MetadataService) CreateTable(ctx context.Context, req *pb.CreateTableRequest) (*pb.Table, error) {
	schemaID, err := uuid.Parse(req.SchemaId)
	if err != nil {
		return nil, fmt.Errorf("invalid schema ID: %w", err)
	}

	table, err := s.queries.CreateTable(ctx, generated.CreateTableParams{
		SchemaID:       schemaID,
		TableName:      req.TableName,
		DisplayName:    req.DisplayName,
		Description:    sql.NullString{String: req.Description, Valid: req.Description != ""},
		SupportsImport: req.SupportsImport,
		CreatedBy:      req.CreatedBy,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	// Convert to the expected row type for mapper
	tableRow := generated.GetTablesBySchemaRow{
		ID:                table.ID,
		SchemaID:          table.SchemaID,
		TableName:         table.TableName,
		DisplayName:       table.DisplayName,
		Description:       table.Description,
		IsActive:          table.IsActive,
		SupportsImport:    table.SupportsImport,
		CreatedAt:         table.CreatedAt,
		UpdatedAt:         table.UpdatedAt,
		CreatedBy:         table.CreatedBy,
		UpdatedBy:         table.UpdatedBy,
		SchemaName:        sql.NullString{},
		SchemaDisplayName: sql.NullString{},
	}

	return s.mapper.ToProtoTable(tableRow), nil
}

func (s *MetadataService) UpdateTable(ctx context.Context, req *pb.UpdateTableRequest) (*pb.Table, error) {
	tableID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid table ID: %w", err)
	}

	table, err := s.queries.UpdateTable(ctx, generated.UpdateTableParams{
		ID:             tableID,
		DisplayName:    req.DisplayName,
		Description:    sql.NullString{String: req.Description, Valid: req.Description != ""},
		SupportsImport: req.SupportsImport,
		UpdatedBy:      sql.NullString{String: req.UpdatedBy, Valid: req.UpdatedBy != ""},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update table: %w", err)
	}

	// Convert to the expected row type for mapper
	tableRow := generated.GetTablesBySchemaRow{
		ID:                table.ID,
		SchemaID:          table.SchemaID,
		TableName:         table.TableName,
		DisplayName:       table.DisplayName,
		Description:       table.Description,
		IsActive:          table.IsActive,
		SupportsImport:    table.SupportsImport,
		CreatedAt:         table.CreatedAt,
		UpdatedAt:         table.UpdatedAt,
		CreatedBy:         table.CreatedBy,
		UpdatedBy:         table.UpdatedBy,
		SchemaName:        sql.NullString{},
		SchemaDisplayName: sql.NullString{},
	}

	return s.mapper.ToProtoTable(tableRow), nil
}

func (s *MetadataService) DeleteTable(ctx context.Context, tableID uuid.UUID, userID string) error {
	err := s.queries.DeactivateTable(ctx, generated.DeactivateTableParams{
		ID:        tableID,
		UpdatedBy: sql.NullString{String: userID, Valid: userID != ""},
	})
	if err != nil {
		return fmt.Errorf("failed to delete table: %w", err)
	}

	return nil
}

// =============================================================================
// Column Operations
// =============================================================================

func (s *MetadataService) GetColumns(ctx context.Context, tableID uuid.UUID, importableOnly bool) ([]*pb.Column, error) {
	columns, err := s.queries.GetColumnsByTable(ctx, tableID)
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var result []*pb.Column
	for _, column := range columns {
		if !importableOnly || column.SupportsImport {
			result = append(result, s.mapper.ToProtoColumn(column))
		}
	}

	return result, nil
}

func (s *MetadataService) GetColumnByID(ctx context.Context, columnID uuid.UUID) (*pb.Column, error) {
	column, err := s.queries.GetColumnByID(ctx, columnID)
	if err != nil {
		return nil, fmt.Errorf("failed to get column by ID: %w", err)
	}

	return s.mapper.ToProtoColumn(column), nil
}

func (s *MetadataService) CreateColumn(ctx context.Context, req *pb.CreateColumnRequest) (*pb.Column, error) {
	tableID, err := uuid.Parse(req.TableId)
	if err != nil {
		return nil, fmt.Errorf("invalid table ID: %w", err)
	}

	// Convert protobuf DataType to string
	dataType := s.mapper.DataTypeToString(req.DataType)

	column, err := s.queries.CreateColumn(ctx, generated.CreateColumnParams{
		TableID:         tableID,
		ColumnName:      req.ColumnName,
		DisplayName:     req.DisplayName,
		Description:     sql.NullString{String: req.Description, Valid: req.Description != ""},
		DataType:        dataType,
		MaxLength:       sql.NullInt32{Int32: req.MaxLength, Valid: req.MaxLength > 0},
		IsRequired:      req.IsRequired,
		IsPrimaryKey:    req.IsPrimaryKey,
		IsUnique:        req.IsUnique,
		DefaultValue:    sql.NullString{String: req.DefaultValue, Valid: req.DefaultValue != ""},
		ValidationRules: []byte(req.ValidationRules),
		ExampleValues:   req.ExampleValues,
		SupportsImport:  req.SupportsImport,
		ColumnOrder:     req.ColumnOrder,
		CreatedBy:       req.CreatedBy,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create column: %w", err)
	}

	return s.mapper.ToProtoColumn(column), nil
}

func (s *MetadataService) UpdateColumn(ctx context.Context, req *pb.UpdateColumnRequest) (*pb.Column, error) {
	columnID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid column ID: %w", err)
	}

	// Convert protobuf DataType to string
	dataType := s.mapper.DataTypeToString(req.DataType)

	column, err := s.queries.UpdateColumn(ctx, generated.UpdateColumnParams{
		ID:              columnID,
		DisplayName:     req.DisplayName,
		Description:     sql.NullString{String: req.Description, Valid: req.Description != ""},
		DataType:        dataType,
		MaxLength:       sql.NullInt32{Int32: req.MaxLength, Valid: req.MaxLength > 0},
		IsRequired:      req.IsRequired,
		IsUnique:        req.IsUnique,
		DefaultValue:    sql.NullString{String: req.DefaultValue, Valid: req.DefaultValue != ""},
		ValidationRules: []byte(req.ValidationRules),
		ExampleValues:   req.ExampleValues,
		SupportsImport:  req.SupportsImport,
		ColumnOrder:     req.ColumnOrder,
		UpdatedBy:       sql.NullString{String: req.UpdatedBy, Valid: req.UpdatedBy != ""},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update column: %w", err)
	}

	return s.mapper.ToProtoColumn(column), nil
}

func (s *MetadataService) DeleteColumn(ctx context.Context, columnID uuid.UUID, userID string) error {
	err := s.queries.DeactivateColumn(ctx, generated.DeactivateColumnParams{
		ID:        columnID,
		UpdatedBy: sql.NullString{String: userID, Valid: userID != ""},
	})
	if err != nil {
		return fmt.Errorf("failed to delete column: %w", err)
	}

	return nil
}

// =============================================================================
// Relationship Operations
// =============================================================================

func (s *MetadataService) CreateRelationship(ctx context.Context, req *pb.CreateRelationshipRequest) (*pb.TableRelationship, error) {
	sourceTableID, err := uuid.Parse(req.SourceTableId)
	if err != nil {
		return nil, fmt.Errorf("invalid source table ID: %w", err)
	}

	sourceColumnID, err := uuid.Parse(req.SourceColumnId)
	if err != nil {
		return nil, fmt.Errorf("invalid source column ID: %w", err)
	}

	targetTableID, err := uuid.Parse(req.TargetTableId)
	if err != nil {
		return nil, fmt.Errorf("invalid target table ID: %w", err)
	}

	targetColumnID, err := uuid.Parse(req.TargetColumnId)
	if err != nil {
		return nil, fmt.Errorf("invalid target column ID: %w", err)
	}

	relationshipType := s.mapper.RelationshipTypeToString(req.RelationshipType)

	relationship, err := s.queries.CreateRelationship(ctx, generated.CreateRelationshipParams{
		Name:             sql.NullString{String: req.Name, Valid: req.Name != ""},
		SourceTableID:    sourceTableID,
		SourceColumnID:   sourceColumnID,
		TargetTableID:    targetTableID,
		TargetColumnID:   targetColumnID,
		RelationshipType: relationshipType,
		CreatedBy:        req.CreatedBy,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create relationship: %w", err)
	}

	return s.mapper.ToProtoTableRelationship(relationship), nil
}

func (s *MetadataService) UpdateRelationship(ctx context.Context, req *pb.UpdateRelationshipRequest) (*pb.TableRelationship, error) {
	relationshipID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid relationship ID: %w", err)
	}

	relationshipType := s.mapper.RelationshipTypeToString(req.RelationshipType)

	relationship, err := s.queries.UpdateRelationship(ctx, generated.UpdateRelationshipParams{
		ID:               relationshipID,
		Name:             sql.NullString{String: req.Name, Valid: req.Name != ""},
		RelationshipType: relationshipType,
		UpdatedBy:        sql.NullString{String: req.UpdatedBy, Valid: req.UpdatedBy != ""},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update relationship: %w", err)
	}

	return s.mapper.ToProtoTableRelationship(relationship), nil
}

func (s *MetadataService) DeleteRelationship(ctx context.Context, relationshipID uuid.UUID, userID string) error {
	err := s.queries.DeleteRelationship(ctx, generated.DeleteRelationshipParams{
		ID:        relationshipID,
		UpdatedBy: sql.NullString{String: userID, Valid: userID != ""},
	})
	if err != nil {
		return fmt.Errorf("failed to delete relationship: %w", err)
	}

	return nil
}

func (s *MetadataService) GetRelationshipsByTable(ctx context.Context, tableID uuid.UUID) ([]*pb.TableRelationship, error) {
	relationships, err := s.queries.ListRelationshipsByTable(ctx, tableID)
	if err != nil {
		return nil, fmt.Errorf("failed to get relationships by table: %w", err)
	}

	var result []*pb.TableRelationship
	for _, relationship := range relationships {
		result = append(result, s.mapper.ToProtoTableRelationshipWithDetails(relationship))
	}

	return result, nil
}

func (s *MetadataService) GetAllRelationships(ctx context.Context) ([]*pb.TableRelationship, error) {
	relationships, err := s.queries.ListAllRelationships(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all relationships: %w", err)
	}

	var result []*pb.TableRelationship
	for _, relationship := range relationships {
		result = append(result, s.mapper.ToProtoTableRelationshipWithDetails(relationship))
	}

	return result, nil
}

// =============================================================================
// Business Term Operations
// =============================================================================

func (s *MetadataService) CreateBusinessTerm(ctx context.Context, req *pb.CreateBusinessTermRequest) (*pb.BusinessTerm, error) {
	var ownerUserID uuid.NullUUID
	if req.OwnerUserId != "" {
		id, err := uuid.Parse(req.OwnerUserId)
		if err != nil {
			return nil, fmt.Errorf("invalid owner user ID: %w", err)
		}
		ownerUserID = uuid.NullUUID{UUID: id, Valid: true}
	}

	businessTerm, err := s.queries.CreateBusinessTerm(ctx, generated.CreateBusinessTermParams{
		Term:            req.Term,
		Definition:      req.Definition,
		BusinessContext: sql.NullString{String: req.BusinessContext, Valid: req.BusinessContext != ""},
		Domain:          sql.NullString{String: req.Domain, Valid: req.Domain != ""},
		OwnerUserID:     ownerUserID,
		CreatedBy:       req.CreatedBy,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create business term: %w", err)
	}

	return s.mapper.ToProtoBusinessTerm(businessTerm), nil
}

func (s *MetadataService) UpdateBusinessTerm(ctx context.Context, req *pb.UpdateBusinessTermRequest) (*pb.BusinessTerm, error) {
	businessTermID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid business term ID: %w", err)
	}

	var ownerUserID uuid.NullUUID
	if req.OwnerUserId != "" {
		id, err := uuid.Parse(req.OwnerUserId)
		if err != nil {
			return nil, fmt.Errorf("invalid owner user ID: %w", err)
		}
		ownerUserID = uuid.NullUUID{UUID: id, Valid: true}
	}

	businessTerm, err := s.queries.UpdateBusinessTerm(ctx, generated.UpdateBusinessTermParams{
		ID:              businessTermID,
		Definition:      req.Definition,
		BusinessContext: sql.NullString{String: req.BusinessContext, Valid: req.BusinessContext != ""},
		Domain:          sql.NullString{String: req.Domain, Valid: req.Domain != ""},
		OwnerUserID:     ownerUserID,
		UpdatedBy:       sql.NullString{String: req.UpdatedBy, Valid: req.UpdatedBy != ""},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update business term: %w", err)
	}

	return s.mapper.ToProtoBusinessTerm(businessTerm), nil
}

func (s *MetadataService) DeleteBusinessTerm(ctx context.Context, businessTermID uuid.UUID, userID string) error {
	err := s.queries.DeleteBusinessTerm(ctx, generated.DeleteBusinessTermParams{
		ID:        businessTermID,
		UpdatedBy: sql.NullString{String: userID, Valid: userID != ""},
	})
	if err != nil {
		return fmt.Errorf("failed to delete business term: %w", err)
	}

	return nil
}

func (s *MetadataService) GetBusinessTerms(ctx context.Context) ([]*pb.BusinessTerm, error) {
	businessTerms, err := s.queries.ListBusinessTerms(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get business terms: %w", err)
	}

	var result []*pb.BusinessTerm
	for _, businessTerm := range businessTerms {
		result = append(result, s.mapper.ToProtoBusinessTerm(businessTerm))
	}

	return result, nil
}

func (s *MetadataService) GetBusinessTermsByDomain(ctx context.Context, domain string) ([]*pb.BusinessTerm, error) {
	businessTerms, err := s.queries.ListBusinessTermsByDomain(ctx, sql.NullString{String: domain, Valid: domain != ""})
	if err != nil {
		return nil, fmt.Errorf("failed to get business terms by domain: %w", err)
	}

	var result []*pb.BusinessTerm
	for _, businessTerm := range businessTerms {
		result = append(result, s.mapper.ToProtoBusinessTerm(businessTerm))
	}

	return result, nil
}

func (s *MetadataService) SearchBusinessTerms(ctx context.Context, query string) ([]*pb.BusinessTerm, error) {
	businessTerms, err := s.queries.SearchBusinessTerms(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search business terms: %w", err)
	}

	var result []*pb.BusinessTerm
	for _, businessTerm := range businessTerms {
		result = append(result, s.mapper.ToProtoBusinessTerm(businessTerm))
	}

	return result, nil
}

func (s *MetadataService) LinkColumnToBusinessTerm(ctx context.Context, req *pb.LinkColumnToBusinessTermRequest) (*pb.ColumnBusinessTerm, error) {
	columnID, err := uuid.Parse(req.ColumnId)
	if err != nil {
		return nil, fmt.Errorf("invalid column ID: %w", err)
	}

	businessTermID, err := uuid.Parse(req.BusinessTermId)
	if err != nil {
		return nil, fmt.Errorf("invalid business term ID: %w", err)
	}

	link, err := s.queries.LinkColumnToBusinessTerm(ctx, generated.LinkColumnToBusinessTermParams{
		ColumnID:       columnID,
		BusinessTermID: businessTermID,
		CreatedBy:      req.CreatedBy,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to link column to business term: %w", err)
	}

	return &pb.ColumnBusinessTerm{
		Id:             link.ID.String(),
		ColumnId:       link.ColumnID.String(),
		BusinessTermId: link.BusinessTermID.String(),
		CreatedAt:      timestamppb.New(link.CreatedAt),
		CreatedBy:      link.CreatedBy,
	}, nil
}

func (s *MetadataService) UnlinkColumnFromBusinessTerm(ctx context.Context, columnID, businessTermID uuid.UUID) error {
	err := s.queries.UnlinkColumnFromBusinessTerm(ctx, generated.UnlinkColumnFromBusinessTermParams{
		ColumnID:       columnID,
		BusinessTermID: businessTermID,
	})
	if err != nil {
		return fmt.Errorf("failed to unlink column from business term: %w", err)
	}

	return nil
}

func (s *MetadataService) GetBusinessTermsForColumn(ctx context.Context, columnID uuid.UUID) ([]*pb.BusinessTerm, error) {
	businessTerms, err := s.queries.GetBusinessTermsForColumn(ctx, columnID)
	if err != nil {
		return nil, fmt.Errorf("failed to get business terms for column: %w", err)
	}

	var result []*pb.BusinessTerm
	for _, businessTerm := range businessTerms {
		result = append(result, s.mapper.ToProtoBusinessTerm(businessTerm))
	}

	return result, nil
}

func (s *MetadataService) GetColumnsForBusinessTerm(ctx context.Context, businessTermID uuid.UUID) ([]*pb.Column, error) {
	columns, err := s.queries.GetColumnsForBusinessTerm(ctx, businessTermID)
	if err != nil {
		return nil, fmt.Errorf("failed to get columns for business term: %w", err)
	}

	var result []*pb.Column
	for _, column := range columns {
		// Convert the detailed column info to a standard column for the mapper
		standardColumn := generated.ColumnsMetadata{
			ID:              column.ID,
			TableID:         column.TableID,
			ColumnName:      column.ColumnName,
			DisplayName:     column.DisplayName,
			Description:     column.Description,
			DataType:        column.DataType,
			MaxLength:       column.MaxLength,
			IsRequired:      column.IsRequired,
			IsPrimaryKey:    column.IsPrimaryKey,
			IsUnique:        column.IsUnique,
			DefaultValue:    column.DefaultValue,
			ValidationRules: column.ValidationRules,
			ExampleValues:   column.ExampleValues,
			IsActive:        column.IsActive,
			SupportsImport:  column.SupportsImport,
			ColumnOrder:     column.ColumnOrder,
			CreatedAt:       column.CreatedAt,
			UpdatedAt:       column.UpdatedAt,
			CreatedBy:       column.CreatedBy,
			UpdatedBy:       column.UpdatedBy,
		}
		result = append(result, s.mapper.ToProtoColumn(standardColumn))
	}

	return result, nil
}