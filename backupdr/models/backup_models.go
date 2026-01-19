package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// BackupPolicy defines backup schedules and retention
type BackupPolicy struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`

	// Backup scope
	BackupType   BackupType `json:"backup_type" db:"backup_type"`
	TargetType   TargetType `json:"target_type" db:"target_type"` // DATABASE, FILES, APPLICATION
	TargetName   string     `json:"target_name" db:"target_name"`
	IncludeRules []string   `json:"include_rules" db:"include_rules"`
	ExcludeRules []string   `json:"exclude_rules" db:"exclude_rules"`

	// Schedule configuration
	ScheduleType    ScheduleType  `json:"schedule_type" db:"schedule_type"`
	CronExpression  string        `json:"cron_expression" db:"cron_expression"`
	IntervalMinutes int           `json:"interval_minutes" db:"interval_minutes"`
	BackupWindow    BackupWindow  `json:"backup_window" db:"backup_window"`

	// Retention policies
	RetentionPolicy RetentionPolicy `json:"retention_policy" db:"retention_policy"`

	// Storage configuration
	StorageConfig   StorageConfig   `json:"storage_config" db:"storage_config"`
	Compression     CompressionType `json:"compression" db:"compression"`
	Encryption      EncryptionConfig `json:"encryption" db:"encryption"`

	// Performance settings
	MaxParallelJobs int           `json:"max_parallel_jobs" db:"max_parallel_jobs"`
	BandwidthLimit  int64         `json:"bandwidth_limit" db:"bandwidth_limit"` // bytes per second
	TimeoutMinutes  int           `json:"timeout_minutes" db:"timeout_minutes"`

	// Notification settings
	NotifyOnSuccess bool     `json:"notify_on_success" db:"notify_on_success"`
	NotifyOnFailure bool     `json:"notify_on_failure" db:"notify_on_failure"`
	NotifyChannels  []string `json:"notify_channels" db:"notify_channels"`

	// Policy metadata
	IsActive      bool            `json:"is_active" db:"is_active"`
	LastBackup    *time.Time      `json:"last_backup" db:"last_backup"`
	NextBackup    *time.Time      `json:"next_backup" db:"next_backup"`

	CreatedBy     uuid.UUID       `json:"created_by" db:"created_by"`
	UpdatedBy     uuid.UUID       `json:"updated_by" db:"updated_by"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
	Metadata      json.RawMessage `json:"metadata" db:"metadata"`
}

type BackupType string

const (
	BackupTypeFull         BackupType = "FULL"
	BackupTypeIncremental  BackupType = "INCREMENTAL"
	BackupTypeDifferential BackupType = "DIFFERENTIAL"
	BackupTypeSnapshot     BackupType = "SNAPSHOT"
	BackupTypeContinuous   BackupType = "CONTINUOUS"
)

type TargetType string

const (
	TargetTypeDatabase    TargetType = "DATABASE"
	TargetTypeFiles       TargetType = "FILES"
	TargetTypeApplication TargetType = "APPLICATION"
	TargetTypeSystem      TargetType = "SYSTEM"
	TargetTypeVolume      TargetType = "VOLUME"
)

type ScheduleType string

const (
	ScheduleTypeCron     ScheduleType = "CRON"
	ScheduleTypeInterval ScheduleType = "INTERVAL"
	ScheduleTypeManual   ScheduleType = "MANUAL"
	ScheduleTypeEvent    ScheduleType = "EVENT"
)

type BackupWindow struct {
	StartTime string `json:"start_time"` // HH:MM format
	EndTime   string `json:"end_time"`   // HH:MM format
	TimeZone  string `json:"time_zone"`
	WeekDays  []int  `json:"week_days"`  // 0=Sunday, 1=Monday, etc.
}

type RetentionPolicy struct {
	KeepDaily   int `json:"keep_daily"`   // Number of daily backups to keep
	KeepWeekly  int `json:"keep_weekly"`  // Number of weekly backups to keep
	KeepMonthly int `json:"keep_monthly"` // Number of monthly backups to keep
	KeepYearly  int `json:"keep_yearly"`  // Number of yearly backups to keep
	MaxAge      int `json:"max_age"`      // Maximum age in days
}

