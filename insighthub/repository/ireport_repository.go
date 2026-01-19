package repository

import (
	"context"

	"github.com/google/uuid"

	"p9e.in/ugcl/insighthub/models"
)

// IReportRepository defines the interface for report data access operations
type IReportRepository interface {
	// Report CRUD operations
	CreateReport(ctx context.Context, report *models.Report) (*models.Report, error)
	GetReportByID(ctx context.Context, id uuid.UUID) (*models.Report, error)
	UpdateReport(ctx context.Context, report *models.Report) (*models.Report, error)
	DeleteReport(ctx context.Context, id uuid.UUID, deletedBy string) error
	SearchReports(ctx context.Context, query string, limit, offset int32) ([]*models.Report, int32, error)
	CloneReport(ctx context.Context, sourceID uuid.UUID, newName, createdBy string) (*models.Report, error)

	// Report field operations
	AddReportField(ctx context.Context, field *models.ReportField) (*models.ReportField, error)
	UpdateReportField(ctx context.Context, field *models.ReportField) (*models.ReportField, error)
	RemoveReportField(ctx context.Context, id uuid.UUID) error
	GetReportFields(ctx context.Context, reportID uuid.UUID) ([]*models.ReportField, error)

	// Report filter operations
	AddReportFilter(ctx context.Context, filter *models.ReportFilter) (*models.ReportFilter, error)
	UpdateReportFilter(ctx context.Context, filter *models.ReportFilter) (*models.ReportFilter, error)
	RemoveReportFilter(ctx context.Context, id uuid.UUID) error
	GetReportFilters(ctx context.Context, reportID uuid.UUID) ([]*models.ReportFilter, error)

	// Report group operations
	AddReportGroup(ctx context.Context, group *models.ReportGroup) (*models.ReportGroup, error)
	UpdateReportGroup(ctx context.Context, group *models.ReportGroup) (*models.ReportGroup, error)
	RemoveReportGroup(ctx context.Context, id uuid.UUID) error
	GetReportGroups(ctx context.Context, reportID uuid.UUID) ([]*models.ReportGroup, error)

	// Report sort operations
	AddReportSort(ctx context.Context, sort *models.ReportSort) (*models.ReportSort, error)
	UpdateReportSort(ctx context.Context, sort *models.ReportSort) (*models.ReportSort, error)
	RemoveReportSort(ctx context.Context, id uuid.UUID) error
	GetReportSorts(ctx context.Context, reportID uuid.UUID) ([]*models.ReportSort, error)

	// Report chart operations
	SetReportChart(ctx context.Context, chart *models.ReportChart) (*models.ReportChart, error)
	RemoveReportChart(ctx context.Context, reportID uuid.UUID) error
	GetReportChart(ctx context.Context, reportID uuid.UUID) (*models.ReportChart, error)

	// Report permission operations
	GrantReportPermission(ctx context.Context, permission *models.ReportPermission) (*models.ReportPermission, error)
	RevokeReportPermission(ctx context.Context, id uuid.UUID) error
	GetReportPermissions(ctx context.Context, reportID uuid.UUID) ([]*models.ReportPermission, error)
}

// RepositoryManager combines all repository interfaces for insighthub module
type RepositoryManager struct {
	Report IReportRepository
}

// NewRepositoryManager creates a new repository manager instance
func NewRepositoryManager(reportRepo IReportRepository) *RepositoryManager {
	return &RepositoryManager{
		Report: reportRepo,
	}
}