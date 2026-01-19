# Document Management System (DMS) Architecture

## Overview
The Document Management System (DMS) is a centralized service for storing, managing, and tracking all file uploads across the application. Instead of each module managing its own files, DMS provides a unified interface for document operations.

## Why Centralized DMS?

### Problems with Distributed File Management
❌ **Without DMS (Current Approach):**
- User module stores identity documents
- Project module stores project files
- Form module stores attachments
- **Issues:**
  - Duplicate code for file upload/storage
  - Inconsistent file naming and organization
  - No centralized audit trail
  - Difficult to implement features like virus scanning, file versioning
  - Hard to track storage quotas per tenant

✅ **With DMS (Proposed Approach):**
- Single source of truth for all documents
- Centralized MinIO integration
- Unified audit logging
- Consistent access control
- Easy to add features (versioning, watermarking, OCR)

---

## Architecture

### High-Level Design

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Modules                       │
│  (User, Project, Form, Vendor, Contractor, etc.)            │
└──────────────┬──────────────────────────────────────────────┘
               │
               │ References document_id
               ▼
┌─────────────────────────────────────────────────────────────┐
│              Document Management Service (DMS)               │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  API Layer (gRPC/Connect)                            │  │
│  │  - UploadDocument                                    │  │
│  │  - GetDocument                                       │  │
│  │  - GetDocumentURL (pre-signed)                       │  │
│  │  - UpdateDocument                                    │  │
│  │  - DeleteDocument                                    │  │
│  │  - ListDocuments                                     │  │
│  │  - VerifyDocument (admin)                            │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Service Layer                                       │  │
│  │  - Validation (file type, size, virus scan)         │  │
│  │  - Storage orchestration                            │  │
│  │  - Metadata management                              │  │
│  │  - Access control                                   │  │
│  │  - Audit logging                                    │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Repository Layer                                    │  │
│  │  - PostgreSQL (metadata)                            │  │
│  │  - MinIO (actual files)                             │  │
│  └──────────────────────────────────────────────────────┘  │
└────────────────┬─────────────────────────────────┬──────────┘
                 │                                 │
                 ▼                                 ▼
        ┌────────────────┐              ┌──────────────────┐
        │   PostgreSQL   │              │      MinIO       │
        │   (Metadata)   │              │   (File Store)   │
        └────────────────┘              └──────────────────┘
