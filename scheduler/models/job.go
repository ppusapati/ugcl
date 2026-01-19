package models

import (
	"time"

	"github.com/google/uuid"
)

// Job represents a scheduled job in the system
type Job struct {
	ID             uuid.UUID
	Name           string
	Description    string
	CronExpression string
	IsActive       bool

	// Target service configuration
	TargetService  string // "insightviewer", "masters", etc.
	TargetMethod   string // "ExecuteReport", "RunDataSync", etc.
	TargetPayload  map[string]interface{} // Service-specific parameters

	// Execution tracking
	NextRunAt      *time.Time
	LastRunAt      *time.Time
	LastRunStatus  *string // "success", "failed", "timeout"
	RunCount       int32
	FailureCount   int32

	// Configuration
	TimeoutSeconds  int32
	MaxRetries     int32
	RetryInterval  int32 // seconds

	// Metadata
	CreatedBy      string
	CreatedAt      time.Time
	UpdatedBy      string
	UpdatedAt      time.Time

	// Notification settings
	NotifyOnSuccess   bool
	NotifyOnFailure   bool
	NotificationChannels []string // email, slack, webhook
	NotificationRecipients []string
}

// JobExecution represents a single execution instance of a job
type JobExecution struct {
	ID        uuid.UUID
	JobID     uuid.UUID
	Status    string // "running", "success", "failed", "timeout", "cancelled"
	StartedAt time.Time
	CompletedAt *time.Time
	DurationMs  *int32

	// Request/Response tracking
	RequestPayload  map[string]interface{}
	ResponseData    map[string]interface{}
	ErrorMessage    *string
	RetryAttempt    int32

	// Execution context
	TriggeredBy string // "scheduler", "manual", "webhook"
	HostName    string // Which instance executed this job

	CreatedAt time.Time
}

// JobExecutionLog represents detailed execution logs
type JobExecutionLog struct {
	ID          uuid.UUID
	ExecutionID uuid.UUID
	Level       string // "info", "warn", "error", "debug"
	Message     string
	Details     map[string]interface{}
	Timestamp   time.Time
}