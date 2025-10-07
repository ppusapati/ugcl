# Employee Management Module

## 1. Module Overview

The Employee module manages permanent, temporary, and probationary employees within the UGCL platform. It handles employee records, organizational assignments, reporting hierarchies, personal information, employment history, and skill tracking.

**Purpose:** Comprehensive employee lifecycle management from onboarding to exit, with integration into organizational hierarchy and identity management.

**Key Features:**
- Employee profile management (personal, professional, financial)
- Organizational hierarchy assignment (division, branch, department)
- Reporting structure (manager, department head)
- Employment lifecycle tracking (joining, confirmation, leaving)
- Skills and certifications management
- Statutory compliance (PAN, Aadhaar, UAN, ESIC)
- Bank account management for payroll
- Emergency contact information
- Employee status management
- Integration with user identity system

## 2. Architecture

### Module Structure
```
employee/
├── proto/
│   └── employee.proto        # Service and message definitions
├── db/
│   ├── schema/
│   │   └── employees.sql    # Database schema
│   └── generated/           # SQLC generated code
├── repository/              # Data access layer
│   └── employee_repository.go
├── services/                # Business logic layer
│   └── employee_service.go
├── handlers/                # gRPC handlers
│   └── employee_handler.go
├── mappers/                 # DTO mappers
│   └── employee_mappers.go
└── module.go                # FX module definition
```

### Database Schema

**Tables:**
- `employees` - Complete employee information

**Key Relationships:**
- Employee → User (N:1 via user_id)
- Employee → Division (N:1 via division_id)
- Employee → Branch (N:1 via branch_id)
- Employee → Department (N:1 via department_id)
- Employee → Manager (N:1 self-reference via manager_id)

### Key Dependencies
- `identity/user` - User account management
- `identity/entity` - Entity abstraction
- `organization` - Organizational hierarchy
- Connect-Go RPC framework
- PostgreSQL with array and JSONB support

## 3. Quick Start

### Create an Employee

```go
import (
    employeev2 "p9e.in/ugcl/employee/api/v2/employee"
    userv2 "p9e.in/ugcl/identity/user/api/v2/user"
    "connectrpc.com/connect"
)

client := employeev2.NewEmployeeServiceClient(httpClient, baseURL)

// Create employee with user account
resp, err := client.CreateEmployee(ctx, connect.NewRequest(&employeev2.CreateEmployeeRequest{
    Employee: &employeev2.Employee{
        EmployeeCode:   "EMP001",
        DivisionId:     wrapperspb.String("division-uuid"),
        BranchId:       wrapperspb.String("branch-uuid"),
        DepartmentId:   wrapperspb.String("dept-uuid"),
        Designation:    "Software Engineer",
        JobTitle:       wrapperspb.String("Senior Software Engineer"),
        JobGrade:       wrapperspb.String("E3"),
        EmployeeType:   wrapperspb.String("PERMANENT"),
        ManagerId:      wrapperspb.String("manager-employee-uuid"),
        DateOfJoining:  timestamppb.New(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
        Status:         "ACTIVE",
        Skills:         []string{"Go", "PostgreSQL", "gRPC"},
    },
    User: &userv2.User{
        Username:       wrapperspb.String("john.doe"),
        Email:          wrapperspb.String("john.doe@ugcl.com"),
        Fullname:       wrapperspb.String("John Doe"),
        Phone:          wrapperspb.String("+91-9876543210"),
        Password:       "SecurePassword123!",
        ConfirmPassword: "SecurePassword123!",
    },
}))
```

### Common Operations

**Get employee by code:**
```go
employee, err := client.GetEmployee(ctx, connect.NewRequest(&employeev2.EmployeeIdentifier{
    Identifier: &employeev2.EmployeeIdentifier_EmployeeCode{
        EmployeeCode: "EMP001",
    },
}))
```

**List employees by department:**
```go
employees, err := client.GetEmployeesByDepartment(ctx, connect.NewRequest(&employeev2.DepartmentIdentifier{
    DepartmentId: "dept-uuid",
}))
```

