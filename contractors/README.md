# Contractors Management Module

## 1. Module Overview

The Contractors module manages third-party contractor companies and their personnel working with UGCL. It handles contractor company profiles, contract lifecycle, project associations, site assignments, and compliance tracking.

**Purpose:** Centralized management of contractor companies, their contracts, work assignments, and statutory compliance for effective vendor and workforce management.

**Key Features:**
- Contractor company profile management
- Contract lifecycle tracking (start date, end date, renewals)
- Project and site associations
- Statutory compliance (GST, PAN)
- Contact person management via user accounts
- Company categorization
- Contract status management
- Multi-project and multi-site support

## 2. Architecture

### Module Structure
```
contractors/
├── proto/
│   └── contractor.proto      # Service and message definitions
├── db/
│   ├── schema/
│   │   └── contractors.sql  # Database schema
│   └── generated/           # SQLC generated code (planned)
├── repository/              # Data access layer
│   └── contractor_repository.go
├── services/                # Business logic layer
│   └── contractor_service.go
├── handlers/                # gRPC handlers
│   └── contractor_handler.go
├── mappers/                 # DTO mappers
│   └── contractor_mappers.go
├── models/                  # Domain models
│   └── contractor.go
└── module.go                # FX module definition
```

### Database Schema

**Tables:**
- `contractors` - Contractor company information and contracts

**Key Relationships:**
- Contractor → User (N:1 via person_id for contact person)
- Contractor → Projects (N:N via associated_project array)
- Contractor → Sites (N:N via working_site array)

### Key Dependencies
- `identity/user` - User account for contact persons
- `identity/entity` - Entity abstraction
- `projects` - Project references (future)
- Connect-Go RPC framework
- PostgreSQL with array and JSONB support

## 3. Quick Start

### Create a Contractor

```go
import (
    contractorv2 "p9e.in/ugcl/contractors/api/v2/contractor"
    userv2 "p9e.in/ugcl/identity/user/api/v2/user"
    "connectrpc.com/connect"
)

client := contractorv2.NewContractorServiceClient(httpClient, baseURL)

// Create contractor with contact person user account
resp, err := client.CreateContractor(ctx, connect.NewRequest(&contractorv2.CreateContractorRequest{
    Contractor: &contractorv2.Contractor{
        CompanyName:         "ABC Construction Pvt Ltd",
        CompanyType:         "Private Limited",
        Gst:                 "27AABCU9603R1ZM",
        Pan:                 "AABCU9603R",
        Category:            "Construction",
        AssociatedProject:   []string{"project-uuid-1", "project-uuid-2"},
        WorkingSite:         []string{"site-uuid-1", "site-uuid-2"},
        ContractStartDate:   timestamppb.New(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
        ContractEndDate:     timestamppb.New(time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)),
        Status:              "active",
        Metadata: &structpb.Struct{
            Fields: map[string]*structpb.Value{
                "license_number": structpb.NewStringValue("LIC12345"),
                "insurance":      structpb.NewStringValue("Active"),
            },
        },
    },
    User: &userv2.User{
        Username:       wrapperspb.String("contact.abc"),
        Email:          wrapperspb.String("contact@abcconstruction.com"),
        Fullname:       wrapperspb.String("Ramesh Kumar"),
        Phone:          wrapperspb.String("+91-9876543210"),
        Password:       "ContactPassword123!",
        ConfirmPassword: "ContactPassword123!",
    },
}))
```

### Common Operations

**Get contractor:**
```go
contractor, err := client.GetContractor(ctx, connect.NewRequest(&contractorv2.ContractorIdentifier{
    Id: "contractor-uuid",
}))
```

**List contractors with filters:**
```go
contractors, err := client.ListContractors(ctx, connect.NewRequest(&contractorv2.ListContractorsRequest{
    PageSize:   50,
    PageOffset: 0,
    Filter: &contractorv2.ContractorFilter{
        Status:      "active",
        Category:    "Construction",
        CompanyName: "ABC",
    },
    Sort: []string{"company_name", "contract_start_date"},
}))
```

