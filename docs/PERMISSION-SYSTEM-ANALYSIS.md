# Permission System Analysis

**Date:** 2025-10-05
**Status:** Analysis Complete
**Next Steps:** Design unified permission system with organizational scope

---

## Executive Summary

This document analyzes the current permission architecture across auth, user, tenant, and entity modules. The analysis identifies **5 critical gaps** requiring immediate attention and proposes a unified permission system that supports organizational hierarchy, entity-based access control, and temporal constraints.

### Key Findings

✅ **Strengths:**
- Strong foundation with permission definitions (PermissionDef)
- Flexible subject-based permissions (user/role)
- Entity-role bindings with temporal constraints
- Multi-tenant isolation built-in

❌ **Critical Gaps:**
1. **No organizational scope** in permission checks (division/branch/department)
2. **No entity-based permissions** in existing permission service
3. **Disconnected systems** (EntityRoleBinding vs Permission service)
4. **Missing permission inheritance** from organizational hierarchy
5. **No unified permission resolution** strategy

---

## Current Architecture Analysis

### 1. Auth Module (`identity/auth/proto/auth.proto`)

**Purpose:** Authentication, token management, session management

#### Key Components

**UserSession** (Lines 134-147):
```protobuf
message UserSession {
  string user_id = 1;
  repeated string roles = 5;
  repeated string permissions = 6;  // ⚠️ Flat list, no scope
  optional string tenant_id = 7;
  // ❌ MISSING: division_id, branch_id, department_id, entity_id
}
```

**ValidateTokenRequest** (Lines 100-104):
```protobuf
message ValidateTokenRequest {
  string token = 1;
  optional string required_permission = 2;  // ⚠️ Simple string check
  optional string resource = 3;             // ⚠️ No organizational scope
}
```

**ValidateTokenResponse** (Lines 106-112):
```protobuf
message ValidateTokenResponse {
  repeated string permissions = 4;  // ⚠️ Flat list
  optional string tenant_id = 5;    // ✅ Tenant isolation
  // ❌ MISSING: organizational context, entity info
}
```

#### Findings

✅ **Strengths:**
- Strong session management with device tracking
- Token validation with basic permission checks
- Multi-tenant support in sessions

❌ **Gaps:**
- **No organizational context** in UserSession (missing division/branch/department)
- **No entity information** in token validation
- **Simple string-based permission checks** without scope
- **No hierarchical permission resolution**

---

### 2. User Module - Permission Service

#### 2.1 Permission Service (`identity/user/proto/permission.proto`)

**Purpose:** Runtime permission assignment and checking

**Permission Model** (Lines 66-74):
```protobuf
message Permission {
  string namespace = 1;    // e.g., "dms", "formbuilder"
  string resource = 2;     // e.g., "document", "form"
  string action = 3;       // e.g., "read", "write", "delete"
  string subject = 4;      // e.g., "user:123", "role:admin"
  Effect effect = 5;       // GRANT or FORBIDDEN
  string tenant_id = 6;    // ✅ Tenant isolation
  string def_name = 7;     // Link to PermissionDef
  // ❌ MISSING: division_id, branch_id, department_id, entity_id
}
```

**CheckPermissionRequest** (Lines 47-52):
```protobuf
message CheckPermissionRequest {
  string subject = 1;      // "user:123" or "role:admin"
  string namespace = 2;
  string resource = 3;
  string action = 4;
  // ❌ MISSING: organizational scope, resource instance ID
}
```

#### Findings

✅ **Strengths:**
- Clean namespace/resource/action hierarchy
- Subject-based permissions (user or role)
- Effect-based grants (GRANT/FORBIDDEN)
- Link to permission definitions

