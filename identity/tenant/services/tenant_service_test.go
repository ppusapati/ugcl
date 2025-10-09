package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	tenantpb "p9e.in/ugcl/identity/tenant/api/v1/tenant"
	"p9e.in/ugcl/identity/tenant/mappers"
	"p9e.in/ugcl/identity/tenant/models"
	"p9e.in/ugcl/identity/tenant/repository"
	"p9e.in/ugcl/identity/tenant/uow"
)

// -------------------- Mocks --------------------

type mockTenantRepo struct {
	CreateTenantFunc        func(ctx context.Context, tenant *models.Tenant) (*models.Tenant, error)
	GetTenantByIDFunc       func(ctx context.Context, id uuid.UUID) (*models.Tenant, error)
	GetTenantByNameFunc     func(ctx context.Context, name string) (*models.Tenant, error)
	UpdateTenantFunc        func(ctx context.Context, tenant *models.Tenant) (*models.Tenant, error)
	DeactivateTenantFunc    func(ctx context.Context, id uuid.UUID) error
	ListTenantsFunc         func(ctx context.Context, limit, offset int32) ([]*models.Tenant, error)
	ListTenantsByRegionFunc func(ctx context.Context, region string, limit, offset int32) ([]*models.Tenant, error)
	SearchTenantsByNameFunc func(ctx context.Context, searchTerm string, limit, offset int32) ([]*models.Tenant, error)
}

func (m *mockTenantRepo) CreateTenant(ctx context.Context, tenant *models.Tenant) (*models.Tenant, error) {
	if m.CreateTenantFunc != nil {
		return m.CreateTenantFunc(ctx, tenant)
	}
	return tenant, nil
}
func (m *mockTenantRepo) GetTenantByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	if m.GetTenantByIDFunc != nil {
		return m.GetTenantByIDFunc(ctx, id)
	}
	return nil, nil
}
func (m *mockTenantRepo) GetTenantByName(ctx context.Context, name string) (*models.Tenant, error) {
	if m.GetTenantByNameFunc != nil {
		return m.GetTenantByNameFunc(ctx, name)
	}
	return nil, nil
}
func (m *mockTenantRepo) UpdateTenant(ctx context.Context, tenant *models.Tenant, _ interface{}) (*models.Tenant, error) {
	if m.UpdateTenantFunc != nil {
		return m.UpdateTenantFunc(ctx, tenant)
	}
	return tenant, nil
}
func (m *mockTenantRepo) DeactivateTenant(ctx context.Context, id uuid.UUID) error {
	if m.DeactivateTenantFunc != nil {
		return m.DeactivateTenantFunc(ctx, id)
	}
	return nil
}
func (m *mockTenantRepo) ListTenants(ctx context.Context, limit, offset int32) ([]*models.Tenant, error) {
	if m.ListTenantsFunc != nil {
		return m.ListTenantsFunc(ctx, limit, offset)
	}
	return nil, nil
}
func (m *mockTenantRepo) ListTenantsByRegion(ctx context.Context, region string, limit, offset int32) ([]*models.Tenant, error) {
	if m.ListTenantsByRegionFunc != nil {
		return m.ListTenantsByRegionFunc(ctx, region, limit, offset)
	}
	return nil, nil
}
func (m *mockTenantRepo) SearchTenantsByName(ctx context.Context, searchTerm string, limit, offset int32) ([]*models.Tenant, error) {
	if m.SearchTenantsByNameFunc != nil {
		return m.SearchTenantsByNameFunc(ctx, searchTerm, limit, offset)
	}
	return nil, nil
}

// Admin user repo

type mockTenantAdminUserRepo struct {
	CreateFunc func(ctx context.Context, adminUser *models.TenantAdminUser) (*models.TenantAdminUser, error)
}

func (m *mockTenantAdminUserRepo) CreateTenantAdminUser(ctx context.Context, adminUser *models.TenantAdminUser) (*models.TenantAdminUser, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, adminUser)
	}
	return adminUser, nil
}
func (m *mockTenantAdminUserRepo) GetTenantAdminUsers(ctx context.Context, tenantID uuid.UUID) ([]*models.TenantAdminUser, error) {
	return nil, nil
}
func (m *mockTenantAdminUserRepo) GetTenantPrimaryAdmin(ctx context.Context, tenantID uuid.UUID) (*models.TenantAdminUser, error) {
	return nil, nil
}
func (m *mockTenantAdminUserRepo) GetUserTenantAdminRoles(ctx context.Context, userID string) ([]*models.TenantAdminUser, error) {
	return nil, nil
}
func (m *mockTenantAdminUserRepo) UpdateTenantAdminUser(ctx context.Context, adminUser *models.TenantAdminUser) (*models.TenantAdminUser, error) {
	return adminUser, nil
}
func (m *mockTenantAdminUserRepo) SetPrimaryAdmin(ctx context.Context, tenantID uuid.UUID, userID string) error {
	return nil
}
func (m *mockTenantAdminUserRepo) DeleteTenantAdminUser(ctx context.Context, tenantID uuid.UUID, userID string) error {
	return nil
}

// UOW mock

type mockUOW struct {
	TenantRepo    *mockTenantRepo
	AdminUserRepo *mockTenantAdminUserRepo
	begun         bool
	committed     bool
	rolledBack    bool
	beginErr      error
	commitErr     error
}

