package repository

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"p9e.in/ugcl/backupdr/models"
)

// backupDRRepository implements IBackupDRRepository
type BackupDRRepository struct {
	db *sql.DB
	// In-memory stores for development (replace with database when schema is ready)
	mu                sync.RWMutex
	policies          map[uuid.UUID]*models.BackupPolicy
	jobs              map[uuid.UUID]*models.BackupJob
	instances         map[uuid.UUID]*models.BackupInstance
	restoreRequests   map[uuid.UUID]*models.RestoreRequest
	drPlans           map[uuid.UUID]*models.DRPlan
	drTests           map[uuid.UUID]*models.DRTest
}

// NewBackupDRRepository creates a new backup DR repository
func NewBackupDRRepository(db *sql.DB) IBackupDRRepository {
	return &BackupDRRepository{
		db:              db,
		policies:        make(map[uuid.UUID]*models.BackupPolicy),
		jobs:            make(map[uuid.UUID]*models.BackupJob),
		instances:       make(map[uuid.UUID]*models.BackupInstance),
		restoreRequests: make(map[uuid.UUID]*models.RestoreRequest),
		drPlans:         make(map[uuid.UUID]*models.DRPlan),
		drTests:         make(map[uuid.UUID]*models.DRTest),
	}
}

// Backup Policy management

func (r *BackupDRRepository) CreateBackupPolicy(ctx context.Context, policy *models.BackupPolicy) (*models.BackupPolicy, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if policy.ID == uuid.Nil {
		policy.ID = uuid.New()
	}
	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()

	// Store in memory map
	r.policies[policy.ID] = policy
	return policy, nil
}

func (r *BackupDRRepository) GetBackupPolicyByID(ctx context.Context, id uuid.UUID) (*models.BackupPolicy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	policy, exists := r.policies[id]
	if !exists {
		return nil, fmt.Errorf("backup policy not found")
	}
	return policy, nil
}

func (r *BackupDRRepository) GetPoliciesByTarget(ctx context.Context, targetType models.TargetType, targetName string) ([]*models.BackupPolicy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.BackupPolicy
	for _, policy := range r.policies {
		if policy.TargetType == targetType && policy.TargetName == targetName && policy.IsActive {
			result = append(result, policy)
		}
	}
	return result, nil
}

func (r *BackupDRRepository) UpdateBackupPolicy(ctx context.Context, policy *models.BackupPolicy) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.policies[policy.ID]; !exists {
		return fmt.Errorf("backup policy not found")
	}

	policy.UpdatedAt = time.Now()
	r.policies[policy.ID] = policy
	return nil
}

func (r *BackupDRRepository) DeleteBackupPolicy(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.policies[id]; !exists {
		return fmt.Errorf("backup policy not found")
	}

	delete(r.policies, id)
	return nil
}

func (r *BackupDRRepository) ListBackupPolicies(ctx context.Context, filters *models.BackupPolicyFilters) ([]*models.BackupPolicy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.BackupPolicy
	for _, policy := range r.policies {
		if r.matchesPolicyFilters(policy, filters) {
			result = append(result, policy)
		}
	}
	return result, nil
}

func (r *BackupDRRepository) GetActivePolicies(ctx context.Context) ([]*models.BackupPolicy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.BackupPolicy
	for _, policy := range r.policies {
		if policy.IsActive {
			result = append(result, policy)
		}
	}
	return result, nil
}

// Helper method to match policy filters
func (r *BackupDRRepository) matchesPolicyFilters(policy *models.BackupPolicy, filters *models.BackupPolicyFilters) bool {
	if filters == nil {
		return true
	}

	if filters.TargetType != nil && *filters.TargetType != policy.TargetType {
		return false
	}

	if filters.BackupType != nil && *filters.BackupType != policy.BackupType {
		return false
	}

	if filters.IsActive != nil && *filters.IsActive != policy.IsActive {
		return false
	}

	if filters.ScheduleType != nil && *filters.ScheduleType != policy.ScheduleType {
		return false
	}

	return true
}

// Backup Job Management

func (r *BackupDRRepository) CreateBackupJob(ctx context.Context, job *models.BackupJob) (*models.BackupJob, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if job.ID == uuid.Nil {
		job.ID = uuid.New()
	}

	r.jobs[job.ID] = job
	return job, nil
}

