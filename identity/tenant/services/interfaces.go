package services

import (
	"context"

	tenantpb "p9e.in/ugcl/identity/tenant/api/v1/tenant"
)

// ITenantService defines the business logic interface for tenant operations
type ITenantService interface {
	// Public tenant operations
	CreateTenant(ctx context.Context, req *tenantpb.CreateTenantRequest) (*tenantpb.Tenant, error)
	UpdateTenant(ctx context.Context, req *tenantpb.UpdateTenantRequest) (*tenantpb.Tenant, error)
	DeleteTenant(ctx context.Context, req *tenantpb.DeleteTenantRequest) (*tenantpb.DeleteTenantReply, error)
	GetTenant(ctx context.Context, req *tenantpb.GetTenantRequest) (*tenantpb.Tenant, error)
	GetTenantPublic(ctx context.Context, req *tenantpb.GetTenantPublicRequest) (*tenantpb.TenantInfo, error)
	ListTenant(ctx context.Context, req *tenantpb.ListTenantRequest) (*tenantpb.ListTenantReply, error)
	GetCurrentTenant(ctx context.Context, req *tenantpb.GetCurrentTenantRequest) (*tenantpb.GetCurrentTenantReply, error)

	// Internal tenant operations (for inter-service communication)
	GetTenantForInternalUse(ctx context.Context, req *tenantpb.GetTenantRequest) (*tenantpb.Tenant, error)
}