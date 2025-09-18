package handler

import (
	"p9e.in/ugcl/identity/tenant/api/v1/tenant/tenantconnect"
	"p9e.in/ugcl/identity/tenant/services"
)

// HandlerContainer holds all tenant handlers
type HandlerContainer struct {
	TenantHandler         tenantconnect.TenantServiceHandler
	TenantInternalHandler tenantconnect.TenantInternalServiceHandler
}

// NewHandlerContainer creates a new handler container with all tenant handlers
func NewHandlerContainer(tenantService services.ITenantService) *HandlerContainer {
	return &HandlerContainer{
		TenantHandler:         NewTenantHandler(tenantService),
		TenantInternalHandler: NewTenantInternalHandler(tenantService),
	}
}

// ProvideHandlerContainer is a provider function for fx dependency injection
func ProvideHandlerContainer(tenantService services.ITenantService) *HandlerContainer {
	return NewHandlerContainer(tenantService)
}