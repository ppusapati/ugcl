package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	tenantpb "p9e.in/ugcl/identity/tenant/api/v1/tenant"
	"p9e.in/ugcl/identity/tenant/mappers"
	"p9e.in/ugcl/identity/tenant/models"
	"p9e.in/ugcl/identity/tenant/repository"
	"p9e.in/ugcl/identity/tenant/uow"
)

// TenantService implements the ITenantService interface
type TenantService struct {
	repositoryContainer *repository.RepositoryContainer
	uowFactory          uow.UnitOfWorkFactory
	protobufMapper      *mappers.ProtobufMapper
	tenantMapper        *mappers.TenantMapper
}

// NewTenantService creates a new tenant service
func NewTenantService(
	repositoryContainer *repository.RepositoryContainer,
	uowFactory uow.UnitOfWorkFactory,
	protobufMapper *mappers.ProtobufMapper,
	tenantMapper *mappers.TenantMapper,
) ITenantService {
	return &TenantService{
		repositoryContainer: repositoryContainer,
		uowFactory:          uowFactory,
		protobufMapper:      protobufMapper,
		tenantMapper:        tenantMapper,
	}
}

// CreateTenant creates a new tenant with admin user setup
func (s *TenantService) CreateTenant(ctx context.Context, req *tenantpb.CreateTenantRequest) (*tenantpb.Tenant, error) {
	// Validate request
	if err := s.protobufMapper.ValidateCreateTenantRequest(req); err != nil {
		return nil, err
	}

	// Convert protobuf to domain model
	tenant := s.protobufMapper.FromProtobufCreateTenantRequest(req)

	// Use unit of work for transactional operations
	unitOfWork := s.uowFactory.Create()
	if err := unitOfWork.Begin(ctx); err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer unitOfWork.Rollback()

	// Create tenant
	createdTenant, err := unitOfWork.TenantRepository().CreateTenant(ctx, tenant)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Create admin user if provided
	if req.AdminEmail != nil || req.AdminUsername != nil {
		adminUser := &models.TenantAdminUser{
			TenantID:  createdTenant.ID,
			IsPrimary: true,
		}

		if req.AdminEmail != nil {
			adminUser.Email = &req.AdminEmail.Value
		}
		if req.AdminUsername != nil {
			adminUser.Username = &req.AdminUsername.Value
		}
		if req.AdminUserId != nil {
			adminUser.UserID = req.AdminUserId.Value
		}

		_, err = unitOfWork.TenantAdminUserRepository().CreateTenantAdminUser(ctx, adminUser)
		if err != nil {
			return nil, fmt.Errorf("failed to create admin user: %w", err)
		}
	}

	// Commit transaction
	if err := unitOfWork.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Convert back to protobuf
	return s.protobufMapper.ToProtobufTenant(createdTenant, nil, nil), nil
}

// UpdateTenant updates an existing tenant
func (s *TenantService) UpdateTenant(ctx context.Context, req *tenantpb.UpdateTenantRequest) (*tenantpb.Tenant, error) {
	// Validate request
	if err := s.protobufMapper.ValidateUpdateTenantRequest(req); err != nil {
		return nil, err
	}

	// Convert protobuf to domain model
	tenant, _, _, updateMask := s.protobufMapper.FromProtobufUpdateTenantRequest(req)

	// Update tenant
	updatedTenant, err := s.repositoryContainer.Tenant.UpdateTenant(ctx, tenant, updateMask)
	if err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	// Get related data for response
	connectionStrings, _ := s.repositoryContainer.TenantConnectionString.GetTenantConnectionStrings(ctx, updatedTenant.ID)
	features, _ := s.repositoryContainer.TenantFeature.GetTenantFeatures(ctx, updatedTenant.ID)

	// Convert back to protobuf
	return s.protobufMapper.ToProtobufTenant(updatedTenant, connectionStrings, features), nil
}

// DeleteTenant deactivates a tenant
func (s *TenantService) DeleteTenant(ctx context.Context, req *tenantpb.DeleteTenantRequest) (*tenantpb.DeleteTenantReply, error) {
	tenantID := s.protobufMapper.ParseDeleteTenantRequest(req)

	// Deactivate tenant instead of hard delete
	err := s.repositoryContainer.Tenant.DeactivateTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete tenant: %w", err)
	}

	return s.protobufMapper.ToProtobufDeleteTenantReply(req.Id), nil
}

