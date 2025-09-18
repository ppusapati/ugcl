package monitoring

import (
	"context"

	"go.uber.org/fx"
	"p9e.in/ugcl/monitoring/config"
	"p9e.in/ugcl/monitoring/services"
	"p9e.in/ugcl/packages/events/domain"
)

// registerMonitoringServices registers all monitoring services with the application
func RegisterMonitoringServices(
	lc fx.Lifecycle,
	subscriber *domain.EventSubscriber,
) {
	// Create monitoring configuration
	monitoringConfig := config.DefaultMonitoringConfig()

	// Create SLA monitor service (event-sourced)
	slaMonitor := services.NewSLAMonitorService(subscriber)

	// Configure SLA monitor
	slaConfig := monitoringConfig.GetSLAConfig()
	if slaConfig.Enabled {
		slaMonitor.SetMonitoringInterval(slaConfig.CheckInterval)
	}

	// Register lifecycle hooks
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if !slaConfig.Enabled {
				return nil
			}
			return slaMonitor.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return slaMonitor.Stop(ctx)
		},
	})
}