❌ **Gaps:**
- **No organizational scope fields** (division/branch/department)
- **No entity-based subjects** (only user/role supported)
- **No resource instance scoping** (can't say "document:abc123")
- **No hierarchical scope resolution** (e.g., division-level access)

---

#### 2.2 Permission Definition Service (`identity/user/proto/permissiondef.proto`)

**Purpose:** Define reusable permission templates

**PermissionDef** (Lines 38-48):
```protobuf
message PermissionDef {
  string name = 1;         // Primary key, e.g., "dms.document.read"
  string namespace = 2;
  string resource = 3;
  string action = 4;
  string scope = 5;        // ⚠️ Pattern-based scope ("*" or pattern)
  int32 version = 6;
  string tenant_id = 7;    // ✅ Tenant-specific definitions
  string description = 8;
  PermissionSide side = 9; // HOST_ONLY, TENANT_ONLY, BOTH
}
```

**PermissionSide Enum** (Lines 72-76):
```protobuf
enum PermissionSide {
  BOTH = 0;
  HOST_ONLY = 1;      // ✅ Host tenant permissions
  TENANT_ONLY = 2;    // ✅ Regular tenant permissions
}
```

#### Findings

✅ **Strengths:**
- Reusable permission templates
- Version tracking for permission evolution
- Host/tenant permission separation
- Scope patterns (though limited)

❌ **Gaps:**
- **Scope is pattern-based string**, not structured organizational hierarchy
- **No built-in organizational scope resolution**
- **No entity type support** in definitions

---

#### 2.3 Role Service (`identity/user/proto/role.proto`)

**Purpose:** Role management and role-permission binding

**Role** (Lines 38-44):
```protobuf
message Role {
  string id = 1;
  string name = 2;
  string parent_id = 3;     // ✅ Role hierarchy support
  bool is_preserved = 4;
  google.protobuf.Struct metadata = 5;
  // ❌ MISSING: organizational scope (role at division/branch/dept level)
}
```

**RoleWithPermissions** (Lines 70-73):
```protobuf
message RoleWithPermissions {
  Role role = 1;
  repeated Permission permissions = 2;  // ✅ Links to permissions
}
```

#### Findings

✅ **Strengths:**
- Role hierarchy with parent_id
- Role-permission association
- Metadata for extensibility

❌ **Gaps:**
- **Roles are global**, not scoped to organizational units
- **No way to say** "Manager role at Division A"
- **No temporal constraints** on role assignments (handled in EntityRoleBinding)

---

#### 2.4 User Service (`identity/user/proto/user.proto`)

**Purpose:** User management and user-role binding

**User** (Lines 45-68):
```protobuf
message User {
  repeated string tenant_ids = 10;       // ✅ Multi-tenant users
  repeated UserTenantRole tenant_roles = 11;  // ✅ Per-tenant roles
  // ❌ MISSING: entity_id linkage, organizational scope
}
```

**UserTenantRole** (Lines 70-75):
```protobuf
message UserTenantRole {
  string tenant_id = 1;
  repeated string roles = 2;
  bool is_active = 3;
  google.protobuf.Timestamp assigned_at = 4;
  // ❌ MISSING: division_id, branch_id, department_id scope
}
```

#### Findings

✅ **Strengths:**
- Multi-tenant user support
- Per-tenant role assignments
- Role activation tracking

❌ **Gaps:**
- **UserTenantRole lacks organizational scope**
- **No link to Entity system** (user_id should map to entity_id)
- **No temporal constraints** on role validity

---

### 3. Tenant Module (`identity/tenant/proto/tenant.proto`)

**Purpose:** Tenant management and isolation

**Tenant** (Lines 104-114):
```protobuf
message Tenant {
  string id = 1;
  string name = 2;
  string region = 4;
  TenantDB tenant_db = 8;  // SHARED/SEPERATEDB/SEPERATESCHEMA
  // ❌ MISSING: No permission configuration
}
```

#### Findings

✅ **Strengths:**
- Multi-tenant isolation strategy (shared/separate DB)
- Regional tenants
- Extensible features

❌ **Gaps:**
- **No tenant-level permission configuration**
- **No default roles/permissions** on tenant creation
- **No organizational structure** (divisions/branches managed separately in organization module)

---

### 4. Entity Module (`identity/entity/proto/entity.proto`)

**Purpose:** Unified identity abstraction for all actors

**Entity** (Lines 47-69):
```protobuf
message Entity {
  string id = 1;
  string tenant_id = 2;
  EntityType entity_type = 3;   // ✅ Employee, Contractor, Device, etc.
  string user_id = 4;            // ✅ Links to user
  string reference_id = 5;       // ✅ Points to domain record
  ReferenceSource reference_source = 6;
  EntityStatus status = 7;

  // ✅ ORGANIZATIONAL CONTEXT
  google.protobuf.StringValue division_id = 8;
  google.protobuf.StringValue branch_id = 9;
  google.protobuf.StringValue department_id = 10;

  map<string, string> metadata = 11;
}
```

**EntityRoleBinding** (Lines 72-92):
```protobuf
message EntityRoleBinding {
  string entity_id = 3;
  string role_id = 4;

  // ✅ OPTIONAL SCOPE LIMITATIONS
  google.protobuf.StringValue division_id = 5;
  google.protobuf.StringValue branch_id = 6;
  google.protobuf.StringValue department_id = 7;

  // ✅ TIME-BOUND ACCESS
  google.protobuf.Timestamp valid_from = 8;
  google.protobuf.Timestamp valid_until = 9;
}
```

**EntityType Enum** (Lines 11-23):
```protobuf
enum EntityType {
  ENTITY_TYPE_EMPLOYEE = 1;
  ENTITY_TYPE_CONTRACTOR = 2;
  ENTITY_TYPE_VENDOR = 3;
  ENTITY_TYPE_CLIENT = 4;
  ENTITY_TYPE_DEVICE = 5;
  ENTITY_TYPE_DRONE = 6;
  ENTITY_TYPE_BOT = 7;
  ENTITY_TYPE_SYSTEM = 8;
  ENTITY_TYPE_ADMIN = 9;
  ENTITY_TYPE_AGENT = 10;
}
```

#### Findings

✅ **Strengths:**
- **Complete organizational context** (division/branch/department)
- **Entity-role bindings with scope limitations**
- **Temporal constraints** (valid_from/valid_until)
- **Polymorphic entity types** (employee, device, bot, etc.)
- **Reference to domain records** (employee_id, device_id, etc.)

❌ **Gaps:**
- **EntityRoleBinding is disconnected from Permission service**
- **No way to check permissions using entity_id**
- **No inheritance from organizational hierarchy**

---

## Critical Integration Gaps

### Gap 1: Disconnected Permission Systems

**Problem:** Two parallel permission systems exist:

1. **Permission Service** (identity/user)
   - Subject = "user:123" or "role:admin"
   - No organizational scope
   - No entity support

2. **EntityRoleBinding** (identity/entity)
   - Entity → Role assignments
   - Organizational scope (division/branch/department)
   - Temporal constraints
   - **Not integrated with Permission.Check()**

**Impact:** Cannot check if "entity:abc can read document scoped to Division A"

---

### Gap 2: No Organizational Scope in Permission Checks

**Problem:** `CheckPermissionRequest` has no organizational scope:

```protobuf
message CheckPermissionRequest {
  string subject = 1;
  string namespace = 2;
  string resource = 3;
  string action = 4;
  // ❌ MISSING: division_id, branch_id, department_id, resource_id
}
```

**Impact:**
- Cannot check division-level permissions
- Cannot implement hierarchical access (division → branch → department)
- Cannot scope permissions to specific organizational units

---

### Gap 3: No Entity-Based Permission Subjects

**Problem:** Permission subjects only support "user:123" and "role:admin", not "entity:abc"

**Impact:**
- Devices, bots, contractors cannot have direct permissions
- Must map entity → user → role → permissions (complex)
- Cannot differentiate between user acting as employee vs contractor

---

### Gap 4: No Resource Instance Scoping

**Problem:** Permissions apply to resource types, not instances:

```protobuf
// Current: Can check "user:123 can read document"
// ❌ CANNOT: Check "user:123 can read document:xyz123"
```

**Impact:**
- Cannot implement owner-based permissions
- Cannot implement shared document permissions
- Cannot implement organizational document access

---

### Gap 5: No Permission Inheritance

**Problem:** No hierarchical permission resolution:

- Division-level role should grant permissions to all branches
- Branch-level role should grant permissions to all departments
- No automatic inheritance mechanism

**Impact:**
- Must manually assign permissions at each level
- Difficult to manage large organizational hierarchies
- Inconsistent access control

---

## Proposed Unified Permission System

### Design Principles

1. **Entity-First:** All actors are entities (users become entities)
2. **Organizational Scope:** All permissions scoped to division/branch/department
3. **Hierarchical Inheritance:** Division permissions inherit to branch/department
4. **Resource Instance Support:** Permissions on specific resource instances
5. **Temporal Constraints:** Time-bound permissions from EntityRoleBinding
6. **Backward Compatible:** Existing Permission service enhanced, not replaced

---

### Enhanced Permission Model

#### 1. Enhanced Permission Message

```protobuf
message Permission {
  string namespace = 1;
  string resource = 2;
  string action = 3;
  string subject = 4;  // Enhanced: "entity:abc", "user:123", "role:admin"
  Effect effect = 5;
  string tenant_id = 6;
  string def_name = 7;

  // NEW: Organizational scope
  google.protobuf.StringValue division_id = 8;
  google.protobuf.StringValue branch_id = 9;
  google.protobuf.StringValue department_id = 10;

  // NEW: Resource instance scoping
  google.protobuf.StringValue resource_id = 11;  // Specific resource instance

  // NEW: Temporal constraints
  google.protobuf.Timestamp valid_from = 12;
  google.protobuf.Timestamp valid_until = 13;

  // NEW: Inheritance control
  bool allow_inheritance = 14;  // Allow child org units to inherit
}
```

#### 2. Enhanced CheckPermissionRequest

```protobuf
message CheckPermissionRequest {
  string subject = 1;  // "entity:abc", "user:123", "role:admin"
  string namespace = 2;
  string resource = 3;
  string action = 4;

  // NEW: Organizational context
  google.protobuf.StringValue division_id = 5;
  google.protobuf.StringValue branch_id = 6;
  google.protobuf.StringValue department_id = 7;

  // NEW: Resource instance
  google.protobuf.StringValue resource_id = 8;

  // NEW: Check mode
  PermissionCheckMode mode = 9;  // EXACT, WITH_INHERITANCE, ANY
}

enum PermissionCheckMode {
  PERMISSION_CHECK_MODE_EXACT = 0;          // Exact match only
  PERMISSION_CHECK_MODE_WITH_INHERITANCE = 1; // Check with org hierarchy
  PERMISSION_CHECK_MODE_ANY = 2;            // Any matching permission
}
```

#### 3. Enhanced CheckPermissionResponse

```protobuf
message CheckPermissionResponse {
  Effect effect = 1;

  // NEW: Matched permission details
  Permission matched_permission = 2;

  // NEW: Resolution path
  repeated PermissionResolutionStep resolution_path = 3;
}

message PermissionResolutionStep {
  string source = 1;  // "direct", "role", "entity_role", "inherited"
  string source_id = 2;  // ID of the source (role_id, entity_id, etc.)
  string scope = 3;  // "division", "branch", "department", "resource"
}
```

---

### Permission Resolution Algorithm

#### Step 1: Direct Entity Permissions
```
1. Check if entity has direct permission for (namespace, resource, action, resource_id)
2. If FORBIDDEN → return FORBIDDEN (explicit deny wins)
3. If GRANT → continue to check scope
```

#### Step 2: Entity Role Bindings
```
1. Get all EntityRoleBindings for entity
2. Filter by:
   - valid_from <= now <= valid_until
   - organizational scope matches or is parent
3. Get roles from bindings
4. Check role permissions
```

#### Step 3: Organizational Inheritance
```
1. If request has department_id:
   - Check department-level permissions
   - Check branch-level permissions (parent)
   - Check division-level permissions (grandparent)
2. If request has branch_id:
   - Check branch-level permissions
   - Check division-level permissions (parent)
3. If request has division_id:
   - Check division-level permissions
```

#### Step 4: Resource Ownership
```
1. If resource_id provided:
   - Check if entity owns resource
   - Check if resource shared with entity
   - Check organizational scope of resource
```

#### Step 5: Final Resolution
```
1. Collect all matching permissions
2. Apply effect precedence: FORBIDDEN > GRANT
3. Return most specific match
4. Return resolution path for audit
```

---

### Implementation Roadmap

#### Phase 1: Proto Updates (P2 Tasks 16-18)

**Task 16:** Update Permission proto
- Add organizational scope fields
- Add resource_id field
- Add temporal constraint fields
- Add inheritance control flag

**Task 17:** Update CheckPermissionRequest/Response
- Add organizational context
- Add resource instance scoping
- Add check mode enum
- Add resolution path response

**Task 18:** Update UserSession in auth proto
- Add entity_id field
- Add division_id, branch_id, department_id
- Add organizational context

#### Phase 2: Service Implementation (P2 Tasks 19-20)

**Task 19:** Implement unified permission resolution service
- EntityPermissionResolver with hierarchical logic
- Integration with EntityRoleBinding
- Temporal constraint checking
- Organizational inheritance

**Task 20:** Update existing services to use new permission checks
- Update middleware to populate organizational context
- Update DMS to check document-level permissions
- Update FormBuilder to check form-level permissions
- Update all service authorization logic

---

## Migration Strategy

### Backward Compatibility

1. **Existing Permission checks continue to work**
   - Subject "user:123" still supported
   - No organizational scope = tenant-wide access
   - Existing PermissionDef remain valid

2. **Gradual Enhancement**
   - Services can adopt organizational scope incrementally
   - Old-style permissions coexist with new-style
   - No breaking changes to existing APIs

### Migration Steps

1. **Update protos** (non-breaking: all new fields optional)
2. **Implement new permission resolver** (parallel to existing)
3. **Update auth middleware** to populate entity/org context
4. **Migrate high-priority services** (DMS, FormBuilder)
5. **Migrate remaining services**
6. **Deprecate old-style permission checks** (6+ months)

---

## Security Considerations

### 1. Principle of Least Privilege
- Default to no access
- Explicit grants required
- FORBIDDEN always wins over GRANT

### 2. Temporal Constraints
- Automatically revoke expired permissions
- Background job to clean up expired bindings
- Audit log for permission expiration

### 3. Organizational Isolation
- Division-level permissions don't cross division boundaries
- Inheritance only flows downward (division → branch → dept)
- No cross-organizational access without explicit grants

### 4. Audit Trail
- Log all permission checks with resolution path
- Track who granted permissions
- Alert on permission changes for sensitive resources

---

## Database Schema Impact

### New Indexes Required

```sql
-- Permission table
CREATE INDEX idx_permissions_entity_org
  ON permissions(tenant_id, subject, division_id, branch_id, department_id);

CREATE INDEX idx_permissions_resource_instance
  ON permissions(tenant_id, namespace, resource, resource_id);

CREATE INDEX idx_permissions_temporal
  ON permissions(tenant_id, valid_from, valid_until)
  WHERE valid_until IS NOT NULL;

-- EntityRoleBinding already has good indexes
-- Organization hierarchy queries (handled by organization module)
```

---

## Example Use Cases

### Use Case 1: Division Manager Access

**Scenario:** Manager at Division A should access all documents in Division A

**Solution:**
```
1. Entity for Manager links to Division A
2. EntityRoleBinding: entity → "Manager" role, scoped to division_id
3. Permission: "Manager" role → GRANT dms.document.read with allow_inheritance=true
4. Check: entity can read document in Branch B of Division A (inherited)
```

### Use Case 2: Shared Document

**Scenario:** Document shared with specific entity

**Solution:**
```
1. Document owned by Entity A in Division A
2. Create Permission: entity_id=B → GRANT dms.document.read, resource_id=doc123
3. Entity B can read doc123 even if in different division
4. Check includes resource_id in query
```

### Use Case 3: Temporary Contractor Access

**Scenario:** Contractor needs access for 3 months

**Solution:**
```
1. Entity for Contractor (entity_type=CONTRACTOR)
2. EntityRoleBinding: entity → "Contractor" role
   - valid_from = start_date
   - valid_until = start_date + 90 days
   - scoped to specific department
3. Permissions auto-expire after 90 days
```

---

## Next Steps

1. ✅ **Complete this analysis** (DONE)
2. 🔄 **Review with team** (get approval for design)
3. ⏳ **Update proto files** (P2 tasks 16-18)
4. ⏳ **Implement permission resolver** (P2 task 19)
5. ⏳ **Migrate services** (P2 task 20)

---

## Appendix: Reference Files

- `identity/auth/proto/auth.proto` - Authentication and sessions
- `identity/user/proto/permission.proto` - Permission service
- `identity/user/proto/permissiondef.proto` - Permission definitions
- `identity/user/proto/role.proto` - Role management
- `identity/user/proto/user.proto` - User management
- `identity/tenant/proto/tenant.proto` - Tenant management
- `identity/entity/proto/entity.proto` - Entity and role bindings

---

**Document Version:** 1.0
**Last Updated:** 2025-10-05
**Author:** Claude Code (AI Assistant)
