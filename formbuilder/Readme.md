# Form Builder and Approval Workflows Module

## Module Overview

The Form Builder module provides a comprehensive form management system with dynamic form creation, validation, workflow integration, and approval management capabilities. It enables users to design complex forms with conditional logic, field dependencies, validation rules, and integrated approval workflows with SLA tracking.

### Key Features

- **Dynamic Form Builder**: Create forms programmatically with extensive field types and validation
- **Conditional Logic**: Field dependencies, show/hide rules, and cross-field validations
- **Workflow Integration**: Built-in workflow engine with state transitions and approval steps
- **SLA Tracking**: Service Level Agreement monitoring with escalation and breach notifications
- **Approval Management**: Multi-level approval workflows with delegation and audit trails
- **Field Types**: 17+ field types including TEXT, NUMBER, EMAIL, DATE, FILE, DROPDOWN, JSON, NESTED_FORM
- **API Integration**: Dynamic dropdown options from external APIs with caching
- **Audit Logging**: Complete audit trail of all form instance changes
- **Attachment Management**: File upload support with metadata tracking
- **Comment System**: Internal and external comments on form instances

## Architecture

### Module Structure

```
formbuilder/
├── api/v2/
│   ├── form_builder/         # Generated gRPC code for forms
│   ├── workflow/             # Generated gRPC code for workflows
│   └── approval/             # Generated gRPC code for approvals
├── db/
│   ├── generated/            # SQLC generated queries
│   ├── queries/              # SQL queries
│   └── schema/               # Database schema
├── handlers/                 # gRPC handlers
│   ├── approval.go
│   └── approval_helpers.go
├── models/                   # Domain models
├── proto/                    # Protobuf definitions
│   ├── formbuilder.proto
│   ├── forminstance.proto
│   ├── workflow.proto
│   └── approval.proto
├── services/                 # Business logic
│   ├── iservices.go
│   ├── types.go
│   └── approval.go
└── module.go                 # Dependency injection
```

### Database Schema

#### Core Tables

1. **forms** - Form definitions and metadata
   - Form structure (steps, fields, dependencies)
   - Versioning support (schema_version)
   - Workflow integration (workflow_id)
   - Access control (allowed_roles)
   - Audit settings

2. **form_instances** - Form submissions/data
   - Current state tracking
   - Field values (JSONB)
   - Assignment tracking
   - Audit metadata
   - Soft delete support

3. **workflows** - Workflow state machines
   - Initial state configuration
   - State definitions
   - Global transitions
   - Metadata and versioning

4. **sla_rules** - Service Level Agreements
   - State-specific rules
   - Duration thresholds
   - Escalation levels
   - Applicable roles
   - Conditional logic

5. **escalations** - Escalation definitions
   - Trigger conditions
   - State transitions
   - Notification rules
   - Auto-escalation settings

6. **audit_logs** - Change tracking
   - User actions
   - State transitions
   - Field changes
   - IP address and user agent
   - Timestamp tracking

7. **approval_actions** - Approval decisions
   - Action types (APPROVE, REJECT, DELEGATE, REQUEST_INFO, WITHDRAW, REASSIGN)
   - Comments and attachments
   - Delegation tracking
   - IP and user agent logging

8. **approval_delegates** - Delegation management
   - Delegator-delegate relationships
   - Entity type filtering
   - Date range validity
   - Active status tracking

9. **approval_reports** - Analytics and metrics
   - Daily aggregations
   - Processing time statistics
   - SLA compliance rates
   - Bottleneck identification

### Dependencies

- **PostgreSQL**: Database with UUID and pgcrypto extensions
- **Notification**: Alert delivery for approvals and SLA events
- **Scheduler**: SLA monitoring and escalation processing

## Quick Start

### Creating a Form

```go
import (
    formbuilder "p9e.in/ugcl/formbuilder/api/v2/form_builder"
    "google.golang.org/protobuf/types/known/anypb"
)

client := formbuilder.NewFormBuilderClient(conn)

// Create form definition
form := &formbuilder.FormDefinition{
    Metadata: &formbuilder.FormMetadata{
        Title: "Employee Onboarding",
        Description: "New employee onboarding form",
        AllowedRoles: []string{"HR", "Manager"},
        Audit: true,
        Module: "hr",
    },
    Steps: []*formbuilder.FormStep{
        {
            Id: "personal_info",
            Label: "Personal Information",
            Fields: []*formbuilder.FormField{
                {
                    Id: "full_name",
                    Type: formbuilder.FieldType_TEXT,
                    Label: "Full Name",
                    Required: true,
                    Validation: &formbuilder.Validation{
                        MinLength: 2,
                        MaxLength: 100,
                    },
                },
                {
                    Id: "email",
                    Type: formbuilder.FieldType_EMAIL,
                    Label: "Email Address",
                    Required: true,
                },
                {
                    Id: "department",
                    Type: formbuilder.FieldType_DROPDOWN,
                    Label: "Department",
                    Required: true,
                    OptionsSource: "/api/departments",
                },
            },
        },
    },
}

resp, err := client.CreateForm(ctx, &formbuilder.CreateFormRequest{
    FormDefinition: form,
    CreateTable: true,
})
```