**Update employee information:**
```go
updated, err := client.UpdateEmployee(ctx, connect.NewRequest(&employeev2.UpdateEmployeeRequest{
    Employee: &employeev2.Employee{
        Id:          employeeUUID,
        Designation: "Lead Software Engineer",
        JobGrade:    wrapperspb.String("E4"),
        Status:      "ACTIVE",
    },
    UpdateMask: &fieldmaskpb.FieldMask{
        Paths: []string{"designation", "job_grade"},
    },
}))
```

## 4. API Reference

### Employee Operations
- `CreateEmployee` - Create new employee with user account
- `UpdateEmployee` - Update employee details with field mask
- `GetEmployee` - Get employee by ID or employee code
- `ListEmployees` - List employees with filtering and pagination
- `DeleteEmployee` - Soft delete employee record
- `GetEmployeeByUserId` - Find employee by user ID
- `GetEmployeesByDepartment` - Get all employees in department
- `GetEmployeesByManager` - Get direct reports of manager

## 5. Database Schema

### Table: employees

```sql
-- Identity fields
- id (BIGSERIAL, PK)
- uuid (UUID, UNIQUE, generated)
- employee_code (VARCHAR(50), UNIQUE)
- user_id (UUID, FK → identity.users)

-- Organizational hierarchy
- division_id (UUID, FK → organization.divisions)
- branch_id (UUID, FK → organization.branches)
- department_id (UUID, FK → organization.departments)

-- Job information
- designation (VARCHAR(100))
- job_title (VARCHAR(150))
- job_grade (VARCHAR(50))
- employee_type (VARCHAR(50)) -- PERMANENT, TEMPORARY, PROBATION

-- Reporting structure
- manager_id (UUID, FK → employees.uuid)
- department_head_id (UUID, FK → employees.uuid)

-- Employment dates
- date_of_joining (TIMESTAMP)
- date_of_confirmation (TIMESTAMP)
- date_of_leaving (TIMESTAMP)

-- Statutory information
- pan (VARCHAR(20), UNIQUE)
- aadhaar (VARCHAR(20))
- uan (VARCHAR(50)) -- PF Universal Account Number
- esic_number (VARCHAR(50))

-- Bank details
- bank_name (VARCHAR(100))
- bank_account_number (VARCHAR(50))
- bank_ifsc (VARCHAR(20))

-- Address
- current_address (TEXT)
- permanent_address (TEXT)

-- Emergency contact
- emergency_contact_name (VARCHAR(100))
- emergency_contact_phone (VARCHAR(20))
- emergency_contact_relation (VARCHAR(50))

-- Work location
- work_location (VARCHAR(150))
- office_phone (VARCHAR(20))
- extension (VARCHAR(10))

-- Status
- status (VARCHAR(50)) -- ACTIVE, INACTIVE, ON_LEAVE, TERMINATED, RESIGNED
- termination_reason (TEXT)

-- Skills and qualifications
- skills (TEXT[]) -- Array of skills
- certifications (TEXT[]) -- Array of certifications
- highest_qualification (VARCHAR(100))

-- Metadata and audit
- metadata (JSONB)
- created_at, updated_at, created_by, updated_by
- deleted_at (soft delete)
```

### Constraints
- `chk_employment_dates` - date_of_leaving > date_of_joining
- `chk_confirmation_date` - date_of_confirmation >= date_of_joining
- PAN must be unique across all employees

### Indexes
- `idx_employees_user_id` - Fast user lookup
- `idx_employees_employee_code` - Unique code lookup
- `idx_employees_division`, `branch`, `department` - Organizational queries
- `idx_employees_manager` - Reporting hierarchy queries
- `idx_employees_status` - Filter by employment status
- `idx_employees_skills` - GIN index for skill search
- `idx_employees_certifications` - GIN index for certification search

## 6. Configuration

### Environment Variables
```env
# Database connection managed by packages/database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=ugcl
```

### Module-Specific Settings
- Employee code generation pattern (configurable)
- Probation period duration (default: 6 months)
- Mandatory fields for different employee types

## 7. Examples

### Complete Employee Onboarding