type StorageConfig struct {
	StorageType     StorageType `json:"storage_type"`
	LocalPath       string      `json:"local_path,omitempty"`
	S3Config        *S3Config   `json:"s3_config,omitempty"`
	AzureConfig     *AzureConfig `json:"azure_config,omitempty"`
	GCPConfig       *GCPConfig  `json:"gcp_config,omitempty"`
	NetworkConfig   *NetworkConfig `json:"network_config,omitempty"`
}

type StorageType string

const (
	StorageTypeLocal   StorageType = "LOCAL"
	StorageTypeS3      StorageType = "S3"
	StorageTypeAzure   StorageType = "AZURE"
	StorageTypeGCP     StorageType = "GCP"
	StorageTypeNetwork StorageType = "NETWORK"
)

type S3Config struct {
	Region          string `json:"region"`
	Bucket          string `json:"bucket"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	Endpoint        string `json:"endpoint,omitempty"`
	UseSSL          bool   `json:"use_ssl"`
	PathStyle       bool   `json:"path_style"`
}

type AzureConfig struct {
	AccountName   string `json:"account_name"`
	AccountKey    string `json:"account_key"`
	ContainerName string `json:"container_name"`
	Endpoint      string `json:"endpoint,omitempty"`
}

type GCPConfig struct {
	ProjectID   string `json:"project_id"`
	BucketName  string `json:"bucket_name"`
	Credentials string `json:"credentials"` // JSON key file content
}

type NetworkConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Protocol string `json:"protocol"` // FTP, SFTP, SCP, NFS, SMB
	Path     string `json:"path"`
}

type CompressionType string

const (
	CompressionNone  CompressionType = "NONE"
	CompressionGZip  CompressionType = "GZIP"
	CompressionZstd  CompressionType = "ZSTD"
	CompressionLZ4   CompressionType = "LZ4"
	CompressionBzip2 CompressionType = "BZIP2"
)

type EncryptionConfig struct {
	Enabled    bool             `json:"enabled"`
	Algorithm  EncryptionAlgo   `json:"algorithm"`
	KeySource  KeySource        `json:"key_source"`
	KeyID      string           `json:"key_id"`
	Passphrase string           `json:"passphrase,omitempty"`
}

type EncryptionAlgo string

const (
	EncryptionAES256     EncryptionAlgo = "AES256"
	EncryptionAES128     EncryptionAlgo = "AES128"
	EncryptionChaCha20   EncryptionAlgo = "CHACHA20"
)

type KeySource string

const (
	KeySourceManual KeySource = "MANUAL"
	KeySourceKMS    KeySource = "KMS"
	KeySourceVault  KeySource = "VAULT"
)

// BackupJob represents an individual backup execution
type BackupJob struct {
	ID       uuid.UUID `json:"id" db:"id"`
	PolicyID uuid.UUID `json:"policy_id" db:"policy_id"`

	// Job details
	JobType     BackupType  `json:"job_type" db:"job_type"`
	Status      JobStatus   `json:"status" db:"status"`
	Priority    JobPriority `json:"priority" db:"priority"`

	// Execution timing
	ScheduledAt time.Time  `json:"scheduled_at" db:"scheduled_at"`
	StartedAt   *time.Time `json:"started_at" db:"started_at"`
	CompletedAt *time.Time `json:"completed_at" db:"completed_at"`
	Duration    *time.Duration `json:"duration" db:"duration"`

	// Progress tracking
	TotalSize        int64   `json:"total_size" db:"total_size"`
	ProcessedSize    int64   `json:"processed_size" db:"processed_size"`
	CompressedSize   int64   `json:"compressed_size" db:"compressed_size"`
	ProgressPercent  float64 `json:"progress_percent" db:"progress_percent"`
	EstimatedFinish  *time.Time `json:"estimated_finish" db:"estimated_finish"`

	// File tracking
	TotalFiles       int64   `json:"total_files" db:"total_files"`
	ProcessedFiles   int64   `json:"processed_files" db:"processed_files"`
	SkippedFiles     int64   `json:"skipped_files" db:"skipped_files"`
	ErroredFiles     int64   `json:"errored_files" db:"errored_files"`

	// Storage information
	BackupPath       string  `json:"backup_path" db:"backup_path"`
	BackupSize       int64   `json:"backup_size" db:"backup_size"`
	CompressionRatio float64 `json:"compression_ratio" db:"compression_ratio"`
	Checksum         string  `json:"checksum" db:"checksum"`

	// Performance metrics
	ThroughputMBps   float64 `json:"throughput_mbps" db:"throughput_mbps"`
	AverageSpeed     float64 `json:"average_speed" db:"average_speed"`
	NetworkUsage     int64   `json:"network_usage" db:"network_usage"`
	CPUUsage         float64 `json:"cpu_usage" db:"cpu_usage"`

	// Error handling
	ErrorMessage     string  `json:"error_message" db:"error_message"`
	RetryCount       int     `json:"retry_count" db:"retry_count"`
	MaxRetries       int     `json:"max_retries" db:"max_retries"`

	// Dependencies
	ParentJobID      *uuid.UUID `json:"parent_job_id" db:"parent_job_id"`
	BaselineJobID    *uuid.UUID `json:"baseline_job_id" db:"baseline_job_id"` // For incremental backups

	ExecutedBy       uuid.UUID  `json:"executed_by" db:"executed_by"`
	JobMetadata      json.RawMessage `json:"job_metadata" db:"job_metadata"`
}

type JobStatus string

const (
	JobStatusScheduled  JobStatus = "SCHEDULED"
	JobStatusQueued     JobStatus = "QUEUED"
	JobStatusRunning    JobStatus = "RUNNING"
	JobStatusCompleted  JobStatus = "COMPLETED"
	JobStatusFailed     JobStatus = "FAILED"
	JobStatusCancelled  JobStatus = "CANCELLED"
	JobStatusPaused     JobStatus = "PAUSED"
	JobStatusRetrying   JobStatus = "RETRYING"
)

type JobPriority string

const (
	PriorityLow      JobPriority = "LOW"
	PriorityNormal   JobPriority = "NORMAL"
	PriorityHigh     JobPriority = "HIGH"
	PriorityCritical JobPriority = "CRITICAL"
)

// DisasterRecoveryPlan defines comprehensive DR procedures
type DisasterRecoveryPlan struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`

	// Plan classification
	PlanType    DRPlanType `json:"plan_type" db:"plan_type"`
	Severity    DRSeverity `json:"severity" db:"severity"`
	Scope       DRScope    `json:"scope" db:"scope"`

	// Recovery objectives
	RTO         time.Duration `json:"rto" db:"rto"`         // Recovery Time Objective
	RPO         time.Duration `json:"rpo" db:"rpo"`         // Recovery Point Objective
	MTTR        time.Duration `json:"mttr" db:"mttr"`       // Mean Time To Recovery
	Availability float64       `json:"availability" db:"availability"` // Target availability %

	// Recovery procedures
	PreRecoverySteps  json.RawMessage `json:"pre_recovery_steps" db:"pre_recovery_steps"`
	RecoverySteps     json.RawMessage `json:"recovery_steps" db:"recovery_steps"`
	PostRecoverySteps json.RawMessage `json:"post_recovery_steps" db:"post_recovery_steps"`
	RollbackSteps     json.RawMessage `json:"rollback_steps" db:"rollback_steps"`

	// Resource requirements
	ResourceRequirements json.RawMessage `json:"resource_requirements" db:"resource_requirements"`
	Dependencies         []uuid.UUID     `json:"dependencies" db:"dependencies"`
	Prerequisites        json.RawMessage `json:"prerequisites" db:"prerequisites"`

	// Testing and validation
	LastTested          *time.Time      `json:"last_tested" db:"last_tested"`
	TestSchedule        string          `json:"test_schedule" db:"test_schedule"`
	ValidationCriteria  json.RawMessage `json:"validation_criteria" db:"validation_criteria"`

	// Communication plan
	NotificationList    []string        `json:"notification_list" db:"notification_list"`
	EscalationMatrix    json.RawMessage `json:"escalation_matrix" db:"escalation_matrix"`
	CommunicationPlan   json.RawMessage `json:"communication_plan" db:"communication_plan"`

	// Plan metadata
	IsActive            bool            `json:"is_active" db:"is_active"`
	Version             int             `json:"version" db:"version"`
	ApprovedBy          *uuid.UUID      `json:"approved_by" db:"approved_by"`
	ApprovedAt          *time.Time      `json:"approved_at" db:"approved_at"`

	CreatedBy           uuid.UUID       `json:"created_by" db:"created_by"`
	UpdatedBy           uuid.UUID       `json:"updated_by" db:"updated_by"`
	CreatedAt           time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at" db:"updated_at"`
	Metadata            json.RawMessage `json:"metadata" db:"metadata"`
}

