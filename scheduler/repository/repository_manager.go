package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"p9e.in/ugcl/scheduler/models"
)

// RepositoryManager combines all repository interfaces for scheduler module
type RepositoryManager struct {
	Job IJobRepository
}

// NewRepositoryManager creates a new repository manager instance
func NewRepositoryManager(jobRepo IJobRepository) *RepositoryManager {
	return &RepositoryManager{
		Job: jobRepo,
	}
}

// Interface compliance methods for engine.RepositoryManager interface
func (rm *RepositoryManager) CreateJobExecution(ctx context.Context, execution *models.JobExecution) (*models.JobExecution, error) {
	return rm.Job.CreateJobExecution(ctx, execution)
}

func (rm *RepositoryManager) UpdateJobExecution(ctx context.Context, execution *models.JobExecution) (*models.JobExecution, error) {
	return rm.Job.UpdateJobExecution(ctx, execution)
}

func (rm *RepositoryManager) UpdateJobStats(ctx context.Context, jobID uuid.UUID, status string) error {
	return rm.Job.UpdateJobStats(ctx, jobID, status)
}