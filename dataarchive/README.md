# Data Archive & Retention Module

## Module Overview

The Data Archive module provides comprehensive data lifecycle management, retention policy enforcement, and compliance auditing capabilities for the UGCL system. It automates the archiving of historical data, enforces regulatory retention requirements, manages legal holds, and provides detailed compliance reporting.

### Key Features

- **Retention Policy Management**: Time-based, size-based, event-based, and compliance-driven retention policies
- **Automated Archival**: Scheduled execution with batch processing and parallel job support
- **Legal Hold Management**: Track and enforce legal holds on data with case management
- **Data Inventory**: Automated discovery and classification of data sources
- **Compression & Encryption**: Multiple compression algorithms with AES-256 encryption
- **Compliance Auditing**: Automated compliance checks with violation tracking
- **Storage Tiering**: Hot, warm, cold, and frozen storage tiers
- **Data Restoration**: Controlled restoration with approval workflows
- **Access Logging**: Complete audit trail of archived data access

## Architecture

### Module Structure

```
dataarchive/
├── api/v1/dataarchive/       # Generated gRPC code
├── db/
│   ├── generated/            # SQLC generated queries
│   ├── queries/              # SQL queries
│   └── schema/               # Database schema
├── handlers/                 # gRPC handlers
├── models/                   # Domain models
│   └── retention_policy.go
├── proto/                    # Protobuf definitions
├── repository/               # Data access layer
├── services/                 # Business logic
│   ├── retention_policy_manager.go
│   ├── archival_executor.go
│   └── compliance_auditor.go
└── module.go                 # Dependency injection
```

### Database Schema

#### Core Tables

1. **retention_policies** - Retention policy definitions
   - Scope (entity type, schema, database)
   - Retention periods (retention, archival, grace)
   - Archival methods (cold storage, compression, export, delete)
   - Compliance and legal hold settings
   - Scheduling configuration

2. **archival_jobs** - Archival job execution tracking
   - Job status and progress
   - Record counts and size metrics
   - Compression statistics
   - Error handling

3. **legal_holds** - Legal hold management
   - Case information and legal basis
   - Affected entities and date ranges
   - Impact on retention policies
   - Audit trail

4. **data_inventory** - Data discovery and classification
   - Table metadata and statistics
   - Data classification (PII, PHI, PCI)
   - Regulatory requirements
   - Scan results

5. **archived_data** - Archived data tracking
   - Storage location and format
   - Encryption and checksums
   - Legal hold status
   - Access tracking

6. **compliance_audits** - Compliance audit records
   - Audit results and scores
   - Violations and recommendations
   - Report generation

7. **archive_access_logs** - Access audit trail
   - User access tracking
   - Purpose and justification
   - Approval workflows
   - Download tracking

### Dependencies

- **PostgreSQL**: Database with UUID and pgcrypto extensions
- **Scheduler**: Automated policy execution
- **Notification**: Compliance alerts and notifications
- **Storage**: Multi-tier storage backends

## Quick Start

### Creating a Retention Policy