### Starting an Approval Process

```go
import approval "p9e.in/ugcl/formbuilder/api/v2/approval"

approvalClient := approval.NewApprovalServiceClient(conn)

resp, err := approvalClient.StartApprovalProcess(ctx, &approval.StartApprovalProcessRequest{
    EntityType: "employee_onboarding",
    EntityId: instanceId,
    RequestedBy: userId,
    Priority: approval.Priority_PRIORITY_NORMAL,
    Comments: "Please review new employee onboarding",
    Metadata: requestMetadata,
})
```

## API Reference

### FormBuilder Service

#### CreateForm
Creates a new form definition with optional database table generation.

**Request:**
```protobuf
message CreateFormRequest {
  FormDefinition form_definition = 1;
  bool create_table = 2;
}
```

**Response:**
```protobuf
message CreateFormResponse {
  string form_id = 1;
  string table_name = 2;
  bool success = 3;
  string message = 4;
}
```

#### GetForm / UpdateForm / DeleteForm / ListForms
Standard CRUD operations for form management.

#### GetFieldOptions
Retrieves dropdown options for a field with caching support.

**Request:**
```protobuf
message GetFieldOptionsRequest {
  string form_id = 1;
  string field_id = 2;
  map<string, string> context = 3;
}
```

### Approval Service

#### StartApprovalProcess
Initiates a new approval workflow.

**Request:**
```protobuf
message StartApprovalProcessRequest {
  string entity_type = 1;
  string entity_id = 2;
  string requested_by = 3;
  Priority priority = 4;
  google.protobuf.Any request_data = 5;
  string comments = 6;
}
```

#### ProcessApprovalAction
Processes an approval action (approve, reject, delegate).

**Request:**
```protobuf
message ProcessApprovalActionRequest {
  string instance_id = 1;
  string approver_id = 2;
  ActionType action = 3; // APPROVE, REJECT, DELEGATE, REQUEST_INFO
  string comments = 4;
  string ip_address = 5;
}
```

#### GetApprovalMetrics
Retrieves approval analytics.

**Response:**
```protobuf
message GetApprovalMetricsResponse {
  int32 total_requests = 1;
  int32 approved_count = 2;
  int32 rejected_count = 3;
  int32 pending_count = 4;
  int32 escalated_count = 5;
  double avg_processing_time_hours = 6;
  double sla_compliance_rate = 7;
  repeated MetricData daily_metrics = 8;
  repeated BottleneckInfo bottlenecks = 9;
}
```

## Database Schema

### Key Tables

```sql
-- Forms table
CREATE TABLE forms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    version VARCHAR(50) NOT NULL DEFAULT '1.0.0',
    created_by VARCHAR(255) NOT NULL,
    allowed_roles JSONB DEFAULT '[]'::jsonb,
    audit BOOLEAN DEFAULT false,
    table_name VARCHAR(255),
    module VARCHAR(255),
    schema_version INTEGER DEFAULT 1,
    core_fields JSONB DEFAULT '[]'::jsonb,
    steps JSONB NOT NULL DEFAULT '[]'::jsonb,
    dependencies JSONB DEFAULT '[]'::jsonb,
    cross_field_validations JSONB DEFAULT '[]'::jsonb,
    workflow_id UUID REFERENCES workflows(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Form instances table
CREATE TABLE form_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    form_id UUID NOT NULL REFERENCES forms(id) ON DELETE CASCADE,
    current_state VARCHAR(100) NOT NULL DEFAULT 'draft',
    field_values JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_by VARCHAR(255) NOT NULL,
    assigned_to VARCHAR(255),
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Approval actions table
CREATE TABLE approval_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    form_instance_id UUID NOT NULL REFERENCES form_instances(id) ON DELETE CASCADE,
    approver_id VARCHAR(255) NOT NULL,
    action VARCHAR(50) NOT NULL,
    comments TEXT,
    acted_at TIMESTAMPTZ DEFAULT NOW(),
    ip_address INET,
    user_agent TEXT,
    delegated_from VARCHAR(255),
    attachment_urls JSONB DEFAULT '[]'::jsonb,
    metadata JSONB DEFAULT '{}'::jsonb,
    CHECK (action IN ('APPROVE', 'REJECT', 'DELEGATE', 'REQUEST_INFO', 'WITHDRAW', 'REASSIGN'))
);
```

## Configuration

### Environment Variables

```bash
# Database
FORMBUILDER_DB_HOST=localhost
FORMBUILDER_DB_PORT=5432
FORMBUILDER_DB_NAME=ugcl_formbuilder
FORMBUILDER_DB_USER=postgres
FORMBUILDER_DB_PASSWORD=secret

# Service
FORMBUILDER_GRPC_PORT=50051
FORMBUILDER_HTTP_PORT=8080

# Cache
FORMBUILDER_CACHE_ENABLED=true
FORMBUILDER_CACHE_TTL=3600
FORMBUILDER_REDIS_URL=redis://localhost:6379

# SLA Processing
FORMBUILDER_SLA_CHECK_INTERVAL=60s
FORMBUILDER_SLA_ENABLED=true

# Notifications
FORMBUILDER_NOTIFICATION_ENABLED=true
```

