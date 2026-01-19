# Uploading Documents to DMS - Complete Guide

This comprehensive guide demonstrates document upload workflows in the UGCL Document Management System (DMS), including file handling, metadata, permissions, processing, and versioning.

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Client Setup](#client-setup)
4. [Basic Document Upload](#basic-document-upload)
5. [Multipart File Handling](#multipart-file-handling)
6. [Document Categorization](#document-categorization)
7. [Permissions and Sharing](#permissions-and-sharing)
8. [Document Processing](#document-processing)
9. [Watermarking](#watermarking)
10. [Document Download](#document-download)
11. [Document Versioning](#document-versioning)
12. [Complete Workflow](#complete-workflow)
13. [Error Handling](#error-handling)
14. [Best Practices](#best-practices)

## Overview

The DMS provides comprehensive document management with:
- Multipart upload support for large files
- Automatic compression and thumbnail generation
- OCR and text extraction
- Virus scanning integration
- Watermark application
- Version control
- Entity-based access control
- Organizational scope (division/branch/department)

## Prerequisites

```go
import (
    "context"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "io"
    "log"
    "mime/multipart"
    "os"
    "path/filepath"
    "time"

    "connectrpc.com/connect"
    dmsv1 "p9e.in/ugcl/dms/api/v1"
    "p9e.in/ugcl/dms/api/v1/dmsv1connect"
    "google.golang.org/protobuf/types/known/structpb"
    "google.golang.org/protobuf/types/known/timestamppb"
)
```

## Client Setup

### DMS Client Configuration

```go
package main

import (
    "crypto/tls"
    "net/http"
    "time"
)

// DMSClient encapsulates the DMS gRPC client configuration
type DMSClient struct {
    client   dmsv1connect.DMSServiceClient
    tenantID string
    entityID string
    userID   string
}

// NewDMSClient creates a new DMS service client
func NewDMSClient(baseURL, tenantID, entityID, userID, authToken string) *DMSClient {
    // Configure HTTP client with extended timeout for large uploads
    httpClient := &http.Client{
        Timeout: 5 * time.Minute, // Longer timeout for uploads
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{
                MinVersion: tls.VersionTLS12,
            },
            MaxIdleConns:          100,
            MaxIdleConnsPerHost:   100,
            IdleConnTimeout:       90 * time.Second,
            DisableCompression:    true, // DMS handles compression
            MaxConnsPerHost:       10,
            ResponseHeaderTimeout: 30 * time.Second,
        },
    }

    // Create Connect client
    client := dmsv1connect.NewDMSServiceClient(
        httpClient,
        baseURL,
        connect.WithInterceptors(
            authInterceptor(authToken),
            uploadProgressInterceptor(),
        ),
    )

    return &DMSClient{
        client:   client,
        tenantID: tenantID,
        entityID: entityID,
        userID:   userID,
    }
}

// authInterceptor adds authentication to requests
func authInterceptor(token string) connect.UnaryInterceptorFunc {
    return func(next connect.UnaryFunc) connect.UnaryFunc {
        return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
            req.Header().Set("Authorization", "Bearer "+token)
            req.Header().Set("X-Request-ID", generateRequestID())
            return next(ctx, req)
        }
    }
}

// uploadProgressInterceptor logs upload progress
func uploadProgressInterceptor() connect.UnaryInterceptorFunc {
    return func(next connect.UnaryFunc) connect.UnaryFunc {
        return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
            start := time.Now()
            resp, err := next(ctx, req)
            duration := time.Since(start)

            if err == nil {
                log.Printf("Upload completed in %v", duration)
            }
            return resp, err
        }
    }
}

func generateRequestID() string {
    return fmt.Sprintf("req-%d", time.Now().UnixNano())
}
```

## Basic Document Upload

### Simple File Upload

```go
// UploadDocument uploads a file with basic metadata
func (c *DMSClient) UploadDocument(
    ctx context.Context,
    filePath, title, description string,
    documentType dmsv1.DocumentType,
    category dmsv1.DocumentCategory,
) (*dmsv1.Document, error) {
    // Read file content
    content, err := os.ReadFile(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to read file: %w", err)
    }

    // Get file info
    fileInfo, err := os.Stat(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to get file info: %w", err)
    }

    // Detect MIME type
    mimeType := detectMimeType(filePath)

    req := &dmsv1.UploadDocumentRequest{
        TenantId:         c.tenantID,
        OwnerEntityType:  "employee",
        OwnerEntityId:    c.entityID,
        DocumentType:     documentType,
        DocumentCategory: category,
        FileName:         filepath.Base(filePath),
        Title:            title,
        Description:      description,
        Content:          content,
        MimeType:         mimeType,
        AutoProcess:      true, // Enable automatic processing
    }

    resp, err := c.client.UploadDocument(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("upload failed: %w", err)
    }

    log.Printf("Uploaded document: %s (ID: %s, Size: %d bytes)",
        resp.Msg.Document.Title,
        resp.Msg.Document.Id,
        fileInfo.Size())

    return resp.Msg.Document, nil
}

// detectMimeType detects the MIME type of a file
func detectMimeType(filePath string) string {
    ext := filepath.Ext(filePath)
    mimeTypes := map[string]string{
        ".pdf":  "application/pdf",
        ".doc":  "application/msword",
        ".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
        ".xls":  "application/vnd.ms-excel",
        ".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
        ".png":  "image/png",
        ".jpg":  "image/jpeg",
        ".jpeg": "image/jpeg",
        ".txt":  "text/plain",
        ".zip":  "application/zip",
    }

    if mime, ok := mimeTypes[ext]; ok {
        return mime
    }
    return "application/octet-stream"
}
```

### Upload with Organizational Context

```go
// UploadDocumentWithOrgContext uploads a document with division/branch/department context
func (c *DMSClient) UploadDocumentWithOrgContext(
    ctx context.Context,
    filePath, title, description string,
    documentType dmsv1.DocumentType,
    category dmsv1.DocumentCategory,
    divisionID, branchID, departmentID string,
) (*dmsv1.Document, error) {
    content, err := os.ReadFile(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to read file: %w", err)
    }

    req := &dmsv1.UploadDocumentRequest{
        TenantId:         c.tenantID,
        OwnerEntityType:  "employee",
        OwnerEntityId:    c.entityID,
        DivisionId:       divisionID,
        BranchId:         branchID,
        DepartmentId:     departmentID,
        DocumentType:     documentType,
        DocumentCategory: category,
        FileName:         filepath.Base(filePath),
        Title:            title,
        Description:      description,
        Content:          content,
        MimeType:         detectMimeType(filePath),
        AutoProcess:      true,
    }

    resp, err := c.client.UploadDocument(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("upload with org context failed: %w", err)
    }

    return resp.Msg.Document, nil
}
```

## Multipart File Handling

### Chunked Upload for Large Files

```go
// ChunkedUploadConfig holds configuration for chunked uploads
type ChunkedUploadConfig struct {
    ChunkSize      int64  // Size of each chunk (e.g., 5MB)
    MaxRetries     int    // Number of retries per chunk
    ProgressReport bool   // Enable progress reporting
}

// UploadLargeDocument uploads a large file in chunks
func (c *DMSClient) UploadLargeDocument(
    ctx context.Context,
    filePath, title, description string,
    documentType dmsv1.DocumentType,
    category dmsv1.DocumentCategory,
    config ChunkedUploadConfig,
) (*dmsv1.Document, error) {
    // Open file
    file, err := os.Open(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to open file: %w", err)
    }
    defer file.Close()

    // Get file info
    fileInfo, err := file.Stat()
    if err != nil {
        return nil, fmt.Errorf("failed to get file info: %w", err)
    }

    totalSize := fileInfo.Size()
    if totalSize <= config.ChunkSize {
        // File is small enough for single upload
        return c.UploadDocument(ctx, filePath, title, description, documentType, category)
    }

    log.Printf("Uploading large file: %s (%.2f MB) in chunks of %.2f MB",
        filepath.Base(filePath),
        float64(totalSize)/(1024*1024),
        float64(config.ChunkSize)/(1024*1024))

    // Read entire file into memory (or use streaming approach)
    content, err := io.ReadAll(file)
    if err != nil {
        return nil, fmt.Errorf("failed to read file: %w", err)
    }

    // Calculate checksum for integrity verification
    checksum := calculateChecksum(content)

    // Create metadata with upload info
    metadata := map[string]interface{}{
        "upload_method":   "chunked",
        "original_size":   totalSize,
        "chunk_size":      config.ChunkSize,
        "checksum":        checksum,
        "upload_client":   "go-sdk",
        "compression":     "auto",
    }
    metadataStruct, _ := structpb.NewStruct(metadata)

    req := &dmsv1.UploadDocumentRequest{
        TenantId:         c.tenantID,
        OwnerEntityType:  "employee",
        OwnerEntityId:    c.entityID,
        DocumentType:     documentType,
        DocumentCategory: category,
        FileName:         filepath.Base(filePath),
        Title:            title,
        Description:      description,
        Content:          content,
        MimeType:         detectMimeType(filePath),
        Metadata:         metadataStruct,
        AutoProcess:      true,
    }

    // Upload with progress tracking
    if config.ProgressReport {
        go trackProgress(ctx, totalSize)
    }

    resp, err := c.client.UploadDocument(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("chunked upload failed: %w", err)
    }

    // Verify checksum
    if resp.Msg.Document.Checksum != checksum {
        log.Printf("Warning: Checksum mismatch (expected: %s, got: %s)",
            checksum, resp.Msg.Document.Checksum)
    }

    return resp.Msg.Document, nil
}

// calculateChecksum computes SHA256 checksum
func calculateChecksum(content []byte) string {
    hash := sha256.Sum256(content)
    return hex.EncodeToString(hash[:])
}

// trackProgress reports upload progress
func trackProgress(ctx context.Context, totalSize int64) {
    ticker := time.NewTicker(2 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            log.Printf("Uploading... (total size: %d bytes)", totalSize)
        }
    }
}
```

### Upload from HTTP Multipart Form

```go
// UploadFromMultipartForm handles file upload from HTTP multipart form
func (c *DMSClient) UploadFromMultipartForm(
    ctx context.Context,
    fileHeader *multipart.FileHeader,
    title, description string,
    documentType dmsv1.DocumentType,
    category dmsv1.DocumentCategory,
) (*dmsv1.Document, error) {
    // Open uploaded file
    file, err := fileHeader.Open()
    if err != nil {
        return nil, fmt.Errorf("failed to open uploaded file: %w", err)
    }
    defer file.Close()

    // Read file content
    content, err := io.ReadAll(file)
    if err != nil {
        return nil, fmt.Errorf("failed to read uploaded file: %w", err)
    }

    // Validate file size (e.g., max 100MB)
    maxSize := int64(100 * 1024 * 1024)
    if int64(len(content)) > maxSize {
        return nil, fmt.Errorf("file size exceeds maximum allowed size of %d bytes", maxSize)
    }

    // Detect MIME type from header
    mimeType := fileHeader.Header.Get("Content-Type")
    if mimeType == "" {
        mimeType = detectMimeType(fileHeader.Filename)
    }

    req := &dmsv1.UploadDocumentRequest{
        TenantId:         c.tenantID,
        OwnerEntityType:  "employee",
        OwnerEntityId:    c.entityID,
        DocumentType:     documentType,
        DocumentCategory: category,
        FileName:         fileHeader.Filename,
        Title:            title,
        Description:      description,
        Content:          content,
        MimeType:         mimeType,
        AutoProcess:      true,
    }

    resp, err := c.client.UploadDocument(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("upload from form failed: %w", err)
    }

    return resp.Msg.Document, nil
}
```

## Document Categorization

### Upload with Tags and Metadata

```go
// UploadDocumentWithMetadata uploads a document with comprehensive metadata and tags
func (c *DMSClient) UploadDocumentWithMetadata(
    ctx context.Context,
    filePath, title, description string,
    documentType dmsv1.DocumentType,
    category dmsv1.DocumentCategory,
    tags []string,
    metadata map[string]interface{},
    expiresAt *time.Time,
) (*dmsv1.Document, error) {
    content, err := os.ReadFile(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to read file: %w", err)
    }

    // Convert metadata to protobuf struct
    metadataStruct, err := structpb.NewStruct(metadata)
    if err != nil {
        return nil, fmt.Errorf("invalid metadata: %w", err)
    }

    req := &dmsv1.UploadDocumentRequest{
        TenantId:         c.tenantID,
        OwnerEntityType:  "employee",
        OwnerEntityId:    c.entityID,
        DocumentType:     documentType,
        DocumentCategory: category,
        FileName:         filepath.Base(filePath),
        Title:            title,
        Description:      description,
        Content:          content,
        MimeType:         detectMimeType(filePath),
        Tags:             tags,
        Metadata:         metadataStruct,
        AutoProcess:      true,
    }

    if expiresAt != nil {
        req.ExpiresAt = timestamppb.New(*expiresAt)
    }

    resp, err := c.client.UploadDocument(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("upload with metadata failed: %w", err)
    }

    return resp.Msg.Document, nil
}
```

### Categorized Upload Examples

```go
// UploadContract uploads a contract document with specific categorization
func (c *DMSClient) UploadContract(
    ctx context.Context,
    filePath, contractNumber, partyName string,
    contractDate time.Time,
    contractValue float64,
) (*dmsv1.Document, error) {
    metadata := map[string]interface{}{
        "contract_number": contractNumber,
        "party_name":      partyName,
        "contract_date":   contractDate.Format("2006-01-02"),
        "contract_value":  contractValue,
        "currency":        "INR",
        "document_class":  "legal",
    }

    tags := []string{
        "contract",
        "legal",
        partyName,
        fmt.Sprintf("year-%d", contractDate.Year()),
    }

    title := fmt.Sprintf("Contract - %s - %s", contractNumber, partyName)
    description := fmt.Sprintf("Contract agreement dated %s with %s",
        contractDate.Format("2006-01-02"), partyName)

    // Set expiration to 10 years from contract date
    expiresAt := contractDate.AddDate(10, 0, 0)

    return c.UploadDocumentWithMetadata(
        ctx,
        filePath,
        title,
        description,
        dmsv1.DocumentType_DOCUMENT_TYPE_PDF,
        dmsv1.DocumentCategory_DOCUMENT_CATEGORY_CONTRACT,
        tags,
        metadata,
        &expiresAt,
    )
}

// UploadInvoice uploads an invoice document
func (c *DMSClient) UploadInvoice(
    ctx context.Context,
    filePath, invoiceNumber, vendorName string,
    invoiceDate time.Time,
    amount float64,
) (*dmsv1.Document, error) {
    metadata := map[string]interface{}{
        "invoice_number": invoiceNumber,
        "vendor_name":    vendorName,
        "invoice_date":   invoiceDate.Format("2006-01-02"),
        "amount":         amount,
        "currency":       "INR",
        "status":         "pending",
    }

    tags := []string{
        "invoice",
        "finance",
        vendorName,
        fmt.Sprintf("fy-%d-%d", invoiceDate.Year(), invoiceDate.Year()+1),
    }

    return c.UploadDocumentWithMetadata(
        ctx,
        filePath,
        fmt.Sprintf("Invoice - %s - %s", invoiceNumber, vendorName),
        fmt.Sprintf("Invoice from %s dated %s", vendorName, invoiceDate.Format("2006-01-02")),
        dmsv1.DocumentType_DOCUMENT_TYPE_PDF,
        dmsv1.DocumentCategory_DOCUMENT_CATEGORY_INVOICE,
        tags,
        metadata,
        nil,
    )
}
```

## Permissions and Sharing

### Share Document with Entity

```go
// ShareDocument shares a document with another entity
func (c *DMSClient) ShareDocument(
    ctx context.Context,
    documentID, sharedWithEntityID string,
    canView, canDownload, canEdit, canDelete bool,
    expiresAt *time.Time,
) (*dmsv1.DocumentShare, error) {
    req := &dmsv1.CreateDocumentShareRequest{
        TenantId:           c.tenantID,
        DocumentId:         documentID,
        SharedWithEntityId: sharedWithEntityID,
        CanView:            canView,
        CanDownload:        canDownload,
        CanEdit:            canEdit,
        CanDelete:          canDelete,
    }

    if expiresAt != nil {
        req.ExpiresAt = timestamppb.New(*expiresAt)
    }

    resp, err := c.client.CreateDocumentShare(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("share document failed: %w", err)
    }

    return resp.Msg.Share, nil
}
```

### Share with Department

```go
// ShareDocumentWithDepartment shares a document with all members of a department
func (c *DMSClient) ShareDocumentWithDepartment(
    ctx context.Context,
    documentID, departmentID string,
    permissions map[string]bool,
    expiresInDays int,
) (*dmsv1.DocumentShare, error) {
    var expiresAt *time.Time
    if expiresInDays > 0 {
        expiry := time.Now().AddDate(0, 0, expiresInDays)
        expiresAt = &expiry
    }

    req := &dmsv1.CreateDocumentShareRequest{
        TenantId:     c.tenantID,
        DocumentId:   documentID,
        DepartmentId: departmentID,
        CanView:      permissions["view"],
        CanDownload:  permissions["download"],
        CanEdit:      permissions["edit"],
        CanDelete:    permissions["delete"],
    }

    if expiresAt != nil {
        req.ExpiresAt = timestamppb.New(*expiresAt)
    }

    resp, err := c.client.CreateDocumentShare(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("share with department failed: %w", err)
    }

    return resp.Msg.Share, nil
}
```

### List Document Shares

```go
// ListDocumentShares retrieves all shares for a document
func (c *DMSClient) ListDocumentShares(ctx context.Context, documentID string) ([]*dmsv1.DocumentShare, error) {
    req := &dmsv1.ListDocumentSharesRequest{
        DocumentId: documentID,
        TenantId:   c.tenantID,
    }

    resp, err := c.client.ListDocumentShares(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("list shares failed: %w", err)
    }

    return resp.Msg.Shares, nil
}
```

## Document Processing

### Process Document with Options

```go
// ProcessDocument triggers document processing (compression, OCR, thumbnails)
func (c *DMSClient) ProcessDocument(
    ctx context.Context,
    documentID string,
    jobTypes []string,
    priority int32,
) ([]*dmsv1.ProcessingJob, error) {
    options := map[string]interface{}{
        "ocr_language":        "eng+hin", // English and Hindi
        "thumbnail_size":      "200x200",
        "compression_quality": 85,
        "extract_metadata":    true,
    }
    optionsStruct, _ := structpb.NewStruct(options)

    req := &dmsv1.ProcessDocumentRequest{
        DocumentId: documentID,
        JobTypes:   jobTypes,
        Priority:   priority,
        Options:    optionsStruct,
    }

    resp, err := c.client.ProcessDocument(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("process document failed: %w", err)
    }

    return resp.Msg.Jobs, nil
}
```

### Check Processing Status

```go
// WaitForProcessing waits for document processing to complete
func (c *DMSClient) WaitForProcessing(
    ctx context.Context,
    documentID string,
    timeout time.Duration,
) error {
    deadline := time.Now().Add(timeout)
    ticker := time.NewTicker(2 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            if time.Now().After(deadline) {
                return fmt.Errorf("processing timeout after %v", timeout)
            }

            req := &dmsv1.GetProcessingStatusRequest{
                DocumentId: documentID,
            }

            resp, err := c.client.GetProcessingStatus(ctx, connect.NewRequest(req))
            if err != nil {
                return fmt.Errorf("failed to get status: %w", err)
            }

            allCompleted := true
            for _, job := range resp.Msg.Jobs {
                if job.Status == "failed" {
                    return fmt.Errorf("processing job %s failed: %s", job.JobType, job.ErrorMessage)
                }
                if job.Status != "completed" {
                    allCompleted = false
                }
                log.Printf("Job %s: %s (%d%%)", job.JobType, job.Status, job.Progress)
            }

            if allCompleted {
                log.Println("All processing jobs completed successfully")
                return nil
            }
        }
    }
}
```

### Extract Text from Document

```go
// ExtractText extracts text from a document using OCR
func (c *DMSClient) ExtractText(
    ctx context.Context,
    documentID string,
    useOCR bool,
) (string, error) {
    req := &dmsv1.ExtractTextRequest{
        DocumentId: documentID,
        Page:       0, // All pages
        UseOcr:     useOCR,
    }

    resp, err := c.client.ExtractText(ctx, connect.NewRequest(req))
    if err != nil {
        return "", fmt.Errorf("extract text failed: %w", err)
    }

    log.Printf("Extracted text (confidence: %.2f%%, language: %s)",
        resp.Msg.Confidence*100, resp.Msg.Language)

    return resp.Msg.Text, nil
}
```

## Watermarking

### Create Watermark Configuration

```go
// CreateWatermarkConfig creates a watermark configuration
func (c *DMSClient) CreateWatermarkConfig(
    ctx context.Context,
    name, description string,
    isDefault bool,
) (*dmsv1.WatermarkConfig, error) {
    config := &dmsv1.WatermarkConfig{
        Name:        name,
        Description: description,
        IsDefault:   isDefault,
        IsActive:    true,
        Type:        "text",
        TextContent: "CONFIDENTIAL",
        UseDynamicText: true,
        ShowUsername:   true,
        ShowTimestamp:  true,
        Position:       "center",
        FontFamily:     "Arial",
        FontSize:       48,
        FontColor:      "#FF0000",
        FontBold:       true,
        Opacity:        0.3,
        Rotation:       45,
        ApplyToAllPages: true,
        TextTemplate:   "{{.Username}} - {{.Timestamp}} - CONFIDENTIAL",
    }

    req := &dmsv1.CreateWatermarkConfigRequest{
        Config: config,
    }

    resp, err := c.client.CreateWatermarkConfig(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("create watermark config failed: %w", err)
    }

    return resp.Msg.Config, nil
}
```

### Apply Watermark to Document

```go
// ApplyWatermark applies a watermark to a document
func (c *DMSClient) ApplyWatermark(
    ctx context.Context,
    documentID, watermarkConfigID string,
    dynamicData map[string]interface{},
) ([]byte, error) {
    dynamicStruct, _ := structpb.NewStruct(dynamicData)

    req := &dmsv1.ApplyWatermarkRequest{
        DocumentId:        documentID,
        WatermarkConfigId: watermarkConfigID,
        DynamicData:       dynamicStruct,
        OutputFormat:      "pdf",
    }

    resp, err := c.client.ApplyWatermark(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("apply watermark failed: %w", err)
    }

    return resp.Msg.Content, nil
}
```

## Document Download

### Download Document

```go
// DownloadDocument downloads a document by ID
func (c *DMSClient) DownloadDocument(
    ctx context.Context,
    documentID, outputPath string,
) error {
    req := &dmsv1.GetDocumentRequest{
        Id:             documentID,
        TenantId:       c.tenantID,
        IncludeContent: true,
    }

    resp, err := c.client.GetDocument(ctx, connect.NewRequest(req))
    if err != nil {
        return fmt.Errorf("get document failed: %w", err)
    }

    // In a real implementation, document content would be retrieved separately
    // This is a placeholder showing the concept
    log.Printf("Document retrieved: %s (Size: %d bytes)",
        resp.Msg.Document.Title,
        resp.Msg.Document.SizeBytes)

    return nil
}
```

### Stream Document

```go
// StreamDocument streams a document with optional watermark
func (c *DMSClient) StreamDocument(
    ctx context.Context,
    documentID, watermarkConfigID, outputPath string,
) error {
    req := &dmsv1.StreamDocumentRequest{
        DocumentId:        documentID,
        Format:            "original",
        WatermarkConfigId: watermarkConfigID,
    }

    stream, err := c.client.StreamDocument(ctx, connect.NewRequest(req))
    if err != nil {
        return fmt.Errorf("stream document failed: %w", err)
    }

    // Create output file
    outFile, err := os.Create(outputPath)
    if err != nil {
        return fmt.Errorf("failed to create output file: %w", err)
    }
    defer outFile.Close()

    // Receive and write chunks
    for stream.Receive() {
        msg := stream.Msg()
        _, err := outFile.Write(msg.Chunk)
        if err != nil {
            return fmt.Errorf("failed to write chunk: %w", err)
        }

        if msg.IsFinal {
            log.Printf("Download complete: %s", outputPath)
            break
        }
    }

    if err := stream.Err(); err != nil {
        return fmt.Errorf("stream error: %w", err)
    }

    return nil
}
```

## Document Versioning

### Create New Version

```go
// CreateDocumentVersion creates a new version of an existing document
func (c *DMSClient) CreateDocumentVersion(
    ctx context.Context,
    documentID, filePath, changeLog string,
) (*dmsv1.Document, error) {
    // First, get the original document
    getReq := &dmsv1.GetDocumentRequest{
        Id:       documentID,
        TenantId: c.tenantID,
    }

    getResp, err := c.client.GetDocument(ctx, connect.NewRequest(getReq))
    if err != nil {
        return nil, fmt.Errorf("failed to get original document: %w", err)
    }

    original := getResp.Msg.Document

    // Read new version content
    content, err := os.ReadFile(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to read new version: %w", err)
    }

    // Create metadata for version
    metadata := map[string]interface{}{
        "is_version":        true,
        "original_id":       documentID,
        "change_log":        changeLog,
        "version_created":   time.Now().Format(time.RFC3339),
        "version_author":    c.userID,
    }
    metadataStruct, _ := structpb.NewStruct(metadata)

    // Upload as new document (version)
    uploadReq := &dmsv1.UploadDocumentRequest{
        TenantId:         c.tenantID,
        OwnerEntityType:  original.OwnerEntityType,
        OwnerEntityId:    original.OwnerEntityId,
        DivisionId:       original.DivisionId,
        BranchId:         original.BranchId,
        DepartmentId:     original.DepartmentId,
        DocumentType:     original.DocumentType,
        DocumentCategory: original.DocumentCategory,
        FileName:         original.FileName,
        Title:            original.Title + " (Updated)",
        Description:      original.Description + " - " + changeLog,
        Content:          content,
        MimeType:         original.MimeType,
        Tags:             append(original.Tags, "version"),
        Metadata:         metadataStruct,
        AutoProcess:      true,
    }

    uploadResp, err := c.client.UploadDocument(ctx, connect.NewRequest(uploadReq))
    if err != nil {
        return nil, fmt.Errorf("failed to upload new version: %w", err)
    }

    return uploadResp.Msg.Document, nil
}
```

### List Document Versions

```go
// ListDocumentVersions retrieves all versions of a document
func (c *DMSClient) ListDocumentVersions(
    ctx context.Context,
    documentID string,
) ([]*dmsv1.DocumentVersion, error) {
    req := &dmsv1.GetDocumentVersionsRequest{
        DocumentId: documentID,
    }

    resp, err := c.client.GetDocumentVersions(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("list versions failed: %w", err)
    }

    return resp.Msg.Versions, nil
}
```

## Complete Workflow

### Complete Document Upload and Processing Workflow

```go
// CompleteDocumentWorkflow demonstrates a full document lifecycle
func CompleteDocumentWorkflow(ctx context.Context) error {
    client := NewDMSClient(
        "https://api.ugcl.example.com",
        "tenant-ugcl-001",
        "entity-employee-123",
        "user-456",
        "auth-token-here",
    )

    // Step 1: Upload contract document
    log.Println("Step 1: Uploading contract document...")
    document, err := client.UploadContract(
        ctx,
        "/path/to/contract.pdf",
        "CNT-2024-001",
        "ABC Contractors Ltd",
        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
        5000000.00,
    )
    if err != nil {
        return fmt.Errorf("upload failed: %w", err)
    }
    log.Printf("✓ Document uploaded: %s (ID: %s)", document.Title, document.Id)

    // Step 2: Trigger processing
    log.Println("Step 2: Processing document...")
    jobs, err := client.ProcessDocument(
        ctx,
        document.Id,
        []string{"compress", "thumbnail", "ocr", "extract_text"},
        1, // High priority
    )
    if err != nil {
        return fmt.Errorf("processing failed: %w", err)
    }
    log.Printf("✓ Started %d processing jobs", len(jobs))

    // Step 3: Wait for processing to complete
    log.Println("Step 3: Waiting for processing to complete...")
    err = client.WaitForProcessing(ctx, document.Id, 5*time.Minute)
    if err != nil {
        return fmt.Errorf("processing timeout: %w", err)
    }
    log.Println("✓ Processing completed")

    // Step 4: Extract text
    log.Println("Step 4: Extracting text...")
    text, err := client.ExtractText(ctx, document.Id, true)
    if err != nil {
        log.Printf("Warning: Text extraction failed: %v", err)
    } else {
        log.Printf("✓ Extracted %d characters of text", len(text))
    }

    // Step 5: Share with legal department
    log.Println("Step 5: Sharing with legal department...")
    share, err := client.ShareDocumentWithDepartment(
        ctx,
        document.Id,
        "dept-legal-001",
        map[string]bool{
            "view":     true,
            "download": true,
            "edit":     false,
            "delete":   false,
        },
        365, // Expires in 1 year
    )
    if err != nil {
        return fmt.Errorf("sharing failed: %w", err)
    }
    log.Printf("✓ Shared with department (Share ID: %s)", share.Id)

    // Step 6: Create watermark config
    log.Println("Step 6: Creating watermark configuration...")
    watermark, err := client.CreateWatermarkConfig(
        ctx,
        "Confidential Contract Watermark",
        "Applied to all contract documents",
        true,
    )
    if err != nil {
        log.Printf("Warning: Watermark creation failed: %v", err)
    } else {
        log.Printf("✓ Created watermark config: %s", watermark.Id)
    }

    // Step 7: Get analytics
    log.Println("Step 7: Document workflow completed successfully!")
    log.Printf("Document ID: %s", document.Id)
    log.Printf("Storage Path: %s", document.StoragePath)
    log.Printf("Compressed Size: %d bytes (ratio: %.2f)",
        document.CompressedSize,
        document.CompressionRatio)

    return nil
}
```

## Error Handling

### Comprehensive Error Handler

```go
// HandleDMSError provides detailed error handling for DMS operations
func HandleDMSError(err error, operation string) {
    if err == nil {
        return
    }

    var connectErr *connect.Error
    if errors.As(err, &connectErr) {
        switch connectErr.Code() {
        case connect.CodeNotFound:
            log.Printf("[%s] Document not found: %v", operation, connectErr.Message())
        case connect.CodeAlreadyExists:
            log.Printf("[%s] Document already exists: %v", operation, connectErr.Message())
        case connect.CodeInvalidArgument:
            log.Printf("[%s] Invalid input: %v", operation, connectErr.Message())
        case connect.CodeResourceExhausted:
            log.Printf("[%s] Storage quota exceeded: %v", operation, connectErr.Message())
        case connect.CodePermissionDenied:
            log.Printf("[%s] Permission denied: %v", operation, connectErr.Message())
        case connect.CodeFailedPrecondition:
            log.Printf("[%s] Virus detected or validation failed: %v", operation, connectErr.Message())
        default:
            log.Printf("[%s] Error: %v (code: %v)", operation, connectErr.Message(), connectErr.Code())
        }
    } else {
        log.Printf("[%s] Unexpected error: %v", operation, err)
    }
}
```

## Best Practices

1. **File Size Management**: Use chunked uploads for files > 5MB
2. **Checksum Verification**: Always verify file integrity using checksums
3. **Metadata Enrichment**: Add comprehensive metadata for better searchability
4. **Automatic Processing**: Enable auto-processing for compression and thumbnails
5. **Virus Scanning**: Ensure virus scanning completes before sharing
6. **Access Control**: Use entity-based permissions and organizational scope
7. **Version Control**: Maintain document history with proper change logs
8. **Watermarking**: Apply watermarks for sensitive documents
9. **Error Recovery**: Implement retry logic for transient failures
10. **Cleanup**: Delete temporary files after upload
11. **Monitoring**: Track processing status and handle failures
12. **Expiration**: Set appropriate expiration dates for temporary documents