type DRPlanType string

const (
	DRPlanTypeDataRecovery     DRPlanType = "DATA_RECOVERY"
	DRPlanTypeSystemFailover   DRPlanType = "SYSTEM_FAILOVER"
	DRPlanTypeSiteFailover     DRPlanType = "SITE_FAILOVER"
	DRPlanTypeBusinessContinuity DRPlanType = "BUSINESS_CONTINUITY"
	DRPlanTypeCyberIncident    DRPlanType = "CYBER_INCIDENT"
)

type DRSeverity string

const (
	DRSeverityMinor    DRSeverity = "MINOR"
	DRSeverityMajor    DRSeverity = "MAJOR"
	DRSeverityCritical DRSeverity = "CRITICAL"
	DRSeverityCatastrophic DRSeverity = "CATASTROPHIC"
)

type DRScope string

const (
	DRScopeApplication DRScope = "APPLICATION"
	DRScopeDatabase    DRScope = "DATABASE"
	DRScopeSystem      DRScope = "SYSTEM"
	DRScopeDatacenter  DRScope = "DATACENTER"
	DRScopeEnterprise  DRScope = "ENTERPRISE"
)

// RecoveryExecution tracks DR plan executions
type RecoveryExecution struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	PlanID          uuid.UUID       `json:"plan_id" db:"plan_id"`
	ExecutionType   ExecutionType   `json:"execution_type" db:"execution_type"`
	TriggerReason   string          `json:"trigger_reason" db:"trigger_reason"`

	// Execution details
	Status          ExecutionStatus `json:"status" db:"status"`
	StartedAt       time.Time       `json:"started_at" db:"started_at"`
	CompletedAt     *time.Time      `json:"completed_at" db:"completed_at"`
	EstimatedRTO    time.Duration   `json:"estimated_rto" db:"estimated_rto"`
	ActualRTO       *time.Duration  `json:"actual_rto" db:"actual_rto"`

	// Progress tracking
	TotalSteps      int             `json:"total_steps" db:"total_steps"`
	CompletedSteps  int             `json:"completed_steps" db:"completed_steps"`
	FailedSteps     int             `json:"failed_steps" db:"failed_steps"`
	CurrentStep     string          `json:"current_step" db:"current_step"`

	// Results and metrics
	SuccessRate     float64         `json:"success_rate" db:"success_rate"`
	DataRecovered   int64           `json:"data_recovered" db:"data_recovered"`
	SystemsRecovered int            `json:"systems_recovered" db:"systems_recovered"`

	// Communication tracking
	NotificationsSent json.RawMessage `json:"notifications_sent" db:"notifications_sent"`
	StakeholdersNotified []string     `json:"stakeholders_notified" db:"stakeholders_notified"`

	// Issues and resolutions
	IssuesEncountered json.RawMessage `json:"issues_encountered" db:"issues_encountered"`
	ResolutionActions json.RawMessage `json:"resolution_actions" db:"resolution_actions"`

	ExecutedBy      uuid.UUID       `json:"executed_by" db:"executed_by"`
	ExecutionLog    json.RawMessage `json:"execution_log" db:"execution_log"`
	Metadata        json.RawMessage `json:"metadata" db:"metadata"`
}

