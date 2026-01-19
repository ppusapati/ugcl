package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	pb "p9e.in/ugcl/identity/user/api/v2/permission"
	"p9e.in/ugcl/identity/user/mappers"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// UnifiedPermissionResolver implements IPermissionResolver with hierarchical resolution
type UnifiedPermissionResolver struct {
	permissionService  IPermissionService
	orgHierarchy       IOrganizationHierarchyProvider
	entityRoleProvider IEntityRoleProvider
	resourceOwnership  IResourceOwnershipProvider
}

// NewUnifiedPermissionResolver creates a new unified permission resolver
func NewUnifiedPermissionResolver(
	permissionService IPermissionService,
	orgHierarchy IOrganizationHierarchyProvider,
	entityRoleProvider IEntityRoleProvider,
	resourceOwnership IResourceOwnershipProvider,
) *UnifiedPermissionResolver {
	return &UnifiedPermissionResolver{
		permissionService:  permissionService,
		orgHierarchy:       orgHierarchy,
		entityRoleProvider: entityRoleProvider,
		resourceOwnership:  resourceOwnership,
	}
}

// ResolvePermission checks if a subject has permission with full context
func (r *UnifiedPermissionResolver) ResolvePermission(ctx context.Context, req *pb.CheckPermissionRequest) (*pb.CheckPermissionResponse, error) {
	// Determine check mode
	mode := req.Mode
	if mode == pb.PermissionCheckMode_PERMISSION_CHECK_MODE_UNSPECIFIED {
		mode = pb.PermissionCheckMode_PERMISSION_CHECK_MODE_WITH_INHERITANCE
	}

	// Build resolution path
	resolutionPath := []*pb.PermissionResolutionStep{}

	// Step 1: Check direct permissions
	effect, permission, path := r.checkDirectPermission(ctx, req)
	resolutionPath = append(resolutionPath, path...)
	if effect == pb.Effect_FORBIDDEN {
		return &pb.CheckPermissionResponse{
			Effect:            pb.Effect_FORBIDDEN,
			MatchedPermission: permission,
			ResolutionPath:    resolutionPath,
			Reason:            "Explicit FORBIDDEN permission found",
		}, nil
	}
	if effect == pb.Effect_GRANT {
		return &pb.CheckPermissionResponse{
			Effect:            pb.Effect_GRANT,
			MatchedPermission: permission,
			ResolutionPath:    resolutionPath,
			Reason:            "Direct permission granted",
		}, nil
	}

	// Step 2: Check entity-based permissions (if subject is entity:xxx)
	if strings.HasPrefix(req.Subject, "entity:") {
		entityID := strings.TrimPrefix(req.Subject, "entity:")
		effect, permission, path := r.checkEntityPermissions(ctx, entityID, req)
		resolutionPath = append(resolutionPath, path...)
		if effect == pb.Effect_FORBIDDEN {
			return &pb.CheckPermissionResponse{
				Effect:            pb.Effect_FORBIDDEN,
				MatchedPermission: permission,
				ResolutionPath:    resolutionPath,
				Reason:            "Entity permission denied",
			}, nil
		}
		if effect == pb.Effect_GRANT {
			return &pb.CheckPermissionResponse{
				Effect:            pb.Effect_GRANT,
				MatchedPermission: permission,
				ResolutionPath:    resolutionPath,
				Reason:            "Entity permission granted",
			}, nil
		}
	}

	// Step 3: Check role-based permissions
	effect, permission, path = r.checkRolePermissions(ctx, req)
	resolutionPath = append(resolutionPath, path...)
	if effect == pb.Effect_FORBIDDEN {
		return &pb.CheckPermissionResponse{
			Effect:            pb.Effect_FORBIDDEN,
			MatchedPermission: permission,
			ResolutionPath:    resolutionPath,
			Reason:            "Role permission denied",
		}, nil
	}
	if effect == pb.Effect_GRANT {
		return &pb.CheckPermissionResponse{
			Effect:            pb.Effect_GRANT,
			MatchedPermission: permission,
			ResolutionPath:    resolutionPath,
			Reason:            "Role permission granted",
		}, nil
	}

	// Step 4: Check organizational inheritance (if mode allows)
	if mode == pb.PermissionCheckMode_PERMISSION_CHECK_MODE_WITH_INHERITANCE {
		effect, permission, path := r.checkInheritedPermissions(ctx, req)
		resolutionPath = append(resolutionPath, path...)
		if effect == pb.Effect_FORBIDDEN {
			return &pb.CheckPermissionResponse{
				Effect:            pb.Effect_FORBIDDEN,
				MatchedPermission: permission,
				ResolutionPath:    resolutionPath,
				Reason:            "Inherited permission denied",
			}, nil
		}
		if effect == pb.Effect_GRANT {
			return &pb.CheckPermissionResponse{
				Effect:            pb.Effect_GRANT,
				MatchedPermission: permission,
				ResolutionPath:    resolutionPath,
				Reason:            "Inherited permission granted",
			}, nil
		}
	}

	// Step 5: Check resource ownership (if resource_id provided)
	if req.ResourceId != nil && req.ResourceId.Value != "" {
		effect, permission, path := r.checkResourceOwnership(ctx, req)
		resolutionPath = append(resolutionPath, path...)
		if effect == pb.Effect_GRANT {
			return &pb.CheckPermissionResponse{
				Effect:            pb.Effect_GRANT,
				MatchedPermission: permission,
				ResolutionPath:    resolutionPath,
				Reason:            "Resource ownership granted",
			}, nil
		}
	}

	// Default: no permission
	return &pb.CheckPermissionResponse{
		Effect:         pb.Effect_UNKNOWN,
		ResolutionPath: resolutionPath,
		Reason:         "No matching permission found",
	}, nil
}