func (r *BackupDRRepository) GetBackupJobByID(ctx context.Context, id uuid.UUID) (*models.BackupJob, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	job, exists := r.jobs[id]
	if !exists {
		return nil, fmt.Errorf("backup job not found")
	}
	return job, nil
}

func (r *BackupDRRepository) GetJobsByPolicy(ctx context.Context, policyID uuid.UUID) ([]*models.BackupJob, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.BackupJob
	for _, job := range r.jobs {
		if job.PolicyID == policyID {
			result = append(result, job)
		}
	}
	return result, nil
}

func (r *BackupDRRepository) UpdateBackupJob(ctx context.Context, job *models.BackupJob) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.jobs[job.ID]; !exists {
		return fmt.Errorf("backup job not found")
	}

	r.jobs[job.ID] = job
	return nil
}

func (r *BackupDRRepository) ListBackupJobs(ctx context.Context, filters *models.BackupJobFilters) ([]*models.BackupJob, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.BackupJob
	for _, job := range r.jobs {
		if r.matchesJobFilters(job, filters) {
			result = append(result, job)
		}
	}
	return result, nil
}

func (r *BackupDRRepository) GetActiveJobs(ctx context.Context) ([]*models.BackupJob, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.BackupJob
	for _, job := range r.jobs {
		if job.Status == models.JobStatusRunning || job.Status == models.JobStatusQueued || job.Status == models.JobStatusScheduled {
			result = append(result, job)
		}
	}
	return result, nil
}

func (r *BackupDRRepository) GetJobsForScheduling(ctx context.Context, nextRunBefore time.Time) ([]*models.BackupJob, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.BackupJob
	for _, job := range r.jobs {
		if job.Status == models.JobStatusScheduled && job.ScheduledAt.Before(nextRunBefore) {
			result = append(result, job)
		}
	}
	return result, nil
}

// Helper method to match job filters
func (r *BackupDRRepository) matchesJobFilters(job *models.BackupJob, filters *models.BackupJobFilters) bool {
	if filters == nil {
		return true
	}

	if filters.PolicyID != nil && *filters.PolicyID != job.PolicyID {
		return false
	}

	if filters.JobType != nil && *filters.JobType != job.JobType {
		return false
	}

	if filters.Status != nil && *filters.Status != job.Status {
		return false
	}

	if filters.Priority != nil && *filters.Priority != job.Priority {
		return false
	}

	if filters.ScheduledAfter != nil && job.ScheduledAt.Before(*filters.ScheduledAfter) {
		return false
	}

	if filters.ScheduledBefore != nil && job.ScheduledAt.After(*filters.ScheduledBefore) {
		return false
	}

	if filters.StartedAfter != nil && (job.StartedAt == nil || job.StartedAt.Before(*filters.StartedAfter)) {
		return false
	}

	if filters.StartedBefore != nil && (job.StartedAt == nil || job.StartedAt.After(*filters.StartedBefore)) {
		return false
	}

	return true
}

// Backup Instance Management

func (r *BackupDRRepository) CreateBackupInstance(ctx context.Context, instance *models.BackupInstance) (*models.BackupInstance, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if instance.ID == uuid.Nil {
		instance.ID = uuid.New()
	}
	instance.CreatedAt = time.Now()

	r.instances[instance.ID] = instance
	return instance, nil
}

func (r *BackupDRRepository) GetBackupInstanceByID(ctx context.Context, id uuid.UUID) (*models.BackupInstance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	instance, exists := r.instances[id]
	if !exists {
		return nil, fmt.Errorf("backup instance not found")
	}
	return instance, nil
}

func (r *BackupDRRepository) GetInstancesByJob(ctx context.Context, jobID uuid.UUID) ([]*models.BackupInstance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.BackupInstance
	for _, instance := range r.instances {
		if instance.JobID == jobID {
			result = append(result, instance)
		}
	}
	return result, nil
}

func (r *BackupDRRepository) GetInstancesByDateRange(ctx context.Context, jobID uuid.UUID, startDate, endDate time.Time) ([]*models.BackupInstance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.BackupInstance
	for _, instance := range r.instances {
		if instance.JobID == jobID &&
			!instance.CreatedAt.Before(startDate) &&
			!instance.CreatedAt.After(endDate) {
			result = append(result, instance)
		}
	}
	return result, nil
}

