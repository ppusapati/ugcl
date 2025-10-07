# UGCL Backend v2 Documentation

Welcome to the UGCL Backend v2 documentation. This directory contains comprehensive documentation for the backend system.

## Documentation Structure

### API Documentation
- **[api/](./api/)** - Auto-generated API documentation from Protocol Buffer definitions
  - Comprehensive HTML documentation for all services and message types
  - Generated using `protoc-gen-doc`
  - See [api/README.md](./api/README.md) for details

### Architecture & Planning Documents
- **[dms-architecture.md](./dms-architecture.md)** - Document Management System architecture
- **[organizational-hierarchy-plan.md](./organizational-hierarchy-plan.md)** - Organizational structure planning
- **[implementation-roadmap.md](./implementation-roadmap.md)** - Implementation roadmap and timeline
- **[technical-specifications.md](./technical-specifications.md)** - Technical specifications
- **[user-permissions-analysis.md](./user-permissions-analysis.md)** - User permissions system analysis

### Module Documentation
- **[PERSONNEL-MODULE-RESTRUCTURING.md](./PERSONNEL-MODULE-RESTRUCTURING.md)** - Personnel module restructuring guide
- **[README-ORGANIZATIONAL-SYSTEM.md](./README-ORGANIZATIONAL-SYSTEM.md)** - Organizational system overview

### Implementation Summaries
- **[PERMISSION-IMPLEMENTATION-SUMMARY.md](./PERMISSION-IMPLEMENTATION-SUMMARY.md)** - Permission system implementation
- **[PERMISSION-SYSTEM-ANALYSIS.md](./PERMISSION-SYSTEM-ANALYSIS.md)** - Permission system analysis
- **[IMPLEMENTATION-SUMMARY.md](./IMPLEMENTATION-SUMMARY.md)** - General implementation summary
- **[FINAL-SESSION-SUMMARY.md](./FINAL-SESSION-SUMMARY.md)** - Final session summary

### Planning & Status
- **[SUMMARY-PLANNING-SESSION.md](./SUMMARY-PLANNING-SESSION.md)** - Planning session summaries
- **[SESSION-TODO-LIST.md](./SESSION-TODO-LIST.md)** - Session-based TODO tracking
- **[TODO-MASTER-LIST.md](./TODO-MASTER-LIST.md)** - Master TODO list
- **[QUICK-STATUS.md](./QUICK-STATUS.md)** - Quick status overview
- **[documentation-strategy-options.md](./documentation-strategy-options.md)** - Documentation strategy options

## Quick Start

### Viewing API Documentation

1. Generate the documentation:
   ```bash
   make docs-api
   ```

2. Serve it locally:
   ```bash
   make docs-serve
   ```

3. Open your browser to: `http://localhost:8000/docs/api/`

### Documentation Maintenance

| Task | Command |
|------|---------|
| Generate API docs | `make docs-api` |
| Generate all docs | `make docs` |
| Clean generated docs | `make docs-clean` |
| Serve docs locally | `make docs-serve` |

## Module Overview

The UGCL Backend v2 system consists of the following major modules:

### Core Infrastructure
- **Core** - Core utilities and shared functionality
- **Packages** - Shared packages and common types

### Identity & Access Management
- **Identity/Auth** - Authentication services
- **Identity/User** - User management and permissions
- **Identity/Tenant** - Multi-tenant management
- **Identity/Entity** - Entity management

### Human Resources Management
- **Employee** - Employee management and records
- **Contractors** - Contractor management
- **Vendors** - Vendor management

### Organizational Management
- **Organization** - Division, branch, and department hierarchy

### Document & Data Management
- **DMS** - Document Management System
- **DataBridge** - Data integration and bridging
- **DataArchive** - Long-term data archival
- **BackupDR** - Backup and disaster recovery

### Analytics & Insights
- **InsightHub** - Central insights and analytics hub
- **InsightViewer** - Insights visualization
- **MetaSearch** - Advanced search capabilities

### Workflow & Forms
- **FormBuilder** - Dynamic form creation and management
- **Approval Workflow** - Approval process management (integrated with FormBuilder)

### Operations
- **Scheduler** - Task scheduling and cron jobs
- **Notification** - Multi-channel notifications
- **Masters** - Master data management
- **Pipeline** - Data pipeline management

### Projects
- **Projects** - Project management and tracking

## Architecture Patterns

The system follows these key architectural patterns:

1. **Microservices Architecture** - Each module is independently deployable
2. **Event-Driven Design** - Modules communicate via events
3. **Domain-Driven Design** - Clear bounded contexts per module
4. **API-First Design** - Protocol Buffers for all service definitions
5. **Multi-Tenancy** - Built-in tenant isolation and management

## Development Workflow

### 1. API Design
- Define services in `.proto` files
- Document all fields, messages, and services
- Generate code with `make proto-generate`

### 2. Database Schema
- Define schema in `db/schema/` directory
- Write queries in `db/queries/` directory
- Generate code with `make sqlc-generate-all`

### 3. Documentation
- Add comments to proto files
- Generate API docs with `make docs-api`
- Update relevant markdown files

### 4. Testing
- Write unit tests for services
- Integration tests for workflows
- Document test scenarios

## Contributing

When adding or updating modules:

1. Update the relevant `.proto` files with clear documentation
2. Update database schemas if needed
3. Regenerate code and documentation
4. Update this README if adding new modules
5. Add any new architectural decisions to technical-specifications.md

## Related Documentation

- [Main Project README](../Readme.md)
- [API Documentation](./api/README.md)
- [Report System README](../REPORT_SYSTEM_README.md)

## Support & Contact

For questions or issues, please refer to the specific module documentation or contact the development team.

---

Last Updated: 2025-10-05
