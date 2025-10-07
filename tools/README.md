# Tools Module

## 1. Module Overview

The Tools module is a special Go module that declares development tool dependencies required for the UGCL platform. It ensures that development tools like database migration utilities and code generators are tracked in go.mod and can be installed consistently across development environments.

**Purpose:** Manage development tool dependencies using the Go modules system, ensuring all developers have access to the same versions of development tools.

**Key Features:**
- Declares tool dependencies in go.mod
- Enables consistent tool versions across team
- Supports `go install` for tool installation
- Uses `//go:build tools` build constraint to exclude from production builds

## 2. Declared Tools

### Atlas

**Package:** `ariga.io/atlas/cmd/atlas`

**Purpose:** Database schema migration tool

**Usage:**
```bash
# Install atlas
go install ariga.io/atlas/cmd/atlas@latest

# Generate migrations
atlas migrate diff --env local

# Apply migrations
atlas migrate apply --env local

# View schema
atlas schema inspect --env local
```

**Documentation:** [Atlas Docs](https://atlasgo.io/)

### GORM Schema Provider for Atlas

**Package:** `ariga.io/atlas-provider-gorm/gormschema`

**Purpose:** Allows Atlas to read GORM models and generate migrations

**Usage:**
```bash
# Used automatically by Atlas when configured
# See atlas.hcl for configuration
```

**Documentation:** [GORM Provider](https://atlasgo.io/guides/orms/gorm)

## 3. Installing Tools

### Install All Tools

```bash
# From the project root
go install $(go list -f '{{join .Imports " "}}' tools/tools.go)
```

### Install Specific Tool

```bash
# Install Atlas
go install ariga.io/atlas/cmd/atlas@latest

# Install GORM schema provider
go install ariga.io/atlas-provider-gorm/gormschema@latest
```

### Verify Installation

```bash
# Check atlas is installed
atlas version

# List installed tools
go version -m $(which atlas)
```

## 4. Adding New Tools

To add a new development tool:

**1. Add import to tools.go:**
```go
//go:build tools

package tools

import (
    _ "ariga.io/atlas-provider-gorm/gormschema"
    _ "ariga.io/atlas/cmd/atlas"
    _ "github.com/golangci/golangci-lint/cmd/golangci-lint"  // New tool
)
```

**2. Update dependencies:**
```bash
go mod tidy
```

**3. Install the tool:**
```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

## 5. Common Development Tools

### Recommended Tools to Add

**Code Quality:**
```go
_ "github.com/golangci/golangci-lint/cmd/golangci-lint"  // Linter
_ "golang.org/x/tools/cmd/goimports"                     // Import organizer
_ "mvdan.cc/gofumpt"                                     // Code formatter
```

**Code Generation:**
```go
_ "github.com/sqlc-dev/sqlc/cmd/sqlc"                    // SQL to Go
_ "google.golang.org/protobuf/cmd/protoc-gen-go"        // Proto to Go
_ "connectrpc.com/connect/cmd/protoc-gen-connect-go"    // Connect RPC
```

**Testing:**
```go
_ "github.com/onsi/ginkgo/v2/ginkgo"                     // BDD testing
_ "gotest.tools/gotestsum"                               // Test runner
```

**Documentation:**
```go
_ "github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc"  // Proto docs
_ "github.com/swaggo/swag/cmd/swag"                           // Swagger docs
```

## 6. Tool Configuration

### Atlas Configuration (atlas.hcl)

```hcl
env "local" {
  src = "gorm://ent/schema"
  dev = "docker://postgres/15/dev"
  migration {
    dir = "file://migrations"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}

env "prod" {
  url = getenv("DATABASE_URL")
  migration {
    dir = "file://migrations"
  }
}
```

### Makefile Integration

```makefile
# Install development tools
.PHONY: tools
tools:
	@echo "Installing development tools..."
	@go install $(shell go list -f '{{join .Imports " "}}' tools/tools.go)

# Run linter
.PHONY: lint
lint:
	@golangci-lint run ./...

# Format code
.PHONY: fmt
fmt:
	@gofumpt -w .
	@goimports -w .

# Generate code
.PHONY: generate
generate:
	@buf generate
	@sqlc generate

# Run migrations
.PHONY: migrate
migrate:
	@atlas migrate apply --env local
```

## 7. CI/CD Integration

### GitHub Actions

```yaml
name: CI

on: [push, pull_request]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.25'
      - name: Install tools
        run: make tools
      - name: Run linter
        run: make lint

  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.25'
      - name: Install tools
        run: make tools
      - name: Run tests
        run: go test -v ./...

  migrations:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Install Atlas
        run: go install ariga.io/atlas/cmd/atlas@latest
      - name: Validate migrations
        run: atlas migrate validate --env local
```

## 8. Best Practices

### Tool Management

1. **Pin tool versions:** Use specific versions in go.mod
2. **Document tools:** Add comments explaining tool purpose
3. **Automate installation:** Use Makefile or scripts
4. **Keep tools updated:** Regularly update tool versions
5. **Test tool changes:** Verify tools work after updates

### Build Constraint

The `//go:build tools` constraint ensures:
- Tools are not included in production binaries
- Dependencies are tracked in go.mod
- Tools can be installed with `go install`

### Version Control

```gitignore
# Don't commit installed binaries
bin/
.bin/

# Don't commit tool configs (if machine-specific)
.tool-versions
```

## 9. Troubleshooting

### Common Issues

**Issue: Tool not found after install**
- **Cause:** `$GOPATH/bin` not in PATH
- **Solution:** Add to PATH: `export PATH=$PATH:$(go env GOPATH)/bin`

**Issue: Version conflict**
- **Cause:** Multiple tool versions installed
- **Solution:** Use `go install` with specific version

**Issue: Permission denied**
- **Cause:** No write access to `$GOPATH/bin`
- **Solution:** Check permissions or install to custom location

**Issue: go.mod conflicts**
- **Cause:** Tool dependencies conflict with project deps
- **Solution:** Update dependencies with `go mod tidy`

### Debugging

```bash
# Check where tools are installed
which atlas
which sqlc

# Check tool versions
go version -m $(which atlas)

# List all tools in go.mod
go list -m all | grep -E 'atlas|sqlc|protoc'

# Clean and reinstall
go clean -modcache
make tools
```

## 10. Additional Resources

- [Go Tools as Dependencies](https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module)
- [Atlas Documentation](https://atlasgo.io/)
- [SQLC Documentation](https://docs.sqlc.dev/)
- [Buf Documentation](https://docs.buf.build/)
- [GolangCI-Lint](https://golangci-lint.run/)

## Summary

The tools module is a lightweight but essential part of the development workflow. It ensures that all developers use consistent versions of development tools, making builds reproducible and reducing "works on my machine" issues.

**Key Points:**
- Declares tool dependencies in a separate package
- Uses `//go:build tools` constraint to exclude from builds
- Enables `go install` for consistent tool versions
- Currently includes Atlas for database migrations
- Can be extended with additional development tools as needed
