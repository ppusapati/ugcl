package mappers

import (
	"database/sql"

	"google.golang.org/protobuf/types/known/timestamppb"

	pb "p9e.in/ugcl/masters/api/proto"
	"p9e.in/ugcl/masters/db/generated"
)

// MetadataMapper handles conversions between database models and protobuf models
type MetadataMapper struct{}

// NewMetadataMapper creates a new metadata mapper instance
func NewMetadataMapper() *MetadataMapper {
	return &MetadataMapper{}
}

// =============================================================================
// Schema Mappings
// =============================================================================

func (m *MetadataMapper) ToProtoSchema(schema generated.SchemasMetadata) *pb.Schema {
	return &pb.Schema{
		Id:          schema.ID.String(),
		SchemaName:  schema.SchemaName,
		DisplayName: schema.DisplayName,
		Description: schema.Description.String,
		IsActive:    schema.IsActive,
		CreatedAt:   timestamppb.New(schema.CreatedAt),
		UpdatedAt:   timestamppb.New(schema.UpdatedAt),
		CreatedBy:   schema.CreatedBy,
		UpdatedBy:   schema.UpdatedBy.String,
	}
}

// =============================================================================
// Table Mappings
// =============================================================================

func (m *MetadataMapper) ToProtoTable(table generated.GetTablesBySchemaRow) *pb.Table {
	return &pb.Table{
		Id:                  table.ID.String(),
		SchemaId:            table.SchemaID.String(),
		TableName:           table.TableName,
		DisplayName:         table.DisplayName,
		Description:         table.Description.String,
		IsActive:            table.IsActive,
		SupportsImport:      table.SupportsImport,
		CreatedAt:           timestamppb.New(table.CreatedAt),
		UpdatedAt:           timestamppb.New(table.UpdatedAt),
		CreatedBy:           table.CreatedBy,
		UpdatedBy:           table.UpdatedBy.String,
		SchemaName:          table.SchemaName.String,
		SchemaDisplayName:   table.SchemaDisplayName.String,
	}
}

// =============================================================================
// Column Mappings
// =============================================================================

func (m *MetadataMapper) ToProtoColumn(column generated.ColumnsMetadata) *pb.Column {
	return &pb.Column{
		Id:              column.ID.String(),
		TableId:         column.TableID.String(),
		ColumnName:      column.ColumnName,
		DisplayName:     column.DisplayName,
		Description:     column.Description.String,
		DataType:        m.StringToDataType(column.DataType),
		MaxLength:       column.MaxLength.Int32,
		IsRequired:      column.IsRequired,
		IsPrimaryKey:    column.IsPrimaryKey,
		IsUnique:        column.IsUnique,
		DefaultValue:    column.DefaultValue.String,
		ValidationRules: string(column.ValidationRules),
		ExampleValues:   column.ExampleValues,
		IsActive:        column.IsActive,
		SupportsImport:  column.SupportsImport,
		ColumnOrder:     column.ColumnOrder,
		CreatedAt:       timestamppb.New(column.CreatedAt),
		UpdatedAt:       timestamppb.New(column.UpdatedAt),
		CreatedBy:       column.CreatedBy,
		UpdatedBy:       column.UpdatedBy.String,
	}
}

// =============================================================================
// Relationship Mappings
// =============================================================================

func (m *MetadataMapper) ToProtoTableRelationship(relationship generated.TableRelationships) *pb.TableRelationship {
	return &pb.TableRelationship{
		Id:               relationship.ID.String(),
		Name:             relationship.Name.String,
		SourceTableId:    relationship.SourceTableID.String(),
		SourceColumnId:   relationship.SourceColumnID.String(),
		TargetTableId:    relationship.TargetTableID.String(),
		TargetColumnId:   relationship.TargetColumnID.String(),
		RelationshipType: m.StringToRelationshipType(relationship.RelationshipType),
		IsActive:         relationship.IsActive,
		CreatedAt:        timestamppb.New(relationship.CreatedAt),
		UpdatedAt:        timestamppb.New(relationship.UpdatedAt),
		CreatedBy:        relationship.CreatedBy,
		UpdatedBy:        relationship.UpdatedBy.String,
	}
}

func (m *MetadataMapper) ToProtoTableRelationshipWithDetails(relationship generated.ListRelationshipsByTableRow) *pb.TableRelationship {
	return &pb.TableRelationship{
		Id:                         relationship.ID.String(),
		Name:                       relationship.Name.String,
		SourceTableId:              relationship.SourceTableID.String(),
		SourceColumnId:             relationship.SourceColumnID.String(),
		TargetTableId:              relationship.TargetTableID.String(),
		TargetColumnId:             relationship.TargetColumnID.String(),
		RelationshipType:           m.StringToRelationshipType(relationship.RelationshipType),
		IsActive:                   relationship.IsActive,
		CreatedAt:                  timestamppb.New(relationship.CreatedAt),
		UpdatedAt:                  timestamppb.New(relationship.UpdatedAt),
		CreatedBy:                  relationship.CreatedBy,
		UpdatedBy:                  relationship.UpdatedBy.String,
		SourceTableName:            relationship.SourceTableName.String,
		SourceTableDisplayName:     relationship.SourceTableDisplayName.String,
		SourceColumnName:           relationship.SourceColumnName.String,
		SourceColumnDisplayName:    relationship.SourceColumnDisplayName.String,
		TargetTableName:            relationship.TargetTableName.String,
		TargetTableDisplayName:     relationship.TargetTableDisplayName.String,
		TargetColumnName:           relationship.TargetColumnName.String,
		TargetColumnDisplayName:    relationship.TargetColumnDisplayName.String,
	}
}

