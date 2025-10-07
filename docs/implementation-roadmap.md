# Implementation Roadmap - Organizational Hierarchy & User Management

## Overview
This roadmap outlines the step-by-step implementation plan for the organizational hierarchy and user management system.

**Estimated Timeline**: 8-10 weeks
**Team Size**: 2-3 developers

---

## Phase 1: Foundation Setup (Week 1-2)

### Week 1: Database Schema & MinIO Setup

#### Day 1-2: Organizational Hierarchy Tables
- [ ] Create `divisions` table with indexes
- [ ] Create `branches` table with location support
- [ ] Create `departments` table with hierarchy support
- [ ] Add database migrations
- [ ] Write SQLC queries for CRUD operations

**Files to create:**
```
identity/tenant/db/schema/
├── 002_organizational_hierarchy.sql
└── queries/
    ├── divisions.sql
    ├── branches.sql
    └── departments.sql
```

#### Day 3-4: User Profile Tables
- [ ] Update `users` table (add user_type, tenant_id)
- [ ] Create `user_profiles` table
- [ ] Create `employee_profiles` table
- [ ] Create `vendor_profiles` table
- [ ] Create `contractor_profiles` table
- [ ] Create `employee_family_details` table
- [ ] Add indexes and constraints

**Files to create:**
```
identity/user/db/sqlc/schema/
├── 002_user_types.sql
├── 003_user_profiles.sql
├── 004_employee_profiles.sql
├── 005_vendor_profiles.sql
└── 006_contractor_profiles.sql
```

#### Day 5: Document Management Tables
- [ ] Create `user_identity_documents` table
- [ ] Create `user_certificates` table
- [ ] Add file path and metadata columns
- [ ] Create SQLC queries

**Files to create:**
```
identity/user/db/sqlc/schema/
└── 007_documents.sql

identity/user/db/sqlc/queries/
├── documents.sql
└── certificates.sql
```

### Week 2: MinIO Integration & Audit Logging

#### Day 1-2: MinIO Setup
- [ ] Create MinIO configuration package
- [ ] Implement DocumentStorage interface
- [ ] Add file upload/download handlers
- [ ] Implement pre-signed URL generation
- [ ] Add file validation (size, type, virus scan)
- [ ] Write unit tests

**Files to create:**
```
packages/storage/
├── config.go
├── minio.go
├── interface.go
├── validation.go
└── minio_test.go
```

#### Day 3-4: Audit Logging System
- [ ] Create `audit_logs` table with partitioning
- [ ] Create `user_activity_logs` table
- [ ] Implement AuditService interface
- [ ] Add audit middleware for HTTP handlers
- [ ] Create audit log queries
- [ ] Write unit tests

**Files to create:**
```
packages/audit/
├── service.go
├── models.go
├── middleware.go
└── repository.go

packages/audit/db/schema/
└── audit_logs.sql
```

#### Day 5: Assignment & Access Control Tables
- [ ] Create `user_organizational_assignments` table
- [ ] Create `user_project_assignments` table
- [ ] Update `user_tenant_roles` with hierarchy fields
- [ ] Add SQLC queries for assignments
- [ ] Write tests

**Files to create:**
```
identity/user/db/sqlc/schema/
└── 008_assignments.sql

identity/user/db/sqlc/queries/
└── assignments.sql
```

---

## Phase 2: Proto Definitions & Code Generation (Week 3)

### Day 1-2: Organization Service Proto

**File:** `identity/tenant/proto/organization.proto`

- [ ] Define Division message
- [ ] Define Branch message with Address and Location
- [ ] Define Department message
- [ ] Define OrganizationService with RPCs:
  - Division CRUD
  - Branch CRUD
  - Department CRUD
  - Hierarchy queries
- [ ] Generate Go code with protoc
- [ ] Generate Connect handlers

**Commands:**
```bash
make proto-gen-tenant
```

### Day 3-4: Profile Service Proto

**File:** `profile/proto/profile.proto`

- [ ] Define UserProfile message
- [ ] Define EmployeeProfile message
- [ ] Define VendorProfile message
- [ ] Define ContractorProfile message
- [ ] Define IdentityDocument message
- [ ] Define Certificate message
- [ ] Define ProfileService with RPCs
- [ ] Generate Go code

**Commands:**
```bash
make proto-gen-profile
```

### Day 5: Update User Service Proto

**File:** `identity/user/proto/user.proto`

- [ ] Add UserType enum
- [ ] Update User message with user_type and tenant_id
- [ ] Update UserTenantRole with hierarchy fields
- [ ] Add AccessScope enum
- [ ] Generate updated Go code

