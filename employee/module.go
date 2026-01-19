package employee

import (
	"go.uber.org/fx"

	"p9e.in/ugcl/employee/db/generated"
	"p9e.in/ugcl/employee/handlers"
	"p9e.in/ugcl/employee/repository"
	"p9e.in/ugcl/employee/services"
	"p9e.in/ugcl/packages/database/sqlc"
)

// Module bundles all Employee dependencies
var Module = fx.Module("employee",
	fx.Provide(
		NewEmployeeQueries,
		fx.Annotate(
			repository.NewEmployeeRepository,
			fx.As(new(repository.IEmployeeRepository)),
		),
		fx.Annotate(
			services.NewEmployeeService,
			fx.As(new(services.IEmployeeService)),
		),
		handlers.NewEmployeeHandler,
	),
)

// NewEmployeeQueries creates employee queries using the database manager
func NewEmployeeQueries(dbManager *sqlc.DatabaseManager) *generated.Queries {
	return dbManager.GetEmployeeQueries()
}
