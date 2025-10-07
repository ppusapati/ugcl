# Core Module

## 1. Module Overview

The Core module provides essential shared functionality, domain models, middleware, and utilities that are used across the UGCL platform. It serves as the central foundation for common business logic, request/response handling, authentication middleware, and standardized API patterns.

**Purpose:** Provide core business logic, shared models, API utilities, middleware, and helper functions that are universally applicable across all microservices in the UGCL ecosystem.

**Key Features:**
- Core domain models and DTOs
- Authentication and authorization middleware
- Request/response standardization
- API versioning support
- Configuration management
- JWT token handling
- Common helper utilities
- Proto message definitions for core types
- Database connection and GORM integration
- Dependency injection with Uber FX

## 2. Architecture

### Module Structure
```
core/
├── api/                    # API layer utilities
│   ├── handlers/          # Base HTTP/gRPC handlers
│   ├── validators/        # Request validators
│   └── formatters/        # Response formatters
├── config/                # Configuration management
│   ├── config.go          # Config structures
│   ├── loader.go          # Config loading from files/env
│   └── database.go        # Database configuration
├── helper/                # Helper utilities
│   ├── jwt.go             # JWT token utilities
│   ├── hash.go            # Password hashing
│   ├── uuid.go            # UUID generation
│   ├── time.go            # Time formatting utilities
│   └── string.go          # String manipulation
├── middleware/            # Middleware components
│   ├── auth.go            # Authentication middleware
│   ├── tenant.go          # Multi-tenant context middleware
│   ├── logging.go         # Request logging
│   ├── recovery.go        # Panic recovery
│   ├── cors.go            # CORS handling
│   └── validation.go      # Request validation
├── models/                # Core domain models
│   ├── user.go            # User models
│   ├── tenant.go          # Tenant models
│   ├── response.go        # API response models
│   ├── pagination.go      # Pagination models
│   └── error.go           # Error models
├── proto/                 # Proto definitions
│   ├── common.proto       # Common message types
│   ├── error.proto        # Error types
│   └── pagination.proto   # Pagination types
├── services/              # Core services
│   ├── jwt_service.go     # JWT service
│   ├── hash_service.go    # Hashing service
│   └── config_service.go  # Configuration service
└── go.mod                 # Module dependencies
```

### Key Dependencies
- **Connect-RPC:** `connectrpc.com/connect` - RPC framework
- **JWT:** `github.com/golang-jwt/jwt/v5` - JSON Web Tokens
- **FX:** `go.uber.org/fx` - Dependency injection
- **GORM:** `gorm.io/gorm` - ORM for database
- **Protobuf:** `google.golang.org/protobuf` - Protocol buffers
- **Config:** `gopkg.in/yaml.v3` - YAML configuration
- **Env:** `github.com/joho/godotenv` - Environment variables

## 3. Quick Start

### Load Configuration

```go
import (
    "p9e.in/ugcl/core/config"
)

// Load from YAML file
cfg, err := config.LoadConfig("config.yaml")
if err != nil {
    log.Fatal(err)
}

// Access configuration
dbConfig := cfg.Database
serverPort := cfg.Server.Port
jwtSecret := cfg.JWT.Secret
```

### Use JWT Service

```go
import (
    "p9e.in/ugcl/core/services"
)

// Create JWT service
jwtService := services.NewJWTService(services.JWTConfig{
    Secret:           "your-secret-key",
    AccessTokenTTL:   15 * time.Minute,
    RefreshTokenTTL:  7 * 24 * time.Hour,
    Issuer:           "ugcl-platform",
})

// Generate access token
token, err := jwtService.GenerateAccessToken(userID, claims)

// Validate token
claims, err := jwtService.ValidateToken(tokenString)

// Generate refresh token
refreshToken, err := jwtService.GenerateRefreshToken(userID)
```

### Use Middleware

```go
import (
    "p9e.in/ugcl/core/middleware"
    "github.com/labstack/echo/v4"
)

e := echo.New()

// Add authentication middleware
e.Use(middleware.Auth(middleware.AuthConfig{
    JWTSecret: "your-secret",
    Skipper: func(c echo.Context) bool {
        // Skip auth for public routes
        return c.Path() == "/health" || c.Path() == "/login"
    },
}))

// Add tenant context middleware
e.Use(middleware.TenantContext())

// Add logging middleware
e.Use(middleware.RequestLogger())

// Add recovery middleware
e.Use(middleware.Recovery())

// Add CORS middleware
e.Use(middleware.CORS())
```

### Standardized API Responses

