package interfaces

import (
	"context"
	"time"

	"github.com/google/uuid"
	"p9e.in/ugcl/dataarchive/models"
)

type DataArchiveService interface {
	// Retention Policy management
	CreateRetentionPolicy(ctx context.Context, req *CreateRetentionPolicyRequest) (*models.RetentionPolicy, error)
	GetRetentionPolicy(ctx context.Context, id uuid.UUID) (*models.RetentionPolicy, error)
	UpdateRetentionPolicy(ctx context.Context, req *UpdateRetentionPolicyRequest) (*models.RetentionPolicy, error)
	DeleteRetentionPolicy(ctx context.Context, id uuid.UUID) error
	ListRetentionPolicies(ctx context.Context, req *ListRetentionPoliciesRequest) ([]*models.RetentionPolicy, int32, string, error)
	ValidateRetentionPolicy(ctx context.Context, policy *models.RetentionPolicy) error

	// Archival operations
	StartArchivalJob(ctx context.Context, req *StartArchivalJobRequest) (*models.ArchivalJob, error)
	MonitorArchivalJob(ctx context.Context, jobID uuid.UUID) (*models.ArchivalJob, error)
	CancelArchivalJob(ctx context.Context, jobID uuid.UUID, reason string) error
	RetryFailedJob(ctx context.Context, jobID uuid.UUID) (*models.ArchivalJob, error)
	GetArchivalJobStatus(ctx context.Context, jobID uuid.UUID) (*ArchivalJobStatus, error)
	ListArchivalJobs(ctx context.Context, req *ListArchivalJobsRequest) ([]*models.ArchivalJob, int32, string, error)

	// Data restoration
	RequestDataRestore(ctx context.Context, req *RestoreDataRequest) (*RestoreJob, error)
	GetRestoreJobStatus(ctx context.Context, jobID uuid.UUID) (*RestoreJob, error)
	ListRestoreJobs(ctx context.Context, req *ListRestoreJobsRequest) ([]*RestoreJob, int32, string, error)

	// Legal Hold management
	CreateLegalHold(ctx context.Context, req *CreateLegalHoldRequest) (*models.LegalHold, error)
	UpdateLegalHold(ctx context.Context, req *UpdateLegalHoldRequest) (*models.LegalHold, error)
	ReleaseLegalHold(ctx context.Context, id uuid.UUID, reason string) error
	ListLegalHolds(ctx context.Context, req *ListLegalHoldsRequest) ([]*models.LegalHold, int32, string, error)
	CheckLegalHoldStatus(ctx context.Context, entityType string, entityIDs []string) ([]*LegalHoldStatus, error)

	// Data inventory and discovery
	ScanDataInventory(ctx context.Context, req *ScanInventoryRequest) (*ScanJob, error)
	GetDataInventory(ctx context.Context, req *GetInventoryRequest) ([]*models.DataInventory, error)
	ClassifyData(ctx context.Context, req *ClassifyDataRequest) (*ClassificationResult, error)
	GetDataClassificationReport(ctx context.Context, req *ClassificationReportRequest) (*ClassificationReport, error)

	// Compliance and auditing
	RunComplianceAudit(ctx context.Context, req *ComplianceAuditRequest) (*models.ComplianceAudit, error)
	GetComplianceReport(ctx context.Context, req *ComplianceReportRequest) (*ComplianceReport, error)
	GetRetentionMetrics(ctx context.Context, req *RetentionMetricsRequest) (*RetentionMetrics, error)
	ValidateCompliance(ctx context.Context, req *ValidateComplianceRequest) (*ComplianceValidation, error)

	// Scheduled operations
	ProcessScheduledPolicies(ctx context.Context) (*ScheduleProcessResult, error)
	GetNextScheduledExecutions(ctx context.Context, limit int) ([]*ScheduledExecution, error)
	UpdatePolicySchedule(ctx context.Context, policyID uuid.UUID, schedule string) error

	// Storage management
	EstimateStorageSavings(ctx context.Context, req *StorageEstimateRequest) (*StorageEstimate, error)
	GetStorageStatistics(ctx context.Context, req *StorageStatsRequest) (*StorageStatistics, error)
	OptimizeStorage(ctx context.Context, req *StorageOptimizationRequest) (*StorageOptimizationResult, error)
}

