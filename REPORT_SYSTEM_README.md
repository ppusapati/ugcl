# Report Builder & Report Viewer System

## Overview

This is a comprehensive Report Builder and Report Viewer system built following clean microservices architecture. The system consists of three main services and two frontend applications.

## Architecture

### Backend Services

#### 1. **Masters Service** - Metadata Management
- **Location**: `D:\Maheshwari\UGCL\backend\v2\masters\`
- **Purpose**: Manages database metadata, schemas, tables, columns, relationships, and business terms
- **Database**: Enhanced metadata schema with relationships and business glossary
- **API**: Complete CRUD operations via Connect RPC

#### 2. **InsightHub Service** - Report Builder
- **Location**: `D:\Maheshwari\UGCL\backend\v2\insighthub\`
- **Purpose**: Report definition management, field configuration, filters, charts
- **Database**: Report definitions, fields, filters, groups, sorts, charts, permissions
- **API**: Comprehensive report building operations

#### 3. **InsightViewer Service** - Report Viewer
- **Location**: `D:\Maheshwari\UGCL\backend\v2\insightviewer\`
- **Purpose**: Report execution, caching, exports, scheduling, alerts, subscriptions
- **Database**: Report runs, results, exports, schedules, alerts, subscriptions
- **API**: Report execution and viewing operations

### Frontend Applications

#### 1. **InsightHub App** - Report Builder UI
- **Location**: `D:\Maheshwari\UGCL\web\ugcl\apps\insighthub\`
- **Tech Stack**: QwikJS + UnoCSS + Drag & Drop
- **Features**: Visual report designer with drag-and-drop interface

#### 2. **InsightViewer App** - Report Viewer UI
- **Location**: `D:\Maheshwari\UGCL\web\ugcl\apps\insightviewer\`
- **Tech Stack**: QwikJS + UnoCSS + ECharts
- **Features**: Report execution, viewing, exports, scheduling

## Getting Started

### Prerequisites
- Go 1.21+
- Node.js 18+
- PostgreSQL 15+
- Docker (optional)

### Backend Setup

1. **Generate SQLC models**:
   ```bash
   # From backend/v2/
   make sqlc-masters
   make sqlc-generate-all
   ```

2. **Generate protobuf files**:
   ```bash
   buf generate
   ```

3. **Run database migrations**:
   ```bash
   make migrate-apply
   ```

4. **Start services**:
   ```bash
   # Each service can be started independently
   cd masters && go run ./cmd/server
   cd insighthub && go run ./cmd/server
   cd insightviewer && go run ./cmd/server
   ```

### Frontend Setup

1. **Install dependencies**:
   ```bash
   # From web/ugcl/
   npm install
   ```

2. **Start development servers**:
   ```bash
   # InsightHub (Report Builder)
   cd apps/insighthub && npm run dev

   # InsightViewer (Report Viewer)
   cd apps/insightviewer && npm run dev
   ```

## Key Features

### Masters Service
- ✅ Database metadata management
- ✅ Table and column definitions
- ✅ Relationship mapping
- ✅ Business glossary terms
- ✅ Data lineage tracking
- ✅ Data quality rules

### InsightHub (Report Builder)
- ✅ Visual drag-and-drop report designer
- ✅ Field selection and configuration
- ✅ Filter builder with logical operators
- ✅ Grouping and aggregation
- ✅ Sorting configuration
- ✅ Chart setup (ECharts integration)
- ✅ Query preview and validation
- ✅ Report permissions management

### InsightViewer (Report Viewer)
- ✅ Report execution engine
- ✅ Result caching system
- ✅ Export functionality (CSV, Excel, PDF)
- ✅ Report scheduling with cron
- ✅ Alert system with notifications
- ✅ User subscriptions
- ✅ Real-time streaming results

## Development Workflow

### Adding New Features

1. **Backend Changes**:
   ```bash
   # Update proto definitions
   vim service/proto/service.proto

   # Generate code
   buf generate

   # Update database schema if needed
   vim service/db/schema/schema.sql

   # Generate SQLC
   make sqlc-[service]

   # Implement service logic
   vim service/services/service_impl.go
   ```

2. **Frontend Changes**:
   ```bash
   # Add new components
   vim apps/app/src/components/new-component.tsx

   # Update routes
   vim apps/app/src/routes/new-route/index.tsx

   # Style with UnoCSS utilities
   # Use design tokens from @p9e.in/design-tokens
   ```

### Testing

```bash
# Backend tests
cd backend/v2
go test ./...

# Frontend tests
cd web/ugcl
npm test

# Integration tests
cd backend/v2
make test-integration
```

## Database Schema

### Masters Database
- `schemas_metadata` - Database schemas
- `tables_metadata` - Table definitions
- `columns_metadata` - Column definitions
- `table_relationships` - Foreign key relationships
- `business_terms` - Data glossary
- `column_business_terms` - Term associations

### InsightHub Database
- `reports` - Report definitions
- `report_fields` - Field configurations
- `report_filters` - Filter conditions
- `report_groups` - Grouping settings
- `report_sorts` - Sort configurations
- `report_charts` - Chart configurations
- `report_permissions` - Access control

### InsightViewer Database
- `report_runs` - Execution history
- `report_results` - Query results
- `report_cache` - Cached results
- `report_exports` - Export records
- `report_schedules` - Scheduled runs
- `report_alerts` - Alert definitions
- `report_subscriptions` - User subscriptions

## API Documentation

All services use Connect RPC (gRPC-compatible) with REST fallback:

- **Masters Service**: `http://localhost:8081/connect/masters.MastersService/`
- **InsightHub Service**: `http://localhost:8082/connect/insighthub.v1.ReportService/`
- **InsightViewer Service**: `http://localhost:8083/connect/insightviewer.v1.ReportExecutionService/`

## Configuration

Services use environment variables or config files:

```yaml
# config.yaml
database:
  host: localhost
  port: 5432
  name: ugcl_reports
  user: postgres
  password: password

server:
  port: 8081
  cors_enabled: true

cache:
  redis_url: redis://localhost:6379
  ttl: 3600
```

## Deployment

### Docker Deployment
```bash
# Build images
docker build -t ugcl/masters ./masters
docker build -t ugcl/insighthub ./insighthub
docker build -t ugcl/insightviewer ./insightviewer

# Run with docker-compose
docker-compose up -d
```

### Kubernetes Deployment
```bash
# Apply manifests
kubectl apply -f k8s/
```

## Monitoring & Observability

- **Metrics**: Prometheus metrics on `/metrics`
- **Health Checks**: `/health` endpoint on each service
- **Logging**: Structured JSON logging
- **Tracing**: OpenTelemetry support

## Security

- **Authentication**: JWT tokens
- **Authorization**: Role-based access control
- **Data Classification**: Field-level security
- **Audit Logging**: All operations logged
- **Rate Limiting**: Per-user and per-endpoint
- **CORS**: Configurable cross-origin policies

## Performance

- **Caching**: Redis for query results
- **Connection Pooling**: Database connections
- **Query Optimization**: Automatic index suggestions
- **Streaming**: Large result sets streamed
- **Compression**: gRPC compression enabled

## Contributing

1. Follow clean architecture principles
2. Use conventional commit messages
3. Add tests for new features
4. Update documentation
5. Run linters and formatters

```bash
# Format code
make fmt

# Run linters
make lint

# Run tests
make test
```

---

This Report Builder & Viewer system provides a robust foundation for data analytics and reporting in your UGCL ERP/SCADA system!