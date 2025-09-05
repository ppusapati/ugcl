# Monitoring Module

## Overview
The monitoring module provides centralized monitoring services for the application, designed for easy microservices migration and extensibility.

## Architecture

```
monitoring/
├── services/           # Core monitoring services
│   ├── interfaces.go   # Service interfaces
│   ├── sla_monitor.go  # SLA monitoring implementation
│   ├── metrics.go      # Metrics monitoring (future)
│   └── alerts.go       # Alert management (future)
├── repository/         # Monitoring data access
│   ├── interfaces.go   # Repository interfaces
│   └── monitoring.go   # Monitoring repository
├── handlers/           # Monitoring API handlers
│   └── monitoring.go   # HTTP/gRPC handlers
├── proto/             # Protocol buffer definitions
│   └── monitoring.proto
├── config/            # Monitoring configuration
│   └── config.go
└── module.go          # FX module definition
```

## Services

### SLA Monitor Service
- Real-time SLA breach detection
- Multi-level escalation management
- Configurable monitoring intervals
- Extensible notification system

### Future Services
- **Metrics Monitor**: Application performance metrics
- **Health Monitor**: Service health checks
- **Alert Manager**: Centralized alerting
- **Audit Monitor**: Compliance and audit trails

## Microservices Ready
This module is designed to be easily extracted into a separate microservice:
- Clean interfaces and dependency injection
- Minimal external dependencies
- Event-driven architecture support
- Independent configuration management

## Usage

```go
// In your main application
import "p9e.in/ugcl/monitoring"

// Add to FX modules
fx.New(
    monitoring.Module,
    // ... other modules
)
```
