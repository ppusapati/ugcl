// =============================================================================
// services/mapping_service.go - Import mapping service implementation
// =============================================================================
package services

import (
	"context"
	"fmt"

	pb "p9e.in/ugcl/databridge/api/databridge"
	"p9e.in/ugcl/databridge/mappers"
	"p9e.in/ugcl/databridge/repository"

	"github.com/google/uuid"
)

// MappingService implements IMappingService
type MappingService struct {
	repos   *repository.RepositoryManager
	mappers *mappers.MapperManager
}

// NewMappingService creates a new mapping service instance
func NewMappingService(repos *repository.RepositoryManager, mappers *mappers.MapperManager) IMappingService {
	return &MappingService{
		repos:   repos,
		mappers: mappers,
	}
}

// GetMappings retrieves import mappings with pagination and filtering
func (s *MappingService) GetMappings(ctx context.Context, req *pb.GetMappingsRequest) (*pb.GetMappingsResponse, error) {
	var mappings []*pb.ImportMapping
	totalCount := int32(0)

	if req.TableId != nil {
		// Get mappings for specific table
		tableID, err := uuid.Parse(*req.TableId)
		if err != nil {
			return nil, fmt.Errorf("invalid table ID: %w", err)
		}

		dbMappings, err := s.repos.Mapping.GetMappingsByTable(ctx, tableID)
		if err != nil {
			return nil, fmt.Errorf("failed to get mappings for table: %w", err)
		}

		for _, mapping := range dbMappings {
			mappings = append(mappings, s.mappers.Mapping.DBToProto(mapping))
		}
		totalCount = int32(len(mappings))

	} else if req.UserId != nil {
		// Get recent mappings for user
		limit := req.Limit
		if limit <= 0 {
			limit = 50 // Default limit
		}

		dbMappings, err := s.repos.Mapping.GetRecentMappingsByUser(ctx, *req.UserId, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to get mappings for user: %w", err)
		}

		for _, mapping := range dbMappings {
			pbMapping := &pb.ImportMapping{
				Id:                  mapping.ID.String(),
				MappingName:         mapping.MappingName,
				TableId:             mapping.TableID.String(),
				IsActive:            mapping.IsActive,
				CreatedBy:           mapping.CreatedBy,
				UsageCount:          mapping.UsageCount,
			}

			if mapping.Description.Valid {
				pbMapping.Description = mapping.Description.String
			}

			if mapping.UpdatedBy.Valid {
				pbMapping.UpdatedBy = mapping.UpdatedBy.String
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

			mappings = append(mappings, pbMapping)
		}
		totalCount = int32(len(mappings))
	}

	return &pb.GetMappingsResponse{
		Mappings:   mappings,
		TotalCount: totalCount,
	}, nil
}

// GetMappingByID retrieves a specific mapping by ID
func (s *MappingService) GetMappingByID(ctx context.Context, mappingID uuid.UUID) (*pb.ImportMapping, error) {
	mapping, err := s.repos.Mapping.GetMappingByID(ctx, mappingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mapping by ID: %w", err)
	}

	return s.mappers.Mapping.DetailDBToProto(mapping), nil
}

// CreateMapping creates a new import mapping
func (s *MappingService) CreateMapping(ctx context.Context, req *pb.CreateMappingRequest) (*pb.ImportMapping, error) {
	// Validate request
	if err := s.validateCreateMappingRequest(req); err != nil {
		return nil, err
	}

	tableID, err := uuid.Parse(req.TableId)
	if err != nil {
		return nil, fmt.Errorf("invalid table ID: %w", err)
	}

	// Check if table exists
	table, err := s.repos.Table.GetTableByID(ctx, tableID)
	if err != nil || table == nil {
		return nil, fmt.Errorf("table not found")
	}

	// Check if mapping name already exists for this table
	existing, err := s.repos.Mapping.GetMappingByTableAndName(ctx, tableID, req.MappingName)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("mapping with name '%s' already exists for this table", req.MappingName)
	}

	// Create mapping
	params := s.mappers.Mapping.CreateRequestToRepoParams(req)
	if params == nil {
		return nil, fmt.Errorf("failed to convert request to repository parameters")
	}

	mapping, err := s.repos.Mapping.CreateMapping(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create mapping: %w", err)
	}

	return s.mappers.Mapping.DBToProto(mapping), nil
}

