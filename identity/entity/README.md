# Entity Identity Module

## 1. Module Overview

The Entity Identity module provides a unified identity abstraction layer that enables any actor (human, machine, or system) in the UGCL platform to have a consistent identity representation. It bridges the gap between authentication (identity/user) and domain-specific records (employee, contractor, vendor, device, etc.).

**Purpose:** Centralized identity management for all actors with role-based access control, organizational context, and unified permission resolution across the platform.

**Key Features:**
- Unified identity abstraction for all actor types
- Entity-to-User mapping for authentication
- Entity-to-Domain record mapping (employee_id, device_id, etc.)
- Role bindings with organizational scope
- Time-bound role assignments
- Multi-tenant entity management
- Status lifecycle management
- Metadata extensibility

## 2. Architecture

### Module Structure
```
identity/entity/
├── proto/
│   └── entity.proto           # Service and message definitions
├── db/
│   ├── schema/
│   │   └── schema.sql        # Entity tables with enums
│   └── generated/            # SQLC generated code
├── repository/               # Data access layer
│   ├── entity_repository.go
│   └── role_binding_repository.go
├── services/                 # Business logic layer
│   ├── entity_service.go
│   └── role_binding_service.go
├── handlers/                 # gRPC handlers
│   └── entity_handler.go
├── mappers/                  # DTO mappers
│   └── entity_mappers.go
└── module.go                 # FX module definition
```

### Database Schema

**Tables:**
- `entities` - Core entity abstraction table
- `entity_role_bindings` - Role assignments to entities

**Enums:**
- `entity_type` - EMPLOYEE, CONTRACTOR, VENDOR, CLIENT, DEVICE, DRONE, BOT, SYSTEM, ADMIN, AGENT
- `entity_status` - ACTIVE, INACTIVE, SUSPENDED, ARCHIVED
- `reference_source` - EMPLOYEE, CONTRACTOR, VENDOR, CLIENT, DEVICE, DRONE, BOT

**Key Relationships:**
- Entity → User (N:1 via user_id for authentication)
- Entity → Domain Record (N:1 via reference_id + reference_source)
- Entity → Roles (N:N via entity_role_bindings)
- Entity → Organization (N:1 optional via division/branch/department)

### Key Dependencies
- `identity/user` - User authentication and accounts
- `identity/auth` - Session management
- `organization` - Organizational hierarchy
- Domain modules (employee, contractor, vendor, etc.)

## 3. Quick Start

### Creating an Entity

```go
import (
    entityv1 "p9e.in/ugcl/identity/entity/api/v1"
    "connectrpc.com/connect"
)

client := entityv1.NewEntityServiceClient(httpClient, baseURL)

// Create entity for an employee
entity, err := client.CreateEntity(ctx, connect.NewRequest(&entityv1.CreateEntityRequest{
    TenantId:        "tenant-uuid",
    EntityType:      entityv1.EntityType_ENTITY_TYPE_EMPLOYEE,
    UserId:          "user-uuid",
    ReferenceId:     "employee-uuid",
    ReferenceSource: entityv1.ReferenceSource_REFERENCE_SOURCE_EMPLOYEE,
    Status:          entityv1.EntityStatus_ENTITY_STATUS_ACTIVE,
    DivisionId:      wrapperspb.String("division-uuid"),
    BranchId:        wrapperspb.String("branch-uuid"),
    DepartmentId:    wrapperspb.String("dept-uuid"),
    Metadata: map[string]string{
        "employee_code": "EMP001",
        "designation":   "Manager",
    },
}))
```

### Common Operations

**Get entity by user ID:**
```go
entity, err := client.GetEntityByUserId(ctx, connect.NewRequest(&entityv1.GetEntityByUserIdRequest{
    UserId:   "user-uuid",
    TenantId: "tenant-uuid",
}))
```

