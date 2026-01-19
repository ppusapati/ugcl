# Command (CMD) Module

## 1. Module Overview

The CMD module serves as the main entry point and application bootstrap layer for the UGCL platform. It orchestrates the initialization, configuration, and assembly of all microservices using Uber FX dependency injection, providing a unified server that hosts all modules as a modular monolith.

**Purpose:** Bootstrap and run the UGCL application by coordinating all modules, managing their dependencies, lifecycle, and HTTP/gRPC server initialization.

**Key Features:**
- Application bootstrap and initialization
- Module registry and dynamic module loading
- Uber FX-based dependency injection
- HTTP/gRPC server factory and configuration
- Configuration management and loading
- Service discovery and registration
- Graceful shutdown handling
- Database connection management
- Migration execution
- Health check endpoints

## 2. Architecture

### Module Structure
```
cmd/
├── main.go               # Application entry point
├── app_builder.go        # FX application builder
├── module_registry.go    # Module registration and management
├── server_factory.go     # HTTP/gRPC server creation
└── go.mod                # Dependencies
```

### Component Responsibilities

**main.go:**
- Entry point of the application
- Loads environment variables
- Initializes logger
- Creates and runs FX application

**app_builder.go:**
- Defines ApplicationBuilder structure
- Assembles all modules using FX
- Configures infrastructure (database, cache, logging)
- Sets up HTTP/gRPC servers
- Manages application lifecycle

**module_registry.go:**
- Registers all available modules
- Provides module discovery
- Enables/disables modules based on configuration
- Manages module dependencies
- Module health checks

**server_factory.go:**
- Creates HTTP server with Echo framework
- Configures gRPC server
- Sets up Connect-RPC handlers
- Registers middleware
- Configures routes

## 3. Quick Start

### Build and Run

```bash
# Build the application
go build -o ugcl cmd/main.go

# Run the application
./ugcl

# Or run directly
go run cmd/main.go
```

### Environment Setup

```bash
# Copy environment template
cp .env.example .env

# Edit configuration
nano .env
```

### Configuration

```env
# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=ugcl
DB_USER=postgres
DB_PASSWORD=password

# Modules to enable (comma-separated)
ENABLED_MODULES=identity,formbuilder,organization,dms,vendors,projects

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

## 4. Application Bootstrap Flow

### Startup Sequence

```
1. main.go executes
   ↓
2. Load environment variables (.env)
   ↓
3. Initialize logger (p9log)
   ↓
4. Create FX application (app_builder.go)
   ↓
5. Register modules (module_registry.go)
   ↓
6. Initialize infrastructure
   - Database connections
   - Redis cache
   - Message queues
   ↓
7. Run database migrations
   ↓
8. Initialize all module services
   - Identity (user, auth, tenant, entity)
   - Formbuilder
   - Organization
   - DMS
   - Vendors
   - Projects
   - etc.
   ↓
9. Create HTTP/gRPC servers (server_factory.go)
   ↓
10. Register routes and handlers
   ↓
11. Start servers
   ↓
12. Application running ✓
```

### FX Dependency Graph

```
Configuration
    ↓
Infrastructure (DB, Cache, Logger)
    ↓
Repositories
    ↓
Services
    ↓
Handlers
    ↓
HTTP Server
    ↓
Application Running
```

## 5. Module Registry

### Registered Modules

The module registry manages the following modules:

**Identity Services:**
- `identity/user` - User management
- `identity/auth` - Authentication and sessions
- `identity/tenant` - Multi-tenancy
- `identity/entity` - Entity abstraction

**Domain Services:**
- `formbuilder` - Dynamic form creation
- `organization` - Organizational hierarchy
- `dms` - Document management
- `vendors` - Vendor management
- `contractors` - Contractor management
- `employee` - Employee management
- `projects` - Project management

**Support Services:**
- `notification` - Notifications
- `monitoring` - System monitoring
- `scheduler` - Job scheduling
- `metasearch` - Cross-module search

**Data Services:**
- `databridge` - Data import/export
- `dataarchive` - Data archiving
- `backupdr` - Backup and disaster recovery

**Analytics Services:**
- `insighthub` - Analytics engine
- `insightviewer` - Report viewer

### Module Interface

```go
type Module interface {
    // Name returns the module identifier
    Name() string

    // Priority returns module load order (lower = earlier)
    Priority() int

    // Dependencies returns required modules
    Dependencies() []string

    // FXModule returns the FX module option
    FXModule() fx.Option

    // HealthCheck verifies module health
    HealthCheck() bool
}
```

### Adding New Module

```go
// 1. Create module in module_registry.go
func (b *ApplicationBuilder) addCustomModule() fx.Option {
    return fx.Module("custom",
        customModule.Module,
        fx.Invoke(func(srv *CustomService) {
            log.Info("Custom module initialized")
        }),
    )
}

