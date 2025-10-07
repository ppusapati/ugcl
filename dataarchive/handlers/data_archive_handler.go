package handlers

import (
	"context"
	"fmt"
	"log"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"github.com/google/uuid"

	dataarchivepb "p9e.in/ugcl/dataarchive/api/proto"
	"p9e.in/ugcl/dataarchive/models"
	"p9e.in/ugcl/dataarchive/services/interfaces"
)

type DataArchiveHandler struct {
	dataArchiveService interfaces.DataArchiveService
}

func NewDataArchiveHandler(dataArchiveService interfaces.DataArchiveService) *DataArchiveHandler {
	return &DataArchiveHandler{
		dataArchiveService: dataArchiveService,
	}
}

// Retention Policy Management

func (h *DataArchiveHandler) CreateRetentionPolicy(ctx context.Context, req *connect.Request[dataarchivepb.CreateRetentionPolicyRequest]) (*connect.Response[dataarchivepb.CreateRetentionPolicyResponse], error) {
	log.Printf("CreateRetentionPolicy request for policy: %s", req.Msg.Policy.Name)

	if req.Msg.Policy == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("policy is required"))
	}

	serviceReq := h.convertProtoToCreatePolicyRequest(req.Msg.Policy)
	policy, err := h.dataArchiveService.CreateRetentionPolicy(ctx, serviceReq)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&dataarchivepb.CreateRetentionPolicyResponse{
		Policy: h.convertPolicyToProto(policy),
	}), nil
}

func (h *DataArchiveHandler) GetRetentionPolicy(ctx context.Context, req *connect.Request[dataarchivepb.GetRetentionPolicyRequest]) (*connect.Response[dataarchivepb.GetRetentionPolicyResponse], error) {
	log.Printf("GetRetentionPolicy request for ID: %s", req.Msg.Id)

	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("policy ID is required"))
	}

	id, err := h.parseUUID(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid policy ID: %w", err))
	}

	policy, err := h.dataArchiveService.GetRetentionPolicy(ctx, id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&dataarchivepb.GetRetentionPolicyResponse{
		Policy: h.convertPolicyToProto(policy),
	}), nil
}

func (h *DataArchiveHandler) UpdateRetentionPolicy(ctx context.Context, req *connect.Request[dataarchivepb.UpdateRetentionPolicyRequest]) (*connect.Response[dataarchivepb.UpdateRetentionPolicyResponse], error) {
	log.Printf("UpdateRetentionPolicy request for policy: %s", req.Msg.Policy.Id)

	if req.Msg.Policy == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("policy is required"))
	}

	serviceReq := h.convertProtoToUpdatePolicyRequest(req.Msg.Policy)
	policy, err := h.dataArchiveService.UpdateRetentionPolicy(ctx, serviceReq)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&dataarchivepb.UpdateRetentionPolicyResponse{
		Policy: h.convertPolicyToProto(policy),
	}), nil
}

func (h *DataArchiveHandler) DeleteRetentionPolicy(ctx context.Context, req *connect.Request[dataarchivepb.DeleteRetentionPolicyRequest]) (*connect.Response[emptypb.Empty], error) {
	log.Printf("DeleteRetentionPolicy request for ID: %s", req.Msg.Id)

	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("policy ID is required"))
	}

	id, err := h.parseUUID(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid policy ID: %w", err))
	}

	err = h.dataArchiveService.DeleteRetentionPolicy(ctx, id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *DataArchiveHandler) ListRetentionPolicies(ctx context.Context, req *connect.Request[dataarchivepb.ListRetentionPoliciesRequest]) (*connect.Response[dataarchivepb.ListRetentionPoliciesResponse], error) {
	log.Printf("ListRetentionPolicies request")

	serviceReq := &interfaces.ListRetentionPoliciesRequest{
		PageSize:     req.Msg.PageSize,
		PageToken:    req.Msg.PageToken,
		EntityType:   req.Msg.EntityType,
		ActiveOnly:   req.Msg.ActiveOnly,
	}

	policies, totalCount, nextPageToken, err := h.dataArchiveService.ListRetentionPolicies(ctx, serviceReq)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoPolicies := make([]*dataarchivepb.RetentionPolicy, len(policies))
	for i, policy := range policies {
		protoPolicies[i] = h.convertPolicyToProto(policy)
	}

	return connect.NewResponse(&dataarchivepb.ListRetentionPoliciesResponse{
		Policies:      protoPolicies,
		NextPageToken: nextPageToken,
		TotalCount:    totalCount,
	}), nil
}

