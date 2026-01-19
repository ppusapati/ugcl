package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// RetentionPolicy defines data retention and archival rules
type RetentionPolicy struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`

	// Scope definition
	EntityType  string    `json:"entity_type" db:"entity_type"`   // table/collection name
	SchemaName  string    `json:"schema_name" db:"schema_name"`   // database schema
	DatabaseName string   `json:"database_name" db:"database_name"` // database name

	// Retention configuration
	RetentionPeriod    time.Duration      `json:"retention_period" db:"retention_period"`
	ArchivalPeriod     time.Duration      `json:"archival_period" db:"archival_period"`     // Before deletion
	GracePeriod        time.Duration      `json:"grace_period" db:"grace_period"`           // Additional time before deletion

	// Policy behavior
	PolicyType         PolicyType         `json:"policy_type" db:"policy_type"`
	ArchivalMethod     ArchivalMethod     `json:"archival_method" db:"archival_method"`
	CompressionType    CompressionType    `json:"compression_type" db:"compression_type"`
	EncryptArchive     bool               `json:"encrypt_archive" db:"encrypt_archive"`

	// Selection criteria
	SelectionCriteria  json.RawMessage    `json:"selection_criteria" db:"selection_criteria"`  // SQL/NoSQL conditions
	ExclusionRules     json.RawMessage    `json:"exclusion_rules" db:"exclusion_rules"`        // What to exclude

	// Compliance and legal hold
	LegalHoldEnabled   bool               `json:"legal_hold_enabled" db:"legal_hold_enabled"`
	ComplianceLevel    ComplianceLevel    `json:"compliance_level" db:"compliance_level"`
	RegulatoryBasis    []string           `json:"regulatory_basis" db:"regulatory_basis"`      // GDPR, HIPAA, SOX, etc.

	// Execution schedule
	ScheduleEnabled    bool               `json:"schedule_enabled" db:"schedule_enabled"`
	ScheduleCron       string             `json:"schedule_cron" db:"schedule_cron"`
	BatchSize          int                `json:"batch_size" db:"batch_size"`
	ParallelJobs       int                `json:"parallel_jobs" db:"parallel_jobs"`

	// Notification settings
	NotifyOnStart      bool               `json:"notify_on_start" db:"notify_on_start"`
	NotifyOnComplete   bool               `json:"notify_on_complete" db:"notify_on_complete"`
	NotifyOnError      bool               `json:"notify_on_error" db:"notify_on_error"`
	NotificationChannels []string         `json:"notification_channels" db:"notification_channels"`

	// Policy metadata
	IsActive           bool               `json:"is_active" db:"is_active"`
	LastExecuted       *time.Time         `json:"last_executed" db:"last_executed"`
	NextExecution      *time.Time         `json:"next_execution" db:"next_execution"`

	CreatedBy          uuid.UUID          `json:"created_by" db:"created_by"`
	UpdatedBy          uuid.UUID          `json:"updated_by" db:"updated_by"`
	CreatedAt          time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at" db:"updated_at"`
	Metadata           json.RawMessage    `json:"metadata" db:"metadata"`
}

type PolicyType string

const (
	PolicyTypeTimeBasedRetention PolicyType = "TIME_BASED_RETENTION"
	PolicyTypeSizeBasedRetention PolicyType = "SIZE_BASED_RETENTION"
	PolicyTypeEventBasedRetention PolicyType = "EVENT_BASED_RETENTION"
	PolicyTypeComplianceRetention PolicyType = "COMPLIANCE_RETENTION"
	PolicyTypeCustomRetention     PolicyType = "CUSTOM_RETENTION"
)

type ArchivalMethod string

const (
	ArchivalMethodColdStorage ArchivalMethod = "COLD_STORAGE"    // Move to cold storage
	ArchivalMethodCompression ArchivalMethod = "COMPRESSION"     // Compress in place
	ArchivalMethodPartition   ArchivalMethod = "PARTITION"       // Move to archive partition
	ArchivalMethodExport      ArchivalMethod = "EXPORT"          // Export to external system
	ArchivalMethodDelete      ArchivalMethod = "DELETE"          // Direct deletion
)

type CompressionType string

const (
	CompressionTypeNone    CompressionType = "NONE"
	CompressionTypeGZip    CompressionType = "GZIP"
	CompressionTypeZstd    CompressionType = "ZSTD"
	CompressionTypeLZ4     CompressionType = "LZ4"
	CompressionTypeBrotli  CompressionType = "BROTLI"
)

type ComplianceLevel string