// checkDirectPermission checks for direct permission assignment
func (r *UnifiedPermissionResolver) checkDirectPermission(ctx context.Context, req *pb.CheckPermissionRequest) (pb.Effect, *pb.Permission, []*pb.PermissionResolutionStep) {
	// Query permission service for exact match
	permissions, err := r.permissionService.GetPermissions(ctx, []string{req.Subject})
	if err != nil {
		return pb.Effect_UNKNOWN, nil, []*pb.PermissionResolutionStep{
			{
				Source:      "direct",
				Description: fmt.Sprintf("Error checking direct permissions: %v", err),
			},
		}
	}

	// Find matching permission
	for _, modelPerm := range permissions {
		protoPerm := mappers.PermissionModelToProto(modelPerm)
		if r.matchesPermission(protoPerm, req) && r.isTemporallyValid(protoPerm) {
			return protoPerm.Effect, protoPerm, []*pb.PermissionResolutionStep{
				{
					Source:      "direct",
					SourceId:    req.Subject,
					Scope:       "exact",
					Description: fmt.Sprintf("Direct permission on %s.%s.%s", req.Namespace, req.Resource, req.Action),
				},
			}
		}
	}

	return pb.Effect_UNKNOWN, nil, []*pb.PermissionResolutionStep{}
}

