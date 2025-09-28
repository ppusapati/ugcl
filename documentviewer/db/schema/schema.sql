-- =============================================================================
-- DOCUMENT VIEWER DATABASE SCHEMA
-- =============================================================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =============================================================================
-- DOCUMENTS TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    file_name VARCHAR(500) NOT NULL,
    original_name VARCHAR(500) NOT NULL,
    title VARCHAR(1000),
    description TEXT,
    mime_type VARCHAR(255) NOT NULL,
    file_extension VARCHAR(10),
    size_bytes BIGINT NOT NULL,
    compressed_size BIGINT DEFAULT 0,
    checksum VARCHAR(64) UNIQUE NOT NULL,
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
    processing_status VARCHAR(50) DEFAULT 'pending',
    processing_error TEXT,
    processed_at TIMESTAMP WITH TIME ZONE,

    -- Compression info
    compression_type VARCHAR(50) DEFAULT 'tar_zstd',
    compression_ratio DECIMAL(5,4) DEFAULT 0,

    -- OCR info
    ocr_status VARCHAR(50) DEFAULT 'pending',
    ocr_confidence DECIMAL(5,4) DEFAULT 0,
    ocr_processed_at TIMESTAMP WITH TIME ZONE,

    -- Access control
    uploaded_by VARCHAR(255) NOT NULL,
    permissions JSONB,
    tags JSONB,
    categories JSONB,

    -- Security
    virus_scan_status VARCHAR(50) DEFAULT 'pending',
    virus_scan_result VARCHAR(50),
    virus_scan_at TIMESTAMP WITH TIME ZONE,

    -- Watermark info
    watermark_config JSONB,
    has_watermark BOOLEAN DEFAULT false,

    -- Metadata
    metadata JSONB,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Soft delete
    is_deleted BOOLEAN DEFAULT false,
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by VARCHAR(255)
);

-- Indexes for documents table
CREATE INDEX IF NOT EXISTS idx_documents_file_name ON documents(file_name);
CREATE INDEX IF NOT EXISTS idx_documents_mime_type ON documents(mime_type);
CREATE INDEX IF NOT EXISTS idx_documents_uploaded_by ON documents(uploaded_by);
CREATE INDEX IF NOT EXISTS idx_documents_processing_status ON documents(processing_status);
CREATE INDEX IF NOT EXISTS idx_documents_created_at ON documents(created_at);
CREATE INDEX IF NOT EXISTS idx_documents_is_deleted ON documents(is_deleted) WHERE is_deleted = false;
CREATE INDEX IF NOT EXISTS idx_documents_tags ON documents USING GIN(tags);
CREATE INDEX IF NOT EXISTS idx_documents_categories ON documents USING GIN(categories);
CREATE INDEX IF NOT EXISTS idx_documents_metadata ON documents USING GIN(metadata);
CREATE INDEX IF NOT EXISTS idx_documents_full_text ON documents USING GIN(to_tsvector('english', COALESCE(title, '') || ' ' || COALESCE(extracted_text, '')));

-- =============================================================================
-- DOCUMENT VERSIONS TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS document_versions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    file_name VARCHAR(500) NOT NULL,
    storage_path TEXT NOT NULL,
    size_bytes BIGINT NOT NULL,
    checksum VARCHAR(64) NOT NULL,
    change_log TEXT,
    created_by VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(document_id, version)
);

-- Indexes for document_versions table
CREATE INDEX IF NOT EXISTS idx_document_versions_document_id ON document_versions(document_id);
CREATE INDEX IF NOT EXISTS idx_document_versions_created_at ON document_versions(created_at);
CREATE INDEX IF NOT EXISTS idx_document_versions_is_active ON document_versions(is_active) WHERE is_active = true;

-- =============================================================================
-- DOCUMENT ACCESS LOGS TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS document_access_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    action VARCHAR(50) NOT NULL, -- 'view', 'download', 'print', etc.
    ip_address INET,
    user_agent TEXT,
    duration BIGINT DEFAULT 0, -- in milliseconds
    success BOOLEAN DEFAULT true,
    error_message TEXT,

    -- Watermark info for this access
    watermark_applied BOOLEAN DEFAULT false,
    watermark_data JSONB,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for document_access_logs table
