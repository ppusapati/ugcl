# Data Bridge & Integration Module

## Module Overview

The Data Bridge module provides comprehensive data import, mapping, and integration capabilities for the UGCL system. It enables users to import data from CSV files, map fields to database columns, validate data, and execute bulk imports with real-time progress tracking.

### Key Features

- **CSV Import**: Support for CSV file imports with automatic encoding detection
- **Field Mapping**: Visual field mapping between CSV columns and database columns
- **Data Validation**: Comprehensive validation rules with customizable error messages
- **Data Transformation**: Built-in transformations (trim, case conversion, date parsing)
- **Preview & Analysis**: Data analysis and import preview before execution
- **Mapping Templates**: Reusable import mappings for repeated imports
- **Progress Streaming**: Real-time progress updates during import execution
- **Error Handling**: Detailed error reporting with row-level error tracking
- **Batch Processing**: Efficient batch processing for large datasets

## Architecture

### Module Structure

```
databridge/
├── api/databridge/           # Generated gRPC code
├── db/
│   ├── generated/            # SQLC generated queries
│   ├── queries/              # SQL queries
│   └── schema/               # Database schema
├── handlers/                 # gRPC handlers
├── models/                   # Domain models
├── proto/                    # Protobuf definitions
├── repository/               # Data access layer
├── services/                 # Business logic
└── module.go                 # Dependency injection
```

### Database Schema

#### Core Tables

1. **import_mappings** - Reusable CSV to database field mappings
   - Mapping name and description
   - Target table configuration
   - CSV headers and field mappings
   - Transformation and validation rules
   - Usage tracking

2. **import_jobs** - Import job execution tracking
   - Job status and progress
   - File metadata (name, size, hash)
   - Success/failure counts
   - Error details
   - Execution timing

### Dependencies

- **Masters Module**: Table and column metadata
- **Notification**: Import completion notifications
- **Storage**: File upload and storage

## Quick Start

### Analyzing CSV Data

```go
import databridge "p9e.in/ugcl/databridge/api/databridge"

client := databridge.NewDataBridgeServiceClient(conn)

// Read CSV file
csvData, _ := ioutil.ReadFile("employees.csv")

// Analyze the CSV structure
analysisResp, err := client.AnalyzeData(ctx, &databridge.AnalyzeDataRequest{
    CsvData:   csvData,
    FileName:  "employees.csv",
    Delimiter: ",",
    HasHeader: true,
})

if err != nil {
    log.Fatalf("Analysis failed: %v", err)
}

analysis := analysisResp.Analysis
log.Printf("Detected %d columns, %d data rows", len(analysis.Headers), analysis.DataRows)
log.Printf("Encoding: %s, Delimiter: %s", analysis.Encoding, analysis.Delimiter)

for _, colAnalysis := range analysis.ColumnAnalysis {
    log.Printf("Column: %s", colAnalysis.Header)
    log.Printf("  Inferred type: %v", colAnalysis.InferredType)
    log.Printf("  Non-empty: %d, Empty: %d", colAnalysis.NonEmptyCount, colAnalysis.EmptyCount)
    log.Printf("  Max length: %d", colAnalysis.MaxLength)
    log.Printf("  Sample values: %v", colAnalysis.SampleValues)
}
```

### Creating an Import Mapping

```go
// Get available tables and columns
tablesResp, _ := client.GetTables(ctx, &databridge.GetTablesRequest{
    ImportableOnly: true,
})

table := tablesResp.Tables[0] // Select target table

columnsResp, _ := client.GetColumns(ctx, &databridge.GetColumnsRequest{
    TableId:        table.Id,
    ImportableOnly: true,
})

// Create field mappings
fieldMappings := []*databridge.FieldMapping{
    {
        CsvField:       "Employee ID",
        DbColumn:       "employee_id",
        Transformation: "trim",
        Validations: []*databridge.ValidationRule{
            {Type: "required", Message: "Employee ID is required"},
            {Type: "pattern", Value: "^EMP[0-9]{6}$", Message: "Invalid employee ID format"},
        },
    },
    {
        CsvField:       "Full Name",
        DbColumn:       "full_name",
        Transformation: "trim",
        Validations: []*databridge.ValidationRule{
            {Type: "required", Message: "Name is required"},
            {Type: "min_length", Value: "2", Message: "Name too short"},
        },
    },
    {
        CsvField:       "Email",
        DbColumn:       "email",
        Transformation: "trim,lowercase",
        Validations: []*databridge.ValidationRule{
            {Type: "required", Message: "Email is required"},
            {Type: "email", Message: "Invalid email format"},
        },
    },
    {
        CsvField:       "Join Date",
        DbColumn:       "joined_at",
        Transformation: "date",
        Validations: []*databridge.ValidationRule{
            {Type: "date", Value: "2006-01-02", Message: "Invalid date format"},
        },
    },
}

// Create mapping
mappingResp, err := client.CreateMapping(ctx, &databridge.CreateMappingRequest{
    MappingName:  "Employee Import Mapping",
    Description:  "Standard employee data import from HR system",
    TableId:      table.Id,
    CsvHeaders:   []string{"Employee ID", "Full Name", "Email", "Join Date"},
    FieldMappings: fieldMappings,
    CreatedBy:    "user-123",
})
```