// =============================================================================
// Business Term Mappings
// =============================================================================

func (m *MetadataMapper) ToProtoBusinessTerm(businessTerm generated.BusinessTerms) *pb.BusinessTerm {
	var ownerUserID string
	if businessTerm.OwnerUserID.Valid {
		ownerUserID = businessTerm.OwnerUserID.UUID.String()
	}

	return &pb.BusinessTerm{
		Id:              businessTerm.ID.String(),
		Term:            businessTerm.Term,
		Definition:      businessTerm.Definition,
		BusinessContext: businessTerm.BusinessContext.String,
		Domain:          businessTerm.Domain.String,
		OwnerUserId:     ownerUserID,
		IsActive:        businessTerm.IsActive,
		CreatedAt:       timestamppb.New(businessTerm.CreatedAt),
		UpdatedAt:       timestamppb.New(businessTerm.UpdatedAt),
		CreatedBy:       businessTerm.CreatedBy,
		UpdatedBy:       businessTerm.UpdatedBy.String,
	}
}

// =============================================================================
// Data Type Conversion Helpers
// =============================================================================

func (m *MetadataMapper) StringToDataType(dataType string) pb.DataType {
	switch dataType {
	case "TEXT":
		return pb.DataType_DATA_TYPE_TEXT
	case "VARCHAR":
		return pb.DataType_DATA_TYPE_VARCHAR
	case "INTEGER":
		return pb.DataType_DATA_TYPE_INTEGER
	case "DECIMAL", "NUMERIC":
		return pb.DataType_DATA_TYPE_DECIMAL
	case "BOOLEAN":
		return pb.DataType_DATA_TYPE_BOOLEAN
	case "DATE":
		return pb.DataType_DATA_TYPE_DATE
	case "TIMESTAMP", "TIMESTAMPTZ":
		return pb.DataType_DATA_TYPE_TIMESTAMP
	case "UUID":
		return pb.DataType_DATA_TYPE_UUID
	case "JSON", "JSONB":
		return pb.DataType_DATA_TYPE_JSON
	default:
		return pb.DataType_DATA_TYPE_UNSPECIFIED
	}
}

func (m *MetadataMapper) DataTypeToString(dataType pb.DataType) string {
	switch dataType {
	case pb.DataType_DATA_TYPE_TEXT:
		return "TEXT"
	case pb.DataType_DATA_TYPE_VARCHAR:
		return "VARCHAR"
	case pb.DataType_DATA_TYPE_INTEGER:
		return "INTEGER"
	case pb.DataType_DATA_TYPE_DECIMAL:
		return "DECIMAL"
	case pb.DataType_DATA_TYPE_BOOLEAN:
		return "BOOLEAN"
	case pb.DataType_DATA_TYPE_DATE:
		return "DATE"
	case pb.DataType_DATA_TYPE_TIMESTAMP:
		return "TIMESTAMP"
	case pb.DataType_DATA_TYPE_UUID:
		return "UUID"
	case pb.DataType_DATA_TYPE_JSON:
		return "JSONB"
	default:
		return "TEXT"
	}
}

// =============================================================================
// Relationship Type Conversion Helpers
// =============================================================================

func (m *MetadataMapper) StringToRelationshipType(relationshipType string) pb.RelationshipType {
	switch relationshipType {
	case "one_to_one":
		return pb.RelationshipType_RELATIONSHIP_TYPE_ONE_TO_ONE
	case "one_to_many":
		return pb.RelationshipType_RELATIONSHIP_TYPE_ONE_TO_MANY
	case "many_to_one":
		return pb.RelationshipType_RELATIONSHIP_TYPE_MANY_TO_ONE
	case "many_to_many":
		return pb.RelationshipType_RELATIONSHIP_TYPE_MANY_TO_MANY
	default:
		return pb.RelationshipType_RELATIONSHIP_TYPE_UNSPECIFIED
	}
}

func (m *MetadataMapper) RelationshipTypeToString(relationshipType pb.RelationshipType) string {
	switch relationshipType {
	case pb.RelationshipType_RELATIONSHIP_TYPE_ONE_TO_ONE:
		return "one_to_one"
	case pb.RelationshipType_RELATIONSHIP_TYPE_ONE_TO_MANY:
		return "one_to_many"
	case pb.RelationshipType_RELATIONSHIP_TYPE_MANY_TO_ONE:
		return "many_to_one"
	case pb.RelationshipType_RELATIONSHIP_TYPE_MANY_TO_MANY:
		return "many_to_many"
	default:
		return "one_to_many"
	}
}