CREATE INDEX IF NOT EXISTS idx_document_access_logs_document_id ON document_access_logs(document_id);
CREATE INDEX IF NOT EXISTS idx_document_access_logs_user_id ON document_access_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_document_access_logs_action ON document_access_logs(action);
CREATE INDEX IF NOT EXISTS idx_document_access_logs_created_at ON document_access_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_document_access_logs_ip_address ON document_access_logs(ip_address);

-- =============================================================================
-- DOCUMENT PROCESSING JOBS TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS document_processing_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    job_type VARCHAR(50) NOT NULL, -- 'compress', 'thumbnail', 'ocr', etc.
    status VARCHAR(50) DEFAULT 'pending', -- 'pending', 'running', 'completed', 'failed'
    priority INTEGER DEFAULT 0,
    progress INTEGER DEFAULT 0, -- percentage 0-100
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    job_data JSONB,
    result JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for document_processing_jobs table
CREATE INDEX IF NOT EXISTS idx_document_processing_jobs_document_id ON document_processing_jobs(document_id);
CREATE INDEX IF NOT EXISTS idx_document_processing_jobs_status ON document_processing_jobs(status);
CREATE INDEX IF NOT EXISTS idx_document_processing_jobs_job_type ON document_processing_jobs(job_type);
CREATE INDEX IF NOT EXISTS idx_document_processing_jobs_priority ON document_processing_jobs(priority DESC);
CREATE INDEX IF NOT EXISTS idx_document_processing_jobs_created_at ON document_processing_jobs(created_at);

-- =============================================================================
-- WATERMARK CONFIGURATIONS TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS watermark_configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    is_default BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,

    -- Watermark content
    type VARCHAR(50) NOT NULL, -- 'text', 'image', 'combined'
    text_content TEXT,
    image_path TEXT,
    logo_path TEXT,

    -- Dynamic content placeholders
    use_dynamic_text BOOLEAN DEFAULT true,
    show_username BOOLEAN DEFAULT true,
    show_timestamp BOOLEAN DEFAULT true,
    show_company_logo BOOLEAN DEFAULT false,
    show_document_info BOOLEAN DEFAULT false,
    show_ip_address BOOLEAN DEFAULT false,

    -- Positioning
    position VARCHAR(50) DEFAULT 'bottom_right', -- 'top_left', 'top_right', 'bottom_left', 'bottom_right', 'center'
    offset_x INTEGER DEFAULT 10,
    offset_y INTEGER DEFAULT 10,
    rotation DECIMAL(5,2) DEFAULT 0,

    -- Appearance
    font_family VARCHAR(100) DEFAULT 'Arial',
    font_size INTEGER DEFAULT 12,
    font_color VARCHAR(7) DEFAULT '#000000', -- hex color
    font_bold BOOLEAN DEFAULT false,
    font_italic BOOLEAN DEFAULT false,

    -- Transparency and effects
    opacity DECIMAL(3,2) DEFAULT 0.5,
    blend_mode VARCHAR(50) DEFAULT 'normal',
    shadow BOOLEAN DEFAULT false,
    shadow_color VARCHAR(7) DEFAULT '#808080',
    shadow_offset_x INTEGER DEFAULT 2,
    shadow_offset_y INTEGER DEFAULT 2,

    -- Size constraints
    max_width INTEGER DEFAULT 200,
    max_height INTEGER DEFAULT 50,
    scale DECIMAL(3,2) DEFAULT 1.0,

    -- Page settings
    apply_to_all_pages BOOLEAN DEFAULT true,
    page_numbers JSONB, -- specific pages if not all

    -- Template for dynamic text
    text_template TEXT DEFAULT '{{.Username}} - {{.Timestamp}}',

    -- Security settings
    prevent_removal BOOLEAN DEFAULT true,
    encrypt_watermark BOOLEAN DEFAULT false,

    -- Usage tracking
    usage_count BIGINT DEFAULT 0,
    last_used_at TIMESTAMP WITH TIME ZONE,

    -- Owner and permissions
    created_by VARCHAR(255) NOT NULL,
    updated_by VARCHAR(255),
    permissions JSONB,

    -- Additional metadata
    tags JSONB,
    metadata JSONB,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for watermark_configs table
