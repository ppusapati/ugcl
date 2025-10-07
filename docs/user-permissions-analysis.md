# User Permissions & Roles Analysis

## Current State Review

### Existing Proto Files

1. **permission.proto** - Core permission management
2. **role.proto** - Role management with hierarchy
3. **permissiondef.proto** - Permission definitions/templates

---

## Analysis of Current Permission System

### ✅ What's Good (Already Sufficient)

#### 1. **Flexible Permission Model**
```protobuf
message Permission {
  string namespace = 1;   // ✅ Good: Supports multi-module permissions
  string resource = 2;    // ✅ Good: Resource-based access
  string action = 3;      // ✅ Good: CRUD + custom actions
  string subject = 4;     // ✅ Good: Can be user:id or role:id
  Effect effect = 5;      // ✅ Good: GRANT/FORBIDDEN (explicit deny)
  string tenant_id = 6;   // ✅ Good: Multi-tenant support
}
```

**Example Permissions:**
```
namespace: "organization"
resource: "division"
action: "create"
subject: "role:HR_Manager"
effect: GRANT
```

#### 2. **Hierarchical Roles**
```protobuf
message Role {
  string id = 1;
  string name = 2;
  string parent_id = 3;  // ✅ Good: Role inheritance
  bool is_preserved = 4; // ✅ Good: System roles protection
}
```

**Example Hierarchy:**
```
Admin (parent_id: null)
  └── Division_Manager (parent_id: Admin)
      └── Branch_Manager (parent_id: Division_Manager)
          └── Employee (parent_id: Branch_Manager)
```

#### 3. **Permission Definitions (Templates)**
```protobuf
message PermissionDef {
  string name = 1;
  string namespace = 2;
  string resource = 3;
  string action = 4;
  string scope = 5;        // ✅ Good: Wildcard support ("*" or specific)
  int32 version = 6;       // ✅ Good: Permission versioning
  string tenant_id = 7;    // ✅ Good: Tenant-specific permissions
  PermissionSide side = 9; // ✅ Good: HOST vs TENANT distinction
}
```

---

## ❌ What's Missing for Organizational Hierarchy

### 1. **Hierarchical Scope Context**

**Problem**: Current permissions don't have context for organizational hierarchy (Business → Division → Branch → Department).

**Current Limitation:**
```protobuf
// Cannot express: "User can manage employees in Branch X only"
Permission {
  namespace: "user"
  resource: "employee"
  action: "update"
  subject: "user:123"
  effect: GRANT
  // ❌ Missing: Which branch? Which division?
}
```

**Solution**: Need to add scope context to permissions.

---

### 2. **Scope-Based Permissions**

#### Recommended Addition to `permission.proto`:

```protobuf
message Permission {
  string namespace = 1;
  string resource = 2;
  string action = 3;
  string subject = 4;
  Effect effect = 5;
  string tenant_id = 6;
  string def_name = 7;

  // ✨ NEW: Scope Context for Organizational Hierarchy
  PermissionScope scope = 8;
}

message PermissionScope {
  ScopeLevel level = 1;           // BUSINESS, DIVISION, BRANCH, DEPARTMENT
  string division_id = 2;         // Specific division (if level >= DIVISION)
  string branch_id = 3;           // Specific branch (if level = BRANCH)
  string department_id = 4;       // Specific department (if level = DEPARTMENT)
  repeated string division_ids = 5;   // Multiple divisions
  repeated string branch_ids = 6;     // Multiple branches
  repeated string department_ids = 7; // Multiple departments
}

enum ScopeLevel {
  SCOPE_LEVEL_UNSPECIFIED = 0;
  BUSINESS = 1;        // Access across entire business/tenant
  DIVISION = 2;        // Access within specific division(s)
  BRANCH = 3;          // Access within specific branch(es)
  DEPARTMENT = 4;      // Access within specific department(s)
  CUSTOM = 5;          // Custom scope (using metadata)
}
```

#### Example Usage:

