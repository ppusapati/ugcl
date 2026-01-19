package services

import (
	"context"

	"github.com/google/uuid"

	pb "p9e.in/ugcl/insighthub/api/proto"
)

// IReportService defines the interface for report management operations
type IReportService interface {
	// Report CRUD operations
	CreateReport(ctx context.Context, req *pb.CreateReportRequest) (*pb.Report, error)
	GetReport(ctx context.Context, reportID uuid.UUID, userID string, userRoles []string) (*pb.Report, error)
	UpdateReport(ctx context.Context, req *pb.UpdateReportRequest) (*pb.Report, error)
	DeleteReport(ctx context.Context, reportID uuid.UUID, deletedBy string) error
	SearchReports(ctx context.Context, query string, userID string, userRoles []string, limit, offset int32) ([]*pb.Report, int32, error)
	CloneReport(ctx context.Context, req *pb.CloneReportRequest) (*pb.Report, error)

	// Report field management
	AddReportField(ctx context.Context, req *pb.AddReportFieldRequest) (*pb.ReportField, error)
	UpdateReportField(ctx context.Context, req *pb.UpdateReportFieldRequest) (*pb.ReportField, error)
	RemoveReportField(ctx context.Context, fieldID uuid.UUID) error
	ReorderReportFields(ctx context.Context, req *pb.ReorderReportFieldsRequest) error

	// Report filter management
	AddReportFilter(ctx context.Context, req *pb.AddReportFilterRequest) (*pb.ReportFilter, error)
	UpdateReportFilter(ctx context.Context, req *pb.UpdateReportFilterRequest) (*pb.ReportFilter, error)
	RemoveReportFilter(ctx context.Context, filterID uuid.UUID) error

	// Report group management
	AddReportGroup(ctx context.Context, req *pb.AddReportGroupRequest) (*pb.ReportGroup, error)
	UpdateReportGroup(ctx context.Context, req *pb.UpdateReportGroupRequest) (*pb.ReportGroup, error)
	RemoveReportGroup(ctx context.Context, groupID uuid.UUID) error

	// Report sort management
	AddReportSort(ctx context.Context, req *pb.AddReportSortRequest) (*pb.ReportSort, error)
	UpdateReportSort(ctx context.Context, req *pb.UpdateReportSortRequest) (*pb.ReportSort, error)
	RemoveReportSort(ctx context.Context, sortID uuid.UUID) error

	// Report chart management
	SetReportChart(ctx context.Context, req *pb.SetReportChartRequest) (*pb.ReportChart, error)
	RemoveReportChart(ctx context.Context, reportID uuid.UUID) error

	// Report validation and preview
	ValidateReport(ctx context.Context, reportID uuid.UUID) (*pb.ValidateReportResponse, error)
	PreviewReportQuery(ctx context.Context, req *pb.PreviewReportQueryRequest) (*pb.PreviewReportQueryResponse, error)

	// Report permissions
	GrantReportPermission(ctx context.Context, req *pb.GrantReportPermissionRequest) (*pb.ReportPermission, error)
	RevokeReportPermission(ctx context.Context, permissionID uuid.UUID) error
	GetReportPermissions(ctx context.Context, reportID uuid.UUID) ([]*pb.ReportPermission, error)
}

// ServiceManager combines all service interfaces for insighthub module
type ServiceManager struct {
	Report IReportService
}

// NewServiceManager creates a new service manager instance for insighthub module
func NewServiceManager(reportService IReportService) *ServiceManager {
	return &ServiceManager{
		Report: reportService,
	}
}