CREATE INDEX IF NOT EXISTS idx_watermark_configs_name ON watermark_configs(name);
CREATE INDEX IF NOT EXISTS idx_watermark_configs_is_active ON watermark_configs(is_active) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_watermark_configs_is_default ON watermark_configs(is_default) WHERE is_default = true;
CREATE INDEX IF NOT EXISTS idx_watermark_configs_created_by ON watermark_configs(created_by);
CREATE INDEX IF NOT EXISTS idx_watermark_configs_tags ON watermark_configs USING GIN(tags);

-- =============================================================================
-- WATERMARK TEMPLATES TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS watermark_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    category VARCHAR(100),
    is_built_in BOOLEAN DEFAULT false,
    is_public BOOLEAN DEFAULT false,

    -- Template configuration (JSON of WatermarkConfig)
    config JSONB NOT NULL,

    -- Preview image
    preview_path TEXT,

    -- Usage and ratings
    usage_count BIGINT DEFAULT 0,
    rating DECIMAL(3,2) DEFAULT 0,
    rating_count BIGINT DEFAULT 0,

    -- Owner
    created_by VARCHAR(255) NOT NULL,
    tags JSONB,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for watermark_templates table
CREATE INDEX IF NOT EXISTS idx_watermark_templates_name ON watermark_templates(name);
CREATE INDEX IF NOT EXISTS idx_watermark_templates_category ON watermark_templates(category);
CREATE INDEX IF NOT EXISTS idx_watermark_templates_is_public ON watermark_templates(is_public) WHERE is_public = true;
CREATE INDEX IF NOT EXISTS idx_watermark_templates_created_by ON watermark_templates(created_by);
CREATE INDEX IF NOT EXISTS idx_watermark_templates_rating ON watermark_templates(rating DESC);

-- =============================================================================
-- WATERMARK JOBS TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS watermark_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    config_id UUID REFERENCES watermark_configs(id) ON DELETE SET NULL,
    status VARCHAR(50) DEFAULT 'pending',
    priority INTEGER DEFAULT 0,

    -- Job details
    requested_by VARCHAR(255) NOT NULL,
    job_type VARCHAR(50) NOT NULL, -- 'preview', 'download', 'print'
    output_path TEXT,

    -- Dynamic data for watermark
    dynamic_data JSONB,

    -- Processing info
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    progress INTEGER DEFAULT 0,
    error_message TEXT,
    processing_time BIGINT DEFAULT 0, -- milliseconds

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for watermark_jobs table
CREATE INDEX IF NOT EXISTS idx_watermark_jobs_document_id ON watermark_jobs(document_id);
CREATE INDEX IF NOT EXISTS idx_watermark_jobs_config_id ON watermark_jobs(config_id);
CREATE INDEX IF NOT EXISTS idx_watermark_jobs_status ON watermark_jobs(status);
CREATE INDEX IF NOT EXISTS idx_watermark_jobs_requested_by ON watermark_jobs(requested_by);
CREATE INDEX IF NOT EXISTS idx_watermark_jobs_priority ON watermark_jobs(priority DESC);

