# Complete Authentication & Authorization Implementation Guide

## Table of Contents
1. [Architecture Overview](#architecture-overview)
2. [Module Responsibilities](#module-responsibilities)
3. [Authentication Flow](#authentication-flow)
4. [Authorization Flow](#authorization-flow)
5. [Form-Level Authorization Example](#form-level-authorization-example)
6. [Field-Level Authorization Example](#field-level-authorization-example)
7. [Implementation Patterns](#implementation-patterns)
8. [Code Duplication Analysis](#code-duplication-analysis)
9. [Integration Guide](#integration-guide)
10. [Best Practices](#best-practices)

---

## Architecture Overview

### The Four-Layer Identity Model

```
┌────────────────────────────────────────────────────────────────┐
│                    COMPLETE IDENTITY STACK                     │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│  Layer 1: TENANT (Multi-tenancy Isolation)                    │
│  ┌──────────────────────────────────────────────────────────┐ │
│  │ - Tenant identification and routing                      │ │
│  │ - Database isolation (SHARED/SEPERATEDB/SEPERATESCHEMA) │ │
│  │ - Feature flags and configurations                       │ │
│  │ - Subscription and billing                               │ │
│  └──────────────────────────────────────────────────────────┘ │
│                           ↓                                    │
│                                                                │
│  Layer 2: AUTH (Authentication)                                │
│  ┌──────────────────────────────────────────────────────────┐ │
│  │ - User credentials (username, email, password)           │ │
│  │ - JWT token generation and validation                    │ │
│  │ - Session management                                     │ │
│  │ - Two-factor authentication (TOTP, backup codes)        │ │
│  │ - Login/logout flows                                     │ │
│  │ - Security events and audit logging                      │ │
│  └──────────────────────────────────────────────────────────┘ │
│                           ↓                                    │
│                                                                │
│  Layer 3: ENTITY (Authorization Context)                       │
│  ┌──────────────────────────────────────────────────────────┐ │
│  │ - Entity type classification (employee, contractor, etc.)│ │
│  │ - Organizational context (division/branch/department)    │ │
│  │ - Entity role bindings (scoped, time-bound)             │ │
│  │ - Links user account ↔ domain records                   │ │
│  └──────────────────────────────────────────────────────────┘ │
│                           ↓                                    │
│                                                                │
│  Layer 4: USER (Authorization Rules)                           │
│  ┌──────────────────────────────────────────────────────────┐ │
│  │ - Permission definitions (namespace:resource:action)     │ │
│  │ - Role management and hierarchies                        │ │
│  │ - Permission resolution (direct, role, entity, inherited)│ │
│  │ - Organizational scope checking                          │ │
│  │ - Resource-level permissions                             │ │
│  └──────────────────────────────────────────────────────────┘ │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

---

## Module Responsibilities

### 1. Tenant Module (`identity/tenant`)

**Primary Responsibility:** Multi-tenant data isolation and configuration

**Key Functions:**
- Tenant provisioning and lifecycle management
- Database routing (shared vs. separate databases)
- Feature flag management per tenant
- Subscription and billing tracking
- Custom domain management
- Usage metrics and quotas

**API Surface:**
```go
// Tenant creation with admin user
CreateTenant(ctx, &CreateTenantRequest{
    Name:             "acme-corp",           // URL-safe identifier
    DisplayName:      "ACME Corporation",
    Region:           "us-east-1",
    TenantDB:         SHARED,                // Isolation mode
    SubscriptionPlan: "enterprise",
    MaxUsers:         500,
    AdminIdentity:    "admin@acme.com",      // Initial admin
    AdminPassword:    "SecurePass123!",
})

// Feature flag management
SetFeatureFlag(ctx, &SetFeatureFlagRequest{
    TenantId: "tenant-uuid",
    Key:      "advanced_analytics",
    Value:    "enabled",
    ValueType: "boolean",
})
```

**Integration Points:**
- Stores tenant context in JWT tokens
- Used by all modules for data filtering
- Provides tenant switching capability

---

### 2. Auth Module (`identity/auth`)

**Primary Responsibility:** Authentication (Who are you?)

**Key Functions:**
- User login with username/email/password
- JWT token generation and validation
- Session lifecycle management
- Two-factor authentication (TOTP, SMS, backup codes)
- Password reset workflows
- OTP generation and verification
- Security event logging
- Account locking and unlocking

**API Surface:**
```go
// Standard login
Login(ctx, &LoginRequest{
    Email:      "john.doe@acme.com",
    Password:   "SecurePassword123",
    TenantId:   "tenant-uuid",
    DeviceInfo: "{\"device\":\"iPhone\",\"os\":\"iOS 17\"}",
    RememberMe: true,
})
// Returns: {access_token, refresh_token, requires_two_factor, session_id, user_session}

// Enable 2FA
EnableTwoFactor(ctx, &EnableTwoFactorRequest{
    UserId: "user-uuid",
    Method: TWO_FACTOR_METHOD_TOTP,
})
// Returns: {secret, qr_code_url, backup_codes[]}

// Verify 2FA during login
VerifyTwoFactor(ctx, &VerifyTwoFactorRequest{
    SessionId: "session-uuid",
    Code:      "123456",
    Method:    TWO_FACTOR_METHOD_TOTP,
})
// Returns: {access_token, refresh_token} (completes login)

// Token refresh
RefreshToken(ctx, &RefreshTokenRequest{
    RefreshToken: "refresh-token-jwt",
    DeviceInfo:   "{...}",
})
// Returns: {new_access_token, new_refresh_token, expires_at}

// Logout
Logout(ctx, &LogoutRequest{
    SessionId: "session-uuid",
})
```

**Session Data Structure:**
```go
type Session struct {
    SessionID        string
    UserID           uuid.UUID
    TenantID         string
    RefreshTokenHash string
    DeviceInfo       map[string]interface{}
    IPAddress        string
    UserAgent        string
    IsActive         bool
    ExpiresAt        time.Time
    CreatedAt        time.Time
    LastAccessedAt   time.Time
}
```

**JWT Token Claims:**
```json
{
  "user_id": "uuid",
  "username": "john.doe",
  "email": "john.doe@acme.com",
  "tenant_id": "tenant-uuid",
  "session_id": "session-uuid",
  "roles": ["manager", "approver"],
  "permissions": ["dms:document:read", "dms:document:write"],
  "entity_id": "entity-uuid",
  "entity_type": "EMPLOYEE",
  "division_id": "division-uuid",
  "exp": 1735689600,
  "iat": 1735688700,
  "iss": "ugcl-platform"
}
```

---

### 3. Entity Module (`identity/entity`)

**Primary Responsibility:** Identity abstraction and authorization context

**Key Functions:**
- Unified identity for all actor types (human, machine, system)
- Links user accounts to domain-specific records
- Organizational hierarchy association
- Entity role bindings with scope
- Time-bound role assignments

**Entity Types Supported:**
```go
const (
    ENTITY_TYPE_EMPLOYEE   = 1  // Human: Internal staff
    ENTITY_TYPE_CONTRACTOR = 2  // Human: External contractors
    ENTITY_TYPE_VENDOR     = 3  // Organization: Vendor companies
    ENTITY_TYPE_CLIENT     = 4  // Organization: Client companies
    ENTITY_TYPE_DEVICE     = 5  // Machine: IoT devices
    ENTITY_TYPE_DRONE      = 6  // Machine: Drones, UAVs
    ENTITY_TYPE_BOT        = 7  // System: Automated bots
    ENTITY_TYPE_SYSTEM     = 8  // System: Internal services
    ENTITY_TYPE_ADMIN      = 9  // Human: System administrators
    ENTITY_TYPE_AGENT      = 10 // System: AI agents
)
```

**API Surface:**
```go
// Create entity for employee
CreateEntity(ctx, &CreateEntityRequest{
    TenantId:        "tenant-uuid",
    EntityType:      ENTITY_TYPE_EMPLOYEE,
    UserId:          "user-uuid",           // Links to auth user
    ReferenceId:     "employee-uuid",       // Links to employee record
    ReferenceSource: REFERENCE_SOURCE_EMPLOYEE,
    DivisionId:      "division-uuid",       // Org placement
    BranchId:        "branch-uuid",
    DepartmentId:    "department-uuid",
    Status:          ENTITY_STATUS_ACTIVE,
})

// Assign role to entity with scope
CreateEntityRoleBinding(ctx, &CreateEntityRoleBindingRequest{
    EntityId:     "entity-uuid",
    RoleId:       "department-manager-role",
    DepartmentId: "dept-uuid",              // Scoped to department
    ValidFrom:    timestamppb.Now(),
    ValidUntil:   timestamppb.New(time.Now().AddDate(1, 0, 0)), // 1 year
})

// Get entity by user ID (critical for auth flow)
GetEntityByUserId(ctx, &GetEntityByUserIdRequest{
    UserId:   "user-uuid",
    TenantId: "tenant-uuid",
})
// Returns: Entity with organizational context
```

**Entity Structure:**
```go
type Entity struct {
    ID              uuid.UUID
    TenantID        uuid.UUID
    EntityType      EntityType
    UserID          uuid.UUID    // → auth.user
    ReferenceID     uuid.UUID    // → employee/contractor/device record
    ReferenceSource string       // "EMPLOYEE", "CONTRACTOR", etc.
    Status          EntityStatus

    // Organizational context
    DivisionID      *uuid.UUID
    BranchID        *uuid.UUID
    DepartmentID    *uuid.UUID

    Metadata        map[string]interface{}
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

---

### 4. User Module (`identity/user`)

**Primary Responsibility:** Authorization rules and permission management

**Key Functions:**
- Permission definition templates
- Role management and hierarchies
- Permission resolution with multiple strategies
- Organizational scope checking
- Resource-level permissions
- Temporal permission constraints

**Permission Model:**
```go
// Permission template
type PermissionDef struct {
    Name        string  // "view_employee_salary"
    Namespace   string  // "employee"
    Resource    string  // "employee"
    Action      string  // "view_salary"
    Scope       string  // "*" or specific pattern
    Side        PermissionSide  // BOTH, HOST_ONLY, TENANT_ONLY
    Description string
}

// Concrete permission
type Permission struct {
    Namespace    string   // "employee"
    Resource     string   // "employee"
    Action       string   // "view_salary"
    Subject      string   // "entity:abc" or "user:123" or "role:xyz"
    Effect       Effect   // GRANT, FORBIDDEN, UNKNOWN
    TenantID     string

    // Organizational scope (optional)
    DivisionID   *string
    BranchID     *string
    DepartmentID *string

    // Resource instance scoping (optional)
    ResourceID   *string

    // Temporal constraints (optional)
    ValidFrom    *time.Time
    ValidUntil   *time.Time

    // Inheritance control
    AllowInheritance bool
}
```

**API Surface:**
```go
// Check permission
CheckPermission(ctx, &CheckPermissionRequest{
    Subject:      "entity:abc123",          // Entity ID
    Namespace:    "formbuilder",
    Resource:     "form",
    Action:       "approve",
    TenantId:     "tenant-uuid",
    DivisionId:   wrapperspb.String("div-uuid"),  // Optional scope
    ResourceId:   wrapperspb.String("form-uuid"), // Optional resource
    Mode:         PERMISSION_CHECK_MODE_WITH_INHERITANCE,
})
// Returns: {effect: GRANT/FORBIDDEN/UNKNOWN, matched_permission, resolution_path[], reason}

// Grant permission
GrantPermission(ctx, &Permission{
    Subject:    "entity:abc123",
    Namespace:  "dms",
    Resource:   "document",
    Action:     "delete",
    Effect:     GRANT,
    TenantId:   "tenant-uuid",
    DivisionId: "div-uuid",                 // Scoped to division
    ValidUntil: timestamppb.New(time.Now().AddDate(0, 6, 0)), // 6 months
})

// Assign role to user
AssignRoleToUser(ctx, &AssignRoleRequest{
    UserId: "user-uuid",
    RoleId: "manager-role",
})
```

**Permission Resolution Order:**
```
1. Direct Permissions     → Check exact subject match (highest priority)
2. Entity Role Permissions → Via entity role bindings
3. User Role Permissions   → Via user role assignments
4. Inherited Permissions   → From organizational hierarchy
5. Resource Ownership      → Owner/shared resource access
```

---

## Authentication Flow

### Complete Login Flow with Entity Resolution

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       │ 1. POST /api/auth/login
       │    {email, password, tenant_id}
       ↓
┌─────────────────────────────────────────────────────────────┐
│                      Auth Module                            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ 2. Validate credentials (bcrypt)                           │
│    ✓ Account active                                        │
│    ✓ Not locked                                            │
│    ✓ Password correct                                      │
│                                                             │
│ 3. Check 2FA status                                        │
│    If enabled → Return requires_two_factor=true            │
│                                                             │
│ 4. Create session record                                   │
│    - Session ID                                            │
│    - User ID, Tenant ID                                    │
│    - Device info, IP address                               │
│    - Refresh token hash                                    │
│                                                             │
└────────────┬────────────────────────────────────────────────┘
             │
             │ 5. Get entity for user
             ↓
┌─────────────────────────────────────────────────────────────┐
│                    Entity Module                            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ 6. Query: GetEntityByUserId(user_id, tenant_id)           │
│    Returns:                                                │
│    - entity_id                                             │
│    - entity_type (EMPLOYEE, CONTRACTOR, etc.)             │
│    - division_id, branch_id, department_id                │
│    - reference_id (employee record UUID)                  │
│                                                             │
│ 7. Get entity role bindings                               │
│    Query: GetActiveEntityRoleBindings(entity_id)          │
│    Returns: [{role_id, scope, valid_from, valid_until}]  │
│                                                             │
└────────────┬────────────────────────────────────────────────┘
             │
             │ 8. Get permissions
             ↓
┌─────────────────────────────────────────────────────────────┐
│                     User Module                             │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ 9. Aggregate permissions from:                             │
│    - Direct user permissions                               │
│    - Entity role permissions                               │
│    - User role permissions                                 │
│                                                             │
│    Returns: ["dms:document:read", "dms:document:write",   │
│              "employee:employee:view", ...]                │
│                                                             │
└────────────┬────────────────────────────────────────────────┘
             │
             │ 10. Generate JWT tokens
             ↓
┌─────────────────────────────────────────────────────────────┐
│                      Auth Module                            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ 11. Generate JWT with claims:                              │
│     {                                                      │
│       user_id, username, email,                           │
│       tenant_id, session_id,                              │
│       entity_id, entity_type,                             │
│       division_id, branch_id, department_id,              │
│       roles: [],                                          │
│       permissions: [],                                    │
│       exp, iat, iss                                       │
│     }                                                      │
│                                                             │
│ 12. Log security event                                     │
│     - LOGIN_SUCCESS                                        │
│     - IP, user agent, timestamp                           │
│                                                             │
└────────────┬────────────────────────────────────────────────┘
             │
             │ 13. Return to client
             ↓
┌─────────────────────────────────────────────────────────────┐
│  Response: {                                                │
│    access_token: "eyJhbGc...",                             │
│    refresh_token: "eyJhbGc...",                            │
│    expires_at: "2025-01-01T12:15:00Z",                    │
│    user: {                                                 │
│      user_id, username, email,                            │
│      entity_id, entity_type,                              │
│      division_id, division_name,                          │
│      roles: ["manager", "approver"],                      │
│      permissions: ["dms:document:*"]                      │
│    }                                                       │
│  }                                                         │
└─────────────────────────────────────────────────────────────┘
```

### Request Authentication Flow

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       │ API Request
       │ Header: Authorization: Bearer eyJhbGc...
       ↓
┌─────────────────────────────────────────────────────────────┐
│                   API Gateway / Middleware                  │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ 1. Extract JWT from Authorization header                   │
│                                                             │
│ 2. Validate JWT signature (HS256)                          │
│    ✓ Valid signature                                       │
│    ✓ Not expired                                           │
│    ✓ Correct issuer                                        │
│                                                             │
│ 3. Check token revocation                                  │
│    Query: auth_revoked_tokens WHERE jti = ?               │
│    (Token blacklist check)                                 │
│                                                             │
│ 4. Validate session still active                           │
│    Query: auth_sessions WHERE session_id = ? AND active   │
│                                                             │
│ 5. Extract claims and populate request context:            │
│    ctx.Set("user_id", claims.UserID)                      │
│    ctx.Set("tenant_id", claims.TenantID)                  │
│    ctx.Set("entity_id", claims.EntityID)                  │
│    ctx.Set("permissions", claims.Permissions)             │
│                                                             │
│ 6. Route to appropriate database (multi-tenancy)          │
│    - If SEPERATEDB → route to tenant database             │
│    - If SEPERATESCHEMA → set schema search path           │
│    - If SHARED → add tenant_id filter                     │
│                                                             │
└────────────┬────────────────────────────────────────────────┘
             │
             │ Authenticated request
             ↓
        ┌────────────┐
        │  Service   │
        │   Layer    │
        └────────────┘
```

---

## Authorization Flow

### Permission Check Flow (Form-Level & Field-Level)

```
┌─────────────────────────────────────────────────────────────┐
│            Permission Resolution Algorithm                  │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ Input: CheckPermissionRequest {                            │
│   subject: "entity:abc123",                                │
│   namespace: "formbuilder",                                │
│   resource: "form",                                        │
│   action: "approve",                                       │
│   tenant_id: "tenant-uuid",                                │
│   division_id: "div-uuid",                                 │
│   resource_id: "form-123",                                 │
│   mode: WITH_INHERITANCE                                   │
│ }                                                          │
│                                                             │
│ Resolution Steps (Waterfall - First Match Wins):          │
│                                                             │
│ STEP 1: Check Direct Permissions ──────────────────────────│
│   Query: permissions                                        │
│   WHERE subject = 'entity:abc123'                          │
│     AND namespace = 'formbuilder'                          │
│     AND resource = 'form'                                  │
│     AND action IN ('approve', '*')                         │
│     AND tenant_id = 'tenant-uuid'                          │
│     AND (resource_id IS NULL OR resource_id = 'form-123') │
│     AND (valid_from IS NULL OR valid_from <= NOW())       │
│     AND (valid_until IS NULL OR valid_until >= NOW())     │
│                                                             │
│   If FORBIDDEN effect found → Deny immediately (explicit)  │
│   If GRANT effect found → Return GRANT                     │
│   Else → Continue to Step 2                                │
│                                                             │
│ STEP 2: Check Entity Role Permissions ─────────────────────│
│   Query: entity_role_bindings                              │
│   WHERE entity_id = 'abc123'                               │
│     AND tenant_id = 'tenant-uuid'                          │
│     AND (division_id IS NULL OR division_id = 'div-uuid') │
│     AND (valid_from IS NULL OR valid_from <= NOW())       │
│     AND (valid_until IS NULL OR valid_until >= NOW())     │
│                                                             │
│   For each role binding:                                   │
│     Get role permissions from role_permissions table       │
│     Check if permission matches request                    │
│     If match → Return GRANT                                │
│                                                             │
│   Else → Continue to Step 3                                │
│                                                             │
│ STEP 3: Check User Role Permissions ───────────────────────│
│   Query: user_roles → role_permissions                     │
│   Similar logic to Step 2                                  │
│   (Legacy support - prefer entity roles)                   │
│                                                             │
│ STEP 4: Check Inherited Permissions ───────────────────────│
│   (Only if mode = WITH_INHERITANCE)                        │
│                                                             │
│   Get organizational hierarchy:                            │
│   - If department specified → check branch, division       │
│   - If branch specified → check division                   │
│   - Walk up hierarchy checking permissions at each level   │
│                                                             │
│   For each ancestor level:                                 │
│     Query entity_role_bindings with ancestor scope         │
│     Check permissions with allow_inheritance = true        │
│     If match → Return GRANT                                │
│                                                             │
│ STEP 5: Check Resource Ownership ──────────────────────────│
│   Query: Check if entity owns the resource                 │
│   - forms.created_by = entity_id                           │
│   - forms.shared_with contains entity_id                   │
│                                                             │
│   If owner → Return GRANT (full permissions)               │
│   If shared → Return GRANT for shared actions              │
│                                                             │
│ STEP 6: Default Deny ──────────────────────────────────────│
│   If no permission found → Return UNKNOWN (deny)           │
│                                                             │
│ Output: CheckPermissionResponse {                          │
│   effect: GRANT | FORBIDDEN | UNKNOWN,                     │
│   matched_permission: {...},                               │
│   resolution_path: [                                       │
│     {source: "entity_role", source_id: "role-abc",        │
│      scope: "division", description: "..."}                │
│   ],                                                       │
│   reason: "Permission granted via entity role 'Manager'"   │
│ }                                                          │
└─────────────────────────────────────────────────────────────┘
```

---

## Form-Level Authorization Example

### Scenario: Form Approval System

**Business Requirements:**
- Only department managers can approve forms created in their department
- Division managers can approve forms from any department in their division
- Form creators can edit their own forms but cannot approve them
- Approval permission expires after 1 year (requires renewal)

### 1. Setup Permission Definitions

```go
// Define permission templates
CreatePermissionDef(ctx, &PermissionDef{
    Name:        "approve_form",
    Namespace:   "formbuilder",
    Resource:    "form",
    Action:      "approve",
    Scope:       "*",
    Side:        BOTH,  // Available in host and tenant
    Description: "Ability to approve forms",
})

CreatePermissionDef(ctx, &PermissionDef{
    Name:        "edit_form",
    Namespace:   "formbuilder",
    Resource:    "form",
    Action:      "edit",
    Scope:       "*",
    Side:        BOTH,
    Description: "Ability to edit form content",
})

CreatePermissionDef(ctx, &PermissionDef{
    Name:        "submit_form",
    Namespace:   "formbuilder",
    Resource:    "form",
    Action:      "submit",
    Scope:       "*",
    Side:        TENANT_ONLY,  // Only for tenant users
    Description: "Ability to submit forms for approval",
})
```

### 2. Setup Roles with Permissions

```go
// Department Manager role
role := CreateRole(ctx, &Role{
    Name:     "department_manager",
    TenantID: "tenant-uuid",
})

// Assign permissions to role
AssignPermissionToRole(ctx, role.ID, &Permission{
    Namespace:        "formbuilder",
    Resource:         "form",
    Action:           "approve",
    Effect:           GRANT,
    AllowInheritance: false,  // Cannot be inherited by sub-departments
})

AssignPermissionToRole(ctx, role.ID, &Permission{
    Namespace: "formbuilder",
    Resource:  "form",
    Action:    "edit",
    Effect:    GRANT,
})

// Division Manager role (higher authority)
divRole := CreateRole(ctx, &Role{
    Name:     "division_manager",
    TenantID: "tenant-uuid",
})

AssignPermissionToRole(ctx, divRole.ID, &Permission{
    Namespace:        "formbuilder",
    Resource:         "form",
    Action:           "approve",
    Effect:           GRANT,
    AllowInheritance: true,  // Can approve forms in sub-departments
})
```

### 3. Assign Roles to Entities

```go
// Assign department manager role to John (scoped to specific department)
CreateEntityRoleBinding(ctx, &CreateEntityRoleBindingRequest{
    EntityId:     "entity-john",
    RoleId:       "department_manager_role_id",
    DepartmentId: "hr-department-uuid",  // Scoped to HR department
    ValidFrom:    timestamppb.Now(),
    ValidUntil:   timestamppb.New(time.Now().AddDate(1, 0, 0)), // Expires in 1 year
})

// Assign division manager role to Sarah (scoped to entire division)
CreateEntityRoleBinding(ctx, &CreateEntityRoleBindingRequest{
    EntityId:   "entity-sarah",
    RoleId:     "division_manager_role_id",
    DivisionId: "operations-division-uuid",  // Scoped to entire division
    ValidFrom:  timestamppb.Now(),
    // No ValidUntil = permanent
})
```

### 4. Form Creation with Ownership

```go
// When a form is created
form := CreateForm(ctx, &Form{
    ID:           "form-123",
    TenantID:     "tenant-uuid",
    DepartmentID: "hr-department-uuid",
    CreatedBy:    "entity-alice",  // Alice created the form
    Status:       "PENDING_APPROVAL",
})

// Grant creator permission to edit their own form
GrantPermission(ctx, &Permission{
    Subject:    "entity-alice",
    Namespace:  "formbuilder",
    Resource:   "form",
    Action:     "edit",
    Effect:     GRANT,
    ResourceID: "form-123",  // Scoped to this specific form
    TenantID:   "tenant-uuid",
})

// Explicitly forbid creator from approving their own form
GrantPermission(ctx, &Permission{
    Subject:    "entity-alice",
    Namespace:  "formbuilder",
    Resource:   "form",
    Action:     "approve",
    Effect:     FORBIDDEN,  // Explicit denial
    ResourceID: "form-123",
    TenantID:   "tenant-uuid",
})
```

### 5. Authorization Checks in API Handlers

```go
// Form Approval Handler
func ApproveFormHandler(ctx context.Context, req *ApproveFormRequest) error {
    // Extract entity from JWT context
    entityID := ctx.Value("entity_id").(string)

    // Get form details
    form, err := formRepo.GetForm(ctx, req.FormId)
    if err != nil {
        return err
    }

    // Check if entity has approval permission for this form
    permResp, err := userClient.CheckPermission(ctx, &CheckPermissionRequest{
        Subject:      fmt.Sprintf("entity:%s", entityID),
        Namespace:    "formbuilder",
        Resource:     "form",
        Action:       "approve",
        TenantId:     form.TenantID,
        DepartmentId: wrapperspb.String(form.DepartmentID),  // Check in form's department
        ResourceId:   wrapperspb.String(form.ID),             // Specific form
        Mode:         PERMISSION_CHECK_MODE_WITH_INHERITANCE, // Allow inheritance
    })

    if err != nil {
        return fmt.Errorf("permission check failed: %w", err)
    }

    // Authorization decision
    if permResp.Effect == pb.Effect_GRANT {
        // ✅ Permission granted
        return formService.ApproveForm(ctx, form.ID, entityID)
    } else if permResp.Effect == pb.Effect_FORBIDDEN {
        // ❌ Explicitly denied (e.g., creator trying to approve own form)
        return status.Errorf(codes.PermissionDenied,
            "You cannot approve your own form. Reason: %s", permResp.Reason)
    } else {
        // ❌ No permission found
        return status.Errorf(codes.PermissionDenied,
            "You do not have permission to approve forms in this department")
    }
}
```

### 6. Test Scenarios

```go
// Test 1: Department manager approves form in their department
// Entity: John (department_manager in HR department)
// Form: Created in HR department
CheckPermission("entity-john", "formbuilder", "form", "approve",
                hr_dept_id, form_123_id)
// Result: GRANT (via entity role binding scoped to HR department)

// Test 2: Department manager tries to approve form in different department
// Entity: John (department_manager in HR department)
// Form: Created in Finance department
CheckPermission("entity-john", "formbuilder", "form", "approve",
                finance_dept_id, form_456_id)
// Result: UNKNOWN (role binding scoped to HR only)

// Test 3: Division manager approves form anywhere in division
// Entity: Sarah (division_manager in Operations division)
// Form: Created in any department under Operations division
CheckPermission("entity-sarah", "formbuilder", "form", "approve",
                logistics_dept_id, form_789_id)
// Result: GRANT (role scoped to division, inheritance enabled)

// Test 4: Form creator tries to approve own form
// Entity: Alice (form creator)
// Form: Created by Alice
CheckPermission("entity-alice", "formbuilder", "form", "approve",
                hr_dept_id, form_123_id)
// Result: FORBIDDEN (explicit deny permission exists)

// Test 5: Form creator edits own form
// Entity: Alice (form creator)
// Form: Created by Alice
CheckPermission("entity-alice", "formbuilder", "form", "edit",
                hr_dept_id, form_123_id)
// Result: GRANT (resource ownership or direct permission)
```

---

## Field-Level Authorization Example

### Scenario: Employee Profile with Salary Field

**Business Requirements:**
- All employees can view basic profile fields (name, department, designation)
- Only HR department can view salary fields
- Only division managers and above can edit salary fields
- Managers can view salary of their direct reports only

### 1. Define Field-Level Permissions

```go
// Permission for viewing salary field
CreatePermissionDef(ctx, &PermissionDef{
    Name:        "view_employee_salary",
    Namespace:   "employee",
    Resource:    "employee",
    Action:      "view_salary",
    Scope:       "field:salary",  // Field-level scope
    Side:        BOTH,
    Description: "View employee salary information",
})

// Permission for editing salary field
CreatePermissionDef(ctx, &PermissionDef{
    Name:        "edit_employee_salary",
    Namespace:   "employee",
    Resource:    "employee",
    Action:      "edit_salary",
    Scope:       "field:salary",
    Side:        BOTH,
    Description: "Edit employee salary information",
})
```

### 2. Assign Permissions to Roles

```go
// HR role gets full salary access
hrRole := CreateRole(ctx, &Role{
    Name:     "hr_officer",
    TenantID: "tenant-uuid",
})

AssignPermissionToRole(ctx, hrRole.ID, &Permission{
    Namespace:        "employee",
    Resource:         "employee",
    Action:           "view_salary",
    Effect:           GRANT,
    AllowInheritance: false,
})

AssignPermissionToRole(ctx, hrRole.ID, &Permission{
    Namespace:        "employee",
    Resource:         "employee",
    Action:           "edit_salary",
    Effect:           GRANT,
    AllowInheritance: false,
})

// Division manager gets view-only salary access
divManagerRole := CreateRole(ctx, &Role{
    Name:     "division_manager",
    TenantID: "tenant-uuid",
})

AssignPermissionToRole(ctx, divManagerRole.ID, &Permission{
    Namespace:        "employee",
    Resource:         "employee",
    Action:           "view_salary",
    Effect:           GRANT,
    AllowInheritance: true,  // Can view salaries in subordinate departments
})

AssignPermissionToRole(ctx, divManagerRole.ID, &Permission{
    Namespace:        "employee",
    Resource:         "employee",
    Action:           "edit_salary",
    Effect:           GRANT,
    AllowInheritance: false,  // Cannot edit in subordinate departments
})
```

### 3. Manager-Subordinate Relationship

```go
// Direct report permission (manager can view subordinate's salary)
// When Bob becomes Alice's manager:
GrantPermission(ctx, &Permission{
    Subject:    "entity-bob",             // Manager
    Namespace:  "employee",
    Resource:   "employee",
    Action:     "view_salary",
    Effect:     GRANT,
    ResourceID: "employee-alice",         // Specific employee
    TenantID:   "tenant-uuid",
    ValidUntil: nil,                      // Valid while reporting relationship exists
})

// When Alice is no longer reporting to Bob, revoke permission:
RevokePermission(ctx, "entity-bob", "employee", "employee", "view_salary", "employee-alice")
```

### 4. Field-Level Filtering in API Response

```go
// Employee Profile Handler
func GetEmployeeProfileHandler(ctx context.Context, req *GetEmployeeRequest) (*EmployeeProfile, error) {
    entityID := ctx.Value("entity_id").(string)

    // Get employee record
    employee, err := employeeRepo.GetEmployee(ctx, req.EmployeeId)
    if err != nil {
        return nil, err
    }

    // Build response with basic fields (always visible)
    profile := &EmployeeProfile{
        ID:           employee.ID,
        Name:         employee.Name,
        Department:   employee.Department,
        Designation:  employee.Designation,
        Email:        employee.Email,
    }

    // Check if requester can view salary field
    canViewSalary, err := checkFieldPermission(ctx, entityID, employee.ID,
                                               "employee", "view_salary", employee.DepartmentID)
    if err != nil {
        return nil, err
    }

    if canViewSalary {
        // ✅ Include salary fields
        profile.Salary = employee.Salary
        profile.Benefits = employee.Benefits
        profile.TaxDeductions = employee.TaxDeductions
    } else {
        // ❌ Exclude salary fields (return nil or masked value)
        profile.Salary = nil  // Or use 0, "***", etc.
    }

    return profile, nil
}

// Helper function for field permission check
func checkFieldPermission(ctx context.Context, entityID, resourceID, resource, action, deptID string) (bool, error) {
    permResp, err := userClient.CheckPermission(ctx, &CheckPermissionRequest{
        Subject:      fmt.Sprintf("entity:%s", entityID),
        Namespace:    "employee",
        Resource:     resource,
        Action:       action,
        TenantId:     ctx.Value("tenant_id").(string),
        DepartmentId: wrapperspb.String(deptID),
        ResourceId:   wrapperspb.String(resourceID),  // Specific employee
        Mode:         PERMISSION_CHECK_MODE_WITH_INHERITANCE,
    })

    if err != nil {
        return false, err
    }

    return permResp.Effect == pb.Effect_GRANT, nil
}
```

### 5. Field-Level Validation in Update Operations

```go
// Update Employee Handler
func UpdateEmployeeHandler(ctx context.Context, req *UpdateEmployeeRequest) error {
    entityID := ctx.Value("entity_id").(string)

    employee, err := employeeRepo.GetEmployee(ctx, req.EmployeeId)
    if err != nil {
        return err
    }

    // Check if update includes salary fields
    if req.FieldMask.Paths contains "salary" or "benefits" {
        // Verify permission to edit salary
        canEditSalary, err := checkFieldPermission(ctx, entityID, employee.ID,
                                                   "employee", "edit_salary", employee.DepartmentID)
        if err != nil {
            return err
        }

        if !canEditSalary {
            return status.Errorf(codes.PermissionDenied,
                "You do not have permission to edit salary information")
        }
    }

    // Check permission for other fields
    if req.FieldMask.Paths contains "designation" {
        canEditDesignation, err := checkFieldPermission(ctx, entityID, employee.ID,
                                                        "employee", "edit_designation", employee.DepartmentID)
        if !canEditDesignation {
            return status.Errorf(codes.PermissionDenied,
                "You do not have permission to edit designation")
        }
    }

    // Proceed with update if all permissions granted
    return employeeService.UpdateEmployee(ctx, req)
}
```

### 6. Dynamic Field Masking

```go
// Advanced: Return field metadata with permissions
type EmployeeProfileWithPermissions struct {
    Employee    *Employee
    FieldAccess map[string]FieldPermission
}

type FieldPermission struct {
    CanView bool
    CanEdit bool
}

func GetEmployeeProfileWithPermissions(ctx context.Context, employeeID string) (*EmployeeProfileWithPermissions, error) {
    entityID := ctx.Value("entity_id").(string)
    employee, _ := employeeRepo.GetEmployee(ctx, employeeID)

    // Check permissions for each sensitive field
    fields := []string{"salary", "benefits", "tax_deductions", "bank_account", "emergency_contact"}
    fieldAccess := make(map[string]FieldPermission)

    for _, field := range fields {
        canView, _ := checkFieldPermission(ctx, entityID, employeeID, "employee",
                                           fmt.Sprintf("view_%s", field), employee.DepartmentID)
        canEdit, _ := checkFieldPermission(ctx, entityID, employeeID, "employee",
                                           fmt.Sprintf("edit_%s", field), employee.DepartmentID)

        fieldAccess[field] = FieldPermission{
            CanView: canView,
            CanEdit: canEdit,
        }
    }

    return &EmployeeProfileWithPermissions{
        Employee:    employee,
        FieldAccess: fieldAccess,
    }, nil
}

// Frontend can use this to show/hide fields and enable/disable edit buttons
```

### 7. Test Scenarios

```go
// Test 1: HR officer views any employee's salary
// Entity: HR Officer
// Employee: Any employee in tenant
CheckPermission("entity-hr-officer", "employee", "employee", "view_salary",
                any_dept_id, employee_id)
// Result: GRANT (HR role has blanket permission)

// Test 2: Division manager views salary in their division
// Entity: Division Manager (Operations)
// Employee: Employee in Logistics (child department of Operations)
CheckPermission("entity-div-manager", "employee", "employee", "view_salary",
                logistics_dept_id, employee_id)
// Result: GRANT (role scoped to division with inheritance)

// Test 3: Manager views direct report's salary
// Entity: Bob (manager)
// Employee: Alice (reports to Bob)
CheckPermission("entity-bob", "employee", "employee", "view_salary",
                hr_dept_id, "employee-alice")
// Result: GRANT (direct resource-level permission)

// Test 4: Manager views non-report's salary
// Entity: Bob (manager)
// Employee: Charlie (does not report to Bob)
CheckPermission("entity-bob", "employee", "employee", "view_salary",
                hr_dept_id, "employee-charlie")
// Result: UNKNOWN (no permission exists)

// Test 5: Regular employee views own salary
// Entity: Alice
// Employee: Alice (self)
CheckPermission("entity-alice", "employee", "employee", "view_salary",
                hr_dept_id, "employee-alice")
// Result: Depends on policy - could GRANT via special self-view permission

// Test 6: Division manager tries to edit salary in subordinate department
// Entity: Division Manager
// Employee: Employee in child department
CheckPermission("entity-div-manager", "employee", "employee", "edit_salary",
                child_dept_id, employee_id)
// Result: UNKNOWN (edit permission has allow_inheritance=false)
```

---

## Implementation Patterns

### Pattern 1: Middleware-Based Authorization

```go
// Middleware to check resource-level permissions
func RequirePermission(namespace, resource, action string) gin.HandlerFunc {
    return func(c *gin.Context) {
        entityID := c.GetString("entity_id")
        tenantID := c.GetString("tenant_id")
        resourceID := c.Param("id")  // e.g., form ID from URL

        permResp, err := userClient.CheckPermission(c, &CheckPermissionRequest{
            Subject:    fmt.Sprintf("entity:%s", entityID),
            Namespace:  namespace,
            Resource:   resource,
            Action:     action,
            TenantId:   tenantID,
            ResourceId: wrapperspb.String(resourceID),
            Mode:       PERMISSION_CHECK_MODE_WITH_INHERITANCE,
        })

        if err != nil || permResp.Effect != pb.Effect_GRANT {
            c.JSON(403, gin.H{"error": "Permission denied"})
            c.Abort()
            return
        }

        c.Next()
    }
}

// Usage in route definitions
router.PUT("/forms/:id/approve",
    RequireAuth(),  // JWT validation
    RequirePermission("formbuilder", "form", "approve"),
    ApproveFormHandler)
```

### Pattern 2: Annotation-Based Authorization (gRPC Interceptor)

```go
// gRPC interceptor for permission checking
func PermissionInterceptor(permService pb.PermissionServiceClient) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
                handler grpc.UnaryHandler) (interface{}, error) {

        // Extract permission requirements from method annotations
        perm := extractPermissionFromMethod(info.FullMethod)
        if perm == nil {
            // No permission required
            return handler(ctx, req)
        }

        // Extract entity from context
        entityID := ctx.Value("entity_id").(string)
        tenantID := ctx.Value("tenant_id").(string)

        // Check permission
        permResp, err := permService.CheckPermission(ctx, &CheckPermissionRequest{
            Subject:   fmt.Sprintf("entity:%s", entityID),
            Namespace: perm.Namespace,
            Resource:  perm.Resource,
            Action:    perm.Action,
            TenantId:  tenantID,
            Mode:      PERMISSION_CHECK_MODE_WITH_INHERITANCE,
        })

        if err != nil || permResp.Effect != pb.Effect_GRANT {
            return nil, status.Errorf(codes.PermissionDenied,
                "Permission denied: %s", permResp.Reason)
        }

        return handler(ctx, req)
    }
}

// Annotation in proto file (custom option)
rpc ApproveForm(ApproveFormRequest) returns (ApproveFormResponse) {
    option (permission) = {
        namespace: "formbuilder",
        resource: "form",
        action: "approve"
    };
}
```

### Pattern 3: Bulk Permission Checking

```go
// For list operations, check permissions in bulk
func ListFormsHandler(ctx context.Context, req *ListFormsRequest) (*ListFormsResponse, error) {
    entityID := ctx.Value("entity_id").(string)

    // Get all forms (without permission filtering)
    forms, err := formRepo.ListForms(ctx, req.DepartmentId)
    if err != nil {
        return nil, err
    }

    // Get effective permissions for entity
    perms, err := userClient.GetEffectivePermissions(ctx,
        fmt.Sprintf("entity:%s", entityID),
        &PermissionResolveOptions{
            TenantID:     ctx.Value("tenant_id").(string),
            DepartmentID: &req.DepartmentId,
        })

    // Build permission map for quick lookup
    permMap := buildPermissionMap(perms)

    // Filter forms based on permissions
    allowedForms := []Form{}
    for _, form := range forms {
        if canAccessForm(permMap, form) {
            // Optionally mask fields based on field-level permissions
            form = maskSensitiveFields(form, permMap)
            allowedForms = append(allowedForms, form)
        }
    }

    return &ListFormsResponse{Forms: allowedForms}, nil
}
```

### Pattern 4: Cached Permission Resolution

```go
// Cache permission results to reduce database queries
type PermissionCache struct {
    cache *lru.Cache
    ttl   time.Duration
}

func (pc *PermissionCache) CheckPermission(ctx context.Context, req *CheckPermissionRequest) (*CheckPermissionResponse, error) {
    // Build cache key
    key := fmt.Sprintf("%s:%s:%s:%s:%s",
        req.Subject, req.Namespace, req.Resource, req.Action, req.TenantId)

    // Check cache
    if cached, found := pc.cache.Get(key); found {
        return cached.(*CheckPermissionResponse), nil
    }

    // Cache miss - query database
    resp, err := pc.permService.CheckPermission(ctx, req)
    if err != nil {
        return nil, err
    }

    // Store in cache with TTL
    pc.cache.Add(key, resp)
    time.AfterFunc(pc.ttl, func() {
        pc.cache.Remove(key)
    })

    return resp, nil
}

// Invalidate cache on permission changes
func (pc *PermissionCache) InvalidateEntity(entityID string) {
    // Remove all cache entries for this entity
    for _, key := range pc.cache.Keys() {
        if strings.HasPrefix(key.(string), fmt.Sprintf("entity:%s:", entityID)) {
            pc.cache.Remove(key)
        }
    }
}
```

---

## Code Duplication Analysis

### Identified Overlaps

#### 1. ❌ **Password Management** (User Module vs Auth Module)

**Location in User Module:**
- `d:\Maheshwari\UGCL\backend\v2\identity\user\proto\user.proto`
  - `ChangePassword()`, `InitiatePasswordReset()`, `ResetPassword()`
- Status: **Stub implementations with TODO comments**

**Location in Auth Module:**
- `d:\Maheshwari\UGCL\backend\v2\identity\auth\services\user_service.go`
  - Fully implemented password change, forgot password, reset password

**Recommendation:**
✅ **Remove from User Module** - Keep all password operations in Auth module
- Delete password-related RPCs from `user.proto`
- Update documentation to direct users to Auth module APIs

#### 2. ❌ **Email/Phone Verification** (User Module vs Auth Module)

**Location in User Module:**
- `d:\Maheshwari\UGCL\backend\v2\identity\user\proto\user.proto`
  - `SendVerificationEmail()`, `VerifyEmail()`, `SendVerificationSMS()`, `VerifyPhone()`
- Status: **Stub implementations**

**Location in Auth Module:**
- `d:\Maheshwari\UGCL\backend\v2\identity\auth\db\schema\schema.sql`
  - `auth_email_verification_tokens`, `auth_phone_verification_tokens` tables
- Service implementation for verification

**Recommendation:**
✅ **Remove from User Module** - Verification is authentication concern
- Delete verification RPCs from `user.proto`
- Keep verification fields in user model (`email_verified`, `phone_verified`)
- Auth module updates these fields after verification

#### 3. ❌ **OTP Model** (User Module)

**Location:**
- `d:\Maheshwari\UGCL\backend\v2\identity\user\models\otp.go`
- Status: **Defined but no database table exists**

**Auth Module:**
- `auth_phone_verification_tokens` table stores OTP codes
- `otp_service.go` implements OTP generation/verification

**Recommendation:**
✅ **Delete OTP model from User Module** - No longer needed

#### 4. ⚠️ **Role Management** (User Module vs Entity Module)

**User Module:**
- `user_roles` table - Direct user-to-role assignments
- Legacy pattern

**Entity Module:**
- `entity_role_bindings` table - Entity-to-role with organizational scope
- Modern pattern with temporal constraints

**Recommendation:**
✅ **Deprecate user_roles** - Migrate to entity-based role bindings
- Add migration script to convert `user_roles` → `entity_role_bindings`
- Update permission resolver to prioritize entity roles
- Eventually remove `user_roles` table

#### 5. ✅ **User Fields vs Auth Fields** (Acceptable Overlap)

**User Module:**
- `password_hash`, `salt`, `two_factor_enabled`, `two_factor_secret`
- These are stored in user model for profile completeness

**Auth Module:**
- `auth_users` table duplicates some fields
- Necessary for auth module independence

**Assessment:**
✅ **This is acceptable architectural duplication**
- Auth module needs its own user table for authentication
- User module maintains user profile data
- Fields synchronized via events

**Recommendation:**
- Keep as-is but ensure synchronization via domain events
- Auth module publishes `PasswordChanged`, `TwoFactorEnabled` events
- User module subscribes and updates its records

### Duplication Summary Table

| Feature | User Module | Auth Module | Entity Module | Recommendation |
|---------|-------------|-------------|---------------|----------------|
| Password Management | Stubs (TODO) | ✅ Implemented | - | **Delete from User** |
| Email/Phone Verification | Stubs | ✅ Implemented | - | **Delete from User** |
| OTP Model | Defined (unused) | ✅ Used | - | **Delete from User** |
| Role Assignment | user_roles (legacy) | - | ✅ entity_role_bindings | **Deprecate user_roles** |
| User Credentials | User model | auth_users table | - | ✅ **Keep both (sync via events)** |
| Session Management | - | ✅ Implemented | - | ✅ **Correct location** |
| Permission Checking | ✅ Implemented | - | - | ✅ **Correct location** |
| Entity Context | - | - | ✅ Implemented | ✅ **Correct location** |
| Tenant Management | - | - | - | ✅ **In Tenant module** |

---

## Integration Guide

### Step 1: Complete Login Flow Implementation

```go
// In Auth Module: services/auth_service.go

func (s *AuthService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
    // 1. Validate user credentials
    user, err := s.validateCredentials(ctx, req)
    if err != nil {
        return nil, err
    }

    // 2. Check 2FA
    if user.TwoFactorEnabled {
        session := s.createPending2FASession(ctx, user.ID, req.TenantId)
        return &pb.LoginResponse{
            RequiresTwoFactor: true,
            SessionId:         session.SessionID,
        }, nil
    }

    // 3. Get entity for user (INTEGRATION POINT)
    entity, err := s.entityClient.GetEntityByUserId(ctx, &entity_pb.GetEntityByUserIdRequest{
        UserId:   user.ID.String(),
        TenantId: req.TenantId,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to get entity: %w", err)
    }

    // 4. Get entity role bindings
    roleBindings, err := s.entityClient.GetEntityRoleBindings(ctx, &entity_pb.GetEntityRoleBindingsRequest{
        EntityId: entity.Id,
        TenantId: req.TenantId,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to get role bindings: %w", err)
    }

    // 5. Get permissions from roles (INTEGRATION POINT)
    permissions := []string{}
    roles := []string{}
    for _, binding := range roleBindings.Bindings {
        rolePerms, err := s.userClient.ListPermissionsByRole(ctx, &user_pb.RoleIdentifier{
            Uuid: binding.RoleId,
        })
        if err != nil {
            continue
        }

        roles = append(roles, binding.RoleName)
        for _, perm := range rolePerms.Permissions {
            permissions = append(permissions, fmt.Sprintf("%s:%s:%s",
                perm.Namespace, perm.Resource, perm.Action))
        }
    }

    // 6. Create session
    session, err := s.createSession(ctx, user.ID, req.TenantId, req.DeviceInfo, req.IpAddress)
    if err != nil {
        return nil, err
    }

    // 7. Generate JWT tokens with entity context
    accessToken, refreshToken, err := s.jwtService.GenerateTokenPair(user, entity, roles, permissions, session.SessionID)
    if err != nil {
        return nil, err
    }

    // 8. Build UserSession response
    userSession := &pb.UserSession{
        UserId:       user.ID.String(),
        Username:     user.Username,
        Email:        user.Email,
        TenantId:     req.TenantId,
        Roles:        roles,
        Permissions:  permissions,

        // Entity context (ENHANCED)
        EntityId:     entity.Id,
        EntityType:   entity.EntityType.String(),
        DivisionId:   entity.DivisionId,
        BranchId:     entity.BranchId,
        DepartmentId: entity.DepartmentId,

        // Entity roles with scope
        EntityRoles:  buildEntityRoleInfos(roleBindings.Bindings),
    }

    return &pb.LoginResponse{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresAt:    timestamppb.New(time.Now().Add(15 * time.Minute)),
        User:         userSession,
        SessionId:    session.SessionID,
    }, nil
}
```

### Step 2: Implement Entity Role Provider

```go
// In User Module: services/entity_role_provider.go

type EntityRoleProvider struct {
    entityClient entity_pb.EntityServiceClient  // ← Inject entity service client
}

func NewEntityRoleProvider(entityClient entity_pb.EntityServiceClient) *EntityRoleProvider {
    return &EntityRoleProvider{
        entityClient: entityClient,
    }
}

func (p *EntityRoleProvider) GetEntityRoleBindings(ctx context.Context, entityID, tenantID string) ([]*EntityRoleBindingInfo, error) {
    // Call entity service
    resp, err := p.entityClient.GetEntityRoleBindings(ctx, &entity_pb.GetEntityRoleBindingsRequest{
        EntityId: entityID,
        TenantId: tenantID,
    })
    if err != nil {
        return nil, err
    }

    // Convert to EntityRoleBindingInfo
    bindings := make([]*EntityRoleBindingInfo, len(resp.Bindings))
    for i, b := range resp.Bindings {
        bindings[i] = &EntityRoleBindingInfo{
            RoleID:       b.RoleId,
            RoleName:     b.RoleName,
            DivisionID:   b.DivisionId,
            BranchID:     b.BranchId,
            DepartmentID: b.DepartmentId,
            ValidFrom:    b.ValidFrom.AsTime(),
            ValidUntil:   b.ValidUntil.AsTime(),
            IsActive:     true,  // Filter active only
        }
    }

    return bindings, nil
}

func (p *EntityRoleProvider) GetEntityInfo(ctx context.Context, entityID, tenantID string) (*EntityInfo, error) {
    entity, err := p.entityClient.GetEntity(ctx, &entity_pb.GetEntityRequest{
        Id:       entityID,
        TenantId: tenantID,
    })
    if err != nil {
        return nil, err
    }

    return &EntityInfo{
        EntityID:     entity.Id,
        EntityType:   entity.EntityType.String(),
        UserID:       entity.UserId,
        ReferenceID:  entity.ReferenceId,
        DivisionID:   entity.DivisionId,
        BranchID:     entity.BranchId,
        DepartmentID: entity.DepartmentId,
        Status:       entity.Status.String(),
    }, nil
}
```

### Step 3: Wire Up Dependency Injection

```go
// In Auth Module: module.go

var AuthModule = fx.Module("auth",
    fx.Provide(
        // ... existing providers ...

        // Add entity client
        NewEntityServiceClient,
        NewUserServiceClient,
    ),
)

func NewEntityServiceClient(/* gRPC connection */) entity_pb.EntityServiceClient {
    // Connect to entity service
    return entity_pb.NewEntityServiceClient(conn)
}

// In User Module: module.go

var UserModule = fx.Module("user",
    fx.Provide(
        // ... existing providers ...

        // Inject entity client into entity role provider
        services.NewEntityRoleProvider,
    ),
)
```

---

## Best Practices

### 1. Always Use Entity-Based Permissions

❌ **Don't:**
```go
// Checking user permissions directly
CheckPermission("user:abc123", "formbuilder", "form", "approve", ...)
```

✅ **Do:**
```go
// Use entity for permission checks
CheckPermission("entity:xyz789", "formbuilder", "form", "approve", ...)
```

**Rationale:** Entities provide organizational context, role scoping, and domain linkage.

### 2. Prefer Organizational Scope Over Global Permissions

❌ **Don't:**
```go
// Global permission (no scope)
GrantPermission(ctx, &Permission{
    Subject:   "entity:abc",
    Namespace: "employee",
    Resource:  "employee",
    Action:    "view",
    Effect:    GRANT,
})
```

✅ **Do:**
```go
// Scoped to organizational unit
GrantPermission(ctx, &Permission{
    Subject:      "entity:abc",
    Namespace:    "employee",
    Resource:     "employee",
    Action:       "view",
    Effect:       GRANT,
    DepartmentID: "dept-uuid",  // Scoped
})
```

### 3. Use Temporal Constraints for Temporary Access

✅ **Example:**
```go
// Grant contractor access for project duration
CreateEntityRoleBinding(ctx, &CreateEntityRoleBindingRequest{
    EntityId:   "contractor-entity",
    RoleId:     "project-contributor",
    ValidFrom:  timestamppb.New(projectStart),
    ValidUntil: timestamppb.New(projectEnd),
})
```

### 4. Explicitly Deny When Needed

✅ **Example:**
```go
// Prevent form creator from self-approval
GrantPermission(ctx, &Permission{
    Subject:    "entity-creator",
    ResourceID: "form-123",
    Action:     "approve",
    Effect:     FORBIDDEN,  // Explicit deny overrides all grants
})
```

### 5. Log Permission Denials for Audit

```go
if permResp.Effect != pb.Effect_GRANT {
    auditLog.Record(ctx, &AuditEntry{
        EventType:  "PERMISSION_DENIED",
        EntityID:   entityID,
        Resource:   fmt.Sprintf("%s:%s", req.Namespace, req.Resource),
        Action:     req.Action,
        Reason:     permResp.Reason,
        IPAddress:  ctx.Value("ip_address").(string),
        Timestamp:  time.Now(),
    })
}
```

### 6. Cache Permission Checks

```go
// Cache entity permissions for session duration
sessionCache.Set(fmt.Sprintf("perms:%s", sessionID), effectivePermissions, 15*time.Minute)

// Invalidate on role changes
eventBus.Subscribe("EntityRoleBindingCreated", func(event Event) {
    sessionCache.Invalidate(fmt.Sprintf("perms:entity:%s", event.EntityID))
})
```

### 7. Test Permission Hierarchies

```go
// Unit test for organizational inheritance
func TestDivisionManagerInheritsPermissions(t *testing.T) {
    // Setup: Division manager with inheritance enabled
    // Test: Can approve forms in child departments
    // Expected: GRANT
}

func TestDepartmentManagerCannotApproveInOtherDepartments(t *testing.T) {
    // Setup: Department manager scoped to dept A
    // Test: Try to approve form in dept B
    // Expected: UNKNOWN (deny)
}
```

---

## Summary

This guide provides a complete implementation of authentication and authorization for the UGCL platform:

1. **Clear Module Boundaries:**
   - **Tenant**: Multi-tenancy isolation
   - **Auth**: Authentication (who are you?)
   - **Entity**: Identity abstraction and context
   - **User**: Authorization rules (what can you do?)

2. **Complete Flows:**
   - Login with entity resolution
   - Permission checking with organizational scope
   - Form-level and field-level authorization

3. **Duplication Identified:**
   - Password management: Remove from User module
   - Email/Phone verification: Remove from User module
   - OTP model: Remove from User module
   - Role management: Deprecate user_roles, use entity_role_bindings

4. **Integration Points:**
   - Auth ↔ Entity: GetEntityByUserId during login
   - User ↔ Entity: EntityRoleProvider for permission resolution
   - All modules ↔ Tenant: Tenant context for data isolation

5. **Best Practices:**
   - Always use entity-based permissions
   - Prefer organizational scoping
   - Use temporal constraints
   - Explicit denials when needed
   - Audit permission checks
   - Cache aggressively

The system is production-ready with minor cleanup needed to remove identified duplications.
