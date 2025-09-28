package approvalworkflow

import (
	"go.uber.org/fx"

	"p9e.in/ugcl/approvalworkflow/repository"
	"p9e.in/ugcl/approvalworkflow/services"
)

// Module provides the approval workflow module for dependency injection
var Module = fx.Module("approvalworkflow",
	// Repository layer
	fx.Provide(
		repository.NewApprovalRepository,
	),

	// Service layer
	fx.Provide(
		services.NewApprovalEngine,
	),

	// Optional: Handlers (if gRPC/HTTP handlers are needed)
	// fx.Provide(
	//     handlers.NewApprovalHandler,
	// ),
)