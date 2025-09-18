package tenant

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"

	sqlcgen "p9e.in/ugcl/identity/tenant/db/generated"
	"p9e.in/ugcl/identity/tenant/handler"
	"p9e.in/ugcl/identity/tenant/mappers"
	"p9e.in/ugcl/identity/tenant/repository"
	"p9e.in/ugcl/identity/tenant/services"
	"p9e.in/ugcl/identity/tenant/uow"
	"p9e.in/ugcl/packages/database/sqlc"
)

// TenantModule bundles all tenant service dependencies
var TenantModule = fx.Module("tenant",
	fx.Provide(
		// Mappers
		mappers.NewTenantMapper,
		mappers.NewProtobufMapper,

		// Database queries
		NewTenantQueries,

		// Database pool
		NewPgxPool,

		// Repository container
		repository.ProvideRepositoryContainer,

		// Unit of work
		NewUnitOfWorkFactory,

		// Service layer
		services.NewTenantService,
		fx.Annotate(
			services.NewTenantService,
			fx.As(new(services.ITenantService)),
		),

		// Handler layer
		handler.NewTenantHandler,
		handler.NewTenantInternalHandler,
		handler.ProvideHandlerContainer,
	),
)

// NewPgxPool creates a new pgx pool from the database manager
func NewPgxPool(dm *sqlc.DatabaseManager) *pgxpool.Pool {
	return dm.Pool
}

// NewTenantQueries creates tenant queries using the database manager
func NewTenantQueries(dbManager *sqlc.DatabaseManager) *sqlcgen.Queries {
	return dbManager.GetTenantQueries()
}

// NewUnitOfWorkFactory creates a new unit of work factory for tenant service
func NewUnitOfWorkFactory(pool *pgxpool.Pool, mapper *mappers.TenantMapper) uow.UnitOfWorkFactory {
	return uow.NewSqlcUnitOfWorkFactory(pool, mapper)
}
