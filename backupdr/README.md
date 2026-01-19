# Backup & Disaster Recovery Module

## Module Overview

The Backup & Disaster Recovery (BackupDR) module provides comprehensive backup management and disaster recovery capabilities for the UGCL system. It enables automated backup scheduling, disaster recovery planning, system health monitoring, and restore operations with support for multiple storage backends.

### Key Features

- **Automated Backup Policies**: Schedule-driven backup execution with cron expressions and custom intervals
- **Multiple Backup Types**: Full, incremental, differential, snapshot, and continuous backups
- **Disaster Recovery Planning**: Create and test comprehensive DR plans with defined RTOs and RPOs
- **Multi-Storage Support**: Local, S3, Azure Blob Storage, Google Cloud Storage, and network storage
- **Advanced Compression**: Support for GZIP, ZSTD, LZ4, and BZIP2 compression algorithms
- **Encryption**: Built-in encryption with AES-256 and configurable key management
- **System Health Monitoring**: Real-time monitoring of system health and backup status
- **Restore Workflows**: Approval-based restore requests with full audit trails
- **Analytics & Reporting**: Comprehensive metrics, dashboards, and overdue backup tracking

## Architecture

### Module Structure

```
backupdr/
├── api/v1/backupdr/          # Generated gRPC code
├── db/
│   ├── generated/            # SQLC generated queries
│   ├── queries/              # SQL queries
│   └── schema/               # Database schema
├── handlers/                 # gRPC handlers
├── models/                   # Domain models
├── proto/                    # Protobuf definitions
├── repository/               # Data access layer
├── services/                 # Business logic
│   └── backup_orchestrator.go
└── module.go                 # Dependency injection
```

### Database Schema

#### Core Tables

1. **backup_policies** - Backup policy definitions
   - Scheduling configuration (cron, interval)
   - Retention policies (daily, weekly, monthly, yearly)
   - Storage and encryption settings
   - Notification preferences

2. **backup_jobs** - Backup execution tracking
   - Job status and progress
   - File and size metrics
   - Performance statistics (throughput, CPU usage)
   - Error handling and retry logic

3. **disaster_recovery_plans** - DR plan definitions
   - RTO (Recovery Time Objective) and RPO (Recovery Point Objective)
   - Recovery procedures and steps
   - Validation and testing schedules
   - Communication and escalation plans

4. **recovery_executions** - DR execution tracking
   - Step-by-step progress
   - Success rates and metrics
   - Stakeholder notifications
   - Issue tracking and resolutions

5. **restore_requests** - Restore request management
   - Approval workflows
   - Point-in-time recovery
   - Progress tracking
   - Audit logging

6. **system_health** - System health monitoring
   - Resource utilization metrics
   - Availability tracking
   - Backup status
   - Alert thresholds

7. **backup_storage_locations** - Storage backend configuration
   - Capacity and usage tracking
   - Performance characteristics
   - Health monitoring
   - Geographic distribution

### Dependencies

- **PostgreSQL**: Database with UUID and pgcrypto extensions
- **Scheduler**: Integration for automated backup execution
- **Notification**: Alert and notification delivery
- **Storage Providers**: S3, Azure, GCP SDK support

## Quick Start

### Creating a Backup Policy

