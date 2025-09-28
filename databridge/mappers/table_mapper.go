// =============================================================================
// mappers/table_mapper.go - Table entity mapping implementation
// =============================================================================
package mappers

import (
	pb "p9e.in/ugcl/databridge/api/databridge"
	db "p9e.in/ugcl/databridge/db/generated"
	"p9e.in/ugcl/databridge/repository"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TableMapper implements ITableMapper
type TableMapper struct{}

// NewTableMapper creates a new table mapper instance
func NewTableMapper() ITableMapper {
	return &TableMapper{}
}

// DBToProto converts database model to protobuf (for GetTablesBySchema result)
func (m *TableMapper) DBToProto(table *db.GetTablesBySchemaRow) *pb.Table {
	if table == nil {
		return nil
	}

	pbTable := &pb.Table{
		Id:                  table.ID.String(),
		SchemaId:            table.SchemaID.String(),
		TableName:           table.TableName,
		DisplayName:         table.DisplayName,
		IsActive:            table.IsActive,
		SupportsImport:      table.SupportsImport,
		CreatedAt:           timestamppb.New(table.CreatedAt.Time),
		UpdatedAt:           timestamppb.New(table.UpdatedAt.Time),
		CreatedBy:           table.CreatedBy,
		SchemaName:          table.SchemaName,
		SchemaDisplayName:   table.SchemaDisplayName,
	}

	if table.Description.Valid {
		pbTable.Description = table.Description.String
	}

	if table.UpdatedBy.Valid {
		pbTable.UpdatedBy = table.UpdatedBy.String
	}

	return pbTable
}

// SingleDBToProto converts single table database model to protobuf
func (m *TableMapper) SingleDBToProto(table *db.TablesMetadatum) *pb.Table {
	if table == nil {
		return nil
	}

	pbTable := &pb.Table{
		Id:             table.ID.String(),
		SchemaId:       table.SchemaID.String(),
		TableName:      table.TableName,
		DisplayName:    table.DisplayName,
		IsActive:       table.IsActive,
		SupportsImport: table.SupportsImport,
		CreatedAt:      timestamppb.New(table.CreatedAt.Time),
		UpdatedAt:      timestamppb.New(table.UpdatedAt.Time),
		CreatedBy:      table.CreatedBy,
	}

	if table.Description.Valid {
		pbTable.Description = table.Description.String
	}

	if table.UpdatedBy.Valid {
		pbTable.UpdatedBy = table.UpdatedBy.String
	}

	return pbTable
}

// ImportableDBToProto converts importable tables result to protobuf
func (m *TableMapper) ImportableDBToProto(table *db.GetAllImportableTablesRow) *pb.Table {
	if table == nil {
		return nil
	}

	pbTable := &pb.Table{
		Id:                  table.ID.String(),
		SchemaId:            table.SchemaID.String(),
		TableName:           table.TableName,
		DisplayName:         table.DisplayName,
		IsActive:            table.IsActive,
		SupportsImport:      table.SupportsImport,
		CreatedAt:           timestamppb.New(table.CreatedAt.Time),
		UpdatedAt:           timestamppb.New(table.UpdatedAt.Time),
		CreatedBy:           table.CreatedBy,
		SchemaName:          table.SchemaName,
		SchemaDisplayName:   table.SchemaDisplayName,
	}

	if table.Description.Valid {
		pbTable.Description = table.Description.String
	}

	if table.UpdatedBy.Valid {
		pbTable.UpdatedBy = table.UpdatedBy.String
	}

	return pbTable
}

// ProtoToDB converts protobuf to database model (mainly for reference)
func (m *TableMapper) ProtoToDB(table *pb.Table) *db.TablesMetadatum {
	if table == nil {
		return nil
	}

	// This is mainly for reference - typically we use repository parameter structs
	return &db.TablesMetadatum{
		TableName:      table.TableName,
		DisplayName:    table.DisplayName,
		IsActive:       table.IsActive,
		SupportsImport: table.SupportsImport,
		CreatedBy:      table.CreatedBy,
	}
}

// CreateRequestToRepoParams converts create request to repository parameters
func (m *TableMapper) CreateRequestToRepoParams(req *pb.CreateTableRequest, schemaID uuid.UUID) *repository.CreateTableParams {
	if req == nil {
		return nil
	}

	params := &repository.CreateTableParams{
		SchemaID:       schemaID,
		TableName:      req.TableName,
		DisplayName:    req.DisplayName,
		SupportsImport: req.SupportsImport,
		CreatedBy:      req.CreatedBy,
	}

	if req.Description != "" {
		params.Description = &req.Description
	}

	return params
}

// UpdateRequestToRepoParams converts update request to repository parameters
func (m *TableMapper) UpdateRequestToRepoParams(req *pb.UpdateTableRequest, tableID uuid.UUID) *repository.UpdateTableParams {
	if req == nil {
		return nil
	}

	params := &repository.UpdateTableParams{
		ID:             tableID,
		DisplayName:    req.DisplayName,
		SupportsImport: req.SupportsImport,
		UpdatedBy:      req.UpdatedBy,
	}

	if req.Description != "" {
		params.Description = &req.Description
	}

	return params
}