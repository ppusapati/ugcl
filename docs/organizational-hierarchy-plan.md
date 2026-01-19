# Organizational Hierarchy & User Management - System Design

## 1. Business Requirements Summary

### 1.1 Organizational Structure
```
Tenant (Business)
├── Divisions (Multiple)
│   ├── Branches (Multiple per Division)
│   └── Departments (Can be Division-specific OR Cross-cutting)
└── Departments (Business-level, cross-cutting all divisions)
```

**Key Points:**
- **Tenant** = Business entity (already exists)
- **Division** = Major business unit (e.g., Manufacturing, Sales, Services)
- **Branch** = Physical location/office under a division
- **Department** = Functional unit (HR, Finance, IT, etc.)
  - Can be **Business-level** (cross-cutting all divisions)
  - Can be **Division-level** (specific to one division)

### 1.2 User Types & Access Patterns

| User Type | Belongs To | Access Scope | Multi-Branch Access |
|-----------|------------|--------------|---------------------|
| **Employee** | Single Tenant | Single/Multiple Branches | Yes (Regional Manager, Area Manager) |
| **Vendor** | External | Multiple Projects/Branches | Yes |
| **Contractor** | External | Multiple Projects | Yes (temporary) |
| **Admin** | Single Tenant | Business/Division/Branch/Department | Yes (hierarchical) |

### 1.3 Profile Data Requirements

| Data Category | Employee | Vendor | Contractor | Storage Type |
|---------------|----------|--------|------------|--------------|
| Authentication | ✓ | ✓ | ✓ | DB (users table) |
| Basic Profile | ✓ | ✓ | ✓ | DB (user_profiles) |
| DOB, Identity Docs | ✓ | ✓ | ✓ | Files + DB references |
| Certificates | ✓ | ✓ | ✓ | Files + DB references |
| Family Details | ✓ | ✗ | ✗ | DB (employee_family) |
| Emergency Contact | ✓ | ✓ | ✓ | DB (user_profiles) |
| Employment Details | ✓ | ✗ | ✗ | DB (employee_profiles) |
| Vendor Details | ✗ | ✓ | ✗ | DB (vendor_profiles) |
| Contract Details | ✗ | ✗ | ✓ | DB (contractor_profiles) |

### 1.4 Access Control Rules
- ✓ Users belong to **ONE tenant only** (no cross-tenant access)
- ✓ Users can have **different roles** in different divisions/branches
- ✓ Permissions **cascade down** the hierarchy
- ✓ Users can be assigned at **multiple levels** with exceptions
- ✓ Branch-specific exceptions can override division-level access

---

## 2. Data Model Design

### 2.1 Organizational Hierarchy Tables

#### 2.1.1 `tenants` (Already exists)
Represents the Business entity.

#### 2.1.2 `divisions` (NEW)
```sql
CREATE TABLE divisions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL,              -- DIV001, MFG, SALES
    name VARCHAR(255) NOT NULL,             -- Manufacturing Division
    description TEXT,
    head_user_id BIGINT,                    -- Division head/manager
    is_active BOOLEAN DEFAULT true,
    display_order INTEGER DEFAULT 0,
    metadata JSONB DEFAULT '{}',            -- Custom fields
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, code)
);
```

#### 2.1.3 `branches` (NEW)
```sql
CREATE TABLE branches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    division_id UUID NOT NULL REFERENCES divisions(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL,              -- BR001, MUM-001
    name VARCHAR(255) NOT NULL,             -- Mumbai Branch
    branch_type VARCHAR(50),                -- HQ, Regional, Local

    -- Location Details
    address_line1 TEXT,
    address_line2 TEXT,
    city VARCHAR(100),
    state VARCHAR(100),
    country VARCHAR(100) DEFAULT 'India',
    postal_code VARCHAR(20),
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),

    -- Contact Details
    phone VARCHAR(20),
    email VARCHAR(255),

    -- Management
    branch_manager_user_id BIGINT,          -- Branch manager
    is_active BOOLEAN DEFAULT true,
    display_order INTEGER DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, code)
);
```

