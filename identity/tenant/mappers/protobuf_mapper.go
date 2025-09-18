package mappers

import (
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	tenantpb "p9e.in/ugcl/identity/tenant/api/v1/tenant"
	"p9e.in/ugcl/identity/tenant/models"
)

// ProtobufMapper handles conversions between protobuf messages and domain models
type ProtobufMapper struct{}

func NewProtobufMapper() *ProtobufMapper {
	return &ProtobufMapper{}
}

// Tenant conversions

func (m *ProtobufMapper) FromProtobufCreateTenantRequest(req *tenantpb.CreateTenantRequest) *models.Tenant {
	tenant := &models.Tenant{
		Name:         req.Name,
		DisplayName:  req.DisplayName,
		Region:       req.Region,
		MaxUsers:     100, // Default values
		MaxStorageGB: 10,
		TenantDB:     0, // Default to SHARED
		IsActive:     true,
	}

	if req.Logo != "" {
		tenant.Logo = &req.Logo
	}

	// Handle database separation
	if req.SeparateDb {
		tenant.TenantDB = 1 // SEPERATEDB
	}

	return tenant
}

func (m *ProtobufMapper) FromProtobufUpdateTenantRequest(req *tenantpb.UpdateTenantRequest) (*models.Tenant, []*models.TenantConnectionString, []*models.TenantFeature, *fieldmaskpb.FieldMask) {
	updateTenant := req.Tenant

	// Parse UUID from string
	tenantID, _ := uuid.Parse(updateTenant.Id)

	tenant := &models.Tenant{
		ID:          tenantID,
		Name:        updateTenant.Name,
		DisplayName: updateTenant.DisplayName,
	}

	if updateTenant.Logo != "" {
		tenant.Logo = &updateTenant.Logo
	}

	// Handle connection strings conversion
	var connectionStrings []*models.TenantConnectionString
	if len(updateTenant.Conn) > 0 {
		connectionStrings = make([]*models.TenantConnectionString, len(updateTenant.Conn))
		for i, conn := range updateTenant.Conn {
			connectionStrings[i] = &models.TenantConnectionString{
				TenantID: tenantID,
				Key:      conn.Key,
				Value:    conn.Value,
			}
		}
	}

	// Handle features conversion
	var features []*models.TenantFeature
	if len(updateTenant.Features) > 0 {
		features = make([]*models.TenantFeature, len(updateTenant.Features))
		for i, feature := range updateTenant.Features {
			features[i] = &models.TenantFeature{
				TenantID:  tenantID,
				Key:       feature.Key,
				Value:     feature.Value,
				ValueType: "string", // Default type
			}
		}
	}

	return tenant, connectionStrings, features, req.UpdateMask
}

func (m *ProtobufMapper) ToProtobufTenant(tenant *models.Tenant, connectionStrings []*models.TenantConnectionString, features []*models.TenantFeature) *tenantpb.Tenant {
	pbTenant := &tenantpb.Tenant{
		Id:          tenant.ID.String(),
		Name:        tenant.Name,
		DisplayName: tenant.DisplayName,
		Region:      tenant.Region,
		TenantDb:    tenantpb.TenantDB(tenant.TenantDB),
	}

	if tenant.Logo != nil {
		pbTenant.Logo = *tenant.Logo
	}

	// Convert connection strings
	if connectionStrings != nil {
		pbTenant.Conn = make([]*tenantpb.TenantConnectionString, len(connectionStrings))
		for i, cs := range connectionStrings {
			pbTenant.Conn[i] = &tenantpb.TenantConnectionString{
				Key:   cs.Key,
				Value: cs.Value,
			}
		}
	}

	// Convert features
	if features != nil {
		pbTenant.Features = make([]*tenantpb.TenantFeature, len(features))
		for i, feature := range features {
			pbTenant.Features[i] = &tenantpb.TenantFeature{
				Key:   feature.Key,
				Value: feature.Value,
			}
		}
	}

	return pbTenant
}

