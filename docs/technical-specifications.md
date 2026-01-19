# Technical Specifications - Organizational Hierarchy & User Management

## 1. MinIO Document Storage

### 1.1 MinIO Configuration

```go
// packages/storage/minio/config.go
type MinioConfig struct {
    Endpoint        string // e.g., "localhost:9000"
    AccessKeyID     string
    SecretAccessKey string
    UseSSL          bool
    BucketName      string // e.g., "ugcl-documents"
    Region          string // e.g., "us-east-1"
}
```

### 1.2 Bucket Structure

```
ugcl-documents/
├── {tenant_id}/
│   ├── employee/
│   │   └── {user_id}/
│   │       ├── documents/
│   │       │   ├── aadhar_{document_id}.pdf
│   │       │   ├── pan_{document_id}.pdf
│   │       │   └── passport_{document_id}.pdf
│   │       ├── certificates/
│   │       │   ├── degree_{certificate_id}.pdf
│   │       │   └── training_{certificate_id}.pdf
│   │       └── profile/
│   │           └── avatar_{timestamp}.jpg
│   ├── vendor/
│   │   └── {user_id}/
│   │       └── documents/
│   │           ├── gst_{document_id}.pdf
│   │           ├── pan_{document_id}.pdf
│   │           └── company_registration_{document_id}.pdf
│   └── contractor/
│       └── {user_id}/
│           └── documents/
│               ├── contract_{document_id}.pdf
│               └── identity_{document_id}.pdf
```

### 1.3 File Naming Convention

```
{document_type}_{document_id}_{timestamp}.{extension}

Examples:
- aadhar_123e4567-e89b-12d3-a456-426614174000_1704067200.pdf
- pan_223e4567-e89b-12d3-a456-426614174001_1704067200.pdf
- degree_323e4567-e89b-12d3-a456-426614174002_1704067200.pdf
```

### 1.4 Document Upload Flow

```
1. Client → API: Upload request with file metadata
2. API → Validate file (type, size, virus scan)
3. API → Generate document_id (UUID)
4. API → Create DB record in user_identity_documents or user_certificates
5. API → Upload file to MinIO with path: {tenant_id}/{user_type}/{user_id}/documents/{filename}
6. MinIO → Returns object key
7. API → Update DB record with file_path and file_url
8. API → Generate pre-signed URL (valid for 1 hour)
9. API → Return response with document_id and pre-signed URL
10. API → Create audit log entry
```

### 1.5 Pre-signed URL Generation

```go
// Generate pre-signed URL for document access (valid for 1 hour)
func (s *DocumentService) GetDocumentURL(ctx context.Context, documentID uuid.UUID) (string, error) {
    // 1. Get document from DB
    doc, err := s.repo.GetDocument(ctx, documentID)
    if err != nil {
        return "", err
    }

    // 2. Check permissions
    if !s.canAccessDocument(ctx, doc) {
        return "", ErrUnauthorized
    }

    // 3. Generate pre-signed URL from MinIO
    url, err := s.minio.PresignedGetObject(
        ctx,
        s.config.BucketName,
        doc.FilePath,
        time.Hour, // 1 hour expiry
        nil,
    )

    // 4. Audit log
    s.auditLog.LogDocumentAccess(ctx, doc.ID, doc.UserID)

    return url.String(), nil
}
```

### 1.6 File Validation Rules

| File Type | Max Size | Allowed Extensions | Virus Scan |
|-----------|----------|-------------------|------------|
| Identity Documents | 5 MB | pdf, jpg, jpeg, png | Yes |
| Certificates | 5 MB | pdf, jpg, jpeg, png | Yes |
| Profile Photos | 2 MB | jpg, jpeg, png | Yes |
| Company Documents | 10 MB | pdf, jpg, jpeg, png | Yes |

### 1.7 MinIO Service Interface

```go
package storage

import (
    "context"
    "io"
    "time"
)

type DocumentStorage interface {
    // Upload document to storage
    Upload(ctx context.Context, path string, reader io.Reader, size int64, contentType string) error

    // Generate pre-signed GET URL
    GetPresignedURL(ctx context.Context, path string, expiry time.Duration) (string, error)

    // Generate pre-signed PUT URL (for direct client uploads)
    GetPresignedUploadURL(ctx context.Context, path string, expiry time.Duration) (string, error)

    // Delete document
    Delete(ctx context.Context, path string) error

    // Check if document exists
    Exists(ctx context.Context, path string) (bool, error)

    // Get document metadata
    GetMetadata(ctx context.Context, path string) (*ObjectMetadata, error)
}

type ObjectMetadata struct {
    Size         int64
    ContentType  string
    LastModified time.Time
    ETag         string
}
```