#### 2.1.4 `departments` (NEW)
```sql
CREATE TABLE departments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    division_id UUID NULL REFERENCES divisions(id) ON DELETE CASCADE,  -- NULL = Business-level
    code VARCHAR(50) NOT NULL,              -- HR, FIN, IT
    name VARCHAR(255) NOT NULL,             -- Human Resources
    description TEXT,
    department_type VARCHAR(50),            -- Operational, Support, Administrative
    head_user_id BIGINT,                    -- Department head (GM, VP)
    parent_department_id UUID REFERENCES departments(id), -- For sub-departments
    is_active BOOLEAN DEFAULT true,
    display_order INTEGER DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, division_id, code),
    CHECK (
        -- Either business-level (division_id IS NULL)
        -- OR division-level (division_id IS NOT NULL)
        (division_id IS NULL AND parent_department_id IS NULL) OR
        (division_id IS NOT NULL)
    )
);
```

**Department Scope Examples:**
- `HR, division_id=NULL` → Business-level HR (across all divisions)
- `HR, division_id=DIV001` → Manufacturing Division's HR
- `IT, division_id=NULL` → Business-level IT
- `Finance, division_id=NULL` → Business-level Finance

### 2.2 User & Profile Tables

#### 2.2.1 `users` (Update existing)
Core authentication table - **already exists**, minor updates needed:

```sql
-- Add user_type enum
CREATE TYPE user_type AS ENUM ('employee', 'vendor', 'contractor', 'admin');

-- Add to users table:
ALTER TABLE users ADD COLUMN user_type user_type DEFAULT 'employee';
ALTER TABLE users ADD COLUMN tenant_id UUID REFERENCES tenants(id); -- Single tenant
```

#### 2.2.2 `user_profiles` (NEW - Common profile data)
```sql
CREATE TABLE user_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Basic Info
    date_of_birth DATE,
    blood_group VARCHAR(10),
    nationality VARCHAR(100) DEFAULT 'Indian',

    -- Emergency Contact
    emergency_contact_name VARCHAR(255),
    emergency_contact_relation VARCHAR(100),
    emergency_contact_phone VARCHAR(20),
    emergency_contact_email VARCHAR(255),

    -- Address
    current_address_line1 TEXT,
    current_address_line2 TEXT,
    current_city VARCHAR(100),
    current_state VARCHAR(100),
    current_postal_code VARCHAR(20),

    permanent_address_line1 TEXT,
    permanent_address_line2 TEXT,
    permanent_city VARCHAR(100),
    permanent_state VARCHAR(100),
    permanent_postal_code VARCHAR(20),

    -- Metadata
    profile_photo_url TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id)
);
```

#### 2.2.3 `user_identity_documents` (NEW - KYC documents)
```sql
CREATE TABLE user_identity_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    document_type VARCHAR(50) NOT NULL,     -- AADHAR, PAN, PASSPORT, DRIVING_LICENSE
    document_number VARCHAR(100) NOT NULL,
    document_file_path TEXT,                -- File storage path
    document_file_url TEXT,                 -- Accessible URL
    issued_date DATE,
    expiry_date DATE,
    issuing_authority VARCHAR(255),
    is_verified BOOLEAN DEFAULT false,
    verified_at TIMESTAMP,
    verified_by BIGINT,                     -- Admin user who verified
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, document_type)
);
```

#### 2.2.4 `employee_profiles` (NEW - Employee-specific)
```sql
CREATE TABLE employee_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Employment Details
    employee_code VARCHAR(50) UNIQUE NOT NULL,
    date_of_joining DATE NOT NULL,
    date_of_leaving DATE,
    employment_type VARCHAR(50),            -- Permanent, Contract, Intern
    designation VARCHAR(255),
    grade VARCHAR(50),

    -- Primary Assignment
    primary_branch_id UUID REFERENCES branches(id),
    primary_department_id UUID REFERENCES departments(id),
    primary_division_id UUID REFERENCES divisions(id),

    -- Reporting
    reporting_manager_user_id BIGINT REFERENCES users(id),

    -- Status
    is_active BOOLEAN DEFAULT true,
    probation_end_date DATE,
    confirmation_date DATE,

    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id)
);
```

