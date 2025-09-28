package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"p9e.in/ugcl/insightviewer/models"
)

// ISchedulingRepository defines the interface for report scheduling data access operations
type ISchedulingRepository interface {
	// Report Schedule operations
	CreateReportSchedule(ctx context.Context, schedule *models.ReportSchedule) (*models.ReportSchedule, error)
	GetReportScheduleByID(ctx context.Context, id uuid.UUID) (*models.ReportSchedule, error)
	UpdateReportSchedule(ctx context.Context, schedule *models.ReportSchedule) (*models.ReportSchedule, error)
	DeleteReportSchedule(ctx context.Context, id uuid.UUID) error
	GetReportSchedules(ctx context.Context, reportID uuid.UUID) ([]*models.ReportSchedule, error)
	GetActiveSchedules(ctx context.Context) ([]*models.ReportSchedule, error)
	UpdateNextRunTime(ctx context.Context, scheduleID uuid.UUID, nextRunAt *time.Time) error

	// Schedule Run operations
	CreateScheduleRun(ctx context.Context, run *models.ScheduleRun) (*models.ScheduleRun, error)
	GetScheduleRunByID(ctx context.Context, id uuid.UUID) (*models.ScheduleRun, error)
	UpdateScheduleRun(ctx context.Context, run *models.ScheduleRun) (*models.ScheduleRun, error)
	GetScheduleRuns(ctx context.Context, scheduleID uuid.UUID, limit, offset int32) ([]*models.ScheduleRun, int32, error)

	// Report Alert operations
	CreateReportAlert(ctx context.Context, alert *models.ReportAlert) (*models.ReportAlert, error)
	GetReportAlertByID(ctx context.Context, id uuid.UUID) (*models.ReportAlert, error)
	UpdateReportAlert(ctx context.Context, alert *models.ReportAlert) (*models.ReportAlert, error)
	DeleteReportAlert(ctx context.Context, id uuid.UUID) error
	GetReportAlerts(ctx context.Context, reportID uuid.UUID) ([]*models.ReportAlert, error)
	GetActiveAlerts(ctx context.Context) ([]*models.ReportAlert, error)
	UpdateAlertTrigger(ctx context.Context, alertID uuid.UUID, lastTriggeredAt time.Time) error

	// Alert Instance operations
	CreateAlertInstance(ctx context.Context, instance *models.AlertInstance) (*models.AlertInstance, error)
	GetAlertInstanceByID(ctx context.Context, id uuid.UUID) (*models.AlertInstance, error)
	UpdateAlertInstance(ctx context.Context, instance *models.AlertInstance) (*models.AlertInstance, error)
	GetAlertInstances(ctx context.Context, alertID uuid.UUID, limit, offset int32) ([]*models.AlertInstance, int32, error)
	ResolveAlertInstance(ctx context.Context, instanceID uuid.UUID, resolvedBy string) error
}