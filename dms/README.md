# Document Management System (DMS) Module

## 1. Module Overview

The Document Management System (DMS) provides comprehensive document lifecycle management with advanced features including document processing, watermarking, OCR, virus scanning, versioning, and secure sharing capabilities.

**Purpose:** Centralized document storage, processing, and access control with entity-based ownership, organizational context, and advanced security features for the UGCL platform.

**Key Features:**
- Multi-format document upload and storage
- Automatic document processing (compression, thumbnails, OCR)
- Dynamic watermarking with user context
- Document versioning and history
- Secure document sharing with permissions
- Virus scanning and security checks
- Full-text search and metadata extraction
- Document analytics and access logging
- Storage usage tracking
- Entity-based ownership model
- Expiration management

## 2. Architecture

### Module Structure
```
dms/
├── proto/
│   └── dms.proto                # Service and message definitions
├── db/
│   ├── schema/
│   │   └── schema.sql          # Database schema with enums
│   └── generated/              # SQLC generated code
├── repository/                 # Data access layer
│   ├── document_repository.go
│   ├── share_repository.go
│   └── watermark_repository.go
├── services/                   # Business logic layer
│   ├── document_service.go
│   ├── processing_service.go
│   ├── watermark_service.go
│   └── sharing_service.go
├── handlers/                   # gRPC handlers
│   └── dms_handler.go
├── mappers/                    # DTO mappers
│   └── dms_mappers.go
├── utils/                      # Utility functions
│   ├── compression/
│   │   └── tar_zstd.go        # TAR+Zstandard compression
│   ├── processing/
│   │   ├── pdf_processor.go
│   │   └── ocr_processor.go
│   └── watermark/
│       └── generator.go
├── events/                     # Event definitions
│   └── document_events.go
├── models/                     # Domain models
│   ├── document.go
│   └── watermark.go
├── config/                     # Configuration
│   └── storage_config.go
├── module.go                   # FX module definition
└── module_sqlc.go             # SQLC provider
```

### Database Schema

**Tables:**
- `documents` - Main document storage with metadata
- `document_shares` - Document sharing and permissions
- `watermark_configs` (planned) - Watermark templates
- `processing_jobs` (planned) - Async processing queue
- `document_access_logs` (planned) - Access audit trail

**Enums:**
- `document_type` - PDF, IMAGE, VIDEO, AUDIO, etc.
- `document_category` - CONTRACT, INVOICE, REPORT, etc.
- `processing_status` - PENDING, PROCESSING, COMPLETED, FAILED
- `ocr_status` - PENDING, PROCESSING, COMPLETED, FAILED, SKIPPED
- `virus_scan_status` - PENDING, SCANNING, CLEAN, INFECTED, FAILED
- `share_permission` - VIEW, DOWNLOAD, EDIT, DELETE

**Key Relationships:**
- Documents → Shares (1:N, CASCADE delete)
- Documents → Entities (N:1 via owner_entity_id)
- Documents → Organization units (N:1 via division_id, branch_id, department_id)

### Key Dependencies
- `packages/database/sqlc` - Database connection
- `identity/entity` - Entity ownership model
- `organization` - Organizational context
- Storage backend (filesystem, S3, etc.)
- OCR engine (Tesseract)
- Compression (Zstandard)
- Watermarking library

## 3. Quick Start

### Upload a Document

```go
import (
    "context"
    dmsv1 "p9e.in/ugcl/dms/api/v1"
    "connectrpc.com/connect"
)

client := dmsv1.NewDMSServiceClient(httpClient, baseURL)

// Upload with auto-processing
resp, err := client.UploadDocument(ctx, connect.NewRequest(&dmsv1.UploadDocumentRequest{
    TenantId:         "tenant-uuid",
    OwnerEntityType:  "EMPLOYEE",
    OwnerEntityId:    "entity-uuid",
    DivisionId:       "division-uuid",
    BranchId:         "branch-uuid",
    DepartmentId:     "dept-uuid",
    DocumentType:     dmsv1.DocumentType_DOCUMENT_TYPE_PDF,
    DocumentCategory: dmsv1.DocumentCategory_DOCUMENT_CATEGORY_CONTRACT,
    FileName:         "contract.pdf",
    Title:            "Vendor Contract 2024",
    Description:      "Annual vendor service contract",
    Content:          fileBytes,
    MimeType:         "application/pdf",
    Tags:             []string{"contract", "vendor", "2024"},
    AutoProcess:      true,
    WatermarkConfigId: "watermark-uuid",
    ExpiresAt:        timestamppb.New(expiryDate),
}))

document := resp.Msg.Document
processingJob := resp.Msg.ProcessingJob
```

