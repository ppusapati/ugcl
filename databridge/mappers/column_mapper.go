// =============================================================================
// mappers/column_mapper.go - Column entity mapping implementation
// =============================================================================
package mappers

import (
	"encoding/json"

	pb "p9e.in/ugcl/databridge/api/databridge"
	db "p9e.in/ugcl/databridge/db/generated"
	"p9e.in/ugcl/databridge/repository"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ColumnMapper implements IColumnMapper
type ColumnMapper struct{}

// NewColumnMapper creates a new column mapper instance
func NewColumnMapper() IColumnMapper {
	return &ColumnMapper{}
}

// DBToProto converts database model to protobuf
func (m *ColumnMapper) DBToProto(column *db.ColumnsMetadatum) *pb.Column {
	if column == nil {
		return nil
	}

	pbColumn := &pb.Column{
		Id:             column.ID.String(),
		TableId:        column.TableID.String(),
		ColumnName:     column.ColumnName,
		DisplayName:    column.DisplayName,
		DataType:       m.DataTypeDBToProto(column.DataType),
		IsRequired:     column.IsRequired,
		IsPrimaryKey:   column.IsPrimaryKey,
		IsUnique:       column.IsUnique,
		IsActive:       column.IsActive,
		SupportsImport: column.SupportsImport,
		ColumnOrder:    column.ColumnOrder,
		CreatedAt:      timestamppb.New(column.CreatedAt.Time),
		UpdatedAt:      timestamppb.New(column.UpdatedAt.Time),
		CreatedBy:      column.CreatedBy,
	}

	if column.Description.Valid {
		pbColumn.Description = column.Description.String
	}

	if column.MaxLength.Valid {
		maxLength := column.MaxLength.Int32
		pbColumn.MaxLength = &maxLength
	}

	if column.DefaultValue.Valid {
		pbColumn.DefaultValue = &column.DefaultValue.String
	}

	if column.UpdatedBy.Valid {
		pbColumn.UpdatedBy = column.UpdatedBy.String
	}

	// Parse validation rules from JSONB
	if len(column.ValidationRules) > 0 {
		var rules []*pb.ValidationRule
		if err := json.Unmarshal(column.ValidationRules, &rules); err == nil {
			pbColumn.ValidationRules = rules
		}
	}

	// Map example values
	if len(column.ExampleValues) > 0 {
		pbColumn.ExampleValues = column.ExampleValues
	}

	return pbColumn
}

// ProtoToDB converts protobuf to database model (mainly for reference)
func (m *ColumnMapper) ProtoToDB(column *pb.Column) *db.ColumnsMetadatum {
	if column == nil {
		return nil
	}

	// This is mainly for reference - typically we use repository parameter structs
	return &db.ColumnsMetadatum{
		ColumnName:     column.ColumnName,
		DisplayName:    column.DisplayName,
		DataType:       m.DataTypeProtoToDB(column.DataType),
		IsRequired:     column.IsRequired,
		IsPrimaryKey:   column.IsPrimaryKey,
		IsUnique:       column.IsUnique,
		IsActive:       column.IsActive,
		SupportsImport: column.SupportsImport,
		ColumnOrder:    column.ColumnOrder,
		CreatedBy:      column.CreatedBy,
	}
}

// CreateRequestToRepoParams converts create request to repository parameters
func (m *ColumnMapper) CreateRequestToRepoParams(req *pb.CreateColumnRequest, tableID uuid.UUID) *repository.CreateColumnParams {
	if req == nil {
		return nil
	}

	params := &repository.CreateColumnParams{
		TableID:        tableID,
		ColumnName:     req.ColumnName,
		DisplayName:    req.DisplayName,
		DataType:       m.DataTypeProtoToDB(req.DataType),
		IsRequired:     req.IsRequired,
		IsPrimaryKey:   req.IsPrimaryKey,
		IsUnique:       req.IsUnique,
		SupportsImport: req.SupportsImport,
		ColumnOrder:    req.ColumnOrder,
		CreatedBy:      req.CreatedBy,
	}

	if req.Description != "" {
		params.Description = &req.Description
	}

	if req.MaxLength != nil {
		params.MaxLength = req.MaxLength
	}

	if req.DefaultValue != nil {
		params.DefaultValue = req.DefaultValue
	}

	// Serialize validation rules to JSONB
	if len(req.ValidationRules) > 0 {
		if rulesJSON, err := json.Marshal(req.ValidationRules); err == nil {
			params.ValidationRules = rulesJSON
		}
	}

	if len(req.ExampleValues) > 0 {
		params.ExampleValues = req.ExampleValues
	}

	return params
}

// UpdateRequestToRepoParams converts update request to repository parameters
func (m *ColumnMapper) UpdateRequestToRepoParams(req *pb.UpdateColumnRequest, columnID uuid.UUID) *repository.UpdateColumnParams {
	if req == nil {
		return nil
	}

	params := &repository.UpdateColumnParams{
		ID:             columnID,
		DisplayName:    req.DisplayName,
		DataType:       m.DataTypeProtoToDB(req.DataType),
		IsRequired:     req.IsRequired,
		IsUnique:       req.IsUnique,
		SupportsImport: req.SupportsImport,
		ColumnOrder:    req.ColumnOrder,
		UpdatedBy:      req.UpdatedBy,
	}

	if req.Description != "" {
		params.Description = &req.Description
	}

	if req.MaxLength != nil {
		params.MaxLength = req.MaxLength
	}

	if req.DefaultValue != nil {
		params.DefaultValue = req.DefaultValue
	}

	// Serialize validation rules to JSONB
	if len(req.ValidationRules) > 0 {
		if rulesJSON, err := json.Marshal(req.ValidationRules); err == nil {
			params.ValidationRules = rulesJSON
		}
	}

	if len(req.ExampleValues) > 0 {
		params.ExampleValues = req.ExampleValues
	}

	return params
}

// DataTypeDBToProto converts database data type to protobuf enum
func (m *ColumnMapper) DataTypeDBToProto(dataType string) pb.DataType {
	switch dataType {
	case "TEXT":
		return pb.DataType_DATA_TYPE_TEXT
	case "VARCHAR":
		return pb.DataType_DATA_TYPE_VARCHAR
	case "CHAR":
		return pb.DataType_DATA_TYPE_CHAR
	case "INTEGER":
		return pb.DataType_DATA_TYPE_INTEGER
	case "BIGINT":
		return pb.DataType_DATA_TYPE_BIGINT
	case "SMALLINT":
		return pb.DataType_DATA_TYPE_SMALLINT
	case "DECIMAL":
		return pb.DataType_DATA_TYPE_DECIMAL
	case "NUMERIC":
		return pb.DataType_DATA_TYPE_NUMERIC
	case "REAL":
		return pb.DataType_DATA_TYPE_REAL
	case "DOUBLE_PRECISION":
		return pb.DataType_DATA_TYPE_DOUBLE_PRECISION
	case "BOOLEAN":
		return pb.DataType_DATA_TYPE_BOOLEAN
	case "DATE":
		return pb.DataType_DATA_TYPE_DATE
	case "TIME":
		return pb.DataType_DATA_TYPE_TIME
	case "TIMESTAMP":
		return pb.DataType_DATA_TYPE_TIMESTAMP
	case "TIMESTAMPTZ":
		return pb.DataType_DATA_TYPE_TIMESTAMPTZ
	case "UUID":
		return pb.DataType_DATA_TYPE_UUID
	case "JSONB":
		return pb.DataType_DATA_TYPE_JSONB
	case "JSON":
		return pb.DataType_DATA_TYPE_JSON
	default:
		return pb.DataType_DATA_TYPE_UNSPECIFIED
	}
}

// DataTypeProtoToDB converts protobuf data type enum to database string
func (m *ColumnMapper) DataTypeProtoToDB(dataType pb.DataType) string {
	switch dataType {
	case pb.DataType_DATA_TYPE_TEXT:
		return "TEXT"
	case pb.DataType_DATA_TYPE_VARCHAR:
		return "VARCHAR"
	case pb.DataType_DATA_TYPE_CHAR:
		return "CHAR"
	case pb.DataType_DATA_TYPE_INTEGER:
		return "INTEGER"
	case pb.DataType_DATA_TYPE_BIGINT:
		return "BIGINT"
	case pb.DataType_DATA_TYPE_SMALLINT:
		return "SMALLINT"
	case pb.DataType_DATA_TYPE_DECIMAL:
		return "DECIMAL"
	case pb.DataType_DATA_TYPE_NUMERIC:
		return "NUMERIC"
	case pb.DataType_DATA_TYPE_REAL:
		return "REAL"
	case pb.DataType_DATA_TYPE_DOUBLE_PRECISION:
		return "DOUBLE_PRECISION"
	case pb.DataType_DATA_TYPE_BOOLEAN:
		return "BOOLEAN"
	case pb.DataType_DATA_TYPE_DATE:
		return "DATE"
	case pb.DataType_DATA_TYPE_TIME:
		return "TIME"
	case pb.DataType_DATA_TYPE_TIMESTAMP:
		return "TIMESTAMP"
	case pb.DataType_DATA_TYPE_TIMESTAMPTZ:
		return "TIMESTAMPTZ"
	case pb.DataType_DATA_TYPE_UUID:
		return "UUID"
	case pb.DataType_DATA_TYPE_JSONB:
		return "JSONB"
	case pb.DataType_DATA_TYPE_JSON:
		return "JSON"
	default:
		return "TEXT" // Default fallback
	}
}