// checkEntityPermissions checks permissions via entity role bindings
func (r *UnifiedPermissionResolver) checkEntityPermissions(ctx context.Context, entityID string, req *pb.CheckPermissionRequest) (pb.Effect, *pb.Permission, []*pb.PermissionResolutionStep) {
	// Get entity role bindings
	bindings, err := r.entityRoleProvider.GetEntityRoleBindings(ctx, entityID, req.TenantId)
	if err != nil {
		return pb.Effect_UNKNOWN, nil, []*pb.PermissionResolutionStep{
			{
				Source:      "entity_role",
				Description: fmt.Sprintf("Error fetching entity roles: %v", err),
			},
		}
	}

	// Check each role binding
	for _, binding := range bindings {
		// Check temporal validity
		if !r.isBindingTemporallyValid(binding) {
			continue
		}

		// Check organizational scope match
		if !r.matchesOrganizationalScope(binding, req) {
			continue
		}

		// Get permissions for this role
		rolePerms, err := r.permissionService.ListPermissionsByRole(ctx, &pb.RoleIdentifier{
			Uuid: binding.RoleID,
		})
		if err != nil {
			continue
		}

		// Check each permission
		for _, perm := range rolePerms.Permissions {
			if r.matchesPermission(perm, req) && r.isTemporallyValid(perm) {
				return perm.Effect, perm, []*pb.PermissionResolutionStep{
					{
						Source:      "entity_role",
						SourceId:    binding.RoleID,
						Scope:       r.getScopeDescription(binding),
						Description: fmt.Sprintf("Permission via entity role '%s'", binding.RoleName),
					},
				}
			}
		}
	}

	return pb.Effect_UNKNOWN, nil, []*pb.PermissionResolutionStep{}
}

// checkRolePermissions checks permissions via role assignments
func (r *UnifiedPermissionResolver) checkRolePermissions(ctx context.Context, req *pb.CheckPermissionRequest) (pb.Effect, *pb.Permission, []*pb.PermissionResolutionStep) {
	// If subject is a role, check role permissions directly
	if strings.HasPrefix(req.Subject, "role:") {
		roleID := strings.TrimPrefix(req.Subject, "role:")
		rolePerms, err := r.permissionService.ListPermissionsByRole(ctx, &pb.RoleIdentifier{
			Uuid: roleID,
		})
		if err != nil {
			return pb.Effect_UNKNOWN, nil, []*pb.PermissionResolutionStep{}
		}

		for _, perm := range rolePerms.Permissions {
			if r.matchesPermission(perm, req) && r.isTemporallyValid(perm) {
				return perm.Effect, perm, []*pb.PermissionResolutionStep{
					{
						Source:      "role",
						SourceId:    roleID,
						Scope:       "role",
						Description: fmt.Sprintf("Permission via role %s", roleID),
					},
				}
			}
		}
	}

	return pb.Effect_UNKNOWN, nil, []*pb.PermissionResolutionStep{}
}

// checkInheritedPermissions checks permissions inherited from parent organizational units
func (r *UnifiedPermissionResolver) checkInheritedPermissions(ctx context.Context, req *pb.CheckPermissionRequest) (pb.Effect, *pb.Permission, []*pb.PermissionResolutionStep) {
	// Get subject (entity or user)
	subject := req.Subject
	if !strings.HasPrefix(subject, "entity:") {
		// For users, we need to get their entity first
		// This would require a user->entity lookup (not implemented in this stub)
		return pb.Effect_UNKNOWN, nil, []*pb.PermissionResolutionStep{}
	}

	entityID := strings.TrimPrefix(subject, "entity:")

	// Get entity info to know its organizational context
	_, err := r.entityRoleProvider.GetEntityInfo(ctx, entityID, req.TenantId)
	if err != nil {
		return pb.Effect_UNKNOWN, nil, []*pb.PermissionResolutionStep{}
	}

	// Build organizational hierarchy to check
	var scopesToCheck []struct {
		DivisionID   *string
		BranchID     *string
		DepartmentID *string
		Scope        string
	}

	// If request has department, check department -> branch -> division
	if req.DepartmentId != nil && req.DepartmentId.Value != "" {
		branchID, divisionID, err := r.orgHierarchy.GetDepartmentAncestors(ctx, req.DepartmentId.Value)
		if err == nil {
			scopesToCheck = append(scopesToCheck,
				struct {
					DivisionID   *string
					BranchID     *string
					DepartmentID *string
					Scope        string
				}{&divisionID, &branchID, nil, "branch"},
				struct {
					DivisionID   *string
					BranchID     *string
					DepartmentID *string
					Scope        string
				}{&divisionID, nil, nil, "division"},
			)
		}
	}

	// If request has branch, check branch -> division
	if req.BranchId != nil && req.BranchId.Value != "" {
		divisionID, err := r.orgHierarchy.GetBranchDivision(ctx, req.BranchId.Value)
		if err == nil {
			scopesToCheck = append(scopesToCheck,
				struct {
					DivisionID   *string
					BranchID     *string
					DepartmentID *string
					Scope        string
				}{&divisionID, nil, nil, "division"},
			)
		}
	}

	// Check permissions at each scope level
	for _, scope := range scopesToCheck {
		// Get entity role bindings at this scope
		bindings, err := r.entityRoleProvider.GetEntityRoleBindings(ctx, entityID, req.TenantId)
		if err != nil {
			continue
		}

		for _, binding := range bindings {
			// Check if binding matches this scope
			if !r.bindingMatchesScope(binding, scope.DivisionID, scope.BranchID, scope.DepartmentID) {
				continue
			}

			// Get role permissions
			rolePerms, err := r.permissionService.ListPermissionsByRole(ctx, &pb.RoleIdentifier{
				Uuid: binding.RoleID,
			})
			if err != nil {
				continue
			}

			for _, perm := range rolePerms.Permissions {
				if r.matchesPermission(perm, req) && r.isTemporallyValid(perm) && perm.AllowInheritance {
					return perm.Effect, perm, []*pb.PermissionResolutionStep{
						{
							Source:      "inherited",
							SourceId:    binding.RoleID,
							Scope:       scope.Scope,
							Description: fmt.Sprintf("Inherited from %s level via role '%s'", scope.Scope, binding.RoleName),
						},
					}
				}
			}
		}
	}

	return pb.Effect_UNKNOWN, nil, []*pb.PermissionResolutionStep{}
}

