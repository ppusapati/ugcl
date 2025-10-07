package dms

import (
	"go.uber.org/fx"

	"p9e.in/ugcl/dms/config"
	"p9e.in/ugcl/dms/events"
	"p9e.in/ugcl/dms/handlers"
	"p9e.in/ugcl/dms/repository"
	"p9e.in/ugcl/dms/services"
	"p9e.in/ugcl/dms/services/interfaces"
	"p9e.in/ugcl/dms/utils/compression"
	"p9e.in/ugcl/dms/utils/processing"
	"p9e.in/ugcl/dms/utils/security"
	"p9e.in/ugcl/dms/utils/watermark"
)

// Module provides the document viewer dependencies
var Module = fx.Options(
	fx.Provide(
		// Config
		config.NewDocumentConfig,
		config.NewStorageConfig,
		config.NewProcessingConfig,
		config.NewWatermarkConfig,

		// Utilities
		compression.NewTarZstdCompressor,
		processing.NewPDFProcessor,
		processing.NewImageProcessor,
		processing.NewOCRProcessor,
		watermark.NewWatermarkGenerator,
		security.NewVirusScanner,
		security.NewAccessController,

		// Repository
		repository.NewDocumentRepository,
		repository.NewStorageRepository,
		repository.NewCacheRepository,

		// Services
		fx.Annotate(
			services.NewDocumentService,
			fx.As(new(interfaces.IDocumentService)),
		),
		fx.Annotate(
			services.NewCompressionService,
			fx.As(new(interfaces.ICompressionService)),
		),
		fx.Annotate(
			services.NewWatermarkService,
			fx.As(new(interfaces.IWatermarkService)),
		),
		fx.Annotate(
			services.NewProcessingService,
			fx.As(new(interfaces.IProcessingService)),
		),
		fx.Annotate(
			services.NewPreviewService,
			fx.As(new(interfaces.IPreviewService)),
		),
		fx.Annotate(
			services.NewOCRService,
			fx.As(new(interfaces.IOCRService)),
		),
		fx.Annotate(
			services.NewStorageService,
			fx.As(new(interfaces.IStorageService)),
		),

		// Handlers
		handlers.NewDocumentHandler,
		handlers.NewPreviewHandler,
		handlers.NewWatermarkHandler,
		handlers.NewUploadHandler,
		handlers.NewDownloadHandler,

		// Events
		events.NewDocumentEventPublisher,
		events.NewProcessingEventSubscriber,
	),
	fx.Invoke(
		// Initialize event subscribers
		func(subscriber *events.ProcessingEventSubscriber) {
			subscriber.Subscribe()
		},
		// Initialize background processors
		func(processingService interfaces.IProcessingService) {
			// Start background processing workers
		},
	),
)