```protobuf
// Example 1: Division Manager - can manage all employees in Manufacturing Division
Permission {
  namespace: "user"
  resource: "employee"
  action: "update"
  subject: "role:Division_Manager"
  effect: GRANT
  tenant_id: "tenant-123"
  scope: {
    level: DIVISION
    division_id: "manufacturing-div-001"
  }
}

// Example 2: Branch Manager - can manage employees in Mumbai Branch only
Permission {
  namespace: "user"
  resource: "employee"
  action: "update"
  subject: "user:456"
  effect: GRANT
  tenant_id: "tenant-123"
  scope: {
    level: BRANCH
    branch_id: "mumbai-branch-001"
  }
}

// Example 3: Regional Manager - manages multiple branches
Permission {
  namespace: "user"
  resource: "employee"
  action: "view"
  subject: "user:789"
  effect: GRANT
  tenant_id: "tenant-123"
  scope: {
    level: BRANCH
    branch_ids: ["mumbai-001", "pune-001", "delhi-001"]
  }
}

// Example 4: HR General Manager - access to HR department across all divisions
Permission {
  namespace: "user"
  resource: "employee"
  action: "*"
  subject: "role:HR_GM"
  effect: GRANT
  tenant_id: "tenant-123"
  scope: {
    level: DEPARTMENT
    department_id: "hr-dept-001"  // Business-level HR department
  }
}
```

---

### 3. **User-Organization Assignments**

#### Need: Track which organizational units a user belongs to

**Current State**: `UserTenantRole` only tracks tenant-level roles:
```protobuf
message UserTenantRole {
  string tenant_id = 1;
  repeated string roles = 2;
  bool is_active = 3;
  google.protobuf.Timestamp assigned_at = 4;
}
```

#### Recommended Update to `user.proto`:

```protobuf
message UserTenantRole {
  string tenant_id = 1;
  repeated string roles = 2;
  bool is_active = 3;
  google.protobuf.Timestamp assigned_at = 4;

  // ✨ NEW: Organizational Context
  UserOrganizationalContext org_context = 5;
}

message UserOrganizationalContext {
  AssignmentLevel level = 1;
  string division_id = 2;
  string branch_id = 3;
  string department_id = 4;

  // For users with access to multiple organizational units
  repeated string division_ids = 5;
  repeated string branch_ids = 6;
  repeated string department_ids = 7;

  // Primary assignment (for employees)
  bool is_primary = 8;
}

enum AssignmentLevel {
  ASSIGNMENT_LEVEL_UNSPECIFIED = 0;
  BUSINESS_LEVEL = 1;
  DIVISION_LEVEL = 2;
  BRANCH_LEVEL = 3;
  DEPARTMENT_LEVEL = 4;
}
```

---

### 4. **Permission Resolution Algorithm**

#### Need: Logic to check if user has permission at a specific organizational level

```go
// Example permission check
func (s *PermissionService) CheckPermission(ctx context.Context, req *CheckPermissionRequest) (*CheckPermissionResponse, error) {
    // 1. Get user's organizational assignments
    userAssignments := s.getUserOrganizationalAssignments(ctx, req.UserId)

    // 2. Get user's roles
    userRoles := s.getUserRoles(ctx, req.UserId)

    // 3. Get all permissions for user + roles
    permissions := s.getPermissions(ctx, req.UserId, userRoles)

    // 4. Filter permissions by organizational scope
    for _, perm := range permissions {
        // Check if permission matches the requested action
        if !s.matchesAction(perm, req.Namespace, req.Resource, req.Action) {
            continue
        }

        // NEW: Check organizational scope
        if req.OrganizationalScope != nil {
            if !s.matchesOrgScope(perm.Scope, req.OrganizationalScope, userAssignments) {
                continue
            }
        }

        // Check effect (GRANT vs FORBIDDEN)
        if perm.Effect == FORBIDDEN {
            return &CheckPermissionResponse{Effect: FORBIDDEN}, nil
        }

        if perm.Effect == GRANT {
            return &CheckPermissionResponse{Effect: GRANT}, nil
        }
    }

    // Default deny
    return &CheckPermissionResponse{Effect: FORBIDDEN}, nil
}

func (s *PermissionService) matchesOrgScope(
    permScope *PermissionScope,
    requestScope *OrganizationalScope,
    userAssignments []*UserOrganizationalContext,
) bool {
    // If permission has no scope, it applies everywhere (business-level)
    if permScope == nil || permScope.Level == BUSINESS {
        return true
    }

    // Check if user is assigned to the required organizational unit
    switch permScope.Level {
    case DIVISION:
        return s.userHasDivisionAccess(userAssignments, permScope.DivisionId)
    case BRANCH:
        return s.userHasBranchAccess(userAssignments, permScope.BranchId)
    case DEPARTMENT:
        return s.userHasDepartmentAccess(userAssignments, permScope.DepartmentId)
    }

    return false
}
```

