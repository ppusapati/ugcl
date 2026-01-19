package models

import (
	"time"

	"github.com/google/uuid"
)

// ReportRun represents a single execution of a report
type ReportRun struct {
	ID         uuid.UUID
	ReportID   uuid.UUID
	RunBy      string
	Status     string // running, completed, failed, cancelled
	Parameters map[string]interface{}
	StartedAt  time.Time
	CompletedAt *time.Time
	DurationMs *int32
	ErrorMsg   *string
	CreatedAt  time.Time
}

// ReportResult represents the output data from a report execution
type ReportResult struct {
	ID         uuid.UUID
	RunID      uuid.UUID
	ResultData map[string]interface{} // JSON data containing columns and rows
	RowCount   int32
	FileSize   *int64 // Size in bytes if stored as file
	FilePath   *string // Path to file if large result set
	CreatedAt  time.Time
}

// ReportCache represents cached report results for performance optimization
type ReportCache struct {
	ID         uuid.UUID
	ReportID   uuid.UUID
	CacheKey   string // Hash of report parameters and filters
	ResultData map[string]interface{}
	CreatedAt  time.Time
	ExpiresAt  time.Time
	HitCount   int32
	LastAccessed time.Time
}