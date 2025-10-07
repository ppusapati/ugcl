# Organizational Hierarchy & User Management System - Documentation Index

## 📋 Overview

This documentation set provides a complete blueprint for implementing a comprehensive organizational hierarchy and user management system for multi-tenant SaaS applications.

**System Capabilities:**
- Multi-level organizational hierarchy (Business → Divisions → Branches, Departments)
- Multiple user types (Employee, Vendor, Contractor, Admin)
- Document management with MinIO storage
- Comprehensive audit logging
- Multi-dimensional reporting and analytics
- Hierarchical access control and permissions
- Background jobs and notifications

---

## 📚 Documentation Structure

### 1. [Organizational Hierarchy Plan](./organizational-hierarchy-plan.md)
**Purpose:** Complete system design and data model

**Contents:**
- Business requirements analysis
- Organizational structure (Divisions, Branches, Departments)
- User types and access patterns
- Complete database schema
  - 15+ tables covering all aspects
  - Relationships and constraints
  - Indexes for performance
- Service architecture
- API design (Proto messages)
- Permission & access control strategy
- Implementation phases

**When to read:** Start here to understand the overall system design and data model.

---

### 2. [Technical Specifications](./technical-specifications.md)
**Purpose:** Deep-dive into technical implementation details

**Contents:**
- **MinIO Document Storage**
  - Configuration and bucket structure
  - File naming conventions
  - Upload/download flow
  - Pre-signed URL generation
  - File validation rules

- **Audit Logging System**
  - Audit log tables and partitioning
  - Events to track (40+ event types)
  - Audit service interface
  - Usage examples

- **Reporting & Analytics**
  - 6 materialized views for reports
  - Report refresh strategies
  - Background job specifications
  - Export capabilities (CSV/PDF/Excel)

- **Performance Optimization**
  - Indexing strategy
  - Query optimization
  - Caching strategy

- **Security & Compliance**
  - Data encryption
  - Access control matrix
  - GDPR compliance

**When to read:** When implementing specific features like document upload, audit logging, or reporting.

---

### 3. [Implementation Roadmap](./implementation-roadmap.md)
**Purpose:** Step-by-step implementation guide

**Contents:**
- **8-Week Implementation Plan**
  - Week 1-2: Foundation (Database, MinIO)
  - Week 3: Proto definitions
  - Week 4: Organization service
  - Week 5: Profile services
  - Week 6: Assignment & access control
  - Week 7: Reporting
  - Week 8: Background jobs
  - Week 9-10: Testing & deployment

- **Detailed Daily Tasks**
  - Specific files to create
  - Implementation checklists
  - Testing requirements

- **File Structure Overview**
  - Complete directory tree
  - Module organization

- **Makefile Targets**
  - Proto generation
  - SQLC generation
  - Database migrations
  - Testing commands

- **Risk Mitigation**
- **Success Criteria**
- **Post-Launch Activities**

**When to read:** When planning sprints and assigning tasks to the development team.

---

## 🎯 Quick Start Guide

### For Project Managers
1. Read the [Organizational Hierarchy Plan](./organizational-hierarchy-plan.md) sections:
   - Section 1: Business Requirements
   - Section 6: Implementation Phases
2. Review the [Implementation Roadmap](./implementation-roadmap.md) for timeline and resource planning

### For Architects
1. Study the complete [Organizational Hierarchy Plan](./organizational-hierarchy-plan.md)
2. Review data models in Section 2
3. Understand service boundaries in Section 3
4. Review permission strategy in Section 4
5. Check [Technical Specifications](./technical-specifications.md) for infrastructure requirements

### For Backend Developers
1. Read [Implementation Roadmap](./implementation-roadmap.md) Phase 1-2 for setup
2. Follow phase-wise implementation from roadmap
3. Reference [Technical Specifications](./technical-specifications.md) while implementing:
   - MinIO integration
   - Audit logging
   - Reporting
4. Use [Organizational Hierarchy Plan](./organizational-hierarchy-plan.md) Section 5 for API contracts

### For DevOps Engineers
1. Review [Technical Specifications](./technical-specifications.md):
   - Section 1: MinIO setup
   - Section 5: Performance optimization
2. Review [Implementation Roadmap](./implementation-roadmap.md):
   - Week 10 Day 4: Deployment preparation
   - Monitoring and logging setup

---

## 🏗️ System Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                        Client Layer                          │
│  (Web App, Mobile App, API Clients)                         │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│                   API Gateway / Load Balancer                │
└──────────────────────┬──────────────────────────────────────┘
                       │
       ┌───────────────┼───────────────┐
       │               │               │
