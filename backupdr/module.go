package backupdr

import (
	"go.uber.org/fx"

	"p9e.in/ugcl/backupdr/repository"
	"p9e.in/ugcl/backupdr/services"
)

// Module provides the backup and disaster recovery module for dependency injection
var Module = fx.Module("backupdr",
	// Repository layer
	fx.Provide(
		repository.NewBackupRepository,
		repository.NewDisasterRecoveryRepository,
	),

	// Service layer
	fx.Provide(
		services.NewBackupOrchestrator,
		services.NewDisasterRecoveryManager,
		services.NewSystemHealthMonitor,
		services.NewStorageManager,
	),

	// Optional: Handlers (if gRPC/HTTP handlers are needed)
	// fx.Provide(
	//     handlers.NewBackupHandler,
	//     handlers.NewDisasterRecoveryHandler,
	// ),
)