```go
import dataarchive "p9e.in/ugcl/dataarchive/api/v1/dataarchive"

client := dataarchive.NewDataArchiveServiceClient(conn)

policy := &dataarchive.RetentionPolicy{
    Name:         "transaction-data-retention",
    Description:  "7-year retention for financial transactions",
    EntityType:   "transactions",
    SchemaName:   "financial",
    DatabaseName: "ugcl_production",

    RetentionPeriod: &duration.Duration{
        Seconds: 7 * 365 * 24 * 60 * 60, // 7 years
    },
    ArchivalPeriod: &duration.Duration{
        Seconds: 1 * 365 * 24 * 60 * 60, // Archive after 1 year
    },
    GracePeriod: &duration.Duration{
        Seconds: 30 * 24 * 60 * 60, // 30 days grace
    },

    PolicyType:      dataarchive.PolicyType_POLICY_TYPE_COMPLIANCE_RETENTION,
    ArchivalMethod:  dataarchive.ArchivalMethod_ARCHIVAL_METHOD_COLD_STORAGE,
    CompressionType: dataarchive.CompressionType_COMPRESSION_TYPE_ZSTD,
    EncryptArchive:  true,

    SelectionCriteria: &structpb.Struct{
        Fields: map[string]*structpb.Value{
            "date_field": structpb.NewStringValue("transaction_date"),
            "conditions": structpb.NewListValue(&structpb.ListValue{
                Values: []*structpb.Value{
                    structpb.NewStringValue("status = 'completed'"),
                },
            }),
        },
    },

    LegalHoldEnabled: true,
    ComplianceLevel:  dataarchive.ComplianceLevel_COMPLIANCE_LEVEL_HIGH,
    RegulatoryBasis:  []string{"SOX", "GDPR", "ISO27001"},

    ScheduleEnabled: true,
    ScheduleCron:    "0 2 * * 0", // Weekly on Sunday at 2 AM
    BatchSize:       10000,
    ParallelJobs:    2,

    NotifyOnComplete: true,
    NotifyOnError:    true,
    NotificationChannels: []string{"EMAIL", "SLACK"},

    IsActive: true,
}

resp, err := client.CreateRetentionPolicy(ctx, &dataarchive.CreateRetentionPolicyRequest{
    Policy: policy,
})
```

### Executing a Retention Policy

```go
// Execute policy immediately with optional dry-run
resp, err := client.ExecuteRetentionPolicy(ctx, &dataarchive.ExecuteRetentionPolicyRequest{
    PolicyId: "550e8400-e29b-41d4-a716-446655440000",
    DryRun:   false, // Set to true to preview without archiving
    DateRange: &dataarchive.DateRange{
        StartDate: timestamppb.New(time.Now().AddDate(-2, 0, 0)),
        EndDate:   timestamppb.New(time.Now().AddDate(-1, 0, 0)),
    },
})

if err != nil {
    log.Fatalf("Failed to execute policy: %v", err)
}

// Monitor job progress
jobID := resp.Job.Id
for {
    jobResp, err := client.GetArchivalJob(ctx, &dataarchive.GetArchivalJobRequest{
        Id: jobID,
    })

    if err != nil {
        log.Printf("Error checking job: %v", err)
        break
    }

    job := jobResp.Job
    log.Printf("Archival progress: %d/%d records (%.2f compression ratio)",
        job.ProcessedRecords, job.TotalRecords, job.CompressionRatio)

    if job.Status == dataarchive.JobStatus_JOB_STATUS_COMPLETED {
        log.Printf("Archival completed! Archived: %d, Deleted: %d, Errors: %d",
            job.ArchivedRecords, job.DeletedRecords, job.ErrorCount)
        break
    } else if job.Status == dataarchive.JobStatus_JOB_STATUS_FAILED {
        log.Printf("Archival failed: %s", job.ErrorMessage)
        break
    }

    time.Sleep(10 * time.Second)
}
```

## API Reference

### DataArchiveService RPCs

#### Retention Policy Management

- **CreateRetentionPolicy** - Create a new retention policy with compliance rules
- **GetRetentionPolicy** - Retrieve a specific retention policy
- **UpdateRetentionPolicy** - Update policy configuration
- **DeleteRetentionPolicy** - Delete a retention policy
- **ListRetentionPolicies** - List all policies with filtering
- **ExecuteRetentionPolicy** - Manually execute a retention policy

#### Archival Job Management

- **GetArchivalJob** - Get archival job details and progress
- **ListArchivalJobs** - List archival jobs with status filtering
- **CancelArchivalJob** - Cancel a running archival job
- **RetryArchivalJob** - Retry a failed archival job

#### Legal Hold Management

- **CreateLegalHold** - Create a legal hold on data
- **GetLegalHold** - Retrieve legal hold details
- **UpdateLegalHold** - Update legal hold information
- **DeleteLegalHold** - Remove a legal hold
- **ListLegalHolds** - List all legal holds with filtering

#### Data Inventory Management

- **CreateDataInventory** - Manually create inventory entry
- **GetDataInventory** - Get inventory for specific table
- **UpdateDataInventory** - Update inventory information
- **ListDataInventories** - List data inventory with filters
- **ScanDataSources** - Trigger automated data discovery scan

