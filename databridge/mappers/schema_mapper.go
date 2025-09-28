// =============================================================================
// mappers/schema_mapper.go - Schema entity mapping implementation
// =============================================================================
package mappers

import (
	pb "p9e.in/ugcl/databridge/api/databridge"
	db "p9e.in/ugcl/databridge/db/generated"
	"p9e.in/ugcl/databridge/repository"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SchemaMapper implements ISchemaMapper
type SchemaMapper struct{}

// NewSchemaMapper creates a new schema mapper instance
func NewSchemaMapper() ISchemaMapper {
	return &SchemaMapper{}
}

// DBToProto converts database model to protobuf
func (m *SchemaMapper) DBToProto(schema *db.SchemasMetadatum) *pb.Schema {
	if schema == nil {
		return nil
	}

	pbSchema := &pb.Schema{
		Id:          schema.ID.String(),
		SchemaName:  schema.SchemaName,
		DisplayName: schema.DisplayName,
		IsActive:    schema.IsActive,
		CreatedAt:   timestamppb.New(schema.CreatedAt.Time),
		UpdatedAt:   timestamppb.New(schema.UpdatedAt.Time),
		CreatedBy:   schema.CreatedBy,
	}

	if schema.Description.Valid {
		pbSchema.Description = schema.Description.String
	}

	if schema.UpdatedBy.Valid {
		pbSchema.UpdatedBy = schema.UpdatedBy.String
	}

	return pbSchema
}

// ProtoToDB converts protobuf to database model
func (m *SchemaMapper) ProtoToDB(schema *pb.Schema) *db.SchemasMetadatum {
	if schema == nil {
		return nil
	}

	// This is mainly for reference - typically we don't convert from proto to DB model directly
	// Instead we use repository parameter structs
	return &db.SchemasMetadatum{
		SchemaName:  schema.SchemaName,
		DisplayName: schema.DisplayName,
		IsActive:    schema.IsActive,
		CreatedBy:   schema.CreatedBy,
	}
}

// CreateRequestToRepoParams converts create request to repository parameters
func (m *SchemaMapper) CreateRequestToRepoParams(req *pb.CreateSchemaRequest) *repository.CreateSchemaParams {
	if req == nil {
		return nil
	}

	params := &repository.CreateSchemaParams{
		SchemaName:  req.SchemaName,
		DisplayName: req.DisplayName,
		CreatedBy:   req.CreatedBy,
	}

	if req.Description != "" {
		params.Description = &req.Description
	}

	return params
}

// UpdateRequestToRepoParams converts update request to repository parameters
func (m *SchemaMapper) UpdateRequestToRepoParams(req *pb.UpdateSchemaRequest, schemaID uuid.UUID) *repository.UpdateSchemaParams {
	if req == nil {
		return nil
	}

	params := &repository.UpdateSchemaParams{
		ID:          schemaID,
		DisplayName: req.DisplayName,
		UpdatedBy:   req.UpdatedBy,
	}

	if req.Description != "" {
		params.Description = &req.Description
	}

	return params
}