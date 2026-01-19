// =============================================================================
// mappers/mapping_mapper.go - Import mapping entity mapping implementation
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

// MappingMapper implements IMappingMapper
type MappingMapper struct{}

// NewMappingMapper creates a new mapping mapper instance
func NewMappingMapper() IMappingMapper {
	return &MappingMapper{}
}

// DBToProto converts database model to protobuf
func (m *MappingMapper) DBToProto(mapping *db.ImportMapping) *pb.ImportMapping {
	if mapping == nil {
		return nil
	}

	pbMapping := &pb.ImportMapping{
		Id:          mapping.ID.String(),
		MappingName: mapping.MappingName,
		TableId:     mapping.TableID.String(),
		IsActive:    mapping.IsActive,
		CreatedAt:   timestamppb.New(mapping.CreatedAt.Time),
		UpdatedAt:   timestamppb.New(mapping.UpdatedAt.Time),
		CreatedBy:   mapping.CreatedBy,
		UsageCount:  mapping.UsageCount,
	}

	if mapping.Description.Valid {
		pbMapping.Description = mapping.Description.String
	}

	if mapping.UpdatedBy.Valid {
		pbMapping.UpdatedBy = mapping.UpdatedBy.String
	}

	if mapping.LastUsedAt.Valid {
		pbMapping.LastUsedAt = timestamppb.New(mapping.LastUsedAt.Time)
	}

	// Parse CSV headers from JSONB
	if len(mapping.CsvHeaders) > 0 {
		var headers []string
		if err := json.Unmarshal(mapping.CsvHeaders, &headers); err == nil {
			pbMapping.CsvHeaders = headers
		}
	}

	// Parse field mappings from JSONB
	if len(mapping.FieldMappings) > 0 {
		var fieldMappings []*pb.FieldMapping
		if err := json.Unmarshal(mapping.FieldMappings, &fieldMappings); err == nil {
			pbMapping.FieldMappings = fieldMappings
		}
	}

	// Parse transformation rules from JSONB
	if len(mapping.TransformationRules) > 0 {
		var transformationRules []*pb.ValidationRule
		if err := json.Unmarshal(mapping.TransformationRules, &transformationRules); err == nil {
			pbMapping.TransformationRules = transformationRules
		}
	}

	// Parse validation rules from JSONB
	if len(mapping.ValidationRules) > 0 {
		var validationRules []*pb.ValidationRule
		if err := json.Unmarshal(mapping.ValidationRules, &validationRules); err == nil {
			pbMapping.ValidationRules = validationRules
		}
	}

	return pbMapping
}

// DetailDBToProto converts detailed mapping query result to protobuf
func (m *MappingMapper) DetailDBToProto(mapping *db.GetMappingByIDRow) *pb.ImportMapping {
	if mapping == nil {
		return nil
	}

	pbMapping := &pb.ImportMapping{
		Id:                  mapping.ID.String(),
		MappingName:         mapping.MappingName,
		TableId:             mapping.TableID.String(),
		IsActive:            mapping.IsActive,
		CreatedAt:           timestamppb.New(mapping.CreatedAt.Time),
		UpdatedAt:           timestamppb.New(mapping.UpdatedAt.Time),
		CreatedBy:           mapping.CreatedBy,
		UsageCount:          mapping.UsageCount,
	}

	if mapping.Description.Valid {
		pbMapping.Description = mapping.Description.String
	}

	if mapping.UpdatedBy.Valid {
		pbMapping.UpdatedBy = mapping.UpdatedBy.String
	}

	if mapping.LastUsedAt.Valid {
		pbMapping.LastUsedAt = timestamppb.New(mapping.LastUsedAt.Time)
	}

	if mapping.TableName.Valid {
		pbMapping.TableName = &mapping.TableName.String
	}

	if mapping.TableDisplayName.Valid {
		pbMapping.TableDisplayName = &mapping.TableDisplayName.String
	}

	if mapping.SchemaName.Valid {
		pbMapping.SchemaName = &mapping.SchemaName.String
	}

	if mapping.SchemaDisplayName.Valid {
		pbMapping.SchemaDisplayName = &mapping.SchemaDisplayName.String
	}

	// Parse JSON fields same as above
	if len(mapping.CsvHeaders) > 0 {
		var headers []string
		if err := json.Unmarshal(mapping.CsvHeaders, &headers); err == nil {
			pbMapping.CsvHeaders = headers
		}
	}

	if len(mapping.FieldMappings) > 0 {
		var fieldMappings []*pb.FieldMapping
		if err := json.Unmarshal(mapping.FieldMappings, &fieldMappings); err == nil {
			pbMapping.FieldMappings = fieldMappings
		}
	}

	if len(mapping.TransformationRules) > 0 {
		var transformationRules []*pb.ValidationRule
		if err := json.Unmarshal(mapping.TransformationRules, &transformationRules); err == nil {
			pbMapping.TransformationRules = transformationRules
		}
	}

	if len(mapping.ValidationRules) > 0 {
		var validationRules []*pb.ValidationRule
		if err := json.Unmarshal(mapping.ValidationRules, &validationRules); err == nil {
			pbMapping.ValidationRules = validationRules
		}
	}

	return pbMapping
}