// DTOs and Request/Response types
type CreateRetentionPolicyRequest struct {
	Name            string                    `json:"name"`
	Description     string                    `json:"description"`
	EntityType      string                    `json:"entity_type"`
	SchemaName      string                    `json:"schema_name"`
	DatabaseName    string                    `json:"database_name"`
	RetentionPeriod time.Duration             `json:"retention_period"`
	ArchivalPeriod  time.Duration             `json:"archival_period"`
	GracePeriod     time.Duration             `json:"grace_period"`
	PolicyType      models.PolicyType         `json:"policy_type"`
	ArchivalMethod  models.ArchivalMethod     `json:"archival_method"`
	CompressionType models.CompressionType    `json:"compression_type"`
	EncryptArchive  bool                      `json:"encrypt_archive"`
	ComplianceLevel models.ComplianceLevel    `json:"compliance_level"`
	RegulatoryBasis []string                  `json:"regulatory_basis"`
	ScheduleCron    string                    `json:"schedule_cron"`
	BatchSize       int                       `json:"batch_size"`
	ParallelJobs    int                       `json:"parallel_jobs"`
}

type UpdateRetentionPolicyRequest struct {
	ID              uuid.UUID                 `json:"id"`
	Name            string                    `json:"name"`
	Description     string                    `json:"description"`
	RetentionPeriod time.Duration             `json:"retention_period"`
	ArchivalPeriod  time.Duration             `json:"archival_period"`
	GracePeriod     time.Duration             `json:"grace_period"`
	PolicyType      models.PolicyType         `json:"policy_type"`
	ArchivalMethod  models.ArchivalMethod     `json:"archival_method"`
	CompressionType models.CompressionType    `json:"compression_type"`
	EncryptArchive  bool                      `json:"encrypt_archive"`
	ComplianceLevel models.ComplianceLevel    `json:"compliance_level"`
	RegulatoryBasis []string                  `json:"regulatory_basis"`
	ScheduleCron    string                    `json:"schedule_cron"`
	IsActive        bool                      `json:"is_active"`
}

type ListRetentionPoliciesRequest struct {
	PageSize     int32  `json:"page_size"`
	PageToken    string `json:"page_token"`
	EntityType   string `json:"entity_type"`
	DatabaseName string `json:"database_name"`
	ActiveOnly   bool   `json:"active_only"`
}

type StartArchivalJobRequest struct {
	PolicyID    uuid.UUID          `json:"policy_id"`
	JobType     models.JobType     `json:"job_type"`
	TargetTable string             `json:"target_table"`
	DateRange   *models.DateRange  `json:"date_range"`
	BatchSize   int                `json:"batch_size"`
	DryRun      bool               `json:"dry_run"`
}

type ArchivalJobStatus struct {
	ID               uuid.UUID    `json:"id"`
	Status           models.JobStatus `json:"status"`
	ProgressPercent  float64      `json:"progress_percent"`
	TotalRecords     int64        `json:"total_records"`
	ProcessedRecords int64        `json:"processed_records"`
	ArchivedRecords  int64        `json:"archived_records"`
	EstimatedFinish  *time.Time   `json:"estimated_finish"`
	ErrorMessage     string       `json:"error_message"`
}

type ListArchivalJobsRequest struct {
	PageSize      int32               `json:"page_size"`
	PageToken     string              `json:"page_token"`
	PolicyID      *uuid.UUID          `json:"policy_id"`
	Status        *models.JobStatus   `json:"status"`
	StartedAfter  *time.Time          `json:"started_after"`
	StartedBefore *time.Time          `json:"started_before"`
}

type RestoreDataRequest struct {
	ArchiveID        uuid.UUID              `json:"archive_id"`
	RestoreType      models.RestoreType     `json:"restore_type"`
	TargetLocation   string                 `json:"target_location"`
	OverwritePolicy  models.OverwritePolicy `json:"overwrite_policy"`
	IncludeFilters   []string               `json:"include_filters"`
	ExcludeFilters   []string               `json:"exclude_filters"`
	PointInTime      *time.Time             `json:"point_in_time"`
	RequiresApproval bool                   `json:"requires_approval"`
}