#### Archived Data Management

- **SearchArchivedData** - Search archived data with filters
- **RestoreArchivedData** - Request restoration of archived data
- **GetArchivedDataInfo** - Get details and access logs
- **DeleteArchivedData** - Permanently delete archived data

#### Compliance and Auditing

- **CreateComplianceAudit** - Initiate a compliance audit
- **GetComplianceAudit** - Get audit results
- **ListComplianceAudits** - List all audits
- **GenerateComplianceReport** - Generate compliance report

#### Analytics and Reporting

- **GetArchiveMetrics** - Get archival statistics
- **GetStorageSavings** - Calculate storage savings from compression
- **GetRetentionDashboard** - Get comprehensive dashboard data

## Database Schema

### Key Indexes

```sql
-- Policy lookups
CREATE INDEX idx_retention_policies_entity_type ON retention_policies(entity_type);
CREATE INDEX idx_retention_policies_next_execution ON retention_policies(next_execution)
    WHERE next_execution IS NOT NULL AND is_active = true;

-- Archival job tracking
CREATE INDEX idx_archival_jobs_policy_id ON archival_jobs(policy_id);
CREATE INDEX idx_archival_jobs_status ON archival_jobs(status);
CREATE INDEX idx_archival_jobs_target_table ON archival_jobs(target_table);

-- Legal hold enforcement
CREATE INDEX idx_legal_holds_active ON legal_holds(is_active) WHERE is_active = true;
CREATE INDEX idx_legal_holds_entity_types ON legal_holds USING GIN(entity_types);

-- Archived data queries
CREATE INDEX idx_archived_data_source ON archived_data(source_database, source_schema, source_table);
CREATE INDEX idx_archived_data_expires_at ON archived_data(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX idx_archived_data_legal_hold ON archived_data(is_on_legal_hold) WHERE is_on_legal_hold = true;

-- Data classification
CREATE INDEX idx_data_inventory_pii ON data_inventory(contains_pii) WHERE contains_pii = true;
```

### Database Functions

```sql
-- Check if data is subject to legal hold
SELECT is_data_on_legal_hold('transactions', 'txn-123', '2024-01-15'::timestamp);

-- Get policies due for execution
SELECT * FROM get_policies_due_for_execution();

-- Calculate storage savings
SELECT * FROM calculate_storage_savings('2024-01-01', '2024-12-31');
```

### Views

```sql
-- Retention policy dashboard
SELECT * FROM retention_policy_dashboard WHERE is_overdue = true;

-- Compliance overview
SELECT * FROM compliance_overview WHERE has_active_policy = false;
```

## Configuration

### Environment Variables

```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=ugcl
DB_USER=ugcl_user
DB_PASSWORD=secure_password

# Archive Storage
ARCHIVE_STORAGE_PATH=/mnt/archives
ARCHIVE_COMPRESSION=ZSTD
ARCHIVE_ENCRYPTION_ENABLED=true
ARCHIVE_ENCRYPTION_KEY_ID=archive-key-001

# Default Retention
DEFAULT_RETENTION_DAYS=2555  # ~7 years
DEFAULT_GRACE_PERIOD_DAYS=30
DEFAULT_BATCH_SIZE=10000
MAX_PARALLEL_JOBS=3

# Compliance
COMPLIANCE_AUDIT_ENABLED=true
COMPLIANCE_AUDIT_SCHEDULE="0 3 1 * *"  # Monthly at 3 AM
COMPLIANCE_REPORT_FORMAT=PDF

# Legal Hold
LEGAL_HOLD_NOTIFICATION_ENABLED=true
LEGAL_HOLD_ALERT_CHANNELS=EMAIL,SLACK

# Data Classification
AUTO_SCAN_ENABLED=true
SCAN_SCHEDULE="0 4 * * 0"  # Weekly on Sunday at 4 AM
PII_DETECTION_ENABLED=true
```

## Examples

### Creating a Legal Hold

