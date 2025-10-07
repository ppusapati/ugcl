package handlers

import (
	"context"
	"fmt"
	"log"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"

	approvalv2 "p9e.in/ugcl/formbuilder/api/v2/approval"
	"p9e.in/ugcl/formbuilder/services"
)

type ApprovalHandler struct {
	approvalService services.IApprovalService
}

func NewApprovalHandler(approvalService services.IApprovalService) *ApprovalHandler {
	return &ApprovalHandler{
		approvalService: approvalService,
	}
}

// StartApprovalProcess initiates a new approval workflow
func (h *ApprovalHandler) StartApprovalProcess(ctx context.Context, req *connect.Request[approvalv2.StartApprovalProcessRequest]) (*connect.Response[approvalv2.StartApprovalProcessResponse], error) {
	log.Printf("StartApprovalProcess request for entity: %s/%s", req.Msg.EntityType, req.Msg.EntityId)

	if req.Msg.EntityType == "" || req.Msg.EntityId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("entity_type and entity_id are required"))
	}

	// Convert proto request to service request
	serviceReq := &services.StartApprovalRequest{
		EntityType:     req.Msg.EntityType,
		EntityID:       req.Msg.EntityId,
		RequestedBy:    req.Msg.RequestedBy,
		Priority:       h.convertProtoPriority(req.Msg.Priority),
		RequestData:    h.convertProtoAnyToMap(req.Msg.RequestData),
		Comments:       req.Msg.Comments,
		AttachmentURLs: req.Msg.AttachmentUrls,
		Metadata:       h.convertProtoAnyToMap(req.Msg.Metadata),
	}

	instance, err := h.approvalService.StartApprovalProcess(ctx, serviceReq)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&approvalv2.StartApprovalProcessResponse{
		Instance: h.convertInstanceToProto(instance),
	}), nil
}

// ProcessApprovalAction processes an approval action
func (h *ApprovalHandler) ProcessApprovalAction(ctx context.Context, req *connect.Request[approvalv2.ProcessApprovalActionRequest]) (*connect.Response[approvalv2.ProcessApprovalActionResponse], error) {
	log.Printf("ProcessApprovalAction request for instance: %s, action: %v", req.Msg.InstanceId, req.Msg.Action)

	if req.Msg.InstanceId == "" || req.Msg.ApproverId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("instance_id and approver_id are required"))
	}

	// Convert proto request to service request
	serviceReq := &services.ProcessApprovalActionRequest{
		InstanceID:     h.parseUUID(req.Msg.InstanceId),
		ApproverID:     req.Msg.ApproverId,
		Action:         h.convertProtoAction(req.Msg.Action),
		Comments:       req.Msg.Comments,
		IPAddress:      req.Msg.IpAddress,
		UserAgent:      req.Msg.UserAgent,
		DelegatedFrom:  h.stringPtr(req.Msg.DelegatedFrom),
		DelegateTo:     h.stringPtr(req.Msg.DelegateTo),
		AttachmentURLs: req.Msg.AttachmentUrls,
		Metadata:       h.convertProtoAnyToMap(req.Msg.Metadata),
	}

	action, updatedInstance, err := h.approvalService.ProcessApprovalAction(ctx, serviceReq)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&approvalv2.ProcessApprovalActionResponse{
		Action:          h.convertActionToProto(action),
		UpdatedInstance: h.convertInstanceToProto(updatedInstance),
	}), nil
}

// GetApprovalInstance gets an approval instance by ID
func (h *ApprovalHandler) GetApprovalInstance(ctx context.Context, req *connect.Request[approvalv2.GetApprovalInstanceRequest]) (*connect.Response[approvalv2.GetApprovalInstanceResponse], error) {
	log.Printf("GetApprovalInstance request for ID: %s", req.Msg.Id)

	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("instance ID is required"))
	}

	// TODO: Implement GetApprovalInstance in service
	// For now, return a placeholder response
	return connect.NewResponse(&approvalv2.GetApprovalInstanceResponse{
		Instance: &approvalv2.ApprovalInstance{Id: req.Msg.Id},
		Actions:  []*approvalv2.ApprovalAction{},
	}), nil
}