#### 2.2.5 `employee_family_details` (NEW)
```sql
CREATE TABLE employee_family_details (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Family Member Info
    name VARCHAR(255) NOT NULL,
    relation VARCHAR(100) NOT NULL,        -- Spouse, Child, Parent, Sibling
    date_of_birth DATE,
    gender gender_type,
    occupation VARCHAR(255),
    phone VARCHAR(20),

    -- Dependent Info
    is_dependent BOOLEAN DEFAULT false,
    is_nominee BOOLEAN DEFAULT false,
    nominee_percentage DECIMAL(5,2),       -- For insurance/benefits

    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 2.2.6 `vendor_profiles` (NEW)
```sql
CREATE TABLE vendor_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Vendor Business Details
    vendor_code VARCHAR(50) UNIQUE NOT NULL,
    company_name VARCHAR(255) NOT NULL,
    company_registration_number VARCHAR(100),
    gst_number VARCHAR(50),
    pan_number VARCHAR(50),

    -- Business Address
    business_address_line1 TEXT,
    business_address_line2 TEXT,
    business_city VARCHAR(100),
    business_state VARCHAR(100),
    business_postal_code VARCHAR(20),

    -- Contact Person
    contact_person_name VARCHAR(255),
    contact_person_designation VARCHAR(100),
    contact_person_phone VARCHAR(20),
    contact_person_email VARCHAR(255),

    -- Vendor Classification
    vendor_type VARCHAR(100),              -- Product, Service, Both
    vendor_category VARCHAR(100),          -- IT, Construction, Supplies, etc.

    -- Onboarding
    onboarded_date DATE,
    is_active BOOLEAN DEFAULT true,

    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id)
);
```

#### 2.2.7 `contractor_profiles` (NEW)
```sql
CREATE TABLE contractor_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Contractor Details
    contractor_code VARCHAR(50) UNIQUE NOT NULL,
    contract_start_date DATE NOT NULL,
    contract_end_date DATE NOT NULL,
    contract_type VARCHAR(100),            -- Fixed-term, Project-based

    -- Skills & Specialization
    specialization VARCHAR(255),
    skills TEXT[],

    -- Rate/Compensation
    billing_rate DECIMAL(10,2),
    billing_currency VARCHAR(3) DEFAULT 'INR',
    billing_cycle VARCHAR(50),             -- Hourly, Daily, Weekly, Monthly

    -- Assignment
    vendor_id UUID,                        -- If contractor is via vendor

    -- Status
    is_active BOOLEAN DEFAULT true,

    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id)
);
```

#### 2.2.8 `user_certificates` (NEW - Professional certificates)
```sql
CREATE TABLE user_certificates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    certificate_name VARCHAR(255) NOT NULL,
    certificate_type VARCHAR(100),         -- Educational, Professional, Training
    issuing_organization VARCHAR(255),
    issue_date DATE,
    expiry_date DATE,                      -- NULL if no expiry
    certificate_number VARCHAR(100),

    -- File Storage
    certificate_file_path TEXT,
    certificate_file_url TEXT,

    -- Verification
    is_verified BOOLEAN DEFAULT false,
    verified_at TIMESTAMP,
    verified_by BIGINT,

    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 2.3 User Access & Assignment Tables

#### 2.3.1 `user_organizational_assignments` (NEW - Multi-level access)
```sql
CREATE TABLE user_organizational_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Hierarchical Assignment (nullable for multi-level)
    division_id UUID REFERENCES divisions(id) ON DELETE CASCADE,
    branch_id UUID REFERENCES branches(id) ON DELETE CASCADE,
    department_id UUID REFERENCES departments(id) ON DELETE CASCADE,

    -- Assignment Details
    assignment_type VARCHAR(50) NOT NULL,  -- Primary, Secondary, Temporary
    assignment_level VARCHAR(50) NOT NULL, -- Business, Division, Branch, Department

    -- Roles at this level
    roles TEXT[] DEFAULT '{}',

    -- Validity
    valid_from DATE DEFAULT CURRENT_DATE,
    valid_until DATE,                      -- NULL = indefinite
    is_active BOOLEAN DEFAULT true,

    -- Priority for conflict resolution
    priority INTEGER DEFAULT 100,          -- Lower = higher priority

    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CHECK (
        -- Must have at least tenant_id
        tenant_id IS NOT NULL AND
        -- Assignment level must match filled fields
        (
            (assignment_level = 'Business' AND division_id IS NULL AND branch_id IS NULL AND department_id IS NULL) OR
            (assignment_level = 'Division' AND division_id IS NOT NULL AND branch_id IS NULL) OR
            (assignment_level = 'Branch' AND branch_id IS NOT NULL) OR
            (assignment_level = 'Department' AND department_id IS NOT NULL)
        )
    )
);
```

