// =============================================================================
// models/service_models.go - Service layer data models
// =============================================================================
package models

import (
	"time"

	pb "p9e.in/ugcl/databridge/api/databridge"

	"github.com/google/uuid"
)

// ValidationError represents a field validation error
type ValidationError struct {
	RowNumber    int32             `json:"row_number"`
	FieldName    string            `json:"field_name"`
	ColumnName   string            `json:"column_name"`
	Value        string            `json:"value"`
	ErrorType    string            `json:"error_type"`
	ErrorMessage string            `json:"error_message"`
	ErrorCode    string            `json:"error_code"`
	Severity     ValidationLevel   `json:"severity"`
}

// ValidationLevel represents the severity of validation errors
type ValidationLevel string

const (
	ValidationLevelError   ValidationLevel = "error"
	ValidationLevelWarning ValidationLevel = "warning"
	ValidationLevelInfo    ValidationLevel = "info"
)

// ValidationResult contains the results of CSV validation
type ValidationResult struct {
	IsValid          bool               `json:"is_valid"`
	TotalRows        int32              `json:"total_rows"`
	ValidRows        int32              `json:"valid_rows"`
	InvalidRows      int32              `json:"invalid_rows"`
	Errors           []*ValidationError `json:"errors"`
	Warnings         []*ValidationError `json:"warnings"`
	SkippedRows      int32              `json:"skipped_rows"`
	ProcessingTimeMs int64              `json:"processing_time_ms"`
}

// ColumnTypeInference represents inferred column information from CSV data
type ColumnTypeInference struct {
	Header           string        `json:"header"`
	InferredType     pb.DataType   `json:"inferred_type"`
	Confidence       float64       `json:"confidence"`
	SampleValues     []string      `json:"sample_values"`
	NullCount        int32         `json:"null_count"`
	UniqueCount      int32         `json:"unique_count"`
	MaxLength        int32         `json:"max_length"`
	MinLength        int32         `json:"min_length"`
	HasDuplicates    bool          `json:"has_duplicates"`
	PatternMatches   []string      `json:"pattern_matches"`
}

// CSVParseResult contains the results of CSV parsing
type CSVParseResult struct {
	Headers       []string    `json:"headers"`
	Rows          [][]string  `json:"rows"`
	TotalRows     int32       `json:"total_rows"`
	DataRows      int32       `json:"data_rows"`
	Encoding      string      `json:"encoding"`
	Delimiter     string      `json:"delimiter"`
	HasHeader     bool        `json:"has_header"`
	ParseErrors   []string    `json:"parse_errors"`
	ParseTimeMs   int64       `json:"parse_time_ms"`
}

// ImportStats contains statistics about an import operation
type ImportStats struct {
	JobID            uuid.UUID         `json:"job_id"`
	Status           string            `json:"status"`
	TotalRows        int32             `json:"total_rows"`
	ProcessedRows    int32             `json:"processed_rows"`
	SuccessfulRows   int32             `json:"successful_rows"`
	FailedRows       int32             `json:"failed_rows"`
	SkippedRows      int32             `json:"skipped_rows"`
	ErrorSummary     []*ValidationError `json:"error_summary"`
	StartTime        time.Time         `json:"start_time"`
	EndTime          *time.Time        `json:"end_time,omitempty"`
	ProcessingTimeMs int64             `json:"processing_time_ms"`
	ThroughputRPS    float64           `json:"throughput_rps"`
}

// TransformationRule defines how to transform CSV field values
type TransformationRule struct {
	Type        string            `json:"type"`        // trim, uppercase, lowercase, date, decimal, etc.
	Parameters  map[string]string `json:"parameters"`  // Additional parameters for transformation
	Description string            `json:"description"`
}

// MappingValidationResult contains validation results for field mappings
type MappingValidationResult struct {
	IsValid           bool                `json:"is_valid"`
	MissingRequired   []string            `json:"missing_required"`
	InvalidMappings   []string            `json:"invalid_mappings"`
	TypeMismatches    []TypeMismatch      `json:"type_mismatches"`
	Suggestions       []MappingSuggestion `json:"suggestions"`
	CompatibilityScore float64            `json:"compatibility_score"`
}

// TypeMismatch represents a data type compatibility issue
type TypeMismatch struct {
	CSVField     string      `json:"csv_field"`
	DBColumn     string      `json:"db_column"`
	CSVType      string      `json:"csv_type"`
	DBType       pb.DataType `json:"db_type"`
	Convertible  bool        `json:"convertible"`
	Confidence   float64     `json:"confidence"`
	SampleValues []string    `json:"sample_values"`
}

