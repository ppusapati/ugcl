# Tenant Management Module

## 1. Module Overview

The Tenant Management module provides multi-tenancy capabilities for the UGCL platform, enabling complete data isolation and customization for different organizations. It manages tenant lifecycle, database configurations, feature flags, and tenant-specific settings.

**Purpose:** Enable Software-as-a-Service (SaaS) multi-tenancy with flexible database isolation strategies, tenant customization, and centralized tenant administration.

**Key Features:**
- Tenant creation and management
- Multiple database isolation strategies (shared DB, separate DB, separate schema)
- Tenant-specific configuration and connection strings
- Feature flag management per tenant
- Tenant logo and branding
- Regional tenant support
- Public tenant information access (for login pages)
- Current tenant context resolution
- Admin user provisioning per tenant

## 2. Architecture

### Module Structure
```
identity/tenant/
├── proto/
│   └── tenant.proto         # Service and message definitions
├── db/
│   ├── schema/
│   │   └── schema.sql      # Tenant tables
│   └── generated/          # SQLC generated code
├── repository/             # Data access layer
│   └── tenant_repository.go
├── services/               # Business logic layer
│   └── tenant_service.go
├── handler/                # gRPC handlers
│   └── tenant_handler.go
├── mappers/                # DTO mappers
│   └── tenant_mappers.go
├── models/                 # Domain models
│   └── tenant.go
├── uow/                    # Unit of work
└── module.go               # FX module definition
```

### Database Schema

**Tables:**
- `tenants` - Tenant master data
- `tenant_connections` - Connection string configurations
- `tenant_features` - Feature flags per tenant

**Key Relationships:**
- Tenant → Connections (1:N key-value pairs)
- Tenant → Features (1:N key-value pairs)
- User → Tenants (N:N via user.tenant_ids)

### Key Dependencies
- `identity/user` - Admin user creation
- Database manager for multi-database support
- All application modules (tenant context)
- Connect-Go RPC framework

## 3. Quick Start

### Create a Tenant

```go
import (
    tenantv1 "p9e.in/ugcl/identity/tenant/api/v1/tenant"
    "connectrpc.com/connect"
)

client := tenantv1.NewTenantServiceClient(httpClient, baseURL)

// Create new tenant with admin user
tenant, err := client.CreateTenant(ctx, connect.NewRequest(&tenantv1.CreateTenantRequest{
    Name:        "acme-corp",
    DisplayName: "ACME Corporation",
    Region:      "us-east-1",
    Logo:        "https://cdn.example.com/acme-logo.png",
    SeparateDb:  true, // Use separate database
    AdminEmail:  wrapperspb.String("admin@acme.com"),
    AdminUsername: wrapperspb.String("admin"),
    AdminPassword: wrapperspb.String("SecureAdminPass123!"),
}))

tenantID := tenant.Msg.Id
```

### Common Operations

**Get tenant:**
```go
tenant, err := client.GetTenant(ctx, connect.NewRequest(&tenantv1.GetTenantRequest{
    IdOrName: "acme-corp", // Can use ID or name
}))
```

**Get public tenant info (for login page):**
```go
publicInfo, err := client.GetTenantPublic(ctx, connect.NewRequest(&tenantv1.GetTenantPublicRequest{
    IdOrName: "acme-corp",
}))

fmt.Printf("Tenant: %s\n", publicInfo.Msg.DisplayName)
fmt.Printf("Logo: %s\n", publicInfo.Msg.Logo)
fmt.Printf("Region: %s\n", publicInfo.Msg.Region)
```

**Update tenant:**
```go
updated, err := client.UpdateTenant(ctx, connect.NewRequest(&tenantv1.UpdateTenantRequest{
    Tenant: &tenantv1.UpdateTenant{
        Id:          tenantID,
        DisplayName: "ACME Corporation Ltd",
        Logo:        "https://cdn.example.com/acme-new-logo.png",
        Features: []*tenantv1.TenantFeature{
            {Key: "advanced_analytics", Value: "enabled"},
            {Key: "api_rate_limit", Value: "10000"},
        },
    },
    UpdateMask: &fieldmaskpb.FieldMask{
        Paths: []string{"display_name", "logo", "features"},
    },
}))
```