---

## 2. Audit Logging System

### 2.1 Audit Log Tables

#### 2.1.1 `audit_logs` (Global audit table)

```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),

    -- Event Details
    event_type VARCHAR(100) NOT NULL,      -- user.created, profile.updated, document.uploaded
    event_category VARCHAR(50) NOT NULL,   -- user, profile, document, organization, assignment
    event_action VARCHAR(50) NOT NULL,     -- create, update, delete, upload, verify

    -- Actor (who did it)
    actor_user_id BIGINT,                  -- User who performed the action
    actor_type VARCHAR(50) DEFAULT 'user', -- user, system, api, webhook

    -- Target (what was affected)
    target_type VARCHAR(50) NOT NULL,      -- user, employee_profile, document, division
    target_id VARCHAR(255) NOT NULL,       -- ID of the affected resource

    -- Change Details
    changes JSONB,                         -- {before: {...}, after: {...}}
    metadata JSONB DEFAULT '{}',           -- Additional context

    -- Request Details
    ip_address INET,
    user_agent TEXT,
    request_id VARCHAR(100),               -- For tracing

    -- Timestamp
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Indexes
    INDEX idx_audit_tenant_id (tenant_id),
    INDEX idx_audit_event_type (event_type),
    INDEX idx_audit_event_category (event_category),
    INDEX idx_audit_actor_user_id (actor_user_id),
    INDEX idx_audit_target (target_type, target_id),
    INDEX idx_audit_created_at (created_at)
);

-- Partition by month for performance
CREATE TABLE audit_logs_2025_01 PARTITION OF audit_logs
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

CREATE TABLE audit_logs_2025_02 PARTITION OF audit_logs
    FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');
-- ... etc
```

#### 2.1.2 `user_activity_logs` (Login/Authentication events)

```sql
CREATE TABLE user_activity_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL REFERENCES users(id),
    tenant_id UUID NOT NULL REFERENCES tenants(id),

    -- Activity Details
    activity_type VARCHAR(50) NOT NULL,    -- login, logout, password_change, 2fa_enable
    status VARCHAR(20) NOT NULL,           -- success, failure, blocked
    failure_reason TEXT,

    -- Context
    ip_address INET,
    user_agent TEXT,
    device_type VARCHAR(50),               -- web, mobile, api
    location_city VARCHAR(100),
    location_country VARCHAR(100),

    -- Session
    session_id VARCHAR(255),

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_activity_user_id (user_id),
    INDEX idx_activity_type (activity_type),
    INDEX idx_activity_status (status),
    INDEX idx_activity_created_at (created_at)
);
```

### 2.2 Audit Events to Track

#### User Management Events
- `user.created` - New user account created
- `user.updated` - User details updated
- `user.deleted` - User account deleted/deactivated
- `user.password_changed` - Password changed
- `user.password_reset` - Password reset requested/completed
- `user.2fa_enabled` - Two-factor authentication enabled
- `user.2fa_disabled` - Two-factor authentication disabled
- `user.login` - User login (success/failure)
- `user.logout` - User logout

#### Profile Events
- `profile.created` - Profile created (employee/vendor/contractor)
- `profile.updated` - Profile updated
- `profile.photo_uploaded` - Profile photo uploaded

#### Document Events
- `document.uploaded` - Document uploaded
- `document.verified` - Document verified by admin
- `document.rejected` - Document rejected
- `document.deleted` - Document deleted
- `document.accessed` - Document viewed/downloaded

#### Certificate Events
- `certificate.uploaded` - Certificate uploaded
- `certificate.verified` - Certificate verified
- `certificate.expiring_soon` - Certificate expiring in 30 days

#### Assignment Events
- `assignment.created` - User assigned to org unit
- `assignment.updated` - Assignment details updated
- `assignment.deleted` - Assignment removed
- `assignment.role_changed` - User role changed

#### Organization Events
- `division.created` - New division created
- `division.updated` - Division updated
- `division.deleted` - Division deleted
- `branch.created` - New branch created
- `branch.updated` - Branch updated
- `branch.deleted` - Branch deleted
- `department.created` - New department created
- `department.updated` - Department updated
- `department.deleted` - Department deleted