**Commands:**
```bash
make proto-gen-user
```

---

## Phase 3: Service Implementation - Organization (Week 4)

### Day 1-2: Division Service

**Files to create:**
```
identity/tenant/services/
├── division_service.go
└── division_service_test.go

identity/tenant/handlers/
├── division_handlers.go
└── division_handlers_test.go
```

**Implementation checklist:**
- [ ] CreateDivision handler with validation
- [ ] UpdateDivision handler with audit logging
- [ ] DeleteDivision handler (soft delete)
- [ ] GetDivision handler
- [ ] ListDivisions with pagination and filters
- [ ] Unit tests for all handlers
- [ ] Integration tests

### Day 3-4: Branch Service

**Files to create:**
```
identity/tenant/services/
├── branch_service.go
└── branch_service_test.go

identity/tenant/handlers/
├── branch_handlers.go
└── branch_handlers_test.go
```

**Implementation checklist:**
- [ ] CreateBranch with location validation
- [ ] UpdateBranch with audit logging
- [ ] DeleteBranch (soft delete)
- [ ] GetBranch
- [ ] ListBranches with filters
- [ ] ListBranchesByDivision
- [ ] Unit and integration tests

### Day 5: Department Service

**Files to create:**
```
identity/tenant/services/
├── department_service.go
└── department_service_test.go

identity/tenant/handlers/
├── department_handlers.go
└── department_handlers_test.go
```

**Implementation checklist:**
- [ ] CreateDepartment with scope validation
- [ ] UpdateDepartment
- [ ] DeleteDepartment
- [ ] GetDepartment
- [ ] ListDepartments
- [ ] ListDepartmentsByDivision
- [ ] Handle business-level vs division-level departments
- [ ] Tests

---

## Phase 4: Service Implementation - User Profiles (Week 5)

### Day 1: Common User Profile Service

**Files to create:**
```
profile/services/
├── user_profile_service.go
└── user_profile_service_test.go

profile/handlers/
├── user_profile_handlers.go
└── user_profile_handlers_test.go

profile/repository/
└── user_profile_repository.go
```

**Implementation checklist:**
- [ ] CreateUserProfile
- [ ] UpdateUserProfile with audit logging
- [ ] GetUserProfile
- [ ] Validate address fields
- [ ] Emergency contact validation
- [ ] Tests

### Day 2: Employee Profile Service

**Files to create:**
```
profile/services/
├── employee_profile_service.go
└── employee_profile_service_test.go

profile/handlers/
├── employee_handlers.go
└── employee_handlers_test.go
```

**Implementation checklist:**
- [ ] CreateEmployeeProfile
- [ ] UpdateEmployeeProfile
- [ ] GetEmployeeProfile
- [ ] Employee code generation logic
- [ ] Validate branch/department/division references
- [ ] Handle probation period
- [ ] Tests

### Day 3: Vendor & Contractor Profile Services

**Files to create:**
```
profile/services/
├── vendor_profile_service.go
├── vendor_profile_service_test.go
├── contractor_profile_service.go
└── contractor_profile_service_test.go

profile/handlers/
├── vendor_handlers.go
└── contractor_handlers.go
```

**Implementation checklist:**
- [ ] CreateVendorProfile with company details
- [ ] GST/PAN validation
- [ ] CreateContractorProfile
- [ ] Contract date validation
- [ ] Auto-deactivation on contract expiry
- [ ] Tests

### Day 4-5: Document Management Service

**Files to create:**
```
profile/services/
├── document_service.go
└── document_service_test.go

profile/handlers/
├── document_handlers.go
└── document_handlers_test.go
```

**Implementation checklist:**
- [ ] UploadIdentityDocument
  - File validation
  - MinIO upload
  - DB record creation
  - Audit logging
- [ ] UploadCertificate
- [ ] VerifyDocument (admin only)
- [ ] GetDocumentURL (pre-signed URL)
- [ ] ListUserDocuments
- [ ] DeleteDocument
- [ ] Tests with mock MinIO

---

## Phase 5: Assignment & Access Control (Week 6)

### Day 1-2: Assignment Service

**Files to create:**
```
identity/user/services/
├── assignment_service.go
└── assignment_service_test.go

identity/user/handlers/
├── assignment_handlers.go
└── assignment_handlers_test.go
```

**Implementation checklist:**
- [ ] AssignUserToOrganization
  - Validate hierarchy (business/division/branch/department)
  - Check for conflicts
  - Priority handling
- [ ] UpdateAssignment
- [ ] RemoveAssignment
- [ ] GetUserAssignments
- [ ] GetAssignmentsByOrganizationalUnit
- [ ] Tests

### Day 3-4: Permission Resolution Service

