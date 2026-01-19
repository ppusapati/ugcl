package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"p9e.in/ugcl/backupdr/models"
	"p9e.in/ugcl/backupdr/repository"
)

// BackupOrchestrator manages backup execution and scheduling
type BackupOrchestrator struct {
	repo            repository.IBackupDRRepository
	storageManager  StorageManager
	notificationSvc NotificationService
	logger          *zap.Logger

	// Runtime state
	mu          sync.RWMutex
	running     bool
	activeJobs  map[uuid.UUID]*JobExecution
	shutdownCh  chan struct{}
}

// JobExecution tracks the execution state of a backup job
type JobExecution struct {
	JobID       uuid.UUID
	StartTime   time.Time
	CancelFunc  context.CancelFunc
	Status      models.JobStatus
}

// StorageManager interface for different storage backends
type StorageManager interface {
	Upload(ctx context.Context, config models.StorageConfig, localPath, remotePath string) error
	Download(ctx context.Context, config models.StorageConfig, remotePath, localPath string) error
	List(ctx context.Context, config models.StorageConfig, path string) ([]string, error)
	Delete(ctx context.Context, config models.StorageConfig, path string) error
	GetSize(ctx context.Context, config models.StorageConfig, path string) (int64, error)
	Exists(ctx context.Context, config models.StorageConfig, path string) (bool, error)
}

// NotificationService interface for sending notifications
type NotificationService interface {
	SendNotification(ctx context.Context, notification *BackupNotification) error
}

// BackupNotification represents a backup-related notification
type BackupNotification struct {
	Type      string    `json:"type"`
	Subject   string    `json:"subject"`
	Message   string    `json:"message"`
	Recipients []string `json:"recipients"`
	JobID     uuid.UUID `json:"job_id"`
	PolicyID  uuid.UUID `json:"policy_id"`
}

func NewBackupOrchestrator(
	repo repository.IBackupDRRepository,
	storageManager StorageManager,
	notificationSvc NotificationService,
	logger *zap.Logger,
) *BackupOrchestrator {
	return &BackupOrchestrator{
		repo:            repo,
		storageManager:  storageManager,
		notificationSvc: notificationSvc,
		logger:          logger,
		activeJobs:      make(map[uuid.UUID]*JobExecution),
		shutdownCh:      make(chan struct{}),
	}
}

// Start initializes the backup orchestrator
func (bo *BackupOrchestrator) Start(ctx context.Context) error {
	bo.mu.Lock()
	defer bo.mu.Unlock()

	if bo.running {
		return fmt.Errorf("backup orchestrator is already running")
	}

	bo.logger.Info("Starting backup orchestrator")
	bo.running = true

	// Start background job processing
	go bo.processScheduledJobs(ctx)
	go bo.monitorActiveJobs(ctx)

	bo.logger.Info("Backup orchestrator started successfully")
	return nil
}

// Stop gracefully shuts down the backup orchestrator
func (bo *BackupOrchestrator) Stop() error {
	bo.mu.Lock()
	defer bo.mu.Unlock()

	if !bo.running {
		return fmt.Errorf("backup orchestrator is not running")
	}

	bo.logger.Info("Stopping backup orchestrator")

	// Cancel all active jobs
	for jobID, execution := range bo.activeJobs {
		bo.logger.Info("Cancelling active job", zap.String("job_id", jobID.String()))
		execution.CancelFunc()
	}

	// Signal shutdown
	close(bo.shutdownCh)
	bo.running = false

	// Wait for active jobs to finish (with timeout)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-shutdownCtx.Done():
			bo.logger.Warn("Shutdown timeout reached, forcing shutdown")
			return nil
		case <-ticker.C:
			if len(bo.activeJobs) == 0 {
				bo.logger.Info("All active jobs completed, shutdown complete")
				return nil
			}
			bo.logger.Info("Waiting for active jobs to complete", zap.Int("active_jobs", len(bo.activeJobs)))
		}
	}
}