```

---

## Database Schema

### Core Tables

#### 1. `documents` (Main metadata table)
```sql
CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,

    -- File Information
    file_name VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL, -- in bytes
    mime_type VARCHAR(100) NOT NULL,
    file_extension VARCHAR(10) NOT NULL,

    -- Storage Information
    storage_path TEXT NOT NULL, -- MinIO object path
    storage_bucket VARCHAR(100) NOT NULL,
    checksum VARCHAR(64), -- SHA256 hash for integrity

    -- Classification
    document_type VARCHAR(50) NOT NULL, -- 'IDENTITY', 'CERTIFICATE', 'PROJECT_FILE', 'FORM_ATTACHMENT'
    document_category VARCHAR(50), -- 'AADHAR', 'PAN', 'PASSPORT', 'DEGREE', 'CONTRACT'

    -- Ownership & Context
    uploaded_by_user_id VARCHAR(255) NOT NULL,
    owner_user_id VARCHAR(255), -- User who owns this document (may differ from uploader)
    owner_entity_type VARCHAR(50), -- 'USER', 'PROJECT', 'FORM', 'VENDOR'
    owner_entity_id VARCHAR(255), -- ID of the owning entity

    -- Access Control
    visibility VARCHAR(20) DEFAULT 'PRIVATE', -- 'PRIVATE', 'TENANT', 'PUBLIC'
    access_roles TEXT[], -- Roles that can access this document

    -- Lifecycle
    status VARCHAR(20) DEFAULT 'ACTIVE', -- 'ACTIVE', 'ARCHIVED', 'DELETED'
    is_verified BOOLEAN DEFAULT false,
    verified_by_user_id VARCHAR(255),
    verified_at TIMESTAMP WITH TIME ZONE,

    -- Expiry (for documents like passports, contracts)
    expires_at TIMESTAMP WITH TIME ZONE,

    -- Versioning
    version INTEGER DEFAULT 1,
    parent_document_id UUID REFERENCES documents(id), -- For versioning
    is_latest_version BOOLEAN DEFAULT true,

    -- Security
    is_encrypted BOOLEAN DEFAULT false,
    encryption_key_id VARCHAR(255), -- Reference to encryption key if encrypted
    virus_scan_status VARCHAR(20) DEFAULT 'PENDING', -- 'PENDING', 'CLEAN', 'INFECTED', 'FAILED'
    virus_scan_at TIMESTAMP WITH TIME ZONE,

    -- Metadata
    metadata JSONB DEFAULT '{}', -- Custom fields per document type
    tags TEXT[],

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- Indexes
    INDEX idx_documents_tenant (tenant_id),
    INDEX idx_documents_type (document_type, document_category),
    INDEX idx_documents_owner (owner_entity_type, owner_entity_id),
    INDEX idx_documents_uploaded_by (uploaded_by_user_id),
    INDEX idx_documents_status (status),
    INDEX idx_documents_expires (expires_at) WHERE expires_at IS NOT NULL,
    INDEX idx_documents_parent (parent_document_id) WHERE parent_document_id IS NOT NULL,
    INDEX idx_documents_metadata ON documents USING GIN(metadata)
);
```

#### 2. `document_access_log` (Audit trail)
```sql
CREATE TABLE document_access_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,

    -- Who accessed
    user_id VARCHAR(255) NOT NULL,

    -- What action
    action VARCHAR(50) NOT NULL, -- 'UPLOAD', 'VIEW', 'DOWNLOAD', 'UPDATE', 'DELETE', 'VERIFY'

    -- Context
    ip_address INET,
    user_agent TEXT,

    -- Timestamp
    accessed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    INDEX idx_doc_access_document (document_id),
    INDEX idx_doc_access_user (user_id),
    INDEX idx_doc_access_tenant (tenant_id),
    INDEX idx_doc_access_time (accessed_at)
);
```

#### 3. `document_shares` (Document sharing)
```sql
CREATE TABLE document_shares (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,

    -- Shared with
    shared_with_user_id VARCHAR(255),
    shared_with_role VARCHAR(100),
    shared_with_email VARCHAR(255), -- For external shares

    -- Permissions
    can_view BOOLEAN DEFAULT true,
    can_download BOOLEAN DEFAULT false,
    can_edit BOOLEAN DEFAULT false,

    -- Lifecycle
    share_token VARCHAR(255) UNIQUE, -- For public/external shares
    expires_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT true,

    -- Metadata
    shared_by_user_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    INDEX idx_doc_shares_document (document_id),
    INDEX idx_doc_shares_user (shared_with_user_id),
    INDEX idx_doc_shares_token (share_token)
);
```

#### 4. `document_thumbnails` (Preview images)
```sql
CREATE TABLE document_thumbnails (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,

    size VARCHAR(20) NOT NULL, -- 'small', 'medium', 'large'
    width INTEGER,
    height INTEGER,
    storage_path TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(document_id, size)
);
```

---

## Document Types & Categories

### Predefined Types

```typescript
enum DocumentType {
  IDENTITY_DOCUMENT = 'IDENTITY_DOCUMENT',
  PROFESSIONAL_CERTIFICATE = 'PROFESSIONAL_CERTIFICATE',
  EDUCATIONAL_CERTIFICATE = 'EDUCATIONAL_CERTIFICATE',
  CONTRACT = 'CONTRACT',
  INVOICE = 'INVOICE',
  RECEIPT = 'RECEIPT',
  PROJECT_FILE = 'PROJECT_FILE',
  FORM_ATTACHMENT = 'FORM_ATTACHMENT',
  PROFILE_PHOTO = 'PROFILE_PHOTO',
  COMPANY_DOCUMENT = 'COMPANY_DOCUMENT',
  OTHER = 'OTHER'
}

enum DocumentCategory {
  // Identity Documents
  AADHAR = 'AADHAR',
  PAN = 'PAN',
  PASSPORT = 'PASSPORT',
  DRIVING_LICENSE = 'DRIVING_LICENSE',
  VOTER_ID = 'VOTER_ID',

  // Professional Certificates
  PMP = 'PMP',
  AWS_CERT = 'AWS_CERT',
  TRAINING_CERT = 'TRAINING_CERT',

  // Educational
  DEGREE = 'DEGREE',
  DIPLOMA = 'DIPLOMA',
  MARKSHEET = 'MARKSHEET',

  // Company Documents
  GST_CERTIFICATE = 'GST_CERTIFICATE',
  INCORPORATION_CERT = 'INCORPORATION_CERT',

  // Contracts
  EMPLOYMENT_CONTRACT = 'EMPLOYMENT_CONTRACT',
  VENDOR_CONTRACT = 'VENDOR_CONTRACT',
  NDA = 'NDA',

  // Others
  CUSTOM = 'CUSTOM'
}
```

---

## Proto Definitions

```protobuf
syntax = "proto3";

package dms.api.v1.dms;

import "google/protobuf/timestamp.proto";
import "google/protobuf/struct.proto";

