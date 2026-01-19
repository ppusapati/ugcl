package handler

import (
	"context"

	"connectrpc.com/connect"

	tenantpb "p9e.in/ugcl/identity/tenant/api/v1/tenant"
	"p9e.in/ugcl/identity/tenant/api/v1/tenant/tenantconnect"
	"p9e.in/ugcl/identity/tenant/services"
)

// TenantHandler implements the TenantService ConnectRPC handlers
type TenantHandler struct {
	tenantService services.ITenantService
}

// NewTenantHandler creates a new tenant handler
func NewTenantHandler(tenantService services.ITenantService) tenantconnect.TenantServiceHandler {
	return &TenantHandler{
		tenantService: tenantService,
	}
}

// CreateTenant handles tenant creation
func (h *TenantHandler) CreateTenant(
	ctx context.Context,
	req *connect.Request[tenantpb.CreateTenantRequest],
) (*connect.Response[tenantpb.Tenant], error) {
	tenant, err := h.tenantService.CreateTenant(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	resp := connect.NewResponse(tenant)
	return resp, nil
}

// UpdateTenant handles tenant updates
func (h *TenantHandler) UpdateTenant(
	ctx context.Context,
	req *connect.Request[tenantpb.UpdateTenantRequest],
) (*connect.Response[tenantpb.Tenant], error) {
	tenant, err := h.tenantService.UpdateTenant(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	resp := connect.NewResponse(tenant)
	return resp, nil
}

// DeleteTenant handles tenant deletion
func (h *TenantHandler) DeleteTenant(
	ctx context.Context,
	req *connect.Request[tenantpb.DeleteTenantRequest],
) (*connect.Response[tenantpb.DeleteTenantReply], error) {
	reply, err := h.tenantService.DeleteTenant(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	resp := connect.NewResponse(reply)
	return resp, nil
}

// GetTenant handles tenant retrieval by ID or name
func (h *TenantHandler) GetTenant(
	ctx context.Context,
	req *connect.Request[tenantpb.GetTenantRequest],
) (*connect.Response[tenantpb.Tenant], error) {
	tenant, err := h.tenantService.GetTenant(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	resp := connect.NewResponse(tenant)
	return resp, nil
}

// GetTenantPublic handles public tenant information retrieval
func (h *TenantHandler) GetTenantPublic(
	ctx context.Context,
	req *connect.Request[tenantpb.GetTenantPublicRequest],
) (*connect.Response[tenantpb.TenantInfo], error) {
	tenantInfo, err := h.tenantService.GetTenantPublic(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	resp := connect.NewResponse(tenantInfo)
	return resp, nil
}

// ListTenant handles tenant listing with filtering and pagination
func (h *TenantHandler) ListTenant(
	ctx context.Context,
	req *connect.Request[tenantpb.ListTenantRequest],
) (*connect.Response[tenantpb.ListTenantReply], error) {
	reply, err := h.tenantService.ListTenant(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	resp := connect.NewResponse(reply)
	return resp, nil
}

// GetCurrentTenant handles current tenant context retrieval
func (h *TenantHandler) GetCurrentTenant(
	ctx context.Context,
	req *connect.Request[tenantpb.GetCurrentTenantRequest],
) (*connect.Response[tenantpb.GetCurrentTenantReply], error) {
	reply, err := h.tenantService.GetCurrentTenant(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	resp := connect.NewResponse(reply)
	return resp, nil
}