func (m *mockUOW) Begin(ctx context.Context) error                              { m.begun = true; return m.beginErr }
func (m *mockUOW) Commit() error                                                { m.committed = true; return m.commitErr }
func (m *mockUOW) Rollback() error                                              { m.rolledBack = true; return nil }
func (m *mockUOW) TenantRepository() repository.ITenantRepository               { return m.TenantRepo }
func (m *mockUOW) TenantFeatureRepository() repository.ITenantFeatureRepository { return nil }
func (m *mockUOW) TenantConnectionStringRepository() repository.ITenantConnectionStringRepository {
	return nil
}
func (m *mockUOW) TenantDomainRepository() repository.ITenantDomainRepository { return nil }
func (m *mockUOW) TenantAdminUserRepository() repository.ITenantAdminUserRepository {
	return m.AdminUserRepo
}
func (m *mockUOW) TenantMetadataRepository() repository.ITenantMetadataRepository         { return nil }
func (m *mockUOW) TenantBillingRepository() repository.ITenantBillingRepository           { return nil }
func (m *mockUOW) TenantUsageMetricsRepository() repository.ITenantUsageMetricsRepository { return nil }
func (m *mockUOW) TenantAuditLogRepository() repository.ITenantAuditLogRepository         { return nil }
func (m *mockUOW) TenantDatabaseSchemaRepository() repository.ITenantDatabaseSchemaRepository {
	return nil
}
func (m *mockUOW) TenantCleanupRepository() repository.ITenantCleanupRepository { return nil }
func (m *mockUOW) TenantReportRepository() repository.ITenantReportRepository   { return nil }

var _ uow.UnitOfWork = (*mockUOW)(nil)

type mockUOWFactory struct{ u *mockUOW }

func (f *mockUOWFactory) Create() uow.UnitOfWork { return f.u }

var _ uow.UnitOfWorkFactory = (*mockUOWFactory)(nil)

// -------------------- Tests --------------------

func TestCreateTenant_ValidationErrors(t *testing.T) {
	repoContainer := &repository.RepositoryContainer{}
	uowMock := &mockUOW{}
	factory := &mockUOWFactory{u: uowMock}
	mapper := mappers.NewProtobufMapper()
	tenantMapper := &mappers.TenantMapper{}
	svc := NewTenantService(repoContainer, factory, mapper, tenantMapper)

	// missing name
	_, err := svc.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{DisplayName: "ACME", Region: "us-east"})
	if err == nil {
		t.Fatalf("expected validation error for name")
	}
	// missing display name
	_, err = svc.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{Name: "acme", Region: "us-east"})
	if err == nil {
		t.Fatalf("expected validation error for display name")
	}
	// missing region
	_, err = svc.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{Name: "acme", DisplayName: "ACME"})
	if err == nil {
		t.Fatalf("expected validation error for region")
	}
}

func TestCreateTenant_Success_WithAdminUser(t *testing.T) {
	createdID := uuid.New()
	tenantRepo := &mockTenantRepo{
		CreateTenantFunc: func(ctx context.Context, tenant *models.Tenant) (*models.Tenant, error) {
			return &models.Tenant{ID: createdID, Name: tenant.Name, DisplayName: tenant.DisplayName, Region: tenant.Region}, nil
		},
	}
	adminRepo := &mockTenantAdminUserRepo{
		CreateFunc: func(ctx context.Context, adminUser *models.TenantAdminUser) (*models.TenantAdminUser, error) {
			if adminUser.TenantID != createdID {
				t.Errorf("expected admin user tenant id %s, got %s", createdID, adminUser.TenantID)
			}
			if !adminUser.IsPrimary {
				t.Errorf("expected primary admin user")
			}
			return adminUser, nil
		},
	}
	uowMock := &mockUOW{TenantRepo: tenantRepo, AdminUserRepo: adminRepo}
	factory := &mockUOWFactory{u: uowMock}
	mapper := mappers.NewProtobufMapper()
	tenantMapper := &mappers.TenantMapper{}
	repoContainer := &repository.RepositoryContainer{}
	svc := NewTenantService(repoContainer, factory, mapper, tenantMapper)

	name, display, region := "acme", "ACME", "us-east"
	email := "admin@acme.com"
	req := &tenantpb.CreateTenantRequest{Name: name, DisplayName: display, Region: region, AdminEmail: &tenantpb.OptionalString{Value: email}}

	resp, err := svc.CreateTenant(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil || resp.Id == "" {
		t.Fatalf("expected tenant response with id")
	}
	if !uowMock.begun || !uowMock.committed || !uowMock.rolledBack {
		t.Errorf("expected Begin, Commit and deferred Rollback to be called")
	}
}

func TestCreateTenant_Fails_WhenRepoError(t *testing.T) {
	tenantRepo := &mockTenantRepo{
		CreateTenantFunc: func(ctx context.Context, tenant *models.Tenant) (*models.Tenant, error) {
			return nil, errors.New("db error")
		},
	}
	uowMock := &mockUOW{TenantRepo: tenantRepo}
	factory := &mockUOWFactory{u: uowMock}
	mapper := mappers.NewProtobufMapper()
	tenantMapper := &mappers.TenantMapper{}
	repoContainer := &repository.RepositoryContainer{}
	svc := NewTenantService(repoContainer, factory, mapper, tenantMapper)

	req := &tenantpb.CreateTenantRequest{Name: "acme", DisplayName: "ACME", Region: "us-east"}
	_, err := svc.CreateTenant(context.Background(), req)
	if err == nil {
		t.Fatalf("expected error from repository")
	}
	if !uowMock.begun {
		t.Errorf("expected transaction to begin")
	}
	if uowMock.committed {
		t.Errorf("did not expect commit on error")
	}
	if !uowMock.rolledBack {
		t.Errorf("expected rollback to be called")
	}
}