// UpdateMapping updates an existing import mapping
func (s *MappingService) UpdateMapping(ctx context.Context, req *pb.UpdateMappingRequest) (*pb.ImportMapping, error) {
	// Validate request
	if err := s.validateUpdateMappingRequest(req); err != nil {
		return nil, err
	}

	mappingID, err := uuid.Parse(req.MappingId)
	if err != nil {
		return nil, fmt.Errorf("invalid mapping ID: %w", err)
	}

	// Check if mapping exists
	existing, err := s.repos.Mapping.GetMappingByID(ctx, mappingID)
	if err != nil || existing == nil {
		return nil, fmt.Errorf("mapping not found")
	}

	// Update mapping
	params := s.mappers.Mapping.UpdateRequestToRepoParams(req, mappingID)
	if params == nil {
		return nil, fmt.Errorf("failed to convert request to repository parameters")
	}

	mapping, err := s.repos.Mapping.UpdateMapping(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update mapping: %w", err)
	}

	return s.mappers.Mapping.DBToProto(mapping), nil
}

// DeleteMapping deactivates an import mapping
func (s *MappingService) DeleteMapping(ctx context.Context, mappingID uuid.UUID, userID string) error {
	// Check if mapping exists
	existing, err := s.repos.Mapping.GetMappingByID(ctx, mappingID)
	if err != nil || existing == nil {
		return fmt.Errorf("mapping not found")
	}

	// Check for active import jobs using this mapping
	activeJobs, err := s.repos.Job.GetActiveJobs(ctx)
	if err != nil {
		return fmt.Errorf("failed to check active jobs: %w", err)
	}

	for _, job := range activeJobs {
		if job.MappingID.String() == mappingID.String() {
			return fmt.Errorf("cannot delete mapping: active import job exists (ID: %s)", job.ID.String())
		}
	}

	// Deactivate mapping
	err = s.repos.Mapping.DeactivateMapping(ctx, mappingID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete mapping: %w", err)
	}

	return nil
}

// ValidateMapping validates CSV data against an import mapping
func (s *MappingService) ValidateMapping(ctx context.Context, req *pb.ValidateMappingRequest) (*pb.ValidateMappingResponse, error) {
	mappingID, err := uuid.Parse(req.MappingId)
	if err != nil {
		return nil, fmt.Errorf("invalid mapping ID: %w", err)
	}

	// Get mapping details
	mapping, err := s.repos.Mapping.GetMappingByID(ctx, mappingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mapping: %w", err)
	}

	// Convert to protobuf mapping
	pbMapping := s.mappers.Mapping.DetailDBToProto(mapping)

	// Use CSV service to validate (we'll create a placeholder validation here)
	// In the full implementation, this would use the CSV service
	response := &pb.ValidateMappingResponse{
		IsValid:      true,
		Errors:       []*pb.ImportError{},
		Warnings:     []string{},
		TotalRows:    0,
		ValidRows:    0,
		InvalidRows:  0,
	}

	// Basic validation - check if CSV headers match mapping headers
	if len(pbMapping.CsvHeaders) == 0 {
		response.IsValid = false
		response.Errors = append(response.Errors, &pb.ImportError{
			RowNumber:    0,
			ErrorMessage: "No CSV headers defined in mapping",
			ErrorType:    "configuration_error",
		})
	}

	return response, nil
}

// Validation helper methods

func (s *MappingService) validateCreateMappingRequest(req *pb.CreateMappingRequest) error {
	if req.MappingName == "" {
		return fmt.Errorf("mapping name is required")
	}
	if req.TableId == "" {
		return fmt.Errorf("table ID is required")
	}
	if req.CreatedBy == "" {
		return fmt.Errorf("created by is required")
	}
	if len(req.FieldMappings) == 0 {
		return fmt.Errorf("at least one field mapping is required")
	}
	return nil
}

func (s *MappingService) validateUpdateMappingRequest(req *pb.UpdateMappingRequest) error {
	if req.MappingId == "" {
		return fmt.Errorf("mapping ID is required")
	}
	if req.MappingName == "" {
		return fmt.Errorf("mapping name is required")
	}
	if req.UpdatedBy == "" {
		return fmt.Errorf("updated by is required")
	}
	if len(req.FieldMappings) == 0 {
		return fmt.Errorf("at least one field mapping is required")
	}
	return nil
}