**Files to create:**
```
identity/user/services/
├── permission_resolver.go
└── permission_resolver_test.go
```

**Implementation checklist:**
- [ ] ResolveUserPermissions
  - Get all assignments ordered by priority
  - Check explicit denies
  - Check explicit grants
  - Check role-based permissions with cascade
  - Cache permission decisions
- [ ] CanAccessResource
- [ ] GetUserAccessScope
- [ ] Tests with various scenarios

### Day 5: Assignment Validation & Business Rules

**Implementation checklist:**
- [ ] Prevent circular reporting (manager cannot report to their team member)
- [ ] Validate assignment dates (valid_from <= valid_until)
- [ ] Check user type compatibility (vendors can't be employees)
- [ ] Ensure at least one primary assignment per employee
- [ ] Tests for all validation rules

---

## Phase 6: Reporting & Analytics (Week 7)

### Day 1-2: Materialized Views & Refresh Jobs

**Files to create:**
```
packages/reporting/db/schema/
├── materialized_views.sql
└── refresh_jobs.sql

packages/reporting/
└── refresh_service.go
```

**Implementation checklist:**
- [ ] Create all materialized views
  - report_headcount_by_division
  - report_headcount_by_branch
  - report_headcount_by_department
  - report_document_expiry_tracker
  - report_new_joiners_leavers
  - report_employee_by_designation
- [ ] Setup pg_cron for auto-refresh
- [ ] Manual refresh API endpoint
- [ ] Tests

### Day 3-4: Report Service Implementation

**Files to create:**
```
packages/reporting/proto/
└── reporting.proto

packages/reporting/services/
├── report_service.go
└── report_service_test.go

packages/reporting/handlers/
├── report_handlers.go
└── report_handlers_test.go
```

**Implementation checklist:**
- [ ] GetHeadcountByDivision
- [ ] GetHeadcountByBranch
- [ ] GetHeadcountByDepartment
- [ ] GetUserTypeDistribution
- [ ] GetEmployeeByDesignation
- [ ] GetDocumentExpiryReport
- [ ] GetExpiringDocuments
- [ ] GetNewJoinersReport
- [ ] GetLeaversReport
- [ ] GetJoinersLeaversTrend
- [ ] Tests

### Day 5: Report Export Service

**Files to create:**
```
packages/reporting/services/
├── export_service.go
└── export_service_test.go
```

**Implementation checklist:**
- [ ] ExportToCSV
- [ ] ExportToPDF
- [ ] ExportToExcel
- [ ] Generate pre-signed URL for download
- [ ] Auto-cleanup old exports
- [ ] Tests

---

## Phase 7: Background Jobs & Notifications (Week 8)

### Day 1-2: Scheduled Jobs

**Files to create:**
```
packages/jobs/
├── document_expiry_job.go
├── contract_expiry_job.go
├── probation_end_job.go
└── scheduler.go
```

**Implementation checklist:**
- [ ] Document expiry notification job (daily at 9 AM)
- [ ] Contract expiry notification job (daily at 9 AM)
- [ ] Probation end notification job (daily at 9 AM)
- [ ] Inactive user cleanup job (weekly)
- [ ] Setup job scheduler (cron/worker)
- [ ] Tests with time mocking

### Day 3-4: Notification Integration

**Files to create:**
```
packages/notification/
├── service.go
├── templates.go
└── service_test.go
```

**Implementation checklist:**
- [ ] Email notification templates
  - Document expiring soon
  - Contract expiring
  - Probation ending
  - User created
  - Password reset
- [ ] SMS notification templates
- [ ] Send notification to user
- [ ] Send notification to role (HR/Admin)
- [ ] Tests

### Day 5: Event-Driven Architecture Setup

**Files to create:**
```
packages/events/
├── publisher.go
├── subscriber.go
└── events.go
```

**Implementation checklist:**
- [ ] Define domain events
  - UserCreated
  - ProfileUpdated
  - DocumentUploaded
  - DocumentVerified
  - AssignmentChanged
- [ ] Event publisher
- [ ] Event subscribers for notifications
- [ ] Tests

---

## Phase 8: Integration, Testing & Documentation (Week 9-10)

### Week 9: Integration & Testing

#### Day 1-2: End-to-End Integration Tests

**Files to create:**
```
tests/integration/
├── organization_test.go
├── profile_test.go
├── assignment_test.go
├── document_test.go
└── reporting_test.go
```

**Test scenarios:**
- [ ] Create tenant → Create divisions → Create branches
- [ ] Create employee → Assign to branch → Verify access
- [ ] Upload document → Verify → Generate URL → Download
- [ ] Create vendor → Assign to project → Check permissions
- [ ] Generate reports → Export to CSV/PDF
- [ ] Full user journey tests

#### Day 3-4: Performance Testing

**Implementation checklist:**
- [ ] Load test with 10,000 users
- [ ] Concurrent assignment tests
- [ ] Report generation performance
- [ ] MinIO upload/download performance
- [ ] Database query optimization
- [ ] Add missing indexes based on slow query log

#### Day 5: Security Testing

**Implementation checklist:**
- [ ] Test access control matrix
- [ ] Test permission cascade
- [ ] Test data encryption
- [ ] Test audit logging
- [ ] Test SQL injection prevention
- [ ] Test file upload security (malicious files)

### Week 10: Documentation & Deployment

#### Day 1-2: API Documentation

**Files to create:**
```
docs/api/
├── organization-service.md
├── profile-service.md
├── user-service.md
├── reporting-service.md
└── postman-collection.json
```

**Documentation checklist:**
- [ ] API reference for all services
- [ ] Request/response examples
- [ ] Error codes and handling
- [ ] Authentication/authorization guide
- [ ] Postman collection

#### Day 3: Developer Guide

**Files to create:**
```
docs/
├── developer-guide.md
├── deployment-guide.md
└── troubleshooting.md
```

**Documentation checklist:**
- [ ] Setup instructions
- [ ] Database migration guide
- [ ] MinIO configuration
- [ ] Environment variables
- [ ] Debugging tips
- [ ] Common issues and solutions

#### Day 4: Deployment Preparation

**Implementation checklist:**
- [ ] Create database migration scripts
- [ ] Setup MinIO buckets and policies
- [ ] Configure environment variables
- [ ] Setup monitoring (Prometheus/Grafana)
- [ ] Setup logging (ELK stack)
- [ ] Setup alerts for critical events
- [ ] Create deployment scripts

#### Day 5: Production Deployment

**Deployment checklist:**
- [ ] Run database migrations
- [ ] Deploy MinIO
- [ ] Deploy services
- [ ] Verify health checks
- [ ] Smoke test critical flows
- [ ] Monitor for errors
- [ ] Handover to operations team

---

## File Structure Overview

```
backend/v2/
├── identity/
│   ├── tenant/
│   │   ├── proto/
│   │   │   ├── tenant.proto (existing)
│   │   │   └── organization.proto (new)
│   │   ├── db/
│   │   │   └── schema/
│   │   │       ├── schema.sql (existing)
│   │   │       └── 002_organizational_hierarchy.sql (new)
│   │   ├── services/
│   │   │   ├── division_service.go (new)
│   │   │   ├── branch_service.go (new)
│   │   │   └── department_service.go (new)
│   │   └── handlers/
│   │       ├── division_handlers.go (new)
│   │       ├── branch_handlers.go (new)
│   │       └── department_handlers.go (new)
│   └── user/
│       ├── proto/
│       │   └── user.proto (update)
│       ├── db/sqlc/
│       │   └── schema/
│       │       ├── schema.sql (update)
│       │       ├── 002_user_types.sql (new)
│       │       ├── 003_user_profiles.sql (new)
│       │       ├── 007_documents.sql (new)
│       │       └── 008_assignments.sql (new)
│       └── services/
│           ├── assignment_service.go (new)
│           └── permission_resolver.go (new)
├── profile/ (NEW MODULE)
│   ├── proto/
│   │   └── profile.proto
│   ├── db/
│   │   └── schema/
│   │       ├── employee_profiles.sql
│   │       ├── vendor_profiles.sql
│   │       └── contractor_profiles.sql
│   ├── services/
│   │   ├── user_profile_service.go
│   │   ├── employee_profile_service.go
│   │   ├── vendor_profile_service.go
│   │   ├── contractor_profile_service.go
│   │   └── document_service.go
│   ├── handlers/
│   │   ├── profile_handlers.go
│   │   ├── employee_handlers.go
│   │   ├── vendor_handlers.go
│   │   └── document_handlers.go
│   └── repository/
│       ├── profile_repository.go
│       └── document_repository.go
├── packages/
│   ├── storage/ (NEW)
│   │   ├── config.go
│   │   ├── minio.go
│   │   ├── interface.go
│   │   └── validation.go
│   ├── audit/ (NEW)
│   │   ├── service.go
│   │   ├── models.go
│   │   ├── middleware.go
│   │   └── repository.go
│   ├── reporting/ (NEW)
│   │   ├── proto/
│   │   │   └── reporting.proto
│   │   ├── db/schema/
│   │   │   └── materialized_views.sql
│   │   ├── services/
│   │   │   ├── report_service.go
│   │   │   ├── export_service.go
│   │   │   └── refresh_service.go
│   │   └── handlers/
│   │       └── report_handlers.go
│   ├── jobs/ (NEW)
│   │   ├── document_expiry_job.go
│   │   ├── contract_expiry_job.go
│   │   ├── probation_end_job.go
│   │   └── scheduler.go
│   ├── notification/ (NEW)
│   │   ├── service.go
│   │   ├── templates.go
│   │   └── email.go
│   └── events/ (NEW)
│       ├── publisher.go
│       ├── subscriber.go
│       └── events.go
├── docs/
│   ├── organizational-hierarchy-plan.md (created)
│   ├── technical-specifications.md (created)
│   ├── implementation-roadmap.md (this file)
│   ├── api/
│   │   ├── organization-service.md
│   │   ├── profile-service.md
│   │   └── reporting-service.md
│   ├── developer-guide.md
│   └── deployment-guide.md
└── tests/
    └── integration/
        ├── organization_test.go
        ├── profile_test.go
        ├── assignment_test.go
        └── reporting_test.go
```

---

## Makefile Targets

Add these to your `Makefile`:

```makefile
# Proto generation
proto-gen-tenant:
	protoc --go_out=. --go-grpc_out=. --connect-go_out=. identity/tenant/proto/*.proto

proto-gen-user:
	protoc --go_out=. --go-grpc_out=. --connect-go_out=. identity/user/proto/*.proto

proto-gen-profile:
	protoc --go_out=. --go-grpc_out=. --connect-go_out=. profile/proto/*.proto

proto-gen-reporting:
	protoc --go_out=. --go-grpc_out=. --connect-go_out=. packages/reporting/proto/*.proto

proto-gen-all: proto-gen-tenant proto-gen-user proto-gen-profile proto-gen-reporting

# SQLC generation
sqlc-gen-tenant:
	cd identity/tenant && sqlc generate

sqlc-gen-user:
	cd identity/user && sqlc generate

sqlc-gen-profile:
	cd profile && sqlc generate

sqlc-gen-all: sqlc-gen-tenant sqlc-gen-user sqlc-gen-profile

# Database migrations
migrate-up:
	migrate -path ./migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path ./migrations -database "$(DB_URL)" down 1

migrate-create:
	migrate create -ext sql -dir ./migrations -seq $(name)

# Tests
test-unit:
	go test -v -short ./...

test-integration:
	go test -v -tags=integration ./tests/integration/...

test-all:
	go test -v ./...

# MinIO setup
minio-setup:
	mc alias set local http://localhost:9000 minioadmin minioadmin
	mc mb local/ugcl-documents
	mc policy set download local/ugcl-documents

# Development
dev-setup: minio-setup migrate-up sqlc-gen-all proto-gen-all

# Build
build:
	go build -o bin/server ./cmd/server

# Run
run:
	go run ./cmd/server
```

---

## Risk Mitigation

### High-Risk Areas

1. **Data Migration**
   - Risk: Existing user data migration failure
   - Mitigation:
     - Backup before migration
     - Test on staging first
     - Write rollback scripts

2. **Permission System**
   - Risk: Incorrect permission resolution causing security issues
   - Mitigation:
     - Extensive unit tests
     - Security review
     - Default deny policy

3. **MinIO Integration**
   - Risk: File upload failures, data loss
   - Mitigation:
     - Retry logic
     - Transaction consistency
     - Regular backups

4. **Performance**
   - Risk: Slow queries with large datasets
   - Mitigation:
     - Load testing
     - Query optimization
     - Proper indexing
     - Caching

---

## Success Criteria

- [ ] All database tables created and migrated
- [ ] All services implemented with >80% code coverage
- [ ] All API endpoints functional and documented
- [ ] MinIO integration working with file upload/download
- [ ] Audit logging capturing all critical events
- [ ] Reports generating correctly with accurate data
- [ ] Background jobs running on schedule
- [ ] Performance benchmarks met (< 200ms for 95th percentile)
- [ ] Security tests passed
- [ ] Production deployment successful
- [ ] Team trained on new system

---

## Post-Launch Activities

### Week 11-12: Monitoring & Optimization

- [ ] Monitor error rates and performance
- [ ] Gather user feedback
- [ ] Optimize slow queries
- [ ] Add missing features based on feedback
- [ ] Documentation updates

### Ongoing

- [ ] Regular security audits
- [ ] Database maintenance (vacuum, reindex)
- [ ] MinIO cleanup of old files
- [ ] Audit log archival (older than 1 year)
- [ ] Report view refresh monitoring
- [ ] User training sessions