```go
import (
    "p9e.in/ugcl/core/models"
)

// Success response
func GetUser(c echo.Context) error {
    user, err := userService.GetByID(userID)
    if err != nil {
        return c.JSON(500, models.ErrorResponse{
            Code:    "INTERNAL_ERROR",
            Message: "Failed to fetch user",
            Details: err.Error(),
        })
    }

    return c.JSON(200, models.SuccessResponse{
        Success: true,
        Data:    user,
        Message: "User retrieved successfully",
    })
}

// Paginated response
func ListUsers(c echo.Context) error {
    users, total, err := userService.List(page, pageSize)
    if err != nil {
        return err
    }

    return c.JSON(200, models.PaginatedResponse{
        Success: true,
        Data:    users,
        Pagination: models.Pagination{
            Page:       page,
            PageSize:   pageSize,
            TotalCount: total,
            TotalPages: (total + pageSize - 1) / pageSize,
        },
    })
}
```

## 4. API Reference

### Configuration Service

**Load Configuration:**
```go
cfg, err := config.LoadConfig(filePath string) (*Config, error)
cfg, err := config.LoadConfigFromEnv() (*Config, error)
```

**Config Structure:**
```go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    JWT      JWTConfig
    Redis    RedisConfig
    Logging  LoggingConfig
}
```

### JWT Service

**Methods:**
- `GenerateAccessToken(userID string, claims map[string]interface{}) (string, error)`
- `GenerateRefreshToken(userID string) (string, error)`
- `ValidateToken(tokenString string) (*Claims, error)`
- `RefreshAccessToken(refreshToken string) (string, error)`
- `RevokeToken(tokenID string) error`

### Hash Service

**Methods:**
- `HashPassword(password string) (string, error)` - Bcrypt hash
- `ComparePassword(hashedPassword, password string) bool` - Verify password
- `GenerateSalt() string` - Generate random salt
- `HashWithSalt(value, salt string) string` - SHA-256 with salt

### Helper Functions

**UUID:**
```go
id := helper.NewUUID()                    // Generate new UUID
valid := helper.IsValidUUID(uuidString)   // Validate UUID
```

**Time:**
```go
formatted := helper.FormatTime(time.Now(), "2006-01-02")
parsed := helper.ParseTime("2025-10-06", "2006-01-02")
timestamp := helper.UnixTimestamp()
```

**String:**
```go
masked := helper.MaskEmail("user@example.com")       // u***@example.com
masked := helper.MaskPhone("+91-9876543210")         // +91-****3210
truncated := helper.Truncate("long text", 10)         // "long te..."
```

## 5. Middleware Reference

### Auth Middleware

```go
middleware.Auth(middleware.AuthConfig{
    JWTSecret:      "secret",
    TokenExtractor: middleware.FromHeader("Authorization"),
    Skipper:        func(c echo.Context) bool { return false },
    ErrorHandler:   customErrorHandler,
})
```

**Features:**
- Extracts JWT token from header/cookie/query
- Validates token signature and expiry
- Injects user context into request
- Configurable skip conditions
- Custom error handling

### Tenant Context Middleware

```go
middleware.TenantContext()
```

**Features:**
- Extracts tenant ID from header/token/subdomain
- Validates tenant access for user
- Injects tenant context into request
- Supports multi-tenancy isolation

### Request Logger Middleware

```go
middleware.RequestLogger()
```

**Features:**
- Logs request method, path, status, duration
- Includes request ID for tracing
- Logs user ID and tenant ID if available
- Configurable log level

### Recovery Middleware

```go
middleware.Recovery()
```

**Features:**
- Catches panics in handlers
- Logs stack trace
- Returns 500 error response
- Prevents server crash

### CORS Middleware

```go
middleware.CORS()
```

**Features:**
- Configurable allowed origins
- Allowed methods and headers
- Credentials support
- Preflight request handling

### Validation Middleware

```go
middleware.Validation()
```

**Features:**
- Validates request body against schema
- Validates path parameters
- Validates query parameters
- Returns detailed validation errors

## 6. Models Reference

### API Response Models

**SuccessResponse:**
```go
type SuccessResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data"`
    Message string      `json:"message,omitempty"`
}
```

**ErrorResponse:**
```go
type ErrorResponse struct {
    Success bool                   `json:"success"` // Always false
    Code    string                 `json:"code"`
    Message string                 `json:"message"`
    Details string                 `json:"details,omitempty"`
    Errors  []ValidationError      `json:"errors,omitempty"`
}
```

**PaginatedResponse:**
```go
type PaginatedResponse struct {
    Success    bool       `json:"success"`
    Data       interface{} `json:"data"`
    Pagination Pagination `json:"pagination"`
}