// 2. Add to addAllServices() in app_builder.go
func (b *ApplicationBuilder) addAllServices() fx.Option {
    return fx.Options(
        b.addProjectServices(),
        b.addMasterServices(),
        b.addCustomModule(),  // Add here
    )
}

// 3. Register routes if needed
func (r *ServiceRegistry) RegisterCustomRoutes(mux *http.ServeMux) {
    customHandler := // ... get from DI
    mux.Handle("/api/v1/custom/", customHandler)
}
```

## 6. Server Configuration

### HTTP Server

```go
// server_factory.go
func NewHTTPServer(cfg *config.ServerConfig) *echo.Echo {
    e := echo.New()

    // Middleware
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())
    e.Use(middleware.CORS())
    e.Use(middleware.RequestID())

    // Health check
    e.GET("/health", healthCheckHandler)

    return e
}
```

### gRPC Server

```go
func NewGRPCServer(cfg *config.ServerConfig) *grpc.Server {
    opts := []grpc.ServerOption{
        grpc.UnaryInterceptor(
            grpc_middleware.ChainUnaryServer(
                grpc_recovery.UnaryServerInterceptor(),
                grpc_auth.UnaryServerInterceptor(authFunc),
                grpc_prometheus.UnaryServerInterceptor,
            ),
        ),
    }

    return grpc.NewServer(opts...)
}
```

### Connect-RPC Integration

```go
// Register Connect handlers
func RegisterConnectHandlers(mux *http.ServeMux, services *ServiceRegistry) {
    // User service
    path, handler := userv1connect.NewUserServiceHandler(services.UserService)
    mux.Handle(path, handler)

    // Auth service
    path, handler = authv1connect.NewAuthServiceHandler(services.AuthService)
    mux.Handle(path, handler)

    // Form builder service
    path, handler = formv1connect.NewFormServiceHandler(services.FormService)
    mux.Handle(path, handler)

    // ... register all module handlers
}
```

## 7. Database Management

### Connection Initialization

```go
// Provide database for all modules
func NewSQLCDatabaseManager(configs *ExtractedConfigs) (*sqlc.DatabaseManager, error) {
    manager := sqlc.NewDatabaseManager()

    // Register databases for each module
    manager.Register("identity", configs.Identity.Database)
    manager.Register("formbuilder", configs.FormBuilder.Database)
    manager.Register("organization", configs.Organization.Database)
    // ... register all module databases

    return manager, nil
}
```

### Migration Execution

```go
// Run migrations on startup
fx.Invoke(func(migrator *migrations.Migrator) error {
    log.Info("Running database migrations...")

    if err := migrator.MigrateAll(); err != nil {
        log.Error("Migration failed", "error", err)
        return err
    }

    log.Info("Migrations completed successfully")
    return nil
})
```

## 8. Graceful Shutdown

### Shutdown Sequence

```go
func (a *Application) Shutdown(ctx context.Context) error {
    log.Info("Starting graceful shutdown...")

    // 1. Stop accepting new requests
    a.httpServer.Shutdown(ctx)

    // 2. Close database connections
    a.dbManager.CloseAll()

    // 3. Close cache connections
    a.cache.Close()

    // 4. Flush metrics
    a.metrics.Flush()

    // 5. Close event bus
    a.eventBus.Close()

    log.Info("Graceful shutdown completed")
    return nil
}
```

### Signal Handling

```go
// main.go
func main() {
    app := NewApplication()

    // Handle shutdown signals
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-sigChan
        log.Info("Shutdown signal received")
        app.Stop()
    }()

    app.Run()
}
```

## 9. Health Checks

### Health Check Endpoint

```
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-10-06T10:30:00Z",
  "uptime": "2h15m30s",
  "modules": {
    "identity": "healthy",
    "formbuilder": "healthy",
    "organization": "healthy",
    "database": "healthy",
    "cache": "healthy"
  },
  "version": "2.0.0"
}
```

### Readiness Probe

```
GET /ready
```

Returns 200 when all modules are initialized and ready.

### Liveness Probe

```
GET /live
```

Returns 200 when application is running.

## 10. Configuration Management

### Configuration Loading Priority

1. Default values
2. Configuration file (config.yaml)
3. Environment variables (.env)
4. Command-line flags

### Config Structure

```go
type ExtractedConfigs struct {
    Server       *ServerConfig
    Database     *DatabaseConfig
    Identity     *IdentityConfig
    FormBuilder  *FormBuilderConfig
    Organization *OrganizationConfig
    DMS          *DMSConfig
    Vendors      *VendorsConfig
    Projects     *ProjectsConfig
    Logging      *LoggingConfig
    Monitoring   *MonitoringConfig
}
```

### Load Configuration

```go
func NewConfigLoader() (*pkgconfig.Config, error) {
    // Load from file
    cfg, err := file.LoadConfigFile("config.yaml")
    if err != nil {
        return nil, err
    }

    // Override with environment variables
    cfg.OverrideWithEnv()

    // Validate
    if err := cfg.Validate(); err != nil {
        return nil, err
    }

    return cfg, nil
}
```

## 11. Development

### Running in Development

```bash
# Run with live reload (using air)
air

