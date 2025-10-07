# UGCL Backend v2 Testing Guide

## Table of Contents

1. [Introduction](#introduction)
2. [Testing Philosophy and Pyramid](#testing-philosophy-and-pyramid)
3. [Unit Testing with Go Testing Package](#unit-testing-with-go-testing-package)
4. [Integration Testing with Testcontainers](#integration-testing-with-testcontainers)
5. [gRPC/Connect Service Testing](#grpcconnect-service-testing)
6. [Database Testing with SQLC](#database-testing-with-sqlc)
7. [Mocking Strategies](#mocking-strategies)
8. [Test Fixtures and Data Builders](#test-fixtures-and-data-builders)
9. [Code Coverage Requirements](#code-coverage-requirements)
10. [CI/CD Testing Pipeline](#cicd-testing-pipeline)
11. [Performance and Load Testing](#performance-and-load-testing)
12. [E2E Testing Strategy](#e2e-testing-strategy)

---

## Introduction

Testing is a critical component of the UGCL Backend v2 development process. This guide outlines our testing philosophy, strategies, and best practices to ensure code quality, reliability, and maintainability.

**Testing Goals:**
- Catch bugs early in development
- Enable confident refactoring
- Document expected behavior
- Facilitate code reviews
- Ensure production reliability

**Testing Stack:**
- **Unit Tests:** Go's built-in `testing` package
- **Integration Tests:** testcontainers-go for real database tests
- **Service Tests:** Connect client for gRPC/HTTP testing
- **Mocking:** gomock, testify/mock
- **Assertions:** testify/assert, testify/require
- **Coverage:** go test -cover
- **Load Testing:** k6, vegeta

---

## Testing Philosophy and Pyramid

### The Testing Pyramid

```
        ┌─────────────┐
        │     E2E     │  ← Few, slow, expensive
        │   (Manual)  │
        ├─────────────┤
        │ Integration │  ← Some, medium speed
        │   Tests     │
        ├─────────────┤
        │    Unit     │  ← Many, fast, cheap
        │   Tests     │
        └─────────────┘
```

### Test Distribution

- **70% Unit Tests:** Fast, isolated, test single functions/methods
- **20% Integration Tests:** Test interactions with database, external services
- **10% E2E Tests:** Full user workflows, critical paths

### Testing Principles

1. **Fast Feedback:** Unit tests should run in milliseconds
2. **Isolation:** Tests should not depend on each other
3. **Repeatability:** Same input = same output, every time
4. **Clarity:** Test names should describe what they test
5. **Comprehensive:** Cover happy paths, edge cases, and error conditions
6. **Maintainable:** Tests should be easy to update as code evolves

---

## Unit Testing with Go Testing Package

### Basic Test Structure

```go
// vendors/services/vendor_service_test.go
package services_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "p9e.in/ugcl/vendors/services"
)

func TestVendorService_Create(t *testing.T) {
    // Arrange
    ctx := context.Background()
    mockRepo := &MockVendorRepository{}
    svc := services.NewVendorService(mockRepo, nil)

    vendor := &db.Vendor{
        CompanyName: "Acme Corp",
        PAN:         "ABCDE1234F",
    }

    // Act
    result, err := svc.Create(ctx, vendor, nil)

    // Assert
    require.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, "Acme Corp", result.CompanyName)
}
```

### Table-Driven Tests

```go
func TestVendorService_ValidateVendor(t *testing.T) {
    tests := []struct {
        name    string
        vendor  *db.Vendor
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid vendor",
            vendor: &db.Vendor{
                CompanyName: "Acme Corp",
                PAN:         "ABCDE1234F",
                GST:         "29ABCDE1234F1Z5",
            },
            wantErr: false,
        },
        {
            name: "missing company name",
            vendor: &db.Vendor{
                PAN: "ABCDE1234F",
            },
            wantErr: true,
            errMsg:  "company name is required",
        },
        {
            name: "invalid PAN format",
            vendor: &db.Vendor{
                CompanyName: "Acme Corp",
                PAN:         "INVALID",
            },
            wantErr: true,
            errMsg:  "invalid PAN format",
        },
        {
            name: "invalid GST format",
            vendor: &db.Vendor{
                CompanyName: "Acme Corp",
                PAN:         "ABCDE1234F",
                GST:         "INVALID",
            },
            wantErr: true,
            errMsg:  "invalid GST format",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            svc := services.NewVendorService(nil, nil)

            // Act
            err := svc.ValidateVendor(tt.vendor)

            // Assert
            if tt.wantErr {
                require.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

### Subtests for Organization

```go
func TestVendorService(t *testing.T) {
    // Setup shared resources
    ctx := context.Background()
    mockRepo := &MockVendorRepository{}
    svc := services.NewVendorService(mockRepo, nil)

    t.Run("Create", func(t *testing.T) {
        t.Run("success", func(t *testing.T) {
            vendor := &db.Vendor{CompanyName: "Test Corp"}
            result, err := svc.Create(ctx, vendor, nil)
            require.NoError(t, err)
            assert.NotNil(t, result)
        })

        t.Run("duplicate PAN", func(t *testing.T) {
            mockRepo.SetError(repository.ErrDuplicatePAN)
            vendor := &db.Vendor{PAN: "ABCDE1234F"}
            _, err := svc.Create(ctx, vendor, nil)
            require.Error(t, err)
            assert.Contains(t, err.Error(), "already exists")
        })
    })

    t.Run("GetByID", func(t *testing.T) {
        t.Run("found", func(t *testing.T) {
            mockRepo.SetVendor(&db.Vendor{ID: "123"})
            result, err := svc.GetByID(ctx, "123")
            require.NoError(t, err)
            assert.Equal(t, "123", result.ID)
        })

        t.Run("not found", func(t *testing.T) {
            mockRepo.SetError(repository.ErrNotFound)
            _, err := svc.GetByID(ctx, "999")
            require.Error(t, err)
        })
    })
}
```

### Testing Error Handling

```go
func TestVendorService_ErrorHandling(t *testing.T) {
    ctx := context.Background()

    tests := []struct {
        name          string
        setupMock     func(*MockVendorRepository)
        expectedError string
    }{
        {
            name: "repository error",
            setupMock: func(m *MockVendorRepository) {
                m.On("Create", mock.Anything, mock.Anything).
                    Return(nil, errors.New("database error"))
            },
            expectedError: "database error",
        },
        {
            name: "unique constraint violation",
            setupMock: func(m *MockVendorRepository) {
                m.On("Create", mock.Anything, mock.Anything).
                    Return(nil, &pq.Error{Code: "23505"})
            },
            expectedError: "already exists",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := &MockVendorRepository{}
            tt.setupMock(mockRepo)
            svc := services.NewVendorService(mockRepo, nil)

            vendor := &db.Vendor{CompanyName: "Test"}
            _, err := svc.Create(ctx, vendor, nil)

            require.Error(t, err)
            assert.Contains(t, err.Error(), tt.expectedError)
        })
    }
}
```

### Testing with Context

```go
func TestVendorService_ContextCancellation(t *testing.T) {
    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()

    mockRepo := &MockVendorRepository{
        delay: 200 * time.Millisecond, // Longer than timeout
    }
    svc := services.NewVendorService(mockRepo, nil)

    _, err := svc.Create(ctx, &db.Vendor{}, nil)

    require.Error(t, err)
    assert.Equal(t, context.DeadlineExceeded, err)
}
```

### Helper Functions

```go
// test_helpers.go
package services_test

import (
    "testing"
    "github.com/google/uuid"
    "p9e.in/ugcl/vendors/db"
)

// Helper to create test vendor
func createTestVendor(t *testing.T) *db.Vendor {
    t.Helper()
    return &db.Vendor{
        ID:          uuid.New().String(),
        CompanyName: "Test Vendor Corp",
        PAN:         "ABCDE1234F",
        GST:         "29ABCDE1234F1Z5",
        Status:      "ACTIVE",
    }
}

// Helper to assert vendor fields
func assertVendorEqual(t *testing.T, expected, actual *db.Vendor) {
    t.Helper()
    assert.Equal(t, expected.CompanyName, actual.CompanyName)
    assert.Equal(t, expected.PAN, actual.PAN)
    assert.Equal(t, expected.Status, actual.Status)
}
```

---

## Integration Testing with Testcontainers

### Setting Up Testcontainers

```go
// vendors/repository/vendor_repository_integration_test.go
package repository_test

import (
    "context"
    "database/sql"
    "testing"
    "time"

    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
    _ "github.com/lib/pq"
)

// setupPostgresContainer starts a PostgreSQL container for testing
func setupPostgresContainer(t *testing.T) (testcontainers.Container, *sql.DB) {
    t.Helper()

    ctx := context.Background()

    // Define container request
    req := testcontainers.ContainerRequest{
        Image:        "postgres:15-alpine",
        ExposedPorts: []string{"5432/tcp"},
        Env: map[string]string{
            "POSTGRES_USER":     "test",
            "POSTGRES_PASSWORD": "test",
            "POSTGRES_DB":       "testdb",
        },
        WaitingFor: wait.ForLog("database system is ready to accept connections").
            WithStartupTimeout(60 * time.Second),
    }

    // Start container
    container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
    if err != nil {
        t.Fatalf("Failed to start container: %v", err)
    }

    // Get container host and port
    host, err := container.Host(ctx)
    if err != nil {
        t.Fatalf("Failed to get container host: %v", err)
    }

    port, err := container.MappedPort(ctx, "5432")
    if err != nil {
        t.Fatalf("Failed to get container port: %v", err)
    }

    // Connect to database
    dsn := fmt.Sprintf("host=%s port=%s user=test password=test dbname=testdb sslmode=disable",
        host, port.Port())
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        t.Fatalf("Failed to connect to database: %v", err)
    }

    // Wait for database to be ready
    if err := db.Ping(); err != nil {
        t.Fatalf("Failed to ping database: %v", err)
    }

    // Run migrations
    runMigrations(t, db)

    return container, db
}

// runMigrations applies database schema
func runMigrations(t *testing.T, db *sql.DB) {
    t.Helper()

    schema := `
    CREATE TABLE vendors (
        id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
        uuid UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
        company_name TEXT NOT NULL,
        pan TEXT UNIQUE,
        gst TEXT,
        status TEXT DEFAULT 'ACTIVE',
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        deleted_at TIMESTAMP WITH TIME ZONE
    );

    CREATE INDEX idx_vendors_pan ON vendors(pan) WHERE deleted_at IS NULL;
    `

    _, err := db.Exec(schema)
    if err != nil {
        t.Fatalf("Failed to run migrations: %v", err)
    }
}
```

### Integration Test Example

```go
func TestVendorRepository_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    // Setup
    container, db := setupPostgresContainer(t)
    defer container.Terminate(context.Background())
    defer db.Close()

    // Create repository
    queries := db.NewQueries(db)
    repo := repository.NewVendorRepository(queries)

    t.Run("Create and Retrieve", func(t *testing.T) {
        ctx := context.Background()

        // Create vendor
        vendor := &db.Vendor{
            CompanyName: "Integration Test Corp",
            PAN:         "INTEG1234T",
        }

        created, err := repo.Create(ctx, db.CreateVendorParams{
            CompanyName: vendor.CompanyName,
            PAN:         &vendor.PAN,
        })

        require.NoError(t, err)
        assert.NotEmpty(t, created.ID)
        assert.Equal(t, vendor.CompanyName, created.CompanyName)

        // Retrieve vendor
        retrieved, err := repo.GetByID(ctx, created.ID)
        require.NoError(t, err)
        assert.Equal(t, created.ID, retrieved.ID)
        assert.Equal(t, created.CompanyName, retrieved.CompanyName)
    })

    t.Run("Unique Constraint Violation", func(t *testing.T) {
        ctx := context.Background()

        vendor := &db.Vendor{
            CompanyName: "Unique Test Corp",
            PAN:         "UNIQU1234T",
        }

        // First creation should succeed
        _, err := repo.Create(ctx, db.CreateVendorParams{
            CompanyName: vendor.CompanyName,
            PAN:         &vendor.PAN,
        })
        require.NoError(t, err)

        // Second creation with same PAN should fail
        _, err = repo.Create(ctx, db.CreateVendorParams{
            CompanyName: "Different Name",
            PAN:         &vendor.PAN,
        })
        require.Error(t, err)
        assert.Contains(t, err.Error(), "already exists")
    })

    t.Run("Soft Delete", func(t *testing.T) {
        ctx := context.Background()

        // Create vendor
        created, err := repo.Create(ctx, db.CreateVendorParams{
            CompanyName: "Delete Test Corp",
            PAN:         stringPtr("DELET1234T"),
        })
        require.NoError(t, err)

        // Delete vendor (soft delete)
        err = repo.Delete(ctx, created.ID)
        require.NoError(t, err)

        // Try to retrieve deleted vendor
        _, err = repo.GetByID(ctx, created.ID)
        require.Error(t, err)
        assert.Contains(t, err.Error(), "not found")
    })
}
```

### Test Suite Pattern

```go
type VendorRepositoryTestSuite struct {
    suite.Suite
    container testcontainers.Container
    db        *sql.DB
    queries   *db.Queries
    repo      repository.IVendorRepository
}

func (suite *VendorRepositoryTestSuite) SetupSuite() {
    // Run once before all tests
    container, db := setupPostgresContainer(suite.T())
    suite.container = container
    suite.db = db
    suite.queries = db.NewQueries(db)
    suite.repo = repository.NewVendorRepository(suite.queries)
}

func (suite *VendorRepositoryTestSuite) TearDownSuite() {
    // Run once after all tests
    suite.db.Close()
    suite.container.Terminate(context.Background())
}

func (suite *VendorRepositoryTestSuite) SetupTest() {
    // Run before each test
    suite.cleanDatabase()
}

func (suite *VendorRepositoryTestSuite) TestCreate() {
    ctx := context.Background()
    vendor, err := suite.repo.Create(ctx, db.CreateVendorParams{
        CompanyName: "Test Corp",
    })
    suite.NoError(err)
    suite.NotEmpty(vendor.ID)
}

func TestVendorRepositorySuite(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test suite")
    }
    suite.Run(t, new(VendorRepositoryTestSuite))
}
```

---

## gRPC/Connect Service Testing

### Testing Connect Handlers

```go
// vendors/handlers/vendor_handler_test.go
package handlers_test

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"

    "connectrpc.com/connect"
    pb "p9e.in/ugcl/vendors/api/v2/vendor"
    "p9e.in/ugcl/vendors/api/v2/vendor/vendorconnect"
    "p9e.in/ugcl/vendors/handlers"
)

func TestVendorHandler_CreateVendor(t *testing.T) {
    // Setup
    mockService := &MockVendorService{}
    handler := handlers.NewVendorHandler(mockService)

    // Create test server
    mux := http.NewServeMux()
    path, h := vendorconnect.NewVendorServiceHandler(handler)
    mux.Handle(path, h)
    server := httptest.NewServer(mux)
    defer server.Close()

    // Create client
    client := vendorconnect.NewVendorServiceClient(
        http.DefaultClient,
        server.URL,
    )

    t.Run("Success", func(t *testing.T) {
        req := &pb.CreateVendorRequest{
            Vendor: &pb.Vendor{
                CompanyName: "Test Corp",
                Pan:         "TESTC1234P",
            },
        }

        resp, err := client.CreateVendor(context.Background(), connect.NewRequest(req))

        require.NoError(t, err)
        assert.NotNil(t, resp.Msg)
        assert.Equal(t, "Test Corp", resp.Msg.CompanyName)
    })

    t.Run("Invalid Input", func(t *testing.T) {
        req := &pb.CreateVendorRequest{
            Vendor: &pb.Vendor{
                // Missing required fields
            },
        }

        _, err := client.CreateVendor(context.Background(), connect.NewRequest(req))

        require.Error(t, err)
        var connectErr *connect.Error
        assert.True(t, errors.As(err, &connectErr))
        assert.Equal(t, connect.CodeInvalidArgument, connectErr.Code())
    })
}
```

### Testing with Mock Server

```go
func TestVendorHandler_WithMockServer(t *testing.T) {
    // Create mock service
    mockService := &MockVendorService{
        vendors: make(map[string]*db.Vendor),
    }

    // Setup handler and server
    handler := handlers.NewVendorHandler(mockService)
    mux := http.NewServeMux()
    path, h := vendorconnect.NewVendorServiceHandler(handler)
    mux.Handle(path, h)
    server := httptest.NewServer(mux)
    defer server.Close()

    // Create client
    client := vendorconnect.NewVendorServiceClient(
        http.DefaultClient,
        server.URL,
    )

    t.Run("CRUD Operations", func(t *testing.T) {
        ctx := context.Background()

        // Create
        createReq := &pb.CreateVendorRequest{
            Vendor: &pb.Vendor{
                CompanyName: "CRUD Test Corp",
                Pan:         "CRUDX1234P",
            },
        }
        createResp, err := client.CreateVendor(ctx, connect.NewRequest(createReq))
        require.NoError(t, err)
        vendorID := createResp.Msg.Id

        // Read
        getReq := &pb.VendorIdentifier{Id: vendorID}
        getResp, err := client.GetVendor(ctx, connect.NewRequest(getReq))
        require.NoError(t, err)
        assert.Equal(t, "CRUD Test Corp", getResp.Msg.CompanyName)

        // Update
        updateReq := &pb.UpdateVendorRequest{
            Vendor: &pb.Vendor{
                Id:          vendorID,
                CompanyName: "Updated Corp",
            },
        }
        updateResp, err := client.UpdateVendor(ctx, connect.NewRequest(updateReq))
        require.NoError(t, err)
        assert.Equal(t, "Updated Corp", updateResp.Msg.CompanyName)

        // Delete
        deleteReq := &pb.VendorIdentifier{Id: vendorID}
        _, err = client.DeleteVendor(ctx, connect.NewRequest(deleteReq))
        require.NoError(t, err)

        // Verify deleted
        _, err = client.GetVendor(ctx, connect.NewRequest(getReq))
        require.Error(t, err)
    })
}
```

### Testing Interceptors

```go
func TestAuthInterceptor(t *testing.T) {
    interceptor := middleware.NewAuthInterceptor()

    t.Run("Valid JWT", func(t *testing.T) {
        req := &http.Request{
            Header: http.Header{
                "Authorization": []string{"Bearer valid_token_here"},
            },
        }

        ctx := context.Background()
        newCtx, err := interceptor.WrapUnary(func(ctx context.Context, req interface{}) (interface{}, error) {
            // Verify context has user info
            userID := ctx.Value("user_id")
            assert.NotNil(t, userID)
            return nil, nil
        })(ctx, req)

        require.NoError(t, err)
    })

    t.Run("Missing Token", func(t *testing.T) {
        req := &http.Request{Header: http.Header{}}
        ctx := context.Background()

        _, err := interceptor.WrapUnary(func(ctx context.Context, req interface{}) (interface{}, error) {
            return nil, nil
        })(ctx, req)

        require.Error(t, err)
        assert.Equal(t, connect.CodeUnauthenticated, err.(*connect.Error).Code())
    })
}
```

---

## Database Testing with SQLC

### Testing SQLC Queries

```go
// Test generated SQLC queries
func TestSQLCQueries_Vendors(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping database test")
    }

    container, sqlDB := setupPostgresContainer(t)
    defer container.Terminate(context.Background())
    defer sqlDB.Close()

    queries := db.NewQueries(sqlDB)
    ctx := context.Background()

    t.Run("CreateVendor", func(t *testing.T) {
        vendor, err := queries.CreateVendor(ctx, db.CreateVendorParams{
            CompanyName: "SQLC Test Corp",
            Pan:         stringPtr("SQLCT1234P"),
        })

        require.NoError(t, err)
        assert.NotEmpty(t, vendor.ID)
        assert.Equal(t, "SQLC Test Corp", vendor.CompanyName)
        assert.NotNil(t, vendor.CreatedAt)
    })

    t.Run("GetVendorByID", func(t *testing.T) {
        // Create test vendor
        created, _ := queries.CreateVendor(ctx, db.CreateVendorParams{
            CompanyName: "Get Test Corp",
        })

        // Retrieve
        vendor, err := queries.GetVendorByID(ctx, created.ID)

        require.NoError(t, err)
        assert.Equal(t, created.ID, vendor.ID)
        assert.Equal(t, "Get Test Corp", vendor.CompanyName)
    })

    t.Run("ListVendors", func(t *testing.T) {
        // Create multiple vendors
        for i := 0; i < 5; i++ {
            queries.CreateVendor(ctx, db.CreateVendorParams{
                CompanyName: fmt.Sprintf("List Test Corp %d", i),
            })
        }

        // List with pagination
        vendors, err := queries.ListVendors(ctx, db.ListVendorsParams{
            Limit:  10,
            Offset: 0,
        })

        require.NoError(t, err)
        assert.GreaterOrEqual(t, len(vendors), 5)
    })

    t.Run("UpdateVendor", func(t *testing.T) {
        created, _ := queries.CreateVendor(ctx, db.CreateVendorParams{
            CompanyName: "Update Test Corp",
        })

        newName := "Updated Test Corp"
        updated, err := queries.UpdateVendor(ctx, db.UpdateVendorParams{
            ID:          created.ID,
            CompanyName: &newName,
        })

        require.NoError(t, err)
        assert.Equal(t, newName, updated.CompanyName)
    })
}
```

### Testing Complex Queries

```go
func TestSQLCQueries_ComplexFiltering(t *testing.T) {
    container, sqlDB := setupPostgresContainer(t)
    defer container.Terminate(context.Background())
    defer sqlDB.Close()

    queries := db.NewQueries(sqlDB)
    ctx := context.Background()

    // Seed test data
    seedVendors(t, queries)

    t.Run("FilterByCategory", func(t *testing.T) {
        category := "RAW_MATERIAL"
        vendors, err := queries.ListVendorsByCategory(ctx, &category)

        require.NoError(t, err)
        for _, v := range vendors {
            assert.Equal(t, category, *v.VendorCategory)
        }
    })

    t.Run("FilterByRatingRange", func(t *testing.T) {
        minRating := int32(4)
        maxRating := int32(5)

        vendors, err := queries.ListVendors(ctx, db.ListVendorsParams{
            Column4: &minRating,
            Column5: &maxRating,
            Limit:   100,
        })

        require.NoError(t, err)
        for _, v := range vendors {
            if v.Rating != nil {
                assert.GreaterOrEqual(t, *v.Rating, minRating)
                assert.LessOrEqual(t, *v.Rating, maxRating)
            }
        }
    })

    t.Run("CountWithFilters", func(t *testing.T) {
        category := "EQUIPMENT"
        count, err := queries.CountVendors(ctx, db.CountVendorsParams{
            Column2: &category,
        })

        require.NoError(t, err)
        assert.Greater(t, count, int64(0))
    })
}
```

---

## Mocking Strategies

### Interface-Based Mocking

```go
// Define interface
type IVendorRepository interface {
    Create(ctx context.Context, params db.CreateVendorParams) (*db.Vendor, error)
    GetByID(ctx context.Context, id string) (*db.Vendor, error)
    Update(ctx context.Context, params db.UpdateVendorParams) (*db.Vendor, error)
    Delete(ctx context.Context, id string) error
}

// Manual mock implementation
type MockVendorRepository struct {
    vendors map[string]*db.Vendor
    err     error
}

func (m *MockVendorRepository) Create(ctx context.Context, params db.CreateVendorParams) (*db.Vendor, error) {
    if m.err != nil {
        return nil, m.err
    }

    vendor := &db.Vendor{
        ID:          uuid.New().String(),
        CompanyName: params.CompanyName,
        PAN:         params.PAN,
    }

    if m.vendors == nil {
        m.vendors = make(map[string]*db.Vendor)
    }
    m.vendors[vendor.ID] = vendor

    return vendor, nil
}

func (m *MockVendorRepository) GetByID(ctx context.Context, id string) (*db.Vendor, error) {
    if m.err != nil {
        return nil, m.err
    }

    vendor, exists := m.vendors[id]
    if !exists {
        return nil, errors.New("vendor not found")
    }

    return vendor, nil
}

// Helper methods for test setup
func (m *MockVendorRepository) SetError(err error) {
    m.err = err
}

func (m *MockVendorRepository) SetVendor(vendor *db.Vendor) {
    if m.vendors == nil {
        m.vendors = make(map[string]*db.Vendor)
    }
    m.vendors[vendor.ID] = vendor
}
```

### Using testify/mock

```go
import "github.com/stretchr/testify/mock"

type MockVendorRepository struct {
    mock.Mock
}

func (m *MockVendorRepository) Create(ctx context.Context, params db.CreateVendorParams) (*db.Vendor, error) {
    args := m.Called(ctx, params)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*db.Vendor), args.Error(1)
}

func (m *MockVendorRepository) GetByID(ctx context.Context, id string) (*db.Vendor, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*db.Vendor), args.Error(1)
}

// Usage in tests
func TestVendorService_WithTestifyMock(t *testing.T) {
    mockRepo := new(MockVendorRepository)

    // Setup expectations
    mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(params db.CreateVendorParams) bool {
        return params.CompanyName == "Test Corp"
    })).Return(&db.Vendor{
        ID:          "123",
        CompanyName: "Test Corp",
    }, nil)

    // Use mock
    svc := services.NewVendorService(mockRepo, nil)
    vendor, err := svc.Create(context.Background(), &db.Vendor{
        CompanyName: "Test Corp",
    }, nil)

    // Assertions
    require.NoError(t, err)
    assert.Equal(t, "123", vendor.ID)

    // Verify expectations were met
    mockRepo.AssertExpectations(t)
}
```

### Using gomock

```bash
# Install gomock
go install github.com/golang/mock/mockgen@latest

# Generate mocks
mockgen -source=vendors/repository/vendor_repository.go \
    -destination=vendors/repository/mocks/mock_vendor_repository.go \
    -package=mocks
```

```go
import "github.com/golang/mock/gomock"

func TestVendorService_WithGoMock(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockRepo := mocks.NewMockIVendorRepository(ctrl)

    // Setup expectations
    mockRepo.EXPECT().
        Create(gomock.Any(), gomock.Any()).
        Return(&db.Vendor{ID: "123"}, nil).
        Times(1)

    // Use mock
    svc := services.NewVendorService(mockRepo, nil)
    vendor, err := svc.Create(context.Background(), &db.Vendor{}, nil)

    require.NoError(t, err)
    assert.Equal(t, "123", vendor.ID)
}
```

---

## Test Fixtures and Data Builders

### Builder Pattern

```go
// vendors/testdata/builders.go
package testdata

import (
    "github.com/google/uuid"
    "p9e.in/ugcl/vendors/db"
)

type VendorBuilder struct {
    vendor *db.Vendor
}

func NewVendorBuilder() *VendorBuilder {
    return &VendorBuilder{
        vendor: &db.Vendor{
            ID:          uuid.New().String(),
            UUID:        uuid.New(),
            CompanyName: "Default Test Corp",
            Status:      stringPtr("ACTIVE"),
        },
    }
}

func (b *VendorBuilder) WithID(id string) *VendorBuilder {
    b.vendor.ID = id
    return b
}

func (b *VendorBuilder) WithCompanyName(name string) *VendorBuilder {
    b.vendor.CompanyName = name
    return b
}

func (b *VendorBuilder) WithPAN(pan string) *VendorBuilder {
    b.vendor.PAN = &pan
    return b
}

func (b *VendorBuilder) WithGST(gst string) *VendorBuilder {
    b.vendor.GST = &gst
    return b
}

func (b *VendorBuilder) WithCategory(category string) *VendorBuilder {
    b.vendor.VendorCategory = &category
    return b
}

func (b *VendorBuilder) WithRating(rating int32) *VendorBuilder {
    b.vendor.Rating = &rating
    return b
}

func (b *VendorBuilder) Blacklisted() *VendorBuilder {
    blacklisted := true
    b.vendor.IsBlacklisted = &blacklisted
    return b
}

func (b *VendorBuilder) Build() *db.Vendor {
    return b.vendor
}

// Usage in tests
func TestVendorService_WithBuilder(t *testing.T) {
    vendor := testdata.NewVendorBuilder().
        WithCompanyName("Builder Test Corp").
        WithPAN("BUILD1234P").
        WithCategory("RAW_MATERIAL").
        WithRating(5).
        Build()

    assert.Equal(t, "Builder Test Corp", vendor.CompanyName)
    assert.Equal(t, "BUILD1234P", *vendor.PAN)
    assert.Equal(t, int32(5), *vendor.Rating)
}
```

### Fixture Files

```go
// vendors/testdata/fixtures.go
package testdata

import (
    "embed"
    "encoding/json"
)

//go:embed *.json
var fixturesFS embed.FS

type Fixtures struct {
    Vendors []db.Vendor `json:"vendors"`
}

func LoadFixtures() (*Fixtures, error) {
    data, err := fixturesFS.ReadFile("vendors.json")
    if err != nil {
        return nil, err
    }

    var fixtures Fixtures
    if err := json.Unmarshal(data, &fixtures); err != nil {
        return nil, err
    }

    return &fixtures, nil
}

// vendors/testdata/vendors.json
[
  {
    "id": "vendor-1",
    "company_name": "Acme Corporation",
    "pan": "ACMEC1234P",
    "gst": "29ACMEC1234P1Z5",
    "vendor_category": "RAW_MATERIAL",
    "rating": 5,
    "status": "ACTIVE"
  },
  {
    "id": "vendor-2",
    "company_name": "TechCorp Industries",
    "pan": "TECHC1234P",
    "gst": "29TECHC1234P1Z5",
    "vendor_category": "EQUIPMENT",
    "rating": 4,
    "status": "ACTIVE"
  }
]

// Usage in tests
func TestVendorService_WithFixtures(t *testing.T) {
    fixtures, err := testdata.LoadFixtures()
    require.NoError(t, err)

    for _, vendor := range fixtures.Vendors {
        // Use fixture data in tests
        assert.NotEmpty(t, vendor.CompanyName)
    }
}
```

### Factory Functions

```go
// test_factories.go
package testdata

func CreateDefaultVendor() *db.Vendor {
    return &db.Vendor{
        ID:          uuid.New().String(),
        CompanyName: "Default Test Vendor",
        Status:      stringPtr("ACTIVE"),
    }
}

func CreateRawMaterialVendor() *db.Vendor {
    vendor := CreateDefaultVendor()
    vendor.VendorCategory = stringPtr("RAW_MATERIAL")
    vendor.Rating = int32Ptr(5)
    return vendor
}

func CreateBlacklistedVendor() *db.Vendor {
    vendor := CreateDefaultVendor()
    vendor.IsBlacklisted = boolPtr(true)
    vendor.Status = stringPtr("BLACKLISTED")
    return vendor
}

// Helper functions
func stringPtr(s string) *string     { return &s }
func int32Ptr(i int32) *int32        { return &i }
func boolPtr(b bool) *bool           { return &b }
```

---

## Code Coverage Requirements

### Running Coverage

```bash
# Generate coverage for all packages
go test -coverprofile=coverage.out ./...

# View coverage in terminal
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Coverage for specific package
go test -cover ./vendors/services

# Coverage with detailed output
go test -coverprofile=coverage.out -covermode=atomic ./...
```

### Coverage Standards

- **Minimum Overall Coverage:** 80%
- **Service Layer:** 90%+ (business logic is critical)
- **Repository Layer:** 85%+ (data access must be reliable)
- **Handler Layer:** 80%+ (API contracts must be validated)
- **Utility Functions:** 70%+ (less critical paths)

### Coverage Report Example

```bash
# Output from: go tool cover -func=coverage.out
p9e.in/ugcl/vendors/services/vendor_service.go:25:  NewVendorService     100.0%
p9e.in/ugcl/vendors/services/vendor_service.go:31:  Create               95.5%
p9e.in/ugcl/vendors/services/vendor_service.go:65:  Update               92.3%
p9e.in/ugcl/vendors/services/vendor_service.go:85:  GetByID              100.0%
p9e.in/ugcl/vendors/services/vendor_service.go:95:  Delete               100.0%
p9e.in/ugcl/vendors/services/vendor_service.go:105: List                 88.9%
total:                                               (statements)         91.2%
```

### Excluding Code from Coverage

```go
// Use build tags to exclude test utilities
//go:build !test

// Use coverage ignore comments (with tools like gocovmerge)
// coverage:ignore
func DebugHelper() {
    // This function won't be counted in coverage
}
```

---

## CI/CD Testing Pipeline

### GitHub Actions Workflow

```yaml
# .github/workflows/test.yml
name: Test Pipeline

on:
  push:
    branches: [ main, develop, v2 ]
  pull_request:
    branches: [ main, develop, v2 ]

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_USER: postgres
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: testdb
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'

      - name: Cache Go modules
        uses: actions/cache@v3
        with:
          path: |
            ~/go/pkg/mod
            ~/.cache/go-build
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: |
            ${{ runner.os }}-go-

      - name: Install dependencies
        run: |
          go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
          curl -sSL "https://github.com/bufbuild/buf/releases/download/v1.28.1/buf-Linux-x86_64" -o /usr/local/bin/buf
          chmod +x /usr/local/bin/buf

      - name: Generate code
        run: |
          buf generate
          make sqlc-generate-all

      - name: Run linters
        run: |
          go vet ./...
          buf lint

      - name: Run unit tests
        run: go test -short -race -coverprofile=coverage.txt -covermode=atomic ./...

      - name: Run integration tests
        env:
          DB_DSN: host=localhost port=5432 user=postgres password=postgres dbname=testdb sslmode=disable
        run: go test -race -coverprofile=coverage-integration.txt -covermode=atomic ./...

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.txt,./coverage-integration.txt
          flags: unittests,integration
          name: codecov-umbrella

      - name: Check coverage threshold
        run: |
          COVERAGE=$(go tool cover -func=coverage.txt | grep total | awk '{print $3}' | sed 's/%//')
          echo "Coverage: $COVERAGE%"
          if (( $(echo "$COVERAGE < 80" | bc -l) )); then
            echo "Coverage $COVERAGE% is below threshold 80%"
            exit 1
          fi

  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.25'

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest
          args: --timeout=5m
```

### Pre-commit Hooks

```bash
# .githooks/pre-commit
#!/bin/bash

echo "Running pre-commit checks..."

# Run tests
echo "Running unit tests..."
go test -short ./...
if [ $? -ne 0 ]; then
    echo "❌ Unit tests failed"
    exit 1
fi

# Run linters
echo "Running go vet..."
go vet ./...
if [ $? -ne 0 ]; then
    echo "❌ go vet failed"
    exit 1
fi

# Check formatting
echo "Checking code formatting..."
UNFORMATTED=$(gofmt -l .)
if [ -n "$UNFORMATTED" ]; then
    echo "❌ The following files need formatting:"
    echo "$UNFORMATTED"
    exit 1
fi

# Check coverage
echo "Checking code coverage..."
go test -cover -coverprofile=coverage.out ./... > /dev/null
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
if (( $(echo "$COVERAGE < 80" | bc -l) )); then
    echo "❌ Coverage ${COVERAGE}% is below threshold 80%"
    exit 1
fi

echo "✅ All pre-commit checks passed!"
exit 0

# Install hook:
# git config core.hooksPath .githooks
```

---

## Performance and Load Testing

### Benchmarking

```go
func BenchmarkVendorService_Create(b *testing.B) {
    mockRepo := &MockVendorRepository{}
    svc := services.NewVendorService(mockRepo, nil)
    vendor := &db.Vendor{CompanyName: "Benchmark Corp"}
    ctx := context.Background()

    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            svc.Create(ctx, vendor, nil)
        }
    })
}

