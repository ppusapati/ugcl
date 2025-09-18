package notification

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"

	db "p9e.in/ugcl/notification/db/generated"
	"p9e.in/ugcl/notification/handlers"
	"p9e.in/ugcl/notification/repository"
	"p9e.in/ugcl/notification/services"
	"p9e.in/ugcl/packages/database/sqlc"
	"p9e.in/ugcl/packages/events/domain"
	"p9e.in/ugcl/packages/p9log"
)

// Module bundles all notification service dependencies
var Module = fx.Module("notification",
	fx.Provide(
		// Database queries
		fx.Annotate(
			func(dbManager *sqlc.DatabaseManager) *db.Queries {
				return dbManager.GetNotificationQueries()
			},
			fx.ResultTags(`name:"notificationQueries"`),
		),

		// Repository layer
		fx.Annotate(
			repository.NewNotificationRepository,
			fx.As(new(repository.INotificationRepository)),
		),

		// Service layer
		fx.Annotate(
			services.NewNotificationService,
			fx.As(new(services.INotificationService)),
		),

		// Handler layer
		handlers.NewNotificationHandler,

		// Event handler for domain events
		handlers.NewNotificationEventHandler,

		// Database pool
		// NewPgxPool,
	),
)

// NewNotificationQueries creates notification queries using the database manager
func NewNotificationQueries(dbManager *sqlc.DatabaseManager) *db.Queries {
	return dbManager.GetNotificationQueries()
}

// NewPgxPool creates a PostgreSQL connection pool
func NewPgxPool(dm *sqlc.DatabaseManager) *pgxpool.Pool {
	return dm.Pool
}

// NotificationHandlers represents all notification handlers
type NotificationHandlers struct {
	NotificationHandler *handlers.NotificationHandler
	EventHandler        *handlers.NotificationEventHandler
}

// NewNotificationHandlers creates a struct containing all notification handlers
func NewNotificationHandlers(
	notificationHandler *handlers.NotificationHandler,
	eventHandler *handlers.NotificationEventHandler,
) *NotificationHandlers {
	return &NotificationHandlers{
		NotificationHandler: notificationHandler,
		EventHandler:        eventHandler,
	}
}

// NotificationServiceParams defines parameters for creating notification service
type NotificationServiceParams struct {
	fx.In

	Repository repository.INotificationRepository
	Publisher  *domain.DomainEventPublisher
	Logger     p9log.Logger
}

// NewNotificationServiceWithParams creates notification service with all dependencies
func NewNotificationServiceWithParams(params NotificationServiceParams) services.INotificationService {
	return services.NewNotificationService(
		params.Repository,
		params.Publisher,
		params.Logger,
	)
}