type ExecutionType string

const (
	ExecutionTypeTest      ExecutionType = "TEST"
	ExecutionTypeActual    ExecutionType = "ACTUAL"
	ExecutionTypeDrill     ExecutionType = "DRILL"
	ExecutionTypePartial   ExecutionType = "PARTIAL"
)

type ExecutionStatus string

const (
	ExecutionStatusInitiated ExecutionStatus = "INITIATED"
	ExecutionStatusInProgress ExecutionStatus = "IN_PROGRESS"
	ExecutionStatusCompleted ExecutionStatus = "COMPLETED"
	ExecutionStatusFailed    ExecutionStatus = "FAILED"
	ExecutionStatusAborted   ExecutionStatus = "ABORTED"
	ExecutionStatusRolledBack ExecutionStatus = "ROLLED_BACK"
)

// SystemHealth tracks system health for proactive DR
type SystemHealth struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	SystemName      string          `json:"system_name" db:"system_name"`
	SystemType      string          `json:"system_type" db:"system_type"`
	HealthStatus    HealthStatus    `json:"health_status" db:"health_status"`

	// Health metrics
	CPUUsage        float64         `json:"cpu_usage" db:"cpu_usage"`
	MemoryUsage     float64         `json:"memory_usage" db:"memory_usage"`
	DiskUsage       float64         `json:"disk_usage" db:"disk_usage"`
	NetworkLatency  float64         `json:"network_latency" db:"network_latency"`
	ResponseTime    float64         `json:"response_time" db:"response_time"`

	// Availability metrics
	Uptime          time.Duration   `json:"uptime" db:"uptime"`
	LastDowntime    *time.Time      `json:"last_downtime" db:"last_downtime"`
	AvailabilityPct float64         `json:"availability_pct" db:"availability_pct"`

	// Backup status
	LastBackupTime  *time.Time      `json:"last_backup_time" db:"last_backup_time"`
	BackupStatus    string          `json:"backup_status" db:"backup_status"`
	BackupSize      int64           `json:"backup_size" db:"backup_size"`

	// Alerts and thresholds
	ActiveAlerts    json.RawMessage `json:"active_alerts" db:"active_alerts"`
	ThresholdBreaches json.RawMessage `json:"threshold_breaches" db:"threshold_breaches"`

	CheckedAt       time.Time       `json:"checked_at" db:"checked_at"`
	Metadata        json.RawMessage `json:"metadata" db:"metadata"`
}