func BenchmarkVendorRepository_GetByID(b *testing.B) {
    // Setup real database
    container, db := setupPostgresContainer(b)
    defer container.Terminate(context.Background())
    defer db.Close()

    queries := db.NewQueries(db)
    repo := repository.NewVendorRepository(queries)

    // Seed data
    created, _ := repo.Create(context.Background(), db.CreateVendorParams{
        CompanyName: "Benchmark Test",
    })

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        repo.GetByID(context.Background(), created.ID)
    }
}
```

### Load Testing with k6

```javascript
// loadtest/vendor_create.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '30s', target: 10 },  // Ramp up to 10 users
    { duration: '1m', target: 50 },   // Ramp up to 50 users
    { duration: '30s', target: 0 },   // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],  // 95% of requests under 500ms
    http_req_failed: ['rate<0.01'],    // Error rate under 1%
  },
};

export default function () {
  const payload = JSON.stringify({
    vendor: {
      company_name: `Load Test Corp ${__VU}-${__ITER}`,
      pan: `LOAD${__VU}${__ITER}P`,
    },
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${__ENV.JWT_TOKEN}`,
    },
  };

  const res = http.post(
    'http://localhost:10011/vendors.api.v2.vendor.VendorService/CreateVendor',
    payload,
    params
  );

  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });

  sleep(1);
}

// Run: k6 run loadtest/vendor_create.js
```

---

## E2E Testing Strategy

### End-to-End Test Example

```go
// e2e/vendor_workflow_test.go
//go:build e2e

package e2e_test

import (
    "context"
    "testing"
    "net/http"

    "connectrpc.com/connect"
    vendorpb "p9e.in/ugcl/vendors/api/v2/vendor"
    "p9e.in/ugcl/vendors/api/v2/vendor/vendorconnect"
    userpb "p9e.in/ugcl/identity/user/api/v2/user"
)

func TestVendorWorkflow_E2E(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping E2E test in short mode")
    }

    // Setup
    baseURL := "http://localhost:10011"
    client := vendorconnect.NewVendorServiceClient(http.DefaultClient, baseURL)
    ctx := context.Background()

    t.Run("Complete Vendor Lifecycle", func(t *testing.T) {
        // 1. Create vendor with user
        createReq := &vendorpb.CreateVendorRequest{
            Vendor: &vendorpb.Vendor{
                CompanyName:    "E2E Test Corporation",
                Pan:            "E2ETE1234P",
                Gst:            "29E2ETE1234P1Z5",
                VendorCategory: "RAW_MATERIAL",
            },
            User: &userpb.User{
                Email:     "e2e@testcorp.com",
                FirstName: "John",
                LastName:  "Doe",
            },
        }

        createResp, err := client.CreateVendor(ctx, connect.NewRequest(createReq))
        require.NoError(t, err)
        vendorID := createResp.Msg.Id

        // 2. Retrieve created vendor
        getResp, err := client.GetVendor(ctx, connect.NewRequest(&vendorpb.VendorIdentifier{
            Id: vendorID,
        }))
        require.NoError(t, err)
        assert.Equal(t, "E2E Test Corporation", getResp.Msg.CompanyName)

        // 3. Update vendor rating
        updateReq := &vendorpb.UpdateVendorRequest{
            Vendor: &vendorpb.Vendor{
                Id:     vendorID,
                Rating: 5,
            },
        }
        updateResp, err := client.UpdateVendor(ctx, connect.NewRequest(updateReq))
        require.NoError(t, err)
        assert.Equal(t, int32(5), updateResp.Msg.Rating)

        // 4. List vendors and verify our vendor is in the list
        listResp, err := client.ListVendors(ctx, connect.NewRequest(&vendorpb.ListVendorsRequest{
            PageSize:   10,
            PageOffset: 0,
        }))
        require.NoError(t, err)
        assert.Greater(t, len(listResp.Msg.Vendors), 0)

        // 5. Delete vendor
        _, err = client.DeleteVendor(ctx, connect.NewRequest(&vendorpb.VendorIdentifier{
            Id: vendorID,
        }))
        require.NoError(t, err)

        // 6. Verify vendor is deleted
        _, err = client.GetVendor(ctx, connect.NewRequest(&vendorpb.VendorIdentifier{
            Id: vendorID,
        }))
        require.Error(t, err)
    })
}

// Run: go test -tags=e2e ./e2e/...
```

---

## Best Practices Summary

### Testing Checklist

- [ ] Write tests before or alongside implementation (TDD)
- [ ] Test happy path and error cases
- [ ] Use table-driven tests for multiple scenarios
- [ ] Mock external dependencies
- [ ] Use testcontainers for database integration tests
- [ ] Maintain > 80% code coverage
- [ ] Run tests in CI/CD pipeline
- [ ] Use meaningful test names
- [ ] Keep tests independent and isolated
- [ ] Clean up resources (defer cleanup)
- [ ] Use test helpers to reduce duplication
- [ ] Document complex test scenarios
- [ ] Run tests with race detector (`-race`)
- [ ] Profile and benchmark critical paths

---

## Troubleshooting

### Common Testing Issues

**Issue: Tests are slow**
- Use `-short` flag to skip integration tests during development
- Run integration tests in parallel
- Use test caching (`go clean -testcache` to clear)

**Issue: Flaky tests**
- Check for race conditions (`go test -race`)
- Ensure proper cleanup of resources
- Avoid using time.Sleep, use channels or polling
- Check for shared mutable state

**Issue: Database tests fail**
- Ensure PostgreSQL container starts properly
- Check port conflicts
- Verify migrations ran successfully
- Clean database between tests

---

## Additional Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Testcontainers Go](https://golang.testcontainers.org/)
- [Testify Documentation](https://github.com/stretchr/testify)
- [GoMock Documentation](https://github.com/golang/mock)
- [k6 Load Testing](https://k6.io/docs/)

---

**Last Updated:** 2025-10-06
**Version:** 2.0