type RestoreJob struct {
	ID              uuid.UUID              `json:"id"`
	ArchiveID       uuid.UUID              `json:"archive_id"`
	Status          models.RestoreStatus   `json:"status"`
	RestoreType     models.RestoreType     `json:"restore_type"`
	TargetLocation  string                 `json:"target_location"`
	ProgressPercent float64                `json:"progress_percent"`
	RestoredFiles   int64                  `json:"restored_files"`
	RestoredSize    int64                  `json:"restored_size"`
	EstimatedFinish *time.Time             `json:"estimated_finish"`
	ErrorMessage    string                 `json:"error_message"`
	CreatedAt       time.Time              `json:"created_at"`
	StartedAt       *time.Time             `json:"started_at"`
	CompletedAt     *time.Time             `json:"completed_at"`
}

type ListRestoreJobsRequest struct {
	PageSize      int32                  `json:"page_size"`
	PageToken     string                 `json:"page_token"`
	ArchiveID     *uuid.UUID             `json:"archive_id"`
	Status        *models.RestoreStatus  `json:"status"`
	RequestedBy   *uuid.UUID             `json:"requested_by"`
	CreatedAfter  *time.Time             `json:"created_after"`
	CreatedBefore *time.Time             `json:"created_before"`
}

type CreateLegalHoldRequest struct {
	Name             string             `json:"name"`
	Description      string             `json:"description"`
	EntityTypes      []string           `json:"entity_types"`
	EntityIDs        []string           `json:"entity_ids"`
	DateRange        *models.DateRange  `json:"date_range"`
	CaseNumber       string             `json:"case_number"`
	LegalBasis       string             `json:"legal_basis"`
	IssuingAuthority string             `json:"issuing_authority"`
	ContactInfo      string             `json:"contact_info"`
	StartDate        time.Time          `json:"start_date"`
	EndDate          *time.Time         `json:"end_date"`
}

type UpdateLegalHoldRequest struct {
	ID               uuid.UUID          `json:"id"`
	Name             string             `json:"name"`
	Description      string             `json:"description"`
	EntityTypes      []string           `json:"entity_types"`
	EntityIDs        []string           `json:"entity_ids"`
	DateRange        *models.DateRange  `json:"date_range"`
	IssuingAuthority string             `json:"issuing_authority"`
	ContactInfo      string             `json:"contact_info"`
	EndDate          *time.Time         `json:"end_date"`
	IsActive         bool               `json:"is_active"`
}

