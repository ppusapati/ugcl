package insighthub

import (
	"context"
	"database/sql"

	"p9e.in/ugcl/insighthub/db/generated"
	"p9e.in/ugcl/insighthub/handlers"
	"p9e.in/ugcl/insighthub/services"
)

type Module struct {
	db          *sql.DB
	queries     *generated.Queries
	services    *services.ServiceManager
	handlers    *handlers.ReportHandler
}

func NewModule(db *sql.DB) *Module {
	// Create the report service directly
	reportService := services.NewReportService(db)
	serviceManager := services.NewServiceManager(reportService)
	reportHandler := handlers.NewReportHandler(serviceManager)

	return &Module{
		db:       db,
		queries:  generated.New(db),
		services: serviceManager,
		handlers: reportHandler,
	}
}

func (m *Module) GetHandler() *handlers.ReportHandler {
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