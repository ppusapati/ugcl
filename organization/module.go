package organization

import (
	"go.uber.org/fx"

	"p9e.in/ugcl/organization/api/v1/organization/organizationconnect"
	db "p9e.in/ugcl/organization/db/generated"
	"p9e.in/ugcl/organization/handlers"
	"p9e.in/ugcl/organization/repository"
	"p9e.in/ugcl/organization/services"
	"p9e.in/ugcl/packages/database/sqlc"
)

// Module bundles all Organization dependencies
var Module = fx.Module("organization",
	// Provide database queries
	fx.Provide(NewOrganizationQueries),

	// Provide repositories
	fx.Provide(
		fx.Annotate(
			repository.NewDivisionRepository,
			fx.As(new(repository.IDivisionRepository)),
		),
		fx.Annotate(
			repository.NewBranchRepository,
			fx.As(new(repository.IBranchRepository)),
		),
		fx.Annotate(
			repository.NewDepartmentRepository,
			fx.As(new(repository.IDepartmentRepository)),
		),
		fx.Annotate(
			repository.NewOrganizationRepository,
			fx.As(new(repository.IOrganizationRepository)),
		),
	),

	// Provide services
	fx.Provide(
		fx.Annotate(
			services.NewDivisionService,
			fx.As(new(services.IDivisionService)),
		),
		fx.Annotate(
			services.NewBranchService,
			fx.As(new(services.IBranchService)),
		),
		fx.Annotate(
			services.NewDepartmentService,
			fx.As(new(services.IDepartmentService)),
		),
	),

	// Provide handlers
	fx.Provide(
		fx.Annotate(
			handlers.NewOrganizationHandler,
			fx.As(new(organizationconnect.OrganizationServiceHandler)),
		),
	),
)

// NewOrganizationQueries creates organization queries using the database manager
func NewOrganizationQueries(dbManager *sqlc.DatabaseManager) *db.Queries {
	return dbManager.GetOrganizationQueries()
}