func (r *BackupDRRepository) UpdateBackupInstance(ctx context.Context, instance *models.BackupInstance) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.instances[instance.ID]; !exists {
		return fmt.Errorf("backup instance not found")
	}

	r.instances[instance.ID] = instance
	return nil
}

func (r *BackupDRRepository) DeleteBackupInstance(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.instances[id]; !exists {
		return fmt.Errorf("backup instance not found")
	}

	delete(r.instances, id)
	return nil
}

func (r *BackupDRRepository) ListBackupInstances(ctx context.Context, filters *models.BackupInstanceFilters) ([]*models.BackupInstance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.BackupInstance
	for _, instance := range r.instances {
		if r.matchesInstanceFilters(instance, filters) {
			result = append(result, instance)
		}
	}
	return result, nil
}

func (r *BackupDRRepository) GetExpiredInstances(ctx context.Context, beforeDate time.Time) ([]*models.BackupInstance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.BackupInstance
	for _, instance := range r.instances {
		if instance.ExpiresAt != nil && instance.ExpiresAt.Before(beforeDate) {
			result = append(result, instance)
		}
	}
	return result, nil
}

// Helper method to match instance filters
func (r *BackupDRRepository) matchesInstanceFilters(instance *models.BackupInstance, filters *models.BackupInstanceFilters) bool {
	if filters == nil {
		return true
	}

	if filters.JobID != nil && *filters.JobID != instance.JobID {
		return false
	}

	if filters.Status != nil && *filters.Status != instance.Status {
		return false
	}

	if filters.CreatedAfter != nil && instance.CreatedAt.Before(*filters.CreatedAfter) {
		return false
	}

	if filters.CreatedBefore != nil && instance.CreatedAt.After(*filters.CreatedBefore) {
		return false
	}

	if filters.ExpiresAfter != nil && (instance.ExpiresAt == nil || instance.ExpiresAt.Before(*filters.ExpiresAfter)) {
		return false
	}

	if filters.ExpiresBefore != nil && (instance.ExpiresAt == nil || instance.ExpiresAt.After(*filters.ExpiresBefore)) {
		return false
	}

	return true
}

// Restore Request Management

func (r *BackupDRRepository) CreateRestoreRequest(ctx context.Context, request *models.RestoreRequest) (*models.RestoreRequest, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if request.ID == uuid.Nil {
		request.ID = uuid.New()
	}
	request.CreatedAt = time.Now()

	r.restoreRequests[request.ID] = request
	return request, nil
}

func (r *BackupDRRepository) GetRestoreRequestByID(ctx context.Context, id uuid.UUID) (*models.RestoreRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	request, exists := r.restoreRequests[id]
	if !exists {
		return nil, fmt.Errorf("restore request not found")
	}
	return request, nil
}

func (r *BackupDRRepository) UpdateRestoreRequest(ctx context.Context, request *models.RestoreRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.restoreRequests[request.ID]; !exists {
		return fmt.Errorf("restore request not found")
	}

	r.restoreRequests[request.ID] = request
	return nil
}

func (r *BackupDRRepository) ListRestoreRequests(ctx context.Context, filters *models.RestoreRequestFilters) ([]*models.RestoreRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.RestoreRequest
	for _, request := range r.restoreRequests {
		if r.matchesRestoreRequestFilters(request, filters) {
			result = append(result, request)
		}
	}
	return result, nil
}

func (r *BackupDRRepository) GetPendingRestoreRequests(ctx context.Context) ([]*models.RestoreRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.RestoreRequest
	for _, request := range r.restoreRequests {
		if request.Status == models.RestoreStatusRequested {
			result = append(result, request)
		}
	}
	return result, nil
}

