-- =============================================================================
-- DataBridge Module - Import and Data Ingestion Schema
-- =============================================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =============================================================================
-- Import mappings - stores CSV field to database column mappings for reuse
-- =============================================================================
CREATE TABLE import_mappings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mapping_name VARCHAR(255) NOT NULL,
    description TEXT,
    table_id UUID NOT NULL, -- References masters.tables_metadata(id)
    csv_headers JSONB NOT NULL DEFAULT '[]'::jsonb, -- Array of CSV header names
    field_mappings JSONB NOT NULL DEFAULT '[]'::jsonb, -- Mapping configuration
    transformation_rules JSONB DEFAULT '[]'::jsonb, -- Data transformation rules
    validation_rules JSONB DEFAULT '[]'::jsonb, -- Import validation rules
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL,
    updated_by VARCHAR(255),
    last_used_at TIMESTAMP WITH TIME ZONE,
    usage_count INTEGER NOT NULL DEFAULT 0,

    -- Ensure unique mapping names per table
    CONSTRAINT uq_mappings_table_name UNIQUE (table_id, mapping_name)
);

-- =============================================================================
-- Import jobs - tracks CSV import operations
-- =============================================================================
CREATE TABLE import_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_name VARCHAR(255) NOT NULL,
    mapping_id UUID NOT NULL REFERENCES import_mappings(id) ON DELETE RESTRICT,
    file_name VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL,
    file_hash VARCHAR(64), -- SHA-256 hash of file content
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, processing, completed, failed, cancelled
    total_rows INTEGER,
    processed_rows INTEGER DEFAULT 0,
    successful_rows INTEGER DEFAULT 0,
    failed_rows INTEGER DEFAULT 0,
    error_details JSONB DEFAULT '[]'::jsonb, -- Array of error records
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL,

    -- Check constraints
    CONSTRAINT chk_import_jobs_status CHECK (status IN (
        'pending', 'processing', 'completed', 'failed', 'cancelled'
    )),
    CONSTRAINT chk_import_jobs_processed_rows CHECK (processed_rows >= 0),
    CONSTRAINT chk_import_jobs_successful_rows CHECK (successful_rows >= 0),
    CONSTRAINT chk_import_jobs_failed_rows CHECK (failed_rows >= 0)
);

-- =============================================================================
-- Create indexes for better performance
-- =============================================================================

-- Import mappings indexes
CREATE INDEX idx_import_mappings_table_id ON import_mappings(table_id);
CREATE INDEX idx_import_mappings_mapping_name ON import_mappings(mapping_name);
CREATE INDEX idx_import_mappings_is_active ON import_mappings(is_active);
CREATE INDEX idx_import_mappings_created_by ON import_mappings(created_by);
CREATE INDEX idx_import_mappings_last_used ON import_mappings(last_used_at DESC);

-- Import jobs indexes
CREATE INDEX idx_import_jobs_mapping_id ON import_jobs(mapping_id);
CREATE INDEX idx_import_jobs_status ON import_jobs(status);
CREATE INDEX idx_import_jobs_created_by ON import_jobs(created_by);
CREATE INDEX idx_import_jobs_created_at ON import_jobs(created_at DESC);
CREATE INDEX idx_import_jobs_status_created ON import_jobs(status, created_at DESC);