package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"p9e.in/ugcl/scheduler/internal/engine"
	"p9e.in/ugcl/scheduler/models"
	"p9e.in/ugcl/scheduler/repository"
)

// SchedulerService is the main service for managing scheduled jobs
type SchedulerService struct {
	repoManager *repository.RepositoryManager
	cronEngine  *engine.CronEngine
	executor    *engine.ServiceExecutor
	running     bool
}

// NewSchedulerService creates a new scheduler service
func NewSchedulerService(repoManager *repository.RepositoryManager) *SchedulerService {
	executor := engine.NewServiceExecutor(repoManager)
	cronEngine := engine.NewCronEngine(executor)

	return &SchedulerService{
		repoManager: repoManager,
		cronEngine:  cronEngine,
		executor:    executor,
	}
}

// Start starts the scheduler service
func (s *SchedulerService) Start(ctx context.Context) error {
	if s.running {
		return nil
	}

	log.Println("Starting scheduler service...")

	// Load and schedule all active jobs
	if err := s.loadActiveJobs(ctx); err != nil {
		return fmt.Errorf("failed to load active jobs: %w", err)
	}

	// Start the cron engine
	if err := s.cronEngine.Start(ctx); err != nil {
		return fmt.Errorf("failed to start cron engine: %w", err)
	}

	s.running = true
	log.Println("Scheduler service started successfully")

	return nil
}

// Stop stops the scheduler service
func (s *SchedulerService) Stop(ctx context.Context) error {
	if !s.running {
		return nil
	}

	log.Println("Stopping scheduler service...")

	// Stop the cron engine
	if err := s.cronEngine.Stop(); err != nil {
		log.Printf("Error stopping cron engine: %v", err)
	}

	// Close executor resources
	if err := s.executor.Close(); err != nil {
		log.Printf("Error closing executor: %v", err)
	}

	s.running = false
	log.Println("Scheduler service stopped")

	return nil
}

// CreateJob creates a new scheduled job
func (s *SchedulerService) CreateJob(ctx context.Context, job *models.Job) (*models.Job, error) {
	// Validate cron expression
	if err := s.cronEngine.ValidateCronExpression(job.CronExpression); err != nil {
		return nil, fmt.Errorf("invalid cron expression: %w", err)
	}

	// Calculate next run time
	if job.IsActive {
		nextRun, err := s.cronEngine.GetNextRunTime(job.CronExpression)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate next run time: %w", err)
		}
		job.NextRunAt = nextRun
	}

	// Set timestamps
	now := time.Now()
	job.ID = uuid.New()
	job.CreatedAt = now
	job.UpdatedAt = now
	job.RunCount = 0
	job.FailureCount = 0

	// Save to database
	createdJob, err := s.repoManager.Job.CreateJob(ctx, job)
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	// Add to scheduler if active
	if createdJob.IsActive {
		if err := s.cronEngine.AddJob(createdJob); err != nil {
			log.Printf("Warning: Failed to add job to scheduler: %v", err)
		}
	}

	log.Printf("Job created: %s (%s)", createdJob.Name, createdJob.ID)
	return createdJob, nil
}

// GetJob retrieves a job by ID
func (s *SchedulerService) GetJob(ctx context.Context, jobID uuid.UUID) (*models.Job, error) {
	return s.repoManager.Job.GetJobByID(ctx, jobID)
}

// UpdateJob updates an existing job
func (s *SchedulerService) UpdateJob(ctx context.Context, job *models.Job) (*models.Job, error) {
	// Validate cron expression
	if err := s.cronEngine.ValidateCronExpression(job.CronExpression); err != nil {
		return nil, fmt.Errorf("invalid cron expression: %w", err)
	}

	// Calculate next run time if active
	if job.IsActive {
		nextRun, err := s.cronEngine.GetNextRunTime(job.CronExpression)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate next run time: %w", err)
		}
		job.NextRunAt = nextRun
	} else {
		job.NextRunAt = nil
	}

	// Update timestamp
	job.UpdatedAt = time.Now()

	// Save to database
	updatedJob, err := s.repoManager.Job.UpdateJob(ctx, job)
	if err != nil {
		return nil, fmt.Errorf("failed to update job: %w", err)
	}

	// Update in scheduler
	if err := s.cronEngine.UpdateJob(updatedJob); err != nil {
		log.Printf("Warning: Failed to update job in scheduler: %v", err)
	}

	log.Printf("Job updated: %s (%s)", updatedJob.Name, updatedJob.ID)
	return updatedJob, nil
}

// DeleteJob deletes a job
func (s *SchedulerService) DeleteJob(ctx context.Context, jobID uuid.UUID) error {
	// Remove from scheduler first
	if err := s.cronEngine.RemoveJob(jobID); err != nil {
		log.Printf("Warning: Failed to remove job from scheduler: %v", err)
	}

	// Delete from database
	if err := s.repoManager.Job.DeleteJob(ctx, jobID); err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}

	log.Printf("Job deleted: %s", jobID)
	return nil
}

// ListJobs lists jobs with optional filtering
func (s *SchedulerService) ListJobs(ctx context.Context, filterService string, activeOnly bool, limit, offset int32) ([]*models.Job, int32, error) {
	return s.repoManager.Job.ListJobs(ctx, filterService, activeOnly, limit, offset)
}

