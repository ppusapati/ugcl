package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"p9e.in/ugcl/backupdr/models"
)

// IBackupDRRepository defines the interface for backup and disaster recovery data operations
type IBackupDRRepository interface {
	// Backup Policy management
	CreateBackupPolicy(ctx context.Context, policy *models.BackupPolicy) (*models.BackupPolicy, error)
	GetBackupPolicyByID(ctx context.Context, id uuid.UUID) (*models.BackupPolicy, error)
	GetPoliciesByTarget(ctx context.Context, targetType models.TargetType, targetName string) ([]*models.BackupPolicy, error)
	UpdateBackupPolicy(ctx context.Context, policy *models.BackupPolicy) error
	DeleteBackupPolicy(ctx context.Context, id uuid.UUID) error
	ListBackupPolicies(ctx context.Context, filters *models.BackupPolicyFilters) ([]*models.BackupPolicy, error)
	GetActivePolicies(ctx context.Context) ([]*models.BackupPolicy, error)

	// Backup Job management
	CreateBackupJob(ctx context.Context, job *models.BackupJob) (*models.BackupJob, error)
	GetBackupJobByID(ctx context.Context, id uuid.UUID) (*models.BackupJob, error)
	GetJobsByPolicy(ctx context.Context, policyID uuid.UUID) ([]*models.BackupJob, error)
	UpdateBackupJob(ctx context.Context, job *models.BackupJob) error
	ListBackupJobs(ctx context.Context, filters *models.BackupJobFilters) ([]*models.BackupJob, error)
	GetActiveJobs(ctx context.Context) ([]*models.BackupJob, error)
	GetJobsForScheduling(ctx context.Context, nextRunBefore time.Time) ([]*models.BackupJob, error)

	// Backup Instance management
	CreateBackupInstance(ctx context.Context, instance *models.BackupInstance) (*models.BackupInstance, error)
	GetBackupInstanceByID(ctx context.Context, id uuid.UUID) (*models.BackupInstance, error)
	GetInstancesByJob(ctx context.Context, jobID uuid.UUID) ([]*models.BackupInstance, error)
	GetInstancesByDateRange(ctx context.Context, jobID uuid.UUID, startDate, endDate time.Time) ([]*models.BackupInstance, error)
	UpdateBackupInstance(ctx context.Context, instance *models.BackupInstance) error
	DeleteBackupInstance(ctx context.Context, id uuid.UUID) error
	ListBackupInstances(ctx context.Context, filters *models.BackupInstanceFilters) ([]*models.BackupInstance, error)
	GetExpiredInstances(ctx context.Context, beforeDate time.Time) ([]*models.BackupInstance, error)

	// Restore Request management
	CreateRestoreRequest(ctx context.Context, request *models.RestoreRequest) (*models.RestoreRequest, error)
	GetRestoreRequestByID(ctx context.Context, id uuid.UUID) (*models.RestoreRequest, error)
	UpdateRestoreRequest(ctx context.Context, request *models.RestoreRequest) error
	ListRestoreRequests(ctx context.Context, filters *models.RestoreRequestFilters) ([]*models.RestoreRequest, error)
	GetPendingRestoreRequests(ctx context.Context) ([]*models.RestoreRequest, error)

	// Disaster Recovery Plan management
	CreateDRPlan(ctx context.Context, plan *models.DRPlan) (*models.DRPlan, error)
	GetDRPlanByID(ctx context.Context, id uuid.UUID) (*models.DRPlan, error)
	UpdateDRPlan(ctx context.Context, plan *models.DRPlan) error
	DeleteDRPlan(ctx context.Context, id uuid.UUID) error
	ListDRPlans(ctx context.Context, filters *models.DRPlanFilters) ([]*models.DRPlan, error)
	GetActiveDRPlans(ctx context.Context) ([]*models.DRPlan, error)

	// DR Test management
	CreateDRTest(ctx context.Context, test *models.DRTest) (*models.DRTest, error)
	GetDRTestByID(ctx context.Context, id uuid.UUID) (*models.DRTest, error)
	GetTestsByPlan(ctx context.Context, planID uuid.UUID) ([]*models.DRTest, error)
	UpdateDRTest(ctx context.Context, test *models.DRTest) error
	ListDRTests(ctx context.Context, filters *models.DRTestFilters) ([]*models.DRTest, error)
	GetOverdueTests(ctx context.Context) ([]*models.DRTest, error)

	// Analytics and reporting
	GetBackupStats(ctx context.Context, filters *models.BackupStatsFilters) (*models.BackupStatistics, error)
	GetStorageUtilization(ctx context.Context, targetType models.TargetType) ([]*models.StorageUtilization, error)
	GetComplianceReport(ctx context.Context, filters *models.ComplianceFilters) (*models.ComplianceReport, error)
}