type ListLegalHoldsRequest struct {
	PageSize    int32      `json:"page_size"`
	PageToken   string     `json:"page_token"`
	EntityTypes []string   `json:"entity_types"`
	CaseNumber  string     `json:"case_number"`
	ActiveOnly  bool       `json:"active_only"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
}

type LegalHoldStatus struct {
	EntityType string               `json:"entity_type"`
	EntityID   string               `json:"entity_id"`
	IsOnHold   bool                 `json:"is_on_hold"`
	HoldIDs    []uuid.UUID          `json:"hold_ids"`
	HoldNames  []string             `json:"hold_names"`
}

type ScanInventoryRequest struct {
	DatabaseName string   `json:"database_name"`
	SchemaNames  []string `json:"schema_names"`
	TableNames   []string `json:"table_names"`
	ScanMethod   string   `json:"scan_method"`
	DeepScan     bool     `json:"deep_scan"`
}

type ScanJob struct {
	ID              uuid.UUID `json:"id"`
	Status          string    `json:"status"`
	ProgressPercent float64   `json:"progress_percent"`
	TablesScanned   int64     `json:"tables_scanned"`
	TotalTables     int64     `json:"total_tables"`
	StartedAt       time.Time `json:"started_at"`
	EstimatedFinish *time.Time `json:"estimated_finish"`
}

type GetInventoryRequest struct {
	DatabaseName       string                     `json:"database_name"`
	SchemaName         string                     `json:"schema_name"`
	TableName          string                     `json:"table_name"`
	DataClassification *models.DataClassification `json:"data_classification"`
	ContainsPII        *bool                      `json:"contains_pii"`
	ContainsPHI        *bool                      `json:"contains_phi"`
	ContainsPCI        *bool                      `json:"contains_pci"`
}

type ClassifyDataRequest struct {
	DatabaseName string   `json:"database_name"`
	SchemaName   string   `json:"schema_name"`
	TableName    string   `json:"table_name"`
	ColumnNames  []string `json:"column_names"`
	SampleSize   int      `json:"sample_size"`
}

type ClassificationResult struct {
	TableClassification   models.DataClassification `json:"table_classification"`
	ColumnClassifications map[string]models.DataClassification `json:"column_classifications"`
	PIIColumns           []string                   `json:"pii_columns"`
	PHIColumns           []string                   `json:"phi_columns"`
	PCIColumns           []string                   `json:"pci_columns"`
	Confidence           float64                    `json:"confidence"`
}

type ClassificationReportRequest struct {
	DatabaseName string                     `json:"database_name"`
	SchemaName   string                     `json:"schema_name"`
	Classification *models.DataClassification `json:"classification"`
	IncludePII   bool                       `json:"include_pii"`
	IncludePHI   bool                       `json:"include_phi"`
	IncludePCI   bool                       `json:"include_pci"`
}

type ClassificationReport struct {
	TotalTables       int64                              `json:"total_tables"`
	ClassifiedTables  int64                              `json:"classified_tables"`
	UnclassifiedTables int64                             `json:"unclassified_tables"`
	Classifications   map[string]int64                   `json:"classifications"`
	PIITables         []string                           `json:"pii_tables"`
	PHITables         []string                           `json:"phi_tables"`
	PCITables         []string                           `json:"pci_tables"`
	RiskScore         float64                            `json:"risk_score"`
}

type ComplianceAuditRequest struct {
	AuditType   models.AuditType  `json:"audit_type"`
	PolicyIDs   []uuid.UUID       `json:"policy_ids"`
	EntityTypes []string          `json:"entity_types"`
	DateRange   *models.DateRange `json:"date_range"`
	DeepScan    bool              `json:"deep_scan"`
}

type ComplianceReport struct {
	AuditID         uuid.UUID                `json:"audit_id"`
	ComplianceScore float64                  `json:"compliance_score"`
	TotalPolicies   int64                    `json:"total_policies"`
	CompliantPolicies int64                  `json:"compliant_policies"`
	ViolationCount  int64                    `json:"violation_count"`
	Violations      []ComplianceViolation    `json:"violations"`
	Recommendations []ComplianceRecommendation `json:"recommendations"`
	GeneratedAt     time.Time                `json:"generated_at"`
}

type ComplianceViolation struct {
	PolicyID    uuid.UUID `json:"policy_id"`
	PolicyName  string    `json:"policy_name"`
	ViolationType string  `json:"violation_type"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	EntityType  string    `json:"entity_type"`
	EntityID    string    `json:"entity_id"`
}

type ComplianceRecommendation struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	ActionType  string `json:"action_type"`
}

type ComplianceReportRequest struct {
	PolicyIDs   []uuid.UUID       `json:"policy_ids"`
	EntityTypes []string          `json:"entity_types"`
	DateRange   *models.DateRange `json:"date_range"`
	Format      string            `json:"format"`
}

type RetentionMetricsRequest struct {
	DateRange    *models.DateRange `json:"date_range"`
	PolicyIDs    []uuid.UUID       `json:"policy_ids"`
	EntityTypes  []string          `json:"entity_types"`
	MetricTypes  []string          `json:"metric_types"`
}

type RetentionMetrics struct {
	TotalDataSize       int64                  `json:"total_data_size"`
	ArchivedDataSize    int64                  `json:"archived_data_size"`
	DeletedDataSize     int64                  `json:"deleted_data_size"`
	StorageSavings      int64                  `json:"storage_savings"`
	CompressionRatio    float64                `json:"compression_ratio"`
	PolicyCompliance    map[string]float64     `json:"policy_compliance"`
	DailyMetrics        []DailyRetentionMetric `json:"daily_metrics"`
}

type DailyRetentionMetric struct {
	Date           time.Time `json:"date"`
	ArchivedSize   int64     `json:"archived_size"`
	DeletedSize    int64     `json:"deleted_size"`
	JobsCompleted  int32     `json:"jobs_completed"`
	JobsFailed     int32     `json:"jobs_failed"`
}