// ScheduleBackup schedules a backup job
func (bo *BackupOrchestrator) ScheduleBackup(ctx context.Context, policyID uuid.UUID) (*models.BackupJob, error) {
	bo.logger.Info("Scheduling backup", zap.String("policy_id", policyID.String()))

	// Get the backup policy
	policy, err := bo.repo.GetBackupPolicyByID(ctx, policyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup policy: %w", err)
	}

	if policy == nil {
		return nil, fmt.Errorf("backup policy not found")
	}

	if !policy.IsActive {
		return nil, fmt.Errorf("backup policy is not active")
	}

	// Create backup job
	job := &models.BackupJob{
		ID:            uuid.New(),
		PolicyID:      policyID,
		JobType:       policy.BackupType,
		Status:        models.JobStatusScheduled,
		Priority:      models.PriorityNormal,
		ScheduledAt:   time.Now(),
		MaxRetries:    3,
		RetryCount:    0,
		ExecutedBy:    uuid.New(), // TODO: Get actual user ID
		JobMetadata:   []byte("{}"),
	}

	// Save the job
	createdJob, err := bo.repo.CreateBackupJob(ctx, job)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup job: %w", err)
	}

	bo.logger.Info("Backup job scheduled successfully",
		zap.String("job_id", createdJob.ID.String()),
		zap.String("policy_id", policyID.String()))

	return createdJob, nil
}

// ExecuteBackup executes a backup job
func (bo *BackupOrchestrator) ExecuteBackup(ctx context.Context, jobID uuid.UUID) error {
	bo.logger.Info("Executing backup", zap.String("job_id", jobID.String()))

	// Get the backup job
	job, err := bo.repo.GetBackupJobByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get backup job: %w", err)
	}

	if job == nil {
		return fmt.Errorf("backup job not found")
	}

	if job.Status != models.JobStatusScheduled && job.Status != models.JobStatusQueued {
		return fmt.Errorf("job is not in a state that can be executed: %s", job.Status)
	}

	// Create execution context with cancellation
	execCtx, cancel := context.WithCancel(ctx)
	execution := &JobExecution{
		JobID:      jobID,
		StartTime:  time.Now(),
		CancelFunc: cancel,
		Status:     models.JobStatusRunning,
	}

	// Track active job
	bo.mu.Lock()
	bo.activeJobs[jobID] = execution
	bo.mu.Unlock()

	// Update job status to running
	job.Status = models.JobStatusRunning
	startTime := time.Now()
	job.StartedAt = &startTime
	err = bo.repo.UpdateBackupJob(ctx, job)
	if err != nil {
		bo.logger.Error("Failed to update job status to running", zap.Error(err))
	}

	// Execute backup in goroutine
	go bo.executeBackupJob(execCtx, job, execution)

	return nil
}

// CancelBackup cancels a running backup job
func (bo *BackupOrchestrator) CancelBackup(ctx context.Context, jobID uuid.UUID) error {
	bo.logger.Info("Cancelling backup", zap.String("job_id", jobID.String()))

	bo.mu.Lock()
	execution, exists := bo.activeJobs[jobID]
	bo.mu.Unlock()

	if !exists {
		return fmt.Errorf("job is not currently executing")
	}

	// Cancel the execution context
	execution.CancelFunc()
	execution.Status = models.JobStatusCancelled

	// Update job status in repository
	job, err := bo.repo.GetBackupJobByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get backup job: %w", err)
	}

	job.Status = models.JobStatusCancelled
	completedTime := time.Now()
	job.CompletedAt = &completedTime
	if job.StartedAt != nil {
		duration := completedTime.Sub(*job.StartedAt)
		job.Duration = &duration
	}

	err = bo.repo.UpdateBackupJob(ctx, job)
	if err != nil {
		bo.logger.Error("Failed to update job status to cancelled", zap.Error(err))
	}

	bo.logger.Info("Backup job cancelled", zap.String("job_id", jobID.String()))
	return nil
}

// GetBackupStatus gets the status of a backup job
func (bo *BackupOrchestrator) GetBackupStatus(ctx context.Context, jobID uuid.UUID) (*models.BackupJob, error) {
	bo.logger.Info("Getting backup status", zap.String("job_id", jobID.String()))

	return bo.repo.GetBackupJobByID(ctx, jobID)
}