**Examples:**
```sql
-- Regional Manager: Access to multiple branches in a division
INSERT INTO user_organizational_assignments (user_id, tenant_id, division_id, assignment_level, roles)
VALUES (123, 'tenant-uuid', 'div-uuid', 'Division', ARRAY['Regional_Manager']);

-- Branch-specific Employee
INSERT INTO user_organizational_assignments (user_id, tenant_id, branch_id, assignment_level, roles)
VALUES (456, 'tenant-uuid', 'branch-uuid', 'Branch', ARRAY['Employee']);

-- Department Head (Business-level HR GM)
INSERT INTO user_organizational_assignments (user_id, tenant_id, department_id, assignment_level, roles)
VALUES (789, 'tenant-uuid', 'hr-dept-uuid', 'Department', ARRAY['General_Manager']);
```

#### 2.3.2 `user_project_assignments` (NEW - For vendors/contractors)
```sql
CREATE TABLE user_project_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id UUID NOT NULL,              -- Reference to project module

    -- Scope within project
    branch_ids UUID[],                     -- Branches this assignment covers
    division_ids UUID[],

    -- Assignment Details
    assignment_start_date DATE NOT NULL,
    assignment_end_date DATE,
    roles TEXT[] DEFAULT '{}',

    is_active BOOLEAN DEFAULT true,

    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 2.4 Permission & Role Tables (Update existing)

#### 2.4.1 Update `user_tenant_roles` (existing table)
```sql
-- ADD columns for hierarchy context
ALTER TABLE user_tenant_roles
    ADD COLUMN division_id UUID REFERENCES divisions(id),
    ADD COLUMN branch_id UUID REFERENCES branches(id),
    ADD COLUMN department_id UUID REFERENCES departments(id),
    ADD COLUMN access_scope VARCHAR(50) DEFAULT 'Business';  -- Business, Division, Branch, Department

-- Add check constraint
ALTER TABLE user_tenant_roles ADD CONSTRAINT check_scope_consistency
CHECK (
    (access_scope = 'Business' AND division_id IS NULL AND branch_id IS NULL AND department_id IS NULL) OR
    (access_scope = 'Division' AND division_id IS NOT NULL AND branch_id IS NULL AND department_id IS NULL) OR
    (access_scope = 'Branch' AND branch_id IS NOT NULL) OR
    (access_scope = 'Department' AND department_id IS NOT NULL)
);
```

---

## 3. Service Architecture

### 3.1 Module Boundaries

```
identity/
├── tenant/          (Existing - manage tenant/business)
├── user/            (Existing - authentication & user management)
└── organization/    (NEW - organizational hierarchy)
    ├── divisions
    ├── branches
    └── departments

profile/             (NEW - user profile management)
├── employee/
├── vendor/
├── contractor/
└── documents/
```

### 3.2 New Services Required

#### 3.2.1 Organization Service
```
- Division CRUD
- Branch CRUD
- Department CRUD
- Hierarchy queries (get all branches in division, etc.)
```

#### 3.2.2 Profile Service
```
- User Profile CRUD (common profile)
- Employee Profile CRUD
- Vendor Profile CRUD
- Contractor Profile CRUD
- Family Details CRUD (employee only)
- Document Management (upload, verify, retrieve)
- Certificate Management
```

#### 3.2.3 Assignment Service
```
- Assign user to organizational units
- Multi-branch/division assignments
- Project assignments for vendors/contractors
- Role assignment at different levels
```

---

## 4. Permission & Access Control Strategy

### 4.1 Permission Hierarchy

```
Business Level (Tenant)
    ↓ (inherits down)
Division Level
    ↓ (inherits down)
Branch Level
    ↓ (can have exceptions)
User Level (explicit grants/denies)
```

### 4.2 Permission Scope Examples

| Role | Scope | Permissions |
|------|-------|-------------|
| CEO | Business | Full access to all divisions, branches, departments |
| Division Head | Division | Full access within assigned division(s) |
| Regional Manager | Division | Access to multiple branches within division |
| Branch Manager | Branch | Full access to assigned branch |
| HR General Manager | Department (Business-level) | Access to HR data across all divisions/branches |
| HR Manager | Department (Division-level) | Access to HR data within specific division |
| Employee | Branch | Access to own branch resources |
| Vendor | Project-based | Access to specific project branches |

### 4.3 Access Resolution Algorithm

```python
def resolve_user_access(user_id, resource, action):
    # 1. Get all assignments for user (ordered by priority)
    assignments = get_user_assignments(user_id, order_by='priority ASC')

    # 2. Check explicit denies first (highest priority)
    if has_explicit_deny(user_id, resource, action):
        return DENY

    # 3. Check explicit grants
    if has_explicit_grant(user_id, resource, action):
        return ALLOW

    # 4. Check role-based permissions (cascade from business → division → branch)
    for assignment in assignments:
        if assignment.has_permission(resource, action):
            return ALLOW

    # 5. Default deny
    return DENY