type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "HEALTHY"
	HealthStatusWarning   HealthStatus = "WARNING"
	HealthStatusCritical  HealthStatus = "CRITICAL"
	HealthStatusUnknown   HealthStatus = "UNKNOWN"
	HealthStatusDown      HealthStatus = "DOWN"
)

// RestoreRequest represents data restoration requests
type RestoreRequest struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	BackupJobID     uuid.UUID       `json:"backup_job_id" db:"backup_job_id"`
	RequestedBy     uuid.UUID       `json:"requested_by" db:"requested_by"`

	// Restore configuration
	RestoreType     RestoreType     `json:"restore_type" db:"restore_type"`
	RestoreScope    RestoreScope    `json:"restore_scope" db:"restore_scope"`
	TargetLocation  string          `json:"target_location" db:"target_location"`
	OverwritePolicy OverwritePolicy `json:"overwrite_policy" db:"overwrite_policy"`

	// Selection criteria
	IncludeFilters  []string        `json:"include_filters" db:"include_filters"`
	ExcludeFilters  []string        `json:"exclude_filters" db:"exclude_filters"`
	PointInTime     *time.Time      `json:"point_in_time" db:"point_in_time"`

	// Status and progress
	Status          RestoreStatus   `json:"status" db:"status"`
	StartedAt       *time.Time      `json:"started_at" db:"started_at"`
	CompletedAt     *time.Time      `json:"completed_at" db:"completed_at"`
	ProgressPercent float64         `json:"progress_percent" db:"progress_percent"`

	// Results
	RestoredFiles   int64           `json:"restored_files" db:"restored_files"`
	RestoredSize    int64           `json:"restored_size" db:"restored_size"`
	ErrorCount      int64           `json:"error_count" db:"error_count"`
	ErrorMessage    string          `json:"error_message" db:"error_message"`

	// Approval workflow
	RequiresApproval bool           `json:"requires_approval" db:"requires_approval"`
	ApprovalStatus   string         `json:"approval_status" db:"approval_status"`
	ApprovedBy       *uuid.UUID     `json:"approved_by" db:"approved_by"`
	ApprovedAt       *time.Time     `json:"approved_at" db:"approved_at"`

	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	Metadata        json.RawMessage `json:"metadata" db:"metadata"`
}

