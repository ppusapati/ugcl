package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"p9e.in/ugcl/scheduler/models"
)

// IJobRepository defines the interface for job data access operations
type IJobRepository interface {
	// Job CRUD operations
	CreateJob(ctx context.Context, job *models.Job) (*models.Job, error)
	GetJobByID(ctx context.Context, id uuid.UUID) (*models.Job, error)
	UpdateJob(ctx context.Context, job *models.Job) (*models.Job, error)
	DeleteJob(ctx context.Context, id uuid.UUID) error
	ListJobs(ctx context.Context, filterService string, activeOnly bool, limit, offset int32) ([]*models.Job, int32, error)
	GetActiveJobs(ctx context.Context) ([]*models.Job, error)

	// Job status operations
	UpdateJobStatus(ctx context.Context, jobID uuid.UUID, isActive bool, updatedBy string) error
	UpdateJobNextRun(ctx context.Context, jobID uuid.UUID, nextRunAt *time.Time) error
	UpdateJobStats(ctx context.Context, jobID uuid.UUID, status string) error

	// Job execution operations
	CreateJobExecution(ctx context.Context, execution *models.JobExecution) (*models.JobExecution, error)
	GetJobExecutionByID(ctx context.Context, id uuid.UUID) (*models.JobExecution, error)
	UpdateJobExecution(ctx context.Context, execution *models.JobExecution) (*models.JobExecution, error)
	GetJobExecutions(ctx context.Context, jobID uuid.UUID, limit, offset int32) ([]*models.JobExecution, int32, error)
	CancelJobExecution(ctx context.Context, executionID uuid.UUID) error

	// Statistics and monitoring
	GetJobStatistics(ctx context.Context, jobID *uuid.UUID, since *time.Time) (*JobStatistics, error)
}

// JobStatistics represents job execution statistics
type JobStatistics struct {
	TotalExecutions     int32
	SuccessfulExecutions int32
	FailedExecutions    int32
	SuccessRate         float64
	AverageDurationMs   float64
	LastExecution       *time.Time
}