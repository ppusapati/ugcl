package handlers

import (
	"context"
	"fmt"
	"log"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"

	backupdrpb "p9e.in/ugcl/backupdr/api/v1/backupdr"
	"p9e.in/ugcl/backupdr/models"
	"p9e.in/ugcl/backupdr/services/interfaces"
)

type BackupDRHandler struct {
	backupDRService interfaces.BackupDRService
}

func NewBackupDRHandler(backupDRService interfaces.BackupDRService) *BackupDRHandler {
	return &BackupDRHandler{
		backupDRService: backupDRService,
	}
}

// Backup Policy Management

func (h *BackupDRHandler) CreateBackupPolicy(ctx context.Context, req *connect.Request[backupdrpb.CreateBackupPolicyRequest]) (*connect.Response[backupdrpb.CreateBackupPolicyResponse], error) {
	log.Printf("CreateBackupPolicy request for policy: %s", req.Msg.Policy.Name)

	if req.Msg.Policy == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("policy is required"))
	}

	serviceReq := h.convertProtoToCreatePolicyRequest(req.Msg.Policy)
	policy, err := h.backupDRService.CreateBackupPolicy(ctx, serviceReq)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&backupdrpb.CreateBackupPolicyResponse{
		Policy: h.convertPolicyToProto(policy),
	}), nil
}

func (h *BackupDRHandler) GetBackupPolicy(ctx context.Context, req *connect.Request[backupdrpb.GetBackupPolicyRequest]) (*connect.Response[backupdrpb.GetBackupPolicyResponse], error) {
	log.Printf("GetBackupPolicy request for ID: %s", req.Msg.Id)

	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("policy ID is required"))
	}

	id, err := h.parseUUID(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid policy ID: %w", err))
	}

	policy, err := h.backupDRService.GetBackupPolicy(ctx, id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&backupdrpb.GetBackupPolicyResponse{
		Policy: h.convertPolicyToProto(policy),
	}), nil
}

func (h *BackupDRHandler) UpdateBackupPolicy(ctx context.Context, req *connect.Request[backupdrpb.UpdateBackupPolicyRequest]) (*connect.Response[backupdrpb.UpdateBackupPolicyResponse], error) {
	log.Printf("UpdateBackupPolicy request for policy: %s", req.Msg.Policy.Id)

	if req.Msg.Policy == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("policy is required"))
	}

	serviceReq := h.convertProtoToUpdatePolicyRequest(req.Msg.Policy)
	policy, err := h.backupDRService.UpdateBackupPolicy(ctx, serviceReq)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&backupdrpb.UpdateBackupPolicyResponse{
		Policy: h.convertPolicyToProto(policy),
	}), nil
}

func (h *BackupDRHandler) DeleteBackupPolicy(ctx context.Context, req *connect.Request[backupdrpb.DeleteBackupPolicyRequest]) (*connect.Response[emptypb.Empty], error) {
	log.Printf("DeleteBackupPolicy request for ID: %s", req.Msg.Id)

	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("policy ID is required"))
	}

	id, err := h.parseUUID(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid policy ID: %w", err))
	}

	err = h.backupDRService.DeleteBackupPolicy(ctx, id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *BackupDRHandler) ListBackupPolicies(ctx context.Context, req *connect.Request[backupdrpb.ListBackupPoliciesRequest]) (*connect.Response[backupdrpb.ListBackupPoliciesResponse], error) {
	log.Printf("ListBackupPolicies request")

	serviceReq := &interfaces.ListBackupPoliciesRequest{
		PageSize:   req.Msg.PageSize,
		PageToken:  req.Msg.PageToken,
		TargetType: h.convertProtoTargetType(req.Msg.TargetType),
		TargetName: req.Msg.TargetName,
		ActiveOnly: req.Msg.ActiveOnly,
	}

	policies, totalCount, nextPageToken, err := h.backupDRService.ListBackupPolicies(ctx, serviceReq)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoPolicies := make([]*backupdrpb.BackupPolicy, len(policies))
	for i, policy := range policies {
		protoPolicies[i] = h.convertPolicyToProto(policy)
	}

	return connect.NewResponse(&backupdrpb.ListBackupPoliciesResponse{
		Policies:      protoPolicies,
		NextPageToken: nextPageToken,
		TotalCount:    totalCount,
	}), nil
}