---

## Summary of Changes Needed

### ✅ Sufficient (No Changes Needed)
1. **Role hierarchy** - Already supports parent-child relationships
2. **Permission effects** - GRANT/FORBIDDEN works well
3. **Namespace/Resource/Action model** - Flexible enough
4. **Permission definitions** - Template system is good

### ❌ Need to Add

#### 1. **Update `permission.proto`**
```protobuf
// Add PermissionScope message
message PermissionScope {
  ScopeLevel level = 1;
  string division_id = 2;
  string branch_id = 3;
  string department_id = 4;
  repeated string division_ids = 5;
  repeated string branch_ids = 6;
  repeated string department_ids = 7;
}

// Add scope field to Permission
message Permission {
  // ... existing fields ...
  PermissionScope scope = 8;  // NEW
}
```

#### 2. **Update `user.proto`**
```protobuf
// Add organizational context to UserTenantRole
message UserTenantRole {
  // ... existing fields ...
  UserOrganizationalContext org_context = 5;  // NEW
}

message UserOrganizationalContext {
  AssignmentLevel level = 1;
  string division_id = 2;
  string branch_id = 3;
  string department_id = 4;
  repeated string division_ids = 5;
  repeated string branch_ids = 6;
  repeated string department_ids = 7;
  bool is_primary = 8;
}
```

#### 3. **New Service: User Assignment Service**
```protobuf
service UserAssignmentService {
  rpc AssignUserToOrganization(AssignUserRequest) returns (UserAssignment);
  rpc RemoveUserAssignment(RemoveAssignmentRequest) returns (google.protobuf.Empty);
  rpc GetUserAssignments(GetUserAssignmentsRequest) returns (GetUserAssignmentsResponse);
  rpc GetUsersInOrganizationalUnit(GetUsersInOrgUnitRequest) returns (GetUsersResponse);
}

message AssignUserRequest {
  string user_id = 1;
  string tenant_id = 2;
  AssignmentLevel level = 3;
  string division_id = 4;
  string branch_id = 5;
  string department_id = 6;
  repeated string roles = 7;
  bool is_primary = 8;
}
```

#### 4. **Database Changes**
```sql
-- Add organizational context to user_tenant_roles
ALTER TABLE user_tenant_roles
ADD COLUMN assignment_level VARCHAR(50),
ADD COLUMN division_id UUID REFERENCES divisions(id),
ADD COLUMN branch_id UUID REFERENCES branches(id),
ADD COLUMN department_id UUID REFERENCES departments(id),
ADD COLUMN is_primary BOOLEAN DEFAULT false;

-- Add scope to permissions table
ALTER TABLE permissions
ADD COLUMN scope_level VARCHAR(50),
ADD COLUMN scope_division_id UUID,
ADD COLUMN scope_branch_id UUID,
ADD COLUMN scope_department_id UUID;
```

---

## Implementation Priority

### Phase 1 (Week 1): Core Changes
- ✅ Update `permission.proto` with `PermissionScope`
- ✅ Update `user.proto` with `UserOrganizationalContext`
- ✅ Database migrations

### Phase 2 (Week 2): Service Layer
- ✅ Implement `UserAssignmentService`
- ✅ Update `PermissionService.CheckPermission()` with scope logic
- ✅ Update permission resolution algorithm

### Phase 3 (Week 3): Integration
- ✅ Integrate with organization module
- ✅ Update existing modules to use scoped permissions
- ✅ Testing

---

## Conclusion

**Is the current permission system sufficient?**

**Answer**: **80% sufficient**, but needs **organizational scope additions**.

**What to do**:
1. Add `PermissionScope` to permission system
2. Add `UserOrganizationalContext` to user assignments
3. Implement scope-aware permission checking
4. Create `UserAssignmentService` for managing org assignments

**Good news**: The foundation is solid. These are additions, not rewrites!