func (m *ProtobufMapper) ToProtobufTenantInfo(tenant *models.Tenant) *tenantpb.TenantInfo {
	tenantInfo := &tenantpb.TenantInfo{
		Id:          tenant.ID.String(),
		Name:        tenant.Name,
		DisplayName: tenant.DisplayName,
		Region:      tenant.Region,
		TenantDb:    tenantpb.TenantDB(tenant.TenantDB),
	}

	if tenant.Logo != nil {
		tenantInfo.Logo = *tenant.Logo
	}

	return tenantInfo
}

func (m *ProtobufMapper) ToProtobufListTenantReply(tenants []*models.Tenant, totalSize, filterSize int32) *tenantpb.ListTenantReply {
	items := make([]*tenantpb.Tenant, len(tenants))
	for i, tenant := range tenants {
		items[i] = m.ToProtobufTenant(tenant, nil, nil) // Basic tenant info without relations
	}

	return &tenantpb.ListTenantReply{
		TotalSize:  totalSize,
		FilterSize: filterSize,
		Items:      items,
	}
}

func (m *ProtobufMapper) ToProtobufDeleteTenantReply(tenantID string) *tenantpb.DeleteTenantReply {
	return &tenantpb.DeleteTenantReply{
		Id: tenantID,
	}
}

func (m *ProtobufMapper) ToProtobufGetCurrentTenantReply(tenant *models.Tenant, isHost bool) *tenantpb.GetCurrentTenantReply {
	tenantInfo := m.ToProtobufTenantInfo(tenant)

	return &tenantpb.GetCurrentTenantReply{
		Tenant: tenantInfo,
		IsHost: isHost,
	}
}

// Request parsing helpers

func (m *ProtobufMapper) ParseGetTenantRequest(req *tenantpb.GetTenantRequest) (string, bool) {
	idOrName := req.IdOrName

	// Try to parse as UUID first
	if _, err := uuid.Parse(idOrName); err == nil {
		return idOrName, true // It's an ID
	}

	return idOrName, false // It's a name
}

func (m *ProtobufMapper) ParseDeleteTenantRequest(req *tenantpb.DeleteTenantRequest) uuid.UUID {
	tenantID, _ := uuid.Parse(req.Id)
	return tenantID
}

func (m *ProtobufMapper) ParseListTenantRequest(req *tenantpb.ListTenantRequest) (limit, offset int32, search string, region string) {
	limit = req.PageSize
	offset = req.PageOffset
	search = req.Search

	// Extract region from filter if present
	if req.Filter != nil && req.Filter.Region != nil {
		if req.Filter.Region.Eq != nil {
			region = req.Filter.Region.Eq.Value
		}
	}

	// Default page size
	if limit <= 0 {
		limit = 50
	}

	return limit, offset, search, region
}

// Validation helpers

func (m *ProtobufMapper) ValidateCreateTenantRequest(req *tenantpb.CreateTenantRequest) error {
	if req.Name == "" {
		return ErrTenantNameRequired
	}
	if req.DisplayName == "" {
		return ErrTenantDisplayNameRequired
	}
	if req.Region == "" {
		return ErrTenantRegionRequired
	}
	return nil
}

func (m *ProtobufMapper) ValidateUpdateTenantRequest(req *tenantpb.UpdateTenantRequest) error {
	if req.Tenant == nil {
		return ErrTenantRequired
	}
	if req.Tenant.Id == "" {
		return ErrTenantIDRequired
	}
	return nil
}

// Error definitions
var (
	ErrTenantNameRequired        = NewValidationError("tenant name is required")
	ErrTenantDisplayNameRequired = NewValidationError("tenant display name is required")
	ErrTenantRegionRequired      = NewValidationError("tenant region is required")
	ErrTenantRequired            = NewValidationError("tenant is required")
	ErrTenantIDRequired          = NewValidationError("tenant ID is required")
)

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func NewValidationError(message string) *ValidationError {
	return &ValidationError{Message: message}
}