// GetTenant retrieves a tenant by ID or name
func (s *TenantService) GetTenant(ctx context.Context, req *tenantpb.GetTenantRequest) (*tenantpb.Tenant, error) {
	idOrName, isID := s.protobufMapper.ParseGetTenantRequest(req)

	var tenant *models.Tenant
	var err error

	if isID {
		tenantID, _ := uuid.Parse(idOrName)
		tenant, err = s.repositoryContainer.Tenant.GetTenantByID(ctx, tenantID)
	} else {
		tenant, err = s.repositoryContainer.Tenant.GetTenantByName(ctx, idOrName)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	// Get related data
	connectionStrings, _ := s.repositoryContainer.TenantConnectionString.GetTenantConnectionStrings(ctx, tenant.ID)
	features, _ := s.repositoryContainer.TenantFeature.GetTenantFeatures(ctx, tenant.ID)

	// Convert to protobuf
	return s.protobufMapper.ToProtobufTenant(tenant, connectionStrings, features), nil
}

// GetTenantPublic retrieves public tenant information
func (s *TenantService) GetTenantPublic(ctx context.Context, req *tenantpb.GetTenantPublicRequest) (*tenantpb.TenantInfo, error) {
	idOrName, isID := s.protobufMapper.ParseGetTenantRequest(&tenantpb.GetTenantRequest{
		IdOrName: req.IdOrName,
	})

	var tenant *models.Tenant
	var err error

	if isID {
		tenantID, _ := uuid.Parse(idOrName)
		tenant, err = s.repositoryContainer.Tenant.GetTenantByID(ctx, tenantID)
	} else {
		tenant, err = s.repositoryContainer.Tenant.GetTenantByName(ctx, idOrName)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	// Convert to public tenant info (no sensitive data)
	return s.protobufMapper.ToProtobufTenantInfo(tenant), nil
}

// ListTenant lists tenants with filtering and pagination
func (s *TenantService) ListTenant(ctx context.Context, req *tenantpb.ListTenantRequest) (*tenantpb.ListTenantReply, error) {
	limit, offset, search, region := s.protobufMapper.ParseListTenantRequest(req)

	var tenants []*models.Tenant
	var err error

	// Apply filtering based on request
	if search != "" {
		tenants, err = s.repositoryContainer.Tenant.SearchTenantsByName(ctx, search, limit, offset)
	} else if region != "" {
		tenants, err = s.repositoryContainer.Tenant.ListTenantsByRegion(ctx, region, limit, offset)
	} else {
		tenants, err = s.repositoryContainer.Tenant.ListTenants(ctx, limit, offset)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}

	// For simplicity, using the same count for total and filter size
	// In a real implementation, you'd run separate count queries
	totalSize := int32(len(tenants))
	filterSize := totalSize

	return s.protobufMapper.ToProtobufListTenantReply(tenants, totalSize, filterSize), nil
}

// GetCurrentTenant retrieves the current tenant context
func (s *TenantService) GetCurrentTenant(ctx context.Context, req *tenantpb.GetCurrentTenantRequest) (*tenantpb.GetCurrentTenantReply, error) {
	// In a real implementation, you would extract tenant context from the request context
	// For now, we'll return a placeholder response

	// This would typically come from JWT token or request headers
	currentTenantID := s.extractTenantFromContext(ctx)
	isHost := s.isHostTenant(ctx)

	if currentTenantID == uuid.Nil {
		return &tenantpb.GetCurrentTenantReply{
			IsHost: true,
		}, nil
	}

	tenant, err := s.repositoryContainer.Tenant.GetTenantByID(ctx, currentTenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current tenant: %w", err)
	}

	return s.protobufMapper.ToProtobufGetCurrentTenantReply(tenant, isHost), nil
}

// GetTenantForInternalUse retrieves tenant for internal service communication
func (s *TenantService) GetTenantForInternalUse(ctx context.Context, req *tenantpb.GetTenantRequest) (*tenantpb.Tenant, error) {
	// This could have different authorization or include additional internal fields
	// For now, delegate to the public method
	return s.GetTenant(ctx, req)
}

// Helper methods

func (s *TenantService) extractTenantFromContext(ctx context.Context) uuid.UUID {
	// Extract tenant ID from context (JWT token, headers, etc.)
	// This is a placeholder implementation
	return uuid.Nil
}

func (s *TenantService) isHostTenant(ctx context.Context) bool {
	// Determine if the current context represents the host tenant
	// This is a placeholder implementation
	return true
}