// Helper method to match restore request filters
func (r *BackupDRRepository) matchesRestoreRequestFilters(request *models.RestoreRequest, filters *models.RestoreRequestFilters) bool {
	if filters == nil {
		return true
	}

	if filters.BackupJobID != nil && *filters.BackupJobID != request.BackupJobID {
		return false
	}

	if filters.RequestedBy != nil && *filters.RequestedBy != request.RequestedBy {
		return false
	}

	if filters.Status != nil && *filters.Status != request.Status {
		return false
	}

	if filters.RestoreType != nil && *filters.RestoreType != request.RestoreType {
		return false
	}

	if filters.RestoreScope != nil && *filters.RestoreScope != request.RestoreScope {
		return false
	}

	if filters.RequiresApproval != nil && *filters.RequiresApproval != request.RequiresApproval {
		return false
	}

	if filters.CreatedAfter != nil && request.CreatedAt.Before(*filters.CreatedAfter) {
		return false
	}

	if filters.CreatedBefore != nil && request.CreatedAt.After(*filters.CreatedBefore) {
		return false
	}

	return true
}

// DR Plan Management

func (r *BackupDRRepository) CreateDRPlan(ctx context.Context, plan *models.DRPlan) (*models.DRPlan, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if plan.ID == uuid.Nil {
		plan.ID = uuid.New()
	}
	plan.CreatedAt = time.Now()
	plan.UpdatedAt = time.Now()

	r.drPlans[plan.ID] = plan
	return plan, nil
}

func (r *BackupDRRepository) GetDRPlanByID(ctx context.Context, id uuid.UUID) (*models.DRPlan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plan, exists := r.drPlans[id]
	if !exists {
		return nil, fmt.Errorf("DR plan not found")
	}
	return plan, nil
}

func (r *BackupDRRepository) UpdateDRPlan(ctx context.Context, plan *models.DRPlan) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.drPlans[plan.ID]; !exists {
		return fmt.Errorf("DR plan not found")
	}

	plan.UpdatedAt = time.Now()
	r.drPlans[plan.ID] = plan
	return nil
}

func (r *BackupDRRepository) DeleteDRPlan(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.drPlans[id]; !exists {
		return fmt.Errorf("DR plan not found")
	}

	delete(r.drPlans, id)
	return nil
}

func (r *BackupDRRepository) ListDRPlans(ctx context.Context, filters *models.DRPlanFilters) ([]*models.DRPlan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.DRPlan
	for _, plan := range r.drPlans {
		if r.matchesDRPlanFilters(plan, filters) {
			result = append(result, plan)
		}
	}
	return result, nil
}

func (r *BackupDRRepository) GetActiveDRPlans(ctx context.Context) ([]*models.DRPlan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.DRPlan
	for _, plan := range r.drPlans {
		if plan.IsActive {
			result = append(result, plan)
		}
	}
	return result, nil
}

// Helper method to match DR plan filters
func (r *BackupDRRepository) matchesDRPlanFilters(plan *models.DRPlan, filters *models.DRPlanFilters) bool {
	if filters == nil {
		return true
	}

	if filters.PlanType != nil && *filters.PlanType != plan.PlanType {
		return false
	}

	if filters.Severity != nil && *filters.Severity != plan.Severity {
		return false
	}

	if filters.Scope != nil && *filters.Scope != plan.Scope {
		return false
	}

	if filters.IsActive != nil && *filters.IsActive != plan.IsActive {
		return false
	}

	if filters.Version != nil && *filters.Version != plan.Version {
		return false
	}

	return true
}

// DR Test Management

func (r *BackupDRRepository) CreateDRTest(ctx context.Context, test *models.DRTest) (*models.DRTest, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if test.ID == uuid.Nil {
		test.ID = uuid.New()
	}
	test.CreatedAt = time.Now()

	r.drTests[test.ID] = test
	return test, nil
}

func (r *BackupDRRepository) GetDRTestByID(ctx context.Context, id uuid.UUID) (*models.DRTest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	test, exists := r.drTests[id]
	if !exists {
		return nil, fmt.Errorf("DR test not found")
	}
	return test, nil
}

func (r *BackupDRRepository) GetTestsByPlan(ctx context.Context, planID uuid.UUID) ([]*models.DRTest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.DRTest
	for _, test := range r.drTests {
		if test.PlanID == planID {
			result = append(result, test)
		}
	}
	return result, nil
}

func (r *BackupDRRepository) UpdateDRTest(ctx context.Context, test *models.DRTest) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.drTests[test.ID]; !exists {
		return fmt.Errorf("DR test not found")
	}

	r.drTests[test.ID] = test
	return nil
}

