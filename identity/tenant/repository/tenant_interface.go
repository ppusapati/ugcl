package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"p9e.in/ugcl/identity/tenant/models"
)

type ITenantRepository interface {
	// Core CRUD operations
	CreateTenant(ctx context.Context, tenant *models.Tenant) (*models.Tenant, error)
	GetTenantByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error)
	GetTenantByName(ctx context.Context, name string) (*models.Tenant, error)
	UpdateTenant(ctx context.Context, tenant *models.Tenant, updateMask *fieldmaskpb.FieldMask) (*models.Tenant, error)
	DeactivateTenant(ctx context.Context, id uuid.UUID) error

	// List and search operations
	ListTenants(ctx context.Context, limit, offset int32) ([]*models.Tenant, error)
	ListTenantsByRegion(ctx context.Context, region string, limit, offset int32) ([]*models.Tenant, error)
	SearchTenantsByName(ctx context.Context, searchTerm string, limit, offset int32) ([]*models.Tenant, error)

	// Statistics and analytics
	GetTenantSummary(ctx context.Context, id uuid.UUID) (*models.TenantSummary, error)
	GetTenantStats(ctx context.Context, id uuid.UUID) (*models.TenantStats, error)
	GetTenantsByRegionStats(ctx context.Context) ([]*models.TenantRegionStats, error)
	GetTenantSubscriptionStats(ctx context.Context) ([]*models.TenantSubscriptionStats, error)

	// Security and compliance
	GetTenantsWithExpiredSubscriptions(ctx context.Context) ([]*models.Tenant, error)
	GetTenantsExceedingLimits(ctx context.Context) ([]*models.Tenant, error)
}

type ITenantFeatureRepository interface {
	CreateTenantFeature(ctx context.Context, feature *models.TenantFeature) (*models.TenantFeature, error)
	GetTenantFeature(ctx context.Context, tenantID uuid.UUID, key string) (*models.TenantFeature, error)
	GetTenantFeatures(ctx context.Context, tenantID uuid.UUID) ([]*models.TenantFeature, error)
	UpdateTenantFeature(ctx context.Context, feature *models.TenantFeature) (*models.TenantFeature, error)
	UpsertTenantFeature(ctx context.Context, feature *models.TenantFeature) (*models.TenantFeature, error)
	DeleteTenantFeature(ctx context.Context, tenantID uuid.UUID, key string) error
}

type ITenantConnectionStringRepository interface {
	CreateTenantConnectionString(ctx context.Context, cs *models.TenantConnectionString) (*models.TenantConnectionString, error)
	GetTenantConnectionString(ctx context.Context, tenantID uuid.UUID, key string) (*models.TenantConnectionString, error)
	GetTenantConnectionStrings(ctx context.Context, tenantID uuid.UUID) ([]*models.TenantConnectionString, error)
	UpdateTenantConnectionString(ctx context.Context, cs *models.TenantConnectionString) (*models.TenantConnectionString, error)
	UpsertTenantConnectionString(ctx context.Context, cs *models.TenantConnectionString) (*models.TenantConnectionString, error)
	DeleteTenantConnectionString(ctx context.Context, tenantID uuid.UUID, key string) error
}

type ITenantDomainRepository interface {
	CreateTenantDomain(ctx context.Context, domain *models.TenantDomain) (*models.TenantDomain, error)
	GetTenantDomain(ctx context.Context, tenantID uuid.UUID, domain string) (*models.TenantDomain, error)
	GetTenantDomains(ctx context.Context, tenantID uuid.UUID) ([]*models.TenantDomain, error)
	GetTenantByDomain(ctx context.Context, domain string) (*models.TenantByDomain, error)
	GetPrimaryDomain(ctx context.Context, tenantID uuid.UUID) (*models.TenantDomain, error)
	SetPrimaryDomain(ctx context.Context, tenantID uuid.UUID, domain string) error
	VerifyDomain(ctx context.Context, tenantID uuid.UUID, domain string) error
	UpdateDomainSSL(ctx context.Context, tenantID uuid.UUID, domain, certificate, privateKey string) error
	DeleteTenantDomain(ctx context.Context, tenantID uuid.UUID, domain string) error
}

type ITenantAdminUserRepository interface {
	CreateTenantAdminUser(ctx context.Context, adminUser *models.TenantAdminUser) (*models.TenantAdminUser, error)
	GetTenantAdminUsers(ctx context.Context, tenantID uuid.UUID) ([]*models.TenantAdminUser, error)
	GetTenantPrimaryAdmin(ctx context.Context, tenantID uuid.UUID) (*models.TenantAdminUser, error)
	GetUserTenantAdminRoles(ctx context.Context, userID string) ([]*models.TenantAdminUser, error)
	UpdateTenantAdminUser(ctx context.Context, adminUser *models.TenantAdminUser) (*models.TenantAdminUser, error)
	SetPrimaryAdmin(ctx context.Context, tenantID uuid.UUID, userID string) error
	DeleteTenantAdminUser(ctx context.Context, tenantID uuid.UUID, userID string) error
}