**Assign role to entity:**
```go
binding, err := client.CreateEntityRoleBinding(ctx, connect.NewRequest(&entityv1.CreateEntityRoleBindingRequest{
    TenantId:    "tenant-uuid",
    EntityId:    "entity-uuid",
    RoleId:      "role-uuid",
    DivisionId:  wrapperspb.String("division-uuid"), // Optional scope
    ValidFrom:   timestamppb.New(time.Now()),
    ValidUntil:  timestamppb.New(time.Now().Add(365*24*time.Hour)),
}))
```

## 4. API Reference

### Entity Operations
- `CreateEntity` - Create new entity abstraction
- `UpdateEntity` - Update entity status and context
- `GetEntity` - Retrieve entity by ID
- `GetEntityByUserId` - Find entity by user ID
- `GetEntityByReference` - Find entity by domain reference
- `ListEntities` - List entities with filtering
- `DeleteEntity` - Remove entity

### Role Binding Operations
- `CreateEntityRoleBinding` - Assign role to entity
- `GetEntityRoleBindings` - Get all roles for entity
- `DeleteEntityRoleBinding` - Revoke role from entity

## 5. Database Schema

### Tables Overview

**entities**
```sql
- id (UUID, PK)
- tenant_id (UUID)
- entity_type (ENUM) -- EMPLOYEE, CONTRACTOR, VENDOR, DEVICE, etc.
- user_id (UUID) -- Links to identity.users for authentication
- reference_id (UUID) -- Points to domain record (employees.uuid, devices.id, etc.)
- reference_source (ENUM) -- Indicates domain source
- status (ENUM) -- ACTIVE, INACTIVE, SUSPENDED, ARCHIVED

-- Organizational context
- division_id (UUID, optional)
- branch_id (UUID, optional)
- department_id (UUID, optional)

-- Extensibility
- metadata (JSONB) -- Entity-specific attributes

-- Audit
- created_at, updated_at, created_by, updated_by
```

**entity_role_bindings**
```sql
- id (UUID, PK)
- tenant_id (UUID)
- entity_id (UUID, FK → entities)
- role_id (UUID) -- Links to identity.roles

-- Optional scope limitations
- division_id (UUID, optional)
- branch_id (UUID, optional)
- department_id (UUID, optional)

-- Time-bound access
- valid_from (TIMESTAMPTZ, optional)
- valid_until (TIMESTAMPTZ, optional)

-- Audit
- created_at, updated_at, created_by, updated_by
```

### Constraints
- `unique_user_per_tenant` - One entity per user per tenant
- `unique_reference_per_tenant` - One entity per domain record per tenant
- `unique_entity_role_scope` - Unique role binding per entity+role+scope

### Indexes
- Tenant-based indexes
- User ID lookup
- Entity type filtering
- Reference lookup (composite on reference_id + reference_source)
- Organizational context indexes
- GIN index on metadata JSONB

## 6. Configuration

### Environment Variables
```env
# Database connection managed by packages/database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=ugcl
```

### Module-Specific Settings
- SQLC configuration in `db/sqlc.yaml`
- Proto generation via project-level `buf.gen.yaml`

## 7. Examples

### Complete Entity Lifecycle