// EnableJob enables a job
func (s *SchedulerService) EnableJob(ctx context.Context, jobID uuid.UUID, updatedBy string) (*time.Time, error) {
	// Get job
	job, err := s.repoManager.Job.GetJobByID(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	if job.IsActive {
		return job.NextRunAt, nil // Already active
	}

	// Calculate next run time
	nextRun, err := s.cronEngine.GetNextRunTime(job.CronExpression)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate next run time: %w", err)
	}

	// Update status in database
	if err := s.repoManager.Job.UpdateJobStatus(ctx, jobID, true, updatedBy); err != nil {
		return nil, fmt.Errorf("failed to enable job: %w", err)
	}

	// Update next run time
	if err := s.repoManager.Job.UpdateJobNextRun(ctx, jobID, nextRun); err != nil {
		return nil, fmt.Errorf("failed to update next run time: %w", err)
	}

	// Add to scheduler
	job.IsActive = true
	job.NextRunAt = nextRun
	if err := s.cronEngine.AddJob(job); err != nil {
		log.Printf("Warning: Failed to add job to scheduler: %v", err)
	}

	log.Printf("Job enabled: %s (%s)", job.Name, jobID)
	return nextRun, nil
}

// DisableJob disables a job
func (s *SchedulerService) DisableJob(ctx context.Context, jobID uuid.UUID, updatedBy string) error {
	// Remove from scheduler
	if err := s.cronEngine.RemoveJob(jobID); err != nil {
		log.Printf("Warning: Failed to remove job from scheduler: %v", err)
	}

	// Update status in database
	if err := s.repoManager.Job.UpdateJobStatus(ctx, jobID, false, updatedBy); err != nil {
		return fmt.Errorf("failed to disable job: %w", err)
	}

	log.Printf("Job disabled: %s", jobID)
	return nil
}

// TriggerJob manually triggers a job execution
func (s *SchedulerService) TriggerJob(ctx context.Context, jobID uuid.UUID, triggeredBy string, overridePayload map[string]interface{}) (*models.JobExecution, error) {
	// Get job
	job, err := s.repoManager.Job.GetJobByID(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	// Override payload if provided
	if overridePayload != nil {
		job.TargetPayload = overridePayload
	}

	// Create execution record
	execution := &models.JobExecution{
		ID:             uuid.New(),
		JobID:          job.ID,
		Status:         "running",
		StartedAt:      time.Now(),
		RequestPayload: job.TargetPayload,
		TriggeredBy:    triggeredBy,
		RetryAttempt:   0,
		CreatedAt:      time.Now(),
	}

	// Save execution record
	execution, err = s.repoManager.Job.CreateJobExecution(ctx, execution)
	if err != nil {
		return nil, fmt.Errorf("failed to create execution record: %w", err)
	}

	// Execute job asynchronously
	go func() {
		if err := s.executor.ExecuteJob(context.Background(), job); err != nil {
			log.Printf("Manual job execution failed for %s: %v", job.Name, err)
		}
	}()

	log.Printf("Job manually triggered: %s (%s) by %s", job.Name, jobID, triggeredBy)
	return execution, nil
}

// GetJobExecutions gets execution history for a job
func (s *SchedulerService) GetJobExecutions(ctx context.Context, jobID uuid.UUID, limit, offset int32) ([]*models.JobExecution, int32, error) {
	return s.repoManager.Job.GetJobExecutions(ctx, jobID, limit, offset)
}

// GetJobExecution gets a specific job execution
func (s *SchedulerService) GetJobExecution(ctx context.Context, executionID uuid.UUID) (*models.JobExecution, error) {
	return s.repoManager.Job.GetJobExecutionByID(ctx, executionID)
}

// GetHealth returns scheduler health status
func (s *SchedulerService) GetHealth() map[string]interface{} {
	return map[string]interface{}{
		"status":             s.getHealthStatus(),
		"running":            s.running,
		"active_jobs":        s.cronEngine.GetActiveJobs(),
		"cron_engine_running": s.cronEngine.IsRunning(),
		"last_heartbeat":     time.Now(),
		"version":            "1.0.0",
	}
}

// GetJobStatistics gets statistics for jobs
func (s *SchedulerService) GetJobStatistics(ctx context.Context, jobID *uuid.UUID, since *time.Time) (*repository.JobStatistics, error) {
	return s.repoManager.Job.GetJobStatistics(ctx, jobID, since)
}

// loadActiveJobs loads all active jobs and adds them to the scheduler
func (s *SchedulerService) loadActiveJobs(ctx context.Context) error {
	jobs, err := s.repoManager.Job.GetActiveJobs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active jobs: %w", err)
	}

	for _, job := range jobs {
		if err := s.cronEngine.AddJob(job); err != nil {
			log.Printf("Warning: Failed to add job %s to scheduler: %v", job.Name, err)
			continue
		}
	}

	log.Printf("Loaded %d active jobs into scheduler", len(jobs))
	return nil
}

// getHealthStatus determines the overall health status
func (s *SchedulerService) getHealthStatus() string {
	if !s.running {
		return "unhealthy"
	}

	if !s.cronEngine.IsRunning() {
		return "degraded"
	}

	return "healthy"
}