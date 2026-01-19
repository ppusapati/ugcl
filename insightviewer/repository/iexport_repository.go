package repository

import (
	"context"

	"github.com/google/uuid"

	"p9e.in/ugcl/insightviewer/models"
)

// IExportRepository defines the interface for report export data access operations
type IExportRepository interface {
	// Report Export operations
	CreateReportExport(ctx context.Context, export *models.ReportExport) (*models.ReportExport, error)
	GetReportExportByID(ctx context.Context, id uuid.UUID) (*models.ReportExport, error)
	UpdateReportExport(ctx context.Context, export *models.ReportExport) (*models.ReportExport, error)
	CompleteReportExport(ctx context.Context, exportID uuid.UUID, status string, filePath *string, fileSize *int64) (*models.ReportExport, error)
	ListReportExports(ctx context.Context, reportID uuid.UUID, limit, offset int32) ([]*models.ReportExport, int32, error)
	DeleteReportExport(ctx context.Context, id uuid.UUID) error
	IncrementDownloadCount(ctx context.Context, exportID uuid.UUID) error

	// Export Template operations
	CreateExportTemplate(ctx context.Context, template *models.ExportTemplate) (*models.ExportTemplate, error)
	GetExportTemplateByID(ctx context.Context, id uuid.UUID) (*models.ExportTemplate, error)
	UpdateExportTemplate(ctx context.Context, template *models.ExportTemplate) (*models.ExportTemplate, error)
	DeleteExportTemplate(ctx context.Context, id uuid.UUID) error
	ListExportTemplates(ctx context.Context, limit, offset int32) ([]*models.ExportTemplate, int32, error)
}