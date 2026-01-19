package dataarchive

import (
	"go.uber.org/fx"

	"p9e.in/ugcl/dataarchive/repository"
	"p9e.in/ugcl/dataarchive/services"
)

// Module provides the data archiving module for dependency injection
var Module = fx.Module("dataarchive",
	// Repository layer
	fx.Provide(
		repository.NewArchiveRepository,
	),

	// Service layer
	fx.Provide(
		services.NewRetentionPolicyManager,
		services.NewArchivalExecutor,
		services.NewComplianceAuditor,
	),

	// Optional: Handlers (if gRPC/HTTP handlers are needed)
	// fx.Provide(
	//     handlers.NewArchiveHandler,
	// ),
)