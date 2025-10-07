package dms

import (
	"go.uber.org/fx"

	"p9e.in/ugcl/dms/api/v1/dmsv1connect"
	db "p9e.in/ugcl/dms/db/generated"
	"p9e.in/ugcl/dms/handlers"
	"p9e.in/ugcl/dms/repository"
	"p9e.in/ugcl/dms/services"
	"p9e.in/ugcl/packages/database/sqlc"
)

// NewDMSQueries creates DMS SQLC queries from database manager
func NewDMSQueries(dbManager *sqlc.DatabaseManager) *db.Queries {
	return db.New(dbManager.Pool)
}

// ModuleSQLC provides FX dependency injection for DMS module with SQLC
var ModuleSQLC = fx.Module("dms",
	// Provide SQLC queries
	fx.Provide(NewDMSQueries),

	// Provide repositories
	fx.Provide(
		fx.Annotate(
			repository.NewDocumentRepository,
			fx.As(new(repository.IDocumentRepository)),
		),
		fx.Annotate(
			repository.NewDocumentShareRepository,
			fx.As(new(repository.IDocumentShareRepository)),
		),
	),

	// Provide services
	fx.Provide(
		fx.Annotate(
			services.NewDocumentService,
			fx.As(new(services.IDocumentService)),
		),
		fx.Annotate(
			services.NewDocumentShareService,
			fx.As(new(services.IDocumentShareService)),
		),
		fx.Annotate(
			services.NewWatermarkService,
			fx.As(new(services.IWatermarkService)),
		),
		fx.Annotate(
			services.NewAnalyticsService,
			fx.As(new(services.IAnalyticsService)),
		),
	),

	// Provide handlers
	fx.Provide(
		fx.Annotate(
			handlers.NewDMSHandler,
			fx.As(new(dmsv1connect.DMSServiceHandler)),
		),
	),
)