// ListActiveBackups lists all active backup jobs
func (bo *BackupOrchestrator) ListActiveBackups(ctx context.Context) ([]*models.BackupJob, error) {
	bo.logger.Info("Listing active backups")

	return bo.repo.GetActiveJobs(ctx)
}

// Background processing methods

// processScheduledJobs continuously checks for scheduled jobs and executes them
func (bo *BackupOrchestrator) processScheduledJobs(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-bo.shutdownCh:
			return
		case <-ticker.C:
			bo.processScheduledJobsBatch(ctx)
		}
	}
}

// processScheduledJobsBatch processes a batch of scheduled jobs
func (bo *BackupOrchestrator) processScheduledJobsBatch(ctx context.Context) {
	// Get jobs that are scheduled to run
	nextRunBefore := time.Now().Add(1 * time.Minute) // Jobs scheduled to run in the next minute
	jobs, err := bo.repo.GetJobsForScheduling(ctx, nextRunBefore)
	if err != nil {
		bo.logger.Error("Failed to get jobs for scheduling", zap.Error(err))
		return
	}

	for _, job := range jobs {
		// Check if job is already being executed
		bo.mu.RLock()
		_, isActive := bo.activeJobs[job.ID]
		bo.mu.RUnlock()

		if !isActive {
			bo.logger.Info("Executing scheduled job", zap.String("job_id", job.ID.String()))
			err := bo.ExecuteBackup(ctx, job.ID)
			if err != nil {
				bo.logger.Error("Failed to execute scheduled job",
					zap.String("job_id", job.ID.String()),
					zap.Error(err))
			}
		}
	}
}

// monitorActiveJobs monitors active jobs and cleans up completed ones
func (bo *BackupOrchestrator) monitorActiveJobs(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second) // Check every 10 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-bo.shutdownCh:
			return
		case <-ticker.C:
			bo.cleanupCompletedJobs(ctx)
		}
	}
}

// cleanupCompletedJobs removes completed job executions from active jobs map
func (bo *BackupOrchestrator) cleanupCompletedJobs(ctx context.Context) {
	bo.mu.Lock()
	defer bo.mu.Unlock()

	for jobID, execution := range bo.activeJobs {
		if execution.Status == models.JobStatusCompleted ||
			execution.Status == models.JobStatusFailed ||
			execution.Status == models.JobStatusCancelled {
			delete(bo.activeJobs, jobID)
			bo.logger.Debug("Cleaned up completed job execution",
				zap.String("job_id", jobID.String()),
				zap.String("status", string(execution.Status)))
		}
	}
}