// ListApprovalInstances lists approval instances with filters
func (h *ApprovalHandler) ListApprovalInstances(ctx context.Context, req *connect.Request[approvalv2.ListApprovalInstancesRequest]) (*connect.Response[approvalv2.ListApprovalInstancesResponse], error) {
	log.Printf("ListApprovalInstances request")

	// TODO: Implement ListApprovalInstances in service
	return connect.NewResponse(&approvalv2.ListApprovalInstancesResponse{
		Instances:     []*approvalv2.ApprovalInstance{},
		NextPageToken: "",
		TotalCount:    0,
	}), nil
}

// CancelApprovalInstance cancels an approval instance
func (h *ApprovalHandler) CancelApprovalInstance(ctx context.Context, req *connect.Request[approvalv2.CancelApprovalInstanceRequest]) (*connect.Response[emptypb.Empty], error) {
	log.Printf("CancelApprovalInstance request for ID: %s", req.Msg.Id)

	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("instance ID is required"))
	}

	// TODO: Implement CancelApprovalInstance in service
	return connect.NewResponse(&emptypb.Empty{}), nil
}

// CreateDelegate creates a new delegation relationship
func (h *ApprovalHandler) CreateDelegate(ctx context.Context, req *connect.Request[approvalv2.CreateDelegateRequest]) (*connect.Response[approvalv2.CreateDelegateResponse], error) {
	log.Printf("CreateDelegate request")

	if req.Msg.Delegate == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("delegate is required"))
	}

	// Convert proto to service request
	serviceReq := h.convertProtoToCreateDelegateRequest(req.Msg.Delegate)
	delegate, err := h.approvalService.CreateDelegate(ctx, serviceReq)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&approvalv2.CreateDelegateResponse{
		Delegate: h.convertDelegateToProto(delegate),
	}), nil
}

// UpdateDelegate updates a delegation relationship
func (h *ApprovalHandler) UpdateDelegate(ctx context.Context, req *connect.Request[approvalv2.UpdateDelegateRequest]) (*connect.Response[approvalv2.UpdateDelegateResponse], error) {
	log.Printf("UpdateDelegate request for ID: %s", req.Msg.Delegate.Id)

	if req.Msg.Delegate == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("delegate is required"))
	}

	// TODO: Implement UpdateDelegate
	return connect.NewResponse(&approvalv2.UpdateDelegateResponse{
		Delegate: req.Msg.Delegate,
	}), nil
}

// DeleteDelegate deletes a delegation relationship
func (h *ApprovalHandler) DeleteDelegate(ctx context.Context, req *connect.Request[approvalv2.DeleteDelegateRequest]) (*connect.Response[emptypb.Empty], error) {
	log.Printf("DeleteDelegate request for ID: %s", req.Msg.Id)

	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("delegate ID is required"))
	}

	err := h.approvalService.DeleteDelegate(ctx, h.parseUUID(req.Msg.Id))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

// ListDelegates lists delegation relationships
func (h *ApprovalHandler) ListDelegates(ctx context.Context, req *connect.Request[approvalv2.ListDelegatesRequest]) (*connect.Response[approvalv2.ListDelegatesResponse], error) {
	log.Printf("ListDelegates request")

	// TODO: Implement ListDelegates
	return connect.NewResponse(&approvalv2.ListDelegatesResponse{
		Delegates:     []*approvalv2.ApprovalDelegate{},
		NextPageToken: "",
		TotalCount:    0,
	}), nil
}

// GetApprovalMetrics gets approval metrics
func (h *ApprovalHandler) GetApprovalMetrics(ctx context.Context, req *connect.Request[approvalv2.GetApprovalMetricsRequest]) (*connect.Response[approvalv2.GetApprovalMetricsResponse], error) {
	log.Printf("GetApprovalMetrics request")

	// Convert proto request to service request
	serviceReq := &services.ApprovalMetricsRequest{
		EntityType: req.Msg.EntityType,
		WorkflowID: h.parseUUIDPtr(req.Msg.WorkflowId),
		StartDate:  req.Msg.StartDate.AsTime(),
		EndDate:    req.Msg.EndDate.AsTime(),
	}

	metrics, err := h.approvalService.GetApprovalMetrics(ctx, serviceReq)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&approvalv2.GetApprovalMetricsResponse{
		TotalRequests:          metrics.TotalRequests,
		ApprovedCount:          metrics.ApprovedCount,
		RejectedCount:          metrics.RejectedCount,
		PendingCount:           metrics.PendingCount,
		EscalatedCount:         metrics.EscalatedCount,
		AvgProcessingTimeHours: metrics.AvgProcessingTimeHours,
		SlaComplianceRate:      metrics.SLAComplianceRate,
		DailyMetrics:           []*approvalv2.MetricData{}, // TODO: Implement conversion
		Bottlenecks:            []*approvalv2.BottleneckInfo{}, // TODO: Implement conversion
	}), nil
}

