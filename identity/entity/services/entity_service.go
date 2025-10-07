package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	pb "p9e.in/ugcl/identity/entity/api/v1"
	db "p9e.in/ugcl/identity/entity/db/generated"
	"p9e.in/ugcl/identity/entity/mappers"
	"p9e.in/ugcl/identity/entity/repository"
)

// EntityService implements IEntityService
type EntityService struct {
	entityRepo repository.IEntityRepository
}

// NewEntityService creates a new entity service
func NewEntityService(entityRepo repository.IEntityRepository) *EntityService {
	return &EntityService{
		entityRepo: entityRepo,
	}
}

func (s *EntityService) CreateEntity(ctx context.Context, req *pb.CreateEntityRequest) (*pb.Entity, error) {
	// Validate request
	if req.TenantId == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.UserId == "" {
		return nil, fmt.Errorf("user_id is required")
	}
	if req.ReferenceId == "" {
		return nil, fmt.Errorf("reference_id is required")
	}

	// TODO: Get current user ID from context for audit
	currentUserID := uuid.New() // Placeholder - should come from auth context

	// Check if entity already exists for this user
	tenantID, _ := uuid.Parse(req.TenantId)
	userID, _ := uuid.Parse(req.UserId)
	existing, _ := s.entityRepo.GetByUserID(ctx, db.GetEntityByUserIdParams{
		UserID:   userID,
		TenantID: tenantID,
	})

	if existing != nil {
		return nil, fmt.Errorf("entity already exists for user_id: %s", req.UserId)
	}

	// Convert request to DB params
	params, err := mappers.CreateEntityRequestToDBParams(req, currentUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to map request: %w", err)
	}

	// Create entity
	dbEntity, err := s.entityRepo.Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create entity: %w", err)
	}

	// Convert to proto
	entity, err := mappers.EntityDBToProto(dbEntity)
	if err != nil {
		return nil, fmt.Errorf("failed to map entity: %w", err)
	}

	return entity, nil
}

func (s *EntityService) UpdateEntity(ctx context.Context, req *pb.UpdateEntityRequest) (*pb.Entity, error) {
	// Validate request
	if req.Id == "" {
		return nil, fmt.Errorf("id is required")
	}

	entityID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid entity id: %w", err)
	}

	// TODO: Get tenant_id and current user from context
	currentUserID := uuid.New()
	tenantID := uuid.New() // Should come from auth context

	// Check if entity exists
	existing, err := s.entityRepo.GetByID(ctx, db.GetEntityParams{
		ID:       entityID,
		TenantID: tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("entity not found: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("entity not found with id: %s", req.Id)
	}

	// Convert request to DB params
	params, err := mappers.UpdateEntityRequestToDBParams(req, currentUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to map request: %w", err)
	}
	params.TenantID = tenantID

	// Update entity
	dbEntity, err := s.entityRepo.Update(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update entity: %w", err)
	}

	// Convert to proto
	entity, err := mappers.EntityDBToProto(dbEntity)
	if err != nil {
		return nil, fmt.Errorf("failed to map entity: %w", err)
	}

	return entity, nil
}

func (s *EntityService) GetEntity(ctx context.Context, req *pb.GetEntityRequest) (*pb.Entity, error) {
	entityID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid entity id: %w", err)
	}

	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant id: %w", err)
	}

	dbEntity, err := s.entityRepo.GetByID(ctx, db.GetEntityParams{
		ID:       entityID,
		TenantID: tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get entity: %w", err)
	}

	entity, err := mappers.EntityDBToProto(dbEntity)
	if err != nil {
		return nil, fmt.Errorf("failed to map entity: %w", err)
	}

	return entity, nil
}

func (s *EntityService) GetEntityByUserId(ctx context.Context, req *pb.GetEntityByUserIdRequest) (*pb.Entity, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant id: %w", err)
	}

	dbEntity, err := s.entityRepo.GetByUserID(ctx, db.GetEntityByUserIdParams{
		UserID:   userID,
		TenantID: tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get entity: %w", err)
	}

	entity, err := mappers.EntityDBToProto(dbEntity)
	if err != nil {
		return nil, fmt.Errorf("failed to map entity: %w", err)
	}

	return entity, nil
}