type RestoreType string

const (
	RestoreTypeFull        RestoreType = "FULL"
	RestoreTypePartial     RestoreType = "PARTIAL"
	RestoreTypePointInTime RestoreType = "POINT_IN_TIME"
	RestoreTypeInstant     RestoreType = "INSTANT"
)

type RestoreScope string

const (
	RestoreScopeDatabase RestoreScope = "DATABASE"
	RestoreScopeTable    RestoreScope = "TABLE"
	RestoreScopeFiles    RestoreScope = "FILES"
	RestoreScopeSystem   RestoreScope = "SYSTEM"
)

type OverwritePolicy string

const (
	OverwritePolicySkip      OverwritePolicy = "SKIP"
	OverwritePolicyOverwrite OverwritePolicy = "OVERWRITE"
	OverwritePolicyRename    OverwritePolicy = "RENAME"
	OverwritePolicyPrompt    OverwritePolicy = "PROMPT"
)

type RestoreStatus string

const (
	RestoreStatusRequested  RestoreStatus = "REQUESTED"
	RestoreStatusApproved   RestoreStatus = "APPROVED"
	RestoreStatusQueued     RestoreStatus = "QUEUED"
	RestoreStatusRunning    RestoreStatus = "RUNNING"
	RestoreStatusCompleted  RestoreStatus = "COMPLETED"
	RestoreStatusFailed     RestoreStatus = "FAILED"
	RestoreStatusCancelled  RestoreStatus = "CANCELLED"
)

// Filter structs for repository queries
type BackupPolicyFilters struct {
	TargetType   *TargetType    `json:"target_type,omitempty"`
	TargetName   string         `json:"target_name,omitempty"`
	BackupType   *BackupType    `json:"backup_type,omitempty"`
	IsActive     *bool          `json:"is_active,omitempty"`
	ScheduleType *ScheduleType  `json:"schedule_type,omitempty"`
}

type BackupJobFilters struct {
	PolicyID      *uuid.UUID     `json:"policy_id,omitempty"`
	JobType       *BackupType    `json:"job_type,omitempty"`
	Status        *JobStatus     `json:"status,omitempty"`
	Priority      *JobPriority   `json:"priority,omitempty"`
	ScheduledAfter *time.Time    `json:"scheduled_after,omitempty"`
	ScheduledBefore *time.Time   `json:"scheduled_before,omitempty"`
	StartedAfter   *time.Time    `json:"started_after,omitempty"`
	StartedBefore  *time.Time    `json:"started_before,omitempty"`
}

type DRPlanFilters struct {
	PlanType  *DRPlanType `json:"plan_type,omitempty"`
	Severity  *DRSeverity `json:"severity,omitempty"`
	Scope     *DRScope    `json:"scope,omitempty"`
	IsActive  *bool       `json:"is_active,omitempty"`
	Version   *int        `json:"version,omitempty"`
}

type RecoveryExecutionFilters struct {
	PlanID         *uuid.UUID       `json:"plan_id,omitempty"`
	ExecutionType  *ExecutionType   `json:"execution_type,omitempty"`
	Status         *ExecutionStatus `json:"status,omitempty"`
	StartedAfter   *time.Time       `json:"started_after,omitempty"`
	StartedBefore  *time.Time       `json:"started_before,omitempty"`
	ExecutedBy     *uuid.UUID       `json:"executed_by,omitempty"`
}

