package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	db "p9e.in/ugcl/identity/entity/db/generated"
	"p9e.in/ugcl/identity/entity/mappers"
	"p9e.in/ugcl/identity/entity/repository"
	pb "p9e.in/ugcl/identity/entity/api/v1"
)

// EntityRoleBindingService implements IEntityRoleBindingService
type EntityRoleBindingService struct {
	roleBindingRepo repository.IEntityRoleBindingRepository
	entityRepo      repository.IEntityRepository
}

// NewEntityRoleBindingService creates a new entity role binding service
func NewEntityRoleBindingService(
	roleBindingRepo repository.IEntityRoleBindingRepository,
	entityRepo repository.IEntityRepository,
) *EntityRoleBindingService {
	return &EntityRoleBindingService{
		roleBindingRepo: roleBindingRepo,
		entityRepo:      entityRepo,
	}
}

func (s *EntityRoleBindingService) CreateEntityRoleBinding(ctx context.Context, req *pb.CreateEntityRoleBindingRequest) (*pb.EntityRoleBinding, error) {
	// Validate request
	if req.TenantId == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.EntityId == "" {
		return nil, fmt.Errorf("entity_id is required")
	}
	if req.RoleId == "" {
		return nil, fmt.Errorf("role_id is required")
	}

	// Verify entity exists
	entityID, _ := uuid.Parse(req.EntityId)
	tenantID, _ := uuid.Parse(req.TenantId)

	existing, err := s.entityRepo.GetByID(ctx, db.GetEntityParams{
		ID:       entityID,
		TenantID: tenantID,
	})
	if err != nil || existing == nil {
		return nil, fmt.Errorf("entity not found: %s", req.EntityId)
	}

	// TODO: Get current user ID from context for audit
	currentUserID := uuid.New() // Placeholder

	// Convert request to DB params
	params, err := mappers.CreateEntityRoleBindingRequestToDBParams(req, currentUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to map request: %w", err)
	}

	// Create role binding
	dbBinding, err := s.roleBindingRepo.Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create role binding: %w", err)
	}

	// Convert to proto
	binding := mappers.EntityRoleBindingDBToProto(dbBinding)

	return binding, nil
}

func (s *EntityRoleBindingService) GetEntityRoleBindings(ctx context.Context, req *pb.GetEntityRoleBindingsRequest) (*pb.GetEntityRoleBindingsResponse, error) {
	entityID, err := uuid.Parse(req.EntityId)
	if err != nil {
		return nil, fmt.Errorf("invalid entity id: %w", err)
	}

	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant id: %w", err)
	}

	dbBindings, err := s.roleBindingRepo.GetByEntityID(ctx, db.GetEntityRoleBindingsParams{
		EntityID: entityID,
		TenantID: tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get role bindings: %w", err)
	}

	bindings := make([]*pb.EntityRoleBinding, 0, len(dbBindings))
	for i := range dbBindings {
		binding := mappers.EntityRoleBindingDBToProto(&dbBindings[i])
		bindings = append(bindings, binding)
	}

	return &pb.GetEntityRoleBindingsResponse{
		Bindings: bindings,
	}, nil
}

func (s *EntityRoleBindingService) DeleteEntityRoleBinding(ctx context.Context, req *pb.DeleteEntityRoleBindingRequest) (*pb.DeleteEntityRoleBindingResponse, error) {
	bindingID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid binding id: %w", err)
	}

	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant id: %w", err)
	}

	err = s.roleBindingRepo.Delete(ctx, db.DeleteEntityRoleBindingParams{
		ID:       bindingID,
		TenantID: tenantID,
	})
	if err != nil {
		return &pb.DeleteEntityRoleBindingResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to delete role binding: %v", err),
		}, nil
	}

	return &pb.DeleteEntityRoleBindingResponse{
		Success: true,
		Message: "Role binding deleted successfully",
	}, nil
}
