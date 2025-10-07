# Planning Session Summary - October 4, 2025

## Session Overview
Comprehensive planning session covering:
1. Organization module implementation
2. Document Management System (DMS) architecture
3. User permissions analysis
4. Documentation strategy

---

## 1. ✅ Organization Module (COMPLETED)

### Created Files
```
organization/
├── db/
│   ├── schema/schema.sql              ✅ Divisions, Branches, Departments
│   ├── queries/
│   │   ├── divisions.sql              ✅ CRUD operations
│   │   ├── branches.sql               ✅ CRUD + location queries
│   │   └── departments.sql            ✅ CRUD + hierarchy queries
│   ├── generated/                     ✅ SQLC models (auto-generated)
│   └── sqlc.yaml                      ✅ SQLC configuration
├── proto/organization.proto           ✅ gRPC service definitions
├── api/v1/organization/              ✅ Generated proto code
├── go.mod                             ✅ Module dependencies
└── [Pending]
    ├── repository/                    ⏳ Repository layer
    ├── services/                      ⏳ Service layer
    ├── handlers/                      ⏳ Connect handlers
    ├── mappers/                       ⏳ Proto ↔ DB mappers
    └── module.go                      ⏳ FX wiring
```

### Database Schema Highlights
- **3 core tables**: `divisions`, `branches`, `departments`
- **4 views**: Division summary, Branch summary, Department hierarchy, Full org hierarchy
- **Hierarchical support**: Departments can be business-level or division-level
- **Location tracking**: Branches have lat/long for geo-queries
- **Soft deletes**: `is_active` flag instead of hard deletes

### Next Steps for Organization Module
1. Create repository layer
2. Create service layer with business logic
3. Create mappers (proto ↔ database)
4. Create Connect handlers
5. Wire everything in `module.go` using FX
6. Test integration

---

## 2. 📁 Document Management System (DMS) - PLANNED

### Purpose
Centralized service for all file uploads across the application.

### Why DMS?
**Before DMS:**
- Each module manages its own files
- Duplicate code for file upload/storage
- Inconsistent access control
- Hard to implement features like versioning, virus scanning

**With DMS:**
- Single source of truth for all documents
- Unified MinIO integration
- Centralized audit trail
- Consistent access control
- Easy to add features (OCR, watermarking, thumbnails)

### Core Entities

#### `documents` Table
```sql
- id (UUID)
- tenant_id
- file_name, file_size, mime_type
- storage_path (MinIO path)
- document_type (IDENTITY, CERTIFICATE, PROJECT_FILE, etc.)
- document_category (AADHAR, PAN, DEGREE, etc.)
- owner_entity_type (USER, PROJECT, FORM)
- owner_entity_id
- is_verified, verified_by_user_id
- expires_at (for passports, contracts)
- version, parent_document_id (versioning)
- virus_scan_status
- metadata (JSONB)
```

#### Supporting Tables
- `document_access_log` - Audit trail
- `document_shares` - Document sharing
- `document_thumbnails` - Preview images

### Integration Example: User Identity Documents

**Before:**
```sql
CREATE TABLE user_identity_documents (
    document_file_path TEXT,  -- Direct MinIO path
    ...
);
```

**After:**
```sql
CREATE TABLE user_identity_documents (
    document_id UUID,  -- Reference to DMS
    document_number VARCHAR(100),
    ...
);
```

**Flow:**
1. User uploads Aadhar → User Service
2. User Service → calls `DMS.UploadDocument()`
3. DMS stores file in MinIO, returns `document_id`
4. User Service stores `document_id` in `user_identity_documents`

### Implementation Priority
**Phase 1** (Week 1): Create DMS module
**Phase 2** (Week 2-3): Migrate existing modules
**Phase 3** (Week 4+): Enhanced features (versioning, OCR, thumbnails)

### Decision
✅ **RECOMMENDED**: Implement DMS as separate module

---

## 3. 🔐 User Permissions Analysis - NEEDS UPDATES

### Current State: 80% Sufficient

#### ✅ What Works
- Flexible permission model (namespace, resource, action)
- Hierarchical roles (parent_id)
- Effect-based (GRANT/FORBIDDEN)
- Permission definitions (templates)
- Multi-tenant support

