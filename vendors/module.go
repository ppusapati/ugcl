package vendors

import (
	"go.uber.org/fx"

	"p9e.in/ugcl/packages/database/sqlc"
	db "p9e.in/ugcl/vendors/db/generated"
	"p9e.in/ugcl/vendors/handlers"
	"p9e.in/ugcl/vendors/repository"
	"p9e.in/ugcl/vendors/services"
)

// Module bundles all Vendor dependencies
var Module = fx.Module("vendors",
	fx.Provide(
		NewVendorQueries,
		fx.Annotate(
			repository.NewVendorRepository,
			fx.As(new(repository.IVendorRepository)),
		),
		fx.Annotate(
			services.NewVendorService,
			fx.As(new(services.IVendorService)),
		),
		handlers.NewVendorHandler,
	),
)

// NewVendorQueries creates vendor queries using the database manager
func NewVendorQueries(dbManager *sqlc.DatabaseManager) *db.Queries {
	return dbManager.GetVendorQueries()
}
