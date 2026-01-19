# Organization Module

## 1. Module Overview

The Organization module provides comprehensive organizational hierarchy management for the UGCL platform. It manages divisions, branches, and departments, enabling multi-level organizational structures with support for both business-level and division-level departments.

**Purpose:** Centralized management of organizational units, geographical locations, and hierarchical relationships to support role-based access control, reporting structures, and operational workflows.

**Key Features:**
- Multi-tenant organizational hierarchy management
- Three-tier structure: Divisions > Branches > Departments
- Flexible department hierarchy (business-level and division-level)
- Geospatial support for branch locations
- Contact and address management
- Hierarchical queries with recursive CTE support
- Comprehensive summary views for reporting

## 2. Architecture

### Module Structure
```
organization/
├── proto/                    # Protocol buffer definitions
│   └── organization.proto    # Service and message definitions
├── db/
│   ├── schema/
│   │   └── schema.sql       # Database schema with views
│   └── generated/           # SQLC generated code
├── repository/              # Data access layer
│   ├── division_repository.go
│   ├── branch_repository.go
│   ├── department_repository.go
│   └── organization_repository.go
├── services/                # Business logic layer
│   ├── division_service.go
│   ├── branch_service.go
│   └── department_service.go
├── handlers/                # gRPC handlers
│   └── organization_handler.go
├── mappers/                 # DTO mappers
│   └── organization_mappers.go
└── module.go                # FX module definition
```

### Database Schema

**Tables:**
- `divisions` - Top-level organizational units
- `branches` - Physical locations under divisions
- `departments` - Functional units (business or division-level)

**Views:**
- `division_summary` - Division stats with branch/department counts
- `branch_summary` - Branch details with division info
- `department_hierarchy` - Recursive department tree with paths
- `organization_full_hierarchy` - Complete org structure

**Key Relationships:**
- Divisions → Branches (1:N, CASCADE delete)
- Divisions → Departments (1:N, CASCADE delete, optional)
- Departments → Sub-departments (self-referential)

### Key Dependencies
- `packages/database/sqlc` - Database connection management
- `identity/user` - User references for heads/managers
- Connect-Go RPC framework
- PostgreSQL with JSONB and GIS extensions

## 3. Quick Start

### Using the Service

```go
import (
    "context"
    "p9e.in/ugcl/organization/api/v1/organization/organizationconnect"
    "connectrpc.com/connect"
)

// Create a division
client := organizationconnect.NewOrganizationServiceClient(httpClient, baseURL)

division, err := client.CreateDivision(ctx, connect.NewRequest(&organization.CreateDivisionRequest{
    TenantId:    "tenant-uuid",
    Code:        "DIV-001",
    Name:        "Operations Division",
    Description: "Main operations division",
    HeadUserId:  "user-uuid",
    IsActive:    true,
    DisplayOrder: 1,
}))
```

### Common Operations

**List divisions with summary:**
```go
summaries, err := client.ListDivisionSummaries(ctx, connect.NewRequest(&organization.ListDivisionSummariesRequest{
    TenantId: "tenant-uuid",
    IsActive: &wrapperspb.BoolValue{Value: true},
}))
```

**Create a branch:**
```go
branch, err := client.CreateBranch(ctx, connect.NewRequest(&organization.CreateBranchRequest{
    DivisionId:  "division-uuid",
    TenantId:    "tenant-uuid",
    Code:        "BR-HQ",
    Name:        "Headquarters",
    BranchType:  "HQ",
    Address: &organization.Address{
        Line1:      "123 Main Street",
        City:       "Mumbai",
        State:      "Maharashtra",
        Country:    "India",
        PostalCode: "400001",
    },
    Location: &organization.Location{
        Latitude:  19.0760,
        Longitude: 72.8777,
    },
}))
```

## 4. API Reference

### Division Operations
- `CreateDivision` - Create a new division
- `UpdateDivision` - Update division details
- `GetDivision` - Get division by ID
- `GetDivisionByCode` - Get division by code
- `ListDivisions` - List divisions with pagination
- `ListDivisionSummaries` - Get division summaries with counts
- `DeleteDivision` - Soft delete a division

### Branch Operations
- `CreateBranch` - Create a new branch
- `UpdateBranch` - Update branch details
- `GetBranch` - Get branch by ID
- `GetBranchByCode` - Get branch by code
- `ListBranches` - List branches with pagination
- `ListBranchesByDivision` - Get branches for a division
- `ListBranchesByCity` - Get branches in a city
- `ListBranchesByState` - Get branches in a state
- `ListBranchSummaries` - Get branch summaries
- `DeleteBranch` - Soft delete a branch

