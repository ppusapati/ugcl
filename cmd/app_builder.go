package main

import (
	"net/http"
	"os"

	"go.uber.org/fx"
	"p9e.in/ugcl/core/config"
	"p9e.in/ugcl/core/middleware"
	formbuilder "p9e.in/ugcl/formbuilder"
	auth "p9e.in/ugcl/identity/auth"
	tenant "p9e.in/ugcl/identity/tenant"
	user "p9e.in/ugcl/identity/user"
	phandlers "p9e.in/ugcl/masters/pipeline/handlers"
	prepo "p9e.in/ugcl/masters/pipeline/repository"
	pservices "p9e.in/ugcl/masters/pipeline/services"
	"p9e.in/ugcl/monitoring"
	"p9e.in/ugcl/notification"
	"p9e.in/ugcl/organization"
	entity "p9e.in/ugcl/identity/entity"
	"p9e.in/ugcl/dms"

	migrations "p9e.in/ugcl/migrations" // commented out for dev
	conf "p9e.in/ugcl/packages/api/v1/config"
	pkgconfig "p9e.in/ugcl/packages/config"
	"p9e.in/ugcl/packages/config/file"
	"p9e.in/ugcl/packages/database/sqlc"
	"p9e.in/ugcl/packages/p9log"
	projects "p9e.in/ugcl/projects"
	vendors "p9e.in/ugcl/vendors"
)

// ApplicationBuilder creates and configures the Fx application
type ApplicationBuilder struct {
	modules []fx.Option
}

// NewApplication creates a new application with all modules
func NewApplication() *fx.App {
	builder := &ApplicationBuilder{}

	return fx.New(
		builder.addConfiguration(),
		// Core infrastructure
		builder.addInfrastructure(),
		// Authentication module
		builder.addAuthenticationModule(),
		// All service modules
		builder.addAllServices(),
		// HTTP layer
		builder.addHTTPLayer(),
		// Lifecycle management
		// builder.addLifecycleManagement(),
	)
}

func (b *ApplicationBuilder) addConfiguration() fx.Option {
	return fx.Options(
		fx.Provide(
			NewConfigLoader,
			NewBootstrapConfig,
			ExtractConfigs,
			NewAuthConfig, // Add auth config
		),
	)
}

// Add authentication module
func (b *ApplicationBuilder) addAuthenticationModule() fx.Option {
	return fx.Module("authentication",
		fx.Provide(middleware.NewAuthService),
		fx.Invoke(func(authService *middleware.AuthService) {
			// Log auth configuration on startup
			authService.LogConfig()
		}),
	)
}

// Add infrastructure modules
func (b *ApplicationBuilder) addInfrastructure() fx.Option {
	return fx.Options(
		config.DatabaseModule,
		fx.Provide(NewSQLCDatabaseManager),
		fx.Provide(NewHTTPMux),
		fx.Provide(NewServiceRegistry), // Add service registry
		migrations.Module,              // Add migrations - commented out for dev (requires Atlas CLI)
	)
}

// Add all service modules in organized groups
func (b *ApplicationBuilder) addAllServices() fx.Option {
	return fx.Options(
		// Project services group
		b.addProjectServices(),

		// Master services group
		b.addMasterServices(),
		b.addIdentityServices(),
		b.addVendorServices(),
		// Utility services group
		// b.addUtilityServices(),
		b.addFormBuilderServices(),
		b.addNotificationServices(),
		b.addOrganizationServices(),
		b.addDMSServices(),
	)
}

// Add project services
func (b *ApplicationBuilder) addProjectServices() fx.Option {
	return fx.Module("project-services",
		// Dairy Site service
		projects.Module,
	)
}

// Add master services
func (b *ApplicationBuilder) addMasterServices() fx.Option {
	return fx.Module("master-services",
		// Pipeline service
		prepo.PipelineRepositoryModule,
		prepo.NodeRepositoryModule,
		prepo.SiteRepositoryModule,
		prepo.ZoneRepositoryModule,

		pservices.PipelineServiceModule,
		pservices.HierarchyServiceModule,
		phandlers.PipelineHandlerModule,
	)
}

func (b *ApplicationBuilder) addIdentityServices() fx.Option {
	return fx.Module("identity-services",
		user.UserModule,
		tenant.TenantModule,
		auth.AuthModule,
		entity.Module,
	)
}