type SystemHealthFilters struct {
	SystemName     string         `json:"system_name,omitempty"`
	SystemType     string         `json:"system_type,omitempty"`
	HealthStatus   *HealthStatus  `json:"health_status,omitempty"`
	CheckedAfter   *time.Time     `json:"checked_after,omitempty"`
	CheckedBefore  *time.Time     `json:"checked_before,omitempty"`
}

type RestoreRequestFilters struct {
	BackupJobID    *uuid.UUID     `json:"backup_job_id,omitempty"`
	RequestedBy    *uuid.UUID     `json:"requested_by,omitempty"`
	Status         *RestoreStatus `json:"status,omitempty"`
	RestoreType    *RestoreType   `json:"restore_type,omitempty"`
	RestoreScope   *RestoreScope  `json:"restore_scope,omitempty"`
	RequiresApproval *bool        `json:"requires_approval,omitempty"`
	CreatedAfter   *time.Time     `json:"created_after,omitempty"`
	CreatedBefore  *time.Time     `json:"created_before,omitempty"`
}

// BackupInstance represents a specific backup execution instance
type BackupInstance struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	JobID           uuid.UUID       `json:"job_id" db:"job_id"`
	BackupPath      string          `json:"backup_path" db:"backup_path"`
	BackupSize      int64           `json:"backup_size" db:"backup_size"`
	ChecksumType    string          `json:"checksum_type" db:"checksum_type"`
	ChecksumValue   string          `json:"checksum_value" db:"checksum_value"`
	Status          InstanceStatus  `json:"status" db:"status"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	ExpiresAt       *time.Time      `json:"expires_at" db:"expires_at"`
	Metadata        json.RawMessage `json:"metadata" db:"metadata"`
}

type InstanceStatus string

const (
	InstanceStatusValid   InstanceStatus = "VALID"
	InstanceStatusCorrupt InstanceStatus = "CORRUPT"
	InstanceStatusExpired InstanceStatus = "EXPIRED"
	InstanceStatusDeleted InstanceStatus = "DELETED"
)

type BackupInstanceFilters struct {
	JobID         *uuid.UUID      `json:"job_id,omitempty"`
	Status        *InstanceStatus `json:"status,omitempty"`
	CreatedAfter  *time.Time      `json:"created_after,omitempty"`
	CreatedBefore *time.Time      `json:"created_before,omitempty"`
	ExpiresAfter  *time.Time      `json:"expires_after,omitempty"`
	ExpiresBefore *time.Time      `json:"expires_before,omitempty"`
}

// DRPlan alias for DisasterRecoveryPlan for interface consistency
type DRPlan = DisasterRecoveryPlan

// DRTest represents disaster recovery test execution
type DRTest struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	PlanID          uuid.UUID       `json:"plan_id" db:"plan_id"`
	TestName        string          `json:"test_name" db:"test_name"`
	TestType        DRTestType      `json:"test_type" db:"test_type"`
	Status          DRTestStatus    `json:"status" db:"status"`
	ScheduledAt     time.Time       `json:"scheduled_at" db:"scheduled_at"`
	StartedAt       *time.Time      `json:"started_at" db:"started_at"`
	CompletedAt     *time.Time      `json:"completed_at" db:"completed_at"`
	TestResults     json.RawMessage `json:"test_results" db:"test_results"`
	IssuesFound     json.RawMessage `json:"issues_found" db:"issues_found"`
	Recommendations json.RawMessage `json:"recommendations" db:"recommendations"`
	ExecutedBy      uuid.UUID       `json:"executed_by" db:"executed_by"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	Metadata        json.RawMessage `json:"metadata" db:"metadata"`
}

type DRTestType string

const (
	DRTestTypeTabletop      DRTestType = "TABLETOP"
	DRTestTypeWalkthrough   DRTestType = "WALKTHROUGH"
	DRTestTypeSimulation    DRTestType = "SIMULATION"
	DRTestTypeParallel      DRTestType = "PARALLEL"
	DRTestTypeFullInterrupt DRTestType = "FULL_INTERRUPT"
)

type DRTestStatus string