// Backup Execution

func (h *BackupDRHandler) ExecuteBackup(ctx context.Context, req *connect.Request[backupdrpb.ExecuteBackupRequest]) (*connect.Response[backupdrpb.ExecuteBackupResponse], error) {
	log.Printf("ExecuteBackup request for policy: %s", req.Msg.PolicyId)

	if req.Msg.PolicyId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("policy ID is required"))
	}

	policyID, err := h.parseUUID(req.Msg.PolicyId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid policy ID: %w", err))
	}

	serviceReq := &interfaces.ExecuteBackupRequest{
		PolicyID:   policyID,
		BackupType: h.convertProtoBackupType(req.Msg.BackupType),
		Priority:   h.convertProtoJobPriority(req.Msg.Priority),
	}

	job, err := h.backupDRService.ExecuteBackup(ctx, serviceReq)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&backupdrpb.ExecuteBackupResponse{
		Job: h.convertJobToProto(job),
	}), nil
}

func (h *BackupDRHandler) GetBackupJob(ctx context.Context, req *connect.Request[backupdrpb.GetBackupJobRequest]) (*connect.Response[backupdrpb.GetBackupJobResponse], error) {
	log.Printf("GetBackupJob request for ID: %s", req.Msg.Id)

	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("job ID is required"))
	}

	id, err := h.parseUUID(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid job ID: %w", err))
	}

	job, err := h.backupDRService.GetBackupJob(ctx, id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&backupdrpb.GetBackupJobResponse{
		Job: h.convertJobToProto(job),
	}), nil
}

func (h *BackupDRHandler) ListBackupJobs(ctx context.Context, req *connect.Request[backupdrpb.ListBackupJobsRequest]) (*connect.Response[backupdrpb.ListBackupJobsResponse], error) {
	log.Printf("ListBackupJobs request")

	serviceReq := &interfaces.ListBackupJobsRequest{
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

	if req.Msg.ScheduledAfter != nil {
		scheduledAfter := req.Msg.ScheduledAfter.AsTime()
		serviceReq.ScheduledAfter = &scheduledAfter
	}

	if req.Msg.ScheduledBefore != nil {
		scheduledBefore := req.Msg.ScheduledBefore.AsTime()
		serviceReq.ScheduledBefore = &scheduledBefore
	}

	jobs, totalCount, nextPageToken, err := h.backupDRService.ListBackupJobs(ctx, serviceReq)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoJobs := make([]*backupdrpb.BackupJob, len(jobs))
	for i, job := range jobs {
		protoJobs[i] = h.convertJobToProto(job)
	}

	return connect.NewResponse(&backupdrpb.ListBackupJobsResponse{
		Jobs:          protoJobs,
		NextPageToken: nextPageToken,
		TotalCount:    totalCount,
	}), nil
}

func (h *BackupDRHandler) CancelBackupJob(ctx context.Context, req *connect.Request[backupdrpb.CancelBackupJobRequest]) (*connect.Response[emptypb.Empty], error) {
	log.Printf("CancelBackupJob request for ID: %s", req.Msg.Id)

	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("job ID is required"))
	}

	id, err := h.parseUUID(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid job ID: %w", err))
	}

	err = h.backupDRService.CancelBackupJob(ctx, id, req.Msg.Reason)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *BackupDRHandler) RetryBackupJob(ctx context.Context, req *connect.Request[backupdrpb.RetryBackupJobRequest]) (*connect.Response[backupdrpb.RetryBackupJobResponse], error) {
	log.Printf("RetryBackupJob request for ID: %s", req.Msg.Id)

	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("job ID is required"))
	}

	id, err := h.parseUUID(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid job ID: %w", err))
	}

	job, err := h.backupDRService.RetryBackupJob(ctx, id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&backupdrpb.RetryBackupJobResponse{
		Job: h.convertJobToProto(job),
	}), nil
}

// Placeholder implementations for remaining handlers