```go
legalHold := &dataarchive.LegalHold{
    Name:         "SEC Investigation 2024-001",
    Description:  "Hold all financial data for SEC investigation",
    EntityTypes:  []string{"transactions", "ledger_entries", "invoices"},
    EntityIds:    []string{}, // Empty means all entities

    DateRange: &dataarchive.DateRange{
        StartDate: timestamppb.New(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
        EndDate:   timestamppb.New(time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)),
    },

    CaseNumber:        "SEC-2024-001",
    LegalBasis:        "SEC investigation pursuant to Section 21(a)",
    IssuingAuthority:  "Securities and Exchange Commission",
    ContactInfo:       "legal@example.com, +1-555-0100",

    IsActive:   true,
    StartDate:  timestamppb.New(time.Now()),
    EndDate:    nil, // Open-ended

    AffectedPolicies: []string{}, // Will be populated automatically
}

resp, err := client.CreateLegalHold(ctx, &dataarchive.CreateLegalHoldRequest{
    LegalHold: legalHold,
})

if err != nil {
    log.Fatalf("Failed to create legal hold: %v", err)
}

log.Printf("Legal hold created: %s", resp.LegalHold.Id)
log.Printf("Affected policies: %v", resp.LegalHold.AffectedPolicies)
```

### Running Data Discovery Scan

```go
// Scan all databases for data classification
scanResp, err := client.ScanDataSources(ctx, &dataarchive.ScanDataSourcesRequest{
    DatabaseNames: []string{"ugcl_production", "ugcl_analytics"},
    SchemaNames:   []string{"public", "financial", "hr"},
    DeepScan:      true, // Enable deep content analysis
})

if err != nil {
    log.Fatalf("Scan failed: %v", err)
}

log.Printf("Discovered %d tables", scanResp.DiscoveredTables)

for _, inv := range scanResp.Inventories {
    log.Printf("Table: %s.%s.%s", inv.DatabaseName, inv.SchemaName, inv.TableName)
    log.Printf("  Records: %d, Size: %.2f MB",
        inv.RecordCount, float64(inv.DataSize)/(1024*1024))
    log.Printf("  Classification: %v", inv.DataClassification)
    log.Printf("  Contains PII: %v, PHI: %v, PCI: %v",
        inv.ContainsPii, inv.ContainsPhi, inv.ContainsPci)
    log.Printf("  Regulatory: %v", inv.RegulatoryRequirements)
}
```

### Generating Compliance Report

```go
report, err := client.GenerateComplianceReport(ctx, &dataarchive.GenerateComplianceReportRequest{
    PolicyIds: []string{
        "policy-1",
        "policy-2",
        "policy-3",
    },
    DateRange: &dataarchive.DateRange{
        StartDate: timestamppb.New(time.Now().AddDate(-1, 0, 0)),
        EndDate:   timestamppb.New(time.Now()),
    },
    ReportFormat: "PDF",
    Regulations:  []string{"GDPR", "SOX", "HIPAA"},
})

if err != nil {
    log.Fatalf("Failed to generate report: %v", err)
}

log.Printf("Compliance report generated: %s", report.ReportId)
log.Printf("Download URL: %s", report.DownloadUrl)
log.Printf("Expires at: %v", report.ExpiresAt)
```

### Searching and Restoring Archived Data

```go
// Search archived data
searchResp, err := client.SearchArchivedData(ctx, &dataarchive.SearchArchivedDataRequest{
    PageSize:    100,
    SourceTable: "transactions",
    ArchivedDateRange: &dataarchive.DateRange{
        StartDate: timestamppb.New(time.Now().AddDate(-1, 0, 0)),
        EndDate:   timestamppb.New(time.Now()),
    },
    StorageTier:      dataarchive.StorageTier_STORAGE_TIER_COLD,
    OnLegalHoldOnly:  false,
})

if err != nil {
    log.Fatalf("Search failed: %v", err)
}

log.Printf("Found %d archived records", searchResp.TotalCount)

// Request restoration of specific record
if len(searchResp.ArchivedData) > 0 {
    archived := searchResp.ArchivedData[0]

    restoreResp, err := client.RestoreArchivedData(ctx, &dataarchive.RestoreArchivedDataRequest{
        ArchivedDataId:  archived.Id,
        TargetLocation:  "restore_temp",
        RequestedBy:     "user-123",
        Justification:   "Required for audit investigation",
    })

    if err != nil {
        log.Fatalf("Restore request failed: %v", err)
    }

    log.Printf("Restore job initiated: %s", restoreResp.RestoreJobId)
    log.Printf("Estimated completion: %s", restoreResp.EstimatedCompletion)
}
```