### 2.3 Audit Service Interface

```go
package audit

import (
    "context"
    "encoding/json"
)

type AuditService interface {
    Log(ctx context.Context, event *AuditEvent) error
    LogUserActivity(ctx context.Context, activity *UserActivity) error
    Query(ctx context.Context, filter *AuditFilter) ([]*AuditEvent, error)
}

type AuditEvent struct {
    TenantID      string
    EventType     string
    EventCategory string
    EventAction   string
    ActorUserID   *int64
    ActorType     string
    TargetType    string
    TargetID      string
    Changes       *ChangeLog
    Metadata      map[string]interface{}
    IPAddress     string
    UserAgent     string
    RequestID     string
}

type ChangeLog struct {
    Before map[string]interface{} `json:"before"`
    After  map[string]interface{} `json:"after"`
}

type UserActivity struct {
    UserID        int64
    TenantID      string
    ActivityType  string
    Status        string
    FailureReason string
    IPAddress     string
    UserAgent     string
    DeviceType    string
    SessionID     string
}

type AuditFilter struct {
    TenantID      string
    EventCategory string
    EventType     string
    ActorUserID   *int64
    TargetType    string
    TargetID      string
    StartDate     time.Time
    EndDate       time.Time
    Limit         int
    Offset        int
}
```

### 2.4 Audit Log Usage Example

```go
// Example: Log profile update
func (s *ProfileService) UpdateEmployeeProfile(ctx context.Context, req *UpdateEmployeeProfileRequest) (*EmployeeProfile, error) {
    // Get current profile
    currentProfile, err := s.repo.GetEmployeeProfile(ctx, req.UserID)
    if err != nil {
        return nil, err
    }

    // Update profile
    updatedProfile, err := s.repo.UpdateEmployeeProfile(ctx, req)
    if err != nil {
        return nil, err
    }

    // Create audit log
    s.audit.Log(ctx, &audit.AuditEvent{
        TenantID:      s.getTenantID(ctx),
        EventType:     "profile.updated",
        EventCategory: "profile",
        EventAction:   "update",
        ActorUserID:   s.getActorUserID(ctx),
        ActorType:     "user",
        TargetType:    "employee_profile",
        TargetID:      updatedProfile.ID.String(),
        Changes: &audit.ChangeLog{
            Before: profileToMap(currentProfile),
            After:  profileToMap(updatedProfile),
        },
        IPAddress: s.getIPAddress(ctx),
        UserAgent: s.getUserAgent(ctx),
        RequestID: s.getRequestID(ctx),
    })

    return updatedProfile, nil
}
```

---

## 3. Reporting Requirements

### 3.1 Report Tables (Materialized Views)

#### 3.1.1 `report_headcount_by_division`

```sql
CREATE MATERIALIZED VIEW report_headcount_by_division AS
SELECT
    t.id as tenant_id,
    t.name as tenant_name,
    d.id as division_id,
    d.name as division_name,
    COUNT(DISTINCT ep.user_id) FILTER (WHERE u.user_type = 'employee' AND u.is_active = true) as active_employees,
    COUNT(DISTINCT ep.user_id) FILTER (WHERE u.user_type = 'employee' AND u.is_active = false) as inactive_employees,
    COUNT(DISTINCT vp.user_id) FILTER (WHERE u.user_type = 'vendor' AND u.is_active = true) as active_vendors,
    COUNT(DISTINCT cp.user_id) FILTER (WHERE u.user_type = 'contractor' AND u.is_active = true) as active_contractors,
    COUNT(DISTINCT ep.user_id) FILTER (WHERE u.user_type = 'employee') as total_employees
FROM tenants t
LEFT JOIN divisions d ON t.id = d.tenant_id
LEFT JOIN employee_profiles ep ON d.id = ep.primary_division_id
LEFT JOIN vendor_profiles vp ON d.tenant_id = vp.tenant_id
LEFT JOIN contractor_profiles cp ON d.tenant_id = cp.tenant_id
LEFT JOIN users u ON u.id = ep.user_id OR u.id = vp.user_id OR u.id = cp.user_id
WHERE d.is_active = true
GROUP BY t.id, t.name, d.id, d.name;

CREATE INDEX idx_report_headcount_division ON report_headcount_by_division(tenant_id, division_id);
```

#### 3.1.2 `report_headcount_by_branch`