func (r *BackupDRRepository) ListDRTests(ctx context.Context, filters *models.DRTestFilters) ([]*models.DRTest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.DRTest
	for _, test := range r.drTests {
		if r.matchesDRTestFilters(test, filters) {
			result = append(result, test)
		}
	}
	return result, nil
}

func (r *BackupDRRepository) GetOverdueTests(ctx context.Context) ([]*models.DRTest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.DRTest
	now := time.Now()
	for _, test := range r.drTests {
		if test.Status == models.DRTestStatusScheduled && test.ScheduledAt.Before(now) {
			result = append(result, test)
		}
	}
	return result, nil
}

// Helper method to match DR test filters
func (r *BackupDRRepository) matchesDRTestFilters(test *models.DRTest, filters *models.DRTestFilters) bool {
	if filters == nil {
		return true
	}

	if filters.PlanID != nil && *filters.PlanID != test.PlanID {
		return false
	}

	if filters.TestType != nil && *filters.TestType != test.TestType {
		return false
	}

	if filters.Status != nil && *filters.Status != test.Status {
		return false
	}

	if filters.ScheduledAfter != nil && test.ScheduledAt.Before(*filters.ScheduledAfter) {
		return false
	}

	if filters.ScheduledBefore != nil && test.ScheduledAt.After(*filters.ScheduledBefore) {
		return false
	}

	if filters.ExecutedBy != nil && *filters.ExecutedBy != test.ExecutedBy {
		return false
	}

	return true
}

// Analytics and Reporting

func (r *BackupDRRepository) GetBackupStats(ctx context.Context, filters *models.BackupStatsFilters) (*models.BackupStatistics, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Calculate statistics from in-memory data
	var totalPolicies, activePolicies int64
	for _, policy := range r.policies {
		totalPolicies++
		if policy.IsActive {
			activePolicies++
		}
	}

	var totalJobs, successfulJobs, failedJobs int64
	var totalBackupSize int64
	var totalDuration time.Duration
	var jobCount int

	for _, job := range r.jobs {
		if r.matchesStatsFilters(job, filters) {
			totalJobs++
			totalBackupSize += job.BackupSize
			if job.Duration != nil {
				totalDuration += *job.Duration
				jobCount++
			}
			if job.Status == models.JobStatusCompleted {
				successfulJobs++
			} else if job.Status == models.JobStatusFailed {
				failedJobs++
			}
		}
	}

	var averageBackupTime time.Duration
	if jobCount > 0 {
		averageBackupTime = totalDuration / time.Duration(jobCount)
	}

	return &models.BackupStatistics{
		TotalPolicies:     totalPolicies,
		ActivePolicies:    activePolicies,
		TotalJobs:         totalJobs,
		SuccessfulJobs:    successfulJobs,
		FailedJobs:        failedJobs,
		TotalBackupSize:   totalBackupSize,
		AverageBackupTime: averageBackupTime,
		CompressionRatio:  0.7, // Default compression ratio
		StorageUtilization: 0.8, // Default utilization
		LastCalculatedAt:  time.Now(),
	}, nil
}

func (r *BackupDRRepository) GetStorageUtilization(ctx context.Context, targetType models.TargetType) ([]*models.StorageUtilization, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Group by target name and calculate utilization
	utilization := make(map[string]*models.StorageUtilization)

	for _, policy := range r.policies {
		if policy.TargetType == targetType {
			if _, exists := utilization[policy.TargetName]; !exists {
				utilization[policy.TargetName] = &models.StorageUtilization{
					TargetType:    targetType,
					TargetName:    policy.TargetName,
					TotalSize:     1000000000, // 1GB default
					UsedSize:      0,
					LastUpdated:   time.Now(),
				}
			}
		}
	}

	// Calculate used size from jobs
	for _, job := range r.jobs {
		if policy, exists := r.policies[job.PolicyID]; exists && policy.TargetType == targetType {
			if util, exists := utilization[policy.TargetName]; exists {
				util.UsedSize += job.BackupSize
				util.BackupCount++
				if util.OldestBackup == nil || job.ScheduledAt.Before(*util.OldestBackup) {
					util.OldestBackup = &job.ScheduledAt
				}
				if util.NewestBackup == nil || job.ScheduledAt.After(*util.NewestBackup) {
					util.NewestBackup = &job.ScheduledAt
				}
			}
		}
	}

	// Calculate utilization percentages and available space
	var result []*models.StorageUtilization
	for _, util := range utilization {
		util.AvailableSize = util.TotalSize - util.UsedSize
		if util.TotalSize > 0 {
			util.UtilizationPct = float64(util.UsedSize) / float64(util.TotalSize) * 100
		}
		result = append(result, util)
	}

	return result, nil
}