func (h *DataArchiveHandler) ExecuteRetentionPolicy(ctx context.Context, req *connect.Request[dataarchivepb.ExecuteRetentionPolicyRequest]) (*connect.Response[dataarchivepb.ExecuteRetentionPolicyResponse], error) {
	log.Printf("ExecuteRetentionPolicy request for policy: %s", req.Msg.PolicyId)

	if req.Msg.PolicyId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("policy ID is required"))
	}

	policyID, err := h.parseUUID(req.Msg.PolicyId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid policy ID: %w", err))
	}

	serviceReq := &interfaces.StartArchivalJobRequest{
		PolicyID:  policyID,
		DryRun:    req.Msg.DryRun,
	}

	if req.Msg.DateRange != nil {
		serviceReq.DateRange = h.convertProtoDateRange(req.Msg.DateRange)
	}

	job, err := h.dataArchiveService.StartArchivalJob(ctx, serviceReq)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&dataarchivepb.ExecuteRetentionPolicyResponse{
		Job: h.convertJobToProto(job),
	}), nil
}

// Archival Job Management

func (h *DataArchiveHandler) GetArchivalJob(ctx context.Context, req *connect.Request[dataarchivepb.GetArchivalJobRequest]) (*connect.Response[dataarchivepb.GetArchivalJobResponse], error) {
	log.Printf("GetArchivalJob request for ID: %s", req.Msg.Id)

	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("job ID is required"))
	}

	id, err := h.parseUUID(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid job ID: %w", err))
	}

	job, err := h.dataArchiveService.MonitorArchivalJob(ctx, id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&dataarchivepb.GetArchivalJobResponse{
		Job: h.convertJobToProto(job),
	}), nil
}

func (h *DataArchiveHandler) ListArchivalJobs(ctx context.Context, req *connect.Request[dataarchivepb.ListArchivalJobsRequest]) (*connect.Response[dataarchivepb.ListArchivalJobsResponse], error) {
	log.Printf("ListArchivalJobs request")

	serviceReq := &interfaces.ListArchivalJobsRequest{
		PageSize:  req.Msg.PageSize,
		PageToken: req.Msg.PageToken,
	}

	if req.Msg.PolicyId != "" {
		policyID, err := h.parseUUID(req.Msg.PolicyId)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid policy ID: %w", err))
		}
		serviceReq.PolicyID = &policyID
	}

	if req.Msg.CreatedAfter != nil {
		createdAfter := req.Msg.CreatedAfter.AsTime()
		serviceReq.StartedAfter = &createdAfter
	}

	if req.Msg.CreatedBefore != nil {
		createdBefore := req.Msg.CreatedBefore.AsTime()
		serviceReq.StartedBefore = &createdBefore
	}

	jobs, totalCount, nextPageToken, err := h.dataArchiveService.ListArchivalJobs(ctx, serviceReq)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoJobs := make([]*dataarchivepb.ArchivalJob, len(jobs))
	for i, job := range jobs {
		protoJobs[i] = h.convertJobToProto(job)
	}

	return connect.NewResponse(&dataarchivepb.ListArchivalJobsResponse{
		Jobs:          protoJobs,
		NextPageToken: nextPageToken,
		TotalCount:    totalCount,
	}), nil
}

func (h *DataArchiveHandler) CancelArchivalJob(ctx context.Context, req *connect.Request[dataarchivepb.CancelArchivalJobRequest]) (*connect.Response[emptypb.Empty], error) {
	log.Printf("CancelArchivalJob request for ID: %s", req.Msg.Id)

	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("job ID is required"))
	}

	id, err := h.parseUUID(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid job ID: %w", err))
	}

	err = h.dataArchiveService.CancelArchivalJob(ctx, id, req.Msg.Reason)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *DataArchiveHandler) RetryArchivalJob(ctx context.Context, req *connect.Request[dataarchivepb.RetryArchivalJobRequest]) (*connect.Response[dataarchivepb.RetryArchivalJobResponse], error) {
	log.Printf("RetryArchivalJob request for ID: %s", req.Msg.Id)

	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("job ID is required"))
	}

	id, err := h.parseUUID(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid job ID: %w", err))
	}

	job, err := h.dataArchiveService.RetryFailedJob(ctx, id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&dataarchivepb.RetryArchivalJobResponse{
		Job: h.convertJobToProto(job),
	}), nil
}

// Placeholder implementations for other handlers