### Common Operations

**Retrieve document with preview:**
```go
// Get document metadata
doc, err := client.GetDocument(ctx, connect.NewRequest(&dmsv1.GetDocumentRequest{
    Id:              "doc-uuid",
    TenantId:        "tenant-uuid",
    IncludeContent:  false,
    IncludeVersions: true,
}))

// Get watermarked preview
preview, err := client.GetDocumentPreview(ctx, connect.NewRequest(&dmsv1.GetDocumentPreviewRequest{
    DocumentId:       "doc-uuid",
    WatermarkConfigId: "watermark-uuid",
    UserContext:      "user-name",
    WatermarkData: &structpb.Struct{
        Fields: map[string]*structpb.Value{
            "username": structpb.NewStringValue("John Doe"),
            "timestamp": structpb.NewStringValue(time.Now().Format(time.RFC3339)),
        },
    },
}))
```

**Share a document:**
```go
share, err := client.CreateDocumentShare(ctx, connect.NewRequest(&dmsv1.CreateDocumentShareRequest{
    TenantId:          "tenant-uuid",
    DocumentId:        "doc-uuid",
    SharedWithUserId:  "recipient-user-uuid",
    CanView:           true,
    CanDownload:       true,
    CanEdit:           false,
    CanDelete:         false,
    ExpiresAt:         timestamppb.New(time.Now().Add(7*24*time.Hour)),
}))
```

## 4. API Reference

### Document CRUD Operations
- `UploadDocument` - Upload new document with auto-processing
- `GetDocument` - Retrieve document by ID with optional content/versions
- `GetDocuments` - List/search documents with filters
- `UpdateDocument` - Update document metadata
- `DeleteDocument` - Soft or hard delete document
- `GetDocumentVersions` - Retrieve version history

### Document Processing
- `ProcessDocument` - Queue processing jobs (compress, thumbnail, OCR)
- `GetProcessingStatus` - Check processing job status
- `GenerateThumbnail` - Generate thumbnail for specific page
- `ExtractText` - Extract text with optional OCR

### Document Viewing
- `GetDocumentPreview` - Get preview with watermark
- `GetDocumentPage` - Render specific page as image
- `StreamDocument` - Stream document with chunking

### Watermark Operations
- `CreateWatermarkConfig` - Create watermark template
- `GetWatermarkConfigs` - List watermark configurations
- `UpdateWatermarkConfig` - Update watermark template
- `DeleteWatermarkConfig` - Remove watermark configuration
- `ApplyWatermark` - Apply watermark to document

### Sharing Operations
- `CreateDocumentShare` - Share document with user/entity
- `GetDocumentShare` - Get share details
- `ListDocumentShares` - List shares for document
- `DeleteDocumentShare` - Revoke share access

### Analytics Operations
- `GetDocumentAccess` - Retrieve access logs
- `GetDocumentAnalytics` - Get usage analytics
- `GetStorageUsage` - Track storage consumption

## 5. Database Schema

### Tables Overview

**documents**
```sql
- id (UUID, PK)
- tenant_id (UUID)
- owner_entity_type (VARCHAR(50))
- owner_entity_id (UUID)
- division_id, branch_id, department_id (UUID, organizational context)
- document_type (ENUM), document_category (ENUM)
- file_name, original_name, title, description
- mime_type, file_extension, size_bytes, compressed_size, checksum
- storage_path, compressed_path, thumbnail_path, preview_path
- page_count, word_count, extracted_text, language
- processing_status (ENUM), processing_error, processed_at
- compression_type, compression_ratio
- ocr_status (ENUM), ocr_confidence, ocr_processed_at
- virus_scan_status (ENUM), virus_scan_result, virus_scanned_at
- watermark_config (JSONB), has_watermark
- metadata (JSONB), tags (TEXT[])
- expires_at, is_expired (GENERATED)
- uploaded_by, permissions (JSONB)
- is_deleted, deleted_at, deleted_by (soft delete)
- created_at, updated_at, created_by, updated_by
```

**document_shares**
```sql
- id (UUID, PK)
- tenant_id (UUID)
- document_id (UUID, FK → documents)
- shared_with_entity_id, shared_with_user_id, shared_with_email
- division_id, branch_id, department_id (organizational scope)
- can_view, can_download, can_edit, can_delete (permissions)
- share_link, share_password
- expires_at, is_expired (GENERATED)
- access_count, last_accessed_at
- created_at, updated_at, created_by, updated_by
```