service DocumentService {
  // Core Operations
  rpc UploadDocument(UploadDocumentRequest) returns (Document);
  rpc GetDocument(GetDocumentRequest) returns (Document);
  rpc GetDocumentURL(GetDocumentRequest) returns (DocumentURLResponse);
  rpc UpdateDocument(UpdateDocumentRequest) returns (Document);
  rpc DeleteDocument(DeleteDocumentRequest) returns (google.protobuf.Empty);
  rpc ListDocuments(ListDocumentsRequest) returns (ListDocumentsResponse);

  // Document Management
  rpc VerifyDocument(VerifyDocumentRequest) returns (Document);
  rpc ArchiveDocument(ArchiveDocumentRequest) returns (Document);
  rpc CreateDocumentVersion(CreateVersionRequest) returns (Document);
  rpc GetDocumentVersions(GetDocumentRequest) returns (ListDocumentsResponse);

  // Sharing
  rpc ShareDocument(ShareDocumentRequest) returns (DocumentShare);
  rpc RevokeShare(RevokeShareRequest) returns (google.protobuf.Empty);
  rpc ListDocumentShares(GetDocumentRequest) returns (ListSharesResponse);

  // Thumbnails
  rpc GenerateThumbnail(GenerateThumbnailRequest) returns (Thumbnail);
  rpc GetThumbnail(GetThumbnailRequest) returns (ThumbnailResponse);
}

message Document {
  string id = 1;
  string tenant_id = 2;
  string file_name = 3;
  int64 file_size = 4;
  string mime_type = 5;
  string file_extension = 6;
  DocumentType document_type = 7;
  string document_category = 8;
  string uploaded_by_user_id = 9;
  string owner_user_id = 10;
  string owner_entity_type = 11;
  string owner_entity_id = 12;
  string visibility = 13;
  string status = 14;
  bool is_verified = 15;
  string verified_by_user_id = 16;
  google.protobuf.Timestamp verified_at = 17;
  google.protobuf.Timestamp expires_at = 18;
  int32 version = 19;
  string parent_document_id = 20;
  bool is_latest_version = 21;
  string virus_scan_status = 22;
  google.protobuf.Struct metadata = 23;
  repeated string tags = 24;
  google.protobuf.Timestamp created_at = 25;
  google.protobuf.Timestamp updated_at = 26;
}

message UploadDocumentRequest {
  string tenant_id = 1;
  bytes file_data = 2;
  string file_name = 3;
  string mime_type = 4;
  DocumentType document_type = 5;
  string document_category = 6;
  string owner_user_id = 7;
  string owner_entity_type = 8;
  string owner_entity_id = 9;
  string visibility = 10;
  google.protobuf.Struct metadata = 11;
  repeated string tags = 12;
  google.protobuf.Timestamp expires_at = 13;
}
```

---

## Integration with Other Modules

### Example: User Identity Documents

**Before (Without DMS):**
```sql
-- In user module
CREATE TABLE user_identity_documents (
    id UUID PRIMARY KEY,
    user_id BIGINT,
    document_type VARCHAR(50),
    document_number VARCHAR(100),
    document_file_path TEXT, -- Direct MinIO path
    ...
);
```

**After (With DMS):**
```sql
-- In user module (simplified)
CREATE TABLE user_identity_documents (
    id UUID PRIMARY KEY,
    user_id BIGINT,
    document_type VARCHAR(50),
    document_number VARCHAR(100),
    document_id UUID, -- Reference to DMS
    ...
);
```

**Usage Flow:**
1. User uploads Aadhar card via User Service
2. User Service calls `DMS.UploadDocument()` with:
   - `document_type = IDENTITY_DOCUMENT`
   - `document_category = AADHAR`
   - `owner_entity_type = USER`
   - `owner_entity_id = user_id`
3. DMS returns `document_id`
4. User Service stores `document_id` in `user_identity_documents` table

---

## Benefits of DMS Approach

### ✅ Centralization
- Single place to manage all files
- Consistent file handling logic
- Easier to maintain and update

### ✅ Security
- Centralized access control
- Virus scanning for all uploads
- Encryption at rest
- Audit trail for all document access

### ✅ Features
- Document versioning
- Thumbnail generation
- Document sharing
- Expiry tracking
- Bulk operations

### ✅ Storage Management
- Track storage per tenant
- Implement quotas
- Optimize storage (deduplication via checksum)
- Easy migration to different storage providers

### ✅ Compliance
- GDPR: Easy to delete/anonymize user documents
- Audit logs for compliance
- Document retention policies

---

## Migration Strategy

### Phase 1: Create DMS Module (Week 1)
- ✅ Database schema
- ✅ Proto definitions
- ✅ Basic CRUD operations
- ✅ MinIO integration

### Phase 2: Migrate Existing Modules (Week 2-3)
- Update User module to use DMS for identity documents
- Update Project module for project files
- Update Form module for attachments

### Phase 3: Enhanced Features (Week 4+)
- Document versioning
- Thumbnail generation
- OCR for text extraction
- Document templates
- Bulk operations

---

## Recommendation

**✅ IMPLEMENT DMS** as a separate module for the following reasons:

1. **Scalability**: As you add more modules (HR, Finance, Legal), all can use DMS
2. **Maintainability**: One place to fix bugs or add features
3. **Consistency**: Same file handling everywhere
4. **Future-proof**: Easy to add advanced features (AI/ML for document classification, OCR, etc.)

**Next Steps:**
1. Review and approve this architecture
2. Create DMS module following organization module pattern
3. Update user profile design to reference DMS
4. Migrate existing file upload logic to DMS