```sql
CREATE MATERIALIZED VIEW report_headcount_by_branch AS
SELECT
    b.tenant_id,
    b.division_id,
    d.name as division_name,
    b.id as branch_id,
    b.name as branch_name,
    b.city,
    COUNT(DISTINCT ep.user_id) FILTER (WHERE u.is_active = true) as active_employees,
    COUNT(DISTINCT ep.user_id) FILTER (WHERE u.is_active = false) as inactive_employees,
    COUNT(DISTINCT ep.user_id) FILTER (WHERE ep.probation_end_date > CURRENT_DATE) as on_probation,
    COUNT(DISTINCT ep.user_id) FILTER (WHERE ep.employment_type = 'Permanent') as permanent_employees,
    COUNT(DISTINCT ep.user_id) FILTER (WHERE ep.employment_type = 'Contract') as contract_employees
FROM branches b
LEFT JOIN divisions d ON b.division_id = d.id
LEFT JOIN employee_profiles ep ON b.id = ep.primary_branch_id
LEFT JOIN users u ON u.id = ep.user_id
WHERE b.is_active = true
GROUP BY b.tenant_id, b.division_id, d.name, b.id, b.name, b.city;

CREATE INDEX idx_report_headcount_branch ON report_headcount_by_branch(tenant_id, branch_id);
```

#### 3.1.3 `report_headcount_by_department`

```sql
CREATE MATERIALIZED VIEW report_headcount_by_department AS
SELECT
    dept.tenant_id,
    dept.division_id,
    dept.id as department_id,
    dept.name as department_name,
    dept.department_type,
    COUNT(DISTINCT ep.user_id) FILTER (WHERE u.is_active = true) as active_employees,
    COUNT(DISTINCT ep.user_id) as total_employees
FROM departments dept
LEFT JOIN employee_profiles ep ON dept.id = ep.primary_department_id
LEFT JOIN users u ON u.id = ep.user_id
WHERE dept.is_active = true
GROUP BY dept.tenant_id, dept.division_id, dept.id, dept.name, dept.department_type;

CREATE INDEX idx_report_headcount_department ON report_headcount_by_department(tenant_id, department_id);
```

#### 3.1.4 `report_document_expiry_tracker`

```sql
CREATE MATERIALIZED VIEW report_document_expiry_tracker AS
SELECT
    u.tenant_id,
    u.id as user_id,
    u.username,
    u.fullname,
    u.user_type,
    uid.id as document_id,
    uid.document_type,
    uid.document_number,
    uid.expiry_date,
    CASE
        WHEN uid.expiry_date < CURRENT_DATE THEN 'Expired'
        WHEN uid.expiry_date <= CURRENT_DATE + INTERVAL '30 days' THEN 'Expiring Soon'
        WHEN uid.expiry_date <= CURRENT_DATE + INTERVAL '90 days' THEN 'Expiring in 3 Months'
        ELSE 'Valid'
    END as expiry_status,
    uid.is_verified
FROM users u
INNER JOIN user_identity_documents uid ON u.id = uid.user_id
WHERE uid.expiry_date IS NOT NULL
UNION ALL
SELECT
    u.tenant_id,
    u.id as user_id,
    u.username,
    u.fullname,
    u.user_type,
    uc.id as document_id,
    uc.certificate_type as document_type,
    uc.certificate_number as document_number,
    uc.expiry_date,
    CASE
        WHEN uc.expiry_date < CURRENT_DATE THEN 'Expired'
        WHEN uc.expiry_date <= CURRENT_DATE + INTERVAL '30 days' THEN 'Expiring Soon'
        WHEN uc.expiry_date <= CURRENT_DATE + INTERVAL '90 days' THEN 'Expiring in 3 Months'
        ELSE 'Valid'
    END as expiry_status,
    uc.is_verified
FROM users u
INNER JOIN user_certificates uc ON u.id = uc.user_id
WHERE uc.expiry_date IS NOT NULL;

CREATE INDEX idx_report_doc_expiry ON report_document_expiry_tracker(tenant_id, expiry_status, expiry_date);
```

#### 3.1.5 `report_new_joiners_leavers`

