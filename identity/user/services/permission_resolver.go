package services

import (
	"context"
	"time"

	pb "p9e.in/ugcl/identity/user/api/v2/permission"
)

// IPermissionResolver provides unified permission resolution with organizational hierarchy support
type IPermissionResolver interface {
	// ResolvePermission checks if a subject has permission with full context
	ResolvePermission(ctx context.Context, req *pb.CheckPermissionRequest) (*pb.CheckPermissionResponse, error)

	// ResolveEntityPermission checks permissions for an entity with organizational scope
	ResolveEntityPermission(ctx context.Context, entityID, namespace, resource, action string, opts *PermissionResolveOptions) (*pb.CheckPermissionResponse, error)

	// GetEffectivePermissions returns all effective permissions for a subject
	GetEffectivePermissions(ctx context.Context, subject string, opts *PermissionResolveOptions) ([]*pb.Permission, error)
}

// PermissionResolveOptions provides additional context for permission resolution
type PermissionResolveOptions struct {
	TenantID     string
	DivisionID   *string
	BranchID     *string
	DepartmentID *string
	ResourceID   *string
	CheckMode    pb.PermissionCheckMode
	Timestamp    *time.Time // For checking temporal constraints
}

// IOrganizationHierarchyProvider provides organizational hierarchy information
type IOrganizationHierarchyProvider interface {
	// GetDivisionAncestors returns all ancestor divisions (parent, grandparent, etc.)
	GetDivisionAncestors(ctx context.Context, divisionID string) ([]string, error)

	// GetBranchDivision returns the division ID for a branch
	GetBranchDivision(ctx context.Context, branchID string) (string, error)

	// GetDepartmentBranch returns the branch ID for a department
	GetDepartmentBranch(ctx context.Context, departmentID string) (string, error)

	// GetDepartmentAncestors returns branch and division IDs for a department
	GetDepartmentAncestors(ctx context.Context, departmentID string) (branchID string, divisionID string, err error)
}

// IEntityRoleProvider provides entity role binding information
type IEntityRoleProvider interface {
	// GetEntityRoleBindings returns all role bindings for an entity
	GetEntityRoleBindings(ctx context.Context, entityID, tenantID string) ([]*EntityRoleBindingInfo, error)

	// GetEntityInfo returns entity information including organizational context
	GetEntityInfo(ctx context.Context, entityID, tenantID string) (*EntityInfo, error)
}

// EntityRoleBindingInfo represents an entity's role binding with scope
type EntityRoleBindingInfo struct {
	RoleID       string
	RoleName     string
	DivisionID   *string
	BranchID     *string
	DepartmentID *string
	ValidFrom    *time.Time
	ValidUntil   *time.Time
	IsActive     bool
}

// EntityInfo represents entity information with organizational context
type EntityInfo struct {
	EntityID     string
	EntityType   string
	UserID       string
	ReferenceID  string
	DivisionID   *string
	BranchID     *string
	DepartmentID *string
	Status       string
}

// IResourceOwnershipProvider checks resource ownership and sharing
type IResourceOwnershipProvider interface {
	// CheckResourceOwnership checks if an entity owns a resource
	CheckResourceOwnership(ctx context.Context, entityID, namespace, resource, resourceID string) (bool, error)

	// CheckResourceSharing checks if a resource is shared with an entity
	CheckResourceSharing(ctx context.Context, entityID, namespace, resource, resourceID string) (bool, []string, error)

	// GetResourceOrganizationalScope returns the organizational context of a resource
	GetResourceOrganizationalScope(ctx context.Context, namespace, resource, resourceID string) (*OrganizationalScope, error)
}

// OrganizationalScope represents the organizational context of a resource
type OrganizationalScope struct {
	TenantID     string
	DivisionID   *string
	BranchID     *string
	DepartmentID *string
	OwnerID      *string // Entity ID of the owner
}