func (h *DataArchiveHandler) CreateLegalHold(ctx context.Context, req *connect.Request[dataarchivepb.CreateLegalHoldRequest]) (*connect.Response[dataarchivepb.CreateLegalHoldResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *DataArchiveHandler) GetLegalHold(ctx context.Context, req *connect.Request[dataarchivepb.GetLegalHoldRequest]) (*connect.Response[dataarchivepb.GetLegalHoldResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *DataArchiveHandler) UpdateLegalHold(ctx context.Context, req *connect.Request[dataarchivepb.UpdateLegalHoldRequest]) (*connect.Response[dataarchivepb.UpdateLegalHoldResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *DataArchiveHandler) DeleteLegalHold(ctx context.Context, req *connect.Request[dataarchivepb.DeleteLegalHoldRequest]) (*connect.Response[emptypb.Empty], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *DataArchiveHandler) ListLegalHolds(ctx context.Context, req *connect.Request[dataarchivepb.ListLegalHoldsRequest]) (*connect.Response[dataarchivepb.ListLegalHoldsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *DataArchiveHandler) CreateDataInventory(ctx context.Context, req *connect.Request[dataarchivepb.CreateDataInventoryRequest]) (*connect.Response[dataarchivepb.CreateDataInventoryResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *DataArchiveHandler) GetDataInventory(ctx context.Context, req *connect.Request[dataarchivepb.GetDataInventoryRequest]) (*connect.Response[dataarchivepb.GetDataInventoryResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *DataArchiveHandler) UpdateDataInventory(ctx context.Context, req *connect.Request[dataarchivepb.UpdateDataInventoryRequest]) (*connect.Response[dataarchivepb.UpdateDataInventoryResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *DataArchiveHandler) ListDataInventories(ctx context.Context, req *connect.Request[dataarchivepb.ListDataInventoriesRequest]) (*connect.Response[dataarchivepb.ListDataInventoriesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *DataArchiveHandler) ScanDataSources(ctx context.Context, req *connect.Request[dataarchivepb.ScanDataSourcesRequest]) (*connect.Response[dataarchivepb.ScanDataSourcesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *DataArchiveHandler) SearchArchivedData(ctx context.Context, req *connect.Request[dataarchivepb.SearchArchivedDataRequest]) (*connect.Response[dataarchivepb.SearchArchivedDataResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *DataArchiveHandler) RestoreArchivedData(ctx context.Context, req *connect.Request[dataarchivepb.RestoreArchivedDataRequest]) (*connect.Response[dataarchivepb.RestoreArchivedDataResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *DataArchiveHandler) GetArchivedDataInfo(ctx context.Context, req *connect.Request[dataarchivepb.GetArchivedDataInfoRequest]) (*connect.Response[dataarchivepb.GetArchivedDataInfoResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *DataArchiveHandler) DeleteArchivedData(ctx context.Context, req *connect.Request[dataarchivepb.DeleteArchivedDataRequest]) (*connect.Response[emptypb.Empty], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *DataArchiveHandler) CreateComplianceAudit(ctx context.Context, req *connect.Request[dataarchivepb.CreateComplianceAuditRequest]) (*connect.Response[dataarchivepb.CreateComplianceAuditResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *DataArchiveHandler) GetComplianceAudit(ctx context.Context, req *connect.Request[dataarchivepb.GetComplianceAuditRequest]) (*connect.Response[dataarchivepb.GetComplianceAuditResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *DataArchiveHandler) ListComplianceAudits(ctx context.Context, req *connect.Request[dataarchivepb.ListComplianceAuditsRequest]) (*connect.Response[dataarchivepb.ListComplianceAuditsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *DataArchiveHandler) GenerateComplianceReport(ctx context.Context, req *connect.Request[dataarchivepb.GenerateComplianceReportRequest]) (*connect.Response[dataarchivepb.GenerateComplianceReportResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *DataArchiveHandler) GetArchiveMetrics(ctx context.Context, req *connect.Request[dataarchivepb.GetArchiveMetricsRequest]) (*connect.Response[dataarchivepb.GetArchiveMetricsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *DataArchiveHandler) GetStorageSavings(ctx context.Context, req *connect.Request[dataarchivepb.GetStorageSavingsRequest]) (*connect.Response[dataarchivepb.GetStorageSavingsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *DataArchiveHandler) GetRetentionDashboard(ctx context.Context, req *connect.Request[dataarchivepb.GetRetentionDashboardRequest]) (*connect.Response[dataarchivepb.GetRetentionDashboardResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

// Helper conversion methods (stubs - would need full implementation)

func (h *DataArchiveHandler) parseUUID(s string) (uuid.UUID, error) {
	// Implementation would parse UUID from string
	return uuid.Parse(s)
}

func (h *DataArchiveHandler) convertProtoToCreatePolicyRequest(proto *dataarchivepb.RetentionPolicy) *interfaces.CreateRetentionPolicyRequest {
	// Implementation would convert protobuf to service request
	return nil
}

func (h *DataArchiveHandler) convertProtoToUpdatePolicyRequest(proto *dataarchivepb.RetentionPolicy) *interfaces.UpdateRetentionPolicyRequest {
	// Implementation would convert protobuf to service request
	return nil
}

func (h *DataArchiveHandler) convertPolicyToProto(policy *models.RetentionPolicy) *dataarchivepb.RetentionPolicy {
	// Implementation would convert model to protobuf
	return nil
}

func (h *DataArchiveHandler) convertJobToProto(job *models.ArchivalJob) *dataarchivepb.ArchivalJob {
	// Implementation would convert model to protobuf
	return nil
}

func (h *DataArchiveHandler) convertProtoDateRange(proto *dataarchivepb.DateRange) *models.DateRange {
	// Implementation would convert protobuf date range to model
	return nil
}