package config

import (
	"time"

	"p9e.in/ugcl/monitoring/services"
)

// MonitoringConfiguration implements services.MonitoringConfig
type MonitoringConfiguration struct {
	SLA     SLAConfiguration     `json:"sla" yaml:"sla"`
	Metrics MetricsConfiguration `json:"metrics" yaml:"metrics"`
	Health  HealthConfiguration  `json:"health" yaml:"health"`
	Alert   AlertConfiguration   `json:"alert" yaml:"alert"`
}

type SLAConfiguration struct {
	Enabled           bool          `json:"enabled" yaml:"enabled"`
	CheckInterval     time.Duration `json:"check_interval" yaml:"check_interval"`
	BatchSize         int           `json:"batch_size" yaml:"batch_size"`
	MaxRetries        int           `json:"max_retries" yaml:"max_retries"`
	NotificationDelay time.Duration `json:"notification_delay" yaml:"notification_delay"`
}

type MetricsConfiguration struct {
	Enabled         bool          `json:"enabled" yaml:"enabled"`
	CollectInterval time.Duration `json:"collect_interval" yaml:"collect_interval"`
	RetentionPeriod time.Duration `json:"retention_period" yaml:"retention_period"`
}

type HealthConfiguration struct {
	Enabled       bool          `json:"enabled" yaml:"enabled"`
	CheckInterval time.Duration `json:"check_interval" yaml:"check_interval"`
	Timeout       time.Duration `json:"timeout" yaml:"timeout"`
}

type AlertConfiguration struct {
	Enabled      bool     `json:"enabled" yaml:"enabled"`
	Channels     []string `json:"channels" yaml:"channels"`
	RateLimiting bool     `json:"rate_limiting" yaml:"rate_limiting"`
}

// GetSLAConfig implements services.MonitoringConfig
func (c *MonitoringConfiguration) GetSLAConfig() services.SLAConfig {
	return services.SLAConfig{
		Enabled:           c.SLA.Enabled,
		CheckInterval:     c.SLA.CheckInterval,
		BatchSize:         c.SLA.BatchSize,
		MaxRetries:        c.SLA.MaxRetries,
		NotificationDelay: c.SLA.NotificationDelay,
	}
}

// GetMetricsConfig implements services.MonitoringConfig
func (c *MonitoringConfiguration) GetMetricsConfig() services.MetricsConfig {
	return services.MetricsConfig{
		Enabled:         c.Metrics.Enabled,
		CollectInterval: c.Metrics.CollectInterval,
		RetentionPeriod: c.Metrics.RetentionPeriod,
	}
}

// GetHealthConfig implements services.MonitoringConfig
func (c *MonitoringConfiguration) GetHealthConfig() services.HealthConfig {
	return services.HealthConfig{
		Enabled:       c.Health.Enabled,
		CheckInterval: c.Health.CheckInterval,
		Timeout:       c.Health.Timeout,
	}
}

// GetAlertConfig implements services.MonitoringConfig
func (c *MonitoringConfiguration) GetAlertConfig() services.AlertConfig {
	return services.AlertConfig{
		Enabled:      c.Alert.Enabled,
		Channels:     c.Alert.Channels,
		RateLimiting: c.Alert.RateLimiting,
	}
}

// DefaultMonitoringConfig returns a default monitoring configuration
func DefaultMonitoringConfig() *MonitoringConfiguration {
	return &MonitoringConfiguration{
		SLA: SLAConfiguration{
			Enabled:           true,
			CheckInterval:     30 * time.Second,
			BatchSize:         100,
			MaxRetries:        3,
			NotificationDelay: 5 * time.Second,
		},
		Metrics: MetricsConfiguration{
			Enabled:         false, // Disabled by default
			CollectInterval: 1 * time.Minute,
			RetentionPeriod: 24 * time.Hour,
		},
		Health: HealthConfiguration{
			Enabled:       false, // Disabled by default
			CheckInterval: 30 * time.Second,
			Timeout:       10 * time.Second,
		},
		Alert: AlertConfiguration{
			Enabled:      true,
			Channels:     []string{"email", "sms"},
			RateLimiting: true,
		},
	}
}
