# UGCL Backend v2 - System Architecture Overview

## Document Information

- **Version:** 2.0
- **Last Updated:** 2025-10-06
- **Status:** Active
- **Audience:** Solution Architects, Backend Engineers, DevOps Engineers

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [System Overview](#system-overview)
3. [High-Level Architecture](#high-level-architecture)
4. [Module Categories](#module-categories)
5. [Technology Stack](#technology-stack)
6. [System Components](#system-components)
7. [Deployment Architecture](#deployment-architecture)
8. [Data Flow Architecture](#data-flow-architecture)
9. [Security Architecture](#security-architecture)
10. [Integration Patterns](#integration-patterns)
11. [Scalability & Performance](#scalability--performance)
12. [Monitoring & Observability](#monitoring--observability)

---

## Executive Summary

The UGCL Backend v2 is a modern, cloud-native enterprise platform built using a modular monolith architecture pattern. The system is designed to manage complex organizational hierarchies, personnel, projects, documents, workflows, and analytics for large-scale industrial operations, particularly in the dairy and agricultural sectors.

### Key Characteristics

- **Architecture Style:** Modular Monolith with potential for future microservices extraction
- **Primary Technology:** Go (Golang) 1.21+
- **Communication Protocol:** Connect Protocol (gRPC-compatible, HTTP/2)
- **Database:** PostgreSQL 15+ with multi-tenancy support
- **Dependency Injection:** Uber FX framework
- **Code Generation:** SQLC for type-safe database access
- **Object Storage:** MinIO for documents and media
- **Deployment Model:** Single binary with modular initialization

### Business Domains Supported

1. **Identity & Access Management** - Multi-tenant user authentication and authorization
2. **Personnel Management** - Employees, contractors, vendors
3. **Organization Structure** - Divisions, branches, departments
4. **Content Management** - Document management with OCR and metadata
5. **Analytics & Insights** - Business intelligence and reporting
6. **Operations** - Projects, workflows, notifications, scheduling

---

## System Overview

### System Context Diagram

```
                                    ┌─────────────────────────────────────┐
                                    │                                     │
                                    │      UGCL Backend v2 Platform       │
                                    │                                     │
                                    └─────────────────────────────────────┘
                                                    │
                    ┌───────────────────────────────┼───────────────────────────────┐
                    │                               │                               │
            ┌───────▼────────┐              ┌──────▼──────┐               ┌────────▼────────┐
            │                │              │             │               │                 │
            │  Web Client    │              │  Mobile     │               │  Partner        │
            │  (React/Vue)   │              │  Apps       │               │  Systems        │
            │                │              │  (iOS/And.) │               │  (API Clients)  │
            └───────┬────────┘              └──────┬──────┘               └────────┬────────┘
                    │                               │                               │
                    └───────────────────────────────┼───────────────────────────────┘
                                                    │
                                            ┌───────▼────────┐
                                            │                │
                                            │  API Gateway   │
                                            │  (Connect/gRPC)│
                                            │                │
                                            └───────┬────────┘
                                                    │
                    ┌───────────────────────────────┼───────────────────────────────┐
                    │                               │                               │
            ┌───────▼────────┐              ┌──────▼──────┐               ┌────────▼────────┐
            │                │              │             │               │                 │
            │  PostgreSQL    │              │   MinIO     │               │  External       │
            │  Database      │              │   Object    │               │  Services       │
            │  (Multi-tenant)│              │   Storage   │               │  (Email/SMS)    │
            │                │              │             │               │                 │
            └────────────────┘              └─────────────┘               └─────────────────┘
```

### Core Design Principles

1. **Modularity:** Clear module boundaries with well-defined interfaces
2. **Multi-tenancy:** Tenant isolation at database and application levels
3. **Type Safety:** Strongly typed code with compile-time guarantees
4. **Performance:** Efficient database queries with connection pooling
5. **Security:** Defense in depth with multiple authentication layers
6. **Auditability:** Complete audit trails for all business operations
7. **Extensibility:** Plugin-based architecture for custom modules
8. **Maintainability:** Clean code with comprehensive documentation

---

## High-Level Architecture

### Modular Monolith Pattern

The system is built as a modular monolith - a single deployable application composed of loosely-coupled, highly-cohesive modules. This provides the development simplicity of a monolith with the organizational benefits of microservices.

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                          UGCL Backend v2 Application                                 │
│                                                                                      │
│  ┌────────────────────────────────────────────────────────────────────────────────┐│
│  │                          HTTP/gRPC Server Layer                                 ││
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       ││
│  │  │   Connect    │  │   Health     │  │  Reflection  │  │     CORS     │       ││
│  │  │  Handlers    │  │   Checks     │  │   Service    │  │  Middleware  │       ││
│  │  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘       ││
│  └────────────────────────────────────────────────────────────────────────────────┘│
│                                         │                                            │
│  ┌────────────────────────────────────────────────────────────────────────────────┐│
│  │                      Authentication & Authorization Layer                       ││
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       ││
│  │  │     JWT      │  │   API Key    │  │  Permission  │  │     RBAC     │       ││
│  │  │   Validator  │  │   Handler    │  │   Resolver   │  │   Enforcer   │       ││
│  │  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘       ││
│  └────────────────────────────────────────────────────────────────────────────────┘│
│                                         │                                            │
│  ┌────────────────────────────────────────────────────────────────────────────────┐│
│  │                            Business Module Layer                                ││
│  │                                                                                 ││
│  │  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐                ││
│  │  │    Identity     │  │   Personnel     │  │  Organization   │                ││
│  │  │    Modules      │  │    Modules      │  │    Modules      │                ││
│  │  │  ┌───────────┐  │  │  ┌───────────┐  │  │  ┌───────────┐  │                ││
│  │  │  │   Auth    │  │  │  │ Employee  │  │  │  │ Division  │  │                ││
│  │  │  │   User    │  │  │  │Contractor │  │  │  │  Branch   │  │                ││
│  │  │  │  Tenant   │  │  │  │  Vendor   │  │  │  │Department │  │                ││
│  │  │  │  Entity   │  │  │  └───────────┘  │  │  └───────────┘  │                ││
│  │  │  └───────────┘  │  └─────────────────┘  └─────────────────┘                ││
│  │  └─────────────────┘                                                            ││
│  │                                                                                 ││
│  │  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐                ││
│  │  │    Content      │  │   Analytics     │  │   Operations    │                ││
│  │  │    Modules      │  │    Modules      │  │    Modules      │                ││
│  │  │  ┌───────────┐  │  │  ┌───────────┐  │  │  ┌───────────┐  │                ││
│  │  │  │    DMS    │  │  │  │ InsightHub│  │  │  │  Project  │  │                ││
│  │  │  │MetaSearch │  │  │  │InsightView│  │  │  │FormBuilder│  │                ││
│  │  │  └───────────┘  │  │  └───────────┘  │  │  │Notification│ │                ││
│  │  └─────────────────┘  └─────────────────┘  │  │ Scheduler │  │                ││
│  │                                             │  └───────────┘  │                ││
│  │  ┌─────────────────┐  ┌─────────────────┐  └─────────────────┘                ││
│  │  │      Data       │  │    Masters      │                                      ││
│  │  │   Management    │  │     Data        │                                      ││
│  │  │  ┌───────────┐  │  │  ┌───────────┐  │                                      ││
│  │  │  │ BackupDR  │  │  │  │ Pipeline  │  │                                      ││
│  │  │  │DataArchive│  │  │  │  Masters  │  │                                      ││
│  │  │  │DataBridge │  │  │  └───────────┘  │                                      ││
│  │  │  └───────────┘  │  └─────────────────┘                                      ││
│  │  └─────────────────┘                                                            ││
│  └────────────────────────────────────────────────────────────────────────────────┘│
│                                         │                                            │
│  ┌────────────────────────────────────────────────────────────────────────────────┐│
│  │                         Infrastructure Layer                                    ││
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       ││
│  │  │  Database    │  │    Object    │  │    Cache     │  │   Message    │       ││
│  │  │   Manager    │  │   Storage    │  │   Manager    │  │    Queue     │       ││
│  │  │  (SQLC)      │  │   (MinIO)    │  │              │  │              │       ││
│  │  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘       ││
│  └────────────────────────────────────────────────────────────────────────────────┘│
│                                         │                                            │
│  ┌────────────────────────────────────────────────────────────────────────────────┐│
│  │                         Data Persistence Layer                                  ││
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       ││
│  │  │  PostgreSQL  │  │    MinIO     │  │    Redis     │  │   RabbitMQ   │       ││
│  │  │  Database    │  │   Buckets    │  │   (Future)   │  │   (Future)   │       ││
│  │  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘       ││
│  └────────────────────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────────────────────┘
```

---

## Module Categories

The system is organized into six major module categories, each containing related business domains:

### 1. Identity Modules

**Purpose:** Authentication, authorization, and user management

**Modules:**
- **auth** - Authentication service (JWT, API keys)
- **user** - User management and permissions
- **tenant** - Multi-tenant management
- **entity** - Universal entity abstraction

**Key Responsibilities:**
- User login and session management
- JWT token generation and validation
- API key authentication
- Role-based access control (RBAC)
- Permission resolution and enforcement
- Tenant isolation and management
- Entity-based polymorphic identity

**Database Tables:** users, roles, permissions, permission_defs, tenants, entities, entity_role_bindings

**Proto Services:**
```protobuf
service AuthService {
  rpc Login(LoginRequest) returns (LoginResponse);
  rpc Logout(LogoutRequest) returns (LogoutResponse);
  rpc RefreshToken(RefreshTokenRequest) returns (RefreshTokenResponse);
  rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
}

service UserService {
  rpc CreateUser(CreateUserRequest) returns (User);
  rpc GetUser(GetUserRequest) returns (User);
  rpc UpdateUser(UpdateUserRequest) returns (User);
  rpc DeleteUser(DeleteUserRequest) returns (DeleteUserResponse);
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);
}

service TenantService {
  rpc CreateTenant(CreateTenantRequest) returns (Tenant);
  rpc GetTenant(GetTenantRequest) returns (Tenant);
  rpc UpdateTenant(UpdateTenantRequest) returns (Tenant);
  rpc ListTenants(ListTenantsRequest) returns (ListTenantsResponse);
}

service EntityService {
  rpc CreateEntity(CreateEntityRequest) returns (Entity);
  rpc GetEntity(GetEntityRequest) returns (Entity);
  rpc GetEntityByUserId(GetEntityByUserIdRequest) returns (Entity);
  rpc GetEntityByReference(GetEntityByReferenceRequest) returns (Entity);
  rpc ListEntities(ListEntitiesRequest) returns (ListEntitiesResponse);
}
```

### 2. Personnel Modules

**Purpose:** Management of human and organizational actors

**Modules:**
- **employee** - Employee lifecycle management
- **contractors** - Contractor management
- **vendors** - Vendor/supplier management

**Key Responsibilities:**
- Employee onboarding and offboarding
- Contractor engagement management
- Vendor registration and qualification
- Personnel records and documentation
- Role assignments and access provisioning
- Performance tracking integration

**Database Tables:** employees, contractors, vendors (domain-specific schemas)

**Entity Integration:**
Each personnel record creates a corresponding Entity record:
- Employee → ENTITY_TYPE_EMPLOYEE → User
- Contractor → ENTITY_TYPE_CONTRACTOR → User
- Vendor → ENTITY_TYPE_VENDOR → User

### 3. Organization Modules

**Purpose:** Organizational structure and hierarchy management

**Modules:**
- **organization** - Divisions, branches, departments

**Key Responsibilities:**
- Division management (business units)
- Branch management (physical locations)
- Department management (functional units)
- Organizational hierarchy traversal
- Location-based access control
- Reporting structure management

**Database Tables:** divisions, branches, departments

**Hierarchy Model:**
```
Business
  └── Division (e.g., "Dairy Operations")
       ├── Branch (e.g., "Mumbai Factory")
       │    └── Department (e.g., "Quality Control")
       └── Branch (e.g., "Delhi Distribution Center")
            └── Department (e.g., "Logistics")

Business-Level Departments (cross-division)
  ├── Finance
  ├── Human Resources
  └── IT
```

### 4. Content Modules

**Purpose:** Document and content management

**Modules:**
- **dms** - Document Management System
- **metasearch** - Universal search service

**Key Responsibilities:**
- Document upload and storage
- OCR processing for text extraction
- Watermarking and security
- Metadata management and tagging
- Document versioning
- Full-text search across all documents
- Entity-based document ownership
- Permission-based access control

**Database Tables:** documents, document_shares

**Storage Architecture:**
- **MinIO:** Binary file storage
- **PostgreSQL:** Metadata and indexing
- **OCR Engine:** Text extraction
- **Search Index:** Elasticsearch integration (future)

### 5. Analytics Modules

**Purpose:** Business intelligence and reporting

**Modules:**
- **insighthub** - Analytics engine and KPI tracking
- **insightviewer** - Report generation and visualization

**Key Responsibilities:**
- KPI calculation and tracking
- Dashboard data aggregation
- Report generation
- Data visualization support
- Trend analysis
- Operational metrics

**Key Metrics:**
- Production efficiency
- Quality control metrics
- Personnel utilization
- Project progress
- Financial KPIs
- Compliance metrics

### 6. Operations Modules

**Purpose:** Day-to-day operational workflows

**Modules:**
- **projects** - Project management (dairy sites, etc.)
- **formbuilder** - Dynamic form creation and workflow
- **notification** - Multi-channel notifications
- **scheduler** - Task and job scheduling
- **masters** - Master data management (pipelines, etc.)
- **backupdr** - Backup and disaster recovery
- **dataarchive** - Data retention and archival
- **databridge** - Data integration and ETL

**Key Responsibilities:**
- Project lifecycle management
- Dynamic form design and submission
- Workflow automation and approvals
- Email, SMS, and push notifications
- Scheduled job execution
- Master data synchronization
- Automated backups
- Data archival policies
- Third-party system integration

---

## Technology Stack

### Backend Framework

**Language:** Go 1.21+

**Why Go?**
- High performance and concurrency
- Strong typing and compile-time safety
- Excellent tooling and ecosystem
- Native cloud support
- Efficient memory management
- Fast compilation times

### Communication Protocol

**Protocol:** Connect Protocol (gRPC-compatible over HTTP/2)

**Why Connect?**
- gRPC compatibility with HTTP/2 benefits
- Browser-friendly (works with standard HTTP)
- Better error handling than REST
- Streaming support
- Strong typing with Protocol Buffers
- Automatic client generation
- Built-in reflection and health checks

**Example Connect Handler:**
```go
// Service registration with authentication
func (r *ServiceRegistry) registerDMSService(dmsHandler dmshandlers.DMSServiceHandler) {
    dmsOptions := append(r.getCommonConnectOptions(),
        connect.WithInterceptors(
            r.authService.RequireApp([]string{"WebApp", "MobileApp", "InternalOps"}),
        ),
    )

    dmsPath, dmsServiceHandler := dmsv1connect.NewDMSServiceHandler(
        dmsHandler, dmsOptions...,
    )
    r.mux.Handle(dmsPath, dmsServiceHandler)
}
```

### Database Layer

**Database:** PostgreSQL 15+

**Why PostgreSQL?**
- ACID compliance
- Rich data types (JSONB, arrays, enums)
- Full-text search capabilities
- Excellent performance for complex queries
- Mature ecosystem and tooling
- Native JSON support
- Advanced indexing strategies
- Robust transaction management

**Query Builder:** SQLC

**Why SQLC?**
- Type-safe SQL queries at compile time
- No ORM overhead
- Write SQL, get Go code
- Easy to review and optimize queries
- No runtime reflection
- Full control over SQL
- Excellent performance

**Example SQLC Query:**
```sql
-- name: GetUserByID :one
SELECT * FROM users
WHERE uuid = $1 AND deleted_at IS NULL;

-- name: ListUsersByTenant :many
SELECT * FROM users
WHERE tenant_id = $1 AND is_active = true
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
```

**Generated Go Code:**
```go
func (q *Queries) GetUserByID(ctx context.Context, uuid uuid.UUID) (User, error)
func (q *Queries) ListUsersByTenant(ctx context.Context, arg ListUsersByTenantParams) ([]User, error)
```

### Dependency Injection

**Framework:** Uber FX

**Why FX?**
- Explicit dependency graph
- Compile-time safety
- Lifecycle management
- Easy testing
- Clear module boundaries
- Prevents circular dependencies

**Example Module:**
```go
// Module definition
var Module = fx.Module("dms",
    fx.Provide(
        repository.NewDMSRepository,
        services.NewDMSService,
        handlers.NewDMSHandler,
    ),
)

// Application assembly
func NewApplication() *fx.App {
    return fx.New(
        config.DatabaseModule,
        identity.AuthModule,
        identity.UserModule,
        dms.Module,
        organization.Module,
        // HTTP layer
        fx.Invoke(RegisterAllServices),
    )
}
```

### Object Storage

**Storage:** MinIO

**Why MinIO?**
- S3-compatible API
- Self-hosted option
- High performance
- Built-in replication
- Encryption at rest
- Kubernetes native

**Use Cases:**
- Document storage
- Image and video files
- Backup archives
- Generated reports

### Configuration Management

**Format:** YAML

**Configuration Layers:**
- Environment variables
- Configuration files
- Command-line flags
- Default values

**Example Configuration:**
```yaml
server:
  http:
    addr: 0.0.0.0:8080
    timeout: 30s
  grpc:
    addr: 0.0.0.0:9090

data:
  postgres:
    host: localhost
    port: 5432
    dbname: ugcl
    user: postgres
    password: ${DB_PASSWORD}
    search_path: testing2

  minio:
    endpoint: localhost:9000
    access_key: ${MINIO_ACCESS_KEY}
    secret_key: ${MINIO_SECRET_KEY}
    use_ssl: false

auth:
  jwt:
    secret: ${JWT_SECRET}
    expiry: 24h
  api_keys:
    - app: WebApp
      key: ${WEBAPP_API_KEY}
      ip_whitelist: ["*"]
```

### Logging and Observability

**Logging:** Custom p9log package (structured logging)

**Metrics:** Prometheus (planned)

**Tracing:** OpenTelemetry (planned)

**Health Checks:** gRPC Health Checking Protocol

---

## System Components

### Core Components

#### 1. Application Builder

**Location:** `cmd/app_builder.go`

**Purpose:** Constructs the application dependency graph using FX

**Key Functions:**
- Module registration
- Dependency wiring
- Lifecycle management
- Configuration loading

**Architecture:**
```go
type ApplicationBuilder struct {
    modules []fx.Option
}

func (b *ApplicationBuilder) addAllServices() fx.Option {
    return fx.Options(
        b.addProjectServices(),
        b.addMasterServices(),
        b.addIdentityServices(),
        b.addVendorServices(),
        b.addFormBuilderServices(),
        b.addNotificationServices(),
        b.addOrganizationServices(),
        b.addDMSServices(),
    )
}
```

#### 2. Service Registry

**Location:** `cmd/module_registry.go`

**Purpose:** Centralized HTTP/gRPC service registration

**Responsibilities:**
- Route registration
- Middleware application
- Authentication enforcement
- Service discovery

**Example:**
```go
type ServiceRegistry struct {
    mux         *http.ServeMux
    authService *middleware.AuthService
    services    []string
}

func RegisterAllServices(params RegisterAllServicesParams) {
    registry := params.Registry

    if params.DairySiteHandler != nil {
        registry.registerProjectServices(params.DairySiteHandler)
    }

    if params.EntityHandler != nil {
        registry.registerEntityService(params.EntityHandler)
    }

    registry.registerHealthAndReflection()
    registry.logRegisteredServices()
}
```

#### 3. Database Manager

**Location:** `packages/database/sqlc/provider.go`

**Purpose:** Connection pool management and query provider

**Features:**
- Centralized connection pool
- Per-module query access
- Transaction support
- Connection lifecycle management

**Architecture:**
```go
type DatabaseManager struct {
    Pool *pgxpool.Pool
}

func (m *DatabaseManager) GetUserQueries() *userDB.Queries {
    return userDB.New(m.Pool)
}

func (m *DatabaseManager) GetOrganizationQueries() *organizationDB.Queries {
    return organizationDB.New(m.Pool)
}

// Connection pool configuration
config.MaxConns = 10
config.MinConns = 2
config.MaxConnIdleTime = 30 * time.Minute
config.MaxConnLifetime = 2 * time.Hour
```

#### 4. Authentication Service

**Location:** `core/middleware/auth.go`

**Purpose:** Unified authentication and authorization

**Authentication Methods:**
- JWT tokens (for user sessions)
- API keys (for service-to-service)

**Features:**
- Token validation
- Role checking
- Permission enforcement
- Context injection

**Flow:**
```
Client Request
    │
    ▼
Authentication Interceptor
    │
    ├─> JWT Token? ──> Validate ──> Extract Claims
    │                                    │
    └─> API Key? ───> Validate ──> Extract App Info
                                         │
                                         ▼
                                  Inject to Context
                                         │
                                         ▼
                                  Permission Check
                                         │
                                         ▼
                                  Handler Execution
```

### Module Components

Each module follows a consistent layered architecture:

```
Module/
├── api/                    # Generated gRPC/Connect code
│   └── v1/
│       └── {service}/
│           ├── {service}.pb.go
│           └── {service}connect/
│               └── {service}.connect.go
├── proto/                  # Protocol Buffer definitions
│   └── {service}.proto
├── handlers/               # gRPC/Connect handlers
│   └── {service}_handler.go
├── services/               # Business logic
│   └── {service}_service.go
├── repository/             # Data access layer
│   └── {service}_repository.go
├── models/                 # Domain models
│   └── {service}_models.go
├── db/
│   ├── schema/            # SQL schema
│   │   └── schema.sql
│   ├── queries/           # SQLC queries
│   │   └── {service}.sql
│   └── generated/         # SQLC generated code
│       ├── models.go
│       ├── querier.go
│       └── {service}.sql.go
├── mappers/               # DTO/Model conversions
│   └── {service}_mapper.go
└── module.go              # FX module definition
```

**Example Module Structure (DMS):**
```
dms/
├── api/v1/dmsv1connect/
│   └── dms.connect.go
├── proto/
│   └── dms.proto
├── handlers/
│   └── dms_handler.go
├── services/
│   ├── dms_service.go
│   └── ocr_service.go
├── repository/
│   └── document_repository.go
├── models/
│   ├── document.go
│   └── document_share.go
├── db/
│   ├── schema/schema.sql
│   ├── queries/documents.sql
│   └── generated/
│       ├── models.go
│       └── documents.sql.go
└── module.go
```

---

## Deployment Architecture

### Deployment Model: Modular Monolith

The application is deployed as a single Go binary, providing simplicity while maintaining internal modularity.

#### Deployment Topology

```
                        ┌─────────────────────────────────────┐
                        │         Load Balancer               │
                        │      (nginx/HAProxy/ALB)            │
                        └─────────────┬───────────────────────┘
                                      │
                    ┌─────────────────┼─────────────────┐
                    │                 │                 │
         ┌──────────▼──────────┐  ┌──▼──────────────┐ │
         │  UGCL Backend v2    │  │  UGCL Backend v2 │ │
         │    Instance 1       │  │    Instance 2    │ │... (N instances)
         │  ┌──────────────┐   │  │  ┌──────────┐   │ │
         │  │ All Modules  │   │  │  │ All      │   │ │
         │  │ + HTTP/gRPC  │   │  │  │ Modules  │   │ │
         │  └──────────────┘   │  │  └──────────┘   │ │
         └──────────┬──────────┘  └──────┬──────────┘ │
                    │                    │             │
                    └────────────────────┼─────────────┘
                                         │
                    ┌────────────────────┼─────────────────────┐
                    │                    │                     │
         ┌──────────▼──────────┐  ┌─────▼──────────┐  ┌──────▼─────────┐
         │   PostgreSQL        │  │     MinIO      │  │    Redis       │
         │   (Primary)         │  │   (Primary)    │  │   (Future)     │
         │   + Read Replicas   │  │   + Replicas   │  │                │
         └─────────────────────┘  └────────────────┘  └────────────────┘
```

### Container Architecture

**Docker Image:**
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ugcl-backend ./cmd

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/ugcl-backend .
COPY configs.yaml .
EXPOSE 8080 9090
CMD ["./ugcl-backend"]
```

**Docker Compose (Development):**
```yaml
version: '3.8'

services:
  backend:
    build: .
    ports:
      - "8080:8080"
      - "9090:9090"
    environment:
      - DB_PASSWORD=${DB_PASSWORD}
      - JWT_SECRET=${JWT_SECRET}
    depends_on:
      - postgres
      - minio
    volumes:
      - ./configs.yaml:/root/configs.yaml

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: ugcl
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  minio:
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: ${MINIO_ACCESS_KEY}
      MINIO_ROOT_PASSWORD: ${MINIO_SECRET_KEY}
    volumes:
      - minio_data:/data
    ports:
      - "9000:9000"
      - "9001:9001"

volumes:
  postgres_data:
  minio_data:
```

### Kubernetes Deployment (Future)

**Deployment Manifest:**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ugcl-backend
spec:
  replicas: 3
  selector:
    matchLabels:
      app: ugcl-backend
  template:
    metadata:
      labels:
        app: ugcl-backend
    spec:
      containers:
      - name: backend
        image: ugcl/backend:v2.0
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 9090
          name: grpc
        env:
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: ugcl-secrets
              key: db-password
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          grpc:
            port: 9090
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          grpc:
            port: 9090
          initialDelaySeconds: 5
          periodSeconds: 5
```

### Environment Configuration

**Development:**
- Single instance
- Local PostgreSQL
- Local MinIO
- Hot reload enabled
- Debug logging

**Staging:**
- 2-3 instances
- Managed PostgreSQL
- Managed object storage
- Info logging
- Performance monitoring

**Production:**
- Auto-scaling (3-10 instances)
- Managed PostgreSQL with read replicas
- Distributed object storage
- Warn/Error logging
- Full observability stack
- Automated backups
- Disaster recovery

---

## Data Flow Architecture

### Request Flow

```
1. Client Request
   │
   ├─> HTTP/2 (Connect Protocol)
   │   or
   └─> gRPC
       │
       ▼
2. Load Balancer
   │
   ▼
3. HTTP Server (Go)
   │
   ▼
4. Authentication Middleware
   │
   ├─> JWT Validation
   │   └─> Extract user, tenant, roles
   │
   └─> API Key Validation
       └─> Extract app, permissions
       │
       ▼
5. Authorization Middleware
   │
   ├─> Check required roles
   └─> Check required app permissions
       │
       ▼
6. Connect Handler
   │
   ▼
7. Service Layer
   │
   ├─> Business logic
   ├─> Validation
   └─> Orchestration
       │
       ▼
8. Repository Layer
   │
   ├─> SQLC Queries
   └─> Database access
       │
       ▼
9. PostgreSQL Database
   │
   ▼
10. Response (Proto)
    │
    ▼
11. Client
```

### Document Upload Flow

```
1. Client → Upload Request (multipart/form-data)
   │
   ▼
2. DMS Handler → Validate file
   │
   ├─> Check file type
   ├─> Check file size
   ├─> Virus scan (future)
   └─> Calculate checksum
       │
       ▼
3. MinIO Service → Store file
   │
   ├─> Generate unique path: tenant_id/entity_id/YYYY/MM/file_id.ext
   └─> Upload to bucket
       │
       ▼
4. Document Repository → Save metadata
   │
   ├─> Insert document record
   ├─> Set processing_status = PENDING
   └─> Commit transaction
       │
       ▼
5. Background Processing (async)
   │
   ├─> OCR Processing (if image/PDF)
   │   ├─> Extract text
   │   ├─> Calculate confidence
   │   └─> Update ocr_status = COMPLETED
   │
   ├─> Thumbnail Generation
   │   └─> Create preview image
   │
   ├─> Compression
   │   └─> Create tar.zstd archive
   │
   └─> Indexing (future)
       └─> Add to search index
       │
       ▼
6. Update document → processing_status = COMPLETED
   │
   ▼
7. Response → Document metadata
```

### Permission Resolution Flow

```
1. User makes request
   │
   ▼
2. Extract user_id from JWT
   │
   ▼
3. Lookup Entity by user_id
   │
   └─> entities WHERE user_id = ?
       │
       ▼
4. Get EntityRoleBindings
   │
   └─> entity_role_bindings WHERE entity_id = ?
       │
       ├─> Filter by organization scope (division/branch/dept)
       ├─> Filter by time bounds (valid_from, valid_until)
       └─> Get list of role_ids
       │
       ▼
5. Resolve Role Hierarchy
   │
   └─> roles WHERE id IN (?) OR parent_id IN (?)
       │
       ▼
6. Get Permission Definitions
   │
   └─> role_permission_defs WHERE role_id IN (?)
       │
       ▼
7. Materialize Permissions
   │
   └─> permissions WHERE def_name IN (?)
       │
       ├─> Apply namespace filter
       ├─> Apply resource filter
       ├─> Apply action filter
       └─> Apply effect (GRANT/FORBIDDEN)
       │
       ▼
8. Check User-level overrides
   │
   └─> user_permissions WHERE user_id = ?
       │
       └─> Override role permissions if exists
       │
       ▼
9. Return final permission decision
   │
   └─> GRANT | FORBIDDEN | UNKNOWN
```

### Multi-Tenant Data Isolation

```
Every Query Pattern:

SELECT * FROM {table}
WHERE tenant_id = $1
  AND deleted_at IS NULL
  AND {other_conditions}

Examples:

-- Get user's documents
SELECT * FROM documents
WHERE tenant_id = 'tenant-123'
  AND owner_entity_id = 'entity-456'
  AND deleted_at IS NULL
ORDER BY created_at DESC;

-- Get division branches
SELECT * FROM branches
WHERE tenant_id = 'tenant-123'
  AND division_id = 'division-789'
  AND is_active = true
ORDER BY display_order;

Tenant Context Injection:

1. JWT contains tenant_id claim
2. Middleware extracts tenant_id
3. Inject into context: ctx = context.WithValue(ctx, "tenant_id", tenantID)
4. Repository reads from context: tenantID := ctx.Value("tenant_id").(string)
5. All queries automatically filter by tenant_id
```

---

## Security Architecture

### Defense in Depth

```
┌─────────────────────────────────────────────────────────────────┐
│ Layer 1: Network Security                                       │
│  - Firewall rules                                               │
│  - VPC isolation                                                │
│  - TLS/SSL encryption                                           │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│ Layer 2: API Gateway                                            │
│  - Rate limiting                                                │
│  - IP whitelisting                                              │
│  - Request validation                                           │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│ Layer 3: Authentication                                         │
│  - JWT validation (HS256)                                       │
│  - API key verification                                         │
│  - Token expiration checks                                      │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│ Layer 4: Authorization                                          │
│  - Role-based access control (RBAC)                             │
│  - Permission enforcement                                       │
│  - Organizational scope filtering                               │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│ Layer 5: Data Access                                            │
│  - Tenant isolation                                             │
│  - Row-level security                                           │
│  - Prepared statements (SQL injection prevention)               │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│ Layer 6: Data Encryption                                        │
│  - Encryption at rest (database)                                │
│  - Encryption in transit (TLS)                                  │
│  - Encrypted backups                                            │
└─────────────────────────────────────────────────────────────────┘
```

### Authentication Mechanisms

#### JWT Authentication

**Token Structure:**
```json
{
  "header": {
    "alg": "HS256",
    "typ": "JWT"
  },
  "payload": {
    "user_id": "uuid-123",
    "tenant_id": "tenant-456",
    "name": "John Doe",
    "email": "john@example.com",
    "phone": "+91XXXXXXXXXX",
    "role": "admin",
    "entity_id": "entity-789",
    "entity_type": "EMPLOYEE",
    "exp": 1704067200,
    "iat": 1703980800
  },
  "signature": "..."
}
```

**Usage:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

#### API Key Authentication

**Configuration:**
```yaml
auth:
  api_keys:
    - app: WebApp
      key: ${WEBAPP_API_KEY}
      allowed_paths: ["/*"]
      allowed_methods: ["GET", "POST", "PUT", "DELETE"]
      ip_whitelist: ["*"]

    - app: MobileApp
      key: ${MOBILE_API_KEY}
      allowed_paths: ["/api/*"]
      allowed_methods: ["GET", "POST"]
      ip_whitelist: ["203.0.113.0/24"]

    - app: InternalOps
      key: ${INTERNAL_API_KEY}
      allowed_paths: ["/*"]
      allowed_methods: ["*"]
      ip_whitelist: ["10.0.0.0/8"]
```

**Usage:**
```
x-api-key: your-api-key-here
```

### Authorization Model

**RBAC + Organizational Scoping:**

```sql
-- Example: User has "manager" role scoped to Division A, Branch B
INSERT INTO entity_role_bindings (
    entity_id,
    role_id,
    division_id,
    branch_id,
    valid_from,
    valid_until
) VALUES (
    'entity-123',
    'role-manager',
    'division-a',
    'branch-b',
    '2024-01-01',
    '2025-01-01'
);

-- Permission check: Can user edit documents in Branch B?
-- 1. Get user's entity
-- 2. Get entity's role bindings filtered by:
--    - division_id = 'division-a'
--    - branch_id = 'branch-b'
--    - NOW() BETWEEN valid_from AND valid_until
-- 3. Get roles and their permissions
-- 4. Check if permission "document:edit" exists with GRANT effect
```

### Audit Trail

All operations are logged with:
- **Who:** user_id, entity_id
- **What:** action, resource, changes
- **When:** timestamp
- **Where:** IP address, user agent
- **Why:** (if provided in request)

**Audit Log Schema:**
```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    entity_id UUID,
    user_id UUID,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id UUID,
    changes JSONB,
    ip_address INET,
    user_agent TEXT,
    timestamp TIMESTAMPTZ DEFAULT NOW()
);
```

---

## Integration Patterns

### Internal Module Communication

**Pattern:** Direct Function Calls (Shared Memory)

Since all modules run in the same process, they communicate via direct Go function calls injected through FX.

**Example:**
```go
// Entity module provides entity service
func NewEntityService(repo EntityRepository) IEntityService {
    return &EntityService{repo: repo}
}

// DMS module depends on entity service
func NewDMSService(
    docRepo DocumentRepository,
    entityService entity.IEntityService, // Injected by FX
) IDMSService {
    return &DMSService{
        docRepo: docRepo,
        entityService: entityService,
    }
}

// Usage in DMS
func (s *DMSService) CreateDocument(ctx context.Context, req *CreateDocumentRequest) error {
    // Validate owner entity exists
    entity, err := s.entityService.GetEntity(ctx, req.OwnerEntityId)
    if err != nil {
        return fmt.Errorf("invalid owner entity: %w", err)
    }

    // Create document
    return s.docRepo.Create(ctx, document)
}
```

### External System Integration

**Pattern:** API Gateway / Client SDKs

For external systems, the platform exposes Connect/gRPC APIs.

**Example External Integration:**
```javascript
// JavaScript client (auto-generated from proto)
import { createPromiseClient } from "@connectrpc/connect";
import { DMSService } from "./gen/dms/v1/dms_connect";

const client = createPromiseClient(DMSService, {
  baseUrl: "https://api.ugcl.com",
  headers: {
    "x-api-key": process.env.API_KEY,
  },
});

// Upload document
const response = await client.uploadDocument({
  tenantId: "tenant-123",
  file: fileBuffer,
  metadata: {
    title: "Invoice Q4 2024",
    category: "INVOICE",
  },
});
```

### Event-Driven Communication (Future)

**Pattern:** Message Queue (RabbitMQ/NATS)

For asynchronous workflows, events will be published to a message queue.

**Use Cases:**
- Document processing (OCR, thumbnails)
- Email notifications
- Scheduled jobs
- Webhook triggers
- Data synchronization

**Example:**
```go
// Publish event
type DocumentUploadedEvent struct {
    DocumentID string
    TenantID   string
    EntityID   string
    FileType   string
}

eventBus.Publish("documents.uploaded", DocumentUploadedEvent{
    DocumentID: doc.ID,
    TenantID:   doc.TenantID,
    EntityID:   doc.OwnerEntityID,
    FileType:   doc.MimeType,
})

// Subscribe to event
eventBus.Subscribe("documents.uploaded", func(event DocumentUploadedEvent) {
    // Trigger OCR processing
    ocrService.ProcessDocument(event.DocumentID)
})
```

---

## Scalability & Performance

### Horizontal Scaling

**Stateless Design:**
- No in-memory session storage
- All state in PostgreSQL/Redis
- Sticky sessions not required
- Easy to add/remove instances

**Load Balancing:**
- Round-robin distribution
- Health check based routing
- Session affinity (optional)

### Database Optimization

**Connection Pooling:**
```go
config.MaxConns = 10       // Max connections per instance
config.MinConns = 2        // Min idle connections
config.MaxConnIdleTime = 30 * time.Minute
config.MaxConnLifetime = 2 * time.Hour
```

**Indexing Strategy:**
- Primary keys on all tables
- Foreign key indexes
- Composite indexes for common queries
- Partial indexes for filtered queries
- GIN indexes for JSONB and arrays

**Query Optimization:**
- Use of prepared statements (SQLC)
- Efficient JOINs with proper indexes
- Pagination for large result sets
- SELECT only required columns
- Avoid N+1 queries

**Read Replicas (Future):**
```
Write Operations → Primary DB
Read Operations → Read Replicas (round-robin)
```

### Caching Strategy (Future)

**Redis Cache:**
- User sessions
- Permission cache
- Frequently accessed master data
- Query result cache

**TTL Strategy:**
```
User sessions: 24 hours
Permissions: 5 minutes
Master data: 1 hour
Query results: 1 minute
```

### Performance Benchmarks

**Target Metrics:**
- API response time: < 100ms (p95)
- Database query time: < 50ms (p95)
- Document upload: < 5s for 10MB file
- Concurrent users: 1000+ per instance
- Throughput: 1000+ req/sec per instance

---

## Monitoring & Observability

### Health Checks

**gRPC Health Protocol:**
```go
checker := grpchealth.NewStaticChecker(
    "p9e.ugcl.identity.api.v2.user.User",
    "p9e.ugcl.dms.api.v1.DMS",
    "p9e.ugcl.organization.api.v1.Organization",
)
mux.Handle(grpchealth.NewHandler(checker))
```

**Health Endpoint:**
```
GET /grpc.health.v1.Health/Check

Response:
{
  "status": "SERVING"
}
```

### Logging

**Structured Logging (p9log):**
```go
logger := p9log.With(
    p9log.NewStdLogger(logFile),
    "Time", p9log.DefaultTimestamp,
    "PID", hostname,
    "Service", "UGCL-Backend",
)

logger.Log("level", "info", "msg", "service started", "port", 8080)
logger.Log("level", "error", "msg", "database connection failed", "error", err)
```

**Log Levels:**
- **DEBUG:** Development only
- **INFO:** Service lifecycle events
- **WARN:** Degraded functionality
- **ERROR:** Operation failures
- **FATAL:** Service crashes

### Metrics (Planned)

**Prometheus Metrics:**
```
# Request metrics
http_requests_total{method, path, status}
http_request_duration_seconds{method, path}

# Database metrics
db_connections_active
db_connections_idle
db_query_duration_seconds{query}

# Business metrics
documents_uploaded_total{tenant_id}
users_active_total{tenant_id}
permissions_checked_total{result}
```

### Tracing (Planned)

**OpenTelemetry:**
- Request tracing across modules
- Database query tracing
- External API call tracing
- Performance bottleneck identification

---

## Conclusion

The UGCL Backend v2 represents a modern, scalable, and secure enterprise platform built with industry best practices. The modular monolith architecture provides the simplicity of a monolith with the organizational benefits of microservices, allowing for future extraction of modules into independent services as needed.

**Key Strengths:**
- Type-safe, high-performance Go codebase
- Connect Protocol for modern RPC communication
- Multi-tenant architecture with strong isolation
- Comprehensive security with multiple authentication methods
- Entity-based polymorphic identity system
- Extensive audit trail and observability
- Modular design for easy maintenance and extension

**Future Roadmap:**
- Event-driven architecture with message queues
- Advanced caching with Redis
- Elasticsearch integration for search
- Kubernetes deployment
- Microservices extraction (if needed)
- Real-time features with WebSockets
- Machine learning integration for analytics

---

**Document Version:** 2.0
**Lines:** 1200+
**Generated:** 2025-10-06
**Maintained By:** UGCL Architecture Team