```go
import backupdr "p9e.in/ugcl/backupdr/api/v1/backupdr"

client := backupdr.NewBackupDRServiceClient(conn)

policy := &backupdr.BackupPolicy{
    Name:        "daily-database-backup",
    Description: "Daily full backup of production database",
    BackupType:  backupdr.BackupType_BACKUP_TYPE_FULL,
    TargetType:  backupdr.TargetType_TARGET_TYPE_DATABASE,
    TargetName:  "ugcl_production",
    ScheduleType: backupdr.ScheduleType_SCHEDULE_TYPE_CRON,
    CronExpression: "0 2 * * *", // Daily at 2 AM
    RetentionPolicy: &backupdr.RetentionPolicy{
        KeepDaily:   7,
        KeepWeekly:  4,
        KeepMonthly: 12,
        KeepYearly:  3,
        MaxAge:      2555, // ~7 years
    },
    StorageConfig: &backupdr.StorageConfig{
        StorageType: backupdr.StorageType_STORAGE_TYPE_S3,
        S3Config: &backupdr.S3Config{
            Region:          "us-east-1",
            Bucket:          "ugcl-backups",
            AccessKeyId:     "AKIAIOSFODNN7EXAMPLE",
            SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
            UseSSL:          true,
        },
    },
    Compression: backupdr.CompressionType_COMPRESSION_TYPE_ZSTD,
    Encryption: &backupdr.EncryptionConfig{
        Enabled:   true,
        Algorithm: "AES-256-GCM",
        KeySource: "KMS",
    },
    NotifyOnSuccess: false,
    NotifyOnFailure: true,
    NotifyChannels:  []string{"EMAIL", "SLACK"},
    IsActive:        true,
}

resp, err := client.CreateBackupPolicy(ctx, &backupdr.CreateBackupPolicyRequest{
    Policy: policy,
})
```

### Executing a Manual Backup

```go
// Execute an immediate backup
resp, err := client.ExecuteBackup(ctx, &backupdr.ExecuteBackupRequest{
    PolicyId:   "550e8400-e29b-41d4-a716-446655440000",
    BackupType: backupdr.BackupType_BACKUP_TYPE_FULL,
    Priority:   backupdr.JobPriority_JOB_PRIORITY_HIGH,
})

if err != nil {
    log.Fatalf("Failed to execute backup: %v", err)
}

// Monitor job progress
jobID := resp.Job.Id
for {
    jobResp, err := client.GetBackupJob(ctx, &backupdr.GetBackupJobRequest{
        Id: jobID,
    })

    if err != nil {
        log.Printf("Error checking job status: %v", err)
        break
    }

    job := jobResp.Job
    log.Printf("Backup progress: %.2f%% (%d/%d files)",
        job.ProgressPercent, job.ProcessedFiles, job.TotalFiles)

    if job.Status == backupdr.JobStatus_JOB_STATUS_COMPLETED {
        log.Printf("Backup completed successfully! Size: %d bytes, Compression ratio: %.2f",
            job.BackupSize, job.CompressionRatio)
        break
    } else if job.Status == backupdr.JobStatus_JOB_STATUS_FAILED {
        log.Printf("Backup failed: %s", job.ErrorMessage)
        break
    }

    time.Sleep(5 * time.Second)
}
```

## API Reference

### BackupDRService RPCs

#### Backup Policy Management

- **CreateBackupPolicy** - Create a new backup policy with scheduling and retention rules
- **GetBackupPolicy** - Retrieve a specific backup policy by ID
- **UpdateBackupPolicy** - Update an existing backup policy
- **DeleteBackupPolicy** - Delete a backup policy (restricts deletion if jobs exist)
- **ListBackupPolicies** - List all backup policies with optional filters

#### Backup Execution

- **ExecuteBackup** - Manually trigger a backup job
- **GetBackupJob** - Get details of a specific backup job
- **ListBackupJobs** - List backup jobs with status filtering
- **CancelBackupJob** - Cancel a running backup job
- **RetryBackupJob** - Retry a failed backup job

#### Restore Operations

- **CreateRestoreRequest** - Submit a new restore request
- **GetRestoreRequest** - Get restore request details
- **ListRestoreRequests** - List restore requests by user or status
- **ApproveRestoreRequest** - Approve a pending restore request
- **ExecuteRestore** - Execute an approved restore operation

#### Disaster Recovery Plans

- **CreateDRPlan** - Create a comprehensive DR plan
- **GetDRPlan** - Retrieve a DR plan with all procedures
- **UpdateDRPlan** - Update DR plan steps and configuration
- **DeleteDRPlan** - Delete a DR plan
- **ListDRPlans** - List all DR plans with filtering
- **TestDRPlan** - Execute a DR plan test

#### Recovery Execution

- **ExecuteRecovery** - Execute a disaster recovery plan
- **GetRecoveryExecution** - Get execution status and progress
- **ListRecoveryExecutions** - List all recovery executions
- **AbortRecovery** - Abort an in-progress recovery