┌──────▼──────┐ ┌─────▼──────┐ ┌─────▼──────────┐
│  Identity   │ │   Profile  │ │   Reporting    │
│  Services   │ │  Services  │ │   Services     │
├─────────────┤ ├────────────┤ ├────────────────┤
│ • Tenant    │ │ • User     │ │ • Headcount    │
│ • User      │ │   Profile  │ │ • Documents    │
│ • Org       │ │ • Employee │ │ • Joiners/     │
│   (Division,│ │ • Vendor   │ │   Leavers      │
│    Branch,  │ │ • Contractor│ │ • Export       │
│    Dept)    │ │ • Document │ │                │
│ • Auth      │ │ • Family   │ │                │
└──────┬──────┘ └─────┬──────┘ └─────┬──────────┘
       │              │              │
       └──────────────┼──────────────┘
                      │
       ┌──────────────┼──────────────────────┐
       │              │                      │
┌──────▼──────┐ ┌────▼─────────┐ ┌─────────▼──────┐
│   Packages  │ │  Background  │ │   Notification │
│             │ │     Jobs     │ │     Service    │
├─────────────┤ ├──────────────┤ ├────────────────┤
│ • Audit     │ │ • Doc Expiry │ │ • Email        │
│ • Storage   │ │ • Contract   │ │ • SMS          │
│   (MinIO)   │ │   Expiry     │ │ • Push         │
│ • Events    │ │ • Probation  │ │                │
│ • Cache     │ │ • Report     │ │                │
│             │ │   Refresh    │ │                │
└──────┬──────┘ └──────┬───────┘ └────────┬───────┘
       │               │                  │
       └───────────────┼──────────────────┘
                       │
       ┌───────────────┴────────────────┐
       │                                │
┌──────▼──────────┐           ┌────────▼─────────┐
│   PostgreSQL    │           │      MinIO       │
│                 │           │                  │
│ • Tenants       │           │ • Documents      │
│ • Users         │           │ • Certificates   │
│ • Profiles      │           │ • Photos         │
│ • Divisions     │           │                  │
│ • Branches      │           │ Buckets:         │
│ • Departments   │           │ ugcl-documents/  │
│ • Assignments   │           │ ├── {tenant}/    │
│ • Audit Logs    │           │    ├── employee/ │
│ • Reports       │           │    ├── vendor/   │
│                 │           │    └── contractor│
└─────────────────┘           └──────────────────┘
```

---

## 📊 Database Schema Overview

### Core Tables (15+)

**Organizational Hierarchy:**
- `divisions` - Business divisions
- `branches` - Physical locations
- `departments` - Functional units (business/division level)

**User Management:**
- `users` - Authentication (existing, updated)
- `user_profiles` - Common profile data
- `employee_profiles` - Employee-specific data
- `vendor_profiles` - Vendor company details
- `contractor_profiles` - Contract details
- `employee_family_details` - Family information

**Documents:**
- `user_identity_documents` - Aadhar, PAN, Passport
- `user_certificates` - Professional certificates

**Access Control:**
- `user_organizational_assignments` - Multi-level assignments
- `user_project_assignments` - Project-based access
- `user_tenant_roles` - Updated with hierarchy fields

**Audit & Tracking:**
- `audit_logs` - All system changes
- `user_activity_logs` - Login/authentication events

**Reporting (Materialized Views):**
- `report_headcount_by_division`
- `report_headcount_by_branch`
- `report_headcount_by_department`
- `report_document_expiry_tracker`
- `report_new_joiners_leavers`
- `report_employee_by_designation`

---

## 🔑 Key Features

### 1. Multi-Level Organizational Hierarchy
```
Tenant (Business)
├── Division 1
│   ├── Branch 1.1
│   ├── Branch 1.2
│   └── Department (Division-level)
├── Division 2
│   └── Branch 2.1
└── Department (Business-level, cross-cutting)
```

### 2. User Type Support
- **Employee**: Full profile, family details, branch assignment
- **Vendor**: Company details, GST/PAN, multi-project access
- **Contractor**: Time-bound contracts, skills, billing
- **Admin**: Hierarchical admin at any level

### 3. Document Management
- Identity documents (Aadhar, PAN, Passport)
- Professional certificates
- Secure storage in MinIO
- Pre-signed URLs for access
- Document verification workflow
- Expiry tracking and notifications

### 4. Comprehensive Audit Trail
- Track all user actions
- Before/after change tracking
- IP address and user agent logging
- Event categorization
- Searchable audit logs

### 5. Powerful Reporting
- Real-time headcount by any dimension
- Document expiry tracking
- Joiners/Leavers analysis
- Export to CSV/PDF/Excel
- Automated report refresh

### 6. Hierarchical Permissions
- Permission cascade (Business → Division → Branch)
- Multi-level assignments
- Role-based access control
- Exception handling

---

## 🔧 Technology Stack

### Backend
- **Language**: Go 1.21+
- **API Protocol**: gRPC + Connect
- **Database**: PostgreSQL 15+
- **ORM**: SQLC (type-safe SQL)
- **Migrations**: golang-migrate

### Storage
- **Object Storage**: MinIO (S3-compatible)
- **File Types**: PDF, JPG, PNG (max 10MB)
- **Security**: Pre-signed URLs, encryption at rest

### Infrastructure
- **Caching**: Redis (optional, for permission cache)
- **Job Scheduler**: pg_cron or custom worker
- **Monitoring**: Prometheus + Grafana
- **Logging**: ELK Stack

---

## 📈 Performance Targets

| Metric | Target | Notes |
|--------|--------|-------|
| API Response Time (p95) | < 200ms | For simple CRUD operations |
| API Response Time (p99) | < 500ms | For complex queries |
| File Upload (10MB) | < 5s | Including MinIO upload |
| Report Generation | < 2s | For materialized views |
| Concurrent Users | 1000+ | Per tenant |
| Database Size | 100GB+ | With 10,000+ users |

---

## 🔒 Security Considerations

### Data Protection
- ✅ Encryption at rest (sensitive fields)
- ✅ Encryption in transit (TLS)
- ✅ Pre-signed URLs for file access (1-hour expiry)
- ✅ Password hashing (bcrypt)
- ✅ Two-factor authentication support

### Access Control
- ✅ Role-based access control (RBAC)
- ✅ Hierarchical permissions
- ✅ Default deny policy
- ✅ Audit logging for all access

### Compliance
- ✅ GDPR right to access
- ✅ GDPR right to erasure (anonymization)
- ✅ Data portability (export)
- ✅ Consent tracking

---

## 🚀 Getting Started

### Prerequisites
```bash
# Install tools
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# Install MinIO
brew install minio/stable/minio  # macOS
# or download from https://min.io/download
```

### Setup Development Environment
```bash
# 1. Clone repository
git clone <repo-url>
cd backend/v2