```

---

## 5. API Design (Proto Messages)

### 5.1 Organization Service (tenant module extension)

```protobuf
// divisions
message Division {
    string id = 1;
    string tenant_id = 2;
    string code = 3;
    string name = 4;
    string description = 5;
    string head_user_id = 6;
    bool is_active = 7;
    int32 display_order = 8;
    google.protobuf.Struct metadata = 9;
    google.protobuf.Timestamp created_at = 10;
    google.protobuf.Timestamp updated_at = 11;
}

// branches
message Branch {
    string id = 1;
    string division_id = 2;
    string tenant_id = 3;
    string code = 4;
    string name = 5;
    string branch_type = 6;
    Address address = 7;
    Location location = 8;
    ContactInfo contact = 9;
    string branch_manager_user_id = 10;
    bool is_active = 11;
    int32 display_order = 12;
    google.protobuf.Struct metadata = 13;
    google.protobuf.Timestamp created_at = 14;
    google.protobuf.Timestamp updated_at = 15;
}

message Address {
    string line1 = 1;
    string line2 = 2;
    string city = 3;
    string state = 4;
    string country = 5;
    string postal_code = 6;
}

message Location {
    double latitude = 1;
    double longitude = 2;
}

message ContactInfo {
    string phone = 1;
    string email = 2;
}

// departments
message Department {
    string id = 1;
    string tenant_id = 2;
    string division_id = 3;  // NULL for business-level
    string code = 4;
    string name = 5;
    string description = 6;
    string department_type = 7;
    string head_user_id = 8;
    string parent_department_id = 9;
    bool is_active = 10;
    int32 display_order = 11;
    google.protobuf.Struct metadata = 12;
    google.protobuf.Timestamp created_at = 13;
    google.protobuf.Timestamp updated_at = 14;
}

service OrganizationService {
    // Divisions
    rpc CreateDivision(CreateDivisionRequest) returns (Division) {}
    rpc UpdateDivision(UpdateDivisionRequest) returns (Division) {}
    rpc DeleteDivision(DeleteDivisionRequest) returns (google.protobuf.Empty) {}
    rpc GetDivision(GetDivisionRequest) returns (Division) {}
    rpc ListDivisions(ListDivisionsRequest) returns (ListDivisionsResponse) {}

    // Branches
    rpc CreateBranch(CreateBranchRequest) returns (Branch) {}
    rpc UpdateBranch(UpdateBranchRequest) returns (Branch) {}
    rpc DeleteBranch(DeleteBranchRequest) returns (google.protobuf.Empty) {}
    rpc GetBranch(GetBranchRequest) returns (Branch) {}
    rpc ListBranches(ListBranchesRequest) returns (ListBranchesResponse) {}
    rpc ListBranchesByDivision(ListBranchesByDivisionRequest) returns (ListBranchesResponse) {}

    // Departments
    rpc CreateDepartment(CreateDepartmentRequest) returns (Department) {}
    rpc UpdateDepartment(UpdateDepartmentRequest) returns (Department) {}
    rpc DeleteDepartment(DeleteDepartmentRequest) returns (google.protobuf.Empty) {}
    rpc GetDepartment(GetDepartmentRequest) returns (Department) {}
    rpc ListDepartments(ListDepartmentsRequest) returns (ListDepartmentsResponse) {}
    rpc ListDepartmentsByDivision(ListDepartmentsByDivisionRequest) returns (ListDepartmentsResponse) {}
}
```

### 5.2 Profile Service (new module)

```protobuf
enum UserType {
    USER_TYPE_UNSPECIFIED = 0;
    USER_TYPE_EMPLOYEE = 1;
    USER_TYPE_VENDOR = 2;
    USER_TYPE_CONTRACTOR = 3;
    USER_TYPE_ADMIN = 4;
}

