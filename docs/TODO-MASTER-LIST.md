# Master TODO List - Backend v2 Modernization

**Last Updated:** October 4, 2025
**Total Tasks:** 64
**Estimated Duration:** 13 weeks (3.25 months)

---

## Overview

This document tracks the complete architectural modernization roadmap for the backend v2 system, covering:

1. **Organization Module** - Hierarchical organizational structure (Divisions → Branches → Departments)
2. **Document Management System (DMS)** - Centralized file storage with MinIO
3. **Permission System Enhancement** - Organizational scope-aware permissions
4. **Entity-Based Identity Architecture** - Universal identity abstraction for humans, machines, and systems
5. **Personnel Domain Module** - Centralized employee, vendor, contractor management
6. **Service Integration** - Migration of existing services to new architecture
7. **Documentation Strategy** - Hybrid approach (auto-generated + manual)

---

## Architecture Vision

```
┌─────────────────────────────────────────────────────────────┐
│                    IDENTITY LAYER                           │
│  ┌────────┐     ┌─────────────┐     ┌────────────────┐    │
│  │ users  │────▶│  entities   │────▶│ entity_roles   │    │
│  │        │     │(abstraction)│     │   (RBAC)       │    │
│  └────────┘     └─────────────┘     └────────────────┘    │
│                         │                                   │
│                         │ Polymorphic entity_id             │
└─────────────────────────┼───────────────────────────────────┘
                          │
        ┌─────────────────┼─────────────────┐
        │                 │                 │
        ▼                 ▼                 ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│ ORGANIZATION │  │  PERSONNEL   │  │     DMS      │
│              │  │              │  │              │
│ ┌──────────┐ │  │ ┌──────────┐ │  │ ┌──────────┐ │
│ │divisions │ │  │ │employees │ │  │ │documents │ │
│ │branches  │ │  │ │vendors   │ │  │ │ (MinIO)  │ │
│ │departments│ │  │ │contractors│ │  │ └──────────┘ │
│ └──────────┘ │  │ └──────────┘ │  └──────────────┘
└──────────────┘  └──────────────┘
        │                 │                 │
        └─────────────────┴─────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              CROSS-CUTTING SERVICES                         │
│  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌──────────┐ │
│  │   HR   │ │Finance │ │ Leave  │ │Projects│ │FormBuilder│
│  └────────┘ └────────┘ └────────┘ └────────┘ └──────────┘ │
└─────────────────────────────────────────────────────────────┘
```

---

## Phase 1: Organization Module (6 tasks)

**Status:** Partially Complete
**Duration:** 1-2 weeks
**Priority:** P0

### Tasks

- [ ] **Task 1:** Complete organization module - create repository layer
- [ ] **Task 2:** Complete organization module - create service layer with business logic
- [ ] **Task 3:** Complete organization module - create mappers (proto ↔ database)
- [ ] **Task 4:** Complete organization module - create Connect handlers
- [ ] **Task 5:** Complete organization module - wire everything in module.go using FX
- [ ] **Task 6:** Test organization module integration

### Deliverable
Working organization service with Division → Branch → Department hierarchy

### Current State
✅ Database schema created
✅ SQLC queries defined
✅ Proto definitions complete
✅ Generated code ready
⏳ Service layer pending

---

## Phase 2: Document Management System (9 tasks)

**Status:** Not Started
**Duration:** 2 weeks
**Priority:** P1

### Tasks

- [ ] **Task 7:** Design DMS (Document Management System) module architecture
- [ ] **Task 8:** Create DMS proto definitions (document.proto) with service definitions
- [ ] **Task 9:** Create DMS database schema (documents, document_access_log, document_shares, document_thumbnails)
- [ ] **Task 10:** Create SQLC queries for DMS operations
- [ ] **Task 11:** Implement DMS repository layer with MinIO integration
- [ ] **Task 12:** Implement DMS service layer (upload, download, versioning, access control)
- [ ] **Task 13:** Create DMS Connect handlers
- [ ] **Task 14:** Wire DMS module with FX dependency injection
- [ ] **Task 15:** Integrate DMS with existing modules (formbuilder, user identity documents)

