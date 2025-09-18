package repository

import (
	sqlcgen "p9e.in/ugcl/identity/tenant/db/generated"
	"p9e.in/ugcl/identity/tenant/mappers"
)

// RepositoryContainer holds all tenant repositories
type RepositoryContainer struct {
	Tenant                 ITenantRepository
	TenantFeature          ITenantFeatureRepository
	TenantConnectionString ITenantConnectionStringRepository
	TenantDomain           ITenantDomainRepository
	TenantAdminUser        ITenantAdminUserRepository
	TenantMetadata         ITenantMetadataRepository
	TenantBilling          ITenantBillingRepository
	TenantUsageMetrics     ITenantUsageMetricsRepository
	TenantAuditLog         ITenantAuditLogRepository
	TenantDatabaseSchema   ITenantDatabaseSchemaRepository
	TenantCleanup          ITenantCleanupRepository
	TenantReport           ITenantReportRepository
}

// NewRepositoryContainer creates a new repository container with all tenant repositories
func NewRepositoryContainer(queries *sqlcgen.Queries, mapper *mappers.TenantMapper) *RepositoryContainer {
	return &RepositoryContainer{
		Tenant:                 NewTenantRepository(queries, mapper),
		TenantFeature:          NewTenantFeatureRepository(queries, mapper),
		TenantConnectionString: NewTenantConnectionStringRepository(queries, mapper),
		TenantDomain:           NewTenantDomainRepository(queries, mapper),
		TenantAdminUser:        NewTenantAdminUserRepository(queries, mapper),
		TenantMetadata:         NewTenantMetadataRepository(queries, mapper),
		TenantBilling:          NewTenantBillingRepository(queries, mapper),
		TenantUsageMetrics:     NewTenantUsageMetricsRepository(queries, mapper),
		TenantAuditLog:         NewTenantAuditLogRepository(queries, mapper),
		TenantDatabaseSchema:   NewTenantDatabaseSchemaRepository(queries, mapper),
		TenantCleanup:          NewTenantCleanupRepository(queries, mapper),
		TenantReport:           NewTenantReportRepository(queries, mapper),
	}
}

// ProvideRepositoryContainer is a provider function for fx dependency injection
func ProvideRepositoryContainer(queries *sqlcgen.Queries, mapper *mappers.TenantMapper) *RepositoryContainer {
	return NewRepositoryContainer(queries, mapper)
}