### Key Indexes
- Tenant-based indexes for multi-tenancy
- Entity ownership indexes
- Organizational context indexes
- Full-text search on file_name
- GIN indexes on tags and metadata JSONB
- Expiration date indexes for cleanup
- Checksum uniqueness per tenant

## 6. Configuration

### Environment Variables
```env
# Storage Configuration
STORAGE_BACKEND=filesystem  # or s3, gcs
STORAGE_PATH=/var/ugcl/documents
STORAGE_MAX_SIZE=104857600  # 100MB max file size

# Processing Configuration
ENABLE_AUTO_COMPRESSION=true
COMPRESSION_LEVEL=3  # Zstandard level
ENABLE_AUTO_OCR=true
OCR_LANGUAGE=eng+hin
ENABLE_VIRUS_SCAN=true

# Watermarking
DEFAULT_WATERMARK_CONFIG=watermark-uuid
WATERMARK_FONT_PATH=/fonts/arial.ttf

# Thumbnails
THUMBNAIL_WIDTH=200
THUMBNAIL_HEIGHT=200
THUMBNAIL_FORMAT=png
```

### Module-Specific Settings
- SQLC configuration in `db/sqlc.yaml`
- Storage backend adapter configuration
- OCR engine settings
- Compression parameters

## 7. Examples

### Complete Document Workflow

```go
// 1. Upload document
uploadResp, err := client.UploadDocument(ctx, connect.NewRequest(&dmsv1.UploadDocumentRequest{
    TenantId:         tenantID,
    OwnerEntityType:  "EMPLOYEE",
    OwnerEntityId:    entityID,
    DocumentType:     dmsv1.DocumentType_DOCUMENT_TYPE_PDF,
    DocumentCategory: dmsv1.DocumentCategory_DOCUMENT_CATEGORY_REPORT,
    FileName:         "monthly-report.pdf",
    Content:          pdfBytes,
    MimeType:         "application/pdf",
    AutoProcess:      true,
}))

docID := uploadResp.Msg.Document.Id

// 2. Monitor processing
for {
    status, err := client.GetProcessingStatus(ctx, connect.NewRequest(&dmsv1.GetProcessingStatusRequest{
        DocumentId: docID,
    }))

    allCompleted := true
    for _, job := range status.Msg.Jobs {
        if job.Status != "COMPLETED" {
            allCompleted = false
            break
        }
    }

    if allCompleted {
        break
    }
    time.Sleep(1 * time.Second)
}

// 3. Share with team
share, err := client.CreateDocumentShare(ctx, connect.NewRequest(&dmsv1.CreateDocumentShareRequest{
    TenantId:         tenantID,
    DocumentId:       docID,
    DepartmentId:     deptID,  // Share with entire department
    CanView:          true,
    CanDownload:      true,
    ExpiresAt:        timestamppb.New(time.Now().Add(30*24*time.Hour)),
}))

// 4. Get analytics
analytics, err := client.GetDocumentAnalytics(ctx, connect.NewRequest(&dmsv1.GetDocumentAnalyticsRequest{
    DocumentId:  docID,
    StartTime:   timestamppb.New(time.Now().Add(-7*24*time.Hour)),
    EndTime:     timestamppb.Now(),
    Granularity: "day",
}))

fmt.Printf("Total views: %d, Downloads: %d\n",
    analytics.Msg.Analytics.TotalViews,
    analytics.Msg.Analytics.TotalDownloads)
```

### Dynamic Watermarking

```go
// Create watermark config
config, err := client.CreateWatermarkConfig(ctx, connect.NewRequest(&dmsv1.CreateWatermarkConfigRequest{
    Config: &dmsv1.WatermarkConfig{
        Name:             "Confidential",
        Type:             "text",
        TextContent:      "CONFIDENTIAL",
        Position:         "center",
        FontSize:         48,
        FontColor:        "#FF0000",
        Opacity:          0.3,
        Rotation:         -45,
        UseDynamicText:   true,
        ShowUsername:     true,
        ShowTimestamp:    true,
        ShowIpAddress:    true,
        ApplyToAllPages:  true,
        TextTemplate:     "{{.Username}} - {{.Timestamp}} - {{.IPAddress}}",
    },
}))

// Apply watermark to document
watermarked, err := client.ApplyWatermark(ctx, connect.NewRequest(&dmsv1.ApplyWatermarkRequest{
    DocumentId:       docID,
    WatermarkConfigId: config.Msg.Config.Id,
    DynamicData: &structpb.Struct{
        Fields: map[string]*structpb.Value{
            "Username":  structpb.NewStringValue("John Doe"),
            "Timestamp": structpb.NewStringValue(time.Now().Format("2006-01-02 15:04:05")),
            "IPAddress": structpb.NewStringValue("192.168.1.100"),
        },
    },
    OutputFormat: "pdf",
}))
```