message UserProfile {
    string id = 1;
    int64 user_id = 2;
    google.protobuf.Timestamp date_of_birth = 3;
    string blood_group = 4;
    string nationality = 5;
    EmergencyContact emergency_contact = 6;
    Address current_address = 7;
    Address permanent_address = 8;
    string profile_photo_url = 9;
    google.protobuf.Struct metadata = 10;
}

message EmergencyContact {
    string name = 1;
    string relation = 2;
    string phone = 3;
    string email = 4;
}

message EmployeeProfile {
    string id = 1;
    int64 user_id = 2;
    string employee_code = 3;
    google.protobuf.Timestamp date_of_joining = 4;
    google.protobuf.Timestamp date_of_leaving = 5;
    string employment_type = 6;
    string designation = 7;
    string grade = 8;
    string primary_branch_id = 9;
    string primary_department_id = 10;
    string primary_division_id = 11;
    string reporting_manager_user_id = 12;
    bool is_active = 13;
    google.protobuf.Timestamp probation_end_date = 14;
    google.protobuf.Timestamp confirmation_date = 15;
}

message VendorProfile {
    string id = 1;
    int64 user_id = 2;
    string vendor_code = 3;
    string company_name = 4;
    string company_registration_number = 5;
    string gst_number = 6;
    string pan_number = 7;
    Address business_address = 8;
    ContactPerson contact_person = 9;
    string vendor_type = 10;
    string vendor_category = 11;
    google.protobuf.Timestamp onboarded_date = 12;
    bool is_active = 13;
}

message ContractorProfile {
    string id = 1;
    int64 user_id = 2;
    string contractor_code = 3;
    google.protobuf.Timestamp contract_start_date = 4;
    google.protobuf.Timestamp contract_end_date = 5;
    string contract_type = 6;
    string specialization = 7;
    repeated string skills = 8;
    double billing_rate = 9;
    string billing_currency = 10;
    string billing_cycle = 11;
    bool is_active = 12;
}

service ProfileService {
    // Common Profile
    rpc CreateUserProfile(CreateUserProfileRequest) returns (UserProfile) {}
    rpc UpdateUserProfile(UpdateUserProfileRequest) returns (UserProfile) {}
    rpc GetUserProfile(GetUserProfileRequest) returns (UserProfile) {}

    // Employee
    rpc CreateEmployeeProfile(CreateEmployeeProfileRequest) returns (EmployeeProfile) {}
    rpc UpdateEmployeeProfile(UpdateEmployeeProfileRequest) returns (EmployeeProfile) {}
    rpc GetEmployeeProfile(GetEmployeeProfileRequest) returns (EmployeeProfile) {}

    // Vendor
    rpc CreateVendorProfile(CreateVendorProfileRequest) returns (VendorProfile) {}
    rpc UpdateVendorProfile(UpdateVendorProfileRequest) returns (VendorProfile) {}
    rpc GetVendorProfile(GetVendorProfileRequest) returns (VendorProfile) {}

    // Contractor
    rpc CreateContractorProfile(CreateContractorProfileRequest) returns (ContractorProfile) {}
    rpc UpdateContractorProfile(UpdateContractorProfileRequest) returns (ContractorProfile) {}
    rpc GetContractorProfile(GetContractorProfileRequest) returns (ContractorProfile) {}

    // Documents
    rpc UploadIdentityDocument(UploadIdentityDocumentRequest) returns (IdentityDocument) {}
    rpc UploadCertificate(UploadCertificateRequest) returns (Certificate) {}
    rpc ListUserDocuments(ListUserDocumentsRequest) returns (ListUserDocumentsResponse) {}
    rpc VerifyDocument(VerifyDocumentRequest) returns (google.protobuf.Empty) {}
}
```

### 5.3 Update User Service (user.proto)

```protobuf
// Extend User message
message User {
    // ... existing fields ...
    UserType user_type = 23;  // NEW
    string tenant_id = 24;    // NEW - single tenant binding
}

// Extend UserTenantRole
message UserTenantRole {
    string tenant_id = 1;
    repeated string roles = 2;
    bool is_active = 3;
    google.protobuf.Timestamp assigned_at = 4;

    // NEW - Hierarchy context
    string division_id = 5;
    string branch_id = 6;
    string department_id = 7;
    AccessScope access_scope = 8;
}