### Deliverable
Centralized document service with MinIO storage, versioning, and access control

### Key Features
- Unified file upload/download API
- Document versioning
- Access control and audit trail
- Virus scanning integration
- Document sharing
- Thumbnail generation

---

## Phase 3: Permission System Enhancement (5 tasks)

**Status:** Not Started
**Duration:** 1 week
**Priority:** P2

### Tasks

- [ ] **Task 16:** Update user permissions - add PermissionScope to permission.proto
- [ ] **Task 17:** Update user permissions - add UserOrganizationalContext to user.proto
- [ ] **Task 18:** Create UserAssignmentService for organizational assignments
- [ ] **Task 19:** Create database migrations for permission scope updates
- [ ] **Task 20:** Implement scope-aware permission checking in service layer

### Deliverable
Permissions that understand organizational context (divisions, branches, departments)

### Example Usage
```protobuf
// Division Manager - can manage employees in Manufacturing Division
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
```

---

## Phase 4: Entity-Based Identity Architecture (8 tasks)

**Status:** Not Started
**Duration:** 2 weeks
**Priority:** P0 (Most Foundational)

### Tasks

- [ ] **Task 21:** Design and implement identity/entity module with Entity abstraction
- [ ] **Task 22:** Create entity.proto with EntityType enum, Entity message, EntityRoleBinding, and EntityService
- [ ] **Task 23:** Create database schema for entities and entity_role_bindings tables
- [ ] **Task 24:** Create SQLC queries for entity operations (CRUD + role bindings)
- [ ] **Task 25:** Implement repository layer for entity module
- [ ] **Task 26:** Implement service layer for entity module with business logic
- [ ] **Task 27:** Create Connect handlers for entity gRPC service
- [ ] **Task 28:** Wire entity module using FX dependency injection

### Deliverable
Universal identity system supporting:
- **Human entities:** Employees, Contractors, Vendors, Clients, Agents
- **Machine entities:** Devices, Drones, Sensors, Gateways
- **System entities:** Bots, Services, Integrations

### Entity Types
```protobuf
enum EntityType {
  // Human entities
  ENTITY_TYPE_EMPLOYEE = 1;
  ENTITY_TYPE_CONTRACTOR = 2;
  ENTITY_TYPE_VENDOR_CONTACT = 3;
  ENTITY_TYPE_CLIENT_CONTACT = 4;
  ENTITY_TYPE_AGENT = 5;

  // Machine entities
  ENTITY_TYPE_DEVICE = 100;
  ENTITY_TYPE_DRONE = 101;
  ENTITY_TYPE_SENSOR = 102;
  ENTITY_TYPE_GATEWAY = 103;

  // System entities
  ENTITY_TYPE_BOT = 200;
  ENTITY_TYPE_SERVICE = 201;
  ENTITY_TYPE_INTEGRATION = 202;
}
```

### Benefits
✅ Unified access control for all entity types
✅ Polymorphic relationships (forms owned by any entity)
✅ Simplified audit trail
✅ IoT/SCADA integration ready
✅ Future-proof (easy to add new entity types)

---

## Phase 5: Personnel Domain Module (8 tasks)

**Status:** Not Started
**Duration:** 2 weeks
**Priority:** P1

### Tasks

- [ ] **Task 29:** Create personnel module with employee.proto, vendor.proto, contractor.proto
- [ ] **Task 30:** Create personnel database schemas (employees, vendors, contractors, employee_family, employee_documents)
- [ ] **Task 31:** Create SQLC queries for personnel operations with organizational queries
- [ ] **Task 32:** Implement personnel repository layer
- [ ] **Task 33:** Implement personnel service layer with entity integration
- [ ] **Task 34:** Create personnel Connect handlers
- [ ] **Task 35:** Migrate contractor from vendors/ to personnel/ module
- [ ] **Task 36:** Update personnel domain services to reference entity_id from identity/entity

### Deliverable
Centralized personnel data service accessible to all other services (HR, Finance, Leave, etc.)

