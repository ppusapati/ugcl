package handler

import (
	"context"

	"connectrpc.com/connect"

	tenantpb "p9e.in/ugcl/identity/tenant/api/v1/tenant"
	"p9e.in/ugcl/identity/tenant/api/v1/tenant/tenantconnect"
	"p9e.in/ugcl/identity/tenant/services"
)

// TenantInternalHandler implements the TenantInternalService ConnectRPC handlers
type TenantInternalHandler struct {
	tenantService services.ITenantService
}

// NewTenantInternalHandler creates a new tenant internal handler
func NewTenantInternalHandler(tenantService services.ITenantService) tenantconnect.TenantInternalServiceHandler {
	return &TenantInternalHandler{
		tenantService: tenantService,
	}
}

// GetTenant handles internal tenant retrieval for remote tenant store
// This is used by other microservices to fetch tenant information
func (h *TenantInternalHandler) GetTenant(
	ctx context.Context,
	req *connect.Request[tenantpb.GetTenantRequest],
) (*connect.Response[tenantpb.Tenant], error) {
	// Use the same service method as the public API but this could have
	// different authorization, rate limiting, or additional internal fields
	tenant, err := h.tenantService.GetTenantForInternalUse(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	resp := connect.NewResponse(tenant)
	return resp, nil
}