**Get current tenant context:**
```go
current, err := client.GetCurrentTenant(ctx, connect.NewRequest(&tenantv1.GetCurrentTenantRequest{}))

tenantInfo := current.Msg.Tenant
isHost := current.Msg.IsHost // True if system admin/host tenant
```

## 4. API Reference

### TenantService (Main Service)
- `CreateTenant` - Create new tenant with optional admin user
- `UpdateTenant` - Update tenant configuration
- `DeleteTenant` - Delete tenant (careful!)
- `GetTenant` - Get tenant by ID or name
- `GetTenantPublic` - Get public tenant info (no auth required)
- `ListTenant` - List tenants with filtering and search
- `GetCurrentTenant` - Get current tenant context from session

### TenantInternalService (Internal APIs)
- `GetTenant` - Internal tenant lookup for remote tenant store

## 5. Database Schema

### Table: tenants

```sql
- id (UUID, PK)
- name (VARCHAR, UNIQUE) -- URL-safe identifier (e.g., acme-corp)
- display_name (VARCHAR) -- Human-readable name
- region (VARCHAR) -- Geographic region
- logo (TEXT) -- Logo URL or base64
- tenant_db (ENUM) -- SHARED, SEPERATEDB, SEPERATESCHEMA
- created_at, updated_at (TIMESTAMP)
```

### Table: tenant_connections

```sql
- id (SERIAL, PK)
- tenant_id (UUID, FK → tenants)
- key (VARCHAR) -- Connection identifier (db_host, db_name, etc.)
- value (TEXT) -- Connection parameter value
```

### Table: tenant_features

```sql
- id (SERIAL, PK)
- tenant_id (UUID, FK → tenants)
- key (VARCHAR) -- Feature identifier
- value (TEXT) -- Feature configuration value
```

### Enums

**TenantDB:**
- `SHARED` - All tenants share same database (tenant_id filtering)
- `SEPERATEDB` - Each tenant has dedicated database
- `SEPERATESCHEMA` - Each tenant has separate schema in same database

### Indexes
- `idx_tenants_name` - Fast name lookup (unique)
- `idx_tenants_region` - Region-based queries
- `idx_tenant_connections_tenant_id` - Connection lookup
- `idx_tenant_features_tenant_id` - Feature lookup

## 6. Configuration

### Environment Variables
```env
# Host database (for tenant registry)
DB_HOST=localhost
DB_PORT=5432
DB_NAME=ugcl_host
DB_USER=postgres
DB_PASSWORD=password

# Default tenant database template
TENANT_DB_TEMPLATE=ugcl_template
TENANT_DB_USER=tenant_user
TENANT_DB_PASSWORD=tenant_pass

# Tenant creation
DEFAULT_TENANT_REGION=us-east-1
ENABLE_AUTO_PROVISIONING=false
```

### Module-Specific Settings
- Database isolation strategy
- Tenant name validation rules
- Logo size limits
- Feature flags catalog

## 7. Examples

### Complete Tenant Setup

