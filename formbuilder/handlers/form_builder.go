package handlers

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "p9e.in/ugcl/formbuilder/api/v2/form_builder"
	pbw "p9e.in/ugcl/formbuilder/api/v2/workflow"
	"p9e.in/ugcl/formbuilder/mappers"
	"p9e.in/ugcl/formbuilder/services"
)

// FormBuilderHandler handles form builder Connect requests
type FormBuilderHandler struct {
	formService services.IFormBuilderService
}

// NewFormBuilderHandler creates a new form builder handler
func NewFormBuilderHandler(formService services.IFormBuilderService) *FormBuilderHandler {
	return &FormBuilderHandler{
		formService: formService,
	}
}

// CreateForm creates a new form definition
func (h *FormBuilderHandler) CreateForm(
	ctx context.Context,
	req *connect.Request[pb.CreateFormRequest],
) (*connect.Response[pb.CreateFormResponse], error) {
	if req.Msg.FormDefinition == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("form definition is required"))
	}
	fmt.Println("FormDefinition: ", req.Msg.FormDefinition.Metadata)
	fmt.Printf("Module field specifically: '%s'\n", req.Msg.FormDefinition.Metadata.Module)
	fmt.Printf("Module field length: %d\n", len(req.Msg.FormDefinition.Metadata.Module))
	// Create form through service
	result, err := h.formService.CreateForm(ctx, req.Msg.FormDefinition, req.Msg.CreateTable)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	response := &pb.CreateFormResponse{
		FormId:    result.FormID.String(),
		TableName: result.TableName,
		Success:   result.Success,
		Message:   result.Message,
	}

	return connect.NewResponse(response), nil
}

// GetForm retrieves a form definition
func (h *FormBuilderHandler) GetForm(
	ctx context.Context,
	req *connect.Request[pbw.GetFormRequest],
) (*connect.Response[pb.GetFormResponse], error) {
	if req.Msg.FormId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("form ID is required"))
	}

	var version *string
	if req.Msg.Version != "" {
		version = &req.Msg.Version
	}

	form, err := h.formService.GetForm(ctx, req.Msg.FormId, version)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	// Convert SQLC struct to proto using mapper
	formProto, err := mappers.FormDefinitionToProto(form)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	response := &pb.GetFormResponse{
		FormDefinition: formProto,
	}

	return connect.NewResponse(response), nil
}

// UpdateForm updates an existing form definition
func (h *FormBuilderHandler) UpdateForm(
	ctx context.Context,
	req *connect.Request[pb.FormDefinition],
) (*connect.Response[pb.CreateFormResponse], error) {
	if req.Msg == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("form definition is required"))
	}

	// Convert proto to SQLC struct using mapper
	form, err := mappers.ProtoToFormDefinition(req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Update form through service
	result, err := h.formService.UpdateForm(ctx, form)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	response := &pb.CreateFormResponse{
		FormId:  result.FormID.String(),
		Success: result.Success,
		Message: result.Message,
	}

	return connect.NewResponse(response), nil
}

// DeleteForm soft deletes a form definition
func (h *FormBuilderHandler) DeleteForm(
	ctx context.Context,
	req *connect.Request[pbw.GetFormRequest],
) (*connect.Response[pb.CreateFormResponse], error) {
	if req.Msg.FormId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("form ID is required"))
	}

	err := h.formService.DeleteForm(ctx, req.Msg.FormId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	response := &pb.CreateFormResponse{
		Success: true,
		Message: "Form deleted successfully",
	}

	return connect.NewResponse(response), nil
}

// ListForms retrieves a paginated list of forms
func (h *FormBuilderHandler) ListForms(
	ctx context.Context,
	req *connect.Request[pb.ListFormsRequest],
) (*connect.Response[pb.ListFormsResponse], error) {
	result, err := h.formService.ListForms(ctx, req.Msg.Page, req.Msg.PageSize, req.Msg.Filter, req.Msg.SortBy)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert forms to proto using mapper
	protoForms := make([]*pb.FormDefinition, len(result.Forms))
	for i, form := range result.Forms {
		protoForm, err := mappers.FormDefinitionToProto(form)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		protoForms[i] = protoForm
	}

	response := &pb.ListFormsResponse{
		Forms:    protoForms,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	}

	return connect.NewResponse(response), nil
}

// MigrateFormVersion migrates form data between versions
func (h *FormBuilderHandler) MigrateFormVersion(
	ctx context.Context,
	req *connect.Request[pb.MigrateFormRequest],
) (*connect.Response[pb.MigrateFormResponse], error) {
	if req.Msg.FormId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("form ID is required"))
	}

	result, err := h.formService.MigrateFormVersion(ctx, req.Msg.FormId, req.Msg.FromVersion, req.Msg.ToVersion, req.Msg.MigrateData)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	response := &pb.MigrateFormResponse{
		Success:         result.Success,
		RecordsMigrated: result.RecordsMigrated,
		Warnings:        result.Warnings,
	}

	return connect.NewResponse(response), nil
}

// GetFieldOptions retrieves dynamic field options
func (h *FormBuilderHandler) GetFieldOptions(
	ctx context.Context,
	req *connect.Request[pb.GetFieldOptionsRequest],
) (*connect.Response[pb.GetFieldOptionsResponse], error) {
	if req.Msg.FormId == "" || req.Msg.FieldId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("form ID and field ID are required"))
	}

	result, err := h.formService.GetFieldOptions(ctx, req.Msg.FormId, req.Msg.FieldId, req.Msg.Context)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert options to proto
	protoOptions := make([]*pb.FieldOption, len(result.Options))
	for i, option := range result.Options {
		protoOptions[i] = &pb.FieldOption{
			Value:    option.Value,
			Label:    option.Label,
			Disabled: option.Disabled,
			Metadata: option.Metadata,
		}
	}

	var cachedAt *timestamppb.Timestamp
	if result.CachedAt != nil {
		cachedAt = timestamppb.New(*result.CachedAt)
	}

	response := &pb.GetFieldOptionsResponse{
		Options:   protoOptions,
		FromCache: result.FromCache,
		CachedAt:  cachedAt,
	}

	return connect.NewResponse(response), nil
}

// RefreshFieldCache refreshes cached field options
func (h *FormBuilderHandler) RefreshFieldCache(
	ctx context.Context,
	req *connect.Request[pb.RefreshCacheRequest],
) (*connect.Response[pb.RefreshCacheResponse], error) {
	if req.Msg.FormId == "" || req.Msg.FieldId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("form ID and field ID are required"))
	}

	err := h.formService.RefreshFieldCache(ctx, req.Msg.FormId, req.Msg.FieldId)
	if err != nil {
		return connect.NewResponse(&pb.RefreshCacheResponse{
			Success: false,
			Message: err.Error(),
		}), nil
	}

	response := &pb.RefreshCacheResponse{
		Success: true,
		Message: "Cache refreshed successfully",
	}

	return connect.NewResponse(response), nil
}
