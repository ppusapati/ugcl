package handlers

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "p9e.in/ugcl/insighthub/api/proto"
	"p9e.in/ugcl/insighthub/api/proto/insighthubv1connect"
	"p9e.in/ugcl/insighthub/services"
)

// ReportHandler implements the InsightHub Connect RPC service
type ReportHandler struct {
	serviceManager *services.ServiceManager
}

// NewReportHandler creates a new report handler instance
func NewReportHandler(serviceManager *services.ServiceManager) *ReportHandler {
	return &ReportHandler{
		serviceManager: serviceManager,
	}
}

// GetPath returns the Connect RPC path for this handler
func (h *ReportHandler) GetPath() (string, any) {
	return insighthubv1connect.NewReportServiceHandler(h)
}

// =============================================================================
// Report CRUD Operations
// =============================================================================

func (h *ReportHandler) CreateReport(ctx context.Context, req *connect.Request[pb.CreateReportRequest]) (*connect.Response[pb.CreateReportResponse], error) {
	report, err := h.serviceManager.Report.CreateReport(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create report: %w", err))
	}

	resp := &pb.CreateReportResponse{
		Report: report,
	}

	return connect.NewResponse(resp), nil
}

func (h *ReportHandler) GetReport(ctx context.Context, req *connect.Request[pb.GetReportRequest]) (*connect.Response[pb.GetReportResponse], error) {
	reportID, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid report ID: %w", err))
	}

	report, err := h.serviceManager.Report.GetReport(ctx, reportID, req.Msg.UserId, req.Msg.UserRoles)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("report not found: %w", err))
	}

	resp := &pb.GetReportResponse{
		Report: report,
	}

	return connect.NewResponse(resp), nil
}

func (h *ReportHandler) UpdateReport(ctx context.Context, req *connect.Request[pb.UpdateReportRequest]) (*connect.Response[pb.UpdateReportResponse], error) {
	report, err := h.serviceManager.Report.UpdateReport(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update report: %w", err))
	}

	resp := &pb.UpdateReportResponse{
		Report: report,
	}

	return connect.NewResponse(resp), nil
}

func (h *ReportHandler) DeleteReport(ctx context.Context, req *connect.Request[pb.DeleteReportRequest]) (*connect.Response[pb.DeleteReportResponse], error) {
	reportID, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid report ID: %w", err))
	}

	err = h.serviceManager.Report.DeleteReport(ctx, reportID, req.Msg.DeletedBy)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to delete report: %w", err))
	}

	resp := &pb.DeleteReportResponse{
		Success: true,
	}

	return connect.NewResponse(resp), nil
}

func (h *ReportHandler) SearchReports(ctx context.Context, req *connect.Request[pb.SearchReportsRequest]) (*connect.Response[pb.SearchReportsResponse], error) {
	reports, totalCount, err := h.serviceManager.Report.SearchReports(ctx, req.Msg.Query, req.Msg.UserId, req.Msg.UserRoles, req.Msg.Limit, req.Msg.Offset)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to search reports: %w", err))
	}

	resp := &pb.SearchReportsResponse{
		Reports:    reports,
		TotalCount: totalCount,
	}

	return connect.NewResponse(resp), nil
}

func (h *ReportHandler) CloneReport(ctx context.Context, req *connect.Request[pb.CloneReportRequest]) (*connect.Response[pb.CloneReportResponse], error) {
	report, err := h.serviceManager.Report.CloneReport(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to clone report: %w", err))
	}

	resp := &pb.CloneReportResponse{
		Report: report,
	}

	return connect.NewResponse(resp), nil
}

// =============================================================================
// Report Field Management
// =============================================================================

func (h *ReportHandler) AddReportField(ctx context.Context, req *connect.Request[pb.AddReportFieldRequest]) (*connect.Response[pb.AddReportFieldResponse], error) {
	field, err := h.serviceManager.Report.AddReportField(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to add report field: %w", err))
	}

	resp := &pb.AddReportFieldResponse{
		Field: field,
	}

	return connect.NewResponse(resp), nil
}

func (h *ReportHandler) UpdateReportField(ctx context.Context, req *connect.Request[pb.UpdateReportFieldRequest]) (*connect.Response[pb.UpdateReportFieldResponse], error) {
	field, err := h.serviceManager.Report.UpdateReportField(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update report field: %w", err))
	}

	resp := &pb.UpdateReportFieldResponse{
		Field: field,
	}

	return connect.NewResponse(resp), nil
}