// checkResourceOwnership checks if the subject owns the resource
func (r *UnifiedPermissionResolver) checkResourceOwnership(ctx context.Context, req *pb.CheckPermissionRequest) (pb.Effect, *pb.Permission, []*pb.PermissionResolutionStep) {
	if req.ResourceId == nil || req.ResourceId.Value == "" {
		return pb.Effect_UNKNOWN, nil, []*pb.PermissionResolutionStep{}
	}

	// Extract entity ID from subject
	var entityID string
	if strings.HasPrefix(req.Subject, "entity:") {
		entityID = strings.TrimPrefix(req.Subject, "entity:")
	} else {
		// For users, would need user->entity lookup
		return pb.Effect_UNKNOWN, nil, []*pb.PermissionResolutionStep{}
	}

	// Check ownership
	isOwner, err := r.resourceOwnership.CheckResourceOwnership(ctx, entityID, req.Namespace, req.Resource, req.ResourceId.Value)
	if err != nil {
		return pb.Effect_UNKNOWN, nil, []*pb.PermissionResolutionStep{}
	}

	if isOwner {
		// Owners get full permissions
		perm := &pb.Permission{
			Namespace:  req.Namespace,
			Resource:   req.Resource,
			Action:     req.Action,
			Subject:    req.Subject,
			Effect:     pb.Effect_GRANT,
			TenantId:   req.TenantId,
			ResourceId: req.ResourceId,
		}
		return pb.Effect_GRANT, perm, []*pb.PermissionResolutionStep{
			{
				Source:      "owner",
				SourceId:    entityID,
				Scope:       "resource",
				Description: "Resource owner has full permissions",
			},
		}
	}

	// Check if resource is shared with entity
	isShared, permissions, err := r.resourceOwnership.CheckResourceSharing(ctx, entityID, req.Namespace, req.Resource, req.ResourceId.Value)
	if err != nil {
		return pb.Effect_UNKNOWN, nil, []*pb.PermissionResolutionStep{}
	}

	if isShared {
		// Check if the requested action is in the shared permissions
		for _, action := range permissions {
			if action == req.Action || action == "*" {
				perm := &pb.Permission{
					Namespace:  req.Namespace,
					Resource:   req.Resource,
					Action:     req.Action,
					Subject:    req.Subject,
					Effect:     pb.Effect_GRANT,
					TenantId:   req.TenantId,
					ResourceId: req.ResourceId,
				}
				return pb.Effect_GRANT, perm, []*pb.PermissionResolutionStep{
					{
						Source:      "shared",
						SourceId:    entityID,
						Scope:       "resource",
						Description: "Resource shared with entity",
					},
				}
			}
		}
	}

	return pb.Effect_UNKNOWN, nil, []*pb.PermissionResolutionStep{}
}

