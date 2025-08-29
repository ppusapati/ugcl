// =============================================================================
// internal/handlers/handlers.go - Handler registration for FX
// =============================================================================
package handlers

import (
	"go.uber.org/fx"
	db "p9e.in/ugcl/formbuilder/db/generated"
	"p9e.in/ugcl/formbuilder/handlers"
	"p9e.in/ugcl/formbuilder/repository"
	"p9e.in/ugcl/formbuilder/services"
	"p9e.in/ugcl/packages/database/sqlc"
)

// HandlersModule provides all handlers for FX dependency injection
var Module = fx.Module("formbuilder",
	fx.Provide(
		// Provide named instances of *db.Queries for different components
		fx.Annotate(
			func(dbManager *sqlc.DatabaseManager) *db.Queries {
				return dbManager.GetFormBuilderQueries()
			},
			fx.ResultTags(`name:"formBuilderQueries"`),
		),
		fx.Annotate(
			func(dbManager *sqlc.DatabaseManager) *db.Queries {
				return dbManager.GetFormInstanceQueries()
			},
			fx.ResultTags(`name:"formInstanceQueries"`),
		),
		fx.Annotate(
			func(dbManager *sqlc.DatabaseManager) *db.Queries {
				return dbManager.GetWorkflowQueries()
			},
			fx.ResultTags(`name:"workflowQueries"`),
		),

		// Form builder handler
		fx.Annotate(
			func(queries *db.Queries, dbManager *sqlc.DatabaseManager) repository.IFormBuilderRepository {
				return repository.NewFormBuilderRepository(queries, dbManager)
			},
			fx.ParamTags(`name:"formBuilderQueries"`, ``),
		),
		fx.Annotate(
			repository.NewFormInstanceRepository,
			fx.ParamTags(`name:"formInstanceQueries"`),
		),
		fx.Annotate(
			repository.NewWorkflowRepository,
			fx.ParamTags(`name:"workflowQueries"`),
		),

		// Form instance handler
		fx.Annotate(
			services.NewFormBuilderService,
			fx.As(new(services.IFormBuilderService)),
		),
		fx.Annotate(
			services.NewFormInstanceService,
			fx.As(new(services.IFormInstanceService)),
		),
		fx.Annotate(
			services.NewWorkflowService,
			fx.As(new(services.IWorkflowService)),
		),
		handlers.NewFormBuilderHandler,
		handlers.NewFormInstanceHandler,
		handlers.NewWorkflowHandler,
		// NewAllHandlers,
		// NewPgxPool,
	),
)

// AllHandlers represents all gRPC handlers
type AllHandlers struct {
	FormBuilder  *handlers.FormBuilderHandler
	FormInstance *handlers.FormInstanceHandler
	Workflow     *handlers.WorkflowHandler
}

// NewAllHandlers creates a struct containing all handlers for server registration
func NewAllHandlers(
	formBuilder *handlers.FormBuilderHandler,
	formInstance *handlers.FormInstanceHandler,
	workflow *handlers.WorkflowHandler,
) *AllHandlers {
	return &AllHandlers{
		FormBuilder:  formBuilder,
		FormInstance: formInstance,
		Workflow:     workflow,
	}
}