#### ❌ What's Missing
**Organizational scope context** - Current permissions don't know about divisions, branches, departments.

### Example Problem
```protobuf
// Cannot express: "User can manage employees in Mumbai Branch only"
Permission {
  namespace: "user"
  resource: "employee"
  action: "update"
  subject: "user:123"
  // ❌ Missing: branch context
}
```

### Proposed Solution

#### 1. Add `PermissionScope` to `permission.proto`
```protobuf
message PermissionScope {
  ScopeLevel level = 1;  // BUSINESS, DIVISION, BRANCH, DEPARTMENT
  string division_id = 2;
  string branch_id = 3;
  string department_id = 4;
  repeated string division_ids = 5;  // For multi-branch managers
  repeated string branch_ids = 6;
  repeated string department_ids = 7;
}

message Permission {
  // ... existing fields ...
  PermissionScope scope = 8;  // NEW
}
```

#### 2. Add `UserOrganizationalContext` to `user.proto`
```protobuf
message UserTenantRole {
  // ... existing fields ...
  UserOrganizationalContext org_context = 5;  // NEW
}

message UserOrganizationalContext {
  AssignmentLevel level = 1;
  string division_id = 2;
  string branch_id = 3;
  string department_id = 4;
  bool is_primary = 8;
}
```

#### 3. New Service: `UserAssignmentService`
```protobuf
service UserAssignmentService {
  rpc AssignUserToOrganization(AssignUserRequest) returns (UserAssignment);
  rpc RemoveUserAssignment(...) returns (...);
  rpc GetUserAssignments(...) returns (...);
}
```

### Example Usage After Updates
```protobuf
// Division Manager - can manage all employees in Manufacturing Division
Permission {
  namespace: "user"
  resource: "employee"
  action: "update"
  subject: "role:Division_Manager"
  scope: {
    level: DIVISION
    division_id: "manufacturing-001"
  }
}

// Regional Manager - manages multiple branches
Permission {
  namespace: "user"
  resource: "employee"
  action: "view"
  subject: "user:789"
  scope: {
    level: BRANCH
    branch_ids: ["mumbai-001", "pune-001"]
  }
}
```

### Implementation Plan
**Phase 1** (Week 1): Proto updates + DB migrations
**Phase 2** (Week 2): Service layer with scope-aware permission checking
**Phase 3** (Week 3): Integration & testing

---

## 4. 📚 Documentation Strategy - OPTIONS PROVIDED

### 5 Options Analyzed

| Option | Cost | Ease of Setup | Maintenance | Recommendation |
|--------|------|---------------|-------------|----------------|
| 1. README.md per module | Free | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | Good for small teams |
| 2. GitHub Wiki | Free | ⭐⭐⭐⭐ | ⭐⭐⭐ | Good for collaboration |
| 3. GitBook/Docusaurus | $6.70/user | ⭐⭐⭐ | ⭐⭐⭐ | Professional docs |
| 4. Auto-generated API docs | Free | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | API reference only |
| 5. **Hybrid** ⭐ | **Free** | **⭐⭐⭐⭐** | **⭐⭐⭐⭐** | **BEST** |

### Recommended: Option 5 (Hybrid Approach)

#### Structure
```
backend/v2/
├── README.md (Project overview)
├── docs/
│   ├── getting-started.md
│   ├── architecture/
│   │   ├── system-overview.md
│   │   ├── database-design.md
│   │   └── inter-module-communication.md
│   ├── guides/
│   │   ├── development-guide.md
│   │   ├── testing-guide.md
│   │   └── deployment-guide.md
│   ├── api/
│   │   ├── organization.html (auto-generated from proto)
│   │   ├── notification.html (auto-generated)
│   │   └── user.html (auto-generated)
│   └── examples/
│       ├── create-division.md
│       └── send-notification.md
├── organization/
│   ├── README.md (Module quick reference)
│   └── ...
├── notification/
│   ├── README.md
│   └── ...
```

#### Documentation Types

| Type | Location | Tool | Purpose |
|------|----------|------|---------|
| Module Overview | `module/README.md` | Markdown | Quick reference |
| API Reference | `docs/api/*.html` | protoc-gen-doc | Auto-generated |
| Architecture | `docs/architecture/` | Markdown | Design decisions |
| Guides | `docs/guides/` | Markdown | How-to guides |
| Examples | `docs/examples/` | Markdown | Code examples |