### Module Structure
```
personnel/
├── proto/
│   ├── employee.proto
│   ├── vendor.proto
│   └── contractor.proto
├── db/
│   ├── schema/
│   │   ├── employees.sql
│   │   ├── vendors.sql
│   │   ├── contractors.sql
│   │   ├── employee_family.sql
│   │   └── employee_documents.sql
│   ├── queries/
│   └── generated/
├── repository/
├── services/
├── handlers/
├── mappers/
└── module.go
```

### Key Features
- Employee management with organizational assignment
- Vendor registration and management
- Contractor lifecycle management
- Family details tracking
- Document management (via DMS)
- Organizational queries (by division, branch, department)

---

## Phase 6: Service Integration & Migration (4 tasks)

**Status:** Not Started
**Duration:** 2 weeks
**Priority:** P2

### Tasks

- [ ] **Task 37:** Update existing services (HR, Finance, Leave) to use personnel API instead of duplicate data
- [ ] **Task 38:** Update existing services to use entity-based permissions
- [ ] **Task 39:** Create migration scripts for transitioning existing data to entity model
- [ ] **Task 40:** Update documentation with entity-based architecture design

### Deliverable
Fully integrated system with:
- No data duplication
- Unified permission model
- All services using entity abstraction

### Migration Strategy
1. Create entity records for existing users
2. Migrate employee data to personnel service
3. Update service APIs to use personnel client
4. Remove duplicate data from services
5. Update permission checks to use entity-based model

---

## Phase 7: Documentation Strategy - Hybrid Approach (24 tasks)

**Status:** Not Started
**Duration:** 3 weeks
**Priority:** P2-P3

### 7.1: Auto-Generated API Documentation (4 tasks)

- [ ] **Task 41:** Install and configure protoc-gen-doc for automated API documentation
- [ ] **Task 42:** Add Makefile targets for generating API docs from proto files
- [ ] **Task 43:** Generate initial API documentation HTML files for all modules
- [ ] **Task 44:** Create docs/api/ directory structure for auto-generated documentation

### 7.2: Module-Level READMEs (5 tasks)

- [ ] **Task 45:** Write module-level README.md for organization module
- [ ] **Task 46:** Write module-level README.md for DMS module
- [ ] **Task 47:** Write module-level README.md for entity module
- [ ] **Task 48:** Write module-level README.md for personnel module
- [ ] **Task 49:** Write module-level README.md files for all other existing modules

### 7.3: Architecture Documentation (4 tasks)

- [ ] **Task 50:** Create docs/architecture/system-overview.md with high-level architecture
- [ ] **Task 51:** Create docs/architecture/database-design.md with ER diagrams and schema docs
- [ ] **Task 52:** Create docs/architecture/inter-module-communication.md documenting service dependencies
- [ ] **Task 53:** Create docs/architecture/entity-identity-model.md explaining entity abstraction

### 7.4: Developer Guides (4 tasks)

- [ ] **Task 54:** Create docs/guides/development-guide.md with setup and coding standards
- [ ] **Task 55:** Create docs/guides/testing-guide.md with testing strategies and examples
- [ ] **Task 56:** Create docs/guides/deployment-guide.md with deployment procedures
- [ ] **Task 57:** Create docs/guides/proto-first-development.md explaining proto-first workflow

### 7.5: Code Examples (4 tasks)

- [ ] **Task 58:** Create docs/examples/create-division-branch-department.md with code examples
- [ ] **Task 59:** Create docs/examples/upload-document-to-dms.md with code examples
- [ ] **Task 60:** Create docs/examples/create-employee-with-entity.md with code examples
- [ ] **Task 61:** Create docs/examples/check-permissions-with-scope.md with code examples

### 7.6: Documentation Infrastructure (3 tasks)

- [ ] **Task 62:** Update root README.md with navigation to all documentation
- [ ] **Task 63:** Create docs/getting-started.md for new developers
- [ ] **Task 64:** Set up CI/CD pipeline to auto-generate docs on proto changes

