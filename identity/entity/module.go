package entity

import (
	"go.uber.org/fx"

	"p9e.in/ugcl/identity/entity/api/v1/entityv1connect"
	db "p9e.in/ugcl/identity/entity/db/generated"
	"p9e.in/ugcl/identity/entity/handlers"
	"p9e.in/ugcl/identity/entity/repository"
	"p9e.in/ugcl/identity/entity/services"
	"p9e.in/ugcl/packages/database/sqlc"
)

// NewEntityQueries creates entity SQLC queries from database manager
func NewEntityQueries(dbManager *sqlc.DatabaseManager) *db.Queries {
	return db.New(dbManager.Pool)
}

// Module provides FX dependency injection for entity module
var Module = fx.Module("entity",
	// Provide SQLC queries
	fx.Provide(NewEntityQueries),

	// Provide repositories
	fx.Provide(
		fx.Annotate(
			repository.NewEntityRepository,
			fx.As(new(repository.IEntityRepository)),
		),
		fx.Annotate(
			repository.NewEntityRoleBindingRepository,
			fx.As(new(repository.IEntityRoleBindingRepository)),
		),
	),

	// Provide services
	fx.Provide(
		fx.Annotate(
			services.NewEntityService,
			fx.As(new(services.IEntityService)),
		),
		fx.Annotate(
			services.NewEntityRoleBindingService,
			fx.As(new(services.IEntityRoleBindingService)),
		),
	),

	// Provide handlers
	fx.Provide(
		fx.Annotate(
			handlers.NewEntityHandler,
			fx.As(new(entityv1connect.EntityServiceHandler)),
		),
	),
)