### Previewing Import

```go
// Preview import with validation
previewResp, err := client.PreviewImport(ctx, &databridge.PreviewImportRequest{
    MappingId:    mappingResp.Mapping.Id,
    CsvData:      csvData,
    FileName:     "employees.csv",
    PreviewRows:  10, // Preview first 10 rows
})

if err != nil {
    log.Fatalf("Preview failed: %v", err)
}

log.Printf("Preview: %d valid rows, %d errors, %d warnings",
    len(previewResp.PreviewRows), len(previewResp.Errors), len(previewResp.Warnings))

// Display preview rows
for _, row := range previewResp.PreviewRows {
    log.Printf("Row %d: %v", row.RowNumber, row.TransformedData)
    if !row.IsValid {
        log.Printf("  Errors: %v", row.RowErrors)
    }
}

// Display errors
for _, err := range previewResp.Errors {
    log.Printf("Error at row %d, field %s: %s",
        err.RowNumber, err.CsvField, err.ErrorMessage)
}
```

### Executing Import

```go
// Create import job
jobResp, err := client.CreateImportJob(ctx, &databridge.CreateImportJobRequest{
    JobName:    "Employee Import - March 2024",
    MappingId:  mappingResp.Mapping.Id,
    FileName:   "employees.csv",
    FileSize:   int64(len(csvData)),
    FileHash:   calculateSHA256(csvData),
    TotalRows:  analysisResp.Analysis.DataRows,
    CreatedBy:  "user-123",
})

// Execute import with streaming progress
stream, err := client.ExecuteImport(ctx, &databridge.ExecuteImportRequest{
    JobId:   jobResp.Job.Id,
    CsvData: csvData,
})

if err != nil {
    log.Fatalf("Import failed: %v", err)
}

// Monitor progress
for {
    resp, err := stream.Recv()
    if err == io.EOF {
        break
    }
    if err != nil {
        log.Fatalf("Stream error: %v", err)
    }

    switch msg := resp.Response.(type) {
    case *databridge.ExecuteImportResponse_Progress:
        progress := msg.Progress
        log.Printf("Progress: %.2f%% (%d/%d rows) - Phase: %s",
            progress.ProgressPercentage,
            progress.ProcessedRows,
            progress.TotalRows,
            progress.CurrentPhase)

        if len(progress.RecentErrors) > 0 {
            log.Printf("  Recent errors: %v", progress.RecentErrors)
        }

    case *databridge.ExecuteImportResponse_Completed:
        completed := msg.Completed
        log.Printf("Import completed!")
        log.Printf("  Total rows: %d", completed.TotalRows)
        log.Printf("  Successful: %d", completed.SuccessfulRows)
        log.Printf("  Failed: %d", completed.FailedRows)
        log.Printf("  Completed at: %v", completed.CompletedAt)

        if len(completed.ErrorSummary) > 0 {
            log.Printf("  Errors summary:")
            for _, err := range completed.ErrorSummary {
                log.Printf("    Row %d: %s", err.RowNumber, err.ErrorMessage)
            }
        }

    case *databridge.ExecuteImportResponse_Failed:
        failed := msg.Failed
        log.Printf("Import failed: %s (%s)", failed.ErrorMessage, failed.ErrorType)
        log.Printf("  Failed at: %v", failed.FailedAt)
    }
}
```

## API Reference

### DataBridgeService RPCs

#### Schema and Table Management
- **GetSchemas** - List all available database schemas
- **GetTables** - List tables in a schema (with import support filter)
- **GetColumns** - Get columns for a table with metadata

#### Import Mapping Management
- **GetMappings** - List all import mappings with filters
- **GetMappingByID** - Retrieve specific mapping
- **CreateMapping** - Create new import mapping
- **UpdateMapping** - Update existing mapping
- **DeleteMapping** - Delete mapping

#### Data Processing
- **AnalyzeData** - Analyze CSV file structure and content
- **ValidateMapping** - Validate mapping against CSV data
- **PreviewImport** - Preview import results with transformations

#### Import Job Management
- **CreateImportJob** - Create new import job
- **GetImportJob** - Get job details and status
- **GetImportJobs** - List jobs with filtering
- **CancelImportJob** - Cancel running import

#### Import Execution
- **ExecuteImport** - Execute import with streaming progress

## Database Schema

### Key Indexes

```sql
-- Mapping lookups
CREATE INDEX idx_import_mappings_table_id ON import_mappings(table_id);
CREATE INDEX idx_import_mappings_created_by ON import_mappings(created_by);
CREATE INDEX idx_import_mappings_last_used ON import_mappings(last_used_at DESC);

-- Job tracking
CREATE INDEX idx_import_jobs_mapping_id ON import_jobs(mapping_id);
CREATE INDEX idx_import_jobs_status ON import_jobs(status);
CREATE INDEX idx_import_jobs_created_at ON import_jobs(created_at DESC);
```