type ValidateComplianceRequest struct {
	PolicyIDs   []uuid.UUID `json:"policy_ids"`
	EntityTypes []string    `json:"entity_types"`
	CheckRules  []string    `json:"check_rules"`
}

type ComplianceValidation struct {
	IsCompliant     bool                    `json:"is_compliant"`
	ValidationScore float64                 `json:"validation_score"`
	CheckResults    []ComplianceCheckResult `json:"check_results"`
	ValidatedAt     time.Time               `json:"validated_at"`
}

type ComplianceCheckResult struct {
	RuleName    string `json:"rule_name"`
	Passed      bool   `json:"passed"`
	Score       float64 `json:"score"`
	Message     string `json:"message"`
	Recommendation string `json:"recommendation"`
}

type ScheduleProcessResult struct {
	ProcessedPolicies int32     `json:"processed_policies"`
	TriggeredJobs     int32     `json:"triggered_jobs"`
	SkippedPolicies   int32     `json:"skipped_policies"`
	Errors            []string  `json:"errors"`
	ProcessedAt       time.Time `json:"processed_at"`
}

type ScheduledExecution struct {
	PolicyID        uuid.UUID `json:"policy_id"`
	PolicyName      string    `json:"policy_name"`
	NextExecution   time.Time `json:"next_execution"`
	LastExecution   *time.Time `json:"last_execution"`
	ExecutionType   string    `json:"execution_type"`
	EstimatedDuration time.Duration `json:"estimated_duration"`
}

type StorageEstimateRequest struct {
	PolicyIDs    []uuid.UUID       `json:"policy_ids"`
	EntityTypes  []string          `json:"entity_types"`
	DateRange    *models.DateRange `json:"date_range"`
	ArchivalMethod models.ArchivalMethod `json:"archival_method"`
	CompressionType models.CompressionType `json:"compression_type"`
}

type StorageEstimate struct {
	CurrentSize       int64   `json:"current_size"`
	EstimatedArchiveSize int64 `json:"estimated_archive_size"`
	EstimatedSavings  int64   `json:"estimated_savings"`
	CompressionRatio  float64 `json:"compression_ratio"`
	SavingsPercentage float64 `json:"savings_percentage"`
}

type StorageStatsRequest struct {
	DatabaseName string            `json:"database_name"`
	SchemaName   string            `json:"schema_name"`
	DateRange    *models.DateRange `json:"date_range"`
	GroupBy      string            `json:"group_by"`
}

type StorageStatistics struct {
	TotalSize        int64                    `json:"total_size"`
	ArchivedSize     int64                    `json:"archived_size"`
	ActiveSize       int64                    `json:"active_size"`
	CompressionRatio float64                  `json:"compression_ratio"`
	StorageBreakdown map[string]StorageBreakdown `json:"storage_breakdown"`
	Trends           []StorageTrend           `json:"trends"`
}

type StorageBreakdown struct {
	Category string `json:"category"`
	Size     int64  `json:"size"`
	Count    int64  `json:"count"`
	Percentage float64 `json:"percentage"`
}

type StorageTrend struct {
	Date      time.Time `json:"date"`
	TotalSize int64     `json:"total_size"`
	Growth    float64   `json:"growth"`
}

type StorageOptimizationRequest struct {
	PolicyIDs      []uuid.UUID `json:"policy_ids"`
	OptimizationType string    `json:"optimization_type"`
	TargetSavings  float64     `json:"target_savings"`
	DryRun         bool        `json:"dry_run"`
}

type StorageOptimizationResult struct {
	OptimizationType    string                       `json:"optimization_type"`
	EstimatedSavings    int64                        `json:"estimated_savings"`
	OptimizationActions []StorageOptimizationAction  `json:"optimization_actions"`
	ExecutionPlan       []OptimizationStep           `json:"execution_plan"`
}

type StorageOptimizationAction struct {
	ActionType  string    `json:"action_type"`
	PolicyID    uuid.UUID `json:"policy_id"`
	Description string    `json:"description"`
	Savings     int64     `json:"savings"`
	Risk        string    `json:"risk"`
}

type OptimizationStep struct {
	StepNumber  int       `json:"step_number"`
	Description string    `json:"description"`
	EstimatedDuration time.Duration `json:"estimated_duration"`
	Dependencies []int    `json:"dependencies"`
}