func (r *BackupDRRepository) GetComplianceReport(ctx context.Context, filters *models.ComplianceFilters) (*models.ComplianceReport, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// For simplicity, generate a compliance report for the first matching policy
	for _, policy := range r.policies {
		if r.matchesComplianceFilters(policy, filters) {
			// Calculate compliance metrics
			var backupCount, failureCount int64
			var lastSuccessfulBackup *time.Time
			var nextScheduledBackup *time.Time

			for _, job := range r.jobs {
				if job.PolicyID == policy.ID {
					backupCount++
					if job.Status == models.JobStatusCompleted {
						if lastSuccessfulBackup == nil || (job.CompletedAt != nil && job.CompletedAt.After(*lastSuccessfulBackup)) {
							lastSuccessfulBackup = job.CompletedAt
						}
					} else if job.Status == models.JobStatusFailed {
						failureCount++
					}

					if job.Status == models.JobStatusScheduled && (nextScheduledBackup == nil || job.ScheduledAt.Before(*nextScheduledBackup)) {
						nextScheduledBackup = &job.ScheduledAt
					}
				}
			}

			successRate := float64(0)
			if backupCount > 0 {
				successRate = float64(backupCount-failureCount) / float64(backupCount) * 100
			}

			// Determine compliance status
			complianceStatus := models.ComplianceStatusCompliant
			var issues, recommendations []string

			if successRate < 95 {
				complianceStatus = models.ComplianceStatusNonCompliant
				issues = append(issues, "Success rate below 95%")
				recommendations = append(recommendations, "Review backup configuration and address failures")
			} else if successRate < 99 {
				complianceStatus = models.ComplianceStatusWarning
				issues = append(issues, "Success rate below 99%")
				recommendations = append(recommendations, "Monitor backup performance closely")
			}

			return &models.ComplianceReport{
				PolicyID:             policy.ID,
				PolicyName:           policy.Name,
				ComplianceStatus:     complianceStatus,
				LastSuccessfulBackup: lastSuccessfulBackup,
				NextScheduledBackup:  nextScheduledBackup,
				BackupCount:          backupCount,
				FailureCount:         failureCount,
				SuccessRate:          successRate,
				RTOCompliance:        true, // Default to compliant
				RPOCompliance:        true, // Default to compliant
				Issues:               issues,
				Recommendations:      recommendations,
				LastChecked:          time.Now(),
			}, nil
		}
	}

	return nil, fmt.Errorf("no matching policy found for compliance report")
}

// Helper method to match stats filters
func (r *BackupDRRepository) matchesStatsFilters(job *models.BackupJob, filters *models.BackupStatsFilters) bool {
	if filters == nil {
		return true
	}

	if filters.PolicyID != nil && *filters.PolicyID != job.PolicyID {
		return false
	}

	if filters.BackupType != nil && *filters.BackupType != job.JobType {
		return false
	}

	if !job.ScheduledAt.After(filters.StartDate) || !job.ScheduledAt.Before(filters.EndDate) {
		return false
	}

	if filters.TargetType != nil {
		if policy, exists := r.policies[job.PolicyID]; !exists || policy.TargetType != *filters.TargetType {
			return false
		}
	}

	return true
}

// Helper method to match compliance filters
func (r *BackupDRRepository) matchesComplianceFilters(policy *models.BackupPolicy, filters *models.ComplianceFilters) bool {
	if filters == nil {
		return true
	}

	if filters.PolicyID != nil && *filters.PolicyID != policy.ID {
		return false
	}

	if filters.TargetType != nil && *filters.TargetType != policy.TargetType {
		return false
	}

	// For CheckedAfter/CheckedBefore, we'll use policy creation time as proxy
	if filters.CheckedAfter != nil && policy.CreatedAt.Before(*filters.CheckedAfter) {
		return false
	}

	if filters.CheckedBefore != nil && policy.CreatedAt.After(*filters.CheckedBefore) {
		return false
	}

	return true
}