const (
	ComplianceLevelNone     ComplianceLevel = "NONE"
	ComplianceLevelStandard ComplianceLevel = "STANDARD"
	ComplianceLevelHigh     ComplianceLevel = "HIGH"
	ComplianceLevelCritical ComplianceLevel = "CRITICAL"
)

// ArchivalJob represents a specific archival execution
type ArchivalJob struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	PolicyID        uuid.UUID       `json:"policy_id" db:"policy_id"`
	JobType         JobType         `json:"job_type" db:"job_type"`
	Status          JobStatus       `json:"status" db:"status"`

	// Job configuration
	TargetTable     string          `json:"target_table" db:"target_table"`
	DateRange       DateRange       `json:"date_range" db:"date_range"`
	BatchSize       int             `json:"batch_size" db:"batch_size"`

	// Execution details
	StartedAt       time.Time       `json:"started_at" db:"started_at"`
	CompletedAt     *time.Time      `json:"completed_at" db:"completed_at"`
	EstimatedEnd    *time.Time      `json:"estimated_end" db:"estimated_end"`

	// Progress tracking
	TotalRecords    int64           `json:"total_records" db:"total_records"`
	ProcessedRecords int64          `json:"processed_records" db:"processed_records"`
	ArchivedRecords int64           `json:"archived_records" db:"archived_records"`
	DeletedRecords  int64           `json:"deleted_records" db:"deleted_records"`
	ErrorCount      int64           `json:"error_count" db:"error_count"`

	// Storage information
	OriginalSize    int64           `json:"original_size" db:"original_size"`      // bytes
	CompressedSize  int64           `json:"compressed_size" db:"compressed_size"`  // bytes
	CompressionRatio float64        `json:"compression_ratio" db:"compression_ratio"`

	// Archive location
	ArchiveLocation string          `json:"archive_location" db:"archive_location"`
	ArchiveFormat   string          `json:"archive_format" db:"archive_format"`
	EncryptionKey   string          `json:"encryption_key" db:"encryption_key"`   // encrypted

	// Error handling
	ErrorMessage    string          `json:"error_message" db:"error_message"`
	RetryCount      int             `json:"retry_count" db:"retry_count"`
	MaxRetries      int             `json:"max_retries" db:"max_retries"`

	// Execution metadata
	ExecutedBy      uuid.UUID       `json:"executed_by" db:"executed_by"`
	JobMetadata     json.RawMessage `json:"job_metadata" db:"job_metadata"`
}

type JobType string

const (
	JobTypeArchive JobType = "ARCHIVE"
	JobTypeRestore JobType = "RESTORE"
	JobTypeDelete  JobType = "DELETE"
	JobTypeAudit   JobType = "AUDIT"
	JobTypePurge   JobType = "PURGE"
)

type JobStatus string

const (
	JobStatusPending    JobStatus = "PENDING"
	JobStatusRunning    JobStatus = "RUNNING"
	JobStatusCompleted  JobStatus = "COMPLETED"
	JobStatusFailed     JobStatus = "FAILED"
	JobStatusCancelled  JobStatus = "CANCELLED"
	JobStatusPartial    JobStatus = "PARTIAL"
)

