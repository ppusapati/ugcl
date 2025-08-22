package uow

import "go.uber.org/fx"

// var UOWModule = fx.Module("uow",
// 	// Provide individual factories
// 	fx.Provide(

// 		fx.Annotate(
// 			NewSQLCUnitOfWorkFactory,
// 			fx.ResultTags(`group:"uow_factories"`),
// 		),
// 	),
// )

var SqlcUOWModule = fx.Module("sqlc_uow",
	fx.Provide(
		fx.Annotate(
			NewSQLCUnitOfWorkFactory,
			fx.As(new(UnitOfWorkFactory)),
		),
	),
)