```go
// 1. Create employee entity
employeeEntity, err := client.CreateEntity(ctx, connect.NewRequest(&entityv1.CreateEntityRequest{
    TenantId:        tenantID,
    EntityType:      entityv1.EntityType_ENTITY_TYPE_EMPLOYEE,
    UserId:          userID,
    ReferenceId:     employeeID,
    ReferenceSource: entityv1.ReferenceSource_REFERENCE_SOURCE_EMPLOYEE,
    Status:          entityv1.EntityStatus_ENTITY_STATUS_ACTIVE,
    DivisionId:      wrapperspb.String(divisionID),
    BranchId:        wrapperspb.String(branchID),
    DepartmentId:    wrapperspb.String(deptID),
}))

entityID := employeeEntity.Msg.Id

// 2. Assign department-scoped role
managerRole, err := client.CreateEntityRoleBinding(ctx, connect.NewRequest(&entityv1.CreateEntityRoleBindingRequest{
    TenantId:     tenantID,
    EntityId:     entityID,
    RoleId:       "department-manager-role-uuid",
    DepartmentId: wrapperspb.String(deptID),
    ValidFrom:    timestamppb.Now(),
}))

// 3. Assign temporary project role
projectRole, err := client.CreateEntityRoleBinding(ctx, connect.NewRequest(&entityv1.CreateEntityRoleBindingRequest{
    TenantId:   tenantID,
    EntityId:   entityID,
    RoleId:     "project-lead-role-uuid",
    ValidFrom:  timestamppb.New(projectStartDate),
    ValidUntil: timestamppb.New(projectEndDate),
}))

// 4. Get all entity roles
bindings, err := client.GetEntityRoleBindings(ctx, connect.NewRequest(&entityv1.GetEntityRoleBindingsRequest{
    EntityId: entityID,
    TenantId: tenantID,
}))

// 5. Suspend entity (e.g., employee on leave)
updated, err := client.UpdateEntity(ctx, connect.NewRequest(&entityv1.UpdateEntityRequest{
    Id:     entityID,
    Status: entityv1.EntityStatus_ENTITY_STATUS_SUSPENDED,
}))

// 6. Reactivate entity
reactivated, err := client.UpdateEntity(ctx, connect.NewRequest(&entityv1.UpdateEntityRequest{
    Id:     entityID,
    Status: entityv1.EntityStatus_ENTITY_STATUS_ACTIVE,
}))
```

### Multi-Type Entity Management

```go
// Create device entity (non-human actor)
deviceEntity, err := client.CreateEntity(ctx, connect.NewRequest(&entityv1.CreateEntityRequest{
    TenantId:        tenantID,
    EntityType:      entityv1.EntityType_ENTITY_TYPE_DEVICE,
    UserId:          deviceServiceAccountUserID,
    ReferenceId:     deviceID,
    ReferenceSource: entityv1.ReferenceSource_REFERENCE_SOURCE_DEVICE,
    Status:          entityv1.EntityStatus_ENTITY_STATUS_ACTIVE,
    Metadata: map[string]string{
        "device_type":   "iot_sensor",
        "serial_number": "SN123456",
        "location":      "warehouse-a",
    },
}))

// Assign device role
deviceRole, err := client.CreateEntityRoleBinding(ctx, connect.NewRequest(&entityv1.CreateEntityRoleBindingRequest{
    TenantId: tenantID,
    EntityId: deviceEntity.Msg.Id,
    RoleId:   "iot-device-role-uuid",
}))
```

## 8. Integration

### With Other Modules

**Identity/User Module:**
- Every entity links to a user account via `user_id`
- User provides authentication, entity provides authorization context
- One user can have multiple entities across tenants

**Identity/Auth Module:**
- Auth session includes entity context
- UserSession contains entity_id, entity_type
- Role resolution uses entity role bindings

**Employee/Contractor/Vendor Modules:**
- Domain modules create entities when creating records
- Entity.reference_id points to domain record UUID
- Entity provides unified identity across domains

**Organization Module:**
- Entities have optional organizational context
- Role bindings can be scoped to division/branch/department
- Permission checks use entity + organization hierarchy

**DMS Module:**
- Documents owned by entities (owner_entity_id)
- Document access based on entity roles
- Sharing targets entities

### Permission Resolution Flow
```
User Login → Auth Session → Entity Lookup
         → Entity Role Bindings → Effective Permissions
         → Organizational Scope → Final Access Decision
```

## 9. Development

### Modifying the Module

**Add New Entity Type:**
1. Update `entity_type` enum in `db/schema/schema.sql`
2. Update `EntityType` enum in `proto/entity.proto`
3. Update `reference_source` if new domain
4. Run migrations and regenerate code

