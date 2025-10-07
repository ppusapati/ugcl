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
	cd identity/entity/db && sqlc generate
	cd contractor/db && sqlc generate
	cd finance/db && sqlc generate
	cd formbuilder/db && sqlc generate
	cd notification/db && sqlc generate
	cd organization/db && sqlc generate
	cd core/database && sqlc generate
	cd projects/db && sqlc generate
	cd databridge/db && sqlc generate
	cd masters/db && sqlc generate
	cd metasearch/db && sqlc generate
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

.PHONY: sqlc-contractor
sqlc-contractor:
	cd contractor/db && sqlc generate

.PHONY: sqlc-entity
sqlc-entity:
	cd identity/entity/db && sqlc generate 

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

.PHONY: sqlc-organization
sqlc-organization:
	cd organization/db && sqlc generate

.PHONY: sqlc-core
sqlc-core:
	cd core/database && sqlc generate

.PHONY: sqlc-databridge
sqlc-databridge:
	cd databridge/db && sqlc generate

.PHONY: sqlc-masters
sqlc-masters:
	cd masters/db && sqlc generate

.PHONY: sqlc-metasearch
sqlc-metasearch:
	cd metasearch/db && sqlc generate

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
	cd identity/entity/db && sqlc verify
	cd vendors/db && sqlc verify
	cd projects/db && sqlc verify
	cd databridge/db && sqlc verify
	cd masters/db && sqlc verify

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
# API Documentation Generation
.PHONY: docs docs-api docs-clean docs-serve

# Generate API documentation from all proto files
docs-api:
	@echo "Generating API documentation from proto files..."
	@mkdir -p docs/api
	@echo "Generating organization documentation..."
	@protoc --proto_path=. --doc_out=docs/api --doc_opt=html,organization.html organization/proto/*.proto packages/proto/*.proto
	@echo "Generating DMS documentation..."
	@protoc --proto_path=. --doc_out=docs/api --doc_opt=html,dms.html dms/proto/*.proto packages/proto/*.proto
	@echo "Generating identity/entity documentation..."
	@protoc --proto_path=. --doc_out=docs/api --doc_opt=html,entity.html identity/entity/proto/*.proto packages/proto/*.proto
	@echo "Generating identity/user documentation..."
	@protoc --proto_path=. --doc_out=docs/api --doc_opt=html,user.html identity/user/proto/*.proto packages/proto/*.proto
	@echo "Generating identity/auth documentation..."
	@protoc --proto_path=. --doc_out=docs/api --doc_opt=html,auth.html identity/auth/proto/*.proto packages/proto/*.proto
	@echo "Generating identity/tenant documentation..."
	@protoc --proto_path=. --doc_out=docs/api --doc_opt=html,tenant.html identity/tenant/proto/*.proto packages/proto/*.proto
	@echo "Generating employee documentation..."
	@protoc --proto_path=. --doc_out=docs/api --doc_opt=html,employee.html employee/proto/*.proto packages/proto/*.proto
	@echo "Generating contractors documentation..."
	@protoc --proto_path=. --doc_out=docs/api --doc_opt=html,contractors.html contractors/proto/*.proto packages/proto/*.proto
	@echo "Generating vendors documentation..."
	@protoc --proto_path=. --doc_out=docs/api --doc_opt=html,vendors.html vendors/proto/*.proto packages/proto/*.proto
	@echo "Generating formbuilder documentation..."
	@protoc --proto_path=. --doc_out=docs/api --doc_opt=html,formbuilder.html formbuilder/proto/*.proto packages/proto/*.proto
	@echo "Generating backupdr documentation..."
	@protoc --proto_path=. --doc_out=docs/api --doc_opt=html,backupdr.html backupdr/proto/*.proto packages/proto/*.proto
	@echo "Generating dataarchive documentation..."
	@protoc --proto_path=. --doc_out=docs/api --doc_opt=html,dataarchive.html dataarchive/proto/*.proto packages/proto/*.proto
	@echo "Generating scheduler documentation..."
	@protoc --proto_path=. --doc_out=docs/api --doc_opt=html,scheduler.html scheduler/proto/*.proto packages/proto/*.proto
	@echo "Generating insighthub documentation..."
	@protoc --proto_path=. --doc_out=docs/api --doc_opt=html,insighthub.html insighthub/proto/*.proto packages/proto/*.proto
	@echo "Generating insightviewer documentation..."
	@protoc --proto_path=. --doc_out=docs/api --doc_opt=html,insightviewer.html insightviewer/proto/*.proto packages/proto/*.proto
	@echo "Generating metasearch documentation..."
	@protoc --proto_path=. --doc_out=docs/api --doc_opt=html,metasearch.html metasearch/proto/*.proto packages/proto/*.proto
	@echo "Generating projects documentation..."
	@protoc --proto_path=. --doc_out=docs/api --doc_opt=html,projects.html projects/protos/*.proto packages/proto/*.proto
	@echo "Generating masters documentation..."
	@protoc --proto_path=. --doc_out=docs/api --doc_opt=html,masters.html masters/proto/*.proto packages/proto/*.proto
	@echo "Generating masters/pipeline documentation..."
	@protoc --proto_path=. --doc_out=docs/api --doc_opt=html,pipeline.html masters/pipeline/proto/*.proto packages/proto/*.proto
	@echo "Generating notification documentation..."
	@protoc --proto_path=. --doc_out=docs/api --doc_opt=html,notification.html notification/proto/*.proto packages/proto/*.proto
	@echo "Generating databridge documentation..."
	@protoc --proto_path=. --doc_out=docs/api --doc_opt=html,databridge.html databridge/proto/*.proto packages/proto/*.proto
	@echo "Generating core documentation..."
	@protoc --proto_path=. --doc_out=docs/api --doc_opt=html,core.html core/proto/*.proto packages/proto/*.proto
	@echo "Generating packages documentation..."
	@protoc --proto_path=. --doc_out=docs/api --doc_opt=html,packages.html packages/proto/*.proto
	@echo "Generating API index..."
	@protoc --proto_path=. --doc_out=docs/api --doc_opt=html,index.html organization/proto/*.proto dms/proto/*.proto identity/entity/proto/*.proto identity/user/proto/*.proto identity/auth/proto/*.proto identity/tenant/proto/*.proto employee/proto/*.proto contractors/proto/*.proto vendors/proto/*.proto formbuilder/proto/*.proto packages/proto/*.proto
	@echo "API documentation generated successfully in docs/api/"

# Generate all documentation
docs: docs-api
	@echo "All documentation generated successfully!"

# Clean documentation
docs-clean:
	@echo "Cleaning documentation..."
	@rm -rf docs/api/*.html
	@echo "Documentation cleaned!"

# Serve documentation locally (requires python)
docs-serve:
	@echo "Serving documentation at http://localhost:8000/docs/api/"
	@cd docs/api && python -m http.server 8000
