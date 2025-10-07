# UGCL Backend v2 Development Guide

## Table of Contents

1. [Introduction](#introduction)
2. [Development Environment Setup](#development-environment-setup)
3. [Project Structure Walkthrough](#project-structure-walkthrough)
4. [Coding Standards and Conventions](#coding-standards-and-conventions)
5. [Git Workflow and Branching Strategy](#git-workflow-and-branching-strategy)
6. [Code Review Checklist](#code-review-checklist)
7. [Local Development with Docker](#local-development-with-docker)
8. [Environment Variables and Configuration](#environment-variables-and-configuration)
9. [Common Development Tasks](#common-development-tasks)
10. [Debugging Techniques](#debugging-techniques)
11. [Performance Profiling](#performance-profiling)
12. [IDE Setup](#ide-setup)

---

## Introduction

Welcome to the UGCL Backend v2 development guide. This guide provides comprehensive instructions for setting up your development environment, understanding the project structure, and following best practices for developing microservices in the UGCL ecosystem.

**Technology Stack:**
- **Language:** Go 1.25.0
- **API Framework:** Connect (gRPC-compatible HTTP/JSON)
- **Database:** PostgreSQL 15+
- **ORM:** SQLC (SQL-first type-safe code generation)
- **Protocol Buffers:** Buf (protobuf management)
- **Dependency Injection:** Uber FX
- **Migration Tool:** Atlas CLI
- **Testing:** Go testing + testcontainers

**Architecture Pattern:**
- Proto-first development
- Repository → Service → Handler layered architecture
- Multi-module monorepo with Go workspaces
- Microservices with shared packages

---

## Development Environment Setup

### Prerequisites

Before you begin, ensure you have the following installed:

#### 1. Go 1.25.0+

```bash
# Download and install Go
# Visit: https://go.dev/dl/

# Verify installation
go version
# Expected: go version go1.25.0 or later
```

#### 2. PostgreSQL 15+

```bash
# Install PostgreSQL
# Windows: Download from https://www.postgresql.org/download/windows/
# macOS: brew install postgresql@15
# Linux: sudo apt install postgresql-15

# Verify installation
psql --version
# Expected: psql (PostgreSQL) 15.x

# Start PostgreSQL service
# Windows: Start from Services
# macOS/Linux: brew services start postgresql@15 or sudo systemctl start postgresql
```

#### 3. Protocol Buffer Tools

```bash
# Install Buf CLI
# Windows (PowerShell as Admin):
Invoke-WebRequest -Uri https://github.com/bufbuild/buf/releases/latest/download/buf-Windows-x86_64.exe -OutFile buf.exe
Move-Item buf.exe C:\Windows\System32\buf.exe

# macOS:
brew install bufbuild/buf/buf

# Linux:
curl -sSL "https://github.com/bufbuild/buf/releases/download/v1.28.1/buf-Linux-x86_64" -o /usr/local/bin/buf
chmod +x /usr/local/bin/buf

# Verify installation
buf --version
```

#### 4. SQLC

```bash
# Install SQLC
# Windows (using Go):
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# macOS:
brew install sqlc

# Linux:
curl -L https://github.com/sqlc-dev/sqlc/releases/download/v1.25.0/sqlc_1.25.0_linux_amd64.tar.gz | tar -xz
sudo mv sqlc /usr/local/bin/

# Verify installation
sqlc version
```

#### 5. Atlas CLI (Database Migration Tool)

```bash
# Install Atlas
# Windows (PowerShell):
Invoke-WebRequest -Uri https://release.ariga.io/atlas/atlas-windows-amd64-latest.exe -OutFile atlas.exe
Move-Item atlas.exe C:\Windows\System32\atlas.exe

# macOS:
brew install ariga/tap/atlas

# Linux:
curl -sSf https://atlasgo.sh | sh

# Verify installation
atlas version
```

#### 6. Docker Desktop (Optional but Recommended)

```bash
# Download and install Docker Desktop
# Visit: https://www.docker.com/products/docker-desktop

# Verify installation
docker --version
docker-compose --version
```

#### 7. Git

```bash
# Install Git
# Visit: https://git-scm.com/downloads

# Configure Git
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"
```

### Clone the Repository

```bash
# Clone the repository
git clone https://github.com/your-org/ugcl-backend-v2.git
cd ugcl-backend-v2

# Checkout the v2 branch
git checkout v2

# Verify Go workspace
cat go.work
```

### Initialize Go Modules

```bash
# Download all dependencies
go work sync

# Tidy all modules
find . -name "go.mod" -execdir go mod tidy \;

# Verify no errors
go work vendor
```

### Database Setup

```bash
# Create local database
createdb ugcl_dev

# Or using psql
psql -U postgres
CREATE DATABASE ugcl_dev;
\q

# Update configs.yaml with your local database
# Edit: data.postgres.host, data.postgres.dbname, etc.
```

### Environment Variables

Create a `.env` file in the project root:

```bash
# Copy the example
cp .env.example .env

# Edit .env with your values
JWT_SECRET=hrZnUNWi3Xh35FG1RZThBkfPrQx0Jhw/RlQhi3RtCtc=
DB_DSN=host=localhost port=5432 user=postgres password=yourpassword dbname=ugcl_dev sslmode=disable search_path=public
MOBILE_APP_KEY=060cf2a4-9842-44b2-a362-20aab5a0879a
PARTNER_PORTAL_KEY=9ecc7d18-d528-4fdf-b237-d4175c4db175
INTERNAL_OPS_KEY=87339ea3-1add-4689-ae57-3128ebd03c4f
```

### Verify Setup

```bash
# Run development setup
make dev-setup

# Expected output:
# ✓ Linting protobuf files...
# ✓ Validating migrations...
# ✓ Generating SQLC code for all modules...
# ✓ Development setup complete
```

---

## Project Structure Walkthrough

The UGCL Backend v2 follows a **multi-module monorepo** structure with Go workspaces:

```
backend/v2/
├── go.work                    # Go workspace file (lists all modules)
├── Makefile                   # Build and development tasks
├── buf.yaml                   # Buf protobuf configuration
├── buf.gen.yaml              # Buf code generation config
├── configs.yaml              # Application configuration
├── Dockerfile                # Container image definition
│
├── cmd/                      # Application entry point
│   ├── main.go              # Main application
│   ├── app_builder.go       # FX dependency injection setup
│   └── module_registry.go   # Service registration
│
├── core/                     # Shared core packages
│   ├── config/              # Configuration management
│   ├── middleware/          # HTTP/gRPC middleware
│   └── proto/               # Core protobuf definitions
│
├── packages/                 # Shared packages
│   ├── database/            # Database connection management
│   ├── config/              # Configuration loaders
│   ├── p9log/               # Logging utilities
│   └── proto/               # Common proto definitions
│
├── identity/                 # Identity module group
│   ├── auth/                # Authentication service
│   │   ├── proto/          # Auth protobuf definitions
│   │   ├── handlers/       # gRPC/Connect handlers
│   │   ├── services/       # Business logic
│   │   ├── repository/     # Data access layer
│   │   ├── db/             # Database schemas and queries
│   │   │   ├── schema/    # SQL schema files
│   │   │   ├── queries/   # SQLC query files
│   │   │   ├── generated/ # SQLC generated code
│   │   │   └── sqlc.yaml  # SQLC configuration
│   │   ├── mappers/        # Proto ↔ DB model conversions
│   │   ├── module.go       # FX module definition
│   │   └── go.mod          # Module dependencies
│   │
│   ├── user/                # User management service
│   ├── tenant/              # Tenant management service
│   └── entity/              # Entity management service
│
├── vendors/                  # Vendor management module
│   ├── proto/
│   │   └── vendor.proto    # Vendor service definition
│   ├── handlers/
│   │   └── vendor_handler.go
│   ├── services/
│   │   └── vendor_service.go
│   ├── repository/
│   │   └── vendor_repository.go
│   ├── db/
│   │   ├── schema/
│   │   │   └── vendors.sql
│   │   ├── queries/
│   │   │   └── vendors.sql
│   │   ├── generated/      # SQLC generated code
│   │   └── sqlc.yaml
│   ├── mappers/
│   │   └── vendor_mapper.go
│   ├── module.go           # FX module
│   └── go.mod
│
├── formbuilder/             # Form builder module
├── projects/                # Project management module
├── organization/            # Organization hierarchy module
├── dms/                     # Document management system
├── notification/            # Notification service
├── scheduler/               # Job scheduler
├── backupdr/               # Backup & disaster recovery
├── dataarchive/            # Data archiving service
├── metasearch/             # Search service
│
├── migrations/              # Database migrations
│   ├── versions/           # Migration files
│   └── aggregated_schema.sql
│
└── docs/                    # Documentation
    ├── api/                # Generated API docs
    ├── guides/             # Developer guides
    └── architecture/       # Architecture diagrams
```

### Module Structure Pattern

Each module follows this standard structure:

```
module-name/
├── proto/                   # Protocol buffer definitions
│   └── module.proto
├── api/v1/                 # Generated protobuf code
│   └── modulev1connect/
├── handlers/               # gRPC/Connect request handlers
│   └── module_handler.go
├── services/               # Business logic layer
│   └── module_service.go
│   └── iservices.go       # Service interfaces
├── repository/             # Data access layer
│   └── module_repository.go
├── db/
│   ├── schema/            # SQL schema definitions
│   │   └── schema.sql
│   ├── queries/           # SQLC query definitions
│   │   └── queries.sql
│   ├── generated/         # SQLC generated code
│   │   ├── models.go
│   │   ├── querier.go
│   │   └── db.go
│   └── sqlc.yaml          # SQLC configuration
├── mappers/               # DTO transformations
│   └── module_mapper.go
├── models/                # Domain models (if needed)
│   └── module_models.go
├── module.go              # FX module definition
├── go.mod                 # Go module file
└── README.md              # Module documentation
```

---

## Coding Standards and Conventions

### Go Code Style

Follow the official [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments):

#### Naming Conventions

```go
// ✅ Good: Exported types use PascalCase
type VendorService struct {
    repo IVendorRepository
}

// ✅ Good: Unexported types use camelCase
type vendorCache struct {
    data map[string]*Vendor
}

// ✅ Good: Interface names
type IVendorRepository interface {
    Create(ctx context.Context, vendor *Vendor) error
}

// ❌ Bad: Don't use underscores
type vendor_service struct {} // Wrong

// ✅ Good: Constants
const (
    MaxPageSize = 100
    DefaultTimeout = 30 * time.Second
)

// ✅ Good: Acronyms should be uppercase
type HTTPAPI struct {} // Good
type HttpApi struct {} // Bad
```

#### Package Names

```go
// ✅ Good: Short, lowercase, no underscores
package vendor
package handler
package repository

// ❌ Bad: Verbose or mixed case
package vendor_management
package VendorService
```

#### Error Handling

```go
// ✅ Good: Wrap errors with context
func (s *VendorService) GetByID(ctx context.Context, id string) (*Vendor, error) {
    vendor, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get vendor %s: %w", id, err)
    }
    return vendor, nil
}

// ✅ Good: Use custom error types for specific errors
var ErrVendorNotFound = errors.New("vendor not found")

// ✅ Good: Check errors early
if err != nil {
    return fmt.Errorf("error context: %w", err)
}

// ❌ Bad: Ignoring errors
vendor, _ := s.repo.GetByID(ctx, id) // Never do this
```

#### Function Comments

```go
// ✅ Good: Document exported functions
// CreateVendor creates a new vendor in the system.
// It first creates a user account and then associates the vendor with that user.
// Returns an error if the user creation fails or if a vendor with the same PAN already exists.
func (s *VendorService) CreateVendor(ctx context.Context, vendor *Vendor) (*Vendor, error) {
    // Implementation
}
```

#### Context Usage

```go
// ✅ Good: Always pass context as first parameter
func (s *VendorService) Create(ctx context.Context, vendor *Vendor) error {
    // Use ctx for cancellation, timeouts, and tracing
}

// ✅ Good: Don't store context in structs
type VendorService struct {
    repo IVendorRepository
    // Don't add: ctx context.Context
}

// ✅ Good: Create timeout contexts for operations
func (s *VendorService) LongOperation(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    return s.repo.ExpensiveQuery(ctx)
}
```

### Protobuf Conventions

#### Message Naming

```protobuf
// ✅ Good: Use PascalCase for messages
message Vendor {
  string id = 1;
  string company_name = 2;
}

// ✅ Good: Request/Response suffix
message CreateVendorRequest {
  Vendor vendor = 1;
}

message CreateVendorResponse {
  Vendor vendor = 1;
}

// ✅ Good: Use descriptive field names (snake_case)
message Vendor {
  string company_name = 1;  // Good
  string CompanyName = 2;   // Bad
  string cname = 3;         // Bad - too abbreviated
}
```

#### Service Definitions

```protobuf
// ✅ Good: Service name matches domain
service VendorService {
  // Use verb + noun pattern
  rpc CreateVendor(CreateVendorRequest) returns (Vendor) {}
  rpc GetVendor(VendorIdentifier) returns (Vendor) {}
  rpc ListVendors(ListVendorsRequest) returns (ListVendorsResponse) {}
  rpc UpdateVendor(UpdateVendorRequest) returns (Vendor) {}
  rpc DeleteVendor(VendorIdentifier) returns (google.protobuf.Empty) {}
}
```

#### Field Numbering

```protobuf
// ✅ Good: Reserve field numbers for deleted fields
message Vendor {
  reserved 2, 15, 9 to 11;
  reserved "middle_name", "old_status";

  string id = 1;
  string company_name = 3;
  // Fields 2, 9-11, 15 are reserved
}

// ✅ Good: Use logical grouping (1-15 for common, 16+ for optional)
message Vendor {
  // Core fields (1-15 = 1 byte encoding)
  string id = 1;
  string company_name = 2;
  string status = 3;

  // Optional/extended fields (16+)
  google.protobuf.Struct metadata = 16;
  repeated string tags = 17;
}
```

### SQL Conventions (SQLC)

#### Query Naming

```sql
-- ✅ Good: Descriptive query names with action prefix
-- name: CreateVendor :one
INSERT INTO vendors (company_name, gst, pan)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetVendorByID :one
SELECT * FROM vendors WHERE id = $1 AND deleted_at IS NULL;

-- name: ListVendors :many
SELECT * FROM vendors WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateVendor :one
UPDATE vendors SET company_name = $1 WHERE id = $2 RETURNING *;

-- name: DeleteVendor :exec
UPDATE vendors SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1;

-- name: CountVendors :one
SELECT COUNT(*) FROM vendors WHERE deleted_at IS NULL;
```

#### Schema Conventions

```sql
-- ✅ Good: Use snake_case for table and column names
CREATE TABLE vendors (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    uuid UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    company_name TEXT NOT NULL,
    vendor_category TEXT,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE, -- Soft delete

    -- Indexes
    CONSTRAINT vendors_company_name_check CHECK (length(company_name) > 0)
);

-- ✅ Good: Create indexes for foreign keys and frequently queried columns
CREATE INDEX idx_vendors_person_id ON vendors(person_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_vendors_category ON vendors(vendor_category) WHERE deleted_at IS NULL;
```

### Repository Pattern

```go
// ✅ Good: Define interface for testability
type IVendorRepository interface {
    Create(ctx context.Context, params db.CreateVendorParams) (*db.Vendor, error)
    GetByID(ctx context.Context, id uuid.UUID) (*db.Vendor, error)
    Update(ctx context.Context, params db.UpdateVendorParams) (*db.Vendor, error)
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, params db.ListVendorsParams) ([]db.Vendor, int64, error)
}

// Implementation
type VendorRepository struct {
    queries *db.Queries
}

func NewVendorRepository(q *db.Queries) IVendorRepository {
    return &VendorRepository{queries: q}
}

func (r *VendorRepository) Create(ctx context.Context, params db.CreateVendorParams) (*db.Vendor, error) {
    vendor, err := r.queries.CreateVendor(ctx, params)
    if err != nil {
        // ✅ Good: Check for database-specific errors
        if pqErr, ok := err.(*pq.Error); ok {
            if pqErr.Code == "23505" { // unique_violation
                return nil, fmt.Errorf("vendor with PAN already exists")
            }
        }
        return nil, fmt.Errorf("failed to create vendor: %w", err)
    }
    return &vendor, nil
}
```

### Service Layer Pattern

```go
// ✅ Good: Service interface for business logic
type IVendorService interface {
    Create(ctx context.Context, vendor *db.Vendor, user *userpb.User) (*db.Vendor, error)
    GetByID(ctx context.Context, id string) (*db.Vendor, error)
    Update(ctx context.Context, vendor *db.Vendor) (*db.Vendor, error)
    Delete(ctx context.Context, id string) error
}

type VendorService struct {
    repo    IVendorRepository
    userSvc IUserService
}

func NewVendorService(repo IVendorRepository, userSvc IUserService) IVendorService {
    return &VendorService{
        repo:    repo,
        userSvc: userSvc,
    }
}

// ✅ Good: Service contains business logic, not just CRUD
func (s *VendorService) Create(ctx context.Context, vendor *db.Vendor, user *userpb.User) (*db.Vendor, error) {
    // Step 1: Create user first
    userResp, err := s.userSvc.RegisterUser(ctx, user)
    if err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }

    // Step 2: Associate vendor with user
    vendor.PersonID = userResp.Uuid.String()

    // Step 3: Create vendor
    return s.repo.Create(ctx, db.CreateVendorParams{
        CompanyName:    vendor.CompanyName,
        PersonID:       vendor.PersonID,
        // ... other fields
    })
}
```

### Handler Pattern (Connect)

```go
// ✅ Good: Handler delegates to service
type VendorHandler struct {
    srvc IVendorService
}

func NewVendorHandler(srvc IVendorService) *VendorHandler {
    return &VendorHandler{srvc: srvc}
}

func (h *VendorHandler) CreateVendor(
    ctx context.Context,
    req *connect.Request[pb.CreateVendorRequest],
) (*connect.Response[pb.Vendor], error) {
    // Validate input
    if req.Msg.Vendor.Id == "" {
        req.Msg.Vendor.Id = uuid.New().String()
    }

    // Map proto to DB model
    dbVendor, err := mappers.ProtoToDBVendor(req.Msg.Vendor)
    if err != nil {
        return nil, connect.NewError(connect.CodeInvalidArgument, err)
    }

    // Call service
    created, err := h.srvc.Create(ctx, dbVendor, req.Msg.User)
    if err != nil {
        return nil, connect.NewError(connect.CodeInternal, err)
    }

    // Map back to proto
    protoVendor, err := mappers.DBToProtoVendor(created)
    if err != nil {
        return nil, connect.NewError(connect.CodeInternal, err)
    }

    return connect.NewResponse(protoVendor), nil
}
```

---

## Git Workflow and Branching Strategy

### Branch Naming

```bash
# Feature branches
feature/vendor-blacklist-management
feature/approval-workflow-integration

# Bug fixes
bugfix/vendor-creation-validation
bugfix/null-pointer-in-search

# Hotfixes (production issues)
hotfix/critical-auth-bypass

# Release branches
release/v2.1.0

# Experimental
experiment/new-caching-strategy
```

### Commit Message Format

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```bash
# Format
<type>(<scope>): <subject>

<body>

<footer>

# Types
feat:     # New feature
fix:      # Bug fix
docs:     # Documentation changes
style:    # Code style changes (formatting, semicolons, etc)
refactor: # Code refactoring
test:     # Adding tests
chore:    # Maintenance tasks (dependencies, build, etc)
perf:     # Performance improvements

# Examples
feat(vendor): add blacklist management feature

Implement vendor blacklisting with reason tracking and audit logs.
- Add is_blacklisted field to vendor model
- Create blacklist service with approval workflow
- Add blacklist history table

Closes #123

---

fix(auth): prevent null pointer exception in JWT validation

The JWT validation middleware was not checking for nil claims
before accessing user_id field.

Fixes #456

---

docs(guides): add comprehensive testing guide

Created testing-guide.md with:
- Unit testing patterns
- Integration testing with testcontainers
- Mocking strategies
- Code coverage requirements
```

### Development Workflow

```bash
# 1. Create a new branch from main
git checkout main
git pull origin main
git checkout -b feature/vendor-rating-system

# 2. Make changes and commit frequently
git add .
git commit -m "feat(vendor): add rating field to vendor model"

# 3. Keep your branch updated
git fetch origin
git rebase origin/main

# 4. Push to remote
git push origin feature/vendor-rating-system

# 5. Create pull request
# Use GitHub UI or gh CLI:
gh pr create --title "Add vendor rating system" --body "Implements #123"

# 6. After review and approval, merge using squash or rebase
# (Done via GitHub UI with "Squash and merge")

# 7. Clean up local branch
git checkout main
git pull origin main
git branch -d feature/vendor-rating-system
```

### Pull Request Best Practices

```markdown
# Pull Request Template

## Description
Brief description of what this PR does.

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Changes Made
- List key changes
- Use bullet points
- Be specific

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing completed
- [ ] Code coverage > 80%

## Database Changes
- [ ] Schema migration added
- [ ] SQLC queries updated
- [ ] Indexes added

## Proto Changes
- [ ] No breaking changes
- [ ] Backward compatible
- [ ] Version incremented if needed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex logic
- [ ] Documentation updated
- [ ] No new warnings generated
- [ ] Tests pass locally
- [ ] Dependent changes merged

## Screenshots (if applicable)

## Related Issues
Closes #123
Related to #456
```

---

## Code Review Checklist

### For Authors (Before Submitting PR)

- [ ] **Self-review:** Read your own code as if you were the reviewer
- [ ] **Tests:** All tests pass locally (`go test ./...`)
- [ ] **Coverage:** Code coverage is above 80% for new code
- [ ] **Linting:** No linter errors (`make proto-lint`, `go vet ./...`)
- [ ] **Documentation:** Updated README or added comments for complex logic
- [ ] **Proto changes:** Ran `buf breaking` to check for breaking changes
- [ ] **SQLC:** Generated code is committed (`make sqlc-generate-all`)
- [ ] **No debug code:** Removed console.log, commented code, TODOs
- [ ] **Security:** No hardcoded secrets, passwords, or API keys
- [ ] **Error handling:** All errors are properly handled and wrapped
- [ ] **Context:** Context is passed to all repository/service calls

### For Reviewers

#### Code Quality

- [ ] Code is readable and maintainable
- [ ] Naming conventions are followed
- [ ] Functions are small and focused (< 50 lines ideally)
- [ ] No code duplication (DRY principle)
- [ ] No overly complex logic (cyclomatic complexity < 10)

#### Architecture

- [ ] Follows Repository → Service → Handler pattern
- [ ] Proper layer separation (no DB calls in handlers)
- [ ] Dependencies injected via FX module
- [ ] Interfaces used for testability

#### Error Handling

- [ ] Errors are wrapped with context
- [ ] Database errors are handled (unique violations, not found, etc.)
- [ ] Connect error codes are appropriate
- [ ] No silent error ignoring

#### Testing

- [ ] Unit tests for business logic
- [ ] Integration tests for database operations
- [ ] Edge cases covered
- [ ] Mocks used appropriately

#### Performance

- [ ] No N+1 queries
- [ ] Database indexes exist for foreign keys
- [ ] Pagination implemented for list operations
- [ ] Context timeouts set for long operations

#### Security

- [ ] Input validation in handlers
- [ ] SQL injection prevented (parameterized queries)
- [ ] Authentication/authorization checked
- [ ] Sensitive data not logged

#### Database

- [ ] Migrations are reversible
- [ ] SQLC queries are efficient
- [ ] Soft deletes used where appropriate
- [ ] Transactions used for multi-step operations

---

## Local Development with Docker

### Docker Compose Setup

Create `docker-compose.yml` in project root:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: ugcl-postgres
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: ugcl_dev
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: ugcl-redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  backend:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: ugcl-backend
    environment:
      - DB_DSN=host=postgres port=5432 user=postgres password=postgres dbname=ugcl_dev sslmode=disable
      - REDIS_ADDR=redis:6379
      - JWT_SECRET=${JWT_SECRET}
    ports:
      - "10011:10011"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_started
    volumes:
      - .:/app

volumes:
  postgres_data:
  redis_data:
```

### Docker Commands

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f backend

# Stop all services
docker-compose down

# Rebuild backend after code changes
docker-compose up -d --build backend

# Execute commands in container
docker-compose exec backend go test ./...

# Access PostgreSQL
docker-compose exec postgres psql -U postgres -d ugcl_dev

# Clean up everything
docker-compose down -v
```

### Development Dockerfile

```dockerfile
# Multi-stage build for development
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# Copy go.work and module files
COPY go.work go.work.sum ./
COPY */go.mod */go.sum ./

# Download dependencies
RUN go work sync

# Copy source code
COPY . .

# Build the application
RUN go build -o /ugcl-server ./cmd

# Production image
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /ugcl-server .
COPY --from=builder /app/configs.yaml .

# Expose ports
EXPOSE 10011

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:10011/health || exit 1

CMD ["./ugcl-server"]
```

---

## Environment Variables and Configuration

### Configuration Management

UGCL uses a hierarchical configuration system:

1. **configs.yaml** - Default configuration
2. **.env file** - Local overrides (not committed)
3. **Environment variables** - Runtime overrides (highest priority)

### configs.yaml Structure

```yaml
# JWT Configuration
jwt:
  secret: ${JWT_SECRET}  # Will be replaced by env var

# API Keys for third-party access
api_keys:
  "mobile-app-key":
    app_name: "MobileApp"
    allowed_paths: ["/p9e.ugcl.identity.api.v2.user.User/*"]
    allowed_methods:
      POST: true
    skip_ip_check: true

# Server Configuration
server:
  http:
    network: tcp
    addr: 0.0.0.0:10011
    port: 10011
    timeout: 600s
  grpc:
    network: tcp
    addr: 0.0.0.0:10012
    port: 10012
    keepalive_time: 60s
    keepalive_timeout: 20s

# Database Configuration
data:
  postgres:
    host: ${DB_HOST:localhost}
    port: ${DB_PORT:5432}
    user: ${DB_USER:postgres}
    password: ${DB_PASSWORD:postgres}
    dbname: ${DB_NAME:ugcl_dev}
    sslmode: ${DB_SSLMODE:disable}
    search_path: ${DB_SCHEMA:public}
    max_connections: 50
    min_connections: 5
    max_connection_lifetime: 1h
    connection_timeout: 30s

  redis:
    addr: ${REDIS_ADDR:localhost:6379}
    password: ${REDIS_PASSWORD:}
    db: 0
    read_timeout: 0.2s
    write_timeout: 0.2s

# Observability
observability:
  tracing:
    provider: jaeger
    endpoint: ${JAEGER_ENDPOINT:http://localhost:14268/api/traces}
    sampling_rate: 0.5
    enabled: true

  metrics:
    provider: prometheus
    endpoint: ${METRICS_ENDPOINT:http://localhost:9090}
    enabled: true
    collection_interval: 15s
```

### Environment Variables Reference

```bash
# Application
PORT=10011
ENV=development  # development, staging, production

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=secure_password
DB_NAME=ugcl_dev
DB_SSLMODE=disable
DB_SCHEMA=public
DB_DSN="host=localhost port=5432 user=postgres password=secure_password dbname=ugcl_dev sslmode=disable"

# Authentication
JWT_SECRET=your_jwt_secret_key_here
JWT_EXPIRY=3600  # seconds

# API Keys
MOBILE_APP_KEY=mobile_secret_key
PARTNER_PORTAL_KEY=partner_secret_key
INTERNAL_OPS_KEY=internal_secret_key

# Redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=

# Observability
JAEGER_ENDPOINT=http://localhost:14268/api/traces
METRICS_ENDPOINT=http://localhost:9090
LOG_LEVEL=debug  # debug, info, warn, error

# External Services
USER_SERVICE_URL=http://localhost:10011
NOTIFICATION_SERVICE_URL=http://localhost:10012
```

---

## Common Development Tasks

### Generate Protobuf Code

```bash
# Generate all protobuf code
make proto-generate

# Or manually with buf
buf generate

# Lint protobuf files
make proto-lint
buf lint

# Check for breaking changes
make proto-breaking
buf breaking --against '.git#branch=main'
```

### Generate SQLC Code

```bash
# Generate all modules
make sqlc-generate-all

# Generate specific module
make sqlc-vendors
cd vendors/db && sqlc generate

# Verify SQLC configuration
cd vendors/db && sqlc verify
```

### Database Migrations

```bash
# Aggregate schemas from all modules
make migrate-aggregate

# Generate new migration
make migrate-generate NAME=add_vendor_rating

# Apply migrations
make migrate-apply

# Check migration status
make migrate-status

# Validate migrations
make migrate-validate
```

### Run the Application

```bash
# Build and run
go run ./cmd

# Or using Makefile
make run

# With hot reload (using air)
air

# Build binary
go build -o bin/ugcl-server ./cmd

# Run binary
./bin/ugcl-server
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run tests for specific module
go test ./vendors/...

# Run specific test
go test -run TestCreateVendor ./vendors/services

# Run tests with race detector
go test -race ./...

# Verbose output
go test -v ./...
```

### Linting and Formatting

```bash
# Format code
go fmt ./...

# Run go vet
go vet ./...

# Run golangci-lint (install first: https://golangci-lint.run/usage/install/)
golangci-lint run

# Auto-fix issues
golangci-lint run --fix
```

### Dependency Management

```bash
# Add new dependency
go get github.com/some/package

# Update dependencies
go get -u ./...

# Tidy modules
go mod tidy

# Vendor dependencies
go mod vendor

# Sync workspace
go work sync
```

---

## Debugging Techniques

### Delve Debugger

Install Delve:

```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

Debug application:

```bash
# Start debugger
dlv debug ./cmd

# Set breakpoint
(dlv) break main.main
(dlv) break vendors/services/vendor_service.go:42

# Continue execution
(dlv) continue

# Step through code
(dlv) next
(dlv) step

# Inspect variables
(dlv) print vendor
(dlv) print vendor.CompanyName

# List breakpoints
(dlv) breakpoints

# Clear breakpoint
(dlv) clear 1
```

### Debug with VS Code

Add `.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch Backend",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd",
      "env": {
        "DB_DSN": "host=localhost port=5432 user=postgres password=postgres dbname=ugcl_dev sslmode=disable"
      },
      "args": []
    },
    {
      "name": "Attach to Process",
      "type": "go",
      "request": "attach",
      "mode": "local",
      "processId": "${command:pickProcess}"
    },
    {
      "name": "Test Current File",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${file}"
    }
  ]
}
```

### Logging Best Practices

```go
import "p9e.in/ugcl/packages/p9log"

// Initialize logger
logger := p9log.NewStdLogger(os.Stdout)
logger = p9log.With(logger, "service", "vendor")

// Log levels
logger.Log(p9log.LevelDebug, "msg", "Vendor created", "vendor_id", vendor.ID)
logger.Log(p9log.LevelInfo, "msg", "Processing request", "user_id", userID)
logger.Log(p9log.LevelWarn, "msg", "Slow query detected", "duration", elapsed)
logger.Log(p9log.LevelError, "msg", "Failed to create vendor", "error", err)

// Structured logging
logger.Log(
    p9log.LevelInfo,
    "msg", "Vendor created successfully",
    "vendor_id", vendor.ID,
    "company_name", vendor.CompanyName,
    "created_by", userID,
    "duration_ms", elapsed.Milliseconds(),
)
```

### Database Query Debugging

```go
// Enable query logging in PostgreSQL
// Add to configs.yaml
data:
  postgres:
    log_queries: true
    log_slow_queries: true
    slow_query_threshold: 100ms

// Use EXPLAIN ANALYZE in psql
EXPLAIN ANALYZE SELECT * FROM vendors WHERE company_name ILIKE '%acme%';

// Check query performance
SELECT query, mean_exec_time, calls
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;
```

---

## Performance Profiling

### CPU Profiling

```go
import (
    "runtime/pprof"
    "os"
)

// Start CPU profiling
f, err := os.Create("cpu.prof")
if err != nil {
    log.Fatal(err)
}
defer f.Close()

pprof.StartCPUProfile(f)
defer pprof.StopCPUProfile()

// Run your code
// ...

// Analyze with:
// go tool pprof cpu.prof
```

### Memory Profiling

```go
import (
    "runtime"
    "runtime/pprof"
)

// Take heap snapshot
f, err := os.Create("mem.prof")
if err != nil {
    log.Fatal(err)
}
defer f.Close()

runtime.GC() // get up-to-date statistics
pprof.WriteHeapProfile(f)

// Analyze with:
// go tool pprof mem.prof
```

### HTTP Profiling (pprof)

```go
import _ "net/http/pprof"

// Add to main.go
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()

// Access in browser:
// http://localhost:6060/debug/pprof/
// http://localhost:6060/debug/pprof/heap
// http://localhost:6060/debug/pprof/goroutine

// Or use go tool:
go tool pprof http://localhost:6060/debug/pprof/heap
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
```

### Benchmarking

```go
// vendor_service_test.go
func BenchmarkCreateVendor(b *testing.B) {
    // Setup
    repo := &MockVendorRepository{}
    svc := NewVendorService(repo, nil)
    vendor := &Vendor{CompanyName: "Test Corp"}
    ctx := context.Background()

    // Reset timer
    b.ResetTimer()

    // Run benchmark
    for i := 0; i < b.N; i++ {
        svc.Create(ctx, vendor, nil)
    }
}

// Run benchmarks
// go test -bench=. ./vendors/services
// go test -bench=BenchmarkCreateVendor -benchmem ./vendors/services
```

---

## IDE Setup

### VS Code

#### Recommended Extensions

```json
// .vscode/extensions.json
{
  "recommendations": [
    "golang.go",                     // Go language support
    "bufbuild.vscode-buf",           // Protobuf support
    "ms-azuretools.vscode-docker",   // Docker support
    "github.copilot",                // AI pair programming
    "eamodio.gitlens",               // Git supercharged
    "redhat.vscode-yaml",            // YAML support
    "ms-vscode.makefile-tools",      // Makefile support
    "esbenp.prettier-vscode"         // Code formatter
  ]
}
```

#### VS Code Settings

```json
// .vscode/settings.json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "workspace",
  "go.formatTool": "goimports",
  "go.formatOnSave": true,
  "go.testFlags": ["-v", "-race"],
  "go.testTimeout": "10m",
  "go.coverOnSave": true,
  "go.coverageDecorator": {
    "type": "gutter"
  },
  "go.toolsManagement.autoUpdate": true,
  "[go]": {
    "editor.codeActionsOnSave": {
      "source.organizeImports": true
    }
  },
  "gopls": {
    "analyses": {
      "unusedparams": true,
      "shadow": true
    },
    "staticcheck": true,
    "usePlaceholders": true
  },
  "files.exclude": {
    "**/.git": true,
    "**/vendor": true,
    "**/*_test.go": false
  },
  "files.watcherExclude": {
    "**/vendor/**": true,
    "**/node_modules/**": true
  }
}
```

#### Tasks Configuration

```json
// .vscode/tasks.json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Generate Proto",
      "type": "shell",
      "command": "make proto-generate",
      "group": "build",
      "presentation": {
        "reveal": "always"
      }
    },
    {
      "label": "Generate SQLC",
      "type": "shell",
      "command": "make sqlc-generate-all",
      "group": "build"
    },
    {
      "label": "Run Tests",
      "type": "shell",
      "command": "go test -v ./...",
      "group": "test"
    },
    {
      "label": "Run Application",
      "type": "shell",
      "command": "go run ./cmd",
      "isBackground": true,
      "group": {
        "kind": "build",
        "isDefault": true
      }
    }
  ]
}
```

### GoLand / IntelliJ IDEA

#### Go Settings

1. **File → Settings → Go → GOROOT**: Set to Go 1.25 installation
2. **File → Settings → Go → Go Modules**: Enable "Enable Go modules integration"
3. **File → Settings → Go → Build Tags**: Add custom build tags if needed

#### Code Style

1. **File → Settings → Editor → Code Style → Go**
   - Tabs: Use tabs
   - Indent: 4
   - Imports: Enable "Add parentheses for a single import"
   - Enable "Format imports on save"

#### Run Configurations

```xml
<!-- .idea/runConfigurations/Run_Backend.xml -->
<component name="ProjectRunConfigurationManager">
  <configuration default="false" name="Run Backend" type="GoApplicationRunConfiguration">
    <module name="ugcl-backend-v2" />
    <working_directory value="$PROJECT_DIR$" />
    <go_parameters value="-v" />
    <kind value="PACKAGE" />
    <package value="p9e.in/ugcl/cmd" />
    <directory value="$PROJECT_DIR$" />
    <filePath value="$PROJECT_DIR$/cmd/main.go" />
    <envs>
      <env name="DB_DSN" value="host=localhost port=5432 user=postgres password=postgres dbname=ugcl_dev sslmode=disable" />
    </envs>
    <method v="2" />
  </configuration>
</component>
```

#### External Tools

**Tools → External Tools → Add:**

**Buf Generate:**
- Program: `buf`
- Arguments: `generate`
- Working directory: `$ProjectFileDir$`

**SQLC Generate:**
- Program: `make`
- Arguments: `sqlc-generate-all`
- Working directory: `$ProjectFileDir$`

---

## Troubleshooting

### Common Issues

#### Issue: "cannot find package"

```bash
# Solution: Sync workspace
go work sync
go mod tidy

# Or download modules
go mod download
```

#### Issue: SQLC generation fails

```bash
# Check SQLC version
sqlc version

# Verify sqlc.yaml is correct
cd module/db
sqlc verify

# Regenerate
sqlc generate
```

#### Issue: Buf generate fails

```bash
# Check buf.yaml syntax
buf lint

# Clear buf cache
buf registry clear-cache

# Regenerate
buf generate
```

#### Issue: Database connection fails

```bash
# Check PostgreSQL is running
pg_isready -h localhost -p 5432

# Test connection
psql -h localhost -p 5432 -U postgres -d ugcl_dev

# Check configs.yaml has correct credentials
```

---

## Next Steps

After setting up your development environment:

1. Read [Testing Guide](./testing-guide.md) for testing strategies
2. Read [Proto-First Development](./proto-first-development.md) for API design
3. Read [Deployment Guide](./deployment-guide.md) for production deployment
4. Explore the [Architecture Documentation](../architecture/) for system design

---

## Additional Resources

- [Go Documentation](https://go.dev/doc/)
- [Buf Documentation](https://docs.buf.build/)
- [SQLC Documentation](https://docs.sqlc.dev/)
- [Connect Documentation](https://connectrpc.com/docs/introduction)
- [Uber FX Documentation](https://uber-go.github.io/fx/)
- [Atlas Migration](https://atlasgo.io/getting-started/)

---

**Last Updated:** 2025-10-06
**Version:** 2.0