// Helper functions

func (r *UnifiedPermissionResolver) matchesPermission(perm *pb.Permission, req *pb.CheckPermissionRequest) bool {
	// Match namespace
	if perm.Namespace != req.Namespace && perm.Namespace != "*" {
		return false
	}

	// Match resource
	if perm.Resource != req.Resource && perm.Resource != "*" {
		return false
	}

	// Match action
	if perm.Action != req.Action && perm.Action != "*" {
		return false
	}

	// Match resource_id if specified in permission
	if perm.ResourceId != nil && perm.ResourceId.Value != "" {
		if req.ResourceId == nil || req.ResourceId.Value != perm.ResourceId.Value {
			return false
		}
	}

	// Match organizational scope
	if perm.DivisionId != nil && perm.DivisionId.Value != "" {
		if req.DivisionId == nil || req.DivisionId.Value != perm.DivisionId.Value {
			return false
		}
	}

	if perm.BranchId != nil && perm.BranchId.Value != "" {
		if req.BranchId == nil || req.BranchId.Value != perm.BranchId.Value {
			return false
		}
	}

	if perm.DepartmentId != nil && perm.DepartmentId.Value != "" {
		if req.DepartmentId == nil || req.DepartmentId.Value != perm.DepartmentId.Value {
			return false
		}
	}

	return true
}

func (r *UnifiedPermissionResolver) isTemporallyValid(perm *pb.Permission) bool {
	now := time.Now()

	if perm.ValidFrom != nil {
		if perm.ValidFrom.AsTime().After(now) {
			return false
		}
	}

	if perm.ValidUntil != nil {
		if perm.ValidUntil.AsTime().Before(now) {
			return false
		}
	}

	return true
}

func (r *UnifiedPermissionResolver) isBindingTemporallyValid(binding *EntityRoleBindingInfo) bool {
	if !binding.IsActive {
		return false
	}

	now := time.Now()

	if binding.ValidFrom != nil {
		if binding.ValidFrom.After(now) {
			return false
		}
	}

	if binding.ValidUntil != nil {
		if binding.ValidUntil.Before(now) {
			return false
		}
	}

	return true
}

func (r *UnifiedPermissionResolver) matchesOrganizationalScope(binding *EntityRoleBindingInfo, req *pb.CheckPermissionRequest) bool {
	// If binding has division constraint
	if binding.DivisionID != nil {
		if req.DivisionId == nil || req.DivisionId.Value != *binding.DivisionID {
			return false
		}
	}

	// If binding has branch constraint
	if binding.BranchID != nil {
		if req.BranchId == nil || req.BranchId.Value != *binding.BranchID {
			return false
		}
	}

	// If binding has department constraint
	if binding.DepartmentID != nil {
		if req.DepartmentId == nil || req.DepartmentId.Value != *binding.DepartmentID {
			return false
		}
	}

	return true
}

func (r *UnifiedPermissionResolver) bindingMatchesScope(binding *EntityRoleBindingInfo, divisionID, branchID, departmentID *string) bool {
	if divisionID != nil && binding.DivisionID != nil && *binding.DivisionID == *divisionID {
		return true
	}
	if branchID != nil && binding.BranchID != nil && *binding.BranchID == *branchID {
		return true
	}
	if departmentID != nil && binding.DepartmentID != nil && *binding.DepartmentID == *departmentID {
		return true
	}
	return false
}