const (
	DRTestStatusScheduled  DRTestStatus = "SCHEDULED"
	DRTestStatusInProgress DRTestStatus = "IN_PROGRESS"
	DRTestStatusCompleted  DRTestStatus = "COMPLETED"
	DRTestStatusFailed     DRTestStatus = "FAILED"
	DRTestStatusCancelled  DRTestStatus = "CANCELLED"
)

type DRTestFilters struct {
	PlanID         *uuid.UUID    `json:"plan_id,omitempty"`
	TestType       *DRTestType   `json:"test_type,omitempty"`
	Status         *DRTestStatus `json:"status,omitempty"`
	ScheduledAfter *time.Time    `json:"scheduled_after,omitempty"`
	ScheduledBefore *time.Time   `json:"scheduled_before,omitempty"`
	ExecutedBy     *uuid.UUID    `json:"executed_by,omitempty"`
}

// Analytics and reporting models
type BackupStatistics struct {
	TotalPolicies        int64             `json:"total_policies"`
	ActivePolicies       int64             `json:"active_policies"`
	TotalJobs            int64             `json:"total_jobs"`
	SuccessfulJobs       int64             `json:"successful_jobs"`
	FailedJobs           int64             `json:"failed_jobs"`
	TotalBackupSize      int64             `json:"total_backup_size"`
	AverageBackupTime    time.Duration     `json:"average_backup_time"`
	CompressionRatio     float64           `json:"compression_ratio"`
	StorageUtilization   float64           `json:"storage_utilization"`
	LastCalculatedAt     time.Time         `json:"last_calculated_at"`
}

type BackupStatsFilters struct {
	StartDate   time.Time    `json:"start_date"`
	EndDate     time.Time    `json:"end_date"`
	TargetType  *TargetType  `json:"target_type,omitempty"`
	BackupType  *BackupType  `json:"backup_type,omitempty"`
	PolicyID    *uuid.UUID   `json:"policy_id,omitempty"`
}

type StorageUtilization struct {
	TargetType      TargetType `json:"target_type"`
	TargetName      string     `json:"target_name"`
	TotalSize       int64      `json:"total_size"`
	UsedSize        int64      `json:"used_size"`
	AvailableSize   int64      `json:"available_size"`
	UtilizationPct  float64    `json:"utilization_pct"`
	BackupCount     int64      `json:"backup_count"`
	OldestBackup    *time.Time `json:"oldest_backup,omitempty"`
	NewestBackup    *time.Time `json:"newest_backup,omitempty"`
	LastUpdated     time.Time  `json:"last_updated"`
}

type ComplianceReport struct {
	PolicyID            uuid.UUID     `json:"policy_id"`
	PolicyName          string        `json:"policy_name"`
	ComplianceStatus    ComplianceStatus `json:"compliance_status"`
	LastSuccessfulBackup *time.Time   `json:"last_successful_backup,omitempty"`
	NextScheduledBackup  *time.Time   `json:"next_scheduled_backup,omitempty"`
	BackupCount         int64         `json:"backup_count"`
	FailureCount        int64         `json:"failure_count"`
	SuccessRate         float64       `json:"success_rate"`
	RTOCompliance       bool          `json:"rto_compliance"`
	RPOCompliance       bool          `json:"rpo_compliance"`
	Issues              []string      `json:"issues"`
	Recommendations     []string      `json:"recommendations"`
	LastChecked         time.Time     `json:"last_checked"`
}

type ComplianceStatus string

const (
	ComplianceStatusCompliant    ComplianceStatus = "COMPLIANT"
	ComplianceStatusNonCompliant ComplianceStatus = "NON_COMPLIANT"
	ComplianceStatusWarning      ComplianceStatus = "WARNING"
	ComplianceStatusUnknown      ComplianceStatus = "UNKNOWN"
)

type ComplianceFilters struct {
	PolicyID         *uuid.UUID        `json:"policy_id,omitempty"`
	ComplianceStatus *ComplianceStatus `json:"compliance_status,omitempty"`
	TargetType       *TargetType       `json:"target_type,omitempty"`
	CheckedAfter     *time.Time        `json:"checked_after,omitempty"`
	CheckedBefore    *time.Time        `json:"checked_before,omitempty"`
}