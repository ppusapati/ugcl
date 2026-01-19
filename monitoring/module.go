package monitoring

import (
	"context"

	"go.uber.org/fx"
	"p9e.in/ugcl/monitoring/config"
	"p9e.in/ugcl/monitoring/services"
)

// Module provides the monitoring module for FX dependency injection
var Module = fx.Module("monitoring",
	fx.Provide(
		// Configuration
		config.DefaultMonitoringConfig,
		fx.Annotate(
			func(cfg *config.MonitoringConfiguration) services.MonitoringConfig {
				return cfg
			},
			fx.As(new(services.MonitoringConfig)),
		),

		// SLA Monitor Service
		fx.Annotate(
			services.NewSLAMonitorService,
			fx.As(new(services.ISLAMonitorService)),
		),

		// Future services can be added here:
		// fx.Annotate(
		//     services.NewMetricsMonitorService,
		//     fx.As(new(services.IMetricsMonitorService)),
		// ),
		// fx.Annotate(
		//     services.NewHealthMonitorService,
		//     fx.As(new(services.IHealthMonitorService)),
		// ),
		// fx.Annotate(
		//     services.NewAlertManagerService,
		//     fx.As(new(services.IAlertManagerService)),
		// ),
	),

	// Lifecycle management for monitoring services
	fx.Invoke(registerMonitoringLifecycle),
)

// registerMonitoringLifecycle registers lifecycle hooks for all monitoring services
func registerMonitoringLifecycle(
	lc fx.Lifecycle,
	slaMonitor services.ISLAMonitorService,
	config services.MonitoringConfig,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			slaConfig := config.GetSLAConfig()
			if !slaConfig.Enabled {
				return nil
			}

			// Configure SLA monitor
			slaMonitor.SetMonitoringInterval(slaConfig.CheckInterval)

			// Start SLA monitoring
			return slaMonitor.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return slaMonitor.Stop(ctx)
		},
	})
}

// MonitoringServices aggregates all monitoring services for easy access
type MonitoringServices struct {
	SLAMonitor services.ISLAMonitorService
	// Future services:
	// MetricsMonitor services.IMetricsMonitorService
	// HealthMonitor  services.IHealthMonitorService
	// AlertManager   services.IAlertManagerService
}

// NewMonitoringServices creates an aggregate of all monitoring services
func NewMonitoringServices(
	slaMonitor services.ISLAMonitorService,
) *MonitoringServices {
	return &MonitoringServices{
		SLAMonitor: slaMonitor,
	}
}
