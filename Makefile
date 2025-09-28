# Proto generation using buf
.PHONY: proto-generate
proto-generate:
	@echo "Generating protobuf files..."
	buf generate

.PHONY: proto-lint
proto-lint:
	@echo "Linting protobuf files..."
	buf lint

.PHONY: proto-breaking
proto-breaking:
	@echo "Checking for breaking changes in protobuf..."
	buf breaking --against '.git#branch=main'

# SQLC generation for all modules
.PHONY: sqlc-generate-all
sqlc-generate-all:
	@echo "Generating SQLC for all modules..."
	cd vendors/db && sqlc generate
	cd identity/user/db/sqlc && sqlc generate
	cd identity/tenant/db && sqlc generate
	cd identity/auth/db && sqlc generate
	cd contractor/db && sqlc generate
	cd finance/db && sqlc generate
	cd formbuilder/db && sqlc generate
	cd notification/db && sqlc generate
	cd core/database && sqlc generate
	cd projects/db && sqlc generate
	cd databridge/db && sqlc generate
	cd masters/db && sqlc generate
	cd searchservice/db && sqlc generate
	cd documentviewer/db && sqlc generate
	cd approvalworkflow/db && sqlc generate
	cd dataarchive/db && sqlc generate
	cd backupdr/db && sqlc generate

# Generate specific modules
.PHONY: sqlc-vendors
sqlc-vendors:
	cd vendors/db && sqlc generate

.PHONY: sqlc-user
sqlc-user:
	cd identity/user/db/sqlc && sqlc generate

.PHONY: sqlc-tenant
sqlc-tenant:
	cd identity/tenant/db && sqlc generate

.PHONY: sqlc-auth
sqlc-auth:
	cd identity/auth/db && sqlc generate

.PHONY: sqlc-projects
sqlc-projects:
	cd projects/db && sqlc generate

.PHONY: sqlc-finance
sqlc-finance:
	cd finance/db && sqlc generate

.PHONY: sqlc-formbuilder
sqlc-formbuilder:
	cd formbuilder/db && sqlc generate

.PHONY: sqlc-notification
sqlc-notification:
	cd notification/db && sqlc generate

.PHONY: sqlc-core
sqlc-core:
	cd core/database && sqlc generate

.PHONY: sqlc-databridge
sqlc-databridge:
	cd databridge/db && sqlc generate

.PHONY: sqlc-masters
sqlc-masters:
	cd masters/db && sqlc generate

.PHONY: sqlc-searchservice
sqlc-searchservice:
	cd searchservice/db && sqlc generate

.PHONY: sqlc-documentviewer
sqlc-documentviewer:
	cd documentviewer/db && sqlc generate

.PHONY: sqlc-approvalworkflow
sqlc-approvalworkflow:
	cd approvalworkflow/db && sqlc generate

.PHONY: sqlc-dataarchive
sqlc-dataarchive:
	cd dataarchive/db && sqlc generate

.PHONY: sqlc-backupdr
sqlc-backupdr:
	cd backupdr/db && sqlc generate

# Verify all configurations
.PHONY: sqlc-verify-all
sqlc-verify-all:
	cd contractor/db && sqlc verify
	cd finance/db && sqlc verify
	cd formbuilder/db && sqlc verify
	cd notification/db && sqlc verify
	cd core/database && sqlc verify
	cd identity/user/db/sqlc && sqlc verify
	cd identity/tenant/db && sqlc verify
	cd identity/auth/db && sqlc verify
	cd projects/db && sqlc verify
	cd databridge/db && sqlc verify

# Combined generation
.PHONY: generate-all
generate-all: proto-generate sqlc-generate-all
	@echo "Generated all proto and SQLC files"

# Development workflow
.PHONY: dev-setup
dev-setup: proto-lint sqlc-verify-all generate-all
	@echo "Development setup complete"

.PHONY: migrate-new migrate-up migrate-down migrate-status
# Makefile for UGCL Migrations

.PHONY: migrate-help migrate-apply migrate-generate migrate-validate migrate-aggregate sqlc-generate dev-setup

# Variables
CONFIG_FILE ?= ./configs.yaml
MIGRATION_CLI = ./cmd/migrate
DB_DSN ?= $(shell grep -A10 "postgres:" $(CONFIG_FILE) | grep -E "(user|password|host|port|dbname)" | tr '\n' ' ')

# Help target
migrate-help:
	@echo "Available migration targets:"
	@echo "  migrate-apply     - Apply all pending migrations"
	@echo "  migrate-generate  - Generate new migration (requires NAME=migration_name)"
	@echo "  migrate-validate  - Validate migration directory integrity"
	@echo "  migrate-aggregate - Aggregate all module schemas"
	@echo "  sqlc-generate     - Generate SQLC code for all modules"
	@echo "  dev-setup         - Complete development setup"
	@echo ""
	@echo "Usage examples:"
	@echo "  make migrate-generate NAME=add_user_table"
	@echo "  make migrate-apply"
	@echo "  make dev-setup"

# Apply migrations
migrate-apply:
	@echo "Applying migrations..."
	go run $(MIGRATION_CLI) -action=apply -config=$(CONFIG_FILE)

# Generate new migration
migrate-generate:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make migrate-generate NAME=migration_name"; \
		exit 1; \
	fi
	@echo "Generating migration: $(NAME)"
	go run $(MIGRATION_CLI) -action=generate -name=$(NAME) -config=$(CONFIG_FILE)

# Validate migrations
migrate-validate:
	@echo "Validating migrations..."
	go run $(MIGRATION_CLI) -action=validate -config=$(CONFIG_FILE)

# Aggregate schemas from all modules
migrate-aggregate:
	@echo "Aggregating schemas from modules..."
	go run $(MIGRATION_CLI) -action=aggregate -config=$(CONFIG_FILE)

# Generate SQLC code for all modules
sqlc-generate:
	@echo "Generating SQLC code for all modules..."
	@find . -name "sqlc.yaml" -exec dirname {} \; | while read module; do \
		echo "Generating SQLC for $$module"; \
		cd $$module && sqlc generate && cd - > /dev/null; \
	done

# Complete development setup
dev-setup: migrate-aggregate migrate-validate sqlc-generate
	@echo "Development setup completed!"
	@echo "Next steps:"
	@echo "  1. Review aggregated schema: migrations/aggregated_schema.sql"
	@echo "  2. Generate migration if needed: make migrate-generate NAME=initial_schema"
	@echo "  3. Apply migrations: make migrate-apply"

# Build migration CLI tool
build-migrate-cli:
	@echo "Building migration CLI..."
	go build -o bin/migrate $(MIGRATION_CLI)

# Clean generated files
clean:
	@echo "Cleaning generated files..."
	rm -f migrations/aggregated_schema.sql
	rm -f migrations/atlas_temp.hcl
	rm -f bin/migrate

# Check migration status
migrate-status:
	@echo "Checking migration status..."
	atlas migrate status --config atlas.hcl --env local

# Lint migrations
migrate-lint:
	@echo "Linting migrations..."
	atlas migrate lint --config atlas.hcl --env local

# Full workflow: aggregate -> validate -> generate -> apply
migrate-full: migrate-aggregate migrate-validate
	@echo "Full migration workflow completed!"
	@echo "If this is a new setup, run: make migrate-generate NAME=initial_schema"