### Deliverable
Comprehensive documentation system with:
- Auto-generated API docs (always in sync)
- Architecture decision records
- Developer onboarding guides
- Practical code examples
- CI/CD automated regeneration

### Documentation Structure
```
backend/v2/
├── README.md                     # Updated with full navigation
├── docs/
│   ├── getting-started.md
│   ├── api/                      # Auto-generated from proto
│   │   ├── index.html
│   │   ├── organization.html
│   │   ├── dms.html
│   │   ├── entity.html
│   │   └── personnel.html
│   ├── architecture/             # Manual architecture docs
│   │   ├── system-overview.md
│   │   ├── database-design.md
│   │   ├── inter-module-communication.md
│   │   └── entity-identity-model.md
│   ├── guides/                   # Developer how-to guides
│   │   ├── development-guide.md
│   │   ├── testing-guide.md
│   │   ├── deployment-guide.md
│   │   └── proto-first-development.md
│   └── examples/                 # Code examples
│       ├── create-division-branch-department.md
│       ├── upload-document-to-dms.md
│       ├── create-employee-with-entity.md
│       └── check-permissions-with-scope.md
├── organization/
│   ├── README.md                 # Module quick reference
│   └── ...
├── dms/
│   ├── README.md
│   └── ...
├── identity/entity/
│   ├── README.md
│   └── ...
└── personnel/
    ├── README.md
    └── ...
```

---

## Implementation Timeline

```
┌────────────────────────────────────────────────────────────┐
│                    13-Week Roadmap                         │
└────────────────────────────────────────────────────────────┘

Week 1-2:   ████████░░ Organization Module (Tasks 1-6)
            ████████░░ DMS Foundation (Tasks 7-11)

Week 3-4:   ░░████████ DMS Completion (Tasks 12-15)
            ░░████████ Permission Enhancement (Tasks 16-20)

Week 5-6:   ████████░░ Entity Architecture (Tasks 21-28)

Week 7-8:   ████████░░ Personnel Module (Tasks 29-34)

Week 9:     ████████░░ Personnel Migration (Tasks 35-36)

Week 10:    ████████░░ Service Integration (Tasks 37-40)

Week 11:    ████████░░ Auto-gen Docs + READMEs (Tasks 41-49)

Week 12:    ████████░░ Architecture Docs + Guides (Tasks 50-57)

Week 13:    ████████░░ Examples + Infrastructure (Tasks 58-64)
```

---

## Priority Matrix

| Priority | Tasks | Phase | Duration | Dependencies |
|----------|-------|-------|----------|--------------|
| **P0** | 1-6 | Organization Module | 1-2 weeks | None |
| **P0** | 21-28 | Entity Architecture | 2 weeks | None |
| **P1** | 29-36 | Personnel Module | 2 weeks | Tasks 1-6, 21-28 |
| **P1** | 7-15 | DMS | 2 weeks | None |
| **P2** | 16-20 | Permission Updates | 1 week | Tasks 1-6 |
| **P2** | 37-40 | Service Integration | 2 weeks | Tasks 21-36 |
| **P2** | 41-49 | Docs: Auto-gen + READMEs | 1 week | Tasks 1-40 |
| **P3** | 50-61 | Docs: Architecture + Examples | 2 weeks | Tasks 1-40 |
| **P3** | 62-64 | Docs: Infrastructure | 3 days | Tasks 41-61 |

---

## Parallel Execution Opportunities

```
┌─────────────────┐     ┌─────────────────┐
│  Organization   │     │  Entity Arch    │
│   (Tasks 1-6)   │  ║  │  (Tasks 21-28)  │  ← Can run in parallel
└─────────────────┘     └─────────────────┘
        │                       │
        └───────────┬───────────┘
                    ▼
        ┌─────────────────────┐
        │  Personnel Module   │
        │   (Tasks 29-36)     │  ← Requires both above
        └─────────────────────┘

┌─────────────────┐
│      DMS        │
│  (Tasks 7-15)   │  ← Can run parallel to Entity work
└─────────────────┘
```

---

## Key Decisions Made

