package insightviewer

import (
	"context"
	"database/sql"

	"p9e.in/ugcl/insightviewer/db/generated"
	"p9e.in/ugcl/insightviewer/handlers"
	"p9e.in/ugcl/insightviewer/services"
)

type Module struct {
	db          *sql.DB
	queries     *generated.Queries
	services    *services.ServiceManager
	handlers    *handlers.ViewerHandler
}

func NewModule(db *sql.DB) *Module {
	// Create the viewer service directly
	viewerService := services.NewViewerService(db)
	serviceManager := services.NewServiceManager(viewerService)
	viewerHandler := handlers.NewViewerHandler(serviceManager)

	return &Module{
		db:       db,
		queries:  generated.New(db),
		services: serviceManager,
		handlers: viewerHandler,
	}
}

func (m *Module) GetHandler() *handlers.ViewerHandler {
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