enum AccessScope {
    ACCESS_SCOPE_UNSPECIFIED = 0;
    ACCESS_SCOPE_BUSINESS = 1;
    ACCESS_SCOPE_DIVISION = 2;
    ACCESS_SCOPE_BRANCH = 3;
    ACCESS_SCOPE_DEPARTMENT = 4;
}
```

---

## 6. Implementation Phases

### Phase 1: Foundation (Week 1-2)
- [ ] Create organizational hierarchy tables (divisions, branches, departments)
- [ ] Update user table with user_type and tenant_id
- [ ] Create SQLC queries for organizational CRUD
- [ ] Create proto definitions for Organization service
- [ ] Generate Connect handlers

### Phase 2: User Profiles (Week 3-4)
- [ ] Create profile tables (user_profiles, employee_profiles, vendor_profiles, contractor_profiles)
- [ ] Create document tables (identity_documents, certificates)
- [ ] Create SQLC queries for profile management
- [ ] Create proto definitions for Profile service
- [ ] Implement file storage for documents
- [ ] Generate Connect handlers

### Phase 3: Access Control (Week 5-6)
- [ ] Create user_organizational_assignments table
- [ ] Update user_tenant_roles with hierarchy fields
- [ ] Implement permission cascade logic
- [ ] Create assignment service APIs
- [ ] Implement access resolution algorithm

### Phase 4: Integration & Testing (Week 7-8)
- [ ] Integration tests for hierarchy queries
- [ ] Multi-level access tests
- [ ] Document upload/retrieval tests
- [ ] Performance optimization
- [ ] API documentation

---

## 7. Key Design Decisions

### 7.1 Why Separate Profile Tables?
- **Different attributes** for each user type
- **Better query performance** (no sparse columns)
- **Easier validation** (type-specific constraints)
- **Clearer data model**

### 7.2 Why `user_organizational_assignments` vs extending `user_tenant_roles`?
- **Multi-level assignments** (user can be assigned at business, division, branch, department levels)
- **Temporal assignments** (valid_from, valid_until for contractors)
- **Priority-based resolution** (handle conflicting permissions)
- **Assignment types** (primary vs secondary)

### 7.3 Department Scoping (Business vs Division level)
- `division_id IS NULL` → Business-level department (e.g., Corporate HR)
- `division_id IS NOT NULL` → Division-specific department (e.g., Manufacturing HR)
- Allows flexibility for different organizational structures

### 7.4 File Storage Strategy
- **Documents stored as files** (not in DB)
- **DB stores metadata + file paths**
- Use cloud storage (S3, Azure Blob, GCS) or local file system
- Store URLs for quick access

---

## 8. Database Indexes (Performance)

```sql
-- Division indexes
CREATE INDEX idx_divisions_tenant_id ON divisions(tenant_id);
CREATE INDEX idx_divisions_code ON divisions(tenant_id, code);
CREATE INDEX idx_divisions_is_active ON divisions(is_active);

-- Branch indexes
CREATE INDEX idx_branches_division_id ON branches(division_id);
CREATE INDEX idx_branches_tenant_id ON branches(tenant_id);
CREATE INDEX idx_branches_code ON branches(tenant_id, code);
CREATE INDEX idx_branches_is_active ON branches(is_active);
CREATE INDEX idx_branches_location ON branches USING GIST(ll_to_earth(latitude, longitude));

-- Department indexes
CREATE INDEX idx_departments_tenant_id ON departments(tenant_id);
CREATE INDEX idx_departments_division_id ON departments(division_id);
CREATE INDEX idx_departments_parent_id ON departments(parent_department_id);
CREATE INDEX idx_departments_code ON departments(tenant_id, code);

-- Assignment indexes
CREATE INDEX idx_org_assignments_user_id ON user_organizational_assignments(user_id);
CREATE INDEX idx_org_assignments_tenant_id ON user_organizational_assignments(tenant_id);
CREATE INDEX idx_org_assignments_division_id ON user_organizational_assignments(division_id);
CREATE INDEX idx_org_assignments_branch_id ON user_organizational_assignments(branch_id);
CREATE INDEX idx_org_assignments_department_id ON user_organizational_assignments(department_id);
CREATE INDEX idx_org_assignments_active ON user_organizational_assignments(is_active, valid_from, valid_until);