**Add Entity Metadata Fields:**
```go
// Entities use JSONB metadata for extensibility
entity.Metadata["custom_field"] = "value"
```

### Generate Code
```bash
# Generate proto stubs
buf generate

# Generate SQLC queries
sqlc generate -f identity/entity/db/sqlc.yaml

# Run all generation
make generate
```

### Testing Guidelines

**Unit Tests:**
```go
func TestCreateEntity(t *testing.T) {
    mockRepo := &MockEntityRepository{}
    svc := services.NewEntityService(mockRepo, logger)

    entity, err := svc.CreateEntity(ctx, req)
    assert.NoError(t, err)
    assert.Equal(t, entityv1.EntityType_ENTITY_TYPE_EMPLOYEE, entity.EntityType)
}
```

**Integration Tests:**
- Test entity creation with user/domain records
- Test role binding with time bounds
- Test organizational scope filtering
- Test cross-tenant isolation

## 10. Troubleshooting

### Common Issues

**Issue: Unique constraint violation on user_id**
- **Cause:** Trying to create multiple entities for same user in same tenant
- **Solution:** Use GetEntityByUserId first, or handle existing entity

**Issue: Reference not found**
- **Cause:** Domain record doesn't exist yet
- **Solution:** Create domain record before entity, or use transaction

**Issue: Role binding not effective**
- **Cause:** Time bounds (valid_from/valid_until) exclude current time
- **Solution:** Check time bounds, ensure valid_from <= now < valid_until

**Issue: Organizational context missing**
- **Cause:** Entity created without division/branch/department
- **Solution:** Update entity with organizational context

### Performance Tips

1. Cache entity lookups by user_id (changes infrequently)
2. Use composite index on reference_id + reference_source
3. Filter by entity_type for domain-specific queries
4. Eager load role bindings for permission checks

### Debugging

**Check entity existence:**
```sql
SELECT * FROM entities
WHERE user_id = 'user-uuid' AND tenant_id = 'tenant-uuid';
```

**Verify role bindings:**
```sql
SELECT erb.*, r.name as role_name
FROM entity_role_bindings erb
JOIN roles r ON erb.role_id = r.id
WHERE erb.entity_id = 'entity-uuid'
  AND (erb.valid_from IS NULL OR erb.valid_from <= NOW())
  AND (erb.valid_until IS NULL OR erb.valid_until > NOW());
```

**Find entities by type:**
```sql
SELECT e.*, u.username
FROM entities e
JOIN users u ON e.user_id = u.uuid
WHERE e.entity_type = 'EMPLOYEE'
  AND e.status = 'ACTIVE'
  AND e.tenant_id = 'tenant-uuid';
```

### Design Patterns

**Entity Creation Pattern:**
```go
// Always create in this order:
// 1. User (identity/user)
// 2. Domain record (employee, contractor, etc.)
// 3. Entity (identity/entity)

tx.Begin()
user := createUser(...)
employee := createEmployee(user.ID, ...)
entity := createEntity(user.ID, employee.ID, ...)
tx.Commit()
```

**Permission Check Pattern:**
```go
// Resolve effective permissions
entity := getEntityByUserId(userID)
roleBindings := getEntityRoleBindings(entity.ID)

// Filter by organizational scope
effectiveRoles := filterByOrganizationScope(roleBindings, division, branch, dept)

// Filter by time bounds
activeRoles := filterByTimeBounds(effectiveRoles, time.Now())

// Aggregate permissions
permissions := aggregatePermissions(activeRoles)
```

## Additional Resources

- [Entity-Component-System Pattern](https://en.wikipedia.org/wiki/Entity_component_system)
- [Multi-Tenant Architecture](https://docs.microsoft.com/en-us/azure/architecture/guide/multitenant/overview)
- [Role-Based Access Control](https://en.wikipedia.org/wiki/Role-based_access_control)
- [Attribute-Based Access Control](https://en.wikipedia.org/wiki/Attribute-based_access_control)
