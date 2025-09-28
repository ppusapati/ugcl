// =============================================================================
// databridge/module.go - DataBridge module with dependency injection (Import Only)
// =============================================================================
package databridge

import (
	"github.com/jackc/pgx/v5"
	"go.uber.org/fx"

	db "p9e.in/ugcl/databridge/db/generated"
	"p9e.in/ugcl/databridge/handlers"
	"p9e.in/ugcl/databridge/mappers"
	"p9e.in/ugcl/databridge/repository"
	"p9e.in/ugcl/databridge/services"
	"p9e.in/ugcl/packages/database/sqlc"
	"p9e.in/ugcl/packages/p9log"
)

// Module bundles all databridge service dependencies (import-focused)
var Module = fx.Module("databridge",
	// Include mapper module
	mappers.MapperModule,

	fx.Provide(
		// Database queries
		fx.Annotate(
			func(dbManager *sqlc.DatabaseManager) *db.Queries {
				return db.New(dbManager.Pool)
			},
			fx.ResultTags(`name:"databridgeQueries"`),
		),

		// Database connection
		fx.Annotate(
			func(dbManager *sqlc.DatabaseManager) *pgx.Conn {
				// Note: In production, you might want to get a connection from the pool
				// For simplicity, we'll use a single connection approach
				conn, err := dbManager.Pool.Acquire(nil)
				if err != nil {
					panic(err)
				}
				return conn.Conn()
			},
			fx.ResultTags(`name:"databridgeConnection"`),
		),

		// Repository layer (import only)
		fx.Annotate(
			func(conn *pgx.Conn) repository.IMappingRepository {
				return repository.NewMappingRepository(conn)
			},
			fx.ParamTags(`name:"databridgeConnection"`),
		),
		fx.Annotate(
			func(conn *pgx.Conn) repository.IJobRepository {
				return repository.NewJobRepository(conn)
			},
			fx.ParamTags(`name:"databridgeConnection"`),
		),
		fx.Annotate(
			func(conn *pgx.Conn) *repository.RepositoryManager {
				return repository.NewRepositoryManager(conn)
			},
			fx.ParamTags(`name:"databridgeConnection"`),
		),

		// Service layer (import-focused)
		fx.Annotate(
			services.NewValidationService,
			fx.As(new(services.IValidationService)),
		),
		fx.Annotate(
			services.NewFileService,
			fx.As(new(services.IFileService)),
		),
		fx.Annotate(
			NewCSVService,
			fx.As(new(services.ICSVService)),
		),
		fx.Annotate(
			NewMappingService,
			fx.As(new(services.IMappingService)),
		),
		fx.Annotate(
			NewImportService,
			fx.As(new(services.IImportService)),
		),
		NewServiceManager,

		// Handler layer
		NewDataBridgeHandler,
	),
)

// Service constructor functions with proper dependency injection

// NewCSVService creates CSV service with dependencies
func NewCSVService(fileService services.IFileService, validationService services.IValidationService) services.ICSVService {
	return services.NewCSVService(fileService, validationService)
}

// NewMappingService creates mapping service with dependencies
func NewMappingService(repos *repository.RepositoryManager, mappers *mappers.MapperManager) services.IMappingService {
	return services.NewMappingService(repos, mappers)
}

// NewImportService creates import service with dependencies
func NewImportService(repos *repository.RepositoryManager, mappers *mappers.MapperManager, csvService services.ICSVService, validationService services.IValidationService) services.IImportService {
	return services.NewImportService(repos, mappers, csvService, validationService)
}

// NewServiceManager creates service manager with import-only dependencies
func NewServiceManager(
	mappingService services.IMappingService,
	csvService services.ICSVService,
	importService services.IImportService,
	validationService services.IValidationService,
	fileService services.IFileService,
) *services.ServiceManager {
	return &services.ServiceManager{
		Mapping:    mappingService,
		CSV:        csvService,
		Import:     importService,
		Validation: validationService,
		File:       fileService,
	}
}

// NewDataBridgeHandler creates the main handler with all dependencies
func NewDataBridgeHandler(serviceManager *services.ServiceManager, logger p9log.Logger) *handlers.DataBridgeHandler {
	return handlers.NewDataBridgeHandler(serviceManager, logger)
}

// DataBridgeHandlers represents all databridge handlers
type DataBridgeHandlers struct {
	DataBridgeHandler *handlers.DataBridgeHandler
}

// NewDataBridgeHandlers creates a struct containing all databridge handlers
func NewDataBridgeHandlers(databridgeHandler *handlers.DataBridgeHandler) *DataBridgeHandlers {
	return &DataBridgeHandlers{
		DataBridgeHandler: databridgeHandler,
	}
}

// DataBridgeServiceParams defines parameters for creating databridge service
type DataBridgeServiceParams struct {
	fx.In

	RepositoryManager *repository.RepositoryManager
	MapperManager     *mappers.MapperManager
	Logger            p9log.Logger
}

// NewDataBridgeServiceWithParams creates databridge service with all dependencies (import-focused)
func NewDataBridgeServiceWithParams(params DataBridgeServiceParams) *services.ServiceManager {
	// Create individual services
	validationService := services.NewValidationService()
	fileService := services.NewFileService()
	csvService := services.NewCSVService(fileService, validationService)
	mappingService := services.NewMappingService(params.RepositoryManager, params.MapperManager)
	importService := services.NewImportService(params.RepositoryManager, params.MapperManager, csvService, validationService)

	return &services.ServiceManager{
		Mapping:    mappingService,
		CSV:        csvService,
		Import:     importService,
		Validation: validationService,
		File:       fileService,
	}
}