### Department Operations
- `CreateDepartment` - Create a new department
- `UpdateDepartment` - Update department details
- `GetDepartment` - Get department by ID
- `GetDepartmentByCode` - Get department by code
- `ListDepartments` - List departments with pagination
- `ListDepartmentsByDivision` - Get division-level departments
- `ListBusinessLevelDepartments` - Get business-level departments
- `ListSubDepartments` - Get sub-departments
- `GetDepartmentHierarchy` - Get complete department tree
- `DeleteDepartment` - Soft delete a department

## 5. Database Schema

### Tables Overview

**divisions**
```sql
- id (UUID, PK)
- tenant_id (UUID)
- code (VARCHAR(50), UNIQUE per tenant)
- name (VARCHAR(255))
- description (TEXT)
- head_user_id (VARCHAR(255))
- is_active (BOOLEAN)
- display_order (INTEGER)
- metadata (JSONB)
- created_at, updated_at, created_by, updated_by
```

**branches**
```sql
- id (UUID, PK)
- division_id (UUID, FK → divisions)
- tenant_id (UUID)
- code (VARCHAR(50), UNIQUE per tenant)
- name (VARCHAR(255))
- branch_type (VARCHAR(50)) -- 'HQ', 'Regional', 'Local', 'Satellite'
- address fields (line1, line2, city, state, country, postal_code)
- location (latitude NUMERIC, longitude NUMERIC)
- contact (phone, email)
- branch_manager_user_id (VARCHAR(255))
- is_active (BOOLEAN)
- display_order (INTEGER)
- metadata (JSONB)
```

**departments**
```sql
- id (UUID, PK)
- tenant_id (UUID)
- division_id (UUID, FK → divisions, NULLABLE for business-level)
- code (VARCHAR(50), UNIQUE per tenant+division)
- name (VARCHAR(255))
- description (TEXT)
- department_type (VARCHAR(50)) -- 'Operational', 'Support', 'Administrative'
- head_user_id (VARCHAR(255))
- parent_department_id (UUID, self-reference for sub-departments)
- is_active (BOOLEAN)
- display_order (INTEGER)
- metadata (JSONB)
```

### Key Relationships
- Divisions CASCADE delete to branches and departments
- Departments support self-referential hierarchy
- Business-level departments have NULL division_id

### Indexes
- Tenant-based indexes for multi-tenancy
- Code-based unique indexes
- Active status partial indexes
- City/state indexes for branch queries
- GIN indexes on JSONB metadata

## 6. Configuration

### Environment Variables
```env
# Database connection managed by packages/database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=ugcl
DB_USER=postgres
DB_PASSWORD=password
```

### Module-Specific Settings
- Managed through SQLC configuration in `db/sqlc.yaml`
- Proto generation via `buf.gen.yaml` in project root

## 7. Examples

### Complete Division Setup

```go
// 1. Create Division
division, err := client.CreateDivision(ctx, connect.NewRequest(&organization.CreateDivisionRequest{
    TenantId:    tenantID,
    Code:        "OPS",
    Name:        "Operations",
    Description: "Operations Division",
    IsActive:    true,
}))

// 2. Create Headquarters Branch
hqBranch, err := client.CreateBranch(ctx, connect.NewRequest(&organization.CreateBranchRequest{
    DivisionId: division.Msg.Id,
    TenantId:   tenantID,
    Code:       "OPS-HQ",
    Name:       "Operations HQ",
    BranchType: "HQ",
    Address: &organization.Address{
        Line1:      "Corporate Plaza",
        City:       "Mumbai",
        State:      "Maharashtra",
        Country:    "India",
        PostalCode: "400001",
    },
}))

// 3. Create Division-Level Department
dept, err := client.CreateDepartment(ctx, connect.NewRequest(&organization.CreateDepartmentRequest{
    TenantId:       tenantID,
    DivisionId:     division.Msg.Id,
    Code:           "OPS-PROD",
    Name:           "Production",
    DepartmentType: "Operational",
    IsActive:       true,
}))

// 4. Create Sub-Department
subDept, err := client.CreateDepartment(ctx, connect.NewRequest(&organization.CreateDepartmentRequest{
    TenantId:            tenantID,
    DivisionId:          division.Msg.Id,
    Code:                "OPS-PROD-QC",
    Name:                "Quality Control",
    ParentDepartmentId:  dept.Msg.Id,
    DepartmentType:      "Operational",
    IsActive:            true,
}))
```

