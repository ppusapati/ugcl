package masters

import (
	"context"
	"database/sql"

	"p9e.in/ugcl/masters/db/generated"
	"p9e.in/ugcl/masters/handlers"
	"p9e.in/ugcl/masters/services"
)

type Module struct {
	db          *sql.DB
	queries     *generated.Queries
	services    *services.ServiceManager
	handlers    *handlers.MastersHandler
}

func NewModule(db *sql.DB) *Module {
	// Create the metadata service directly
	metadataService := services.NewMetadataService(db)
	serviceManager := services.NewServiceManager(metadataService)
	mastersHandler := handlers.NewMastersHandler(serviceManager)

	return &Module{
		db:       db,
		queries:  generated.New(db),
		services: serviceManager,
		handlers: mastersHandler,
	}
}

func (m *Module) GetHandler() *handlers.MastersHandler {
	return m.handlers
}

func (m *Module) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

func (m *Module) HealthCheck(ctx context.Context) error {
	return m.db.PingContext(ctx)
}