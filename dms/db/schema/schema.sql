-- Document type enumeration
CREATE TYPE document_type AS ENUM (
    'PDF',
    'IMAGE',
    'VIDEO',
    'AUDIO',
    'SPREADSHEET',
    'PRESENTATION',
    'TEXT',
    'ARCHIVE',
    'OTHER'
);

-- Document category enumeration
CREATE TYPE document_category AS ENUM (
    'CONTRACT',
    'INVOICE',
    'REPORT',
    'MEMO',
    'LETTER',
    'FORM',
    'POLICY',
    'PROCEDURE',
    'MANUAL',
    'PRESENTATION',
    'DRAWING',
    'PHOTO',
    'VIDEO',
    'AUDIO',
    'OTHER'
);

-- Processing status enumeration
CREATE TYPE processing_status AS ENUM (
    'PENDING',
    'PROCESSING',
    'COMPLETED',
    'FAILED'
);

-- OCR status enumeration
CREATE TYPE ocr_status AS ENUM (
    'PENDING',
    'PROCESSING',
    'COMPLETED',
    'FAILED',
    'SKIPPED'
);

-- Virus scan status enumeration
CREATE TYPE virus_scan_status AS ENUM (
    'PENDING',
    'SCANNING',
    'CLEAN',
    'INFECTED',
    'FAILED'
);

-- Document share permission enumeration
CREATE TYPE share_permission AS ENUM (
    'VIEW',
    'DOWNLOAD',
    'EDIT',
    'DELETE'
);

-- Documents table - main document storage
CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,

    -- Entity-based ownership
    owner_entity_type VARCHAR(50),
    owner_entity_id UUID,

    -- Organizational context
    division_id UUID,
    branch_id UUID,
    department_id UUID,

    -- Document classification
    document_type document_type NOT NULL DEFAULT 'OTHER',
    document_category document_category NOT NULL DEFAULT 'OTHER',

    -- File information
    file_name VARCHAR(500) NOT NULL,
    original_name VARCHAR(500) NOT NULL,
    title VARCHAR(1000),
    description TEXT,
    mime_type VARCHAR(255) NOT NULL,
    file_extension VARCHAR(10),
    size_bytes BIGINT NOT NULL,
    compressed_size BIGINT DEFAULT 0,
    checksum VARCHAR(64) NOT NULL,

    -- Storage paths
    storage_path TEXT NOT NULL,
    compressed_path TEXT,
    thumbnail_path TEXT,
    preview_path TEXT,

    -- Document specific fields
    page_count INTEGER DEFAULT 0,
    word_count INTEGER DEFAULT 0,
    extracted_text TEXT,
    language VARCHAR(10),
    author VARCHAR(255),
    subject VARCHAR(500),
    keywords TEXT,

    -- Processing status
    processing_status processing_status NOT NULL DEFAULT 'PENDING',
    processing_error TEXT,
    processed_at TIMESTAMPTZ,

    -- Compression info
    compression_type VARCHAR(50) DEFAULT 'tar_zstd',
    compression_ratio NUMERIC(5, 2) DEFAULT 0,

    -- OCR info
    ocr_status ocr_status NOT NULL DEFAULT 'PENDING',
    ocr_confidence NUMERIC(5, 2) DEFAULT 0,
    ocr_processed_at TIMESTAMPTZ,

    -- Security
    virus_scan_status virus_scan_status NOT NULL DEFAULT 'PENDING',
    virus_scan_result VARCHAR(50),
    virus_scanned_at TIMESTAMPTZ,

    -- Watermark info
    watermark_config JSONB,
    has_watermark BOOLEAN DEFAULT FALSE,

    -- Metadata and tags
    metadata JSONB DEFAULT '{}',
    tags TEXT[],

    -- Expiration tracking
    expires_at TIMESTAMPTZ,
    is_expired BOOLEAN GENERATED ALWAYS AS (expires_at IS NOT NULL AND expires_at < NOW()) STORED,

    -- Access control
    uploaded_by UUID NOT NULL,
    permissions JSONB DEFAULT '{}',

    -- Soft delete
    is_deleted BOOLEAN DEFAULT FALSE,
    deleted_at TIMESTAMPTZ,
    deleted_by UUID,

    -- Audit fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,

    -- Constraints
    CONSTRAINT unique_checksum_per_tenant UNIQUE(tenant_id, checksum)
);

-- Indexes for documents
CREATE INDEX idx_documents_tenant_id ON documents(tenant_id);
CREATE INDEX idx_documents_owner_entity ON documents(owner_entity_id) WHERE owner_entity_id IS NOT NULL;
CREATE INDEX idx_documents_division ON documents(division_id) WHERE division_id IS NOT NULL;
CREATE INDEX idx_documents_branch ON documents(branch_id) WHERE branch_id IS NOT NULL;
CREATE INDEX idx_documents_department ON documents(department_id) WHERE department_id IS NOT NULL;
CREATE INDEX idx_documents_file_name ON documents(file_name);
CREATE INDEX idx_documents_mime_type ON documents(mime_type);
CREATE INDEX idx_documents_document_type ON documents(document_type);
CREATE INDEX idx_documents_document_category ON documents(document_category);
CREATE INDEX idx_documents_uploaded_by ON documents(uploaded_by);
CREATE INDEX idx_documents_is_deleted ON documents(is_deleted);
CREATE INDEX idx_documents_expires_at ON documents(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX idx_documents_tags ON documents USING GIN(tags);
CREATE INDEX idx_documents_metadata ON documents USING GIN(metadata);
CREATE INDEX idx_documents_created_at ON documents(created_at);

-- Document shares table
CREATE TABLE document_shares (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,

    -- Share target
    shared_with_entity_id UUID,
    shared_with_user_id UUID,
    shared_with_email VARCHAR(255),

    -- Organizational scope
    division_id UUID,
    branch_id UUID,
    department_id UUID,

    -- Share permissions
    can_view BOOLEAN DEFAULT TRUE,
    can_download BOOLEAN DEFAULT FALSE,
    can_edit BOOLEAN DEFAULT FALSE,
    can_delete BOOLEAN DEFAULT FALSE,

    -- Share lifecycle
    share_link VARCHAR(255) UNIQUE,
    share_password VARCHAR(255),
    expires_at TIMESTAMPTZ,
    is_expired BOOLEAN GENERATED ALWAYS AS (expires_at IS NOT NULL AND expires_at < NOW()) STORED,

    -- Access tracking
    access_count INTEGER DEFAULT 0,
    last_accessed_at TIMESTAMPTZ,

    -- Audit fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,

    -- Constraints
    CONSTRAINT share_target_check CHECK (
        shared_with_entity_id IS NOT NULL OR
        shared_with_user_id IS NOT NULL OR
        shared_with_email IS NOT NULL
    )
);

-- Indexes for document_shares
CREATE INDEX idx_document_shares_tenant_id ON document_shares(tenant_id);
CREATE INDEX idx_document_shares_document_id ON document_shares(document_id);
CREATE INDEX idx_document_shares_entity ON document_shares(shared_with_entity_id) WHERE shared_with_entity_id IS NOT NULL;
CREATE INDEX idx_document_shares_user ON document_shares(shared_with_user_id) WHERE shared_with_user_id IS NOT NULL;
CREATE INDEX idx_document_shares_email ON document_shares(shared_with_email) WHERE shared_with_email IS NOT NULL;
CREATE INDEX idx_document_shares_link ON document_shares(share_link) WHERE share_link IS NOT NULL;
CREATE INDEX idx_document_shares_expires_at ON document_shares(expires_at) WHERE expires_at IS NOT NULL;