### Query Department Hierarchy

```go
hierarchy, err := client.GetDepartmentHierarchy(ctx, connect.NewRequest(&organization.GetDepartmentHierarchyRequest{
    TenantId: tenantID,
    IsActive: &wrapperspb.BoolValue{Value: true},
}))

// Returns departments with level, path array, and full_path string
for _, dept := range hierarchy.Msg.Departments {
    fmt.Printf("Level %d: %s (%s)\n", dept.Level, dept.Name, dept.FullPath)
}
```

## 8. Integration

### With Other Modules

**Employee Module:**
- Employees reference division_id, branch_id, department_id
- Manager relationships use employee UUIDs

**Identity/User Module:**
- Division heads reference user IDs
- Branch managers reference user IDs
- Department heads reference user IDs

**Permissions System:**
- Organizational units used for scope-based permissions
- Entity role bindings can be scoped to division/branch/department

**DMS Module:**
- Documents tagged with organizational context
- Access control based on organizational hierarchy

### Dependencies Graph
```
organization
  ├── depends on: identity/user (for user references)
  └── depended by: employee, contractors, vendors, dms, entity
```

## 9. Development

### Modifying the Module

**Add New Fields:**
1. Update `proto/organization.proto`
2. Update `db/schema/schema.sql`
3. Run migrations: `make migrate-up`
4. Regenerate code: `make generate`
5. Update mappers in `mappers/`
6. Update service logic in `services/`

**Generate Code:**
```bash
# Generate proto stubs
buf generate

# Generate SQLC queries
sqlc generate -f organization/db/sqlc.yaml

# Run all code generation
make generate
```

### Testing Guidelines

**Unit Tests:**
- Test services with mocked repositories
- Test mappers for correct proto/model conversion
- Test validation logic

**Integration Tests:**
- Test against real database with test containers
- Test cascade deletes
- Test hierarchy queries

**Example Test:**
```go
func TestCreateDivision(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer db.Close()

    // Create service with test dependencies
    svc := services.NewDivisionService(repo, logger)

    // Test division creation
    division, err := svc.CreateDivision(ctx, req)
    assert.NoError(t, err)
    assert.Equal(t, "DIV-001", division.Code)
}
```

### Code Generation

**SQLC Configuration** (`db/sqlc.yaml`):
```yaml
version: "2"
sql:
  - schema: "db/schema/schema.sql"
    queries: "db/queries/"
    engine: "postgresql"
    gen:
      go:
        package: "generated"
        out: "db/generated"
```

## 10. Troubleshooting

### Common Issues

**Issue: Unique constraint violation on code**
- **Cause:** Code already exists for tenant
- **Solution:** Use unique codes per tenant or check existing codes first

**Issue: CASCADE delete fails**
- **Cause:** Referenced by employees or other modules
- **Solution:** Check foreign key references before deletion

**Issue: Department hierarchy depth limit**
- **Cause:** PostgreSQL recursive CTE may have default depth limit
- **Solution:** Adjust max_recursive_depth if needed

**Issue: Geospatial queries not working**
- **Cause:** Missing PostGIS extension
- **Solution:** Ensure PostGIS is installed and enabled

### Performance Tips

1. Use summary views for reporting instead of joins
2. Leverage partial indexes for active-only queries
3. Use display_order for UI sorting instead of name
4. Cache division/branch lists as they change infrequently

### Debugging

**Enable SQL logging:**
```go
// In database connection setup
db.LogMode(true)
```

**Check hierarchy structure:**
```sql
SELECT * FROM department_hierarchy
WHERE tenant_id = 'your-tenant-id'
ORDER BY full_path;
```

**Verify CASCADE behavior:**
```sql
-- Check what will be deleted
SELECT 'branches' as type, id, name FROM branches WHERE division_id = 'div-id'
UNION ALL
SELECT 'departments', id, name FROM departments WHERE division_id = 'div-id';
```

## Additional Resources

- [Protocol Buffers Documentation](https://protobuf.dev/)
- [SQLC Documentation](https://docs.sqlc.dev/)
- [Connect-Go Documentation](https://connectrpc.com/docs/go/getting-started)
- [PostgreSQL Recursive Queries](https://www.postgresql.org/docs/current/queries-with.html)
