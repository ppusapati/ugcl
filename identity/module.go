package user

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"

	db "p9e.in/ugcl/identity/db/sqlc/generated"
	"p9e.in/ugcl/identity/handler"
	"p9e.in/ugcl/identity/repository"
	"p9e.in/ugcl/identity/services"
	"p9e.in/ugcl/identity/uow"
	"p9e.in/ugcl/packages/database/sqlc"
)

// Module bundles all Contractor dependencies
var UserModule = fx.Module("user",
	fx.Provide(
		NewUserQueries,
		// func(pool *pgxpool.Pool) *db.Queries {
		// 	return db.New(pool)
		// },
		fx.Annotate(
			repository.ProvideRepositoryContainer,
		),

		fx.Annotate(
			services.NewUserService,
			fx.As(new(services.IUserService)),
		),

		// fx.Annotate(
		// handlers.NewContractorHandler,
		handler.NewUserHandler, // optional interface
		// ),
		NewPgxPool,
	),

	uow.SqlcUOWModule,
)

func NewPgxPool(dm *sqlc.DatabaseManager) *pgxpool.Pool {
	return dm.Pool
}

// NewContractorQueries creates contractor queries using the database manager
func NewUserQueries(dbManager *sqlc.DatabaseManager) *db.Queries {
	return dbManager.GetUserQueries()
}

func NewUnitOfWorkFactory(dbManager *sqlc.DatabaseManager) uow.UnitOfWorkFactory {
	return dbManager.GetUserUOW()
}