```go
// 1. Create tenant with separate database
tenant, err := client.CreateTenant(ctx, connect.NewRequest(&tenantv1.CreateTenantRequest{
    Name:         "tech-startup",
    DisplayName:  "Tech Startup Inc",
    Region:       "us-west-2",
    Logo:         "https://cdn.example.com/tech-logo.png",
    SeparateDb:   true,
    AdminEmail:   wrapperspb.String("admin@techstartup.com"),
    AdminUsername: wrapperspb.String("admin"),
    AdminPassword: wrapperspb.String("AdminPass123!"),
}))

tenantID := tenant.Msg.Id

// 2. Configure tenant-specific database connections
updated, err := client.UpdateTenant(ctx, connect.NewRequest(&tenantv1.UpdateTenantRequest{
    Tenant: &tenantv1.UpdateTenant{
        Id: tenantID,
        Conn: []*tenantv1.TenantConnectionString{
            {Key: "db_host", Value: "tenant-db.example.com"},
            {Key: "db_port", Value: "5432"},
            {Key: "db_name", Value: "tech_startup_db"},
            {Key: "db_user", Value: "tech_user"},
            {Key: "db_password", Value: "encrypted_password"},
            {Key: "db_ssl_mode", Value: "require"},
        },
    },
    UpdateMask: &fieldmaskpb.FieldMask{
        Paths: []string{"conn"},
    },
}))

// 3. Enable tenant-specific features
featuresUpdate, err := client.UpdateTenant(ctx, connect.NewRequest(&tenantv1.UpdateTenantRequest{
    Tenant: &tenantv1.UpdateTenant{
        Id: tenantID,
        Features: []*tenantv1.TenantFeature{
            {Key: "advanced_reporting", Value: "enabled"},
            {Key: "api_access", Value: "enabled"},
            {Key: "max_users", Value: "100"},
            {Key: "storage_quota_gb", Value: "500"},
            {Key: "custom_branding", Value: "enabled"},
        },
    },
    UpdateMask: &fieldmaskpb.FieldMask{
        Paths: []string{"features"},
    },
}))
```

### Multi-Region Tenant Management

```go
// List tenants by region
tenants, err := client.ListTenant(ctx, connect.NewRequest(&tenantv1.ListTenantRequest{
    PageSize: 50,
    Filter: &tenantv1.TenantFilter{
        Region: &queryv1.StringFilterOperation{
            Eq: wrapperspb.String("us-east-1"),
        },
    },
    Sort: []string{"display_name"},
}))

fmt.Printf("Tenants in us-east-1: %d\n", tenants.Msg.FilterSize)
for _, t := range tenants.Msg.Items {
    fmt.Printf("- %s (%s)\n", t.DisplayName, t.Name)
}
```

### Tenant Context Resolution

```go
// In middleware or request handler
current, err := client.GetCurrentTenant(ctx, connect.NewRequest(&tenantv1.GetCurrentTenantRequest{}))

if current.Msg.IsHost {
    // System admin access - can see all tenants
    fmt.Println("Host/System Admin Access")
} else {
    // Regular tenant user
    tenantID := current.Msg.Tenant.Id
    tenantName := current.Msg.Tenant.Name
    fmt.Printf("Tenant Context: %s (%s)\n", tenantName, tenantID)

    // Use tenant context for data isolation
    // All queries should filter by tenantID
}
```

### Feature Flag Checks

```go
// Get tenant and check features
tenant, err := client.GetTenant(ctx, connect.NewRequest(&tenantv1.GetTenantRequest{
    IdOrName: tenantName,
}))

features := make(map[string]string)
for _, f := range tenant.Msg.Features {
    features[f.Key] = f.Value
}

// Check if feature is enabled
if features["advanced_reporting"] == "enabled" {
    // Show advanced reporting UI
}

// Check quota limits
maxUsers, _ := strconv.Atoi(features["max_users"])
if currentUserCount >= maxUsers {
    return errors.New("user limit reached")
}
```

## 8. Integration

### With Other Modules

**Identity/User Module:**
- Users belong to tenants via tenant_ids
- Tenant admin user created during tenant setup
- User access scoped by tenant

**Identity/Auth Module:**
- Authentication includes tenant context
- Sessions contain tenant_id
- Tenant-specific login pages

**All Application Modules:**
- All data filtered by tenant_id
- Tenant context required for operations
- Multi-tenant data isolation

**Database Manager:**
- Resolves tenant database connections
- Handles database routing
- Manages connection pools per tenant

### Tenant Context Flow
```
Request → Tenant Identifier (subdomain/header/token)
       → GetCurrentTenant → Tenant Context
       → Database Connection Resolver
       → Tenant-Specific Data Access
```