func (s *EntityService) GetEntityByReference(ctx context.Context, req *pb.GetEntityByReferenceRequest) (*pb.Entity, error) {
	referenceID, err := uuid.Parse(req.ReferenceId)
	if err != nil {
		return nil, fmt.Errorf("invalid reference id: %w", err)
	}

	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant id: %w", err)
	}

	dbEntity, err := s.entityRepo.GetByReference(ctx, db.GetEntityByReferenceParams{
		ReferenceID:     referenceID,
		ReferenceSource: mappers.ReferenceSourceProtoToDB(req.ReferenceSource),
		TenantID:        tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get entity: %w", err)
	}

	entity, err := mappers.EntityDBToProto(dbEntity)
	if err != nil {
		return nil, fmt.Errorf("failed to map entity: %w", err)
	}

	return entity, nil
}

func (s *EntityService) ListEntities(ctx context.Context, req *pb.ListEntitiesRequest) (*pb.ListEntitiesResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant id: %w", err)
	}

	// Set defaults for pagination
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	pageNumber := req.PageNumber
	if pageNumber <= 0 {
		pageNumber = 1
	}

	offset := (pageNumber - 1) * pageSize

	// Build query params
	params := db.ListEntitiesParams{
		TenantID: tenantID,
		Limit:    pageSize,
		Offset:   offset,
	}

	// Optional filters
	if req.EntityType != nil {
		et := mappers.EntityTypeProtoToDB(pb.EntityType(req.EntityType.Value))
		params.EntityType = et
	}
	if req.Status != nil {
		st := mappers.EntityStatusProtoToDB(pb.EntityStatus(req.Status.Value))
		params.Status = st
	}
	if req.DivisionId != nil {
		divID, _ := uuid.Parse(req.DivisionId.Value)
		params.DivisionID = divID
	}
	if req.BranchId != nil {
		branchID, _ := uuid.Parse(req.BranchId.Value)
		params.BranchID = branchID
	}
	if req.DepartmentId != nil {
		deptID, _ := uuid.Parse(req.DepartmentId.Value)
		params.DepartmentID = deptID
	}

	// Get entities
	dbEntities, err := s.entityRepo.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list entities: %w", err)
	}

	// Get total count
	countParams := db.CountEntitiesParams{
		TenantID:     params.TenantID,
		EntityType:   params.EntityType,
		Status:       params.Status,
		DivisionID:   params.DivisionID,
		BranchID:     params.BranchID,
		DepartmentID: params.DepartmentID,
	}
	totalCount, err := s.entityRepo.Count(ctx, countParams)
	if err != nil {
		return nil, fmt.Errorf("failed to count entities: %w", err)
	}

	// Convert to proto
	entities := make([]*pb.Entity, 0, len(dbEntities))
	for i := range dbEntities {
		entity, err := mappers.EntityDBToProto(&dbEntities[i])
		if err != nil {
			return nil, fmt.Errorf("failed to map entity: %w", err)
		}
		entities = append(entities, entity)
	}

	return &pb.ListEntitiesResponse{
		Entities:   entities,
		TotalCount: int32(totalCount),
		PageNumber: pageNumber,
		PageSize:   pageSize,
	}, nil
}

func (s *EntityService) DeleteEntity(ctx context.Context, req *pb.DeleteEntityRequest) (*pb.DeleteEntityResponse, error) {
	entityID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid entity id: %w", err)
	}

	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant id: %w", err)
	}

	// Check if entity exists
	existing, err := s.entityRepo.GetByID(ctx, db.GetEntityParams{
		ID:       entityID,
		TenantID: tenantID,
	})
	if err != nil {
		return &pb.DeleteEntityResponse{
			Success: false,
			Message: fmt.Sprintf("Entity not found: %v", err),
		}, nil
	}
	if existing == nil {
		return &pb.DeleteEntityResponse{
			Success: false,
			Message: "Entity not found",
		}, nil
	}

	// Delete entity (cascades to role bindings)
	err = s.entityRepo.Delete(ctx, db.DeleteEntityParams{
		ID:       entityID,
		TenantID: tenantID,
	})
	if err != nil {
		return &pb.DeleteEntityResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to delete entity: %v", err),
		}, nil
	}

	return &pb.DeleteEntityResponse{
		Success: true,
		Message: "Entity deleted successfully",
	}, nil
}