### ✅ DMS Implementation
**Decision:** Implement DMS as a separate centralized module
**Rationale:** Single source of truth for all documents, unified access control, easier to add features like versioning and virus scanning

### ✅ Documentation Strategy
**Decision:** Hybrid approach (Option E)
**Components:**
- Auto-generated API docs from proto files
- Manual architecture documentation
- Module-level READMEs
- Code examples
- Developer guides

### ✅ User Profile Design
**Decision:** Modify approach - Create Personnel module with entity references
**Rationale:**
- Centralized employee/vendor/contractor data
- Reusable by HR, Finance, Leave services
- Eliminates data duplication

### ✅ Permission System
**Decision:** Add organizational scope to existing permission model
**Changes:**
- Add `PermissionScope` message
- Add `UserOrganizationalContext`
- Create `UserAssignmentService`
- Non-breaking extension of existing model

### ✅ Entity-Based Identity
**Decision:** Implement universal entity abstraction
**Rationale:**
- Handles humans, machines, and systems uniformly
- Future-proof for IoT/SCADA integration
- Polymorphic relationships and unified audit trail
- Easy to add new entity types (drones, bots, partners)

---

## Success Metrics

### Technical Metrics
- [ ] All modules pass integration tests
- [ ] Zero data duplication across services
- [ ] API documentation auto-generates successfully
- [ ] Permission checks work with organizational scope
- [ ] Entity model supports all current and planned entity types

### Documentation Metrics
- [ ] All modules have README.md files
- [ ] API documentation coverage: 100%
- [ ] Architecture docs created for all major systems
- [ ] At least 4 code examples documented
- [ ] CI/CD pipeline regenerates docs on proto changes

### Performance Metrics
- [ ] Personnel API response time < 100ms
- [ ] DMS upload/download latency < 500ms
- [ ] Permission check latency < 50ms
- [ ] Entity lookup latency < 50ms

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Entity model too complex | Medium | High | Start simple, iterate based on feedback |
| Data migration issues | High | High | Create comprehensive migration scripts with rollback |
| Service integration breaks existing functionality | Medium | Critical | Implement feature flags, gradual rollout |
| Documentation becomes stale | Medium | Medium | CI/CD automation, make it part of PR checklist |
| MinIO integration issues | Low | Medium | Test thoroughly in dev environment first |
| Performance degradation | Medium | High | Performance testing at each phase |

---

## Next Steps

### Immediate Actions (This Week)
1. **Choose starting point:**
   - Option A: Task 21 (Entity architecture - most foundational)
   - Option B: Tasks 1-6 (Organization - quick win, already started)
   - Option C: Tasks 41-44 (Set up doc generation infrastructure)

2. **Set up development environment**
3. **Review and approve this roadmap**
4. **Assign ownership for each phase**

### Communication Plan
- Weekly status updates on task completion
- Bi-weekly architecture review sessions
- Monthly demo of completed phases
- Documentation reviews before each phase completion

---

## References

### Related Documents
- [Organizational Hierarchy Plan](organizational-hierarchy-plan.md)
- [Technical Specifications](technical-specifications.md)
- [Implementation Roadmap](implementation-roadmap.md)
- [DMS Architecture](dms-architecture.md)
- [User Permissions Analysis](user-permissions-analysis.md)
- [Documentation Strategy Options](documentation-strategy-options.md)
- [Planning Session Summary](SUMMARY-PLANNING-SESSION.md)

### External Resources
- [protoc-gen-doc](https://github.com/pseudomuto/protoc-gen-doc)
- [Connect Protocol](https://connectrpc.com/)
- [SQLC Documentation](https://docs.sqlc.dev/)
- [Uber FX](https://uber-go.github.io/fx/)
- [MinIO Documentation](https://min.io/docs/)

---

**Document Status:** Living Document
**Review Frequency:** Weekly
**Owner:** Architecture Team
**Last Review:** October 4, 2025



- [ ] Remove password management stubs from user.proto

- [ ] Remove email/phone verification stubs from user.proto

- [ ] Delete unused OTP model from user module

- [ ] Update user service to remove stub implementations

- [ ] Verify no references to removed functionality exist