func (h *ReportHandler) RemoveReportField(ctx context.Context, req *connect.Request[pb.RemoveReportFieldRequest]) (*connect.Response[pb.RemoveReportFieldResponse], error) {
	fieldID, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid field ID: %w", err))
	}

	err = h.serviceManager.Report.RemoveReportField(ctx, fieldID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to remove report field: %w", err))
	}

	resp := &pb.RemoveReportFieldResponse{
		Success: true,
	}

	return connect.NewResponse(resp), nil
}

func (h *ReportHandler) ReorderReportFields(ctx context.Context, req *connect.Request[pb.ReorderReportFieldsRequest]) (*connect.Response[pb.ReorderReportFieldsResponse], error) {
	err := h.serviceManager.Report.ReorderReportFields(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to reorder report fields: %w", err))
	}

	resp := &pb.ReorderReportFieldsResponse{
		Success: true,
	}

	return connect.NewResponse(resp), nil
}

// =============================================================================
// Report Filter Management
// =============================================================================

func (h *ReportHandler) AddReportFilter(ctx context.Context, req *connect.Request[pb.AddReportFilterRequest]) (*connect.Response[pb.AddReportFilterResponse], error) {
	filter, err := h.serviceManager.Report.AddReportFilter(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to add report filter: %w", err))
	}

	resp := &pb.AddReportFilterResponse{
		Filter: filter,
	}

	return connect.NewResponse(resp), nil
}

func (h *ReportHandler) UpdateReportFilter(ctx context.Context, req *connect.Request[pb.UpdateReportFilterRequest]) (*connect.Response[pb.UpdateReportFilterResponse], error) {
	filter, err := h.serviceManager.Report.UpdateReportFilter(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update report filter: %w", err))
	}

	resp := &pb.UpdateReportFilterResponse{
		Filter: filter,
	}

	return connect.NewResponse(resp), nil
}

func (h *ReportHandler) RemoveReportFilter(ctx context.Context, req *connect.Request[pb.RemoveReportFilterRequest]) (*connect.Response[pb.RemoveReportFilterResponse], error) {
	filterID, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid filter ID: %w", err))
	}

	err = h.serviceManager.Report.RemoveReportFilter(ctx, filterID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to remove report filter: %w", err))
	}

	resp := &pb.RemoveReportFilterResponse{
		Success: true,
	}

	return connect.NewResponse(resp), nil
}

// =============================================================================
// Report Group Management
// =============================================================================

func (h *ReportHandler) AddReportGroup(ctx context.Context, req *connect.Request[pb.AddReportGroupRequest]) (*connect.Response[pb.AddReportGroupResponse], error) {
	group, err := h.serviceManager.Report.AddReportGroup(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to add report group: %w", err))
	}

	resp := &pb.AddReportGroupResponse{
		Group: group,
	}

	return connect.NewResponse(resp), nil
}

func (h *ReportHandler) UpdateReportGroup(ctx context.Context, req *connect.Request[pb.UpdateReportGroupRequest]) (*connect.Response[pb.UpdateReportGroupResponse], error) {
	group, err := h.serviceManager.Report.UpdateReportGroup(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update report group: %w", err))
	}

	resp := &pb.UpdateReportGroupResponse{
		Group: group,
	}

	return connect.NewResponse(resp), nil
}

func (h *ReportHandler) RemoveReportGroup(ctx context.Context, req *connect.Request[pb.RemoveReportGroupRequest]) (*connect.Response[pb.RemoveReportGroupResponse], error) {
	groupID, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid group ID: %w", err))
	}

	err = h.serviceManager.Report.RemoveReportGroup(ctx, groupID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to remove report group: %w", err))
	}

	resp := &pb.RemoveReportGroupResponse{
		Success: true,
	}

	return connect.NewResponse(resp), nil
}

// =============================================================================
// Report Sort Management
// =============================================================================

func (h *ReportHandler) AddReportSort(ctx context.Context, req *connect.Request[pb.AddReportSortRequest]) (*connect.Response[pb.AddReportSortResponse], error) {
	sort, err := h.serviceManager.Report.AddReportSort(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to add report sort: %w", err))
	}

	resp := &pb.AddReportSortResponse{
		Sort: sort,
	}

	return connect.NewResponse(resp), nil
}

