package packages

import (
	"go.uber.org/fx"

	"p9e.in/ugcl/packages/database/sqlc"
	db "p9e.in/ugcl/projects/db/generated"
	"p9e.in/ugcl/projects/handlers"
	"p9e.in/ugcl/projects/repository"
	"p9e.in/ugcl/projects/services"
)

// Module bundles all Contractor dependencies
var Module = fx.Module("vendors",
	fx.Provide(
		NewDairySiteQueries,
		// func(pool *pgxpool.Pool) *db.Queries {
		// 	return db.New(pool)
		// },
		fx.Annotate(
			repository.NewDairySiteRepository,
			fx.As(new(repository.IDairySiteRepository)),
		),
		// fx.Annotate(
		// 	uservice.NewUserService, // 👈 constructor for IUserService
		// 	fx.As(new(uservice.IUserService)),
		// ),
		fx.Annotate(
			services.NewDairySiteService,
			fx.As(new(services.IDairySiteService)),
		),
		// fx.Annotate(
		// handlers.NewContractorHandler,
		handlers.NewDairySiteHandler, // optional interface
		// ),
	),
)

// NewContractorQueries creates contractor queries using the database manager
func NewDairySiteQueries(dbManager *sqlc.DatabaseManager) *db.Queries {
	return dbManager.GetDairySiteQueries()
}