## Configuration

### Environment Variables

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=ugcl
DB_USER=ugcl_user
DB_PASSWORD=secure_password

# Import Configuration
MAX_FILE_SIZE=100MB
MAX_ROWS_PER_IMPORT=1000000
BATCH_SIZE=1000
TEMP_UPLOAD_DIR=/tmp/imports

# Validation
ENABLE_STRICT_VALIDATION=true
MAX_ERRORS_BEFORE_ABORT=100

# Performance
PARALLEL_IMPORTS=2
IMPORT_TIMEOUT=3600  # 1 hour
```

## Examples

### Custom Validation Rules

```go
validations := []*databridge.ValidationRule{
    {Type: "required", Message: "Field is required"},
    {Type: "email", Message: "Invalid email"},
    {Type: "min_length", Value: "5", Message: "Too short"},
    {Type: "max_length", Value: "100", Message: "Too long"},
    {Type: "pattern", Value: "^[A-Z]{2}[0-9]{6}$", Message: "Invalid format"},
    {Type: "min", Value: "0", Message: "Must be positive"},
    {Type: "max", Value: "100", Message: "Value too large"},
}
```

### Available Transformations

```go
transformations := []string{
    "trim",           // Remove whitespace
    "uppercase",      // Convert to uppercase
    "lowercase",      // Convert to lowercase
    "date",           // Parse date
    "decimal",        // Parse decimal
    "integer",        // Parse integer
    "boolean",        // Parse boolean
}
```

## Integration

### With Masters Module

```go
// Get table metadata from Masters
func (s *DataBridgeService) GetTableMetadata(tableID string) (*masters.Table, error) {
    return s.mastersClient.GetTableByID(ctx, &masters.GetTableByIDRequest{
        TableId: tableID,
    })
}
```

### With Notification Module

```go
// Notify on import completion
func (s *DataBridgeService) notifyImportComplete(job *models.ImportJob) error {
    return s.notificationClient.SendNotification(ctx, &notification.SendNotificationRequest{
        Notification: &notification.Notification{
            Type:        notification.NotificationType_SYSTEM_ALERT,
            Channel:     notification.NotificationChannel_EMAIL,
            RecipientId: job.CreatedBy,
            Subject:     fmt.Sprintf("Import Job Completed: %s", job.JobName),
            Message:     formatImportSummary(job),
        },
    })
}
```

## Development

### Running Tests

```bash
go test ./databridge/...
go test -tags=integration ./databridge/...
```

### Code Generation

```bash
cd databridge
buf generate

cd db
sqlc generate
```

## Troubleshooting

### Common Issues

#### 1. Import Fails with "Invalid CSV Format"

**Solution**: Check encoding and delimiter
```go
// Try different delimiters
for _, delimiter := range []string{",", ";", "\t", "|"} {
    analysis, err := client.AnalyzeData(ctx, &databridge.AnalyzeDataRequest{
        CsvData:   csvData,
        Delimiter: delimiter,
    })
    if err == nil && len(analysis.Errors) == 0 {
        log.Printf("Detected delimiter: %s", delimiter)
        break
    }
}
```

#### 2. Validation Errors for Dates

**Solution**: Specify correct date format
```go
{
    CsvField:       "Date",
    DbColumn:       "created_at",
    Transformation: "date",
    Validations: []*databridge.ValidationRule{
        {Type: "date", Value: "01/02/2006", Message: "Use MM/DD/YYYY format"},
    },
}
```

#### 3. Slow Import Performance

**Optimization**:
- Increase batch size for larger datasets
- Disable unnecessary validations
- Use bulk insert instead of row-by-row
- Import during off-peak hours

#### 4. Duplicate Key Errors

**Solution**: Add duplicate handling strategy
```sql
-- Use UPSERT for imports
INSERT INTO employees (id, name, email)
VALUES ($1, $2, $3)
ON CONFLICT (id) DO UPDATE
SET name = EXCLUDED.name,
    email = EXCLUDED.email;
```

### Monitoring Queries

```sql
-- Active imports
SELECT
    ij.id,
    ij.job_name,
    ij.status,
    ij.processed_rows,
    ij.total_rows,
    ROUND(100.0 * ij.processed_rows / NULLIF(ij.total_rows, 0), 2) as progress_pct
FROM import_jobs ij
WHERE ij.status = 'processing'
ORDER BY ij.created_at DESC;

-- Import statistics
SELECT
    im.mapping_name,
    COUNT(ij.id) as total_imports,
    SUM(CASE WHEN ij.status = 'completed' THEN 1 ELSE 0 END) as successful,
    SUM(CASE WHEN ij.status = 'failed' THEN 1 ELSE 0 END) as failed,
    AVG(ij.successful_rows) as avg_rows
FROM import_mappings im
LEFT JOIN import_jobs ij ON ij.mapping_id = im.id
GROUP BY im.id, im.mapping_name;
```
