package services

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// IMonitoringService defines the core monitoring service interface
type IMonitoringService interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	GetStatus() MonitoringStatus
}

// ISLAMonitorService defines the SLA monitoring service interface
type ISLAMonitorService interface {
	IMonitoringService
	StartSLAMonitoring(ctx context.Context) error
	CheckSLABreaches(ctx context.Context) error
	ProcessEscalations(ctx context.Context) error
	StopSLAMonitoring()
	SetMonitoringInterval(interval time.Duration)
}

// IMetricsMonitorService defines metrics monitoring (future implementation)
type IMetricsMonitorService interface {
	IMonitoringService
	CollectMetrics(ctx context.Context) error
	GetMetrics(ctx context.Context, timeRange TimeRange) ([]Metric, error)
}

// IHealthMonitorService defines health monitoring (future implementation)
type IHealthMonitorService interface {
	IMonitoringService
	CheckHealth(ctx context.Context) (HealthStatus, error)
	RegisterHealthCheck(name string, check HealthCheckFunc) error
}

// IAlertManagerService defines alert management (future implementation)
type IAlertManagerService interface {
	IMonitoringService
	SendAlert(ctx context.Context, alert Alert) error
	GetActiveAlerts(ctx context.Context) ([]Alert, error)
}

// Common types
type MonitoringStatus struct {
	Service   string    `json:"service"`
	Status    string    `json:"status"` // running, stopped, error
	StartTime time.Time `json:"start_time"`
	LastCheck time.Time `json:"last_check"`
	Error     string    `json:"error,omitempty"`
}

type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type Metric struct {
	Name      string                 `json:"name"`
	Value     float64                `json:"value"`
	Tags      map[string]string      `json:"tags"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type HealthStatus struct {
	Overall string                    `json:"overall"` // healthy, degraded, unhealthy
	Checks  map[string]HealthCheck    `json:"checks"`
}

type HealthCheck struct {
	Status  string    `json:"status"`
	Message string    `json:"message"`
	LastRun time.Time `json:"last_run"`
}

type HealthCheckFunc func(ctx context.Context) HealthCheck

type Alert struct {
	ID          uuid.UUID              `json:"id"`
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Message     string                 `json:"message"`
	Source      string                 `json:"source"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Resolved    bool                   `json:"resolved"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
}

// Event types for inter-service communication
type MonitoringEvent struct {
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// Configuration interfaces
type MonitoringConfig interface {
	GetSLAConfig() SLAConfig
	GetMetricsConfig() MetricsConfig
	GetHealthConfig() HealthConfig
	GetAlertConfig() AlertConfig
}

type SLAConfig struct {
	Enabled           bool          `json:"enabled"`
	CheckInterval     time.Duration `json:"check_interval"`
	BatchSize         int           `json:"batch_size"`
	MaxRetries        int           `json:"max_retries"`
	NotificationDelay time.Duration `json:"notification_delay"`
}

type MetricsConfig struct {
	Enabled         bool          `json:"enabled"`
	CollectInterval time.Duration `json:"collect_interval"`
	RetentionPeriod time.Duration `json:"retention_period"`
}

type HealthConfig struct {
	Enabled       bool          `json:"enabled"`
	CheckInterval time.Duration `json:"check_interval"`
	Timeout       time.Duration `json:"timeout"`
}

type AlertConfig struct {
	Enabled      bool     `json:"enabled"`
	Channels     []string `json:"channels"`
	RateLimiting bool     `json:"rate_limiting"`
}