## Examples

### Multi-Step Form with Conditional Logic

```go
form := &formbuilder.FormDefinition{
    Metadata: &formbuilder.FormMetadata{
        Title: "Purchase Requisition",
        Description: "Request for purchase approval",
        AllowedRoles: []string{"employee", "manager", "finance"},
        Audit: true,
    },
    Steps: []*formbuilder.FormStep{
        {
            Id: "requisition_details",
            Label: "Requisition Details",
            Order: 1,
            Fields: []*formbuilder.FormField{
                {
                    Id: "item_description",
                    Type: formbuilder.FieldType_TEXTAREA,
                    Label: "Item Description",
                    Required: true,
                    Validation: &formbuilder.Validation{
                        MinLength: 10,
                        MaxLength: 500,
                    },
                },
                {
                    Id: "quantity",
                    Type: formbuilder.FieldType_NUMBER,
                    Label: "Quantity",
                    Required: true,
                },
                {
                    Id: "unit_price",
                    Type: formbuilder.FieldType_CURRENCY,
                    Label: "Unit Price",
                    Required: true,
                },
            },
        },
    },
    CrossFieldValidations: []*formbuilder.CrossFieldValidation{
        {
            Fields: []string{"quantity", "unit_price", "total_amount"},
            Rule: "total_amount == quantity * unit_price",
            Message: "Total amount must equal quantity × unit price",
        },
    },
}
```

### Managing Approval Delegation

```go
// Create delegation
delegateResp, err := approvalClient.CreateDelegate(ctx, &approval.CreateDelegateRequest{
    Delegate: &approval.ApprovalDelegate{
        DelegatorId: "user123",
        DelegateId: "user456",
        EntityTypes: []string{"purchase_requisition", "leave_request"},
        StartDate: timestamppb.Now(),
        EndDate: timestamppb.New(time.Now().AddDate(0, 1, 0)),
        Reason: "On vacation",
        IsActive: true,
    },
})
```

## Integration

### With Notification Service

Approval notifications are automatically sent via the notification service:

```go
// Configure in workflow definition
workflow := &formbuilder.Workflow{
    States: []*formbuilder.State{
        {
            Name: "pending_approval",
            Actions: []*formbuilder.Action{
                {
                    Type: "notification",
                    Config: map[string]string{
                        "channel": "email",
                        "template": "approval_request",
                    },
                },
            },
        },
    },
}
```

### With Scheduler Service

SLA monitoring runs via scheduled job:

```go
escalation := &formbuilder.Escalation{
    Name: "Management Escalation",
    TriggerCondition: "time_in_state > 48h",
    FromStates: []string{"pending_approval"},
    ToState: "escalated",
    AutoEscalate: true,
}
```

## Development

### Running Locally

```bash
# Install dependencies
go mod download

# Generate code from proto files
make proto-gen

# Generate SQLC code
make sqlc-gen

# Run migrations
make migrate-up

# Run tests
make test

# Start service
make run
```

### Database Migrations

```bash
# Create new migration
migrate create -ext sql -dir db/migrations -seq add_approval_metrics

# Run migrations
migrate -path db/migrations -database "postgresql://user:pass@localhost/db" up
```

### Testing

```bash
# Unit tests
go test ./...

# Integration tests
go test -tags=integration ./...

# Coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Troubleshooting

### Form Creation Fails

**Problem**: Form creation returns validation errors

**Solution**:
- Check that form metadata has title and created_by
- Verify all required fields in steps have proper validation
- Ensure workflow_id references existing workflow if specified

### Approval Actions Not Processing

**Problem**: Approval actions fail silently

**Solution**:
```sql
-- Check approval instance status
SELECT * FROM form_instances WHERE id = 'instance_id';

-- Verify approver permissions
SELECT * FROM approval_delegates
WHERE delegator_id = 'user_id' AND is_active = true;
```

### SLA Escalations Not Triggering

**Problem**: SLA escalations don't occur as expected

**Solution**:
```sql
-- Verify escalation configuration
SELECT * FROM escalations WHERE auto_escalate = true;

-- Check SLA processing job
SELECT * FROM scheduler.jobs WHERE target_method = 'ProcessEscalations';
```

### Field Options Not Loading

**Problem**: Dropdown fields show no options

**Solution**:
```go
// Clear cache
resp, err := client.RefreshFieldCache(ctx, &formbuilder.RefreshCacheRequest{
    FormId: formId,
    FieldId: fieldId,
})
```

### Performance Optimization

```sql
-- Add indexes for common queries
CREATE INDEX idx_form_instances_state_created
    ON form_instances(current_state, created_at DESC);

CREATE INDEX idx_approval_actions_approver_acted
    ON approval_actions(approver_id, acted_at DESC);
```

### Health Checks

```bash
# gRPC health check
grpcurl -plaintext localhost:50051 grpc.health.v1.Health/Check

# Database connectivity
psql -h localhost -U postgres -d ugcl_formbuilder -c "SELECT 1"
```