**Update contractor:**
```go
updated, err := client.UpdateContractor(ctx, connect.NewRequest(&contractorv2.UpdateContractorRequest{
    Contractor: &contractorv2.Contractor{
        Id:              contractorID,
        Status:          "inactive",
        ContractEndDate: timestamppb.New(time.Now()),
    },
    UpdateMask: &fieldmaskpb.FieldMask{
        Paths: []string{"status", "contract_end_date"},
    },
}))
```

## 4. API Reference

### Contractor Operations
- `CreateContractor` - Create new contractor with contact person
- `UpdateContractor` - Update contractor details with field mask
- `GetContractor` - Get contractor by ID
- `ListContractors` - List contractors with filtering and pagination
- `DeleteContractor` - Soft delete contractor record
- `GetContractorByUserId` - Find contractor by contact person user ID

## 5. Database Schema

### Table: contractors

```sql
-- Identity fields
- id (BIGSERIAL, PK)
- uuid (UUID, UNIQUE, generated)
- company_name (VARCHAR(255))
- company_type (VARCHAR(100)) -- Private Limited, Partnership, Proprietorship, etc.

-- Statutory information
- gst (VARCHAR(50)) -- GST number
- pan (VARCHAR(50), UNIQUE) -- PAN of company
- category (VARCHAR(100)) -- Construction, Electrical, Plumbing, etc.

-- Contact information
- person_id (BIGINT, FK → users.id) -- Contact person user account

-- Work assignments
- associated_project (TEXT[]) -- Array of project UUIDs
- working_site (TEXT[]) -- Array of site UUIDs

-- Contract details
- contract_start_date (TIMESTAMP)
- contract_end_date (TIMESTAMP)

-- Status and metadata
- status (VARCHAR(50)) -- active, inactive, expired, terminated
- metadata (JSONB) -- Additional information (license, insurance, etc.)

-- Audit fields
- created_at, updated_at (TIMESTAMP)
- deleted_at (TIMESTAMP, soft delete)
```

### Constraints
- `chk_contract_dates` - contract_end_date > contract_start_date (or NULL)
- PAN must be unique across all contractors

### Indexes
- `idx_contractors_person_id` - Contact person lookup
- `idx_contractors_status` - Filter by status
- `idx_contractors_company` - Company name search
- `idx_contractors_category` - Category filtering
- `idx_contractors_pan` - PAN lookup
- `idx_contractors_gst` - GST lookup
- `idx_contractors_projects` - GIN index for project associations
- `idx_contractors_sites` - GIN index for site assignments

## 6. Configuration

### Environment Variables
```env
# Database connection managed by packages/database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=ugcl
```

### Module-Specific Settings
- Contract expiry notification period (default: 30 days)
- Auto-expire contracts after end date
- Mandatory compliance documents

## 7. Examples

### Complete Contractor Lifecycle

