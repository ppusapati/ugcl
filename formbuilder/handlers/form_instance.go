// =============================================================================
// internal/handlers/form_instance_handler.go
// =============================================================================
package handlers

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"p9e.in/ugcl/core/middleware"
	pb "p9e.in/ugcl/formbuilder/api/v2/form_instance"
	"p9e.in/ugcl/formbuilder/mappers"
	"p9e.in/ugcl/formbuilder/services"
)

// FormInstanceHandler handles form instance Connect requests
type FormInstanceHandler struct {
	instanceService services.IFormInstanceService
}

// NewFormInstanceHandler creates a new form instance handler
func NewFormInstanceHandler(instanceService services.IFormInstanceService) *FormInstanceHandler {
	return &FormInstanceHandler{
		instanceService: instanceService,
	}
}

// SubmitForm handles form submission
func (h *FormInstanceHandler) SubmitForm(
	ctx context.Context,
	req *connect.Request[pb.SubmitFormRequest],
) (*connect.Response[pb.SubmitFormResponse], error) {
	claims := middleware.GetClaimsFromContext(ctx)
	if claims == nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("authentication required"))
	}

	userID := claims.UserID

	if req.Msg.FormId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("form ID is required"))
	}

	// Convert proto field values to map using helper
	fieldValues, err := mappers.ProtoAnyToMapStringInterface(req.Msg.FieldValues)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid field values: %w", err))
	}

	// Submit form through service
	result, err := h.instanceService.SubmitForm(ctx, req.Msg.FormId, fieldValues, req.Msg.Action, userID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert validation errors to proto
	protoErrors := make([]*pb.ValidationError, len(result.Errors))
	for i, validationErr := range result.Errors {
		protoErrors[i] = &pb.ValidationError{

			Field:   validationErr.Field,
			Message: validationErr.Message,
			Code:    validationErr.Code,
		}
	}

	response := &pb.SubmitFormResponse{
		InstanceId:   result.InstanceID.String(),
		CurrentState: result.CurrentState,
		Success:      result.Success,
		Errors:       protoErrors,
	}

	return connect.NewResponse(response), nil
}

// GetFormInstance retrieves a form instance with related data
func (h *FormInstanceHandler) GetFormInstance(
	ctx context.Context,
	req *connect.Request[pb.GetFormInstanceRequest],
) (*connect.Response[pb.FormInstance], error) {

	if req.Msg.InstanceId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("instance ID is required"))
	}

	// Get instance with related data through service
	instanceData, err := h.instanceService.GetFormInstanceWithRelated(ctx, req.Msg.InstanceId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	// Convert main instance to proto
	protoInstance, err := mappers.FormInstanceToProto(instanceData.Instance)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert audit logs to proto
	protoAudits := make([]*pb.AuditLog, len(instanceData.AuditLogs))
	for i, audit := range instanceData.AuditLogs {
		protoAudit, err := mappers.AuditLogToProto(audit)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		protoAudits[i] = protoAudit
	}

	// Convert attachments to proto
	protoAttachments := make([]*pb.Attachment, len(instanceData.Attachments))
	for i, attachment := range instanceData.Attachments {
		protoAttachment, err := mappers.AttachmentToProto(attachment)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		protoAttachments[i] = protoAttachment
	}

	// Convert comments to proto
	protoComments := make([]*pb.Comment, len(instanceData.Comments))
	for i, comment := range instanceData.Comments {
		protoComment, err := mappers.CommentToProto(comment)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		protoComments[i] = protoComment
	}

	// Populate related data in proto instance
	protoInstance.AuditLogs = protoAudits
	protoInstance.Attachments = protoAttachments
	protoInstance.Comments = protoComments

	return connect.NewResponse(protoInstance), nil
}

// UpdateFormInstance updates an existing form instance
func (h *FormInstanceHandler) UpdateFormInstance(
	ctx context.Context,
	req *connect.Request[pb.UpdateFormInstanceRequest],
) (*connect.Response[pb.SubmitFormResponse], error) {
	claims := middleware.GetClaimsFromContext(ctx)
	if claims == nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("authentication required"))
	}

	userID := claims.UserID
	if req.Msg.InstanceId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("instance ID is required"))
	}

	// Convert proto field values to map using helper
	fieldValues, err := mappers.ProtoAnyToMapStringInterface(req.Msg.FieldValues)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid field values: %w", err))
	}

	// Update form instance through service
	result, err := h.instanceService.UpdateFormInstance(ctx, req.Msg.InstanceId, fieldValues, req.Msg.Action, userID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert validation errors to proto
	protoErrors := make([]*pb.ValidationError, len(result.Errors))
	for i, validationErr := range result.Errors {
		protoErrors[i] = &pb.ValidationError{
			Field:   validationErr.Field,
			Message: validationErr.Message,
			Code:    validationErr.Code,
		}
	}

	response := &pb.SubmitFormResponse{
		InstanceId:   result.InstanceID.String(),
		CurrentState: result.CurrentState,
		Success:      result.Success,
		Errors:       protoErrors,
	}

	return connect.NewResponse(response), nil
}
