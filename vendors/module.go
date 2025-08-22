package vendors

import (
	"go.uber.org/fx"

	"p9e.in/ugcl/packages/database/sqlc"
	db "p9e.in/ugcl/vendors/db/generated"
	"p9e.in/ugcl/vendors/handlers"
	"p9e.in/ugcl/vendors/repository"
	"p9e.in/ugcl/vendors/services"
)

// Module bundles all Contractor dependencies
var Module = fx.Module("vendors",
	fx.Provide(
		NewContractorQueries,
		// func(pool *pgxpool.Pool) *db.Queries {
		// 	return db.New(pool)
		// },
		fx.Annotate(
			repository.NewContractorRepository,
			fx.As(new(repository.IContractorRepository)),
		),
		// fx.Annotate(
		// 	uservice.NewUserService, // 👈 constructor for IUserService
		// 	fx.As(new(uservice.IUserService)),
		// ),
		fx.Annotate(
			services.NewContractorService,
			fx.As(new(services.IContractorService)),
		),
		// fx.Annotate(
		// handlers.NewContractorHandler,
		handlers.NewContractorHandler, // optional interface
		// ),
	),
)

// NewContractorQueries creates contractor queries using the database manager
func NewContractorQueries(dbManager *sqlc.DatabaseManager) *db.Queries {
	return dbManager.GetContractorQueries()
}