-- =============================================================================
-- WATERMARK USAGE TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS watermark_usage (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    config_id UUID NOT NULL REFERENCES watermark_configs(id) ON DELETE CASCADE,
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    action VARCHAR(50) NOT NULL, -- 'preview', 'download', 'print'
    ip_address INET,
    user_agent TEXT,
    success BOOLEAN DEFAULT true,
    error_message TEXT,

    -- Applied watermark details
    watermark_data JSONB,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for watermark_usage table
CREATE INDEX IF NOT EXISTS idx_watermark_usage_config_id ON watermark_usage(config_id);
CREATE INDEX IF NOT EXISTS idx_watermark_usage_document_id ON watermark_usage(document_id);
CREATE INDEX IF NOT EXISTS idx_watermark_usage_user_id ON watermark_usage(user_id);
CREATE INDEX IF NOT EXISTS idx_watermark_usage_action ON watermark_usage(action);
CREATE INDEX IF NOT EXISTS idx_watermark_usage_created_at ON watermark_usage(created_at);

-- =============================================================================
-- TRIGGERS
-- =============================================================================

-- Update updated_at timestamp trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply updated_at triggers to all tables
CREATE TRIGGER update_documents_updated_at BEFORE UPDATE ON documents
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_document_versions_updated_at BEFORE UPDATE ON document_versions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_document_processing_jobs_updated_at BEFORE UPDATE ON document_processing_jobs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_watermark_configs_updated_at BEFORE UPDATE ON watermark_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_watermark_templates_updated_at BEFORE UPDATE ON watermark_templates
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_watermark_jobs_updated_at BEFORE UPDATE ON watermark_jobs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =============================================================================
-- FUNCTIONS
-- =============================================================================

-- Function to get document storage usage by user
CREATE OR REPLACE FUNCTION get_user_storage_usage(user_id_param VARCHAR(255))
RETURNS TABLE(
    total_documents BIGINT,
    total_size_bytes BIGINT,
    compressed_size_bytes BIGINT,
    compression_ratio DECIMAL(5,4)
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        COUNT(*)::BIGINT as total_documents,
        COALESCE(SUM(size_bytes), 0)::BIGINT as total_size_bytes,
        COALESCE(SUM(compressed_size), 0)::BIGINT as compressed_size_bytes,
        CASE
            WHEN COALESCE(SUM(size_bytes), 0) > 0
            THEN (COALESCE(SUM(compressed_size), 0)::DECIMAL / COALESCE(SUM(size_bytes), 1))
            ELSE 0
        END as compression_ratio
    FROM documents
    WHERE uploaded_by = user_id_param
    AND is_deleted = false;
END;
$$ LANGUAGE plpgsql;

-- Function to get document access statistics
CREATE OR REPLACE FUNCTION get_document_access_stats(document_id_param UUID, days_back INTEGER DEFAULT 30)
RETURNS TABLE(
    total_views BIGINT,
    unique_users BIGINT,
    avg_duration DECIMAL(10,2),
    last_accessed TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        COUNT(*)::BIGINT as total_views,
        COUNT(DISTINCT user_id)::BIGINT as unique_users,
        AVG(duration)::DECIMAL(10,2) as avg_duration,
        MAX(created_at) as last_accessed
    FROM document_access_logs
    WHERE document_id = document_id_param
    AND action = 'view'
    AND success = true
    AND created_at >= NOW() - INTERVAL '%s days', days_back;
END;
$$ LANGUAGE plpgsql;

-- =============================================================================
-- VIEWS
-- =============================================================================

-- View for document statistics
CREATE OR REPLACE VIEW document_stats AS
SELECT
    d.id,
    d.file_name,
    d.title,
    d.size_bytes,
    d.page_count,
    d.uploaded_by,
    d.created_at,
    COALESCE(al.view_count, 0) as view_count,
    COALESCE(al.download_count, 0) as download_count,
    COALESCE(al.unique_viewers, 0) as unique_viewers,
    al.last_accessed
FROM documents d
LEFT JOIN (
    SELECT
        document_id,
        COUNT(*) FILTER (WHERE action = 'view') as view_count,
        COUNT(*) FILTER (WHERE action = 'download') as download_count,
        COUNT(DISTINCT user_id) FILTER (WHERE action = 'view') as unique_viewers,
        MAX(created_at) as last_accessed
    FROM document_access_logs
    WHERE success = true
    GROUP BY document_id
) al ON d.id = al.document_id
WHERE d.is_deleted = false;