#### System Health Monitoring

- **GetSystemHealth** - Get current health status of a system
- **ListSystemHealth** - List health status for all monitored systems
- **UpdateSystemHealth** - Update system health metrics

#### Storage Management

- **CreateStorageLocation** - Register a new backup storage location
- **GetStorageLocation** - Get storage location details
- **ListStorageLocations** - List all configured storage locations
- **TestStorageConnection** - Test connectivity to a storage location

#### Analytics and Reporting

- **GetBackupMetrics** - Get backup statistics for a time range
- **GetRecoveryMetrics** - Get disaster recovery metrics
- **GetBackupDashboard** - Get comprehensive backup dashboard data
- **GetOverdueBackups** - List policies with overdue backups

## Database Schema

### Key Indexes

```sql
-- Efficient policy lookup
CREATE INDEX idx_backup_policies_target ON backup_policies(target_type, target_name);
CREATE INDEX idx_backup_policies_next_backup ON backup_policies(next_backup)
    WHERE next_backup IS NOT NULL AND is_active = true;

-- Job monitoring
CREATE INDEX idx_backup_jobs_status ON backup_jobs(status);
CREATE INDEX idx_backup_jobs_scheduled_at ON backup_jobs(scheduled_at);

-- DR plan lookups
CREATE INDEX idx_dr_plans_type ON disaster_recovery_plans(plan_type);
CREATE INDEX idx_dr_plans_severity ON disaster_recovery_plans(severity);

-- System health tracking
CREATE INDEX idx_system_health_latest ON system_health(system_name, checked_at DESC);
```

### Database Functions

```sql
-- Get backup status summary
SELECT * FROM get_backup_status_summary(7); -- Last 7 days

-- Get overdue backups
SELECT * FROM get_overdue_backups();

-- Calculate recovery metrics
SELECT * FROM calculate_recovery_metrics('2024-01-01', '2024-12-31');
```

### Views

```sql
-- Backup dashboard view
SELECT * FROM backup_dashboard WHERE is_active = true;

-- Current system health
SELECT * FROM system_health_current WHERE health_status = 'CRITICAL';
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

# S3 Storage
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
BACKUP_S3_BUCKET=ugcl-backups

# Azure Storage
AZURE_STORAGE_ACCOUNT=ugclbackups
AZURE_STORAGE_KEY=your_storage_key
AZURE_CONTAINER_NAME=backups

# GCP Storage
GCP_PROJECT_ID=ugcl-project
GCP_BUCKET_NAME=ugcl-backups
GCP_CREDENTIALS_FILE=/path/to/credentials.json

# Backup Configuration
BACKUP_MAX_PARALLEL_JOBS=3
BACKUP_DEFAULT_TIMEOUT=28800 # 8 hours in seconds
BACKUP_RETENTION_DAYS=30
BACKUP_COMPRESSION=ZSTD

# Encryption
ENCRYPTION_ENABLED=true
ENCRYPTION_ALGORITHM=AES-256-GCM
KMS_KEY_ID=arn:aws:kms:us-east-1:123456789:key/example

# Notifications
NOTIFY_ON_FAILURE=true
NOTIFY_CHANNELS=EMAIL,SLACK
NOTIFICATION_EMAIL=backups@example.com
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/WEBHOOK/URL
```

## Examples

### Creating a DR Plan