type ITenantMetadataRepository interface {
	CreateTenantMetadata(ctx context.Context, metadata *models.TenantMetadata) (*models.TenantMetadata, error)
	GetTenantMetadata(ctx context.Context, tenantID uuid.UUID) (*models.TenantMetadata, error)
	UpdateTenantMetadata(ctx context.Context, metadata *models.TenantMetadata) (*models.TenantMetadata, error)
	UpsertTenantMetadata(ctx context.Context, metadata *models.TenantMetadata) (*models.TenantMetadata, error)
	DeleteTenantMetadata(ctx context.Context, tenantID uuid.UUID) error
}

type ITenantBillingRepository interface {
	CreateTenantBilling(ctx context.Context, billing *models.TenantBilling) (*models.TenantBilling, error)
	GetTenantBilling(ctx context.Context, tenantID uuid.UUID) (*models.TenantBilling, error)
	UpdateTenantBilling(ctx context.Context, billing *models.TenantBilling) (*models.TenantBilling, error)
	UpdateTenantPaymentMethod(ctx context.Context, tenantID uuid.UUID, paymentMethodID string) error
	UpdateTenantSubscription(ctx context.Context, tenantID uuid.UUID, subscriptionID string) error
	GetTenantsByPaymentMethod(ctx context.Context, paymentMethodID string) ([]*models.TenantWithBilling, error)
	GetTenantsBySubscription(ctx context.Context, subscriptionID string) ([]*models.TenantWithBilling, error)
	DeleteTenantBilling(ctx context.Context, tenantID uuid.UUID) error
}

type ITenantUsageMetricsRepository interface {
	RecordTenantUsageMetric(ctx context.Context, metric *models.TenantUsageMetric) (*models.TenantUsageMetric, error)
	GetTenantUsageMetrics(ctx context.Context, tenantID uuid.UUID, metricName string, limit, offset int32) ([]*models.TenantUsageMetric, error)
	GetTenantUsageMetricsInPeriod(ctx context.Context, tenantID uuid.UUID, start, end time.Time) ([]*models.TenantUsageMetric, error)
	GetLatestTenantUsageMetric(ctx context.Context, tenantID uuid.UUID, metricName string) (*models.TenantUsageMetric, error)
	GetTenantUsageMetricsSummary(ctx context.Context, tenantID uuid.UUID, start, end time.Time) ([]*models.TenantUsageMetricsSummary, error)
	DeleteOldUsageMetrics(ctx context.Context, before time.Time) error
}

type ITenantAuditLogRepository interface {
	CreateTenantAuditEntry(ctx context.Context, entry *models.TenantAuditEntry) (*models.TenantAuditEntry, error)
	GetTenantAuditLog(ctx context.Context, tenantID uuid.UUID, limit, offset int32) ([]*models.TenantAuditEntry, error)
	GetTenantAuditLogByAction(ctx context.Context, tenantID uuid.UUID, action string, limit, offset int32) ([]*models.TenantAuditEntry, error)
	GetTenantAuditLogByActor(ctx context.Context, tenantID uuid.UUID, actorID string, limit, offset int32) ([]*models.TenantAuditEntry, error)
	GetTenantAuditLogInPeriod(ctx context.Context, tenantID uuid.UUID, start, end time.Time, limit, offset int32) ([]*models.TenantAuditEntry, error)
	GetAllTenantsAuditLog(ctx context.Context, start, end time.Time, limit, offset int32) ([]*models.TenantAuditWithTenant, error)
	DeleteOldAuditEntries(ctx context.Context, before time.Time) error
}

type ITenantDatabaseSchemaRepository interface {
	CreateTenantDatabaseSchema(ctx context.Context, schema *models.TenantDatabaseSchema) (*models.TenantDatabaseSchema, error)
	GetTenantDatabaseSchema(ctx context.Context, tenantID uuid.UUID) (*models.TenantDatabaseSchema, error)
	UpdateTenantDatabaseSchema(ctx context.Context, schema *models.TenantDatabaseSchema) (*models.TenantDatabaseSchema, error)
	UpdateTenantMigrationVersion(ctx context.Context, tenantID uuid.UUID, version string) error
	DeactivateTenantDatabaseSchema(ctx context.Context, tenantID uuid.UUID) error
	GetTenantsWithSeparateSchemas(ctx context.Context) ([]*models.TenantDatabaseSchemaWithTenant, error)
}

type ITenantCleanupRepository interface {
	CleanupUnverifiedDomains(ctx context.Context, before time.Time) error
	CleanupInactiveTenantsData(ctx context.Context, before time.Time) error
	CleanupOldTenantAuditLogs(ctx context.Context, before time.Time) error
	CleanupOldUsageMetrics(ctx context.Context, before time.Time) error
	ArchiveInactiveTenants(ctx context.Context, before time.Time) ([]*models.Tenant, error)
	GetOrphanedTenantData(ctx context.Context) ([]*models.OrphanedData, error)
	GetTenantDataSizes(ctx context.Context) ([]*models.TenantDataSize, error)
}

type ITenantReportRepository interface {
	GetTenantDashboardStats(ctx context.Context) (*models.TenantDashboardStats, error)
	GetTenantGrowthStats(ctx context.Context, period string, since time.Time) ([]*models.TenantGrowthStats, error)
	GetTenantFeatureUsage(ctx context.Context) ([]*models.TenantFeatureUsage, error)
	GetTenantsBySubscriptionPlan(ctx context.Context) ([]*models.TenantsBySubscriptionPlan, error)
	GetTenantActivitySummary(ctx context.Context, limit, offset int32) ([]*models.TenantActivitySummary, error)
}