## 8. Integration

### With Other Modules

**Identity/Entity Module:**
- Documents owned by entities (EMPLOYEE, CONTRACTOR, VENDOR)
- Entity-based access control
- User context for watermarking

**Organization Module:**
- Documents tagged with division/branch/department
- Organizational scope for sharing
- Department-wide document access

**Employee/Contractor/Vendor Modules:**
- Personal documents (contracts, certifications)
- Reference documents for entities
- Onboarding document management

**Approval Workflows:**
- Document approval processes
- Version tracking for approvals
- Audit trail integration

### Event Publishing
```go
// Document events
type DocumentEvent struct {
    Type       string    // uploaded, updated, deleted, accessed
    DocumentID string
    TenantID   string
    EntityID   string
    Timestamp  time.Time
    Metadata   map[string]interface{}
}
```

## 9. Development

### Modifying the Module

**Add New Document Type:**
1. Update `document_type` enum in `db/schema/schema.sql`
2. Update `DocumentType` enum in `proto/dms.proto`
3. Add processor in `utils/processing/`
4. Run migrations and regenerate code

**Add Processing Feature:**
1. Define job type in `models/processing_job.go`
2. Implement processor in `services/processing_service.go`
3. Update `ProcessDocument` RPC handler
4. Add tests for new processor

### Generate Code
```bash
# Generate proto stubs
buf generate

# Generate SQLC queries
sqlc generate -f dms/db/sqlc.yaml

# Run all generation
make generate
```

### Testing Guidelines

**Unit Tests:**
```go
func TestDocumentUpload(t *testing.T) {
    mockRepo := &MockDocumentRepository{}
    mockStorage := &MockStorageBackend{}
    svc := services.NewDocumentService(mockRepo, mockStorage)

    doc, err := svc.UploadDocument(ctx, req)
    assert.NoError(t, err)
    assert.Equal(t, "contract.pdf", doc.FileName)
}
```

**Integration Tests:**
- Test with real storage backend (MinIO in test mode)
- Test OCR with sample images
- Test compression ratios
- Test watermark generation

## 10. Troubleshooting

### Common Issues

**Issue: File upload fails with size limit exceeded**
- **Cause:** File exceeds STORAGE_MAX_SIZE
- **Solution:** Adjust config or implement chunked upload

**Issue: OCR not working**
- **Cause:** Tesseract not installed or wrong language pack
- **Solution:** Install tesseract-ocr and language data

**Issue: Watermark not appearing**
- **Cause:** Font path incorrect or unsupported format
- **Solution:** Verify WATERMARK_FONT_PATH and document format

**Issue: Compression fails**
- **Cause:** Insufficient disk space or corrupted file
- **Solution:** Check storage space and validate file integrity

**Issue: Document not found after upload**
- **Cause:** Async processing not complete
- **Solution:** Poll GetProcessingStatus before accessing

### Performance Tips

1. Use compressed_path for large documents
2. Cache thumbnails on CDN
3. Lazy-load OCR (on-demand only)
4. Use streaming for large file downloads
5. Implement document expiration cleanup job

### Debugging

**Check storage paths:**
```go
doc, _ := client.GetDocument(ctx, &dmsv1.GetDocumentRequest{Id: docID})
fmt.Printf("Storage: %s\n", doc.StoragePath)
fmt.Printf("Compressed: %s\n", doc.CompressedPath)
fmt.Printf("Thumbnail: %s\n", doc.ThumbnailPath)
```

**Verify processing status:**
```sql
SELECT id, processing_status, ocr_status, virus_scan_status, processing_error
FROM documents
WHERE id = 'doc-uuid';
```

**Check access logs:**
```sql
SELECT user_id, action, created_at, watermark_applied
FROM document_access_logs
WHERE document_id = 'doc-uuid'
ORDER BY created_at DESC
LIMIT 100;
```

### Security Considerations

1. Always scan files for viruses before serving
2. Apply watermarks for sensitive documents
3. Log all document access
4. Encrypt documents at rest
5. Validate file types (don't trust mime_type)
6. Implement rate limiting for downloads
7. Use signed URLs with expiration

## Additional Resources

- [Zstandard Compression](https://facebook.github.io/zstd/)
- [Tesseract OCR](https://github.com/tesseract-ocr/tesseract)
- [PDF Processing in Go](https://github.com/unidoc/unipdf)
- [Watermarking Best Practices](https://www.w3.org/TR/watermarking/)