# Run with debug logging
LOG_LEVEL=debug go run cmd/main.go

# Run specific modules only
ENABLED_MODULES=identity,formbuilder go run cmd/main.go
```

### Adding New Middleware

```go
// In server_factory.go
e.Use(middleware.Custom(middleware.CustomConfig{
    // Configuration
}))
```

### Debugging

```bash
# Enable debug logs
LOG_LEVEL=debug

# Enable FX debug
FX_DEBUG=true

# Enable SQL query logging
DB_LOG_QUERIES=true

# Profile CPU
go run cmd/main.go -cpuprofile=cpu.prof

# Profile memory
go run cmd/main.go -memprofile=mem.prof
```

## 12. Deployment

### Production Build

```bash
# Build optimized binary
CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o ugcl cmd/main.go

# Build with version
VERSION=2.0.0 go build -ldflags="-X main.Version=$VERSION" -o ugcl cmd/main.go
```

### Docker

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o ugcl cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/ugcl .
COPY --from=builder /app/config.yaml .
EXPOSE 8080
CMD ["./ugcl"]
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ugcl
spec:
  replicas: 3
  selector:
    matchLabels:
      app: ugcl
  template:
    metadata:
      labels:
        app: ugcl
    spec:
      containers:
      - name: ugcl
        image: ugcl:2.0.0
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: host
        livenessProbe:
          httpGet:
            path: /live
            port: 8080
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
```

## 13. Troubleshooting

### Common Issues

**Issue: Application won't start**
- **Cause:** Database connection failed
- **Solution:** Check DB_* environment variables

**Issue: Module not loading**
- **Cause:** Module not in ENABLED_MODULES
- **Solution:** Add module name to ENABLED_MODULES env var

**Issue: Port already in use**
- **Cause:** Another process using port 8080
- **Solution:** Change SERVER_PORT or kill existing process

**Issue: FX dependency error**
- **Cause:** Missing provider or circular dependency
- **Solution:** Check fx.Provide() calls and dependency graph

### Logs

```bash
# View startup logs
tail -f logs/app.log

# Filter by module
grep "module=identity" logs/app.log

# View errors only
grep "level=error" logs/app.log
```

## Additional Resources

- [Uber FX Documentation](https://uber-go.github.io/fx/)
- [Echo Framework](https://echo.labstack.com/)
- [Connect-RPC](https://connectrpc.com/)
- [Go Modules](https://go.dev/blog/using-go-modules)