## Integration

### With Scheduler Module

Automated policy execution:

```go
// Register retention policy with scheduler
func (m *RetentionPolicyManager) RegisterPolicy(policy *models.RetentionPolicy) error {
    if !policy.ScheduleEnabled {
        return nil
    }

    return m.schedulerClient.CreateJob(ctx, &scheduler.CreateJobRequest{
        Name:           fmt.Sprintf("retention-policy-%s", policy.ID),
        Description:    fmt.Sprintf("Execute retention policy: %s", policy.Name),
        CronExpression: policy.ScheduleCron,
        TargetService:  "dataarchive.DataArchiveService",
        TargetMethod:   "ExecuteRetentionPolicy",
        TargetPayload: &structpb.Struct{
            Fields: map[string]*structpb.Value{
                "policy_id": structpb.NewStringValue(policy.ID),
                "dry_run":   structpb.NewBoolValue(false),
            },
        },
        TimeoutSeconds: 7200, // 2 hours
    })
}
```

### With Notification Module

Compliance alerts:

```go
// Send compliance violation notification
func (a *ComplianceAuditor) notifyViolation(audit *models.ComplianceAudit, violation string) error {
    return a.notificationClient.SendNotification(ctx, &notification.SendNotificationRequest{
        Notification: &notification.Notification{
            Type:         notification.NotificationType_SYSTEM_ALERT,
            Channel:      notification.NotificationChannel_EMAIL,
            Priority:     notification.NotificationPriority_HIGH,
            RecipientIds: []string{"compliance-team"},
            Subject:      fmt.Sprintf("Compliance Violation Detected: %s", audit.AuditType),
            Message:      formatViolationMessage(audit, violation),
        },
    })
}
```

### With Legal Module

Legal hold integration:

```go
// Check legal hold before archival
func (e *ArchivalExecutor) canArchive(entityType, entityID string, date time.Time) (bool, error) {
    // Query database function
    var isOnHold bool
    err := e.db.QueryRow(`
        SELECT is_data_on_legal_hold($1, $2, $3)
    `, entityType, entityID, date).Scan(&isOnHold)

    if err != nil {
        return false, err
    }

    return !isOnHold, nil
}
```

## Development

### Running Tests

```bash
# Unit tests
go test ./dataarchive/...

# Integration tests
go test -tags=integration ./dataarchive/...

# Test specific service
go test ./dataarchive/services/...
```

### Code Generation

```bash
# Generate protobuf code
cd dataarchive
buf generate

# Generate SQLC queries
cd dataarchive/db
sqlc generate
```

### Database Migrations

```bash
# Create migration
migrate create -ext sql -dir dataarchive/db/migrations -seq add_data_classification

# Apply migrations
migrate -path dataarchive/db/migrations -database "postgresql://localhost:5432/ugcl" up
```

## Troubleshooting

### Common Issues

#### 1. Archival Job Failing with "Data on Legal Hold"

**Symptoms**: Jobs fail with legal hold errors

**Solution**:
```sql
-- Check active legal holds
SELECT lh.name, lh.case_number, lh.entity_types, lh.date_range_start, lh.date_range_end
FROM legal_holds lh
WHERE lh.is_active = true
  AND 'your_entity_type' = ANY(lh.entity_types);

-- Update legal hold if resolved
UPDATE legal_holds
SET is_active = false,
    end_date = NOW()
WHERE id = 'legal-hold-id';
```

#### 2. Low Compression Ratios

**Symptoms**: Archived data not significantly compressed

**Diagnostic**:
```sql
SELECT
    aj.target_table,
    AVG(aj.compression_ratio) as avg_compression,
    COUNT(*) as job_count
FROM archival_jobs aj
WHERE aj.status = 'COMPLETED'
  AND aj.started_at > NOW() - INTERVAL '30 days'
GROUP BY aj.target_table
ORDER BY avg_compression ASC;
```