func (h *BackupDRHandler) CreateRestoreRequest(ctx context.Context, req *connect.Request[backupdrpb.CreateRestoreRequestRequest]) (*connect.Response[backupdrpb.CreateRestoreRequestResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *BackupDRHandler) GetRestoreRequest(ctx context.Context, req *connect.Request[backupdrpb.GetRestoreRequestRequest]) (*connect.Response[backupdrpb.GetRestoreRequestResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *BackupDRHandler) ListRestoreRequests(ctx context.Context, req *connect.Request[backupdrpb.ListRestoreRequestsRequest]) (*connect.Response[backupdrpb.ListRestoreRequestsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *BackupDRHandler) ApproveRestoreRequest(ctx context.Context, req *connect.Request[backupdrpb.ApproveRestoreRequestRequest]) (*connect.Response[backupdrpb.ApproveRestoreRequestResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *BackupDRHandler) ExecuteRestore(ctx context.Context, req *connect.Request[backupdrpb.ExecuteRestoreRequest]) (*connect.Response[backupdrpb.ExecuteRestoreResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *BackupDRHandler) CreateDRPlan(ctx context.Context, req *connect.Request[backupdrpb.CreateDRPlanRequest]) (*connect.Response[backupdrpb.CreateDRPlanResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *BackupDRHandler) GetDRPlan(ctx context.Context, req *connect.Request[backupdrpb.GetDRPlanRequest]) (*connect.Response[backupdrpb.GetDRPlanResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *BackupDRHandler) UpdateDRPlan(ctx context.Context, req *connect.Request[backupdrpb.UpdateDRPlanRequest]) (*connect.Response[backupdrpb.UpdateDRPlanResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *BackupDRHandler) DeleteDRPlan(ctx context.Context, req *connect.Request[backupdrpb.DeleteDRPlanRequest]) (*connect.Response[emptypb.Empty], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *BackupDRHandler) ListDRPlans(ctx context.Context, req *connect.Request[backupdrpb.ListDRPlansRequest]) (*connect.Response[backupdrpb.ListDRPlansResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *BackupDRHandler) TestDRPlan(ctx context.Context, req *connect.Request[backupdrpb.TestDRPlanRequest]) (*connect.Response[backupdrpb.TestDRPlanResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *BackupDRHandler) ExecuteRecovery(ctx context.Context, req *connect.Request[backupdrpb.ExecuteRecoveryRequest]) (*connect.Response[backupdrpb.ExecuteRecoveryResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *BackupDRHandler) GetRecoveryExecution(ctx context.Context, req *connect.Request[backupdrpb.GetRecoveryExecutionRequest]) (*connect.Response[backupdrpb.GetRecoveryExecutionResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *BackupDRHandler) ListRecoveryExecutions(ctx context.Context, req *connect.Request[backupdrpb.ListRecoveryExecutionsRequest]) (*connect.Response[backupdrpb.ListRecoveryExecutionsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *BackupDRHandler) AbortRecovery(ctx context.Context, req *connect.Request[backupdrpb.AbortRecoveryRequest]) (*connect.Response[emptypb.Empty], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *BackupDRHandler) GetSystemHealth(ctx context.Context, req *connect.Request[backupdrpb.GetSystemHealthRequest]) (*connect.Response[backupdrpb.GetSystemHealthResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *BackupDRHandler) ListSystemHealth(ctx context.Context, req *connect.Request[backupdrpb.ListSystemHealthRequest]) (*connect.Response[backupdrpb.ListSystemHealthResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *BackupDRHandler) UpdateSystemHealth(ctx context.Context, req *connect.Request[backupdrpb.UpdateSystemHealthRequest]) (*connect.Response[backupdrpb.UpdateSystemHealthResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *BackupDRHandler) CreateStorageLocation(ctx context.Context, req *connect.Request[backupdrpb.CreateStorageLocationRequest]) (*connect.Response[backupdrpb.CreateStorageLocationResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *BackupDRHandler) GetStorageLocation(ctx context.Context, req *connect.Request[backupdrpb.GetStorageLocationRequest]) (*connect.Response[backupdrpb.GetStorageLocationResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *BackupDRHandler) ListStorageLocations(ctx context.Context, req *connect.Request[backupdrpb.ListStorageLocationsRequest]) (*connect.Response[backupdrpb.ListStorageLocationsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *BackupDRHandler) TestStorageConnection(ctx context.Context, req *connect.Request[backupdrpb.TestStorageConnectionRequest]) (*connect.Response[backupdrpb.TestStorageConnectionResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *BackupDRHandler) GetBackupMetrics(ctx context.Context, req *connect.Request[backupdrpb.GetBackupMetricsRequest]) (*connect.Response[backupdrpb.GetBackupMetricsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *BackupDRHandler) GetRecoveryMetrics(ctx context.Context, req *connect.Request[backupdrpb.GetRecoveryMetricsRequest]) (*connect.Response[backupdrpb.GetRecoveryMetricsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *BackupDRHandler) GetBackupDashboard(ctx context.Context, req *connect.Request[backupdrpb.GetBackupDashboardRequest]) (*connect.Response[backupdrpb.GetBackupDashboardResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *BackupDRHandler) GetOverdueBackups(ctx context.Context, req *connect.Request[backupdrpb.GetOverdueBackupsRequest]) (*connect.Response[backupdrpb.GetOverdueBackupsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

// Helper conversion methods (stubs - would need full implementation)

func (h *BackupDRHandler) parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

func (h *BackupDRHandler) convertProtoToCreatePolicyRequest(proto *backupdrpb.BackupPolicy) *interfaces.CreateBackupPolicyRequest {
	// Implementation would convert protobuf to service request
	return nil
}

func (h *BackupDRHandler) convertProtoToUpdatePolicyRequest(proto *backupdrpb.BackupPolicy) *interfaces.UpdateBackupPolicyRequest {
	// Implementation would convert protobuf to service request
	return nil
}

func (h *BackupDRHandler) convertPolicyToProto(policy *models.BackupPolicy) *backupdrpb.BackupPolicy {
	// Implementation would convert model to protobuf
	return nil
}

func (h *BackupDRHandler) convertJobToProto(job *models.BackupJob) *backupdrpb.BackupJob {
	// Implementation would convert model to protobuf
	return nil
}

func (h *BackupDRHandler) convertProtoTargetType(protoType backupdrpb.TargetType) string {
	switch protoType {
	case backupdrpb.TargetType_TARGET_TYPE_DATABASE:
		return "DATABASE"
	case backupdrpb.TargetType_TARGET_TYPE_FILES:
		return "FILES"
	case backupdrpb.TargetType_TARGET_TYPE_APPLICATION:
		return "APPLICATION"
	case backupdrpb.TargetType_TARGET_TYPE_SYSTEM:
		return "SYSTEM"
	case backupdrpb.TargetType_TARGET_TYPE_VOLUME:
		return "VOLUME"
	default:
		return ""
	}
}

func (h *BackupDRHandler) convertProtoBackupType(protoType backupdrpb.BackupType) string {
	switch protoType {
	case backupdrpb.BackupType_BACKUP_TYPE_FULL:
		return "FULL"
	case backupdrpb.BackupType_BACKUP_TYPE_INCREMENTAL:
		return "INCREMENTAL"
	case backupdrpb.BackupType_BACKUP_TYPE_DIFFERENTIAL:
		return "DIFFERENTIAL"
	case backupdrpb.BackupType_BACKUP_TYPE_SNAPSHOT:
		return "SNAPSHOT"
	case backupdrpb.BackupType_BACKUP_TYPE_CONTINUOUS:
		return "CONTINUOUS"
	default:
		return "FULL"
	}
}

func (h *BackupDRHandler) convertProtoJobPriority(protoPriority backupdrpb.JobPriority) string {
	switch protoPriority {
	case backupdrpb.JobPriority_JOB_PRIORITY_LOW:
		return "LOW"
	case backupdrpb.JobPriority_JOB_PRIORITY_NORMAL:
		return "NORMAL"
	case backupdrpb.JobPriority_JOB_PRIORITY_HIGH:
		return "HIGH"
	case backupdrpb.JobPriority_JOB_PRIORITY_CRITICAL:
		return "CRITICAL"
	default:
		return "NORMAL"
	}
}
