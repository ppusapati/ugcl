package backupdr

import (
	"go.uber.org/fx"

	"p9e.in/ugcl/backupdr/repository"
	"p9e.in/ugcl/backupdr/services"
)

// Module provides the backup and disaster recovery module for dependency injection
var Module = fx.Module("backupdr",
	fx.Provide(
		// Repository layer
		fx.Annotate(
			repository.NewBackupDRRepository,
			fx.As(new(repository.IBackupDRRepository)),
		),

		// Service layer
		services.NewBackupOrchestrator,
		// TODO: Add other services when implemented:
		// services.NewDisasterRecoveryManager,
		// services.NewSystemHealthMonitor,
		// services.NewStorageManager,
	),

	// Optional: Handlers (if gRPC/HTTP handlers are needed)
	// fx.Provide(
	//     handlers.NewBackupHandler,
	//     handlers.NewDisasterRecoveryHandler,
	// ),
)