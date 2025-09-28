package services

import (
	"context"

	"github.com/google/uuid"

	pb "p9e.in/ugcl/insightviewer/api/proto"
)

// IViewerService defines the interface for report execution and viewing operations
type IViewerService interface {
	// Report execution
	ExecuteReport(ctx context.Context, req *pb.ExecuteReportRequest) (*pb.ExecuteReportResponse, error)
	GetReportResult(ctx context.Context, resultID uuid.UUID) (*pb.ReportResult, error)
	GetReportRuns(ctx context.Context, reportID uuid.UUID, limit, offset int32) ([]*pb.ReportRun, int32, error)

	// Report caching
	CacheReportResult(ctx context.Context, req *pb.CacheReportResultRequest) (*pb.ReportCache, error)
	GetCachedResult(ctx context.Context, cacheKey string) (*pb.ReportCache, error)
	InvalidateCache(ctx context.Context, reportID uuid.UUID) error

	// Report exports
	ExportReport(ctx context.Context, req *pb.ExportReportRequest) (*pb.ReportExport, error)
	GetReportExport(ctx context.Context, exportID uuid.UUID) (*pb.ReportExport, error)
	ListReportExports(ctx context.Context, reportID uuid.UUID, limit, offset int32) ([]*pb.ReportExport, int32, error)

	// Report scheduling
	CreateReportSchedule(ctx context.Context, req *pb.CreateReportScheduleRequest) (*pb.ReportSchedule, error)
	UpdateReportSchedule(ctx context.Context, req *pb.UpdateReportScheduleRequest) (*pb.ReportSchedule, error)
	DeleteReportSchedule(ctx context.Context, scheduleID uuid.UUID) error
	GetReportSchedules(ctx context.Context, reportID uuid.UUID) ([]*pb.ReportSchedule, error)
	GetActiveSchedules(ctx context.Context) ([]*pb.ReportSchedule, error)

	// Report alerts
	CreateReportAlert(ctx context.Context, req *pb.CreateReportAlertRequest) (*pb.ReportAlert, error)
	UpdateReportAlert(ctx context.Context, req *pb.UpdateReportAlertRequest) (*pb.ReportAlert, error)
	DeleteReportAlert(ctx context.Context, alertID uuid.UUID) error
	GetReportAlerts(ctx context.Context, reportID uuid.UUID) ([]*pb.ReportAlert, error)

	// Report subscriptions
	CreateReportSubscription(ctx context.Context, req *pb.CreateReportSubscriptionRequest) (*pb.ReportSubscription, error)
	UpdateReportSubscription(ctx context.Context, req *pb.UpdateReportSubscriptionRequest) (*pb.ReportSubscription, error)
	DeleteReportSubscription(ctx context.Context, subscriptionID uuid.UUID) error
	GetReportSubscriptions(ctx context.Context, reportID uuid.UUID) ([]*pb.ReportSubscription, error)
	GetUserSubscriptions(ctx context.Context, userID string) ([]*pb.ReportSubscription, error)
}

// ServiceManager combines all service interfaces for insightviewer module
type ServiceManager struct {
	Viewer IViewerService
}

// NewServiceManager creates a new service manager instance for insightviewer module
func NewServiceManager(viewerService IViewerService) *ServiceManager {
	return &ServiceManager{
		Viewer: viewerService,
	}
}