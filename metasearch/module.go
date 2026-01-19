package metasearch

import (
	"go.uber.org/fx"

	"p9e.in/ugcl/metasearch/config"
	"p9e.in/ugcl/metasearch/handlers"
	"p9e.in/ugcl/metasearch/indexers"
	"p9e.in/ugcl/metasearch/repository"
	"p9e.in/ugcl/metasearch/services"
	"p9e.in/ugcl/metasearch/services/interfaces"
)

// Module provides the search service dependencies
var Module = fx.Options(
	fx.Provide(
		// Config
		config.NewElasticsearchConfig,

		// Repository
		repository.NewElasticsearchRepository,

		// Services
		fx.Annotate(
			services.NewIndexingService,
			fx.As(new(interfaces.IIndexingService)),
		),

		// Indexers
		indexers.NewDocumentIndexer,

		// Handlers
		handlers.NewSearchHandler,
	),
)