```go
drPlan := &backupdr.DisasterRecoveryPlan{
    Name:        "Database Failover Plan",
    Description: "Complete database failover to secondary datacenter",
    PlanType:    backupdr.DRPlanType_DR_PLAN_TYPE_SYSTEM_FAILOVER,
    Severity:    backupdr.DRSeverity_DR_SEVERITY_CRITICAL,
    Scope:       "DATABASE",
    Rto:         &duration.Duration{Seconds: 3600},    // 1 hour RTO
    Rpo:         &duration.Duration{Seconds: 300},     // 5 minute RPO
    Mttr:        &duration.Duration{Seconds: 7200},    // 2 hour MTTR
    Availability: 0.9999, // 99.99% availability target

    PreRecoverySteps: &structpb.Struct{
        Fields: map[string]*structpb.Value{
            "steps": structpb.NewListValue(&structpb.ListValue{
                Values: []*structpb.Value{
                    structpb.NewStringValue("Verify latest backup integrity"),
                    structpb.NewStringValue("Notify stakeholders of DR initiation"),
                    structpb.NewStringValue("Put primary database in read-only mode"),
                },
            }),
        },
    },

    RecoverySteps: &structpb.Struct{
        Fields: map[string]*structpb.Value{
            "steps": structpb.NewListValue(&structpb.ListValue{
                Values: []*structpb.Value{
                    structpb.NewStringValue("Restore latest backup to secondary site"),
                    structpb.NewStringValue("Apply transaction logs to recovery point"),
                    structpb.NewStringValue("Update DNS to point to secondary"),
                    structpb.NewStringValue("Verify application connectivity"),
                    structpb.NewStringValue("Run smoke tests"),
                },
            }),
        },
    },

    PostRecoverySteps: &structpb.Struct{
        Fields: map[string]*structpb.Value{
            "steps": structpb.NewListValue(&structpb.ListValue{
                Values: []*structpb.Value{
                    structpb.NewStringValue("Monitor system performance"),
                    structpb.NewStringValue("Notify stakeholders of successful failover"),
                    structpb.NewStringValue("Document recovery process"),
                    structpb.NewStringValue("Schedule post-mortem review"),
                },
            }),
        },
    },

    TestSchedule:     "0 3 1 * *", // Test monthly at 3 AM
    NotificationList: []string{"dba-team@example.com", "ops-team@example.com"},
    IsActive:         true,
}

resp, err := client.CreateDRPlan(ctx, &backupdr.CreateDRPlanRequest{
    Plan: drPlan,
})
```

### Monitoring System Health

```go
// Get health status for all systems
healthResp, err := client.ListSystemHealth(ctx, &backupdr.ListSystemHealthRequest{
    PageSize:     50,
    StatusFilter: backupdr.HealthStatus_HEALTH_STATUS_WARNING,
})

for _, health := range healthResp.HealthRecords {
    log.Printf("System: %s, Status: %v, CPU: %.2f%%, Memory: %.2f%%, Disk: %.2f%%",
        health.SystemName,
        health.HealthStatus,
        health.CpuUsage,
        health.MemoryUsage,
        health.DiskUsage)

    if health.HealthStatus != backupdr.HealthStatus_HEALTH_STATUS_HEALTHY {
        log.Printf("  Last backup: %v, Status: %s",
            health.LastBackupTime, health.BackupStatus)
    }
}
```

### Getting Backup Metrics

```go
startDate := time.Now().AddDate(0, -1, 0) // Last month
endDate := time.Now()

metricsResp, err := client.GetBackupMetrics(ctx, &backupdr.GetBackupMetricsRequest{
    StartDate:  timestamppb.New(startDate),
    EndDate:    timestamppb.New(endDate),
    TargetType: "DATABASE",
})

if err != nil {
    log.Fatalf("Failed to get metrics: %v", err)
}

metrics := metricsResp
log.Printf("Backup Metrics (Last Month):")
log.Printf("  Total Policies: %d (Active: %d)", metrics.TotalPolicies, metrics.ActivePolicies)
log.Printf("  Successful Backups: %d", metrics.SuccessfulBackups)
log.Printf("  Failed Backups: %d", metrics.FailedBackups)
log.Printf("  Total Backup Size: %.2f GB", float64(metrics.TotalBackupSize)/(1024*1024*1024))
log.Printf("  Average Duration: %v", metrics.AvgBackupDuration)

for _, daily := range metrics.DailyMetrics {
    log.Printf("  %v: %d successful, %d failed, %.2f GB",
        daily.Date, daily.SuccessfulJobs, daily.FailedJobs,
        float64(daily.TotalSize)/(1024*1024*1024))
}
```

## Integration

### With Scheduler Module

Backup policies are automatically scheduled through the scheduler module:

```go
// Scheduler creates cron jobs based on backup policy schedule
func (s *BackupOrchestrator) RegisterPolicyWithScheduler(policyID string, cronExpr string) error {
    return s.schedulerClient.CreateJob(ctx, &scheduler.CreateJobRequest{
        Name:           fmt.Sprintf("backup-policy-%s", policyID),
        CronExpression: cronExpr,
        TargetService:  "backupdr.BackupDRService",
        TargetMethod:   "ExecuteBackup",
        TargetPayload: &structpb.Struct{
            Fields: map[string]*structpb.Value{
                "policy_id": structpb.NewStringValue(policyID),
            },
        },
    })
}
```

### With Notification Module

Send backup completion notifications:

```go
// Notify on backup completion
func (s *BackupOrchestrator) notifyBackupComplete(job *models.BackupJob) error {
    if job.Status == models.JobStatusCompleted && !job.Policy.NotifyOnSuccess {
        return nil
    }
    if job.Status == models.JobStatusFailed && !job.Policy.NotifyOnFailure {
        return nil
    }

    return s.notificationClient.SendNotification(ctx, &notification.SendNotificationRequest{
        Notification: &notification.Notification{
            Type:          notification.NotificationType_SYSTEM_ALERT,
            Channel:       notification.NotificationChannel_EMAIL,
            Priority:      notification.NotificationPriority_NORMAL,
            RecipientIds:  []string{"ops-team"},
            Subject:       fmt.Sprintf("Backup %s: %s", job.Status, job.Policy.Name),
            Message:       formatBackupNotification(job),
        },
    })
}
```

### With Storage Providers

Multi-cloud storage support:

```go
// Select storage backend based on configuration
func (s *BackupOrchestrator) getStorageBackend(config *models.StorageConfig) (storage.Backend, error) {
    switch config.StorageType {
    case models.StorageTypeS3:
        return storage.NewS3Backend(config.S3Config)
    case models.StorageTypeAzure:
        return storage.NewAzureBackend(config.AzureConfig)
    case models.StorageTypeGCP:
        return storage.NewGCPBackend(config.GCPConfig)
    case models.StorageTypeLocal:
        return storage.NewLocalBackend(config.LocalPath)
    default:
        return nil, fmt.Errorf("unsupported storage type: %s", config.StorageType)
    }
}
```

## Development

### Running Tests

```bash
# Unit tests
go test ./backupdr/...

# Integration tests
go test -tags=integration ./backupdr/...

# With coverage
go test -cover -coverprofile=coverage.out ./backupdr/...
go tool cover -html=coverage.out
```

### Code Generation

```bash
# Generate protobuf code
cd backupdr
buf generate

# Generate SQLC queries
cd backupdr/db
sqlc generate

# Generate mocks for testing
mockgen -source=repository/backup_repository.go -destination=repository/mocks/backup_repository_mock.go
```

### Database Migrations

```bash
# Create a new migration
migrate create -ext sql -dir backupdr/db/migrations -seq add_backup_tags

# Run migrations
migrate -path backupdr/db/migrations -database "postgresql://user:pass@localhost:5432/ugcl?sslmode=disable" up

# Rollback
migrate -path backupdr/db/migrations -database "postgresql://user:pass@localhost:5432/ugcl?sslmode=disable" down 1
```

## Troubleshooting

### Common Issues

#### 1. Backup Job Stuck in Running State

**Symptoms**: Backup job status remains `RUNNING` for extended period

**Solution**:
```sql
-- Check for long-running jobs
SELECT id, policy_id, status, started_at,
       NOW() - started_at as duration
FROM backup_jobs
WHERE status = 'RUNNING'
  AND started_at < NOW() - INTERVAL '24 hours';

-- Manually update stale jobs
UPDATE backup_jobs
SET status = 'FAILED',
    error_message = 'Job timeout - manually marked as failed',
    completed_at = NOW()
WHERE id = 'job-id-here';
```

#### 2. Storage Connection Failures

**Symptoms**: Backups fail with storage connection errors