```go
// 1. Create new contractor
contractor, err := client.CreateContractor(ctx, connect.NewRequest(&contractorv2.CreateContractorRequest{
    Contractor: &contractorv2.Contractor{
        CompanyName:       "XYZ Electricals",
        CompanyType:       "Partnership",
        Gst:               "27XYZAB1234C1Z5",
        Pan:               "XYZAB1234C",
        Category:          "Electrical",
        ContractStartDate: timestamppb.New(time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)),
        ContractEndDate:   timestamppb.New(time.Date(2025, 5, 31, 0, 0, 0, 0, time.UTC)),
        Status:            "active",
    },
    User: &userv2.User{
        Username:       wrapperspb.String("contact.xyz"),
        Email:          wrapperspb.String("contact@xyzelectricals.com"),
        Fullname:       wrapperspb.String("Suresh Patel"),
        Phone:          wrapperspb.String("+91-9123456789"),
        Password:       "Password123!",
        ConfirmPassword: "Password123!",
    },
}))

contractorID := contractor.Msg.Id

// 2. Assign to projects and sites
assigned, err := client.UpdateContractor(ctx, connect.NewRequest(&contractorv2.UpdateContractorRequest{
    Contractor: &contractorv2.Contractor{
        Id:                contractorID,
        AssociatedProject: []string{"electrical-project-uuid"},
        WorkingSite:       []string{"site-a-uuid", "site-b-uuid"},
    },
    UpdateMask: &fieldmaskpb.FieldMask{
        Paths: []string{"associated_project", "working_site"},
    },
}))

// 3. Renew contract
renewed, err := client.UpdateContractor(ctx, connect.NewRequest(&contractorv2.UpdateContractorRequest{
    Contractor: &contractorv2.Contractor{
        Id:              contractorID,
        ContractEndDate: timestamppb.New(time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC)),
    },
    UpdateMask: &fieldmaskpb.FieldMask{
        Paths: []string{"contract_end_date"},
    },
}))

// 4. Terminate contract
terminated, err := client.UpdateContractor(ctx, connect.NewRequest(&contractorv2.UpdateContractorRequest{
    Contractor: &contractorv2.Contractor{
        Id:              contractorID,
        Status:          "terminated",
        ContractEndDate: timestamppb.Now(),
    },
    UpdateMask: &fieldmaskpb.FieldMask{
        Paths: []string{"status", "contract_end_date"},
    },
}))
```

### Query Contractors by Project

```go
// Get all contractors working on specific project
contractors, err := client.ListContractors(ctx, connect.NewRequest(&contractorv2.ListContractorsRequest{
    Filter: &contractorv2.ContractorFilter{
        Status: "active",
    },
}))

// Filter contractors with specific project (application-level filtering)
projectID := "target-project-uuid"
var projectContractors []*contractorv2.Contractor

for _, c := range contractors.Msg.Contractors {
    for _, p := range c.AssociatedProject {
        if p == projectID {
            projectContractors = append(projectContractors, c)
            break
        }
    }
}
```

### Expiring Contracts Report

```go
// Find contracts expiring in next 30 days
thirtyDaysFromNow := time.Now().AddDate(0, 0, 30)

contractors, err := client.ListContractors(ctx, connect.NewRequest(&contractorv2.ListContractorsRequest{
    Filter: &contractorv2.ContractorFilter{
        Status: "active",
    },
}))

var expiring []*contractorv2.Contractor
for _, c := range contractors.Msg.Contractors {
    if c.ContractEndDate != nil {
        endDate := c.ContractEndDate.AsTime()
        if endDate.Before(thirtyDaysFromNow) && endDate.After(time.Now()) {
            expiring = append(expiring, c)
        }
    }
}

fmt.Printf("Contracts expiring in 30 days: %d\n", len(expiring))
```

## 8. Integration

### With Other Modules

**Identity/User Module:**
- Contact person has user account for portal access
- User.id → Contractor.person_id relationship
- Contact person can view/update contractor details

**Identity/Entity Module:**
- Contractor record creates corresponding entity
- Entity.reference_id → Contractor.uuid
- Entity type: CONTRACTOR

**Projects Module (Future):**
- Contractors assigned to projects
- associated_project references project UUIDs
- Project-contractor billing and tracking

**Sites/Locations Module (Future):**
- Contractors work at specific sites
- working_site references site UUIDs
- Site access control based on assignments

**DMS Module:**
- Contractor documents (licenses, insurance, contracts)
- Document ownership via entity
- Compliance document management

### Event Publishing
```go
// Contractor lifecycle events
type ContractorEvent struct {
    Type         string    // created, assigned, renewed, expired, terminated
    ContractorID string
    CompanyName  string
    Timestamp    time.Time
    Changes      map[string]interface{}
}
```

## 9. Development

### Modifying the Module

**Add New Field:**
1. Update `contractors.sql` schema
2. Update `contractor.proto` message
3. Run migration: `make migrate-up`
4. Regenerate SQLC (when configured)
5. Regenerate proto: `buf generate`
6. Update mappers and services