## 9. Development

### Modifying the Module

**Add New Tenant Field:**
1. Update tenant proto message
2. Update tenants table schema
3. Run migration
4. Regenerate SQLC and proto
5. Update mappers

**Add New Feature Flag:**
```go
// Features are key-value pairs
// Add to tenant via UpdateTenant
features := []*tenantv1.TenantFeature{
    {Key: "new_feature", Value: "enabled"},
}
```

### Generate Code
```bash
# Generate proto stubs
buf generate

# Generate SQLC queries
sqlc generate -f identity/tenant/db/sqlc.yaml

# Run migrations
make migrate-up
```

### Testing Guidelines

**Unit Tests:**
```go
func TestCreateTenant(t *testing.T) {
    mockRepo := &MockTenantRepository{}
    mockUserService := &MockUserService{}
    svc := services.NewTenantService(mockRepo, mockUserService)

    tenant, err := svc.CreateTenant(ctx, req)
    assert.NoError(t, err)
    assert.Equal(t, "acme-corp", tenant.Name)
}
```

**Integration Tests:**
- Test tenant creation with database provisioning
- Test tenant isolation
- Test connection string resolution
- Test feature flag management

## 10. Troubleshooting

### Common Issues

**Issue: Duplicate tenant name**
- **Cause:** Tenant name already exists
- **Solution:** Use unique tenant names (URL-safe)

**Issue: Database connection fails**
- **Cause:** Invalid connection string configuration
- **Solution:** Verify tenant_connections settings

**Issue: Data leakage between tenants**
- **Cause:** Missing tenant_id filter in queries
- **Solution:** Always filter by tenant context

**Issue: Tenant not found in session**
- **Cause:** User not assigned to tenant
- **Solution:** Add user to tenant via tenant_ids

**Issue: Cannot delete tenant**
- **Cause:** Tenant has active users/data
- **Solution:** Archive tenant instead, or migrate data first

### Performance Tips

1. Cache tenant configurations
2. Use connection pooling per tenant
3. Index tenant_id in all tables
4. Implement tenant-level caching
5. Monitor database connections per tenant

### Debugging

**Check tenant existence:**
```sql
SELECT id, name, display_name, tenant_db, region
FROM tenants
WHERE name = 'acme-corp';
```

**Verify tenant connections:**
```sql
SELECT key, value
FROM tenant_connections
WHERE tenant_id = 'tenant-uuid'
ORDER BY key;
```

**Check tenant features:**
```sql
SELECT key, value
FROM tenant_features
WHERE tenant_id = 'tenant-uuid'
ORDER BY key;
```

**Find users by tenant:**
```sql
SELECT username, email, tenant_ids
FROM users
WHERE 'tenant-uuid' = ANY(tenant_ids);
```

### Database Isolation Strategies

**SHARED (Shared Database):**
- Pros: Simple, cost-effective, easy maintenance
- Cons: Potential performance impact, data leakage risk
- Use case: Small-medium deployments, similar tenants

**SEPERATEDB (Separate Database):**
- Pros: Complete isolation, independent scaling, easier compliance
- Cons: Higher cost, more complex maintenance
- Use case: Enterprise clients, compliance requirements

**SEPERATESCHEMA (Separate Schema):**
- Pros: Good isolation, shared resources, moderate cost
- Cons: Schema management complexity
- Use case: Balance between isolation and cost

## Additional Resources

- [Multi-Tenancy Architecture Patterns](https://docs.microsoft.com/en-us/azure/architecture/guide/multitenant/overview)
- [Database Per Tenant](https://martinfowler.com/bliki/DatabasePerTenant.html)
- [SaaS Tenant Isolation](https://docs.aws.amazon.com/prescriptive-guidance/latest/saas-multitenant-managed-services/tenant-isolation.html)
- [Feature Flags Best Practices](https://www.martinfowler.com/articles/feature-toggles.html)