type DateRange struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// LegalHold represents legal hold instructions
type LegalHold struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	Name            string          `json:"name" db:"name"`
	Description     string          `json:"description" db:"description"`

	// Hold scope
	EntityTypes     []string        `json:"entity_types" db:"entity_types"`
	EntityIDs       []string        `json:"entity_ids" db:"entity_ids"`
	DateRange       *DateRange      `json:"date_range" db:"date_range"`

	// Legal details
	CaseNumber      string          `json:"case_number" db:"case_number"`
	LegalBasis      string          `json:"legal_basis" db:"legal_basis"`
	IssuingAuthority string         `json:"issuing_authority" db:"issuing_authority"`
	ContactInfo     string          `json:"contact_info" db:"contact_info"`

	// Hold status
	IsActive        bool            `json:"is_active" db:"is_active"`
	StartDate       time.Time       `json:"start_date" db:"start_date"`
	EndDate         *time.Time      `json:"end_date" db:"end_date"`

	// Affected policies
	AffectedPolicies []uuid.UUID    `json:"affected_policies" db:"affected_policies"`

	CreatedBy       uuid.UUID       `json:"created_by" db:"created_by"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
	Metadata        json.RawMessage `json:"metadata" db:"metadata"`
}

// ArchivedData represents archived data records
type ArchivedData struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	PolicyID        uuid.UUID       `json:"policy_id" db:"policy_id"`
	JobID           uuid.UUID       `json:"job_id" db:"job_id"`

	// Original data reference
	SourceTable     string          `json:"source_table" db:"source_table"`
	SourceSchema    string          `json:"source_schema" db:"source_schema"`
	SourceDatabase  string          `json:"source_database" db:"source_database"`
	OriginalID      string          `json:"original_id" db:"original_id"`

	// Archive information
	ArchiveFormat   string          `json:"archive_format" db:"archive_format"`
	StorageLocation string          `json:"storage_location" db:"storage_location"`
	StorageTier     StorageTier     `json:"storage_tier" db:"storage_tier"`

	// Data characteristics
	DataSize        int64           `json:"data_size" db:"data_size"`
	CompressedSize  int64           `json:"compressed_size" db:"compressed_size"`
	IsEncrypted     bool            `json:"is_encrypted" db:"is_encrypted"`
	EncryptionAlgo  string          `json:"encryption_algo" db:"encryption_algo"`

	// Checksums for integrity
	OriginalChecksum string         `json:"original_checksum" db:"original_checksum"`
	ArchiveChecksum  string         `json:"archive_checksum" db:"archive_checksum"`

	// Timeline
	OriginalDate    time.Time       `json:"original_date" db:"original_date"`
	ArchivedAt      time.Time       `json:"archived_at" db:"archived_at"`
	ExpiresAt       *time.Time      `json:"expires_at" db:"expires_at"`

	// Legal hold status
	LegalHoldIDs    []uuid.UUID     `json:"legal_hold_ids" db:"legal_hold_ids"`
	IsOnLegalHold   bool            `json:"is_on_legal_hold" db:"is_on_legal_hold"`

	// Access tracking
	LastAccessedAt  *time.Time      `json:"last_accessed_at" db:"last_accessed_at"`
	AccessCount     int             `json:"access_count" db:"access_count"`

	Metadata        json.RawMessage `json:"metadata" db:"metadata"`
}

type StorageTier string

const (
	StorageTierHot     StorageTier = "HOT"         // Immediate access
	StorageTierWarm    StorageTier = "WARM"        // Infrequent access
	StorageTierCold    StorageTier = "COLD"        // Archive access
	StorageTierFrozen  StorageTier = "FROZEN"      // Deep archive
)

// DataInventory tracks what data exists and its characteristics
type DataInventory struct {
	ID              uuid.UUID       `json:"id" db:"id"`

	// Data source
	DatabaseName    string          `json:"database_name" db:"database_name"`
	SchemaName      string          `json:"schema_name" db:"schema_name"`
	TableName       string          `json:"table_name" db:"table_name"`

	// Data characteristics
	RecordCount     int64           `json:"record_count" db:"record_count"`
	DataSize        int64           `json:"data_size" db:"data_size"`
	OldestRecord    *time.Time      `json:"oldest_record" db:"oldest_record"`
	NewestRecord    *time.Time      `json:"newest_record" db:"newest_record"`

	// Classification
	DataClassification DataClassification `json:"data_classification" db:"data_classification"`
	ContainsPII     bool            `json:"contains_pii" db:"contains_pii"`
	ContainsPHI     bool            `json:"contains_phi" db:"contains_phi"`
	ContainsPCI     bool            `json:"contains_pci" db:"contains_pci"`

	// Retention requirements
	MinRetentionPeriod time.Duration `json:"min_retention_period" db:"min_retention_period"`
	MaxRetentionPeriod time.Duration `json:"max_retention_period" db:"max_retention_period"`
	RegulatoryRequirements []string  `json:"regulatory_requirements" db:"regulatory_requirements"`

	// Discovery metadata
	LastScanned     time.Time       `json:"last_scanned" db:"last_scanned"`
	ScanMethod      string          `json:"scan_method" db:"scan_method"`
	Confidence      float64         `json:"confidence" db:"confidence"`

	Metadata        json.RawMessage `json:"metadata" db:"metadata"`
}

type DataClassification string

const (
	DataClassificationPublic       DataClassification = "PUBLIC"
	DataClassificationInternal     DataClassification = "INTERNAL"
	DataClassificationConfidential DataClassification = "CONFIDENTIAL"
	DataClassificationRestricted   DataClassification = "RESTRICTED"
)

// ComplianceAudit tracks compliance with retention policies
type ComplianceAudit struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	AuditType       AuditType       `json:"audit_type" db:"audit_type"`

	// Audit scope
	PolicyIDs       []uuid.UUID     `json:"policy_ids" db:"policy_ids"`
	EntityTypes     []string        `json:"entity_types" db:"entity_types"`
	DateRange       DateRange       `json:"date_range" db:"date_range"`

	// Audit results
	Status          AuditStatus     `json:"status" db:"status"`
	ComplianceScore float64         `json:"compliance_score" db:"compliance_score"`
	Violations      json.RawMessage `json:"violations" db:"violations"`
	Recommendations json.RawMessage `json:"recommendations" db:"recommendations"`

	// Execution details
	StartedAt       time.Time       `json:"started_at" db:"started_at"`
	CompletedAt     *time.Time      `json:"completed_at" db:"completed_at"`
	ExecutedBy      uuid.UUID       `json:"executed_by" db:"executed_by"`

	// Report details
	ReportLocation  string          `json:"report_location" db:"report_location"`
	ReportFormat    string          `json:"report_format" db:"report_format"`

	Metadata        json.RawMessage `json:"metadata" db:"metadata"`
}

type AuditType string

const (
	AuditTypeCompliance AuditType = "COMPLIANCE"
	AuditTypeIntegrity  AuditType = "INTEGRITY"
	AuditTypeAccess     AuditType = "ACCESS"
	AuditTypeRetention  AuditType = "RETENTION"
)

type AuditStatus string

const (
	AuditStatusScheduled  AuditStatus = "SCHEDULED"
	AuditStatusRunning    AuditStatus = "RUNNING"
	AuditStatusCompleted  AuditStatus = "COMPLETED"
	AuditStatusFailed     AuditStatus = "FAILED"
)

// Filter structs for repository queries
type RetentionPolicyFilters struct {
	EntityType     string       `json:"entity_type,omitempty"`
	SchemaName     string       `json:"schema_name,omitempty"`
	DatabaseName   string       `json:"database_name,omitempty"`
	IsActive       *bool        `json:"is_active,omitempty"`
	PolicyType     *PolicyType  `json:"policy_type,omitempty"`
	ComplianceLevel *ComplianceLevel `json:"compliance_level,omitempty"`
}

type ArchivalJobFilters struct {
	PolicyID    *uuid.UUID `json:"policy_id,omitempty"`
	JobType     *JobType   `json:"job_type,omitempty"`
	Status      *JobStatus `json:"status,omitempty"`
	StartedAfter *time.Time `json:"started_after,omitempty"`
	StartedBefore *time.Time `json:"started_before,omitempty"`
}

type LegalHoldFilters struct {
	IsActive       *bool      `json:"is_active,omitempty"`
	CaseNumber     string     `json:"case_number,omitempty"`
	EntityTypes    []string   `json:"entity_types,omitempty"`
	StartDateAfter *time.Time `json:"start_date_after,omitempty"`
	StartDateBefore *time.Time `json:"start_date_before,omitempty"`
}

type ArchivedDataFilters struct {
	PolicyID       *uuid.UUID   `json:"policy_id,omitempty"`
	JobID          *uuid.UUID   `json:"job_id,omitempty"`
	SourceTable    string       `json:"source_table,omitempty"`
	SourceSchema   string       `json:"source_schema,omitempty"`
	SourceDatabase string       `json:"source_database,omitempty"`
	StorageTier    *StorageTier `json:"storage_tier,omitempty"`
	IsOnLegalHold  *bool        `json:"is_on_legal_hold,omitempty"`
	ArchivedAfter  *time.Time   `json:"archived_after,omitempty"`
	ArchivedBefore *time.Time   `json:"archived_before,omitempty"`
}

type DataInventoryFilters struct {
	DatabaseName       string              `json:"database_name,omitempty"`
	SchemaName         string              `json:"schema_name,omitempty"`
	TableName          string              `json:"table_name,omitempty"`
	DataClassification *DataClassification `json:"data_classification,omitempty"`
	ContainsPII        *bool               `json:"contains_pii,omitempty"`
	ContainsPHI        *bool               `json:"contains_phi,omitempty"`
	ContainsPCI        *bool               `json:"contains_pci,omitempty"`
}

type ComplianceAuditFilters struct {
	AuditType     *AuditType   `json:"audit_type,omitempty"`
	Status        *AuditStatus `json:"status,omitempty"`
	PolicyIDs     []uuid.UUID  `json:"policy_ids,omitempty"`
	EntityTypes   []string     `json:"entity_types,omitempty"`
	StartedAfter  *time.Time   `json:"started_after,omitempty"`
	StartedBefore *time.Time   `json:"started_before,omitempty"`
}