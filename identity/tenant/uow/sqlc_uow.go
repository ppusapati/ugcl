package uow

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	sqlcgen "p9e.in/ugcl/identity/tenant/db/generated"
	"p9e.in/ugcl/identity/tenant/mappers"
	repo "p9e.in/ugcl/identity/tenant/repository"
)

// SqlcUnitOfWork implements UnitOfWork using SQLC and pgx transactions
type SqlcUnitOfWork struct {
	pool   *pgxpool.Pool
	tx     pgx.Tx
	mapper *mappers.TenantMapper

	// Repository instances
	tenantRepo                 repo.ITenantRepository
	tenantFeatureRepo          repo.ITenantFeatureRepository
	tenantConnectionStringRepo repo.ITenantConnectionStringRepository
	tenantDomainRepo           repo.ITenantDomainRepository
	tenantAdminUserRepo        repo.ITenantAdminUserRepository
	tenantMetadataRepo         repo.ITenantMetadataRepository
	tenantBillingRepo          repo.ITenantBillingRepository
	tenantUsageMetricsRepo     repo.ITenantUsageMetricsRepository
	tenantAuditLogRepo         repo.ITenantAuditLogRepository
	tenantDatabaseSchemaRepo   repo.ITenantDatabaseSchemaRepository
	tenantCleanupRepo          repo.ITenantCleanupRepository
	tenantReportRepo           repo.ITenantReportRepository
}

func NewSqlcUnitOfWork(pool *pgxpool.Pool, mapper *mappers.TenantMapper) *SqlcUnitOfWork {
	return &SqlcUnitOfWork{
		pool:   pool,
		mapper: mapper,
	}
}

func (uow *SqlcUnitOfWork) Begin(ctx context.Context) error {
	tx, err := uow.pool.Begin(ctx)
	if err != nil {
		return err
	}

	uow.tx = tx

	// Create SQLC queries with transaction
	queries := sqlcgen.New(tx)

	// Initialize repositories with transactional queries
	uow.tenantRepo = repo.NewTenantRepository(queries, uow.mapper)
	uow.tenantFeatureRepo = repo.NewTenantFeatureRepository(queries, uow.mapper)
	uow.tenantConnectionStringRepo = repo.NewTenantConnectionStringRepository(queries, uow.mapper)
	uow.tenantDomainRepo = repo.NewTenantDomainRepository(queries, uow.mapper)
	uow.tenantAdminUserRepo = repo.NewTenantAdminUserRepository(queries, uow.mapper)
	uow.tenantMetadataRepo = repo.NewTenantMetadataRepository(queries, uow.mapper)
	uow.tenantBillingRepo = repo.NewTenantBillingRepository(queries, uow.mapper)
	uow.tenantUsageMetricsRepo = repo.NewTenantUsageMetricsRepository(queries, uow.mapper)
	uow.tenantAuditLogRepo = repo.NewTenantAuditLogRepository(queries, uow.mapper)
	uow.tenantDatabaseSchemaRepo = repo.NewTenantDatabaseSchemaRepository(queries, uow.mapper)
	uow.tenantCleanupRepo = repo.NewTenantCleanupRepository(queries, uow.mapper)
	uow.tenantReportRepo = repo.NewTenantReportRepository(queries, uow.mapper)

	return nil
}

func (uow *SqlcUnitOfWork) Commit() error {
	if uow.tx == nil {
		return nil
	}
	return uow.tx.Commit(context.Background())
}

func (uow *SqlcUnitOfWork) Rollback() error {
	if uow.tx == nil {
		return nil
	}
	return uow.tx.Rollback(context.Background())
}

// Repository accessors

func (uow *SqlcUnitOfWork) TenantRepository() repo.ITenantRepository {
	return uow.tenantRepo
}

func (uow *SqlcUnitOfWork) TenantFeatureRepository() repo.ITenantFeatureRepository {
	return uow.tenantFeatureRepo
}

func (uow *SqlcUnitOfWork) TenantConnectionStringRepository() repo.ITenantConnectionStringRepository {
	return uow.tenantConnectionStringRepo
}

func (uow *SqlcUnitOfWork) TenantDomainRepository() repo.ITenantDomainRepository {
	return uow.tenantDomainRepo
}

func (uow *SqlcUnitOfWork) TenantAdminUserRepository() repo.ITenantAdminUserRepository {
	return uow.tenantAdminUserRepo
}

func (uow *SqlcUnitOfWork) TenantMetadataRepository() repo.ITenantMetadataRepository {
	return uow.tenantMetadataRepo
}

func (uow *SqlcUnitOfWork) TenantBillingRepository() repo.ITenantBillingRepository {
	return uow.tenantBillingRepo
}

func (uow *SqlcUnitOfWork) TenantUsageMetricsRepository() repo.ITenantUsageMetricsRepository {
	return uow.tenantUsageMetricsRepo
}

func (uow *SqlcUnitOfWork) TenantAuditLogRepository() repo.ITenantAuditLogRepository {
	return uow.tenantAuditLogRepo
}

func (uow *SqlcUnitOfWork) TenantDatabaseSchemaRepository() repo.ITenantDatabaseSchemaRepository {
	return uow.tenantDatabaseSchemaRepo
}

func (uow *SqlcUnitOfWork) TenantCleanupRepository() repo.ITenantCleanupRepository {
	return uow.tenantCleanupRepo
}

func (uow *SqlcUnitOfWork) TenantReportRepository() repo.ITenantReportRepository {
	return uow.tenantReportRepo
}

// SqlcUnitOfWorkFactory creates new SqlcUnitOfWork instances
type SqlcUnitOfWorkFactory struct {
	pool   *pgxpool.Pool
	mapper *mappers.TenantMapper
}

func NewSqlcUnitOfWorkFactory(pool *pgxpool.Pool, mapper *mappers.TenantMapper) UnitOfWorkFactory {
	return &SqlcUnitOfWorkFactory{
		pool:   pool,
		mapper: mapper,
	}
}

func (factory *SqlcUnitOfWorkFactory) Create() UnitOfWork {
	return NewSqlcUnitOfWork(factory.pool, factory.mapper)
}
