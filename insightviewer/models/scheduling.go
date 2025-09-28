package models

import (
	"time"

	"github.com/google/uuid"
)

// ReportSchedule represents a scheduled execution of a report
type ReportSchedule struct {
	ID             uuid.UUID
	ReportID       uuid.UUID
	Name           string
	Description    string
	CronExpression string
	IsActive       bool
	Parameters     map[string]interface{}
	CreatedBy      string
	CreatedAt      time.Time
	UpdatedBy      string
	UpdatedAt      time.Time
	NextRunAt      *time.Time
	LastRunAt      *time.Time
	LastRunStatus  *string
}

// ScheduleRun represents a single execution of a scheduled report
type ScheduleRun struct {
	ID         uuid.UUID
	ScheduleID uuid.UUID
	RunID      uuid.UUID
	Status     string // scheduled, running, completed, failed, skipped
	StartedAt  time.Time
	CompletedAt *time.Time
	ErrorMsg   *string
	CreatedAt  time.Time
}

// ReportAlert represents alert conditions for a report
type ReportAlert struct {
	ID                   uuid.UUID
	ReportID             uuid.UUID
	Name                 string
	Description          string
	AlertType            string // threshold, anomaly, data_quality, execution_failure
	AlertRules           map[string]interface{} // JSON rules for triggering alerts
	IsActive             bool
	NotificationChannels []string // email, slack, webhook, etc.
	CreatedBy            string
	CreatedAt            time.Time
	UpdatedBy            string
	UpdatedAt            time.Time
	LastTriggeredAt      *time.Time
	TriggerCount         int32
}

// AlertInstance represents a triggered alert
type AlertInstance struct {
	ID        uuid.UUID
	AlertID   uuid.UUID
	RunID     *uuid.UUID // The run that triggered this alert
	Severity  string     // low, medium, high, critical
	Message   string
	Details   map[string]interface{}
	Status    string // active, acknowledged, resolved
	CreatedAt time.Time
	ResolvedAt *time.Time
	ResolvedBy *string
}