#### Why Hybrid?
✅ Auto-generated API docs stay in sync
✅ Manual docs for architecture/context
✅ Module READMEs for quick reference
✅ Free and version controlled
✅ Works offline

#### Implementation
```bash
# Install protoc-gen-doc
go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@latest

# Add to Makefile
.PHONY: docs-generate
docs-generate:
	protoc --doc_out=./docs/api --doc_opt=html,index.html organization/proto/*.proto
	protoc --doc_out=./docs/api --doc_opt=html,notification.html notification/proto/*.proto
```

### Next Steps
**Decision needed**: Choose documentation approach (A, B, C, D, or E)

---

## Summary of Deliverables

### Documents Created Today
1. ✅ `docs/organizational-hierarchy-plan.md` - Complete system design
2. ✅ `docs/technical-specifications.md` - MinIO, audit logging, reporting
3. ✅ `docs/implementation-roadmap.md` - 8-10 week plan
4. ✅ `docs/README-ORGANIZATIONAL-SYSTEM.md` - Navigation guide
5. ✅ `docs/dms-architecture.md` - DMS design
6. ✅ `docs/user-permissions-analysis.md` - Permission system review
7. ✅ `docs/documentation-strategy-options.md` - 5 documentation approaches
8. ✅ `docs/SUMMARY-PLANNING-SESSION.md` - This document

### Code Created Today
1. ✅ Organization module foundation
   - Database schema
   - SQLC queries
   - Proto definitions
   - Generated code
   - Workspace configuration

---

## Key Insights

`✶ Insight ─────────────────────────────────────`

**1. Modular Architecture Pattern**
Following notification module's structure ensures consistency:
- Proto-first API design
- SQLC for type-safe database queries
- FX for dependency injection
- Separation: Repository → Service → Handler

**2. Centralization Benefits**
DMS exemplifies "single responsibility" - one module, one
purpose. This reduces code duplication and makes features
like virus scanning available everywhere automatically.

**3. Permission Scope Evolution**
Current permission system is solid but needs organizational
context. The addition is non-breaking - we're extending,
not replacing, which shows good initial design.

`─────────────────────────────────────────────────`

---

## Decisions Needed from You

### 1. ✅ DMS Implementation
**Question**: Approve DMS architecture and proceed with implementation?
**Options**:
- A. Yes, implement DMS as separate module
- B. No, each module manages its own files
- C. Defer decision for now

### 2. 📚 Documentation Strategy
**Question**: Which documentation approach?
**Options**:
- A. README.md only
- B. GitHub Wiki
- C. GitBook/Docusaurus
- D. Auto-generated API docs only
- E. Hybrid (README + Auto-docs + Central docs) ⭐ RECOMMENDED

### 3. 👥 User Profile Design
**Question**: Proceed with user profile tables from planning docs?
**Tables planned**:
- `user_profiles` (common)
- `employee_profiles`
- `vendor_profiles`
- `contractor_profiles`
- `employee_family_details`
- `user_identity_documents` → references DMS
- `user_certificates` → references DMS

**Options**:
- A. Yes, proceed as planned
- B. Make modifications first (specify what)
- C. Discuss further

### 4. 🔐 Permission Updates
**Question**: Approve permission scope additions?
**Changes**:
- Add `PermissionScope` to `permission.proto`
- Add `UserOrganizationalContext` to `user.proto`
- Create `UserAssignmentService`
- Database migrations

**Options**:
- A. Yes, proceed with updates
- B. Review and modify first
- C. Alternative approach

---

## Next Immediate Actions

### Today/This Week
1. ⏳ Get decisions on 4 questions above
2. ⏳ Complete organization module (repository, service, handlers)
3. ⏳ Start DMS module (if approved)
4. ⏳ Start documentation (based on chosen strategy)

### Next Week
1. ⏳ Implement user permission scope updates
2. ⏳ Create user profile tables
3. ⏳ Integrate DMS with user module
4. ⏳ Testing & integration

---

## Questions?

Please review the documents and let me know:
1. Which documentation strategy you prefer
2. If DMS architecture is approved
3. If user profile design is approved
4. If permission updates are approved

I'm ready to start implementing as soon as you give the green light! 🚀