func (r *UnifiedPermissionResolver) getScopeDescription(binding *EntityRoleBindingInfo) string {
	if binding.DepartmentID != nil {
		return "department"
	}
	if binding.BranchID != nil {
		return "branch"
	}
	if binding.DivisionID != nil {
		return "division"
	}
	return "tenant"
}

// ResolveEntityPermission is a convenience method for entity-based permission checks
func (r *UnifiedPermissionResolver) ResolveEntityPermission(ctx context.Context, entityID, namespace, resource, action string, opts *PermissionResolveOptions) (*pb.CheckPermissionResponse, error) {
	req := &pb.CheckPermissionRequest{
		Subject:   "entity:" + entityID,
		Namespace: namespace,
		Resource:  resource,
		Action:    action,
		TenantId:  opts.TenantID,
		Mode:      opts.CheckMode,
	}

	if opts.DivisionID != nil {
		req.DivisionId = wrapperspb.String(*opts.DivisionID)
	}
	if opts.BranchID != nil {
		req.BranchId = wrapperspb.String(*opts.BranchID)
	}
	if opts.DepartmentID != nil {
		req.DepartmentId = wrapperspb.String(*opts.DepartmentID)
	}
	if opts.ResourceID != nil {
		req.ResourceId = wrapperspb.String(*opts.ResourceID)
	}

	return r.ResolvePermission(ctx, req)
}

// GetEffectivePermissions returns all effective permissions for a subject
func (r *UnifiedPermissionResolver) GetEffectivePermissions(ctx context.Context, subject string, opts *PermissionResolveOptions) ([]*pb.Permission, error) {
	// Get direct permissions
	permissions, err := r.permissionService.GetPermissions(ctx, []string{subject})
	if err != nil {
		return nil, err
	}

	effectivePerms := make([]*pb.Permission, 0)

	// Filter by temporal validity
	for _, modelPerm := range permissions {
		protoPerm := mappers.PermissionModelToProto(modelPerm)
		if r.isTemporallyValid(protoPerm) {
			effectivePerms = append(effectivePerms, protoPerm)
		}
	}

	// If subject is entity, add permissions from entity role bindings
	if strings.HasPrefix(subject, "entity:") {
		entityID := strings.TrimPrefix(subject, "entity:")
		bindings, err := r.entityRoleProvider.GetEntityRoleBindings(ctx, entityID, opts.TenantID)
		if err == nil {
			for _, binding := range bindings {
				if !r.isBindingTemporallyValid(binding) {
					continue
				}

				rolePerms, err := r.permissionService.ListPermissionsByRole(ctx, &pb.RoleIdentifier{
					Uuid: binding.RoleID,
				})
				if err != nil {
					continue
				}

				for _, perm := range rolePerms.Permissions {
					if r.isTemporallyValid(perm) {
						// Annotate with binding scope (use proto.Clone to avoid copying lock)
						permCopy := &pb.Permission{
							Namespace:        perm.Namespace,
							Resource:         perm.Resource,
							Action:           perm.Action,
							Subject:          perm.Subject,
							Effect:           perm.Effect,
							TenantId:         perm.TenantId,
							ResourceId:       perm.ResourceId,
							AllowInheritance: perm.AllowInheritance,
						}
						if binding.DivisionID != nil {
							permCopy.DivisionId = wrapperspb.String(*binding.DivisionID)
						}
						if binding.BranchID != nil {
							permCopy.BranchId = wrapperspb.String(*binding.BranchID)
						}
						if binding.DepartmentID != nil {
							permCopy.DepartmentId = wrapperspb.String(*binding.DepartmentID)
						}
						if binding.ValidFrom != nil {
							permCopy.ValidFrom = timestamppb.New(*binding.ValidFrom)
						}
						if binding.ValidUntil != nil {
							permCopy.ValidUntil = timestamppb.New(*binding.ValidUntil)
						}
						effectivePerms = append(effectivePerms, permCopy)
					}
				}
			}
		}
	}

	return effectivePerms, nil
}