// executeBackupJob performs the actual backup execution
func (bo *BackupOrchestrator) executeBackupJob(ctx context.Context, job *models.BackupJob, execution *JobExecution) {
	defer func() {
		// Mark execution as completed
		bo.mu.Lock()
		execution.Status = job.Status
		bo.mu.Unlock()
	}()

	bo.logger.Info("Starting backup job execution",
		zap.String("job_id", job.ID.String()),
		zap.String("job_type", string(job.JobType)))

	// Get the policy for this job
	policy, err := bo.repo.GetBackupPolicyByID(ctx, job.PolicyID)
	if err != nil {
		bo.logger.Error("Failed to get backup policy", zap.Error(err))
		bo.updateJobStatus(ctx, job, models.JobStatusFailed, fmt.Sprintf("Failed to get backup policy: %v", err))
		return
	}

	// Simulate backup execution based on job type
	var backupSize int64
	var backupPath string

	switch job.JobType {
	case models.BackupTypeFull:
		backupSize, backupPath, err = bo.executeFullBackup(ctx, policy)
	case models.BackupTypeIncremental:
		backupSize, backupPath, err = bo.executeIncrementalBackup(ctx, policy)
	case models.BackupTypeDifferential:
		backupSize, backupPath, err = bo.executeDifferentialBackup(ctx, policy)
	case models.BackupTypeSnapshot:
		backupSize, backupPath, err = bo.executeSnapshotBackup(ctx, policy)
	default:
		err = fmt.Errorf("unsupported backup type: %s", job.JobType)
	}

	if err != nil {
		select {
		case <-ctx.Done():
			// Job was cancelled
			bo.updateJobStatus(ctx, job, models.JobStatusCancelled, "Job was cancelled")
			return
		default:
			// Job failed
			bo.logger.Error("Backup execution failed",
				zap.String("job_id", job.ID.String()),
				zap.Error(err))

			// Check if we should retry
			if job.RetryCount < job.MaxRetries {
				bo.scheduleRetry(ctx, job)
			} else {
				bo.updateJobStatus(ctx, job, models.JobStatusFailed, err.Error())
				bo.sendNotification(ctx, job, policy, "BACKUP_FAILED", fmt.Sprintf("Backup job failed: %v", err))
			}
			return
		}
	}

	// Backup completed successfully
	job.BackupSize = backupSize
	job.BackupPath = backupPath
	job.ProcessedSize = backupSize
	job.ProgressPercent = 100.0

	// Create backup instance record
	instance := &models.BackupInstance{
		JobID:         job.ID,
		BackupPath:    backupPath,
		BackupSize:    backupSize,
		ChecksumType:  "SHA256",
		ChecksumValue: "dummy-checksum", // TODO: Calculate actual checksum
		Status:        models.InstanceStatusValid,
		ExpiresAt:     calculateExpirationDate(policy),
		Metadata:      []byte("{}"),
	}

	_, err = bo.repo.CreateBackupInstance(ctx, instance)
	if err != nil {
		bo.logger.Error("Failed to create backup instance record", zap.Error(err))
	}

	bo.updateJobStatus(ctx, job, models.JobStatusCompleted, "Backup completed successfully")
	bo.sendNotification(ctx, job, policy, "BACKUP_SUCCESS", "Backup completed successfully")

	bo.logger.Info("Backup job completed successfully",
		zap.String("job_id", job.ID.String()),
		zap.Int64("backup_size", backupSize),
		zap.String("backup_path", backupPath))
}

// Helper methods

// updateJobStatus updates the job status and timing information
func (bo *BackupOrchestrator) updateJobStatus(ctx context.Context, job *models.BackupJob, status models.JobStatus, errorMessage string) {
	job.Status = status
	completedTime := time.Now()
	job.CompletedAt = &completedTime

	if job.StartedAt != nil {
		duration := completedTime.Sub(*job.StartedAt)
		job.Duration = &duration
	}

	if errorMessage != "" {
		job.ErrorMessage = errorMessage
	}

	err := bo.repo.UpdateBackupJob(ctx, job)
	if err != nil {
		bo.logger.Error("Failed to update job status",
			zap.String("job_id", job.ID.String()),
			zap.String("status", string(status)),
			zap.Error(err))
	}
}

// scheduleRetry schedules a job for retry
func (bo *BackupOrchestrator) scheduleRetry(ctx context.Context, job *models.BackupJob) {
	job.RetryCount++
	job.Status = models.JobStatusRetrying
	// Schedule retry in 5 minutes
	retryTime := time.Now().Add(5 * time.Minute)
	job.ScheduledAt = retryTime

	err := bo.repo.UpdateBackupJob(ctx, job)
	if err != nil {
		bo.logger.Error("Failed to schedule job retry",
			zap.String("job_id", job.ID.String()),
			zap.Error(err))
	}

	bo.logger.Info("Job scheduled for retry",
		zap.String("job_id", job.ID.String()),
		zap.Int("retry_count", job.RetryCount),
		zap.Time("retry_at", retryTime))
}

