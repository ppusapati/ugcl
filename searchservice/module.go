package searchservice

import (
	"go.uber.org/fx"

	"p9e.in/ugcl/searchservice/config"
	"p9e.in/ugcl/searchservice/handlers"
	"p9e.in/ugcl/searchservice/indexers"
	"p9e.in/ugcl/searchservice/repository"
	"p9e.in/ugcl/searchservice/services"
	"p9e.in/ugcl/searchservice/services/interfaces"
	"p9e.in/ugcl/searchservice/events"
)

// Module provides the search service dependencies
var Module = fx.Options(
	fx.Provide(
		// Config
		config.NewElasticsearchConfig,
		config.NewIndexingConfig,
		config.NewCacheConfig,

		// Repository
		repository.NewElasticsearchRepository,
		repository.NewCacheRepository,

		// Services
		fx.Annotate(
			services.NewIndexingService,
			fx.As(new(interfaces.IIndexingService)),
		),
		fx.Annotate(
			services.NewQueryService,
			fx.As(new(interfaces.IQueryService)),
		),
		fx.Annotate(
			services.NewSuggestionService,
			fx.As(new(interfaces.ISuggestionService)),
		),
		fx.Annotate(
			services.NewAnalyticsService,
			fx.As(new(interfaces.IAnalyticsService)),
		),
		fx.Annotate(
			services.NewAdminService,
			fx.As(new(interfaces.IAdminService)),
		),

		// Indexers
		indexers.NewDocumentIndexer,
		indexers.NewFormIndexer,
		indexers.NewUserIndexer,
		indexers.NewNotificationIndexer,
		indexers.NewGenericIndexer,

		// Handlers
		handlers.NewSearchHandler,
		handlers.NewAdminHandler,
		handlers.NewAnalyticsHandler,

		// Events
		events.NewIndexingSubscriber,
	),
	fx.Invoke(
		// Initialize event subscribers
		func(subscriber *events.IndexingSubscriber) {
			subscriber.Subscribe()
		},
	),
)