**Add Contractor Category:**
```sql
-- Categories stored as VARCHAR, no enum constraint
-- Add validation in application layer or use CHECK constraint
ALTER TABLE contractors ADD CONSTRAINT chk_category
    CHECK (category IN ('Construction', 'Electrical', 'Plumbing', 'HVAC', 'Security', 'Cleaning', 'Other'));
```

### Generate Code
```bash
# Generate proto stubs
buf generate

# Generate SQLC queries (when configured)
sqlc generate -f contractors/db/sqlc.yaml

# Run migrations
make migrate-up
```

### Testing Guidelines

**Unit Tests:**
```go
func TestCreateContractor(t *testing.T) {
    mockRepo := &MockContractorRepository{}
    mockUserService := &MockUserService{}
    svc := services.NewContractorService(mockRepo, mockUserService)

    contractor, err := svc.CreateContractor(ctx, req)
    assert.NoError(t, err)
    assert.Equal(t, "ABC Construction", contractor.CompanyName)
    assert.Equal(t, "active", contractor.Status)
}
```

**Integration Tests:**
- Test contractor creation with user
- Test project/site assignments
- Test contract expiry logic
- Test soft delete behavior

## 10. Troubleshooting

### Common Issues

**Issue: Duplicate PAN**
- **Cause:** PAN already registered with another contractor
- **Solution:** Check existing contractors before creation

**Issue: Invalid GST format**
- **Cause:** GST number doesn't match pattern
- **Solution:** Validate GST format (15 characters, alphanumeric)

**Issue: Contract dates invalid**
- **Cause:** End date before start date
- **Solution:** Validate dates before save, constraint will catch it

**Issue: User creation fails for contact person**
- **Cause:** Email/username already exists
- **Solution:** Validate user uniqueness before contractor creation

**Issue: Cannot delete contractor**
- **Cause:** Contractor has active projects or pending invoices
- **Solution:** Complete pending work before deletion, or use soft delete

### Performance Tips

1. Index frequently queried fields (company_name, status)
2. Use GIN indexes for array fields (projects, sites)
3. Cache active contractors list
4. Batch contractor imports with transactions
5. Use partial indexes on deleted_at for active queries

### Debugging

**Check contractor existence:**
```sql
SELECT uuid, company_name, status, contract_start_date, contract_end_date
FROM contractors
WHERE pan = 'AABCU9603R' AND deleted_at IS NULL;
```

**Find expiring contracts:**
```sql
SELECT company_name, contract_end_date,
       contract_end_date - CURRENT_DATE as days_remaining
FROM contractors
WHERE status = 'active'
  AND contract_end_date IS NOT NULL
  AND contract_end_date > CURRENT_DATE
  AND contract_end_date < CURRENT_DATE + INTERVAL '30 days'
ORDER BY contract_end_date;
```

**Verify project assignments:**
```sql
SELECT c.company_name, c.associated_project
FROM contractors c
WHERE c.associated_project && ARRAY['project-uuid']::TEXT[]
  AND c.deleted_at IS NULL;
```

### Business Rules

1. **Contract Validity:** Active contracts must have start_date <= today <= end_date
2. **Renewal:** Create new contract record or extend end_date
3. **Termination:** Update status to 'terminated' and set end_date
4. **Compliance:** GST and PAN mandatory for tax purposes
5. **Contact Person:** Must have valid user account with email/phone

### Validation Rules

**GST Number:**
- Format: 15 characters (2 state code + 10 PAN + 1 entity code + 1 Z + 1 checksum)
- Example: 27AABCU9603R1ZM

**PAN Number:**
- Format: 10 characters (AAA + C + A + 9999 + A)
- Example: AABCU9603R

**Contract Dates:**
- contract_end_date must be NULL or > contract_start_date
- Active contracts should have end_date in future

## Additional Resources

- [GST Information](https://www.gst.gov.in/)
- [PAN Card Details](https://www.incometax.gov.in/iec/foportal/)
- [Contract Management Best Practices](https://www.contractworks.com/contract-management-best-practices)
- [Vendor Management](https://www.investopedia.com/terms/v/vendor.asp)