// sendNotification sends backup-related notifications
func (bo *BackupOrchestrator) sendNotification(ctx context.Context, job *models.BackupJob, policy *models.BackupPolicy, notificationType, message string) {
	if bo.notificationSvc == nil {
		return
	}

	notification := &BackupNotification{
		Type:       notificationType,
		Subject:    fmt.Sprintf("Backup %s - %s", notificationType, policy.Name),
		Message:    message,
		Recipients: []string{"admin@example.com"}, // TODO: Get from policy or configuration
		JobID:      job.ID,
		PolicyID:   job.PolicyID,
	}

	err := bo.notificationSvc.SendNotification(ctx, notification)
	if err != nil {
		bo.logger.Error("Failed to send notification",
			zap.String("job_id", job.ID.String()),
			zap.String("type", notificationType),
			zap.Error(err))
	}
}

// Backup execution methods for different types

// executeFullBackup performs a full backup
func (bo *BackupOrchestrator) executeFullBackup(ctx context.Context, policy *models.BackupPolicy) (int64, string, error) {
	bo.logger.Info("Executing full backup",
		zap.String("target_type", string(policy.TargetType)),
		zap.String("target_name", policy.TargetName))

	// Simulate backup execution
	select {
	case <-ctx.Done():
		return 0, "", ctx.Err()
	case <-time.After(2 * time.Second): // Simulate 2-second backup
		// Return simulated results
		backupSize := int64(1000000) // 1MB
		backupPath := fmt.Sprintf("/backups/%s/full_%d.backup", policy.TargetName, time.Now().Unix())
		return backupSize, backupPath, nil
	}
}

// executeIncrementalBackup performs an incremental backup
func (bo *BackupOrchestrator) executeIncrementalBackup(ctx context.Context, policy *models.BackupPolicy) (int64, string, error) {
	bo.logger.Info("Executing incremental backup",
		zap.String("target_type", string(policy.TargetType)),
		zap.String("target_name", policy.TargetName))

	// Simulate backup execution
	select {
	case <-ctx.Done():
		return 0, "", ctx.Err()
	case <-time.After(1 * time.Second): // Simulate 1-second backup
		// Return simulated results
		backupSize := int64(100000) // 100KB
		backupPath := fmt.Sprintf("/backups/%s/incremental_%d.backup", policy.TargetName, time.Now().Unix())
		return backupSize, backupPath, nil
	}
}

// executeDifferentialBackup performs a differential backup
func (bo *BackupOrchestrator) executeDifferentialBackup(ctx context.Context, policy *models.BackupPolicy) (int64, string, error) {
	bo.logger.Info("Executing differential backup",
		zap.String("target_type", string(policy.TargetType)),
		zap.String("target_name", policy.TargetName))

	// Simulate backup execution
	select {
	case <-ctx.Done():
		return 0, "", ctx.Err()
	case <-time.After(1500 * time.Millisecond): // Simulate 1.5-second backup
		// Return simulated results
		backupSize := int64(500000) // 500KB
		backupPath := fmt.Sprintf("/backups/%s/differential_%d.backup", policy.TargetName, time.Now().Unix())
		return backupSize, backupPath, nil
	}
}

// executeSnapshotBackup performs a snapshot backup
func (bo *BackupOrchestrator) executeSnapshotBackup(ctx context.Context, policy *models.BackupPolicy) (int64, string, error) {
	bo.logger.Info("Executing snapshot backup",
		zap.String("target_type", string(policy.TargetType)),
		zap.String("target_name", policy.TargetName))

	// Simulate backup execution
	select {
	case <-ctx.Done():
		return 0, "", ctx.Err()
	case <-time.After(500 * time.Millisecond): // Simulate 0.5-second backup
		// Return simulated results
		backupSize := int64(2000000) // 2MB
		backupPath := fmt.Sprintf("/backups/%s/snapshot_%d.backup", policy.TargetName, time.Now().Unix())
		return backupSize, backupPath, nil
	}
}

// calculateExpirationDate calculates when a backup instance should expire
func calculateExpirationDate(policy *models.BackupPolicy) *time.Time {
	// Use retention policy to calculate expiration
	if policy.RetentionPolicy.MaxAge > 0 {
		expirationDate := time.Now().AddDate(0, 0, policy.RetentionPolicy.MaxAge)
		return &expirationDate
	}

	// Default to 30 days if no retention policy is set
	expirationDate := time.Now().AddDate(0, 0, 30)
	return &expirationDate
}