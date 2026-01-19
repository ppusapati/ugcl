package repository

import (
	"context"

	"github.com/google/uuid"

	"p9e.in/ugcl/insightviewer/models"
)

// IExecutionRepository defines the interface for report execution data access operations
type IExecutionRepository interface {
	// Report Run operations
	CreateReportRun(ctx context.Context, run *models.ReportRun) (*models.ReportRun, error)
	GetReportRunByID(ctx context.Context, id uuid.UUID) (*models.ReportRun, error)
	UpdateReportRun(ctx context.Context, run *models.ReportRun) (*models.ReportRun, error)
	GetReportRuns(ctx context.Context, reportID uuid.UUID, limit, offset int32) ([]*models.ReportRun, int32, error)
	CompleteReportRun(ctx context.Context, runID uuid.UUID, status string, durationMs *int32, errorMsg *string) (*models.ReportRun, error)

	// Report Result operations
	CreateReportResult(ctx context.Context, result *models.ReportResult) (*models.ReportResult, error)
	GetReportResultByRunID(ctx context.Context, runID uuid.UUID) (*models.ReportResult, error)
	GetReportResultByID(ctx context.Context, id uuid.UUID) (*models.ReportResult, error)
	DeleteReportResult(ctx context.Context, id uuid.UUID) error

	// Report Cache operations
	CreateReportCache(ctx context.Context, cache *models.ReportCache) (*models.ReportCache, error)
	GetCachedResult(ctx context.Context, cacheKey string) (*models.ReportCache, error)
	UpdateCacheAccess(ctx context.Context, cacheKey string) error
	InvalidateReportCache(ctx context.Context, reportID uuid.UUID) error
	CleanExpiredCache(ctx context.Context) error
}