```go
// 1. Create employee record
employee, err := client.CreateEmployee(ctx, connect.NewRequest(&employeev2.CreateEmployeeRequest{
    Employee: &employeev2.Employee{
        EmployeeCode:   "EMP" + nextSequence,
        DivisionId:     wrapperspb.String(divisionID),
        BranchId:       wrapperspb.String(branchID),
        DepartmentId:   wrapperspb.String(deptID),
        Designation:    "Software Engineer",
        JobGrade:       wrapperspb.String("E2"),
        EmployeeType:   wrapperspb.String("PROBATION"),
        ManagerId:      wrapperspb.String(managerUUID),
        DateOfJoining:  timestamppb.Now(),
        Pan:            wrapperspb.String("ABCDE1234F"),
        BankName:       wrapperspb.String("HDFC Bank"),
        BankAccountNumber: wrapperspb.String("12345678901234"),
        BankIfsc:       wrapperspb.String("HDFC0001234"),
        CurrentAddress: wrapperspb.String("123 Main St, Mumbai"),
        WorkLocation:   wrapperspb.String("Mumbai HQ"),
        Status:         "ACTIVE",
        Skills:         []string{"Python", "Django", "PostgreSQL"},
        HighestQualification: wrapperspb.String("B.Tech Computer Science"),
    },
    User: &userv2.User{
        Username:       wrapperspb.String("jane.smith"),
        Email:          wrapperspb.String("jane.smith@ugcl.com"),
        Fullname:       wrapperspb.String("Jane Smith"),
        Phone:          wrapperspb.String("+91-9876543210"),
        Password:       "InitialPassword123!",
        ConfirmPassword: "InitialPassword123!",
    },
}))

employeeUUID := employee.Msg.Uuid.Value

// 2. After probation period, confirm employee
confirmationDate := time.Now().AddDate(0, 6, 0) // 6 months later
confirmed, err := client.UpdateEmployee(ctx, connect.NewRequest(&employeev2.UpdateEmployeeRequest{
    Employee: &employeev2.Employee{
        Uuid:                wrapperspb.String(employeeUUID),
        EmployeeType:        wrapperspb.String("PERMANENT"),
        DateOfConfirmation:  timestamppb.New(confirmationDate),
    },
    UpdateMask: &fieldmaskpb.FieldMask{
        Paths: []string{"employee_type", "date_of_confirmation"},
    },
}))
```

### Reporting Hierarchy Queries

```go
// Get all direct reports of a manager
reports, err := client.GetEmployeesByManager(ctx, connect.NewRequest(&employeev2.ManagerIdentifier{
    ManagerId: managerUUID,
}))

fmt.Printf("Manager has %d direct reports:\n", reports.Msg.TotalCount)
for _, emp := range reports.Msg.Employees {
    fmt.Printf("- %s (%s): %s\n",
        emp.GetFullname(),
        emp.EmployeeCode,
        emp.Designation)
}

// Get all employees in a department
deptEmployees, err := client.GetEmployeesByDepartment(ctx, connect.NewRequest(&employeev2.DepartmentIdentifier{
    DepartmentId: deptID,
}))
```

### Skill-Based Search

```go
// List employees with specific skills
engineers, err := client.ListEmployees(ctx, connect.NewRequest(&employeev2.ListEmployeesRequest{
    PageSize:   50,
    PageOffset: 0,
    Filter: &employeev2.EmployeeFilter{
        DepartmentId: wrapperspb.String("engineering-dept-uuid"),
        Status:       wrapperspb.String("ACTIVE"),
        Skills:       []string{"Go", "Kubernetes"},
    },
    Sort: []string{"designation", "date_of_joining"},
}))
```

## 8. Integration

### With Other Modules

**Identity/User Module:**
- Each employee has a user account for authentication
- User.uuid → Employee.user_id relationship
- Password management through identity/user

**Identity/Entity Module:**
- Employee record creates corresponding entity
- Entity.reference_id → Employee.uuid
- Entity provides unified identity and permissions

**Organization Module:**
- Employees assigned to divisions, branches, departments
- Validates organizational unit existence
- Supports org hierarchy queries

**Payroll/HR Modules (Future):**
- Bank details for salary processing
- UAN, ESIC for statutory compliance
- Employment dates for tenure calculation

### Event Publishing
```go
// Employee lifecycle events
type EmployeeEvent struct {
    Type       string    // created, updated, confirmed, terminated, resigned
    EmployeeID string
    UserID     string
    Timestamp  time.Time
    Changes    map[string]interface{}
}
```

## 9. Development

### Modifying the Module