type Pagination struct {
    Page       int `json:"page"`
    PageSize   int `json:"page_size"`
    TotalCount int `json:"total_count"`
    TotalPages int `json:"total_pages"`
}
```

### Domain Models

**User Model:**
```go
type User struct {
    ID        string    `json:"id"`
    UUID      string    `json:"uuid"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    FullName  string    `json:"full_name"`
    IsActive  bool      `json:"is_active"`
    TenantIDs []string  `json:"tenant_ids"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

**Tenant Model:**
```go
type Tenant struct {
    ID        string    `json:"id"`
    UUID      string    `json:"uuid"`
    Name      string    `json:"name"`
    Domain    string    `json:"domain"`
    IsActive  bool      `json:"is_active"`
    Settings  map[string]interface{} `json:"settings"`
    CreatedAt time.Time `json:"created_at"`
}
```

## 7. Configuration

### Environment Variables

```env
# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_READ_TIMEOUT=30s
SERVER_WRITE_TIMEOUT=30s

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=ugcl
DB_USER=postgres
DB_PASSWORD=password
DB_SSL_MODE=disable
DB_MAX_CONNECTIONS=25

# JWT Configuration
JWT_SECRET=your-secret-key-here
JWT_ISSUER=ugcl-platform
JWT_ACCESS_TOKEN_EXPIRY=15m
JWT_REFRESH_TOKEN_EXPIRY=168h

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

### YAML Configuration

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: "30s"
  write_timeout: "30s"

database:
  host: "localhost"
  port: 5432
  name: "ugcl"
  user: "postgres"
  password: "password"
  ssl_mode: "disable"
  max_connections: 25

jwt:
  secret: "your-secret-key"
  issuer: "ugcl-platform"
  access_token_expiry: "15m"
  refresh_token_expiry: "168h"

logging:
  level: "info"
  format: "json"
```

## 8. Integration

### With Application Modules

All application modules depend on core for:
- **Middleware:** Auth, tenant context, logging
- **Models:** Standardized response formats
- **Helpers:** JWT, hashing, UUID generation
- **Config:** Centralized configuration management

### Dependency Injection with FX

```go
import (
    "p9e.in/ugcl/core/config"
    "p9e.in/ugcl/core/services"
    "go.uber.org/fx"
)

fx.New(
    // Provide core services
    fx.Provide(
        config.LoadConfig,
        services.NewJWTService,
        services.NewHashService,
    ),
    // Provide application services
    fx.Provide(
        NewUserService,
        NewTenantService,
    ),
    // Invoke startup
    fx.Invoke(StartServer),
).Run()
```

## 9. Development

### Adding New Middleware

```go
// core/middleware/custom.go
package middleware

import "github.com/labstack/echo/v4"

func CustomMiddleware(config CustomConfig) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // Pre-processing
            // ...

            // Call next handler
            err := next(c)

            // Post-processing
            // ...

            return err
        }
    }
}
```

### Adding New Helper

```go
// core/helper/custom.go
package helper

func CustomHelper(input string) string {
    // Implementation
    return output
}
```

### Testing

```bash
# Run tests
go test ./core/...

# With coverage
go test -cover ./core/...

# Verbose
go test -v ./core/...
```

## 10. Best Practices

### Configuration
- Use environment variables for sensitive data
- Provide defaults for optional settings
- Validate configuration on startup
- Use typed configuration structures

### Middleware
- Keep middleware focused and single-purpose
- Use middleware.Skipper for conditional execution
- Log errors appropriately
- Handle errors gracefully

### JWT
- Use strong secret keys (256+ bits)
- Set appropriate token expiry times
- Implement token refresh mechanism
- Revoke tokens on logout

### Response Format
- Always use standardized response models
- Include meaningful error codes
- Provide helpful error messages
- Use consistent field naming (snake_case)

## 11. Troubleshooting

### Common Issues

**Issue: JWT validation fails**
- **Cause:** Wrong secret or expired token
- **Solution:** Check JWT_SECRET matches, verify expiry time

**Issue: CORS errors in browser**
- **Cause:** Origin not allowed
- **Solution:** Configure CORS middleware with allowed origins

**Issue: Database connection failed**
- **Cause:** Wrong credentials or host
- **Solution:** Verify DB_* environment variables

**Issue: Config file not found**
- **Cause:** Incorrect file path
- **Solution:** Use absolute path or check working directory

## Additional Resources

- [Echo Framework](https://echo.labstack.com/)
- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)
- [Uber FX](https://uber-go.github.io/fx/)
- [GORM Documentation](https://gorm.io/docs/)
- [Connect-RPC](https://connectrpc.com/)