// ProtoToDB converts protobuf to database model (mainly for reference)
func (m *MappingMapper) ProtoToDB(mapping *pb.ImportMapping) *db.ImportMapping {
	if mapping == nil {
		return nil
	}

	// This is mainly for reference - typically we use repository parameter structs
	return &db.ImportMapping{
		MappingName: mapping.MappingName,
		IsActive:    mapping.IsActive,
		CreatedBy:   mapping.CreatedBy,
		UsageCount:  mapping.UsageCount,
	}
}

// CreateRequestToRepoParams converts create request to repository parameters
func (m *MappingMapper) CreateRequestToRepoParams(req *pb.CreateMappingRequest) *repository.CreateMappingParams {
	if req == nil {
		return nil
	}

	tableID, err := uuid.Parse(req.TableId)
	if err != nil {
		return nil
	}

	params := &repository.CreateMappingParams{
		MappingName: req.MappingName,
		TableID:     tableID,
		CreatedBy:   req.CreatedBy,
	}

	if req.Description != "" {
		params.Description = &req.Description
	}

	// Serialize CSV headers to JSONB
	if len(req.CsvHeaders) > 0 {
		if headersJSON, err := json.Marshal(req.CsvHeaders); err == nil {
			params.CSVHeaders = headersJSON
		}
	}

	// Serialize field mappings to JSONB
	if len(req.FieldMappings) > 0 {
		if mappingsJSON, err := json.Marshal(req.FieldMappings); err == nil {
			params.FieldMappings = mappingsJSON
		}
	}

	// Serialize transformation rules to JSONB
	if len(req.TransformationRules) > 0 {
		if rulesJSON, err := json.Marshal(req.TransformationRules); err == nil {
			params.TransformationRules = rulesJSON
		}
	}

	// Serialize validation rules to JSONB
	if len(req.ValidationRules) > 0 {
		if rulesJSON, err := json.Marshal(req.ValidationRules); err == nil {
			params.ValidationRules = rulesJSON
		}
	}

	return params
}

// UpdateRequestToRepoParams converts update request to repository parameters
func (m *MappingMapper) UpdateRequestToRepoParams(req *pb.UpdateMappingRequest, mappingID uuid.UUID) *repository.UpdateMappingParams {
	if req == nil {
		return nil
	}

	params := &repository.UpdateMappingParams{
		ID:          mappingID,
		MappingName: req.MappingName,
		UpdatedBy:   req.UpdatedBy,
	}

	if req.Description != "" {
		params.Description = &req.Description
	}

	// Serialize CSV headers to JSONB
	if len(req.CsvHeaders) > 0 {
		if headersJSON, err := json.Marshal(req.CsvHeaders); err == nil {
			params.CSVHeaders = headersJSON
		}
	}

	// Serialize field mappings to JSONB
	if len(req.FieldMappings) > 0 {
		if mappingsJSON, err := json.Marshal(req.FieldMappings); err == nil {
			params.FieldMappings = mappingsJSON
		}
	}

	// Serialize transformation rules to JSONB
	if len(req.TransformationRules) > 0 {
		if rulesJSON, err := json.Marshal(req.TransformationRules); err == nil {
			params.TransformationRules = rulesJSON
		}
	}

	// Serialize validation rules to JSONB
	if len(req.ValidationRules) > 0 {
		if rulesJSON, err := json.Marshal(req.ValidationRules); err == nil {
			params.ValidationRules = rulesJSON
		}
	}

	return params
}