func (b *ApplicationBuilder) addVendorServices() fx.Option {
	return fx.Module("vendor-services",
		// Contractor service
		vendors.Module,
	)
}

func (b *ApplicationBuilder) addFormBuilderServices() fx.Option {
	return fx.Module("formbuilder-services",
		// Formbuilder service
		formbuilder.Module,
	)
}

// Add utility services (logging, monitoring, etc.)
func (b *ApplicationBuilder) addUtilityServices() fx.Option {
	return fx.Module("utility-services",
		// Add monitoring module
		fx.Invoke(monitoring.RegisterMonitoringServices),
	)
}

func (b *ApplicationBuilder) addNotificationServices() fx.Option {
	return fx.Module("notification-services",
		// Notification service
		notification.Module,
	)
}

func (b *ApplicationBuilder) addOrganizationServices() fx.Option {
	return fx.Module("organization-services",
		// Organization service
		organization.Module,
	)
}

func (b *ApplicationBuilder) addDMSServices() fx.Option {
	return fx.Module("dms-services",
		// Document Management System service
		dms.ModuleSQLC,
	)
}

// Add HTTP layer
func (b *ApplicationBuilder) addHTTPLayer() fx.Option {
	return fx.Options(
		fx.Invoke(RegisterAllServices), // Use the updated function
		fx.Invoke(SetupServerLifecycle),
	)
}

// Add lifecycle management
// func (b *ApplicationBuilder) addLifecycleManagement() fx.Option {
// 	return fx.Options(
// 		fx.Invoke(migrations.Module),
// 	)
// }

// Provide HTTP mux
func NewHTTPMux() *http.ServeMux {
	return http.NewServeMux()
}

// NewServiceRegistry creates a new service registry with dependency injection
func NewServiceRegistry(mux *http.ServeMux, authService *middleware.AuthService) *ServiceRegistry {
	return &ServiceRegistry{
		mux:         mux,
		authService: authService,
		services:    []string{},
	}
}

// NewConfigLoader creates a configuration loader from config.yaml
func NewConfigLoader() (pkgconfig.Config, error) {
	c := pkgconfig.New(
		pkgconfig.WithSource(
			file.NewSource("./configs.yaml"),
		),
	)

	if err := c.Load(); err != nil {
		return nil, err
	}

	return c, nil
}

// NewBootstrapConfig creates bootstrap configuration from the loaded config
func NewBootstrapConfig(config pkgconfig.Config) (*conf.Bootstrap, error) {
	var bc conf.Bootstrap
	if err := config.Scan(&bc); err != nil {
		return nil, err
	}
	return &bc, nil
}

// NewAuthConfig creates authentication configuration
func NewAuthConfig(config pkgconfig.Config) (*middleware.Config, error) {
	// Try to extract auth config from main config first
	var authConfig middleware.Config
	if err := config.Scan(&authConfig); err == nil && authConfig.JWT.Secret != "" {
		return &authConfig, nil
	}

	// Fallback to separate auth config file
	return middleware.LoadAuthConfigFromFile("./configs.yaml")
}

// ExtractConfigs extracts individual config sections for easier injection
func ExtractConfigs(bc *conf.Bootstrap) (*conf.Server, *conf.Data, *conf.Observability, p9log.Logger) {

	const thresholdSize = 10 * 1024 * 1024 // 10MB
	logFile := p9log.LoggerFile("../log/", thresholdSize)
	defer logFile.(interface{ Close() error }).Close()

	var id, _ = os.Hostname()
	logger := p9log.With(p9log.NewStdLogger(logFile), "Time", p9log.DefaultTimestamp, "PID", id, "Service", "Version", Version)

	var server *conf.Server
	var data *conf.Data
	var observability *conf.Observability

	if bc.Server != nil {
		server = bc.Server
	} else {
		// Provide defaults
		server = &conf.Server{}
	}

	if bc.Data != nil {
		data = bc.Data
	} else {
		// Provide defaults
		data = &conf.Data{}
	}

	if bc.Observability != nil {
		observability = bc.Observability
	} else {
		// Provide defaults
		observability = &conf.Observability{}
	}

	return server, data, observability, logger
}

func NewSQLCDatabaseManager(data *conf.Data) (*sqlc.DatabaseManager, error) {
	return sqlc.NewDatabaseManager(data)
}