```sql
CREATE MATERIALIZED VIEW report_new_joiners_leavers AS
SELECT
    t.id as tenant_id,
    d.id as division_id,
    d.name as division_name,
    b.id as branch_id,
    b.name as branch_name,
    DATE_TRUNC('month', ep.date_of_joining) as month,
    COUNT(DISTINCT ep.user_id) FILTER (WHERE ep.date_of_joining IS NOT NULL) as new_joiners,
    COUNT(DISTINCT ep.user_id) FILTER (WHERE ep.date_of_leaving IS NOT NULL) as leavers,
    COUNT(DISTINCT ep.user_id) FILTER (WHERE ep.date_of_joining IS NOT NULL) -
    COUNT(DISTINCT ep.user_id) FILTER (WHERE ep.date_of_leaving IS NOT NULL) as net_change
FROM tenants t
LEFT JOIN divisions d ON t.id = d.tenant_id
LEFT JOIN branches b ON d.id = b.division_id
LEFT JOIN employee_profiles ep ON b.id = ep.primary_branch_id
WHERE ep.date_of_joining >= DATE_TRUNC('month', CURRENT_DATE - INTERVAL '12 months')
   OR ep.date_of_leaving >= DATE_TRUNC('month', CURRENT_DATE - INTERVAL '12 months')
GROUP BY t.id, d.id, d.name, b.id, b.name, DATE_TRUNC('month', ep.date_of_joining)
ORDER BY month DESC;

CREATE INDEX idx_report_joiners_leavers ON report_new_joiners_leavers(tenant_id, month);
```

#### 3.1.6 `report_employee_by_designation`

```sql
CREATE MATERIALIZED VIEW report_employee_by_designation AS
SELECT
    ep.designation,
    ep.grade,
    COUNT(DISTINCT ep.user_id) FILTER (WHERE u.is_active = true) as active_count,
    COUNT(DISTINCT ep.user_id) as total_count,
    MIN(ep.date_of_joining) as earliest_joining,
    MAX(ep.date_of_joining) as latest_joining
FROM employee_profiles ep
LEFT JOIN users u ON u.id = ep.user_id
GROUP BY ep.designation, ep.grade
ORDER BY total_count DESC;
```

### 3.2 Report Refresh Strategy

```sql
-- Refresh reports daily at 2 AM
CREATE EXTENSION IF NOT EXISTS pg_cron;

SELECT cron.schedule(
    'refresh-headcount-reports',
    '0 2 * * *',  -- Every day at 2 AM
    $$
    REFRESH MATERIALIZED VIEW CONCURRENTLY report_headcount_by_division;
    REFRESH MATERIALIZED VIEW CONCURRENTLY report_headcount_by_branch;
    REFRESH MATERIALIZED VIEW CONCURRENTLY report_headcount_by_department;
    REFRESH MATERIALIZED VIEW CONCURRENTLY report_employee_by_designation;
    $$
);

SELECT cron.schedule(
    'refresh-expiry-reports',
    '0 6 * * *',  -- Every day at 6 AM
    $$
    REFRESH MATERIALIZED VIEW CONCURRENTLY report_document_expiry_tracker;
    $$
);

SELECT cron.schedule(
    'refresh-joiners-leavers-reports',
    '0 3 1 * *',  -- 1st of every month at 3 AM
    $$
    REFRESH MATERIALIZED VIEW CONCURRENTLY report_new_joiners_leavers;
    $$
);
```

### 3.3 Report Service APIs

```protobuf
service ReportService {
    // Headcount Reports
    rpc GetHeadcountByDivision(HeadcountRequest) returns (HeadcountByDivisionResponse) {}
    rpc GetHeadcountByBranch(HeadcountRequest) returns (HeadcountByBranchResponse) {}
    rpc GetHeadcountByDepartment(HeadcountRequest) returns (HeadcountByDepartmentResponse) {}

    // User Distribution
    rpc GetUserTypeDistribution(UserDistributionRequest) returns (UserDistributionResponse) {}
    rpc GetEmployeeByDesignation(EmployeeDesignationRequest) returns (EmployeeDesignationResponse) {}

    // Document Expiry
    rpc GetDocumentExpiryReport(DocumentExpiryRequest) returns (DocumentExpiryResponse) {}
    rpc GetExpiringDocuments(ExpiringDocumentsRequest) returns (ExpiringDocumentsResponse) {}

    // Joiners & Leavers
    rpc GetNewJoinersReport(JoinersLeaversRequest) returns (JoinersLeaversResponse) {}
    rpc GetLeaversReport(JoinersLeaversRequest) returns (JoinersLeaversResponse) {}
    rpc GetJoinersLeaversTrend(JoinersLeaversRequest) returns (JoinersLeaversTrendResponse) {}

    // Reporting Hierarchy
    rpc GetReportingHierarchy(ReportingHierarchyRequest) returns (ReportingHierarchyResponse) {}
    rpc GetTeamMembers(TeamMembersRequest) returns (TeamMembersResponse) {}

    // Export
    rpc ExportReport(ExportReportRequest) returns (ExportReportResponse) {}  // CSV/PDF/Excel
}
```