**Diagnostic**:
```go
// Test storage connection
resp, err := client.TestStorageConnection(ctx, &backupdr.TestStorageConnectionRequest{
    LocationId: "storage-location-id",
})

if !resp.IsHealthy {
    log.Printf("Storage unhealthy: %s (latency: %.2fms)",
        resp.ErrorMessage, resp.LatencyMs)
}
```

**Solutions**:
- Verify credentials are correct and not expired
- Check network connectivity to storage endpoint
- Verify bucket/container exists and permissions are correct
- Check firewall rules allow outbound connections

#### 3. High Backup Failure Rate

**Diagnosis**:
```sql
-- Get failure reasons
SELECT error_message, COUNT(*) as count
FROM backup_jobs
WHERE status = 'FAILED'
  AND started_at > NOW() - INTERVAL '7 days'
GROUP BY error_message
ORDER BY count DESC;
```

**Common Causes**:
- Insufficient disk space on backup storage
- Database locks preventing consistent backup
- Network timeouts during large transfers
- Permission issues accessing source data

**Solutions**:
- Increase backup timeout values
- Schedule backups during low-activity periods
- Enable compression to reduce storage requirements
- Review and adjust retry policies

#### 4. Slow Backup Performance

**Optimization Strategies**:
- Use incremental backups instead of full backups
- Enable compression (ZSTD provides best compression/speed ratio)
- Increase `max_parallel_jobs` for parallel processing
- Use faster storage tier for frequently accessed backups
- Adjust `bandwidth_limit` if network is constrained

```go
// Optimize backup policy
policy.Compression = backupdr.CompressionType_COMPRESSION_TYPE_ZSTD
policy.MaxParallelJobs = 4
policy.BandwidthLimit = 0 // Unlimited
```

#### 5. Recovery Plan Test Failures

**Troubleshooting**:
```go
// Execute test with detailed logging
testResp, err := client.TestDRPlan(ctx, &backupdr.TestDRPlanRequest{
    PlanId:     "dr-plan-id",
    TestType:   backupdr.ExecutionType_EXECUTION_TYPE_TEST,
    Reason:     "Monthly DR drill",
})

// Check execution log
execResp, err := client.GetRecoveryExecution(ctx, &backupdr.GetRecoveryExecutionRequest{
    Id: testResp.Execution.Id,
})

log.Printf("Steps completed: %d/%d",
    execResp.Execution.CompletedSteps,
    execResp.Execution.TotalSteps)
log.Printf("Current step: %s", execResp.Execution.CurrentStep)
log.Printf("Issues: %v", execResp.Execution.IssuesEncountered)
```

### Performance Tuning

```bash
# PostgreSQL configuration for backup workloads
max_connections = 200
shared_buffers = 4GB
effective_cache_size = 12GB
maintenance_work_mem = 1GB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
random_page_cost = 1.1
effective_io_concurrency = 200
work_mem = 20MB
min_wal_size = 1GB
max_wal_size = 4GB
```

### Monitoring Queries

```sql
-- Active backup jobs with progress
SELECT
    bj.id,
    bp.name as policy_name,
    bj.status,
    bj.progress_percent,
    bj.processed_size,
    bj.total_size,
    bj.throughput_mbps,
    NOW() - bj.started_at as duration
FROM backup_jobs bj
JOIN backup_policies bp ON bp.id = bj.policy_id
WHERE bj.status IN ('RUNNING', 'QUEUED')
ORDER BY bj.started_at DESC;

-- Storage capacity by location
SELECT
    name,
    storage_type,
    total_capacity,
    used_capacity,
    available_capacity,
    ROUND(100.0 * used_capacity / NULLIF(total_capacity, 0), 2) as usage_percent
FROM backup_storage_locations
WHERE is_active = true
ORDER BY usage_percent DESC;

-- Recent failures
SELECT
    bp.name,
    bj.error_message,
    bj.started_at,
    bj.retry_count
FROM backup_jobs bj
JOIN backup_policies bp ON bp.id = bj.policy_id
WHERE bj.status = 'FAILED'
  AND bj.started_at > NOW() - INTERVAL '24 hours'
ORDER BY bj.started_at DESC;
```
