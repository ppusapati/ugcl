// =============================================================================
// internal/handlers/workflow_handler.go
// =============================================================================
package handlers

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"p9e.in/ugcl/core/middleware"
	pb "p9e.in/ugcl/formbuilder/api/v2/workflow"
	"p9e.in/ugcl/formbuilder/mappers"
	"p9e.in/ugcl/formbuilder/services"
)

// WorkflowHandler handles workflow Connect requests
type WorkflowHandler struct {
	workflowService services.IWorkflowService
}

// NewWorkflowHandler creates a new workflow handler
func NewWorkflowHandler(workflowService services.IWorkflowService) *WorkflowHandler {
	return &WorkflowHandler{
		workflowService: workflowService,
	}
}

// CreateWorkflow creates a new workflow definition
func (h *WorkflowHandler) CreateWorkflow(
	ctx context.Context,
	req *connect.Request[pb.CreateWorkflowRequest],
) (*connect.Response[pb.CreateWorkflowResponse], error) {

	if req.Msg.Workflow == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("workflow definition is required"))
	}
	wf, e := mappers.ProtoToWorkflow(req.Msg.Workflow)
	if e != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, e)
	}
	result, err := h.workflowService.CreateWorkflow(ctx, wf)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	resp := &pb.CreateWorkflowResponse{
		WorkflowId: result.ID.String(),
		Success:    true,
		Message:    "Workflow created successfully",
	}
	return connect.NewResponse(resp), nil
}

// GetWorkflow retrieves a workflow definition
func (h *WorkflowHandler) GetWorkflow(
	ctx context.Context,
	req *connect.Request[pb.GetWorkflowRequest],
) (*connect.Response[pb.GetWorkflowResponse], error) {

	if req.Msg.WorkflowId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("workflow ID is required"))
	}

	result, err := h.workflowService.GetWorkflow(ctx, req.Msg.WorkflowId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	wf, e := mappers.WorkflowToProto(result)
	if e != nil {
		return nil, connect.NewError(connect.CodeInternal, e)
	}
	resp := &pb.GetWorkflowResponse{
		Workflow: wf,
	}
	return connect.NewResponse(resp), nil
}

// UpdateWorkflow updates an existing workflow definition
func (h *WorkflowHandler) UpdateWorkflow(
	ctx context.Context,
	req *connect.Request[pb.CreateWorkflowRequest],
) (*connect.Response[pb.CreateWorkflowResponse], error) {

	if req.Msg.Workflow == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("workflow definition is required"))
	}
	wf, e := mappers.ProtoToWorkflow(req.Msg.Workflow)
	if e != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, e)
	}
	result, err := h.workflowService.UpdateWorkflow(ctx, wf)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	resp := &pb.CreateWorkflowResponse{
		WorkflowId: result.ID.String(),
		Success:    true,
		Message:    "Workflow updated successfully",
	}
	return connect.NewResponse(resp), nil
}

// DeleteWorkflow deletes a workflow definition
func (h *WorkflowHandler) DeleteWorkflow(
	ctx context.Context,
	req *connect.Request[pb.GetWorkflowRequest],
) (*connect.Response[pb.CreateWorkflowResponse], error) {

	if req.Msg.WorkflowId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("workflow ID is required"))
	}

	err := h.workflowService.DeleteWorkflow(ctx, req.Msg.WorkflowId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	resp := &pb.CreateWorkflowResponse{
		WorkflowId: req.Msg.WorkflowId,
		Success:    true,
		Message:    "Workflow deleted successfully",
	}
	return connect.NewResponse(resp), nil
}

// GetWorkflowStates retrieves workflow states for a form
func (h *WorkflowHandler) GetWorkflowStates(
	ctx context.Context,
	req *connect.Request[pb.GetFormRequest],
) (*connect.Response[pb.GetWorkflowStatesResponse], error) {
	if req.Msg.FormId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("form ID is required"))
	}

	result, err := h.workflowService.GetWorkflowStates(ctx, req.Msg.FormId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert states to proto using mapper
	protoStates := make([]*pb.WorkflowState, len(result.States))
	for i, state := range result.States {
		protoState, err := mappers.WorkflowStateToProto(state)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		protoStates[i] = protoState
	}

	response := &pb.GetWorkflowStatesResponse{
		States:           protoStates,
		CurrentState:     result.CurrentState,
		AvailableActions: result.AvailableActions,
	}

	return connect.NewResponse(response), nil
}

// TransitionWorkflow handles state transitions
func (h *WorkflowHandler) TransitionWorkflow(
	ctx context.Context,
	req *connect.Request[pb.TransitionRequest],
) (*connect.Response[pb.TransitionResponse], error) {

	claims := middleware.GetClaimsFromContext(ctx)
	if claims == nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("authentication required"))
	}

	userID := claims.UserID
	if req.Msg.InstanceId == "" || req.Msg.Event == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("instance ID and event are required"))
	}

	result, err := h.workflowService.TransitionWorkflow(ctx, req.Msg.InstanceId, req.Msg.Event, userID, req.Msg.Context)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert executed actions to proto
	protoActions := make([]*pb.TransitionAction, len(result.ExecutedActions))
	for i, action := range result.ExecutedActions {
		protoActions[i] = &pb.TransitionAction{
			Type:   action.Type,
			Params: action.Params,
		}
	}

	response := &pb.TransitionResponse{
		Success:         result.Success,
		NewState:        result.NewState,
		Message:         result.Message,
		ExecutedActions: protoActions,
	}

	return connect.NewResponse(response), nil
}

// TriggerExternalWorkflow triggers external workflow systems
func (h *WorkflowHandler) TriggerExternalWorkflow(
	ctx context.Context,
	req *connect.Request[pb.ExternalWorkflowRequest],
) (*connect.Response[pb.ExternalWorkflowResponse], error) {
	if req.Msg.InstanceId == "" || req.Msg.Action == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("instance ID and action are required"))
	}

	// Convert proto variables to map using helper
	variables, err := mappers.ProtoAnyToMapStringInterface(req.Msg.Variables)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid variables: %w", err))
	}

	result, err := h.workflowService.TriggerExternalWorkflow(ctx, req.Msg.InstanceId, req.Msg.Action, variables)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	response := &pb.ExternalWorkflowResponse{
		ProcessInstanceId: result.ProcessInstanceID,
		Success:           result.Success,
		Message:           result.Message,
	}

	return connect.NewResponse(response), nil
}