**Solutions**:
- Switch to ZSTD compression (best ratio)
- Verify data is not already compressed
- Check if data contains binary/encrypted content
- Consider different archival format (Parquet, ORC)

#### 3. Storage Tier Migration Issues

**Symptoms**: Data not moving between storage tiers

**Solution**:
```sql
-- Check data eligible for tier migration
SELECT
    storage_tier,
    COUNT(*) as count,
    SUM(data_size) as total_size,
    MIN(archived_at) as oldest,
    MAX(archived_at) as newest
FROM archived_data
GROUP BY storage_tier;

-- Manually migrate old data to frozen tier
UPDATE archived_data
SET storage_tier = 'FROZEN'
WHERE storage_tier = 'COLD'
  AND archived_at < NOW() - INTERVAL '2 years'
  AND is_on_legal_hold = false;
```

#### 4. Compliance Audit Failures

**Diagnostic**:
```sql
-- Review recent audit results
SELECT
    id,
    audit_type,
    status,
    compliance_score,
    violations
FROM compliance_audits
WHERE status = 'FAILED'
ORDER BY started_at DESC
LIMIT 10;
```

**Common Causes**:
- Missing retention policies for sensitive data
- Policies not enforced (inactive)
- Data older than retention period not archived
- Legal holds not properly tracked

#### 5. Data Restoration Performance

**Optimization**:
- Use parallel restoration for large datasets
- Restore to temporary tables first
- Schedule restorations during off-peak hours
- Consider partial restoration for analysis

```go
// Parallel restoration
var wg sync.WaitGroup
for _, archivedID := range archivedIDs {
    wg.Add(1)
    go func(id string) {
        defer wg.Done()
        client.RestoreArchivedData(ctx, &dataarchive.RestoreArchivedDataRequest{
            ArchivedDataId: id,
            TargetLocation: "restore_temp",
        })
    }(archivedID)
}
wg.Wait()
```

### Monitoring Queries

```sql
-- Active archival jobs
SELECT
    aj.id,
    rp.name as policy_name,
    aj.target_table,
    aj.status,
    aj.processed_records,
    aj.total_records,
    ROUND(100.0 * aj.processed_records / NULLIF(aj.total_records, 0), 2) as progress_pct
FROM archival_jobs aj
JOIN retention_policies rp ON rp.id = aj.policy_id
WHERE aj.status IN ('PENDING', 'RUNNING')
ORDER BY aj.started_at;

-- Storage savings summary
SELECT
    SUM(data_size) as original_size,
    SUM(compressed_size) as compressed_size,
    SUM(data_size - compressed_size) as saved_space,
    ROUND(100.0 * SUM(data_size - compressed_size) / NULLIF(SUM(data_size), 0), 2) as savings_pct
FROM archived_data
WHERE archived_at > NOW() - INTERVAL '1 year';

-- Legal holds impact
SELECT
    lh.name,
    lh.case_number,
    COUNT(ad.id) as affected_records,
    SUM(ad.data_size) as total_size
FROM legal_holds lh
LEFT JOIN archived_data ad ON ad.is_on_legal_hold = true
  AND lh.id = ANY(ad.legal_hold_ids)
WHERE lh.is_active = true
GROUP BY lh.id, lh.name, lh.case_number;
```

### Best Practices

1. **Policy Design**:
   - Align retention periods with regulatory requirements
   - Use grace periods to prevent premature deletion
   - Enable legal hold for sensitive data types
   - Schedule during low-activity periods

2. **Performance**:
   - Use appropriate batch sizes (10K-100K records)
   - Enable parallel jobs for large tables
   - Monitor compression ratios
   - Archive to appropriate storage tier

3. **Compliance**:
   - Regular compliance audits (monthly minimum)
   - Document all policies with regulatory basis
   - Maintain complete access logs
   - Test restoration procedures quarterly

4. **Security**:
   - Always enable encryption for archived data
   - Use KMS for key management
   - Implement approval workflows for restoration
   - Log all access to archived data
