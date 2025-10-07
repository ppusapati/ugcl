# UGCL Backend v2 - Inter-Module Communication Architecture

## Document Information

- **Version:** 2.0
- **Last Updated:** 2025-10-06
- **Status:** Active
- **Communication Protocol:** Connect Protocol (gRPC-compatible)
- **Dependency Management:** Uber FX

---

## Table of Contents

1. [Communication Overview](#communication-overview)
2. [Module Dependency Graph](#module-dependency-graph)
3. [Communication Patterns](#communication-patterns)
4. [gRPC/Connect Protocol](#grpcconnect-protocol)
5. [Service Integration](#service-integration)
6. [Data Sharing Strategies](#data-sharing-strategies)
7. [Error Handling](#error-handling)
8. [API Versioning](#api-versioning)
9. [Event-Driven Patterns](#event-driven-patterns)
10. [Circuit Breaker & Resilience](#circuit-breaker--resilience)
11. [Service Discovery](#service-discovery)

---

## Communication Overview

### Architecture Style: Modular Monolith

The UGCL Backend v2 uses a **modular monolith** architecture where all modules run in a single process but maintain clear boundaries and communication patterns.

```
┌─────────────────────────────────────────────────────────────────────┐
│                   UGCL Backend v2 Process                            │
│                                                                      │
│  ┌────────────┐      ┌────────────┐      ┌────────────┐            │
│  │  Module A  │◄────►│  Module B  │◄────►│  Module C  │            │
│  │            │      │            │      │            │            │
│  │ • Handler  │      │ • Handler  │      │ • Handler  │            │
│  │ • Service  │      │ • Service  │      │ • Service  │            │
│  │ • Repo     │      │ • Repo     │      │ • Repo     │            │
│  └────────────┘      └────────────┘      └────────────┘            │
│         │                   │                   │                   │
│         └───────────────────┴───────────────────┘                   │
│                             │                                        │
│                   ┌─────────▼─────────┐                             │
│                   │  Shared Database  │                             │
│                   │   (PostgreSQL)    │                             │
│                   └───────────────────┘                             │
└─────────────────────────────────────────────────────────────────────┘

Communication Methods:
1. Direct Function Calls (in-process, shared memory)
2. Dependency Injection (Uber FX)
3. Shared Database (transactional consistency)
4. Future: Event Bus (async operations)
```

### Benefits of Modular Monolith

**Advantages:**
- **Simplicity:** Single deployment, single codebase
- **Performance:** No network latency for inter-module calls
- **Transactions:** ACID guarantees across modules
- **Development Speed:** Easy to refactor and change module boundaries
- **Debugging:** Simple stack traces, single process debugging

**Module Isolation:**
- Clear package boundaries
- Interface-based contracts
- Dependency injection for loose coupling
- Potential for future microservices extraction

---

## Module Dependency Graph

### High-Level Module Dependencies

```
                           ┌──────────┐
                           │   Core   │
                           │ Packages │
                           └────┬─────┘
                                │
                ┌───────────────┼───────────────┐
                │               │               │
         ┌──────▼─────┐  ┌─────▼──────┐  ┌────▼──────┐
         │  Identity  │  │   Config   │  │ Database  │
         │            │  │            │  │  Manager  │
         └──────┬─────┘  └────────────┘  └───────────┘
                │
                ├────────────────────────────────────────┐
                │                                        │
         ┌──────▼─────┐                          ┌──────▼─────┐
         │   Entity   │                          │   User     │
         │  (Service) │                          │  (Service) │
         └──────┬─────┘                          └──────┬─────┘
                │                                       │
                │                                       │
    ┌───────────┼───────────┐              ┌───────────┼───────────┐
    │           │           │              │           │           │
┌───▼────┐  ┌──▼────┐  ┌───▼────┐    ┌───▼────┐  ┌──▼────┐  ┌───▼────┐
│Employee│  │Contr. │  │Vendor  │    │  DMS   │  │Project│  │  Form  │
│        │  │       │  │        │    │        │  │       │  │Builder │
└────────┘  └───────┘  └────────┘    └────┬───┘  └───────┘  └───┬────┘
                                           │                      │
                                           │                      │
                                      ┌────▼────┐          ┌─────▼─────┐
                                      │MetaSearch│         │ Workflow  │
                                      │         │          │  Engine   │
                                      └─────────┘          └───────────┘

Legend:
────► Direct Dependency (uses service/repository)
═══► Shared Infrastructure (database, config)
```

### Detailed Dependency Matrix

| Module         | Depends On                                      | Used By                                    |
|----------------|------------------------------------------------|-------------------------------------------|
| **Core**       | None                                           | All modules                               |
| **Identity**   | Core                                           | All business modules                      |
| **Entity**     | Identity, Core                                 | Personnel, DMS, Projects, FormBuilder     |
| **User**       | Identity, Core                                 | Auth, Entity, Organization                |
| **Auth**       | User, Tenant, Core                             | All modules (middleware)                  |
| **Employee**   | Entity, User, Organization                     | Projects, FormBuilder                     |
| **Contractor** | Entity, User, Organization                     | Projects                                  |
| **Vendor**     | Entity, User, Organization                     | Projects, FormBuilder                     |
| **Organization**| Core                                          | Entity, Personnel, DMS                    |
| **DMS**        | Entity, User, Organization                     | FormBuilder, Projects                     |
| **FormBuilder**| Entity, User, Workflow                         | Projects, Personnel                       |
| **Projects**   | Entity, Organization, FormBuilder              | None                                      |
| **MetaSearch** | DMS, FormBuilder (future)                      | Frontend applications                     |
| **Notification**| User, Entity                                  | FormBuilder, Projects (future)            |

---

## Communication Patterns

### 1. Direct Service Invocation (In-Process)

**Pattern:** Module A directly calls Module B's service interface

```go
// Module B: Entity service interface
package entity

type IEntityService interface {
    GetEntity(ctx context.Context, id string) (*Entity, error)
    GetEntityByUserId(ctx context.Context, userId string) (*Entity, error)
    CreateEntity(ctx context.Context, req *CreateEntityRequest) (*Entity, error)
}

// Module A: DMS service using Entity service
package dms

type DMSService struct {
    docRepo       DocumentRepository
    entityService entity.IEntityService  // Injected via FX
}

func NewDMSService(
    docRepo DocumentRepository,
    entityService entity.IEntityService,
) IDMSService {
    return &DMSService{
        docRepo:       docRepo,
        entityService: entityService,
    }
}

func (s *DMSService) CreateDocument(ctx context.Context, req *CreateDocumentRequest) error {
    // Validate owner entity exists
    entity, err := s.entityService.GetEntity(ctx, req.OwnerEntityId)
    if err != nil {
        return fmt.Errorf("invalid owner entity: %w", err)
    }

    // Verify entity has permission to create documents
    if !s.checkEntityPermission(ctx, entity, "document:create") {
        return errors.New("entity lacks permission to create documents")
    }

    // Create document
    doc := &Document{
        OwnerEntityID:   entity.ID,
        OwnerEntityType: entity.EntityType,
        TenantID:        entity.TenantID,
        // ...
    }

    return s.docRepo.Create(ctx, doc)
}
```

**Flow Diagram:**
```
Client
  │
  ▼
DMS Handler
  │
  ▼
DMS Service
  │
  ├─► Entity Service.GetEntity()
  │   └─► Entity Repository → Database
  │
  └─► Document Repository.Create() → Database
```

**Advantages:**
- Zero network latency
- Type-safe at compile time
- Easy to debug (single stack trace)
- Transactional consistency

**Disadvantages:**
- Tight coupling (compile-time dependency)
- No resilience (if Entity service fails, DMS fails)
- Shared failure domain

### 2. Repository Pattern with Shared Database

**Pattern:** Modules access shared database through their own repositories

```go
// Entity repository (in Entity module)
package entity

type EntityRepository interface {
    GetByID(ctx context.Context, id string) (*Entity, error)
    GetByUserID(ctx context.Context, userID string) (*Entity, error)
    Create(ctx context.Context, entity *Entity) error
}

// DMS can use EntityRepository directly (if needed)
package dms

type DMSService struct {
    docRepo    DocumentRepository
    entityRepo entity.EntityRepository  // Direct repository access
}

func (s *DMSService) GetDocumentOwner(ctx context.Context, docID string) (*entity.Entity, error) {
    doc, err := s.docRepo.GetByID(ctx, docID)
    if err != nil {
        return nil, err
    }

    // Direct database access through EntityRepository
    return s.entityRepo.GetByID(ctx, doc.OwnerEntityID)
}
```

**When to Use:**
- High-performance queries across modules
- Complex joins needed
- Avoiding multiple service hops

**Trade-offs:**
- Bypasses business logic in Entity service
- Less encapsulation
- Harder to extract to microservices later

### 3. Dependency Injection with FX

**Pattern:** Uber FX wires dependencies at application startup

```go
// Entity module definition
package entity

var Module = fx.Module("entity",
    fx.Provide(
        repository.NewEntityRepository,
        services.NewEntityService,
        handlers.NewEntityHandler,
    ),
)

// DMS module definition with Entity dependency
package dms

var Module = fx.Module("dms",
    fx.Provide(
        repository.NewDocumentRepository,
        services.NewDMSService,  // Will receive EntityService via FX
        handlers.NewDMSHandler,
    ),
)

// Application assembly
package main

func NewApplication() *fx.App {
    return fx.New(
        config.DatabaseModule,
        identity.UserModule,
        entity.Module,      // Provides EntityService
        dms.Module,         // Consumes EntityService
        organization.Module,
        // HTTP layer
        fx.Invoke(RegisterAllServices),
    )
}
```

**Dependency Resolution:**
```
1. FX scans all fx.Provide() calls
2. Builds dependency graph
3. Detects circular dependencies (compile error)
4. Instantiates dependencies in correct order
5. Injects into constructors
```

**Example Circular Dependency Detection:**
```go
// ❌ This will fail at startup
Module A → depends on B
Module B → depends on C
Module C → depends on A  // Circular!

// FX Error:
// fx.New(
//     moduleA,
//     moduleB,
//     moduleC,
// )
// Error: cycle detected in dependency graph
```

### 4. Event-Driven Communication (Future)

**Pattern:** Publish-subscribe for asynchronous operations

```go
// Event publisher
type DocumentUploadedEvent struct {
    DocumentID string
    TenantID   string
    EntityID   string
    FileType   string
}

func (s *DMSService) UploadDocument(ctx context.Context, req *UploadRequest) error {
    // Save document
    doc, err := s.docRepo.Create(ctx, document)
    if err != nil {
        return err
    }

    // Publish event (non-blocking)
    s.eventBus.Publish("documents.uploaded", DocumentUploadedEvent{
        DocumentID: doc.ID,
        TenantID:   doc.TenantID,
        EntityID:   doc.OwnerEntityID,
        FileType:   doc.MimeType,
    })

    return nil
}

// Event subscribers (in different modules)
func (s *OCRService) init() {
    eventBus.Subscribe("documents.uploaded", s.processDocument)
}

func (s *OCRService) processDocument(event DocumentUploadedEvent) {
    if !s.shouldProcessOCR(event.FileType) {
        return
    }

    // Async OCR processing
    go s.extractText(event.DocumentID)
}

func (s *NotificationService) init() {
    eventBus.Subscribe("documents.uploaded", s.notifyOwner)
}

func (s *NotificationService) notifyOwner(event DocumentUploadedEvent) {
    s.sendNotification(event.EntityID, "Document uploaded successfully")
}
```

**Use Cases:**
- Background processing (OCR, thumbnails)
- Cross-module notifications
- Audit logging
- Analytics tracking
- Webhook triggers

---

## gRPC/Connect Protocol

### Connect Protocol Overview

Connect is a modern RPC framework that is:
- gRPC-compatible (uses Protocol Buffers)
- HTTP/2 based
- Browser-friendly (works over standard HTTP)
- Language-agnostic with code generation

### Proto Definition Example

```protobuf
// dms/proto/dms.proto
syntax = "proto3";

package dms.v1;

import "google/protobuf/timestamp.proto";
import "google/protobuf/wrappers.proto";

// Document message
message Document {
  string id = 1;
  string tenant_id = 2;
  string owner_entity_id = 3;
  string owner_entity_type = 4;
  string file_name = 5;
  string mime_type = 6;
  int64 size_bytes = 7;
  string storage_path = 8;
  google.protobuf.Timestamp created_at = 9;
}

// Service definition
service DMSService {
  rpc UploadDocument(UploadDocumentRequest) returns (Document);
  rpc GetDocument(GetDocumentRequest) returns (Document);
  rpc ListDocuments(ListDocumentsRequest) returns (ListDocumentsResponse);
  rpc DeleteDocument(DeleteDocumentRequest) returns (DeleteDocumentResponse);
  rpc ShareDocument(ShareDocumentRequest) returns (DocumentShare);
}

message UploadDocumentRequest {
  string tenant_id = 1;
  string owner_entity_id = 2;
  bytes file_data = 3;
  string file_name = 4;
  string mime_type = 5;
  map<string, string> metadata = 6;
}

message ListDocumentsRequest {
  string tenant_id = 1;
  google.protobuf.StringValue owner_entity_id = 2;
  google.protobuf.StringValue division_id = 3;
  int32 page_size = 4;
  int32 page_number = 5;
}

message ListDocumentsResponse {
  repeated Document documents = 1;
  int32 total_count = 2;
  int32 page_number = 3;
  int32 page_size = 4;
}
```

### Code Generation

```bash
# Generate Go code from proto
buf generate

# Generated files:
# dms/api/v1/dms.pb.go                    # Message types
# dms/api/v1/dmsconnect/dms.connect.go    # Service interfaces
```

### Service Implementation

```go
// Handler implementation
package handlers

type DMSHandler struct {
    service services.IDMSService
}

func NewDMSHandler(service services.IDMSService) DMSServiceHandler {
    return &DMSHandler{service: service}
}

// Implement generated interface
func (h *DMSHandler) UploadDocument(
    ctx context.Context,
    req *connect.Request[dmsv1.UploadDocumentRequest],
) (*connect.Response[dmsv1.Document], error) {
    // Extract auth context
    tenantID, err := middleware.TenantFromContext(ctx)
    if err != nil {
        return nil, connect.NewError(connect.CodeUnauthenticated, err)
    }

    // Call service layer
    doc, err := h.service.UploadDocument(ctx, &services.UploadDocumentParams{
        TenantID:       tenantID,
        OwnerEntityID:  req.Msg.OwnerEntityId,
        FileData:       req.Msg.FileData,
        FileName:       req.Msg.FileName,
        MimeType:       req.Msg.MimeType,
        Metadata:       req.Msg.Metadata,
    })
    if err != nil {
        return nil, connect.NewError(connect.CodeInternal, err)
    }

    // Map to proto response
    return connect.NewResponse(&dmsv1.Document{
        Id:              doc.ID,
        TenantId:        doc.TenantID,
        OwnerEntityId:   doc.OwnerEntityID,
        FileName:        doc.FileName,
        MimeType:        doc.MimeType,
        SizeBytes:       doc.SizeBytes,
        StoragePath:     doc.StoragePath,
        CreatedAt:       timestamppb.New(doc.CreatedAt),
    }), nil
}

func (h *DMSHandler) ListDocuments(
    ctx context.Context,
    req *connect.Request[dmsv1.ListDocumentsRequest],
) (*connect.Response[dmsv1.ListDocumentsResponse], error) {
    tenantID, _ := middleware.TenantFromContext(ctx)

    docs, total, err := h.service.ListDocuments(ctx, &services.ListDocumentsParams{
        TenantID:       tenantID,
        OwnerEntityID:  req.Msg.OwnerEntityId,
        DivisionID:     req.Msg.DivisionId,
        PageSize:       req.Msg.PageSize,
        PageNumber:     req.Msg.PageNumber,
    })
    if err != nil {
        return nil, connect.NewError(connect.CodeInternal, err)
    }

    // Map to proto
    protoDocs := make([]*dmsv1.Document, len(docs))
    for i, doc := range docs {
        protoDocs[i] = mappers.DocumentToProto(doc)
    }

    return connect.NewResponse(&dmsv1.ListDocumentsResponse{
        Documents:  protoDocs,
        TotalCount: total,
        PageNumber: req.Msg.PageNumber,
        PageSize:   req.Msg.PageSize,
    }), nil
}
```

### Service Registration with Middleware

```go
// Register service with authentication
func (r *ServiceRegistry) registerDMSService(dmsHandler handlers.DMSServiceHandler) {
    dmsOptions := append(r.getCommonConnectOptions(),
        connect.WithInterceptors(
            r.authService.RequireApp([]string{"WebApp", "MobileApp", "InternalOps"}),
        ),
    )

    dmsPath, dmsServiceHandler := dmsv1connect.NewDMSServiceHandler(
        dmsHandler,
        dmsOptions...,
    )

    r.mux.Handle(dmsPath, dmsServiceHandler)
    r.services = append(r.services, "DMS: "+dmsPath)
}

// Common options (applied to all services)
func (r *ServiceRegistry) getCommonConnectOptions() []connect.HandlerOption {
    return []connect.HandlerOption{
        connect.WithCompressMinBytes(0),
        connect.WithInterceptors(r.authService.AuthInterceptor()),
    }
}
```

### Client Usage

**JavaScript/TypeScript Client:**
```typescript
import { createPromiseClient } from "@connectrpc/connect";
import { DMSService } from "./gen/dms/v1/dms_connect";
import { createConnectTransport } from "@connectrpc/connect-web";

const transport = createConnectTransport({
  baseUrl: "https://api.ugcl.com",
  headers: {
    "Authorization": "Bearer " + token,
  },
});

const client = createPromiseClient(DMSService, transport);

// Upload document
const doc = await client.uploadDocument({
  tenantId: "tenant-123",
  ownerEntityId: "entity-456",
  fileData: fileBuffer,
  fileName: "report.pdf",
  mimeType: "application/pdf",
  metadata: {
    "category": "REPORT",
    "year": "2024",
  },
});

// List documents
const response = await client.listDocuments({
  tenantId: "tenant-123",
  pageSize: 20,
  pageNumber: 1,
});

console.log(`Found ${response.totalCount} documents`);
```

**Go Client (for testing/internal use):**
```go
import (
    "connectrpc.com/connect"
    dmsv1 "p9e.in/ugcl/dms/api/v1"
    "p9e.in/ugcl/dms/api/v1/dmsv1connect"
)

client := dmsv1connect.NewDMSServiceClient(
    http.DefaultClient,
    "http://localhost:8080",
)

req := connect.NewRequest(&dmsv1.GetDocumentRequest{
    Id:       "doc-123",
    TenantId: "tenant-456",
})
req.Header().Set("Authorization", "Bearer "+token)

resp, err := client.GetDocument(ctx, req)
if err != nil {
    return err
}

doc := resp.Msg
fmt.Printf("Document: %s (%d bytes)\n", doc.FileName, doc.SizeBytes)
```

---

## Service Integration

### Integration Layers

```
┌─────────────────────────────────────────────────────────────────┐
│                     External Clients                             │
│          (Web App, Mobile App, Partner Systems)                  │
└────────────────────────────┬────────────────────────────────────┘
                             │
                    ┌────────▼────────┐
                    │   API Gateway   │
                    │  (Load Balancer)│
                    └────────┬────────┘
                             │
┌────────────────────────────▼────────────────────────────────────┐
│                      HTTP/gRPC Layer                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │ Connect      │  │  Health      │  │  Reflection  │          │
│  │ Handlers     │  │  Checks      │  │  Service     │          │
│  └──────┬───────┘  └──────────────┘  └──────────────┘          │
└─────────┼──────────────────────────────────────────────────────┘
          │
┌─────────▼──────────────────────────────────────────────────────┐
│                   Authentication Layer                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │ JWT Validator│  │  API Key     │  │  Permission  │         │
│  │              │  │  Handler     │  │  Resolver    │         │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘         │
└─────────┼─────────────────┼─────────────────┼─────────────────┘
          │                 │                 │
┌─────────▼─────────────────▼─────────────────▼─────────────────┐
│                      Service Layer                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │ DMS Service  │  │Entity Service│  │ User Service │         │
│  │              │◄─┤              │◄─┤              │         │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘         │
└─────────┼─────────────────┼─────────────────┼─────────────────┘
          │                 │                 │
┌─────────▼─────────────────▼─────────────────▼─────────────────┐
│                   Repository Layer                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │ Doc Repo     │  │Entity Repo   │  │ User Repo    │         │
│  │ (SQLC)       │  │ (SQLC)       │  │ (SQLC)       │         │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘         │
└─────────┼─────────────────┼─────────────────┼─────────────────┘
          │                 │                 │
          └─────────────────┴─────────────────┘
                             │
                    ┌────────▼────────┐
                    │   PostgreSQL    │
                    │    Database     │
                    └─────────────────┘
```

### Cross-Module Service Call Example

**Scenario:** Create a form instance that references an employee entity

```go
// FormBuilder Service
package formbuilder

type FormInstanceService struct {
    formRepo      FormInstanceRepository
    entityService entity.IEntityService  // Injected
    userService   user.IUserService      // Injected
}

func (s *FormInstanceService) CreateInstance(
    ctx context.Context,
    req *CreateInstanceRequest,
) (*FormInstance, error) {
    // 1. Validate user has permission
    claims, _ := middleware.ClaimsFromContext(ctx)

    // 2. Get entity for the user
    entity, err := s.entityService.GetEntityByUserId(ctx, claims.UserID)
    if err != nil {
        return nil, fmt.Errorf("entity not found: %w", err)
    }

    // 3. Validate entity type
    if entity.EntityType != entity.ENTITY_TYPE_EMPLOYEE {
        return nil, errors.New("only employees can create form instances")
    }

    // 4. Get form definition
    form, err := s.formRepo.GetFormByID(ctx, req.FormID)
    if err != nil {
        return nil, err
    }

    // 5. Validate user has required role
    hasRole, err := s.userService.HasRole(ctx, claims.UserID, form.AllowedRoles)
    if err != nil || !hasRole {
        return nil, errors.New("user lacks required role")
    }

    // 6. Create instance
    instance := &FormInstance{
        FormID:       req.FormID,
        CreatedBy:    entity.ID,
        CurrentState: "draft",
        FieldValues:  req.FieldValues,
        Metadata: map[string]interface{}{
            "entity_type": entity.EntityType,
            "division_id": entity.DivisionID,
            "branch_id":   entity.BranchID,
        },
    }

    return s.formRepo.CreateInstance(ctx, instance)
}
```

**Call Flow:**
```
Client Request
    │
    ▼
FormBuilder.CreateInstance()
    │
    ├─► EntityService.GetEntityByUserId()
    │   └─► EntityRepository.GetByUserID() → Database
    │
    ├─► UserService.HasRole()
    │   └─► UserRepository.GetUserRoles() → Database
    │
    └─► FormRepository.CreateInstance() → Database
```

### Data Consistency Patterns

#### Pattern 1: Single Database Transaction

```go
// Atomic operation across modules
func (s *EmployeeService) OnboardEmployee(ctx context.Context, req *OnboardRequest) error {
    // Begin transaction
    tx, err := s.db.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)

    // 1. Create user (Identity module)
    user, err := s.userRepo.CreateWithTx(ctx, tx, &User{
        Username: req.Email,
        Email:    req.Email,
        FullName: req.FullName,
    })
    if err != nil {
        return err
    }

    // 2. Create employee (Personnel module)
    employee, err := s.employeeRepo.CreateWithTx(ctx, tx, &Employee{
        UserID:       user.ID,
        EmployeeCode: req.EmployeeCode,
        DivisionID:   req.DivisionID,
        BranchID:     req.BranchID,
    })
    if err != nil {
        return err
    }

    // 3. Create entity (Identity module)
    entity, err := s.entityRepo.CreateWithTx(ctx, tx, &Entity{
        UserID:        user.ID,
        ReferenceID:   employee.ID,
        EntityType:    "EMPLOYEE",
        DivisionID:    req.DivisionID,
        BranchID:      req.BranchID,
    })
    if err != nil {
        return err
    }

    // 4. Assign role (Identity module)
    err = s.entityRoleRepo.CreateWithTx(ctx, tx, &EntityRoleBinding{
        EntityID:    entity.ID,
        RoleID:      req.RoleID,
        DivisionID:  req.DivisionID,
        BranchID:    req.BranchID,
    })
    if err != nil {
        return err
    }

    // Commit transaction
    return tx.Commit(ctx)
}
```

#### Pattern 2: Saga Pattern (Future for Distributed Transactions)

```go
// For eventual consistency when microservices are needed
type OnboardSaga struct {
    userService     user.IUserService
    employeeService employee.IEmployeeService
    entityService   entity.IEntityService
}

func (s *OnboardSaga) Execute(ctx context.Context, req *OnboardRequest) error {
    var createdUser *User
    var createdEmployee *Employee
    var createdEntity *Entity

    // Compensation functions
    rollback := func() {
        if createdEntity != nil {
            s.entityService.Delete(ctx, createdEntity.ID)
        }
        if createdEmployee != nil {
            s.employeeService.Delete(ctx, createdEmployee.ID)
        }
        if createdUser != nil {
            s.userService.Delete(ctx, createdUser.ID)
        }
    }

    // Step 1: Create user
    user, err := s.userService.Create(ctx, &CreateUserRequest{...})
    if err != nil {
        return err
    }
    createdUser = user

    // Step 2: Create employee
    employee, err := s.employeeService.Create(ctx, &CreateEmployeeRequest{...})
    if err != nil {
        rollback()
        return err
    }
    createdEmployee = employee

    // Step 3: Create entity
    entity, err := s.entityService.Create(ctx, &CreateEntityRequest{...})
    if err != nil {
        rollback()
        return err
    }
    createdEntity = entity

    return nil
}
```

---

## Data Sharing Strategies

### Strategy 1: Service-to-Service (Recommended)

**When to Use:**
- Need business logic encapsulation
- Want to maintain module boundaries
- Planning for future microservices

```go
// DMS needs employee information
func (s *DMSService) GetDocumentWithOwner(ctx context.Context, docID string) (*DocumentWithOwner, error) {
    // Get document
    doc, err := s.docRepo.GetByID(ctx, docID)
    if err != nil {
        return nil, err
    }

    // Get entity through Entity Service
    entity, err := s.entityService.GetEntity(ctx, doc.OwnerEntityID)
    if err != nil {
        return nil, err
    }

    // If entity is employee, get employee details through Employee Service
    var employeeDetails *Employee
    if entity.EntityType == "EMPLOYEE" {
        employeeDetails, err = s.employeeService.GetEmployee(ctx, entity.ReferenceID)
        if err != nil {
            // Non-critical, log and continue
            log.Warn("failed to fetch employee details", "error", err)
        }
    }

    return &DocumentWithOwner{
        Document:  doc,
        Entity:    entity,
        Employee:  employeeDetails,
    }, nil
}
```

### Strategy 2: Direct Repository Access

**When to Use:**
- High-performance requirements
- Complex database joins
- Read-only operations

```go
// Optimized query with join
func (s *DMSService) ListDocumentsWithOwners(ctx context.Context, filters ListFilters) ([]*DocumentWithOwner, error) {
    // Raw SQL with join (using SQLC)
    rows, err := s.db.Query(ctx, `
        SELECT
            d.id, d.file_name, d.created_at,
            e.id as entity_id, e.entity_type,
            u.username, u.fullname
        FROM documents d
        INNER JOIN entities e ON e.id = d.owner_entity_id
        INNER JOIN users u ON u.uuid = e.user_id
        WHERE d.tenant_id = $1
          AND d.deleted_at IS NULL
        ORDER BY d.created_at DESC
        LIMIT $2 OFFSET $3
    `, filters.TenantID, filters.Limit, filters.Offset)

    // Map results...
}
```

### Strategy 3: Shared DTOs/Models

**When to Use:**
- Common data structures across modules
- Avoid duplication

```go
// Shared package: packages/models
package models

type EntityReference struct {
    ID         string
    Type       string
    UserID     string
    Username   string
    DivisionID string
    BranchID   string
}

// Used by multiple modules
package dms
import "p9e.in/ugcl/packages/models"

type Document struct {
    ID    string
    Owner models.EntityReference  // Shared model
}

package formbuilder
import "p9e.in/ugcl/packages/models"

type FormInstance struct {
    ID        string
    CreatedBy models.EntityReference  // Same shared model
}
```

### Strategy 4: CQRS (Command Query Responsibility Segregation)

**When to Use:**
- Different read/write patterns
- Performance optimization
- Complex queries

```go
// Write model (normalized, transactional)
type DocumentWriteModel struct {
    ID            string
    OwnerEntityID string
    TenantID      string
    FileName      string
}

// Read model (denormalized, optimized for queries)
type DocumentReadModel struct {
    ID              string
    FileName        string
    OwnerName       string
    OwnerType       string
    DivisionName    string
    BranchName      string
    Tags            []string
    CreatedAt       time.Time
}

// Separate repositories
type DocumentWriteRepository interface {
    Create(ctx context.Context, doc *DocumentWriteModel) error
    Update(ctx context.Context, doc *DocumentWriteModel) error
    Delete(ctx context.Context, id string) error
}

type DocumentReadRepository interface {
    GetByID(ctx context.Context, id string) (*DocumentReadModel, error)
    List(ctx context.Context, filters ListFilters) ([]*DocumentReadModel, error)
    Search(ctx context.Context, query string) ([]*DocumentReadModel, error)
}

// Materialized view for read model
CREATE MATERIALIZED VIEW documents_read_model AS
SELECT
    d.id,
    d.file_name,
    u.fullname as owner_name,
    e.entity_type as owner_type,
    div.name as division_name,
    br.name as branch_name,
    d.tags,
    d.created_at
FROM documents d
INNER JOIN entities e ON e.id = d.owner_entity_id
INNER JOIN users u ON u.uuid = e.user_id
LEFT JOIN divisions div ON div.id = e.division_id
LEFT JOIN branches br ON br.id = e.branch_id;

-- Refresh periodically
REFRESH MATERIALIZED VIEW CONCURRENTLY documents_read_model;
```

---

## Error Handling

### Error Propagation Strategy

```go
// Domain errors
package errors

var (
    ErrNotFound         = errors.New("resource not found")
    ErrUnauthorized     = errors.New("unauthorized")
    ErrInvalidInput     = errors.New("invalid input")
    ErrConflict         = errors.New("resource conflict")
    ErrInternal         = errors.New("internal server error")
)

// Wrapped errors with context
type AppError struct {
    Code    string
    Message string
    Cause   error
    Context map[string]interface{}
}

func (e *AppError) Error() string {
    return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
}

// Error creation helpers
func NotFound(resource string, id string) *AppError {
    return &AppError{
        Code:    "NOT_FOUND",
        Message: fmt.Sprintf("%s not found", resource),
        Context: map[string]interface{}{
            "resource": resource,
            "id":       id,
        },
    }
}

func Unauthorized(reason string) *AppError {
    return &AppError{
        Code:    "UNAUTHORIZED",
        Message: reason,
    }
}
```

### Connect Error Codes

```go
// Map application errors to Connect error codes
func mapErrorToConnect(err error) *connect.Error {
    var appErr *AppError
    if errors.As(err, &appErr) {
        switch appErr.Code {
        case "NOT_FOUND":
            return connect.NewError(connect.CodeNotFound, err)
        case "UNAUTHORIZED":
            return connect.NewError(connect.CodeUnauthenticated, err)
        case "FORBIDDEN":
            return connect.NewError(connect.CodePermissionDenied, err)
        case "INVALID_INPUT":
            return connect.NewError(connect.CodeInvalidArgument, err)
        case "CONFLICT":
            return connect.NewError(connect.CodeAlreadyExists, err)
        default:
            return connect.NewError(connect.CodeInternal, err)
        }
    }

    // Default to internal error
    return connect.NewError(connect.CodeInternal, err)
}

// Usage in handler
func (h *DMSHandler) GetDocument(
    ctx context.Context,
    req *connect.Request[dmsv1.GetDocumentRequest],
) (*connect.Response[dmsv1.Document], error) {
    doc, err := h.service.GetDocument(ctx, req.Msg.Id)
    if err != nil {
        return nil, mapErrorToConnect(err)
    }

    return connect.NewResponse(mappers.DocumentToProto(doc)), nil
}
```

### Error Logging and Monitoring

```go
// Structured error logging
func (s *DMSService) GetDocument(ctx context.Context, id string) (*Document, error) {
    doc, err := s.docRepo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            // Log warning for not found
            s.logger.Log(
                "level", "warn",
                "msg", "document not found",
                "document_id", id,
                "error", err,
            )
            return nil, NotFound("document", id)
        }

        // Log error for database failures
        s.logger.Log(
            "level", "error",
            "msg", "failed to fetch document",
            "document_id", id,
            "error", err,
        )
        return nil, fmt.Errorf("database error: %w", err)
    }

    return doc, nil
}
```

---

## API Versioning

### Versioning Strategy

**Protocol Buffers Package Versioning:**
```
dms/proto/dms.proto           → package dms.v1
dms/api/v1/dms.pb.go
dms/api/v1/dmsconnect/dms.connect.go

Future:
dms/proto/dms_v2.proto        → package dms.v2
dms/api/v2/dms.pb.go
dms/api/v2/dmsconnect/dms.connect.go
```

### URL Path Versioning

```
/dms.v1.DMSService/UploadDocument
/dms.v2.DMSService/UploadDocument
```

### Backward Compatibility

```protobuf
// v1 API
message Document {
  string id = 1;
  string file_name = 2;
  int64 size_bytes = 3;
}

// v2 API (backward compatible)
message Document {
  string id = 1;
  string file_name = 2;
  int64 size_bytes = 3;
  string checksum = 4;           // New field (optional)
  repeated string tags = 5;      // New field (optional)
  DocumentType type = 6;         // New field (optional)
}

// Non-breaking changes:
// - Adding new optional fields
// - Adding new RPC methods
// - Adding new enum values (with unknown handling)

// Breaking changes (require v2):
// - Removing fields
// - Changing field types
// - Renaming fields
// - Changing field numbers
```

### Deprecation Process

```protobuf
service DMSService {
  // Deprecated: Use UploadDocumentV2 instead
  rpc UploadDocument(UploadDocumentRequest) returns (Document) {
    option deprecated = true;
  }

  rpc UploadDocumentV2(UploadDocumentV2Request) returns (DocumentV2);
}
```

---

## Event-Driven Patterns (Future)

### Event Bus Architecture

```go
// Event bus interface
type EventBus interface {
    Publish(topic string, event interface{}) error
    Subscribe(topic string, handler EventHandler) error
    Unsubscribe(topic string, handler EventHandler) error
}

type EventHandler func(event interface{}) error

// Event definitions
type DomainEvent struct {
    ID        string
    Type      string
    Timestamp time.Time
    TenantID  string
    EntityID  string
    Payload   interface{}
}

// Example events
type DocumentUploadedEvent struct {
    DocumentID string
    FileName   string
    MimeType   string
    SizeBytes  int64
}

type FormSubmittedEvent struct {
    InstanceID   string
    FormID       string
    SubmittedBy  string
    CurrentState string
}

type EmployeeOnboardedEvent struct {
    EmployeeID string
    UserID     string
    EntityID   string
    DivisionID string
}
```

### Event Publisher

```go
func (s *DMSService) UploadDocument(ctx context.Context, req *UploadRequest) (*Document, error) {
    // Upload document
    doc, err := s.docRepo.Create(ctx, document)
    if err != nil {
        return nil, err
    }

    // Publish event (non-blocking)
    go func() {
        err := s.eventBus.Publish("document.uploaded", &DomainEvent{
            ID:        uuid.New().String(),
            Type:      "document.uploaded",
            Timestamp: time.Now(),
            TenantID:  doc.TenantID,
            EntityID:  doc.OwnerEntityID,
            Payload: &DocumentUploadedEvent{
                DocumentID: doc.ID,
                FileName:   doc.FileName,
                MimeType:   doc.MimeType,
                SizeBytes:  doc.SizeBytes,
            },
        })
        if err != nil {
            s.logger.Log("level", "error", "msg", "failed to publish event", "error", err)
        }
    }()

    return doc, nil
}
```

### Event Subscribers

```go
// OCR service subscribes to document uploads
func (s *OCRService) RegisterEventHandlers(eventBus EventBus) {
    eventBus.Subscribe("document.uploaded", s.handleDocumentUploaded)
}

func (s *OCRService) handleDocumentUploaded(event interface{}) error {
    domainEvent := event.(*DomainEvent)
    payload := domainEvent.Payload.(*DocumentUploadedEvent)

    // Process only PDFs and images
    if !s.shouldProcessOCR(payload.MimeType) {
        return nil
    }

    // Async OCR processing
    return s.ProcessDocument(context.Background(), payload.DocumentID)
}

// Notification service subscribes to form submissions
func (s *NotificationService) RegisterEventHandlers(eventBus EventBus) {
    eventBus.Subscribe("form.submitted", s.handleFormSubmitted)
}

func (s *NotificationService) handleFormSubmitted(event interface{}) error {
    domainEvent := event.(*DomainEvent)
    payload := domainEvent.Payload.(*FormSubmittedEvent)

    // Send notification to approvers
    return s.NotifyApprovers(context.Background(), payload.InstanceID)
}
```

---

## Circuit Breaker & Resilience (Future)

### Circuit Breaker Pattern

```go
import "github.com/sony/gobreaker"

// Circuit breaker configuration
type CircuitBreakerConfig struct {
    Name          string
    MaxRequests   uint32
    Interval      time.Duration
    Timeout       time.Duration
    ReadyToTrip   func(counts gobreaker.Counts) bool
}

// Wrap service calls with circuit breaker
func (s *DMSService) GetDocumentWithResilience(ctx context.Context, id string) (*Document, error) {
    result, err := s.circuitBreaker.Execute(func() (interface{}, error) {
        return s.docRepo.GetByID(ctx, id)
    })

    if err != nil {
        // Circuit open or execution failed
        return nil, err
    }

    return result.(*Document), nil
}

// Circuit breaker configuration
cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
    Name:        "DocumentService",
    MaxRequests: 3,
    Interval:    time.Second * 10,
    Timeout:     time.Second * 30,
    ReadyToTrip: func(counts gobreaker.Counts) bool {
        failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
        return counts.Requests >= 3 && failureRatio >= 0.6
    },
})
```

### Retry Pattern

```go
import "github.com/cenkalti/backoff/v4"

func (s *DMSService) UploadToMinIOWithRetry(ctx context.Context, doc *Document) error {
    operation := func() error {
        return s.minioClient.Upload(ctx, doc.StoragePath, doc.FileData)
    }

    // Exponential backoff retry
    backoffConfig := backoff.NewExponentialBackOff()
    backoffConfig.MaxElapsedTime = 30 * time.Second

    return backoff.Retry(operation, backoff.WithContext(backoffConfig, ctx))
}
```

### Timeout Pattern

```go
func (s *DMSService) ProcessDocumentWithTimeout(ctx context.Context, docID string) error {
    // Create timeout context
    timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    // Process with timeout
    resultChan := make(chan error, 1)
    go func() {
        resultChan <- s.ocrService.ExtractText(timeoutCtx, docID)
    }()

    select {
    case err := <-resultChan:
        return err
    case <-timeoutCtx.Done():
        return fmt.Errorf("document processing timeout: %w", timeoutCtx.Err())
    }
}
```

---

## Service Discovery

### Static Configuration (Current)

```yaml
# configs.yaml
services:
  dms:
    endpoint: "http://localhost:8080"
    timeout: 30s

  entity:
    endpoint: "http://localhost:8080"  # Same process
    timeout: 10s

  notification:
    endpoint: "http://localhost:8080"
    timeout: 15s
```

### Dynamic Service Discovery (Future with Kubernetes)

```go
// Kubernetes service discovery
import (
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
)

type ServiceDiscovery struct {
    clientset *kubernetes.Clientset
}

func NewServiceDiscovery() (*ServiceDiscovery, error) {
    config, err := rest.InClusterConfig()
    if err != nil {
        return nil, err
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        return nil, err
    }

    return &ServiceDiscovery{clientset: clientset}, nil
}

func (sd *ServiceDiscovery) GetServiceEndpoint(serviceName string) (string, error) {
    service, err := sd.clientset.CoreV1().Services("default").Get(
        context.Background(),
        serviceName,
        metav1.GetOptions{},
    )
    if err != nil {
        return "", err
    }

    return fmt.Sprintf("http://%s:%d", service.Spec.ClusterIP, service.Spec.Ports[0].Port), nil
}
```

---

## Conclusion

The UGCL Backend v2 inter-module communication architecture is designed for:

- **Simplicity:** Direct function calls within a single process
- **Performance:** No network overhead for module-to-module communication
- **Type Safety:** Compile-time dependency checking via FX
- **Maintainability:** Clear module boundaries and interfaces
- **Extensibility:** Easy to extract modules to microservices in the future
- **Resilience:** Patterns for error handling, retries, and circuit breakers

The modular monolith approach provides the best of both worlds: development simplicity with organizational benefits of microservices architecture.

---

**Document Version:** 2.0
**Lines:** 1500+
**Generated:** 2025-10-06
**Maintained By:** UGCL Architecture Team