---

## 4. Background Jobs & Scheduled Tasks

### 4.1 Document Expiry Notifications

```go
// Run daily at 9 AM
func (j *DocumentExpiryJob) Run(ctx context.Context) error {
    // Get documents expiring in 30 days
    docs, err := j.repo.GetExpiringDocuments(ctx, 30)
    if err != nil {
        return err
    }

    for _, doc := range docs {
        // Send notification to user
        j.notification.Send(ctx, &notification.Message{
            UserID:   doc.UserID,
            Template: "document_expiring_soon",
            Data: map[string]interface{}{
                "document_type": doc.DocumentType,
                "expiry_date":   doc.ExpiryDate,
                "days_left":     doc.DaysUntilExpiry,
            },
        })

        // Notify HR/Admin
        j.notification.SendToRole(ctx, &notification.Message{
            TenantID: doc.TenantID,
            Role:     "HR_Admin",
            Template: "user_document_expiring",
            Data: map[string]interface{}{
                "user_name":     doc.UserName,
                "document_type": doc.DocumentType,
                "expiry_date":   doc.ExpiryDate,
            },
        })
    }

    return nil
}
```

### 4.2 Contract Expiry Notifications

```go
// Run daily to check contractor contract expiry
func (j *ContractExpiryJob) Run(ctx context.Context) error {
    // Get contracts expiring in 15 days
    contracts, err := j.repo.GetExpiringContracts(ctx, 15)
    if err != nil {
        return err
    }

    for _, contract := range contracts {
        // Notify contractor
        j.notification.Send(ctx, &notification.Message{
            UserID:   contract.UserID,
            Template: "contract_expiring_soon",
            Data: map[string]interface{}{
                "contract_end_date": contract.ContractEndDate,
                "days_left":         contract.DaysUntilExpiry,
            },
        })

        // Notify Admin
        j.notification.SendToRole(ctx, &notification.Message{
            TenantID: contract.TenantID,
            Role:     "Admin",
            Template: "contractor_contract_expiring",
            Data: map[string]interface{}{
                "contractor_name": contract.UserName,
                "contract_end_date": contract.ContractEndDate,
            },
        })
    }

    return nil
}
```

### 4.3 Probation End Notifications

```go
// Run daily to check probation ending soon
func (j *ProbationEndJob) Run(ctx context.Context) error {
    // Get employees whose probation ends in 7 days
    employees, err := j.repo.GetProbationEnding(ctx, 7)
    if err != nil {
        return err
    }

    for _, emp := range employees {
        // Notify reporting manager
        j.notification.Send(ctx, &notification.Message{
            UserID:   emp.ReportingManagerUserID,
            Template: "probation_ending_soon",
            Data: map[string]interface{}{
                "employee_name":     emp.FullName,
                "probation_end_date": emp.ProbationEndDate,
            },
        })

        // Notify HR
        j.notification.SendToRole(ctx, &notification.Message{
            TenantID: emp.TenantID,
            Role:     "HR_Manager",
            Template: "employee_probation_ending",
            Data: map[string]interface{}{
                "employee_name":     emp.FullName,
                "employee_code":     emp.EmployeeCode,
                "probation_end_date": emp.ProbationEndDate,
            },
        })
    }

    return nil
}
```

---

## 5. Performance Optimization

### 5.1 Database Indexing Strategy