func (h *ReportHandler) UpdateReportSort(ctx context.Context, req *connect.Request[pb.UpdateReportSortRequest]) (*connect.Response[pb.UpdateReportSortResponse], error) {
	sort, err := h.serviceManager.Report.UpdateReportSort(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update report sort: %w", err))
	}

	resp := &pb.UpdateReportSortResponse{
		Sort: sort,
	}

	return connect.NewResponse(resp), nil
}

func (h *ReportHandler) RemoveReportSort(ctx context.Context, req *connect.Request[pb.RemoveReportSortRequest]) (*connect.Response[pb.RemoveReportSortResponse], error) {
	sortID, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid sort ID: %w", err))
	}

	err = h.serviceManager.Report.RemoveReportSort(ctx, sortID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to remove report sort: %w", err))
	}

	resp := &pb.RemoveReportSortResponse{
		Success: true,
	}

	return connect.NewResponse(resp), nil
}

// =============================================================================
// Report Chart Management
// =============================================================================

func (h *ReportHandler) SetReportChart(ctx context.Context, req *connect.Request[pb.SetReportChartRequest]) (*connect.Response[pb.SetReportChartResponse], error) {
	chart, err := h.serviceManager.Report.SetReportChart(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to set report chart: %w", err))
	}

	resp := &pb.SetReportChartResponse{
		Chart: chart,
	}

	return connect.NewResponse(resp), nil
}

func (h *ReportHandler) RemoveReportChart(ctx context.Context, req *connect.Request[pb.RemoveReportChartRequest]) (*connect.Response[pb.RemoveReportChartResponse], error) {
	reportID, err := uuid.Parse(req.Msg.ReportId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid report ID: %w", err))
	}

	err = h.serviceManager.Report.RemoveReportChart(ctx, reportID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to remove report chart: %w", err))
	}

	resp := &pb.RemoveReportChartResponse{
		Success: true,
	}

	return connect.NewResponse(resp), nil
}

// =============================================================================
// Report Validation and Preview
// =============================================================================

func (h *ReportHandler) ValidateReport(ctx context.Context, req *connect.Request[pb.ValidateReportRequest]) (*connect.Response[pb.ValidateReportResponse], error) {
	reportID, err := uuid.Parse(req.Msg.ReportId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid report ID: %w", err))
	}

	resp, err := h.serviceManager.Report.ValidateReport(ctx, reportID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to validate report: %w", err))
	}

	return connect.NewResponse(resp), nil
}

func (h *ReportHandler) PreviewReportQuery(ctx context.Context, req *connect.Request[pb.PreviewReportQueryRequest]) (*connect.Response[pb.PreviewReportQueryResponse], error) {
	resp, err := h.serviceManager.Report.PreviewReportQuery(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to preview report query: %w", err))
	}

	return connect.NewResponse(resp), nil
}

// =============================================================================
// Report Permission Management
// =============================================================================

func (h *ReportHandler) GrantReportPermission(ctx context.Context, req *connect.Request[pb.GrantReportPermissionRequest]) (*connect.Response[pb.GrantReportPermissionResponse], error) {
	permission, err := h.serviceManager.Report.GrantReportPermission(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to grant report permission: %w", err))
	}

	resp := &pb.GrantReportPermissionResponse{
		Permission: permission,
	}

	return connect.NewResponse(resp), nil
}

func (h *ReportHandler) RevokeReportPermission(ctx context.Context, req *connect.Request[pb.RevokeReportPermissionRequest]) (*connect.Response[pb.RevokeReportPermissionResponse], error) {
	permissionID, err := uuid.Parse(req.Msg.PermissionId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid permission ID: %w", err))
	}

	err = h.serviceManager.Report.RevokeReportPermission(ctx, permissionID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to revoke report permission: %w", err))
	}

	resp := &pb.RevokeReportPermissionResponse{
		Success: true,
	}

	return connect.NewResponse(resp), nil
}

func (h *ReportHandler) GetReportPermissions(ctx context.Context, req *connect.Request[pb.GetReportPermissionsRequest]) (*connect.Response[pb.GetReportPermissionsResponse], error) {
	reportID, err := uuid.Parse(req.Msg.ReportId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid report ID: %w", err))
	}

	permissions, err := h.serviceManager.Report.GetReportPermissions(ctx, reportID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get report permissions: %w", err))
	}

	resp := &pb.GetReportPermissionsResponse{
		Permissions: permissions,
	}

	return connect.NewResponse(resp), nil
}