// GetPendingApprovals gets pending approvals for a user
func (h *ApprovalHandler) GetPendingApprovals(ctx context.Context, req *connect.Request[approvalv2.GetPendingApprovalsRequest]) (*connect.Response[approvalv2.GetPendingApprovalsResponse], error) {
	log.Printf("GetPendingApprovals request for user: %s", req.Msg.UserId)

	if req.Msg.UserId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user ID is required"))
	}

	pendingApprovals, err := h.approvalService.GetPendingApprovalsForUser(ctx, req.Msg.UserId, req.Msg.PageSize, 0)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoApprovals := make([]*approvalv2.PendingApproval, len(pendingApprovals))
	for i, approval := range pendingApprovals {
		protoApprovals[i] = &approvalv2.PendingApproval{
			Instance:        h.convertInstanceToProto(approval.Instance),
			FormTitle:       approval.FormTitle,
			FormDescription: approval.FormDescription,
			// TODO: Add due date and overdue calculation
		}
	}

	return connect.NewResponse(&approvalv2.GetPendingApprovalsResponse{
		Approvals:     protoApprovals,
		NextPageToken: "",
		TotalCount:    int32(len(protoApprovals)),
	}), nil
}

// GetApprovalHistory gets approval history for an instance
func (h *ApprovalHandler) GetApprovalHistory(ctx context.Context, req *connect.Request[approvalv2.GetApprovalHistoryRequest]) (*connect.Response[approvalv2.GetApprovalHistoryResponse], error) {
	log.Printf("GetApprovalHistory request for instance: %s", req.Msg.InstanceId)

	if req.Msg.InstanceId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("instance ID is required"))
	}

	history, err := h.approvalService.GetApprovalHistory(ctx, h.parseUUID(req.Msg.InstanceId))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoActions := make([]*approvalv2.ApprovalAction, len(history))
	for i, entry := range history {
		protoActions[i] = h.convertActionToProto(entry.ApprovalAction)
	}

	return connect.NewResponse(&approvalv2.GetApprovalHistoryResponse{
		Actions:       protoActions,
		Escalations:   []*approvalv2.EscalationEvent{}, // TODO: Implement conversion
		Notifications: []*approvalv2.NotificationEvent{}, // TODO: Implement conversion
	}), nil
}

// ProcessEscalations processes escalations
func (h *ApprovalHandler) ProcessEscalations(ctx context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[approvalv2.ProcessEscalationsResponse], error) {
	log.Printf("ProcessEscalations request")

	processedCount, escalatedInstances, err := h.approvalService.ProcessEscalations(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&approvalv2.ProcessEscalationsResponse{
		ProcessedCount:     processedCount,
		EscalatedInstances: escalatedInstances,
	}), nil
}

// GetEscalatedApprovals gets escalated approvals
func (h *ApprovalHandler) GetEscalatedApprovals(ctx context.Context, req *connect.Request[approvalv2.GetEscalatedApprovalsRequest]) (*connect.Response[approvalv2.GetEscalatedApprovalsResponse], error) {
	log.Printf("GetEscalatedApprovals request")

	instances, err := h.approvalService.GetEscalatedInstances(ctx, req.Msg.PageSize, 0)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoInstances := make([]*approvalv2.ApprovalInstance, len(instances))
	for i, instance := range instances {
		protoInstances[i] = h.convertInstanceToProto(instance.Instance)
	}

	return connect.NewResponse(&approvalv2.GetEscalatedApprovalsResponse{
		Instances:     protoInstances,
		NextPageToken: "",
		TotalCount:    int32(len(protoInstances)),
	}), nil
}