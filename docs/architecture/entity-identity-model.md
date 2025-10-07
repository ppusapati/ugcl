# UGCL Backend v2 - Entity-Identity Model Architecture

## Document Information

- **Version:** 2.0
- **Last Updated:** 2025-10-06
- **Status:** Active
- **Core Concept:** Universal Identity Abstraction for All Actors

---

## Table of Contents

1. [Entity Model Overview](#entity-model-overview)
2. [Problem Statement](#problem-statement)
3. [Entity Abstraction Architecture](#entity-abstraction-architecture)
4. [EntityType Enumeration](#entitytype-enumeration)
5. [Three-Layer Identity Model](#three-layer-identity-model)
6. [Entity Role Bindings](#entity-role-bindings)
7. [Permission Resolution](#permission-resolution)
8. [Polymorphic Relationships](#polymorphic-relationships)
9. [Integration Examples](#integration-examples)
10. [Use Cases](#use-cases)
11. [Migration Path](#migration-path)
12. [Best Practices](#best-practices)

---

## Entity Model Overview

### The Vision

The Entity model provides a **unified identity abstraction** that allows the UGCL platform to treat all actors (humans, machines, systems) consistently for authentication, authorization, and resource ownership.

```
Traditional Approach (Fragmented):
┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐
│ Employee │   │Contractor│   │  Vendor  │   │  Device  │
│ Identity │   │ Identity │   │ Identity │   │ Identity │
└──────────┘   └──────────┘   └──────────┘   └──────────┘
     │              │              │              │
     └──────────────┴──────────────┴──────────────┘
                    │
            Different permission systems,
            duplicate code, inconsistent behavior

Entity Approach (Unified):
                    ┌──────────┐
                    │  Entity  │
                    │(Universal)│
                    └─────┬────┘
                          │
         ┌────────────────┼────────────────┐
         │                │                │
    ┌────▼────┐     ┌─────▼─────┐    ┌────▼────┐
    │Employee │     │Contractor │    │ Device  │
    │ (Domain)│     │ (Domain)  │    │(Domain) │
    └─────────┘     └───────────┘    └─────────┘

Single permission system, shared code, consistent behavior
```

### Core Benefits

1. **Unified Authentication:** All actors authenticate through the same User → Entity model
2. **Consistent Authorization:** Single permission resolution engine for all entity types
3. **Polymorphic Ownership:** Documents, forms, projects can be owned by any entity
4. **Organizational Scoping:** Permissions can be scoped to division/branch/department
5. **Temporal Access:** Time-bound role assignments (valid_from, valid_until)
6. **Future-Proof:** Easy to add new entity types (drones, bots, AI agents)
7. **Audit Trail:** Unified tracking of all actor activities

---

## Problem Statement

### Before Entity Model

**Challenges:**
1. **Fragmented Identity:** Separate authentication for employees, contractors, vendors
2. **Duplicate Code:** Permission logic repeated for each actor type
3. **Inconsistent Behavior:** Different authorization rules for different actors
4. **Ownership Complexity:** Documents/forms need separate owner_employee_id, owner_contractor_id, etc.
5. **Non-Human Actors:** No clear way to handle devices, bots, systems
6. **Organizational Context:** Hard to scope permissions to organizational units

**Example of the Problem:**
```sql
-- Before: Documents table with multiple owner columns
CREATE TABLE documents (
    id UUID PRIMARY KEY,
    owner_employee_id UUID REFERENCES employees(id),      -- Only for employees
    owner_contractor_id UUID REFERENCES contractors(id),  -- Only for contractors
    owner_vendor_id UUID REFERENCES vendors(id),          -- Only for vendors
    ...
);

-- Query becomes complex
SELECT * FROM documents
WHERE (
    owner_employee_id = $1 OR
    owner_contractor_id = $1 OR
    owner_vendor_id = $1
);

-- Adding new actor type (device) requires:
-- 1. New column in documents
-- 2. New foreign key
-- 3. Update all queries
-- 4. Update application logic
```

### After Entity Model

**Solution:**
```sql
-- After: Single polymorphic owner
CREATE TABLE documents (
    id UUID PRIMARY KEY,
    owner_entity_id UUID REFERENCES entities(id),  -- Works for ALL actors
    owner_entity_type VARCHAR(50),                 -- 'EMPLOYEE', 'CONTRACTOR', 'DEVICE', etc.
    ...
);

-- Simple query
SELECT * FROM documents
WHERE owner_entity_id = $1;

-- Adding new actor type (device):
-- 1. Add 'DEVICE' to entity_type enum
-- 2. Create devices table
-- 3. Create entities with reference_source='DEVICE'
-- No changes to documents table or queries!
```

---

## Entity Abstraction Architecture

### Entity Model

```
┌─────────────────────────────────────────────────────────────────┐
│                         Entity Model                             │
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │ Entity (Universal Identity)                                 │ │
│  │                                                              │ │
│  │ • id (UUID)                                                  │ │
│  │ • tenant_id (UUID)                                          │ │
│  │ • entity_type (ENUM: EMPLOYEE, CONTRACTOR, DEVICE, etc.)   │ │
│  │ • user_id (UUID) → Links to users.uuid                     │ │
│  │ • reference_id (UUID) → Points to domain record            │ │
│  │ • reference_source (ENUM: EMPLOYEE, CONTRACTOR, etc.)      │ │
│  │ • status (ENUM: ACTIVE, INACTIVE, SUSPENDED, ARCHIVED)     │ │
│  │                                                              │ │
│  │ • division_id (UUID) → Organizational context              │ │
│  │ • branch_id (UUID)                                         │ │
│  │ • department_id (UUID)                                     │ │
│  │                                                              │ │
│  │ • metadata (JSONB) → Extensible attributes                 │ │
│  │ • created_at, updated_at, created_by, updated_by           │ │
│  └────────────────────────────────────────────────────────────┘ │
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │ EntityRoleBinding (Role Assignment with Scoping)           │ │
│  │                                                              │ │
│  │ • id (UUID)                                                  │ │
│  │ • entity_id (UUID) → References entity                      │ │
│  │ • role_id (UUID) → References role                          │ │
│  │                                                              │ │
│  │ • division_id (UUID) → Optional scope limitation           │ │
│  │ • branch_id (UUID)                                         │ │
│  │ • department_id (UUID)                                     │ │
│  │                                                              │ │
│  │ • valid_from (TIMESTAMP) → Start of access period          │ │
│  │ • valid_until (TIMESTAMP) → End of access period           │ │
│  │                                                              │ │
│  │ • created_at, updated_at, created_by, updated_by           │ │
│  └────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

### Database Schema

```sql
-- Entity types enumeration
CREATE TYPE entity_type AS ENUM (
    'EMPLOYEE',     -- Human employee
    'CONTRACTOR',   -- External contractor
    'VENDOR',       -- Supplier/vendor organization
    'CLIENT',       -- Customer/client
    'DEVICE',       -- IoT device, sensor
    'DRONE',        -- Unmanned aerial vehicle
    'BOT',          -- Software bot, automation
    'SYSTEM',       -- System account
    'ADMIN',        -- Administrator account
    'AGENT'         -- AI agent (future)
);

-- Entity status enumeration
CREATE TYPE entity_status AS ENUM (
    'ACTIVE',       -- Currently active
    'INACTIVE',     -- Temporarily inactive
    'SUSPENDED',    -- Access suspended
    'ARCHIVED'      -- Permanently archived
);

-- Reference source enumeration (where reference_id points)
CREATE TYPE reference_source AS ENUM (
    'EMPLOYEE',     -- employees.id
    'CONTRACTOR',   -- contractors.id
    'VENDOR',       -- vendors.id
    'CLIENT',       -- clients.id
    'DEVICE',       -- devices.id
    'DRONE',        -- drones.id
    'BOT'           -- bots.id
);

-- Entities table
CREATE TABLE entities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,

    -- Entity classification
    entity_type entity_type NOT NULL,
    status entity_status NOT NULL DEFAULT 'ACTIVE',

    -- Identity linkage
    user_id UUID NOT NULL,  -- Links to users.uuid (for authentication)
    reference_id UUID NOT NULL,  -- Points to domain-specific record
    reference_source reference_source NOT NULL,

    -- Organizational context
    division_id UUID,
    branch_id UUID,
    department_id UUID,

    -- Metadata for extensibility
    metadata JSONB DEFAULT '{}',

    -- Audit fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,

    -- Constraints
    CONSTRAINT unique_user_per_tenant UNIQUE(tenant_id, user_id),
    CONSTRAINT unique_reference_per_tenant UNIQUE(tenant_id, reference_id, reference_source)
);

-- Indexes
CREATE INDEX idx_entities_tenant_id ON entities(tenant_id);
CREATE INDEX idx_entities_user_id ON entities(user_id);
CREATE INDEX idx_entities_entity_type ON entities(entity_type);
CREATE INDEX idx_entities_status ON entities(status);
CREATE INDEX idx_entities_reference ON entities(reference_id, reference_source);
CREATE INDEX idx_entities_division ON entities(division_id) WHERE division_id IS NOT NULL;
CREATE INDEX idx_entities_branch ON entities(branch_id) WHERE branch_id IS NOT NULL;
CREATE INDEX idx_entities_department ON entities(department_id) WHERE department_id IS NOT NULL;
CREATE INDEX idx_entities_metadata ON entities USING GIN(metadata);

-- Comments for documentation
COMMENT ON TABLE entities IS 'Unified identity abstraction for all actors (human, machine, system)';
COMMENT ON COLUMN entities.user_id IS 'Links to users table for authentication';
COMMENT ON COLUMN entities.reference_id IS 'Points to domain-specific record (employee_id, device_id, etc.)';
COMMENT ON COLUMN entities.reference_source IS 'Indicates which domain service owns the reference_id';
COMMENT ON COLUMN entities.metadata IS 'Extensible JSON field for entity-specific attributes';
```

### Entity Role Bindings Schema

```sql
-- Entity role bindings with organizational and temporal scoping
CREATE TABLE entity_role_bindings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    entity_id UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    role_id UUID NOT NULL,  -- Links to roles.uuid

    -- Optional scope limitations (NULL = global within tenant)
    division_id UUID,
    branch_id UUID,
    department_id UUID,

    -- Time-bound access (NULL = permanent)
    valid_from TIMESTAMPTZ,
    valid_until TIMESTAMPTZ,

    -- Audit fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,

    -- Constraints
    CONSTRAINT unique_entity_role_scope UNIQUE(
        tenant_id, entity_id, role_id,
        division_id, branch_id, department_id
    )
);

-- Indexes
CREATE INDEX idx_entity_role_bindings_tenant_id ON entity_role_bindings(tenant_id);
CREATE INDEX idx_entity_role_bindings_entity_id ON entity_role_bindings(entity_id);
CREATE INDEX idx_entity_role_bindings_role_id ON entity_role_bindings(role_id);
CREATE INDEX idx_entity_role_bindings_division ON entity_role_bindings(division_id) WHERE division_id IS NOT NULL;
CREATE INDEX idx_entity_role_bindings_branch ON entity_role_bindings(branch_id) WHERE branch_id IS NOT NULL;
CREATE INDEX idx_entity_role_bindings_department ON entity_role_bindings(department_id) WHERE department_id IS NOT NULL;
CREATE INDEX idx_entity_role_bindings_validity ON entity_role_bindings(valid_from, valid_until)
    WHERE valid_from IS NOT NULL OR valid_until IS NOT NULL;

COMMENT ON TABLE entity_role_bindings IS 'Manages role assignments to entities with optional organizational and temporal scoping';
COMMENT ON COLUMN entity_role_bindings.valid_from IS 'Optional start date for time-bound role access';
COMMENT ON COLUMN entity_role_bindings.valid_until IS 'Optional end date for time-bound role access';
```

---

## EntityType Enumeration

### All Supported Entity Types

| Entity Type | Description | Use Case | Example |
|-------------|-------------|----------|---------|
| **EMPLOYEE** | Full-time employee | Internal staff with company benefits | Software Engineer, Manager, Accountant |
| **CONTRACTOR** | External contractor | Temporary workers on contract | Consultant, Freelancer, Temp Worker |
| **VENDOR** | Supplier/vendor organization | External companies providing services | Equipment Supplier, Logistics Provider |
| **CLIENT** | Customer/client | External customers using services | Dairy Farm Owner, Cooperative |
| **DEVICE** | IoT device | Hardware devices (sensors, gateways) | Temperature Sensor, Gateway Device |
| **DRONE** | Unmanned aerial vehicle | Autonomous flying devices | Survey Drone, Delivery Drone |
| **BOT** | Software bot | Automated software agents | Slack Bot, ETL Bot, Backup Bot |
| **SYSTEM** | System account | Internal system services | Migration Service, Analytics Engine |
| **ADMIN** | Administrator | Super admin accounts | Root Admin, System Administrator |
| **AGENT** | AI agent (future) | AI-powered autonomous agents | ChatGPT Integration, ML Model |

### Entity Type Characteristics

```go
// Entity type metadata (stored in entities.metadata)
type EntityTypeMetadata struct {
    // Human vs Machine
    IsHuman   bool
    IsMachine bool

    // Authentication method
    RequiresPassword     bool
    RequiresAPIKey       bool
    RequiresCertificate  bool

    // Capabilities
    CanOwnDocuments     bool
    CanSubmitForms      bool
    CanApprove          bool
    CanGenerateReports  bool

    // Organizational
    RequiresDivision    bool
    RequiresBranch      bool
    RequiresDepartment  bool
}

// Example metadata for different types
var EntityTypeConfigs = map[EntityType]EntityTypeMetadata{
    ENTITY_TYPE_EMPLOYEE: {
        IsHuman:             true,
        RequiresPassword:    true,
        CanOwnDocuments:     true,
        CanSubmitForms:      true,
        CanApprove:          true,
        RequiresDivision:    true,
        RequiresBranch:      true,
    },
    ENTITY_TYPE_DEVICE: {
        IsMachine:           true,
        RequiresAPIKey:      true,
        CanOwnDocuments:     false,
        CanSubmitForms:      false,
        CanGenerateReports:  true,
        RequiresBranch:      true,
    },
    ENTITY_TYPE_BOT: {
        IsMachine:           true,
        RequiresAPIKey:      true,
        CanOwnDocuments:     false,
        CanSubmitForms:      false,
        CanGenerateReports:  true,
    },
}
```

---

## Three-Layer Identity Model

### Layer 1: User (Authentication)

**Purpose:** Authentication credentials and basic profile

```sql
CREATE TABLE users (
    id              BIGSERIAL PRIMARY KEY,
    uuid            UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    username        TEXT,
    email           TEXT,
    password_hash   TEXT NOT NULL,
    is_active       BOOLEAN DEFAULT true,
    ...
);
```

**Responsibilities:**
- Store authentication credentials
- Password management
- Two-factor authentication
- Session management
- Login/logout

**Key Point:** Users are NOT directly associated with business domains. They are pure authentication records.

### Layer 2: Entity (Universal Identity)

**Purpose:** Universal abstraction linking users to business domains

```sql
CREATE TABLE entities (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    entity_type entity_type NOT NULL,
    user_id UUID NOT NULL,  -- → users.uuid
    reference_id UUID NOT NULL,  -- → domain record
    reference_source reference_source NOT NULL,
    ...
);
```

**Responsibilities:**
- Link User to domain-specific record (Employee, Device, etc.)
- Store organizational context (division, branch, department)
- Manage entity lifecycle (active, inactive, suspended)
- Provide unified identity for permissions
- Enable polymorphic relationships

**Key Point:** Entities are the "glue" between authentication (User) and business domains (Employee, Device, etc.).

### Layer 3: Domain Record (Business Context)

**Purpose:** Domain-specific business data

```sql
-- Example: Employee domain
CREATE TABLE employees (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    employee_code VARCHAR(50) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    hire_date DATE NOT NULL,
    job_title VARCHAR(100),
    salary NUMERIC(15, 2),
    ...
);

-- Example: Device domain
CREATE TABLE devices (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    device_code VARCHAR(50) NOT NULL,
    device_type VARCHAR(50) NOT NULL,
    manufacturer VARCHAR(100),
    model VARCHAR(100),
    serial_number VARCHAR(100),
    installed_at TIMESTAMPTZ,
    ...
);
```

**Responsibilities:**
- Store domain-specific business data
- Enforce domain-specific business rules
- Maintain domain-specific relationships
- No authentication or permission logic

**Key Point:** Domain records are independent of identity/permissions. They focus purely on business logic.

### Complete Relationship Flow

```
User (Authentication Layer)
    │
    │ user_id
    ▼
Entity (Identity Abstraction Layer)
    │
    │ reference_id + reference_source
    ▼
Employee / Contractor / Device / Bot (Domain Layer)


Example Flow:

1. John Doe logs in
   → Authenticates as User(uuid=user-123, email=john@example.com)

2. System looks up Entity
   → Entity(id=entity-456, user_id=user-123, entity_type=EMPLOYEE, reference_id=emp-789)

3. System fetches business data
   → Employee(id=emp-789, employee_code=EMP001, job_title=Engineer)

4. System resolves permissions
   → EntityRoleBinding(entity_id=entity-456, role_id=engineer-role)
   → Permissions based on engineer-role

5. John can now:
   - Own documents (as entity-456)
   - Submit forms (as entity-456)
   - View resources (based on engineer-role permissions)
```

### Visual Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                     Complete Identity Flow                       │
└─────────────────────────────────────────────────────────────────┘

User Authentication:
┌──────────────┐
│    User      │
│              │
│ uuid: u-123  │
│ email: john@ │
│ password: ****│
└──────┬───────┘
       │
       │ Links to
       ▼
┌──────────────┐
│   Entity     │
│              │
│ id: e-456    │
│ user_id: u-123│ ◄──── Universal Identity
│ type: EMPLOYEE│
│ ref_id: em-789│
│ division: d-1 │
│ branch: b-2   │
└──────┬───────┘
       │
       │ Points to
       ▼
┌──────────────┐
│  Employee    │
│              │
│ id: em-789   │
│ code: EMP001 │ ◄──── Domain-Specific Data
│ title: Eng.  │
│ salary: 100K │
└──────────────┘

Permission Resolution:
┌──────────────┐
│   Entity     │
│ id: e-456    │
└──────┬───────┘
       │
       │ Has roles
       ▼
┌──────────────────┐
│EntityRoleBinding │
│                  │
│ entity_id: e-456 │
│ role_id: r-admin │ ◄──── Role Assignment with Scope
│ division: d-1    │
│ valid_from: now  │
│ valid_until: +1y │
└──────┬───────────┘
       │
       │ Materializes to
       ▼
┌──────────────────┐
│  Permissions     │
│                  │
│ document:read    │ ◄──── Actual Permissions
│ document:write   │
│ form:submit      │
│ form:approve     │
└──────────────────┘
```

---

## Entity Role Bindings

### Role Assignment with Organizational Scoping

Traditional RBAC assigns roles globally:
```
User → Role → Permissions (global)
```

Entity-based RBAC supports organizational scoping:
```
Entity → Role (scoped to division/branch/dept) → Permissions (limited to scope)
```

### Scoping Examples

#### Example 1: Division-Scoped Manager

```sql
-- John is Manager for Division A only
INSERT INTO entity_role_bindings (
    tenant_id,
    entity_id,
    role_id,
    division_id,
    branch_id,
    department_id,
    valid_from,
    valid_until
) VALUES (
    'tenant-123',
    'entity-john',
    'role-manager',
    'division-a',  -- Scoped to Division A
    NULL,          -- All branches in Division A
    NULL,          -- All departments in Division A
    '2024-01-01',
    '2025-01-01'   -- Valid for 1 year
);

-- Permission check: Can John approve documents in Division A?
-- YES - he's a manager in Division A

-- Permission check: Can John approve documents in Division B?
-- NO - his manager role is scoped to Division A only
```

#### Example 2: Branch-Scoped Supervisor

```sql
-- Sarah is Supervisor for Branch B1 in Division A
INSERT INTO entity_role_bindings (
    tenant_id,
    entity_id,
    role_id,
    division_id,
    branch_id,
    department_id
) VALUES (
    'tenant-123',
    'entity-sarah',
    'role-supervisor',
    'division-a',
    'branch-b1',  -- Scoped to specific branch
    NULL          -- All departments in this branch
);

-- Permission check: Can Sarah supervise employees in Branch B1?
-- YES

-- Permission check: Can Sarah supervise employees in Branch B2?
-- NO - her supervisor role is scoped to Branch B1 only
```

#### Example 3: Department-Scoped Specialist

```sql
-- Mike is QA Specialist for Quality Control dept in Branch B1
INSERT INTO entity_role_bindings (
    tenant_id,
    entity_id,
    role_id,
    division_id,
    branch_id,
    department_id
) VALUES (
    'tenant-123',
    'entity-mike',
    'role-qa-specialist',
    'division-a',
    'branch-b1',
    'dept-quality-control'  -- Scoped to specific department
);

-- Permission check: Can Mike access QC documents in Quality Control dept?
-- YES

-- Permission check: Can Mike access QC documents in Production dept?
-- NO - his role is scoped to Quality Control dept only
```

#### Example 4: Global Role (No Scoping)

```sql
-- Admin has global role (no organizational scoping)
INSERT INTO entity_role_bindings (
    tenant_id,
    entity_id,
    role_id,
    division_id,
    branch_id,
    department_id
) VALUES (
    'tenant-123',
    'entity-admin',
    'role-super-admin',
    NULL,  -- No division limitation
    NULL,  -- No branch limitation
    NULL   -- No department limitation
);

-- Permission check: Can Admin access resources anywhere?
-- YES - global access across all organizational units
```

### Temporal Scoping (Time-Bound Access)

```sql
-- Temporary contractor with time-limited access
INSERT INTO entity_role_bindings (
    tenant_id,
    entity_id,
    role_id,
    division_id,
    valid_from,
    valid_until
) VALUES (
    'tenant-123',
    'entity-contractor-123',
    'role-developer',
    'division-it',
    '2024-06-01 00:00:00',  -- Access starts June 1, 2024
    '2024-12-31 23:59:59'   -- Access ends December 31, 2024
);

-- Permission check on July 15, 2024:
-- YES - current time is within valid range

-- Permission check on January 5, 2025:
-- NO - access has expired
```

### Multiple Role Assignments

An entity can have multiple roles with different scopes:

```sql
-- Jane is both:
-- 1. Manager in Division A
-- 2. Finance Specialist globally
INSERT INTO entity_role_bindings (tenant_id, entity_id, role_id, division_id)
VALUES ('tenant-123', 'entity-jane', 'role-manager', 'division-a');

INSERT INTO entity_role_bindings (tenant_id, entity_id, role_id)
VALUES ('tenant-123', 'entity-jane', 'role-finance-specialist');

-- Permission resolution:
-- - In Division A: Has both manager AND finance specialist permissions
-- - In Division B: Has ONLY finance specialist permissions
```

---

## Permission Resolution

### Permission Resolution Algorithm

```go
func ResolvePermissions(ctx context.Context, entityID string, resource string, action string) (bool, error) {
    // 1. Get entity
    entity, err := entityRepo.GetByID(ctx, entityID)
    if err != nil {
        return false, err
    }

    // 2. Get current organizational context (from request)
    orgContext := OrganizationContextFromRequest(ctx)

    // 3. Get entity role bindings filtered by:
    //    - Organizational scope (division/branch/department)
    //    - Time bounds (valid_from, valid_until)
    bindings, err := entityRoleBindingRepo.GetActiveBindings(ctx, &GetBindingsParams{
        EntityID:     entityID,
        DivisionID:   orgContext.DivisionID,
        BranchID:     orgContext.BranchID,
        DepartmentID: orgContext.DepartmentID,
        CurrentTime:  time.Now(),
    })
    if err != nil {
        return false, err
    }

    // 4. Collect all role IDs
    roleIDs := make([]string, len(bindings))
    for i, binding := range bindings {
        roleIDs[i] = binding.RoleID
    }

    // 5. Resolve role hierarchy (include parent roles)
    allRoleIDs, err := roleRepo.GetRoleHierarchy(ctx, roleIDs)
    if err != nil {
        return false, err
    }

    // 6. Get permission definitions for all roles
    permDefs, err := permissionDefRepo.GetByRoles(ctx, allRoleIDs)
    if err != nil {
        return false, err
    }

    // 7. Materialize permissions
    for _, def := range permDefs {
        if def.Namespace == resource && def.Action == action {
            // Check scope
            if def.Scope == "*" || matchesScope(def.Scope, orgContext) {
                return true, nil  // GRANT
            }
        }
    }

    // 8. Check user-level overrides
    userOverride, err := userPermissionRepo.GetOverride(ctx, entity.UserID, resource, action)
    if err == nil && userOverride.Effect == "GRANT" {
        return true, nil  // User override grants access
    }
    if err == nil && userOverride.Effect == "FORBIDDEN" {
        return false, nil  // User override denies access
    }

    // 9. Default: DENY
    return false, nil
}
```

### Permission Resolution Flow Diagram

```
Client Request (with entity context)
    │
    ▼
┌─────────────────────────────────────┐
│ 1. Extract Entity ID from JWT       │
│    entity_id = "entity-456"         │
└────────────┬────────────────────────┘
             │
             ▼
┌─────────────────────────────────────┐
│ 2. Get Entity Record                │
│    SELECT * FROM entities           │
│    WHERE id = 'entity-456'          │
└────────────┬────────────────────────┘
             │
             ▼
┌─────────────────────────────────────┐
│ 3. Get Active Role Bindings         │
│    SELECT * FROM                    │
│    entity_role_bindings             │
│    WHERE entity_id = 'entity-456'   │
│      AND division_id = $division    │
│      AND branch_id = $branch        │
│      AND NOW() BETWEEN              │
│          valid_from AND valid_until │
└────────────┬────────────────────────┘
             │
             ▼
┌─────────────────────────────────────┐
│ 4. Resolve Role Hierarchy           │
│    Get parent roles recursively     │
│    role-manager → role-supervisor   │
│                → role-user          │
└────────────┬────────────────────────┘
             │
             ▼
┌─────────────────────────────────────┐
│ 5. Get Permission Definitions       │
│    SELECT * FROM                    │
│    role_permission_defs             │
│    WHERE role_id IN (...)           │
└────────────┬────────────────────────┘
             │
             ▼
┌─────────────────────────────────────┐
│ 6. Materialize Permissions          │
│    SELECT * FROM permissions        │
│    WHERE def_name IN (...)          │
│      AND namespace = 'document'     │
│      AND resource = 'file'          │
│      AND action = 'read'            │
└────────────┬────────────────────────┘
             │
             ▼
┌─────────────────────────────────────┐
│ 7. Apply Scope Filter               │
│    If scope = '*': GRANT            │
│    If scope = 'own': Check ownership│
│    If scope = 'branch': Check branch│
└────────────┬────────────────────────┘
             │
             ▼
┌─────────────────────────────────────┐
│ 8. Check User Overrides             │
│    SELECT * FROM user_permissions   │
│    WHERE user_id = ...              │
│      AND permission_id = ...        │
└────────────┬────────────────────────┘
             │
             ▼
┌─────────────────────────────────────┐
│ 9. Return Decision                  │
│    GRANT / DENY                     │
└─────────────────────────────────────┘
```

### Example: Document Access Permission

```go
// Check if entity can read a specific document
func (s *DMSService) CanAccessDocument(ctx context.Context, entityID string, documentID string) (bool, error) {
    // 1. Get document
    doc, err := s.docRepo.GetByID(ctx, documentID)
    if err != nil {
        return false, err
    }

    // 2. Get entity
    entity, err := s.entityService.GetEntity(ctx, entityID)
    if err != nil {
        return false, err
    }

    // 3. Check ownership (scope: 'own')
    if doc.OwnerEntityID == entityID {
        // Entity owns the document
        return s.permissionService.CheckPermission(ctx, entityID, "document", "file", "read:own")
    }

    // 4. Check branch-level access (scope: 'branch')
    if doc.BranchID == entity.BranchID {
        return s.permissionService.CheckPermission(ctx, entityID, "document", "file", "read:branch")
    }

    // 5. Check division-level access (scope: 'division')
    if doc.DivisionID == entity.DivisionID {
        return s.permissionService.CheckPermission(ctx, entityID, "document", "file", "read:division")
    }

    // 6. Check global access (scope: '*')
    return s.permissionService.CheckPermission(ctx, entityID, "document", "file", "read:all")
}
```

### Permission Scopes

| Scope | Description | Use Case |
|-------|-------------|----------|
| **own** | Only own resources | Employee can read their own documents |
| **branch** | Resources in same branch | Manager can read all documents in their branch |
| **division** | Resources in same division | Director can read all documents in their division |
| **department** | Resources in same department | QA can read all QC documents |
| ***** (all) | Global access | Admin can read all documents |

---

## Polymorphic Relationships

### Document Ownership (Polymorphic)

```sql
-- Documents can be owned by ANY entity type
CREATE TABLE documents (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,

    -- Polymorphic ownership
    owner_entity_id UUID NOT NULL,  -- References entities.id
    owner_entity_type VARCHAR(50),  -- 'EMPLOYEE', 'CONTRACTOR', 'DEVICE', etc.

    file_name VARCHAR(500) NOT NULL,
    ...
);

-- Example: Employee-owned document
INSERT INTO documents (id, tenant_id, owner_entity_id, owner_entity_type, file_name)
VALUES ('doc-1', 'tenant-123', 'entity-employee-1', 'EMPLOYEE', 'report.pdf');

-- Example: Device-owned document (sensor data)
INSERT INTO documents (id, tenant_id, owner_entity_id, owner_entity_type, file_name)
VALUES ('doc-2', 'tenant-123', 'entity-device-1', 'DEVICE', 'sensor-data.csv');

-- Example: Bot-owned document (generated report)
INSERT INTO documents (id, tenant_id, owner_entity_id, owner_entity_type, file_name)
VALUES ('doc-3', 'tenant-123', 'entity-bot-1', 'BOT', 'analytics-report.pdf');

-- Query: Get all documents owned by an entity (regardless of type)
SELECT * FROM documents
WHERE owner_entity_id = 'entity-456';

-- Query: Get document with owner details
SELECT
    d.id,
    d.file_name,
    e.entity_type,
    u.username,
    u.fullname
FROM documents d
INNER JOIN entities e ON e.id = d.owner_entity_id
INNER JOIN users u ON u.uuid = e.user_id
WHERE d.tenant_id = 'tenant-123';
```

### Form Submissions (Polymorphic)

```sql
-- Forms can be submitted by any entity
CREATE TABLE form_instances (
    id UUID PRIMARY KEY,
    form_id UUID NOT NULL,

    -- Polymorphic creator
    created_by_entity_id UUID NOT NULL,  -- References entities.id
    created_by_entity_type VARCHAR(50),

    -- Polymorphic assignee
    assigned_to_entity_id UUID,
    assigned_to_entity_type VARCHAR(50),

    current_state VARCHAR(100),
    field_values JSONB,
    ...
);

-- Example: Employee submits form
INSERT INTO form_instances (id, form_id, created_by_entity_id, created_by_entity_type)
VALUES ('inst-1', 'form-1', 'entity-employee-1', 'EMPLOYEE');

-- Example: System bot auto-submits form
INSERT INTO form_instances (id, form_id, created_by_entity_id, created_by_entity_type)
VALUES ('inst-2', 'form-2', 'entity-bot-1', 'BOT');

-- Query: Get all forms submitted by entity
SELECT * FROM form_instances
WHERE created_by_entity_id = 'entity-456';
```

### Project Teams (Polymorphic)

```sql
-- Project team members can be employees, contractors, or even bots
CREATE TABLE project_team_members (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL,

    -- Polymorphic team member
    member_entity_id UUID NOT NULL,
    member_entity_type VARCHAR(50),

    role VARCHAR(100),  -- 'Lead', 'Developer', 'QA', 'Bot'
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    ...
);

-- Example: Mixed team
-- Employee as lead
INSERT INTO project_team_members (project_id, member_entity_id, member_entity_type, role)
VALUES ('proj-1', 'entity-employee-1', 'EMPLOYEE', 'Lead');

-- Contractor as developer
INSERT INTO project_team_members (project_id, member_entity_id, member_entity_type, role)
VALUES ('proj-1', 'entity-contractor-1', 'CONTRACTOR', 'Developer');

-- Bot for automation
INSERT INTO project_team_members (project_id, member_entity_id, member_entity_type, role)
VALUES ('proj-1', 'entity-bot-1', 'BOT', 'CI/CD Bot');

-- Query: Get all team members with details
SELECT
    ptm.role,
    e.entity_type,
    u.username,
    CASE e.entity_type
        WHEN 'EMPLOYEE' THEN emp.fullname
        WHEN 'CONTRACTOR' THEN con.fullname
        WHEN 'BOT' THEN b.name
    END as name
FROM project_team_members ptm
INNER JOIN entities e ON e.id = ptm.member_entity_id
INNER JOIN users u ON u.uuid = e.user_id
LEFT JOIN employees emp ON emp.id = e.reference_id AND e.reference_source = 'EMPLOYEE'
LEFT JOIN contractors con ON con.id = e.reference_id AND e.reference_source = 'CONTRACTOR'
LEFT JOIN bots b ON b.id = e.reference_id AND e.reference_source = 'BOT'
WHERE ptm.project_id = 'proj-1';
```

---

## Integration Examples

### Example 1: Employee Onboarding (Creating Entity)

```go
func (s *EmployeeService) OnboardEmployee(ctx context.Context, req *OnboardEmployeeRequest) error {
    tx, _ := s.db.Begin(ctx)
    defer tx.Rollback(ctx)

    // 1. Create User (authentication layer)
    user, err := s.userRepo.CreateWithTx(ctx, tx, &User{
        Username:     req.Email,
        Email:        req.Email,
        PasswordHash: hashPassword(req.Password),
        FullName:     req.FullName,
        Phone:        req.Phone,
    })
    if err != nil {
        return fmt.Errorf("failed to create user: %w", err)
    }

    // 2. Create Employee (domain layer)
    employee, err := s.employeeRepo.CreateWithTx(ctx, tx, &Employee{
        TenantID:     req.TenantID,
        EmployeeCode: generateEmployeeCode(),
        FirstName:    req.FirstName,
        LastName:     req.LastName,
        HireDate:     req.HireDate,
        JobTitle:     req.JobTitle,
        DivisionID:   req.DivisionID,
        BranchID:     req.BranchID,
        DepartmentID: req.DepartmentID,
    })
    if err != nil {
        return fmt.Errorf("failed to create employee: %w", err)
    }

    // 3. Create Entity (identity abstraction layer)
    entity, err := s.entityRepo.CreateWithTx(ctx, tx, &Entity{
        TenantID:        req.TenantID,
        EntityType:      ENTITY_TYPE_EMPLOYEE,
        UserID:          user.UUID,
        ReferenceID:     employee.ID,
        ReferenceSource: REFERENCE_SOURCE_EMPLOYEE,
        Status:          ENTITY_STATUS_ACTIVE,
        DivisionID:      req.DivisionID,
        BranchID:        req.BranchID,
        DepartmentID:    req.DepartmentID,
        CreatedBy:       getCurrentUserID(ctx),
        UpdatedBy:       getCurrentUserID(ctx),
    })
    if err != nil {
        return fmt.Errorf("failed to create entity: %w", err)
    }

    // 4. Assign Role (with organizational scope)
    _, err = s.entityRoleBindingRepo.CreateWithTx(ctx, tx, &EntityRoleBinding{
        TenantID:    req.TenantID,
        EntityID:    entity.ID,
        RoleID:      req.RoleID,
        DivisionID:  req.DivisionID,
        BranchID:    req.BranchID,
        DepartmentID: req.DepartmentID,
        ValidFrom:   timestamppb.Now(),
        ValidUntil:  nil,  // Permanent
        CreatedBy:   getCurrentUserID(ctx),
        UpdatedBy:   getCurrentUserID(ctx),
    })
    if err != nil {
        return fmt.Errorf("failed to assign role: %w", err)
    }

    // 5. Commit transaction
    return tx.Commit(ctx)
}
```

### Example 2: Device Registration (IoT Device)

```go
func (s *DeviceService) RegisterDevice(ctx context.Context, req *RegisterDeviceRequest) error {
    tx, _ := s.db.Begin(ctx)
    defer tx.Rollback(ctx)

    // 1. Create User (for authentication via API key)
    user, err := s.userRepo.CreateWithTx(ctx, tx, &User{
        Username:     req.DeviceCode,  // Use device code as username
        Email:        fmt.Sprintf("%s@device.ugcl.com", req.DeviceCode),
        PasswordHash: "",  // No password, will use API key
        IsActive:     true,
    })
    if err != nil {
        return fmt.Errorf("failed to create device user: %w", err)
    }

    // 2. Create Device (domain layer)
    device, err := s.deviceRepo.CreateWithTx(ctx, tx, &Device{
        TenantID:      req.TenantID,
        DeviceCode:    req.DeviceCode,
        DeviceType:    req.DeviceType,
        Manufacturer:  req.Manufacturer,
        Model:         req.Model,
        SerialNumber:  req.SerialNumber,
        BranchID:      req.BranchID,  // Device is installed at a branch
        InstalledAt:   timestamppb.Now(),
    })
    if err != nil {
        return fmt.Errorf("failed to create device: %w", err)
    }

    // 3. Create Entity
    entity, err := s.entityRepo.CreateWithTx(ctx, tx, &Entity{
        TenantID:        req.TenantID,
        EntityType:      ENTITY_TYPE_DEVICE,
        UserID:          user.UUID,
        ReferenceID:     device.ID,
        ReferenceSource: REFERENCE_SOURCE_DEVICE,
        Status:          ENTITY_STATUS_ACTIVE,
        BranchID:        req.BranchID,
        Metadata: map[string]interface{}{
            "device_type":   req.DeviceType,
            "serial_number": req.SerialNumber,
        },
        CreatedBy: getCurrentUserID(ctx),
        UpdatedBy: getCurrentUserID(ctx),
    })
    if err != nil {
        return fmt.Errorf("failed to create entity: %w", err)
    }

    // 4. Assign Role (e.g., "Data Collector" role)
    _, err = s.entityRoleBindingRepo.CreateWithTx(ctx, tx, &EntityRoleBinding{
        TenantID:  req.TenantID,
        EntityID:  entity.ID,
        RoleID:    "role-data-collector",
        BranchID:  req.BranchID,  // Scoped to installation branch
        CreatedBy: getCurrentUserID(ctx),
        UpdatedBy: getCurrentUserID(ctx),
    })
    if err != nil {
        return fmt.Errorf("failed to assign role: %w", err)
    }

    // 5. Generate API key for device authentication
    apiKey := generateAPIKey()
    err = s.apiKeyRepo.CreateWithTx(ctx, tx, &APIKey{
        UserID:    user.UUID,
        KeyHash:   hashAPIKey(apiKey),
        ExpiresAt: nil,  // Never expires
    })
    if err != nil {
        return fmt.Errorf("failed to create API key: %w", err)
    }

    // 6. Commit transaction
    if err := tx.Commit(ctx); err != nil {
        return err
    }

    // Return API key to caller (only time it's visible)
    log.Printf("Device registered. API Key: %s", apiKey)
    return nil
}
```

### Example 3: Bot Creation (Automation Agent)

```go
func (s *BotService) CreateBot(ctx context.Context, req *CreateBotRequest) error {
    tx, _ := s.db.Begin(ctx)
    defer tx.Rollback(ctx)

    // 1. Create User (for authentication)
    user, err := s.userRepo.CreateWithTx(ctx, tx, &User{
        Username:     req.BotName,
        Email:        fmt.Sprintf("%s@bot.ugcl.com", req.BotName),
        PasswordHash: "",  // No password, will use API key
        IsActive:     true,
    })
    if err != nil {
        return fmt.Errorf("failed to create bot user: %w", err)
    }

    // 2. Create Bot (domain layer)
    bot, err := s.botRepo.CreateWithTx(ctx, tx, &Bot{
        TenantID:    req.TenantID,
        BotName:     req.BotName,
        BotType:     req.BotType,  // 'OCR', 'ETL', 'Notification', etc.
        Description: req.Description,
        Config:      req.Config,
    })
    if err != nil {
        return fmt.Errorf("failed to create bot: %w", err)
    }

    // 3. Create Entity
    entity, err := s.entityRepo.CreateWithTx(ctx, tx, &Entity{
        TenantID:        req.TenantID,
        EntityType:      ENTITY_TYPE_BOT,
        UserID:          user.UUID,
        ReferenceID:     bot.ID,
        ReferenceSource: REFERENCE_SOURCE_BOT,
        Status:          ENTITY_STATUS_ACTIVE,
        Metadata: map[string]interface{}{
            "bot_type":    req.BotType,
            "capabilities": req.Capabilities,
        },
        CreatedBy: getCurrentUserID(ctx),
        UpdatedBy: getCurrentUserID(ctx),
    })
    if err != nil {
        return fmt.Errorf("failed to create entity: %w", err)
    }

    // 4. Assign Role (e.g., "System Bot" role with specific permissions)
    _, err = s.entityRoleBindingRepo.CreateWithTx(ctx, tx, &EntityRoleBinding{
        TenantID:  req.TenantID,
        EntityID:  entity.ID,
        RoleID:    req.RoleID,  // e.g., "role-system-bot"
        CreatedBy: getCurrentUserID(ctx),
        UpdatedBy: getCurrentUserID(ctx),
    })
    if err != nil {
        return fmt.Errorf("failed to assign role: %w", err)
    }

    // 5. Generate API key
    apiKey := generateAPIKey()
    err = s.apiKeyRepo.CreateWithTx(ctx, tx, &APIKey{
        UserID:    user.UUID,
        KeyHash:   hashAPIKey(apiKey),
        ExpiresAt: nil,
    })
    if err != nil {
        return fmt.Errorf("failed to create API key: %w", err)
    }

    // 6. Commit transaction
    if err := tx.Commit(ctx); err != nil {
        return err
    }

    log.Printf("Bot created. API Key: %s", apiKey)
    return nil
}
```

---

## Use Cases

### Use Case 1: Multi-Type Document Access

**Scenario:** A document can be accessed by employees, contractors, and automated bots for different purposes.

```sql
-- Document owned by employee
INSERT INTO documents (id, owner_entity_id, owner_entity_type, file_name)
VALUES ('doc-1', 'entity-emp-1', 'EMPLOYEE', 'quality-report.pdf');

-- Shared with contractor for review
INSERT INTO document_shares (document_id, shared_with_entity_id, can_view, can_download)
VALUES ('doc-1', 'entity-contractor-1', true, true);

-- OCR bot processes document
-- Bot has entity_type=BOT with permission "document:process"
-- Bot can read document, extract text, update OCR fields
UPDATE documents
SET ocr_status = 'COMPLETED',
    extracted_text = 'Quality metrics...',
    ocr_processed_by = 'entity-ocr-bot'  -- Bot entity
WHERE id = 'doc-1';

-- Query: Who has accessed this document?
SELECT
    e.entity_type,
    u.username,
    ds.can_view,
    ds.can_download,
    ds.last_accessed_at
FROM document_shares ds
INNER JOIN entities e ON e.id = ds.shared_with_entity_id
INNER JOIN users u ON u.uuid = e.user_id
WHERE ds.document_id = 'doc-1';
```

### Use Case 2: Drone Survey Data Collection

**Scenario:** Autonomous drones collect survey data and submit to the system.

```go
// Drone entity submits survey data
func (s *SurveyService) SubmitDroneSurvey(ctx context.Context, req *SubmitSurveyRequest) error {
    // 1. Get drone entity from API key
    droneEntity, err := s.entityService.GetEntityByUserId(ctx, req.DroneUserID)
    if err != nil {
        return err
    }

    // 2. Validate entity is a drone
    if droneEntity.EntityType != ENTITY_TYPE_DRONE {
        return errors.New("only drones can submit survey data")
    }

    // 3. Validate drone has permission
    hasPermission, err := s.permissionService.CheckPermission(
        ctx,
        droneEntity.ID,
        "survey",
        "data",
        "submit",
    )
    if err != nil || !hasPermission {
        return errors.New("drone lacks permission to submit survey data")
    }

    // 4. Create survey record
    survey := &Survey{
        SubmittedByEntityID:   droneEntity.ID,
        SubmittedByEntityType: "DRONE",
        SurveyType:           req.SurveyType,
        Data:                 req.Data,
        Location:             req.Location,
        Timestamp:            timestamppb.Now(),
    }

    // 5. Upload images to DMS (owned by drone)
    for _, image := range req.Images {
        doc, err := s.dmsService.UploadDocument(ctx, &UploadDocumentRequest{
            OwnerEntityID:   droneEntity.ID,
            OwnerEntityType: "DRONE",
            FileData:        image.Data,
            FileName:        image.FileName,
            MimeType:        "image/jpeg",
            Metadata: map[string]string{
                "survey_id": survey.ID,
                "drone_id":  droneEntity.ReferenceID,
            },
        })
        if err != nil {
            return err
        }

        survey.ImageDocumentIDs = append(survey.ImageDocumentIDs, doc.ID)
    }

    return s.surveyRepo.Create(ctx, survey)
}
```

### Use Case 3: Temporary Contractor Access

**Scenario:** A contractor is hired for 6 months with limited access.

```sql
-- 1. Create contractor entity
INSERT INTO entities (id, tenant_id, entity_type, user_id, reference_id, reference_source, status)
VALUES (
    'entity-contractor-123',
    'tenant-abc',
    'CONTRACTOR',
    'user-456',
    'contractor-789',
    'CONTRACTOR',
    'ACTIVE'
);

-- 2. Assign role with time limits and organizational scope
INSERT INTO entity_role_bindings (
    tenant_id,
    entity_id,
    role_id,
    division_id,
    branch_id,
    valid_from,
    valid_until
) VALUES (
    'tenant-abc',
    'entity-contractor-123',
    'role-developer',
    'division-it',
    'branch-hq',
    '2024-01-01 00:00:00',
    '2024-06-30 23:59:59'  -- 6-month contract
);

-- 3. After contract ends (July 1, 2024):
-- Permission check fails due to valid_until expiration
-- Contractor can no longer access resources

-- 4. Optional: Deactivate entity
UPDATE entities
SET status = 'ARCHIVED'
WHERE id = 'entity-contractor-123';
```

### Use Case 4: Cross-Department Collaboration

**Scenario:** An employee has different roles in different departments.

```sql
-- Employee is:
-- 1. Manager in Production Department
-- 2. QA Specialist in Quality Control Department

-- Manager role in Production
INSERT INTO entity_role_bindings (
    tenant_id,
    entity_id,
    role_id,
    division_id,
    branch_id,
    department_id
) VALUES (
    'tenant-123',
    'entity-emp-456',
    'role-manager',
    'division-manufacturing',
    'branch-factory-1',
    'dept-production'
);

-- QA Specialist role in Quality Control
INSERT INTO entity_role_bindings (
    tenant_id,
    entity_id,
    role_id,
    division_id,
    branch_id,
    department_id
) VALUES (
    'tenant-123',
    'entity-emp-456',
    'role-qa-specialist',
    'division-manufacturing',
    'branch-factory-1',
    'dept-quality-control'
);

-- Permission resolution:
-- When accessing Production resources:
--   → Uses Manager permissions
-- When accessing Quality Control resources:
--   → Uses QA Specialist permissions
```

---

## Migration Path

### Phase 1: Add Entity Tables

```sql
-- Add entity tables alongside existing tables
CREATE TABLE entities (...);
CREATE TABLE entity_role_bindings (...);

-- Existing tables remain unchanged
-- employees, contractors, vendors continue to work
```

### Phase 2: Populate Entities

```sql
-- Migrate existing employees to entities
INSERT INTO entities (tenant_id, entity_type, user_id, reference_id, reference_source, ...)
SELECT
    e.tenant_id,
    'EMPLOYEE'::entity_type,
    u.uuid,
    e.id,
    'EMPLOYEE'::reference_source,
    ...
FROM employees e
INNER JOIN users u ON u.id = e.user_id;

-- Migrate existing contractors
INSERT INTO entities (tenant_id, entity_type, user_id, reference_id, reference_source, ...)
SELECT
    c.tenant_id,
    'CONTRACTOR'::entity_type,
    u.uuid,
    c.id,
    'CONTRACTOR'::reference_source,
    ...
FROM contractors c
INNER JOIN users u ON u.id = c.user_id;

-- Migrate role assignments to entity_role_bindings
INSERT INTO entity_role_bindings (tenant_id, entity_id, role_id, ...)
SELECT
    e.tenant_id,
    e.id,
    ur.role_id,
    ...
FROM entities e
INNER JOIN users u ON u.uuid = e.user_id
INNER JOIN user_roles ur ON ur.user_id = u.id;
```

### Phase 3: Update Application Code

```go
// Old code (direct user-role mapping)
func GetUserPermissions(userID string) ([]Permission, error) {
    // ...
}

// New code (entity-based)
func GetEntityPermissions(entityID string, orgContext OrgContext) ([]Permission, error) {
    // Permission resolution with organizational scoping
    // ...
}

// Transition period: Support both
func GetPermissions(ctx context.Context, identifier string, identifierType string) ([]Permission, error) {
    if identifierType == "user" {
        // Legacy: Get entity from user, then permissions
        entity, _ := entityRepo.GetByUserID(ctx, identifier)
        return GetEntityPermissions(entity.ID, orgContext)
    }
    // New: Direct entity lookup
    return GetEntityPermissions(identifier, orgContext)
}
```

### Phase 4: Add Polymorphic Ownership

```sql
-- Add entity columns to existing tables
ALTER TABLE documents
ADD COLUMN owner_entity_id UUID REFERENCES entities(id),
ADD COLUMN owner_entity_type VARCHAR(50);

-- Backfill from existing data
UPDATE documents d
SET owner_entity_id = e.id,
    owner_entity_type = e.entity_type
FROM entities e
WHERE e.reference_id = d.owner_employee_id
  AND e.reference_source = 'EMPLOYEE';

-- After validation, drop old columns
ALTER TABLE documents
DROP COLUMN owner_employee_id,
DROP COLUMN owner_contractor_id;
```

### Phase 5: Deprecate Old Patterns

```go
// Mark old functions as deprecated
// @deprecated Use GetEntityPermissions instead
func GetUserPermissions(userID string) ([]Permission, error) {
    log.Warn("GetUserPermissions is deprecated, use GetEntityPermissions")
    // Fallback to new implementation
    entity, _ := entityRepo.GetByUserID(context.Background(), userID)
    return GetEntityPermissions(entity.ID, OrgContext{})
}
```

---

## Best Practices

### 1. Always Use Entity ID for Ownership

```go
// ✅ CORRECT: Use entity ID
type Document struct {
    OwnerEntityID   string
    OwnerEntityType string
}

// ❌ WRONG: Use user ID directly
type Document struct {
    OwnerUserID string  // Bypasses entity abstraction
}
```

### 2. Include Organizational Context in Requests

```go
// ✅ CORRECT: Pass organizational context
type CreateDocumentRequest struct {
    OwnerEntityID string
    DivisionID    string  // For scoping permissions
    BranchID      string
    DepartmentID  string
}

// ❌ WRONG: No organizational context
type CreateDocumentRequest struct {
    OwnerEntityID string  // Missing context
}
```

### 3. Validate Entity Type for Operations

```go
// ✅ CORRECT: Validate entity type
func (s *FormService) SubmitForm(ctx context.Context, req *SubmitFormRequest) error {
    entity, _ := s.entityService.GetEntity(ctx, req.SubmitterEntityID)

    // Only humans can submit forms
    if entity.EntityType != ENTITY_TYPE_EMPLOYEE && entity.EntityType != ENTITY_TYPE_CONTRACTOR {
        return errors.New("only employees and contractors can submit forms")
    }
    // ...
}

// ❌ WRONG: No validation
func (s *FormService) SubmitForm(ctx context.Context, req *SubmitFormRequest) error {
    // Allows devices, bots to submit forms (probably not desired)
    // ...
}
```

### 4. Use Scope Filters in Permission Checks

```go
// ✅ CORRECT: Check permission with scope
hasPermission, _ := s.permissionService.CheckPermissionWithScope(
    ctx,
    entityID,
    "document",
    "read",
    PermissionScope{
        DivisionID:  document.DivisionID,
        BranchID:    document.BranchID,
    },
)

// ❌ WRONG: No scope check
hasPermission, _ := s.permissionService.CheckPermission(
    ctx,
    entityID,
    "document",
    "read",
)
```

### 5. Handle Time-Bound Access

```go
// ✅ CORRECT: Include time validation
bindings, _ := s.entityRoleBindingRepo.GetActiveBindings(ctx, &GetBindingsParams{
    EntityID:    entityID,
    CurrentTime: time.Now(),  // Filter by valid_from/valid_until
})

// ❌ WRONG: Ignore time bounds
bindings, _ := s.entityRoleBindingRepo.GetAllBindings(ctx, entityID)
// May include expired bindings
```

### 6. Use Metadata for Extensibility

```go
// ✅ CORRECT: Store extra attributes in metadata
entity := &Entity{
    EntityType: ENTITY_TYPE_DEVICE,
    Metadata: map[string]interface{}{
        "device_type":    "temperature_sensor",
        "firmware_version": "1.2.3",
        "location":       "warehouse-a",
    },
}

// Easy to query
SELECT * FROM entities
WHERE metadata->>'device_type' = 'temperature_sensor';
```

### 7. Audit Entity Changes

```go
// ✅ CORRECT: Log entity lifecycle events
func (s *EntityService) DeactivateEntity(ctx context.Context, entityID string) error {
    entity, _ := s.entityRepo.GetByID(ctx, entityID)

    // Update status
    entity.Status = ENTITY_STATUS_INACTIVE
    err := s.entityRepo.Update(ctx, entity)

    // Audit log
    s.auditService.Log(ctx, &AuditLog{
        Action:       "ENTITY_DEACTIVATED",
        ResourceType: "entity",
        ResourceID:   entityID,
        EntityID:     getCurrentEntityID(ctx),
        Changes: map[string]interface{}{
            "status": map[string]string{
                "from": "ACTIVE",
                "to":   "INACTIVE",
            },
        },
    })

    return err
}
```

---

## Conclusion

The Entity-Identity Model is a powerful abstraction that enables the UGCL platform to:

1. **Unify Authentication:** Single authentication layer for all actors
2. **Simplify Permissions:** One permission system for humans, devices, bots
3. **Enable Polymorphism:** Documents, forms, projects owned by any entity type
4. **Support Organizations:** Scope permissions to divisions, branches, departments
5. **Time-Bound Access:** Temporary role assignments with expiration
6. **Future-Proof:** Easy to add new entity types without schema changes

This model provides the foundation for a flexible, scalable, and maintainable identity and access management system.

---

**Document Version:** 2.0
**Lines:** 1600+
**Generated:** 2025-10-06
**Maintained By:** UGCL Architecture Team
