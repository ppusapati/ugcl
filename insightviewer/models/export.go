package models

import (
	"time"

	"github.com/google/uuid"
)

// ReportExport represents an export job for a report result
type ReportExport struct {
	ID        uuid.UUID
	RunID     uuid.UUID
	Format    string // csv, excel, pdf, json
	Status    string // processing, completed, failed
	FilePath  *string
	FileSize  *int64
	CreatedBy string
	CreatedAt time.Time
	CompletedAt *time.Time
	ExpiresAt   *time.Time // When the export file will be deleted
	DownloadCount int32
}

// ExportTemplate represents a reusable export configuration
type ExportTemplate struct {
	ID          uuid.UUID
	Name        string
	Description string
	Format      string
	Settings    map[string]interface{} // Format-specific settings
	IsActive    bool
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedBy   string
	UpdatedAt   time.Time
}