-- Profile indexes
CREATE INDEX idx_employee_profiles_employee_code ON employee_profiles(employee_code);
CREATE INDEX idx_employee_profiles_branch_id ON employee_profiles(primary_branch_id);
CREATE INDEX idx_employee_profiles_department_id ON employee_profiles(primary_department_id);
CREATE INDEX idx_vendor_profiles_vendor_code ON vendor_profiles(vendor_code);
CREATE INDEX idx_contractor_profiles_contractor_code ON contractor_profiles(contractor_code);
CREATE INDEX idx_contractor_profiles_dates ON contractor_profiles(contract_start_date, contract_end_date);

-- Document indexes
CREATE INDEX idx_identity_docs_user_id ON user_identity_documents(user_id);
CREATE INDEX idx_identity_docs_type ON user_identity_documents(document_type);
CREATE INDEX idx_identity_docs_verified ON user_identity_documents(is_verified);
CREATE INDEX idx_certificates_user_id ON user_certificates(user_id);
CREATE INDEX idx_certificates_type ON user_certificates(certificate_type);
```

---

## 9. Migration Strategy

### 9.1 Existing Data Migration
```sql
-- Migrate existing users to have tenant_id
UPDATE users u
SET tenant_id = utr.tenant_id
FROM user_tenant_roles utr
WHERE u.id = utr.user_id
  AND u.tenant_id IS NULL
LIMIT 1;  -- Take first tenant if multiple

-- Set default user_type for existing users
UPDATE users SET user_type = 'employee' WHERE user_type IS NULL;
```

### 9.2 Zero-Downtime Deployment
1. Add new tables without foreign keys
2. Add new columns as nullable
3. Backfill data
4. Add constraints
5. Update application code
6. Deploy

---

## 10. Security Considerations

### 10.1 Document Access Control
- Documents should only be accessible by:
  - Document owner
  - HR admins
  - Reporting managers
  - System admins
- Implement pre-signed URLs for temporary access
- Encrypt sensitive documents at rest

### 10.2 PII Data Protection
- Encrypt Aadhar, PAN, passport numbers
- Audit log all access to identity documents
- Implement data masking for non-authorized users
- GDPR/Data Protection compliance

### 10.3 Multi-level Access Validation
```python
def can_access_user_data(requester_id, target_user_id):
    # 1. Self-access always allowed
    if requester_id == target_user_id:
        return True

    # 2. Check if requester is reporting manager
    if is_reporting_manager(requester_id, target_user_id):
        return True

    # 3. Check if requester has HR/Admin role at appropriate level
    requester_scope = get_highest_scope(requester_id)
    target_scope = get_user_scope(target_user_id)

    if requester_has_hr_or_admin_role(requester_id):
        if requester_scope.covers(target_scope):
            return True

    return False
```

---

## 11. Requirements Clarification (ANSWERED)

1. **Projects**: ✅ Project module will be separate, not part of this implementation
   - For now, keep `user_project_assignments` table with `project_id` as UUID reference
   - Future Project module will integrate via this table

2. **Salary/Compensation**: ✅ Separate Payroll module
   - Remove salary fields from `employee_profiles`
   - Keep employment details (designation, grade, joining date)

3. **File Storage**: ✅ MinIO (S3-compatible object storage)
   - Need MinIO configuration in system
   - Document storage path: `{tenant_id}/{user_type}/{user_id}/documents/{document_id}.{ext}`
   - Pre-signed URLs for secure access

4. **User Onboarding**: ✅ Admin-created only
   - No self-registration
   - Admin creates user → User receives credentials → First login forces password change
   - Email/SMS verification required

5. **Audit Requirements**: ✅ YES - Full audit logging required
   - Track all profile changes (before/after values)
   - Track all assignment changes
   - Track document uploads/verifications
   - Track organizational hierarchy changes

6. **Reports**: ✅ All possible organizational reports needed
   - Headcount by Division/Branch/Department
   - User type distribution (Employee/Vendor/Contractor)
   - Active vs Inactive users
   - Document expiry tracking (certificates, contracts, identity docs)
   - Vendor/Contractor summary by project
   - Employee distribution by designation/grade
   - Branch-wise employee strength
   - Reporting hierarchy charts
   - New joiners/Leavers reports (monthly/quarterly)

---

## Next Steps

Once you review this plan and answer the clarification questions, we can:

1. ✅ Finalize the data model
2. ✅ Create database migration scripts
3. ✅ Generate proto files and SQLC code
4. ✅ Implement service handlers
5. ✅ Add validation and business logic
6. ✅ Create integration tests

Please review and let me know if any adjustments are needed!