```sql
-- Composite indexes for common queries

-- User lookups with filters
CREATE INDEX idx_users_tenant_type_active ON users(tenant_id, user_type, is_active);
CREATE INDEX idx_users_tenant_status ON users(tenant_id, status) WHERE is_active = true;

-- Assignment lookups
CREATE INDEX idx_org_assignments_user_active ON user_organizational_assignments(user_id, is_active, valid_from, valid_until);
CREATE INDEX idx_org_assignments_branch_active ON user_organizational_assignments(branch_id, is_active) WHERE branch_id IS NOT NULL;

-- Profile searches
CREATE INDEX idx_employee_profiles_branch_dept ON employee_profiles(primary_branch_id, primary_department_id) WHERE is_active = true;
CREATE INDEX idx_employee_profiles_manager ON employee_profiles(reporting_manager_user_id) WHERE is_active = true;

-- Document queries
CREATE INDEX idx_identity_docs_expiry ON user_identity_documents(expiry_date) WHERE expiry_date IS NOT NULL;
CREATE INDEX idx_certificates_expiry ON user_certificates(expiry_date) WHERE expiry_date IS NOT NULL;

-- Audit log queries
CREATE INDEX idx_audit_logs_composite ON audit_logs(tenant_id, event_category, created_at DESC);
CREATE INDEX idx_audit_logs_target ON audit_logs(target_type, target_id, created_at DESC);
```

### 5.2 Query Optimization

```sql
-- Use prepared statements
PREPARE get_branch_employees AS
SELECT u.*, ep.*
FROM users u
INNER JOIN employee_profiles ep ON u.id = ep.user_id
WHERE ep.primary_branch_id = $1 AND u.is_active = true;

-- Use CTEs for complex queries
WITH branch_users AS (
    SELECT user_id FROM employee_profiles WHERE primary_branch_id = $1
),
active_users AS (
    SELECT id FROM users WHERE id IN (SELECT user_id FROM branch_users) AND is_active = true
)
SELECT * FROM active_users;
```

### 5.3 Caching Strategy

```go
// Cache frequently accessed data
type CacheKeys struct {
    UserProfile         = "user:profile:{user_id}"
    EmployeeProfile     = "employee:profile:{user_id}"
    UserAssignments     = "user:assignments:{user_id}"
    BranchEmployees     = "branch:employees:{branch_id}"
    DivisionBranches    = "division:branches:{division_id}"
}

// Cache TTLs
const (
    ProfileCacheTTL     = 1 * time.Hour
    AssignmentCacheTTL  = 30 * time.Minute
    ReportCacheTTL      = 15 * time.Minute
)
```

---

## 6. Security & Compliance

### 6.1 Data Encryption

```go
// Encrypt sensitive fields before storing
type EncryptionService interface {
    Encrypt(plaintext string) (string, error)
    Decrypt(ciphertext string) (string, error)
}

// Fields to encrypt:
// - user_identity_documents.document_number (Aadhar, PAN, Passport)
// - vendor_profiles.gst_number
// - vendor_profiles.pan_number
// - user_profiles.emergency_contact_phone
```

### 6.2 Access Control Matrix

| Role | View Own Profile | View Team Profiles | Edit Profiles | Verify Documents | View Reports |
|------|-----------------|-------------------|---------------|------------------|--------------|
| Employee | ✓ | ✗ | Own only | ✗ | Own only |
| Manager | ✓ | Team only | Team only | ✗ | Team only |
| HR Manager | ✓ | Department | Department | ✓ | Department |
| HR GM | ✓ | All | All | ✓ | All |
| Admin | ✓ | All | All | ✓ | All |

### 6.3 GDPR Compliance

- **Right to Access**: Users can request all their data
- **Right to Erasure**: Soft delete with anonymization
- **Data Portability**: Export user data in JSON/CSV
- **Consent Management**: Track consent for data processing

```sql
-- Anonymize user data (GDPR right to be forgotten)
CREATE FUNCTION anonymize_user_data(p_user_id BIGINT) RETURNS void AS $$
BEGIN
    UPDATE users SET
        username = 'deleted_user_' || id,
        email = 'deleted_' || id || '@example.com',
        phone = NULL,
        fullname = 'Deleted User'
    WHERE id = p_user_id;

    UPDATE user_profiles SET
        date_of_birth = NULL,
        current_address_line1 = NULL,
        permanent_address_line1 = NULL,
        emergency_contact_name = NULL,
        emergency_contact_phone = NULL
    WHERE user_id = p_user_id;

    -- Mark documents for deletion
    UPDATE user_identity_documents SET
        document_number = 'REDACTED'
    WHERE user_id = p_user_id;
END;
$$ LANGUAGE plpgsql;
```

---

## Next Steps for Implementation

1. ✅ Setup MinIO configuration in the system
2. ✅ Create audit logging tables and service
3. ✅ Create report materialized views
4. ✅ Implement background jobs for notifications
5. ✅ Add encryption for sensitive data
6. ✅ Implement caching layer
7. ✅ Setup monitoring and alerting
