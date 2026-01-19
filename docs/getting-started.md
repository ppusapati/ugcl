# Getting Started with UGCL Backend v2

Welcome to the UGCL platform! This guide will help you get up and running quickly.

## Prerequisites

Before you begin, ensure you have the following installed:

- **Go 1.25+** - [Download](https://golang.org/dl/)
- **Docker & Docker Compose** - [Download](https://www.docker.com/get-started)
- **Git** - [Download](https://git-scm.com/downloads)
- **Make** - Usually pre-installed on Linux/Mac, [Windows](http://gnuwin32.sourceforge.net/packages/make.htm)

## Quick Setup (5 Minutes)

### 1. Clone the Repository

```bash
git clone https://github.com/your-org/ugcl-backend-v2.git
cd ugcl-backend-v2
```

### 2. Start Dependencies

```bash
# Start PostgreSQL, Redis, and MinIO
docker-compose up -d postgres redis minio
```

### 3. Setup Environment

```bash
# Copy environment template
cp .env.example .env

# The defaults should work for local development
```

### 4. Run Migrations

```bash
# Install Atlas (migration tool)
go install ariga.io/atlas/cmd/atlas@latest

# Run database migrations
make migrate-up
```

### 5. Generate Code

```bash
# Install code generation tools
make tools

# Generate proto and SQLC code
make generate
```

### 6. Run the Application

```bash
# Start the server
go run cmd/main.go

# Or use hot reload (recommended for development)
air
```

### 7. Verify Installation

```bash
# Check health endpoint
curl http://localhost:8080/health

# Expected response:
# {"status":"healthy","modules":{"database":"healthy",...}}
```

🎉 **Success!** The UGCL backend is now running on `http://localhost:8080`

## Project Structure

```
backend/v2/
├── cmd/                    # Application entry point
├── packages/               # Shared utilities
├── core/                   # Core business logic
├── identity/               # Identity & auth modules
│   ├── auth/              # Authentication
│   ├── user/              # User management
│   ├── tenant/            # Multi-tenancy
│   └── entity/            # Entity abstraction
├── formbuilder/           # Dynamic forms
├── organization/          # Org hierarchy
├── dms/                   # Document management
├── vendors/               # Vendor management
├── projects/              # Project management
└── docs/                  # Documentation
```

## Key Concepts

### 1. Proto-First Development

Services are defined in `.proto` files first, then code is generated:

```protobuf
// proto/user.proto
service UserService {
  rpc CreateUser(CreateUserRequest) returns (User);
}
```

Generate code:
```bash
buf generate
```

Learn more: [Proto-First Development Guide](guides/proto-first-development.md)

### 2. Modular Architecture

Each feature is a self-contained module:

```
{module}/
├── proto/          # API definitions
├── db/             # Database layer
├── services/       # Business logic
├── handlers/       # API handlers
└── module.go       # FX module
```

### 3. Dependency Injection

Using Uber FX for dependency injection:

```go
fx.Module("user",
    fx.Provide(
        repository.NewUserRepository,
        services.NewUserService,
        handlers.NewUserHandler,
    ),
)
```

## Common Tasks

### Create a New Module

```bash
# Use the module template
make new-module MODULE=mymodule

# This creates:
# - mymodule/proto/
# - mymodule/db/
# - mymodule/services/
# - mymodule/handlers/
# - mymodule/module.go
```

### Add a Database Table

```sql
-- db/schema/schema.sql
CREATE TABLE my_table (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
```

```bash
# Create migration
atlas migrate diff add_my_table --env local

# Apply migration
make migrate-up
```

### Add SQLC Queries

```sql
-- db/queries/my_table.sql
-- name: GetMyRecord :one
SELECT * FROM my_table WHERE id = $1;

-- name: ListMyRecords :many
SELECT * FROM my_table ORDER BY created_at DESC;
```

```bash
# Generate Go code
sqlc generate -f mymodule/db/sqlc.yaml
```

### Run Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific package tests
go test ./identity/user/...

# Run integration tests
make test-integration
```

### View Logs

```bash
# Application logs
tail -f logs/app.log

# Docker logs
docker-compose logs -f app

# Filter by level
grep "level=error" logs/app.log
```

## Development Workflow

### 1. Create Feature Branch

```bash
git checkout -b feature/UGCL-123-add-feature
```

### 2. Make Changes

- Update `.proto` files for API changes
- Run `buf generate` to regenerate code
- Implement service logic
- Add tests

### 3. Run Checks

```bash
# Format code
make fmt

# Run linter
make lint

# Run tests
make test
```

### 4. Commit and Push

```bash
git add .
git commit -m "feat(module): add new feature"
git push origin feature/UGCL-123-add-feature
```

### 5. Create Pull Request

- Go to GitHub
- Create PR from your branch to `main`
- Wait for CI checks to pass
- Request review

## Useful Commands

```bash
# Development
make run                    # Run application
make test                   # Run tests
make lint                   # Run linter
make fmt                    # Format code

# Code Generation
make generate               # Generate all code
make proto-generate         # Generate proto code only
make sqlc-generate          # Generate SQLC code only

# Database
make migrate-up             # Run migrations
make migrate-down           # Rollback migrations
make migrate-status         # Check migration status

# Docker
make docker-build           # Build Docker image
make docker-up              # Start all services
make docker-down            # Stop all services

# Tools
make tools                  # Install development tools
```

## API Documentation

### Explore APIs

1. **Swagger UI**: http://localhost:8080/swagger/
2. **API Docs**: [docs/api/](api/)
3. **Module READMEs**: See individual module directories

### Example API Call

```bash
# Create a user
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "fullname": "John Doe"
  }'

# Get user
curl http://localhost:8080/api/v1/users/{user-id}
```

## Troubleshooting

### Database Connection Failed

```bash
# Check PostgreSQL is running
docker ps | grep postgres

# Restart PostgreSQL
docker-compose restart postgres

# Check connection
psql -h localhost -U postgres -d ugcl_dev
```

### Port Already in Use

```bash
# Find process using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>

# Or change port in .env
SERVER_PORT=8081
```

### Migration Failed

```bash
# Check migration status
atlas migrate status --env local

# Rollback last migration
atlas migrate down --env local

# Re-apply
atlas migrate apply --env local
```

### Code Generation Issues

```bash
# Clean generated code
rm -rf */api/

# Reinstall tools
make tools

# Regenerate
make generate
```

## Next Steps

Now that you're set up, explore these resources:

1. **[Development Guide](guides/development-guide.md)** - Coding standards and best practices
2. **[Proto-First Development](guides/proto-first-development.md)** - Learn the proto-first workflow
3. **[Testing Guide](guides/testing-guide.md)** - Testing strategies and examples
4. **[Architecture Overview](architecture/system-overview.md)** - System architecture deep dive
5. **[Module READMEs](../README.md#modules)** - Explore individual modules

## Getting Help

- **Documentation**: [docs/](.)
- **Issues**: [GitHub Issues](https://github.com/your-org/ugcl-backend-v2/issues)
- **Discussions**: [GitHub Discussions](https://github.com/your-org/ugcl-backend-v2/discussions)
- **Team Chat**: Slack #ugcl-backend

## Contributing

We welcome contributions! Please read our [Development Guide](guides/development-guide.md) and follow our coding standards.

Happy coding! 🚀
