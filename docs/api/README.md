# API Documentation

This directory contains auto-generated API documentation from Protocol Buffer (protobuf) definitions.

## Overview

The API documentation is automatically generated using `protoc-gen-doc` from all `.proto` files across the backend v2 project modules.

## Documentation Files

The following HTML documentation files are available:

### Core Modules
- **[index.html](./index.html)** - Comprehensive index of all APIs
- **[packages.html](./packages.html)** - Shared packages and common types
- **[core.html](./core.html)** - Core functionality and utilities

### Identity & Authentication
- **[auth.html](./auth.html)** - Authentication services
- **[user.html](./user.html)** - User management and permissions
- **[tenant.html](./tenant.html)** - Tenant/organization management
- **[entity.html](./entity.html)** - Entity management

### Human Resources
- **[employee.html](./employee.html)** - Employee management
- **[contractors.html](./contractors.html)** - Contractor management
- **[vendors.html](./vendors.html)** - Vendor management

### Organizational Structure
- **[organization.html](./organization.html)** - Division, branch, and department management

### Document Management
- **[dms.html](./dms.html)** - Document management system

### Data Management
- **[databridge.html](./databridge.html)** - Data bridging services
- **[dataarchive.html](./dataarchive.html)** - Data archival services
- **[backupdr.html](./backupdr.html)** - Backup and disaster recovery

### Analytics & Insights
- **[insighthub.html](./insighthub.html)** - Insights hub services
- **[insightviewer.html](./insightviewer.html)** - Insights viewer
- **[metasearch.html](./metasearch.html)** - Meta search services

### Forms & Workflows
- **[formbuilder.html](./formbuilder.html)** - Form builder and workflow management

### Operations
- **[scheduler.html](./scheduler.html)** - Scheduling services
- **[notification.html](./notification.html)** - Notification services
- **[masters.html](./masters.html)** - Master data management
- **[pipeline.html](./pipeline.html)** - Pipeline management

### Projects
- **[projects.html](./projects.html)** - Project management

## Generating Documentation

### Prerequisites

1. **protoc-gen-doc** must be installed:
   ```bash
   go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@latest
   ```

2. **protoc** (Protocol Buffer Compiler) must be installed

### Generate All Documentation

To regenerate all API documentation:

```bash
make docs-api
```

Or to generate all documentation (includes all types):

```bash
make docs
```

### Clean Documentation

To remove all generated HTML files:

```bash
make docs-clean
```

### Serve Documentation Locally

To view the documentation in a browser:

```bash
make docs-serve
```

This will start a local HTTP server at `http://localhost:8000/docs/api/`

## Documentation Structure

Each module's documentation includes:

- **Service Definitions**: RPC methods and their request/response types
- **Message Types**: Data structures and their fields
- **Enumerations**: Enumerated types and their values
- **Field Documentation**: Inline comments from proto files

## Updating Documentation

The documentation is automatically generated from `.proto` files. To update:

1. Modify the relevant `.proto` file(s)
2. Add/update comments in the proto file (these become documentation)
3. Run `make docs-api` to regenerate the documentation

## Best Practices

1. **Always document proto files**: Add comments to services, messages, fields, and enums
2. **Keep comments clear and concise**: Documentation is generated from these comments
3. **Regenerate after changes**: Run `make docs-api` after any proto file changes
4. **Review generated docs**: Check the HTML output to ensure clarity

## Makefile Targets

| Target | Description |
|--------|-------------|
| `make docs-api` | Generate API documentation from all proto files |
| `make docs` | Generate all documentation |
| `make docs-clean` | Remove all generated HTML files |
| `make docs-serve` | Serve documentation locally on port 8000 |

## Architecture

The documentation generation process:

1. Scans all `.proto` files in the project
2. Processes them with `protoc` and `protoc-gen-doc`
3. Generates HTML documentation with cross-references
4. Outputs to `docs/api/` directory

## Additional Resources

- [protoc-gen-doc GitHub](https://github.com/pseudomuto/protoc-gen-doc)
- [Protocol Buffers Documentation](https://protobuf.dev/)
- [gRPC Documentation](https://grpc.io/docs/)