**Add New Field:**
1. Update `employees.sql` schema
2. Update `employee.proto` message
3. Run migration: `make migrate-up`
4. Regenerate SQLC: `sqlc generate -f employee/db/sqlc.yaml`
5. Regenerate proto: `buf generate`
6. Update mappers and services

**Add Employment Type:**
```sql
-- Add to employee_type values
ALTER TABLE employees ADD CONSTRAINT chk_employee_type
    CHECK (employee_type IN ('PERMANENT', 'TEMPORARY', 'PROBATION', 'CONTRACT'));
```

### Generate Code
```bash
# Generate proto stubs
buf generate

# Generate SQLC queries
sqlc generate -f employee/db/sqlc.yaml

# Run migrations
make migrate-up
```

### Testing Guidelines

**Unit Tests:**
```go
func TestCreateEmployee(t *testing.T) {
    mockRepo := &MockEmployeeRepository{}
    mockUserService := &MockUserService{}
    svc := services.NewEmployeeService(mockRepo, mockUserService)

    emp, err := svc.CreateEmployee(ctx, req)
    assert.NoError(t, err)
    assert.Equal(t, "EMP001", emp.EmployeeCode)
    assert.Equal(t, "ACTIVE", emp.Status)
}
```

**Integration Tests:**
- Test employee creation with user creation
- Test reporting hierarchy
- Test organizational assignment
- Test soft delete behavior

## 10. Troubleshooting

### Common Issues

**Issue: Duplicate employee code**
- **Cause:** Employee code already exists
- **Solution:** Implement auto-increment or UUID-based code generation

**Issue: User creation fails**
- **Cause:** Email/username already exists
- **Solution:** Validate uniqueness before employee creation

**Issue: Invalid manager assignment**
- **Cause:** Manager doesn't exist or in different department
- **Solution:** Validate manager existence and organizational hierarchy

**Issue: PAN already registered**
- **Cause:** PAN must be unique across employees
- **Solution:** Check if employee with PAN exists, update if needed

**Issue: Cannot delete employee**
- **Cause:** Employee is referenced as manager by others
- **Solution:** Reassign reports before deletion

### Performance Tips

1. Index frequently queried fields (employee_code, user_id)
2. Use GIN indexes for array fields (skills, certifications)
3. Cache organizational hierarchy for validation
4. Batch employee imports with transactions
5. Use partial indexes on deleted_at for active queries

### Debugging

**Check employee existence:**
```sql
SELECT uuid, employee_code, designation, status
FROM employees
WHERE user_id = 'user-uuid' AND deleted_at IS NULL;
```

**Find reporting chain:**
```sql
WITH RECURSIVE reporting_chain AS (
    SELECT id, uuid, employee_code, manager_id, 1 as level
    FROM employees
    WHERE uuid = 'employee-uuid'

    UNION ALL

    SELECT e.id, e.uuid, e.employee_code, e.manager_id, rc.level + 1
    FROM employees e
    JOIN reporting_chain rc ON e.uuid = rc.manager_id
)
SELECT * FROM reporting_chain;
```

**Verify organizational assignment:**
```sql
SELECT
    e.employee_code,
    d.name as division,
    b.name as branch,
    dept.name as department
FROM employees e
LEFT JOIN divisions d ON e.division_id = d.id
LEFT JOIN branches b ON e.branch_id = b.id
LEFT JOIN departments dept ON e.department_id = dept.id
WHERE e.uuid = 'employee-uuid';
```

### Business Rules

1. **Probation Period:** Employees in PROBATION status for > 6 months should be reviewed
2. **Confirmation:** date_of_confirmation must be >= date_of_joining
3. **Termination:** Set date_of_leaving and update status to TERMINATED/RESIGNED
4. **Manager Hierarchy:** Managers must be in same or higher organizational level
5. **Statutory Compliance:** PAN, Aadhaar required for PERMANENT employees

## Additional Resources

- [Indian Labour Laws](https://labour.gov.in/)
- [EPF (Employee Provident Fund)](https://www.epfindia.gov.in/)
- [ESIC (Employee State Insurance)](https://www.esic.nic.in/)
- [PAN Card Information](https://www.incometax.gov.in/iec/foportal/help/individual/return-applicable-1)