# 2. Install dependencies
go mod download

# 3. Setup MinIO
make minio-setup

# 4. Run migrations
make migrate-up

# 5. Generate code
make sqlc-gen-all
make proto-gen-all

# 6. Run server
make run
```

---

## 📞 Support & Contribution

### Documentation Updates
- All documentation is in Markdown
- Keep documentation in sync with code changes
- Use diagrams where helpful (Mermaid supported)

### Code Reviews
- All PRs require approval
- Check against implementation roadmap
- Ensure tests pass (>80% coverage)

### Questions?
- Create an issue in the repository
- Tag with appropriate labels (question, bug, feature)

---

## 📅 Timeline Summary

| Phase | Duration | Deliverables |
|-------|----------|--------------|
| **Phase 1**: Foundation | Week 1-2 | Database schema, MinIO, Audit |
| **Phase 2**: Proto & Codegen | Week 3 | Proto files, generated code |
| **Phase 3**: Organization Service | Week 4 | Division, Branch, Department APIs |
| **Phase 4**: Profile Service | Week 5 | User, Employee, Vendor, Document APIs |
| **Phase 5**: Access Control | Week 6 | Assignment, Permission resolution |
| **Phase 6**: Reporting | Week 7 | Reports, Analytics, Export |
| **Phase 7**: Background Jobs | Week 8 | Notifications, Scheduled tasks |
| **Phase 8**: Testing & Deployment | Week 9-10 | Integration tests, Production deploy |

**Total**: 8-10 weeks with 2-3 developers

---

## 🎓 Learning Resources

### Understanding the System
1. Start with [Business Requirements](./organizational-hierarchy-plan.md#1-business-requirements-summary)
2. Study [Data Models](./organizational-hierarchy-plan.md#2-data-model-design)
3. Review [API Design](./organizational-hierarchy-plan.md#5-api-design-proto-messages)

### Implementation Guides
1. Follow [Implementation Roadmap](./implementation-roadmap.md) phase by phase
2. Reference [Technical Specifications](./technical-specifications.md) for details
3. Use Makefile targets for automation

### Best Practices
- Always audit critical operations
- Use transactions for multi-table updates
- Cache permission decisions
- Validate input at API boundary
- Test with realistic data volumes

---

## ✅ Checklist for Success

### Before Starting Development
- [ ] Read all three documentation files
- [ ] Understand organizational hierarchy
- [ ] Review database schema
- [ ] Setup development environment
- [ ] Understand MinIO integration

### During Development
- [ ] Follow implementation roadmap
- [ ] Write tests alongside code (TDD)
- [ ] Document API changes
- [ ] Use audit logging for all changes
- [ ] Optimize queries with indexes

### Before Production Deployment
- [ ] All tests passing (unit + integration)
- [ ] Performance benchmarks met
- [ ] Security audit completed
- [ ] Documentation updated
- [ ] Monitoring and alerting configured
- [ ] Backup and disaster recovery plan
- [ ] Team trained on new system

---

## 🎯 Next Steps

1. ✅ **Documentation Complete** - You are here
2. ⏭️ **Start Phase 1** - Create database schema
3. ⏭️ **Setup MinIO** - Configure object storage
4. ⏭️ **Implement Services** - Follow roadmap week by week
5. ⏭️ **Deploy to Production** - Week 10

**Ready to start coding!** 🚀

Refer to [Implementation Roadmap](./implementation-roadmap.md) for detailed day-by-day tasks.

---

*Last Updated: 2025-10-04*
*Version: 1.0*
