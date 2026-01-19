package uow

import (
	"context"

	interfaces "p9e.in/ugcl/identity/tenant/repository"
)

// UnitOfWork interface defines the contract for database transactions
type UnitOfWork interface {
	// Transaction management
	Begin(ctx context.Context) error
	Commit() error
	Rollback() error

	// Repository access within transaction
	TenantRepository() interfaces.ITenantRepository
	TenantFeatureRepository() interfaces.ITenantFeatureRepository
	TenantConnectionStringRepository() interfaces.ITenantConnectionStringRepository
	TenantDomainRepository() interfaces.ITenantDomainRepository
	TenantAdminUserRepository() interfaces.ITenantAdminUserRepository
	TenantMetadataRepository() interfaces.ITenantMetadataRepository
	TenantBillingRepository() interfaces.ITenantBillingRepository
	TenantUsageMetricsRepository() interfaces.ITenantUsageMetricsRepository
	TenantAuditLogRepository() interfaces.ITenantAuditLogRepository
	TenantDatabaseSchemaRepository() interfaces.ITenantDatabaseSchemaRepository
	TenantCleanupRepository() interfaces.ITenantCleanupRepository
	TenantReportRepository() interfaces.ITenantReportRepository
}

// UnitOfWorkFactory creates new unit of work instances
type UnitOfWorkFactory interface {
	Create() UnitOfWork
}