// MappingSuggestion provides automatic mapping suggestions
type MappingSuggestion struct {
	CSVField     string  `json:"csv_field"`
	DBColumn     string  `json:"db_column"`
	Confidence   float64 `json:"confidence"`
	Reason       string  `json:"reason"`
	Transformation *TransformationRule `json:"transformation,omitempty"`
}

// ImportProgress represents real-time import progress
type ImportProgress struct {
	JobID              uuid.UUID          `json:"job_id"`
	Phase              ImportPhase        `json:"phase"`
	TotalRows          int32              `json:"total_rows"`
	ProcessedRows      int32              `json:"processed_rows"`
	SuccessfulRows     int32              `json:"successful_rows"`
	FailedRows         int32              `json:"failed_rows"`
	ProgressPercentage float64            `json:"progress_percentage"`
	EstimatedTimeLeft  *time.Duration     `json:"estimated_time_left,omitempty"`
	ThroughputRPS      float64            `json:"throughput_rps"`
	RecentErrors       []*ValidationError `json:"recent_errors"`
	Message            string             `json:"message"`
	Timestamp          time.Time          `json:"timestamp"`
}

// ImportPhase represents the current phase of import operation
type ImportPhase string

const (
	ImportPhaseInitializing ImportPhase = "initializing"
	ImportPhaseParsing      ImportPhase = "parsing"
	ImportPhaseValidating   ImportPhase = "validating"
	ImportPhaseTransforming ImportPhase = "transforming"
	ImportPhaseImporting    ImportPhase = "importing"
	ImportPhaseFinalizing   ImportPhase = "finalizing"
	ImportPhaseCompleted    ImportPhase = "completed"
	ImportPhaseFailed       ImportPhase = "failed"
	ImportPhaseCancelled    ImportPhase = "cancelled"
)

// CSVAnalysisResult contains detailed CSV file analysis
type CSVAnalysisResult struct {
	FileInfo         *FileInfo                 `json:"file_info"`
	Structure        *CSVStructure             `json:"structure"`
	ColumnAnalysis   []*ColumnTypeInference    `json:"column_analysis"`
	DataQuality      *DataQuality              `json:"data_quality"`
	Recommendations  []*ImportRecommendation   `json:"recommendations"`
	AnalysisTimeMs   int64                     `json:"analysis_time_ms"`
}

// FileInfo contains basic file information
type FileInfo struct {
	FileName    string    `json:"file_name"`
	FileSize    int64     `json:"file_size"`
	Encoding    string    `json:"encoding"`
	MimeType    string    `json:"mime_type"`
	Hash        string    `json:"hash"`
	UploadedAt  time.Time `json:"uploaded_at"`
}

// CSVStructure contains CSV structural information
type CSVStructure struct {
	Delimiter     string   `json:"delimiter"`
	HasHeader     bool     `json:"has_header"`
	Headers       []string `json:"headers"`
	TotalRows     int32    `json:"total_rows"`
	DataRows      int32    `json:"data_rows"`
	EmptyRows     int32    `json:"empty_rows"`
	ColumnsCount  int32    `json:"columns_count"`
	MaxRowLength  int32    `json:"max_row_length"`
	MinRowLength  int32    `json:"min_row_length"`
	InconsistentRows int32 `json:"inconsistent_rows"`
}

// DataQuality contains data quality metrics
type DataQuality struct {
	OverallScore       float64               `json:"overall_score"`
	Completeness       float64               `json:"completeness"`
	Consistency        float64               `json:"consistency"`
	Validity           float64               `json:"validity"`
	Uniqueness         float64               `json:"uniqueness"`
	Issues             []*DataQualityIssue   `json:"issues"`
}

// DataQualityIssue represents a data quality problem
type DataQualityIssue struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Column      string `json:"column"`
	Description string `json:"description"`
	Count       int32  `json:"count"`
	Examples    []string `json:"examples"`
}

// ImportRecommendation provides suggestions for import optimization
type ImportRecommendation struct {
	Type        string `json:"type"`
	Priority    string `json:"priority"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Action      string `json:"action"`
	Impact      string `json:"impact"`
}

// BatchImportConfig contains configuration for batch import operations
type BatchImportConfig struct {
	BatchSize        int32         `json:"batch_size"`
	MaxConcurrency   int32         `json:"max_concurrency"`
	TimeoutSeconds   int32         `json:"timeout_seconds"`
	RetryAttempts    int32         `json:"retry_attempts"`
	ValidationMode   string        `json:"validation_mode"` // strict, lenient, skip
	ErrorHandling    string        `json:"error_handling"`  // fail_fast, continue, rollback
	ProgressInterval time.Duration `json:"progress_interval"`
}