# UGCL Backend v2

Modern, modular backend platform for UGCL built with Go, gRPC/Connect, and PostgreSQL.

## 🚀 Quick Start

```bash
# Clone and setup
git clone <repository-url>
cd backend/v2

# Start dependencies
docker-compose up -d postgres redis minio

# Run migrations
make migrate-up

# Start application
go run cmd/main.go
```

Visit [Getting Started Guide](docs/getting-started.md) for detailed setup instructions.

## 📚 Documentation

### For New Developers
- **[Getting Started](docs/getting-started.md)** - Setup and first steps
- **[Development Guide](docs/guides/development-guide.md)** - Coding standards and best practices
- **[Proto-First Development](docs/guides/proto-first-development.md)** - API-first workflow

### Architecture & Design
- **[System Overview](docs/architecture/system-overview.md)** - High-level architecture
- **[Database Design](docs/architecture/database-design.md)** - Schema and relationships
- **[Entity Identity Model](docs/architecture/entity-identity-model.md)** - Universal entity abstraction
- **[Inter-Module Communication](docs/architecture/inter-module-communication.md)** - Service dependencies

### Deployment & Operations
- **[Deployment Guide](docs/guides/deployment-guide.md)** - Production deployment
- **[Testing Guide](docs/guides/testing-guide.md)** - Testing strategies

### Code Examples
- **[Create Organization Structure](docs/examples/create-division-branch-department.md)**
- **[Upload Document to DMS](docs/examples/upload-document-to-dms.md)**
- **[Create Employee with Entity](docs/examples/create-employee-with-entity.md)**
- **[Permission Scoping](docs/examples/check-permissions-with-scope.md)**

## 🏗️ Architecture

UGCL follows a **modular monolith** architecture with proto-first API design:

```
┌─────────────────────────────────────────┐
│         Identity & Auth Layer           │
│  (user, auth, tenant, entity)          │
└─────────────────┬───────────────────────┘
                  │
    ┌─────────────┼─────────────┐
    │             │             │
    ▼             ▼             ▼
┌─────────┐  ┌─────────┐  ┌─────────┐
│  Org &  │  │Document │  │Personnel│
│Projects │  │   Mgmt  │  │  Mgmt   │
└─────────┘  └─────────┘  └─────────┘
```

## 📦 Modules

### Core Infrastructure
- **[cmd](cmd/)** - Application entry point and bootstrap
- **[core](core/)** - Core business logic and middleware
- **[packages](packages/)** - Shared utilities and infrastructure

### Identity & Access
- **[identity/auth](identity/auth/)** - Authentication and sessions
- **[identity/user](identity/user/)** - User management
- **[identity/tenant](identity/tenant/)** - Multi-tenancy
- **[identity/entity](identity/entity/)** - Entity abstraction

### Business Modules
- **[formbuilder](formbuilder/)** - Dynamic form creation
- **[organization](organization/)** - Organizational hierarchy
- **[dms](dms/)** - Document management system
- **[vendors](vendors/)** - Vendor management
- **[contractors](contractors/)** - Contractor management
- **[employee](employee/)** - Employee management
- **[projects](projects/)** - Project management

### Support Services
- **[notification](notification/)** - Notification system
- **[scheduler](scheduler/)** - Job scheduling
- **[monitoring](monitoring/)** - System monitoring
- **[metasearch](metasearch/)** - Cross-module search

### Data Services
- **[databridge](databridge/)** - Data import/export
- **[dataarchive](dataarchive/)** - Data archiving
- **[backupdr](backupdr/)** - Backup & disaster recovery

### Analytics
- **[insighthub](insighthub/)** - Analytics engine
- **[insightviewer](insightviewer/)** - Report viewer

## 🛠️ Tech Stack

- **Language**: Go 1.25+
- **API**: Protocol Buffers + Connect-RPC
- **Database**: PostgreSQL 15+ with SQLC
- **Cache**: Redis 7+
- **Storage**: MinIO (S3-compatible)
- **DI**: Uber FX
- **Migrations**: Atlas
- **Testing**: Go testing + Testify

## 🔧 Development

### Prerequisites
- Go 1.25+
- Docker & Docker Compose
- PostgreSQL 15+
- Make

### Common Commands

```bash
# Development
make run                # Run application
make test               # Run tests
make lint               # Run linter
make fmt                # Format code

# Code Generation
make generate           # Generate proto + SQLC
make proto-generate     # Generate proto only
make sqlc-generate      # Generate SQLC only

# Database
make migrate-up         # Apply migrations
make migrate-down       # Rollback migrations

# Docker
make docker-build       # Build image
make docker-up          # Start services
```

## 📋 Project Status

**Current Version**: 2.0.0  
**Status**: Active Development

### Recent Updates
- ✅ Proto-first API development
- ✅ Entity-based identity system
- ✅ Module documentation complete
- ✅ SQLC code generation
- ✅ FX dependency injection

### Upcoming
- [ ] Auto-generated API documentation
- [ ] CI/CD pipeline setup
- [ ] Performance testing
- [ ] Production deployment

## 📖 Additional Resources

- **[Master TODO List](docs/TODO-MASTER-LIST.md)** - Project roadmap
- **[Technical Specifications](docs/technical-specifications.md)** - Detailed specs
- **[Implementation Roadmap](docs/implementation-roadmap.md)** - Development phases

## 🤝 Contributing

1. Read the [Development Guide](docs/guides/development-guide.md)
2. Create a feature branch: `git checkout -b feature/UGCL-123-description`
3. Make your changes and add tests
4. Run checks: `make lint && make test`
5. Commit: `git commit -m "feat(module): description"`
6. Push and create a Pull Request

## 📄 License

[Add license information]

## 📞 Support

- **Documentation**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/your-org/ugcl/issues)
- **Discussions**: [GitHub Discussions](https://github.com/your-org/ugcl/discussions)

---

**Built with ❤️ by the UGCL Team**
