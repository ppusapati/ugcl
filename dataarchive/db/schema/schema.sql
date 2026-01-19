-- =============================================================================
-- DATA ARCHIVING & RETENTION POLICIES DATABASE SCHEMA
-- =============================================================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =============================================================================
-- RETENTION POLICIES TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS retention_policies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,

    -- Scope definition
    entity_type VARCHAR(100) NOT NULL,
    schema_name VARCHAR(100) NOT NULL,
    database_name VARCHAR(100) NOT NULL,

    -- Retention configuration
    retention_period INTERVAL NOT NULL,
    archival_period INTERVAL,
    grace_period INTERVAL DEFAULT INTERVAL '30 days',

    -- Policy behavior
    policy_type VARCHAR(50) NOT NULL DEFAULT 'TIME_BASED_RETENTION',
    archival_method VARCHAR(50) NOT NULL DEFAULT 'COLD_STORAGE',
    compression_type VARCHAR(20) DEFAULT 'ZSTD',
    encrypt_archive BOOLEAN DEFAULT true,

    -- Selection and exclusion criteria
    selection_criteria JSONB DEFAULT '{}'::jsonb,
    exclusion_rules JSONB DEFAULT '{}'::jsonb,

    -- Compliance and legal hold
    legal_hold_enabled BOOLEAN DEFAULT false,
    compliance_level VARCHAR(20) DEFAULT 'STANDARD',
    regulatory_basis TEXT[] DEFAULT ARRAY[]::TEXT[],

    -- Execution schedule
    schedule_enabled BOOLEAN DEFAULT true,
    schedule_cron VARCHAR(100) DEFAULT '0 2 * * *', -- Daily at 2 AM
    batch_size INTEGER DEFAULT 1000,
    parallel_jobs INTEGER DEFAULT 1,

    -- Notification settings
    notify_on_start BOOLEAN DEFAULT false,
    notify_on_complete BOOLEAN DEFAULT true,
    notify_on_error BOOLEAN DEFAULT true,
    notification_channels TEXT[] DEFAULT ARRAY['EMAIL']::TEXT[],

    -- Policy metadata
    is_active BOOLEAN DEFAULT true,
    last_executed TIMESTAMP WITH TIME ZONE,
    next_execution TIMESTAMP WITH TIME ZONE,

    created_by UUID NOT NULL,
    updated_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::jsonb,

    -- Constraints
    CONSTRAINT chk_policy_type CHECK (policy_type IN ('TIME_BASED_RETENTION', 'SIZE_BASED_RETENTION', 'EVENT_BASED_RETENTION', 'COMPLIANCE_RETENTION', 'CUSTOM_RETENTION')),
    CONSTRAINT chk_archival_method CHECK (archival_method IN ('COLD_STORAGE', 'COMPRESSION', 'PARTITION', 'EXPORT', 'DELETE')),
    CONSTRAINT chk_compression_type CHECK (compression_type IN ('NONE', 'GZIP', 'ZSTD', 'LZ4', 'BROTLI')),
    CONSTRAINT chk_compliance_level CHECK (compliance_level IN ('NONE', 'STANDARD', 'HIGH', 'CRITICAL'))
);

-- =============================================================================
-- ARCHIVAL JOBS TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS archival_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    policy_id UUID NOT NULL REFERENCES retention_policies(id) ON DELETE RESTRICT,
    job_type VARCHAR(20) NOT NULL DEFAULT 'ARCHIVE',
    status VARCHAR(20) DEFAULT 'PENDING',

    -- Job configuration
    target_table VARCHAR(200) NOT NULL,
    date_range_start TIMESTAMP WITH TIME ZONE,
    date_range_end TIMESTAMP WITH TIME ZONE,
    batch_size INTEGER DEFAULT 1000,

    -- Execution details
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    estimated_end TIMESTAMP WITH TIME ZONE,

    -- Progress tracking
    total_records BIGINT DEFAULT 0,
    processed_records BIGINT DEFAULT 0,
    archived_records BIGINT DEFAULT 0,
    deleted_records BIGINT DEFAULT 0,
    error_count BIGINT DEFAULT 0,

    -- Storage information
    original_size BIGINT DEFAULT 0,
    compressed_size BIGINT DEFAULT 0,
    compression_ratio DECIMAL(5,4) DEFAULT 0,

    -- Archive location
    archive_location TEXT,
    archive_format VARCHAR(50) DEFAULT 'PARQUET',
    encryption_key_id VARCHAR(255), -- Reference to key management system

    -- Error handling
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,

    -- Execution metadata
    executed_by UUID NOT NULL,
    job_metadata JSONB DEFAULT '{}'::jsonb,

    CONSTRAINT chk_job_type CHECK (job_type IN ('ARCHIVE', 'RESTORE', 'DELETE', 'AUDIT', 'PURGE')),
    CONSTRAINT chk_job_status CHECK (status IN ('PENDING', 'RUNNING', 'COMPLETED', 'FAILED', 'CANCELLED', 'PARTIAL'))
);

-- =============================================================================
-- LEGAL HOLDS TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS legal_holds (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,

    -- Hold scope
    entity_types TEXT[] NOT NULL,
    entity_ids TEXT[] DEFAULT ARRAY[]::TEXT[],
    date_range_start TIMESTAMP WITH TIME ZONE,
    date_range_end TIMESTAMP WITH TIME ZONE,

    -- Legal details
    case_number VARCHAR(100),
    legal_basis TEXT NOT NULL,
    issuing_authority VARCHAR(255),
    contact_info TEXT,

    -- Hold status
    is_active BOOLEAN DEFAULT true,
    start_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    end_date TIMESTAMP WITH TIME ZONE,

    -- Affected policies
    affected_policies UUID[] DEFAULT ARRAY[]::UUID[],

    created_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::jsonb
);

-- =============================================================================
-- ARCHIVED DATA TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS archived_data (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    policy_id UUID NOT NULL REFERENCES retention_policies(id) ON DELETE RESTRICT,
    job_id UUID NOT NULL REFERENCES archival_jobs(id) ON DELETE RESTRICT,

    -- Original data reference
    source_table VARCHAR(200) NOT NULL,
    source_schema VARCHAR(100) NOT NULL,
    source_database VARCHAR(100) NOT NULL,
    original_id VARCHAR(255) NOT NULL,

    -- Archive information
    archive_format VARCHAR(50) DEFAULT 'PARQUET',
    storage_location TEXT NOT NULL,
    storage_tier VARCHAR(20) DEFAULT 'COLD',

    -- Data characteristics
    data_size BIGINT NOT NULL,
    compressed_size BIGINT,
    is_encrypted BOOLEAN DEFAULT true,
    encryption_algo VARCHAR(50) DEFAULT 'AES-256-GCM',

    -- Checksums for integrity
    original_checksum VARCHAR(128),
    archive_checksum VARCHAR(128),

    -- Timeline
    original_date TIMESTAMP WITH TIME ZONE NOT NULL,
    archived_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,

    -- Legal hold status
    legal_hold_ids UUID[] DEFAULT ARRAY[]::UUID[],
    is_on_legal_hold BOOLEAN DEFAULT false,

    -- Access tracking
    last_accessed_at TIMESTAMP WITH TIME ZONE,
    access_count INTEGER DEFAULT 0,

    metadata JSONB DEFAULT '{}'::jsonb,

    CONSTRAINT chk_storage_tier CHECK (storage_tier IN ('HOT', 'WARM', 'COLD', 'FROZEN')),
    UNIQUE(source_table, source_schema, source_database, original_id)
);

-- =============================================================================
-- DATA INVENTORY TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS data_inventory (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Data source
    database_name VARCHAR(100) NOT NULL,
    schema_name VARCHAR(100) NOT NULL,
    table_name VARCHAR(200) NOT NULL,

    -- Data characteristics
    record_count BIGINT DEFAULT 0,
    data_size BIGINT DEFAULT 0,
    oldest_record TIMESTAMP WITH TIME ZONE,
    newest_record TIMESTAMP WITH TIME ZONE,

    -- Classification
    data_classification VARCHAR(20) DEFAULT 'INTERNAL',
    contains_pii BOOLEAN DEFAULT false,
    contains_phi BOOLEAN DEFAULT false,
    contains_pci BOOLEAN DEFAULT false,

    -- Retention requirements
    min_retention_period INTERVAL,
    max_retention_period INTERVAL,
    regulatory_requirements TEXT[] DEFAULT ARRAY[]::TEXT[],

    -- Discovery metadata
    last_scanned TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    scan_method VARCHAR(50) DEFAULT 'AUTOMATED',
    confidence DECIMAL(3,2) DEFAULT 0.95,

    metadata JSONB DEFAULT '{}'::jsonb,

    CONSTRAINT chk_data_classification CHECK (data_classification IN ('PUBLIC', 'INTERNAL', 'CONFIDENTIAL', 'RESTRICTED')),
    UNIQUE(database_name, schema_name, table_name)
);

-- =============================================================================
-- COMPLIANCE AUDITS TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS compliance_audits (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    audit_type VARCHAR(20) NOT NULL DEFAULT 'COMPLIANCE',

    -- Audit scope
    policy_ids UUID[] DEFAULT ARRAY[]::UUID[],
    entity_types TEXT[] DEFAULT ARRAY[]::TEXT[],
    date_range_start TIMESTAMP WITH TIME ZONE,
    date_range_end TIMESTAMP WITH TIME ZONE,

    -- Audit results
    status VARCHAR(20) DEFAULT 'SCHEDULED',
    compliance_score DECIMAL(5,4) DEFAULT 0,
    violations JSONB DEFAULT '[]'::jsonb,
    recommendations JSONB DEFAULT '[]'::jsonb,

    -- Execution details
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    executed_by UUID NOT NULL,

    -- Report details
    report_location TEXT,
    report_format VARCHAR(20) DEFAULT 'PDF',

    metadata JSONB DEFAULT '{}'::jsonb,

    CONSTRAINT chk_audit_type CHECK (audit_type IN ('COMPLIANCE', 'INTEGRITY', 'ACCESS', 'RETENTION')),
    CONSTRAINT chk_audit_status CHECK (status IN ('SCHEDULED', 'RUNNING', 'COMPLETED', 'FAILED'))
);

-- =============================================================================
-- ARCHIVE ACCESS LOGS TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS archive_access_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    archived_data_id UUID NOT NULL REFERENCES archived_data(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    access_type VARCHAR(20) NOT NULL, -- 'READ', 'RESTORE', 'DOWNLOAD'

    -- Access context
    purpose TEXT,
    justification TEXT,
    approval_required BOOLEAN DEFAULT false,
    approval_id UUID, -- Reference to approval workflow

    -- Technical details
    ip_address INET,
    user_agent TEXT,
    session_id VARCHAR(255),

    -- Timing and status
    requested_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    granted_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) DEFAULT 'REQUESTED',

    -- Data retrieved
    data_size BIGINT,
    download_location TEXT,
    expiry_time TIMESTAMP WITH TIME ZONE,

    metadata JSONB DEFAULT '{}'::jsonb,

    CONSTRAINT chk_access_type CHECK (access_type IN ('READ', 'RESTORE', 'DOWNLOAD', 'AUDIT')),
    CONSTRAINT chk_access_status CHECK (status IN ('REQUESTED', 'APPROVED', 'GRANTED', 'COMPLETED', 'DENIED', 'EXPIRED'))
);

-- =============================================================================
-- INDEXES
-- =============================================================================

-- Retention policies indexes
CREATE INDEX IF NOT EXISTS idx_retention_policies_entity_type ON retention_policies(entity_type);
CREATE INDEX IF NOT EXISTS idx_retention_policies_active ON retention_policies(is_active) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_retention_policies_next_execution ON retention_policies(next_execution) WHERE next_execution IS NOT NULL AND is_active = true;
CREATE INDEX IF NOT EXISTS idx_retention_policies_database_schema ON retention_policies(database_name, schema_name);

-- Archival jobs indexes
CREATE INDEX IF NOT EXISTS idx_archival_jobs_policy_id ON archival_jobs(policy_id);
CREATE INDEX IF NOT EXISTS idx_archival_jobs_status ON archival_jobs(status);
CREATE INDEX IF NOT EXISTS idx_archival_jobs_started_at ON archival_jobs(started_at);
CREATE INDEX IF NOT EXISTS idx_archival_jobs_target_table ON archival_jobs(target_table);

-- Legal holds indexes
CREATE INDEX IF NOT EXISTS idx_legal_holds_active ON legal_holds(is_active) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_legal_holds_entity_types ON legal_holds USING GIN(entity_types);
CREATE INDEX IF NOT EXISTS idx_legal_holds_case_number ON legal_holds(case_number);
CREATE INDEX IF NOT EXISTS idx_legal_holds_date_range ON legal_holds(start_date, end_date);

-- Archived data indexes
CREATE INDEX IF NOT EXISTS idx_archived_data_policy_id ON archived_data(policy_id);
CREATE INDEX IF NOT EXISTS idx_archived_data_job_id ON archived_data(job_id);
CREATE INDEX IF NOT EXISTS idx_archived_data_source ON archived_data(source_database, source_schema, source_table);
CREATE INDEX IF NOT EXISTS idx_archived_data_original_date ON archived_data(original_date);
CREATE INDEX IF NOT EXISTS idx_archived_data_archived_at ON archived_data(archived_at);
CREATE INDEX IF NOT EXISTS idx_archived_data_expires_at ON archived_data(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_archived_data_legal_hold ON archived_data(is_on_legal_hold) WHERE is_on_legal_hold = true;
CREATE INDEX IF NOT EXISTS idx_archived_data_storage_tier ON archived_data(storage_tier);

-- Data inventory indexes
CREATE INDEX IF NOT EXISTS idx_data_inventory_source ON data_inventory(database_name, schema_name, table_name);
CREATE INDEX IF NOT EXISTS idx_data_inventory_classification ON data_inventory(data_classification);
CREATE INDEX IF NOT EXISTS idx_data_inventory_pii ON data_inventory(contains_pii) WHERE contains_pii = true;
CREATE INDEX IF NOT EXISTS idx_data_inventory_last_scanned ON data_inventory(last_scanned);

-- Compliance audits indexes
CREATE INDEX IF NOT EXISTS idx_compliance_audits_type ON compliance_audits(audit_type);
CREATE INDEX IF NOT EXISTS idx_compliance_audits_status ON compliance_audits(status);
CREATE INDEX IF NOT EXISTS idx_compliance_audits_started_at ON compliance_audits(started_at);
CREATE INDEX IF NOT EXISTS idx_compliance_audits_policy_ids ON compliance_audits USING GIN(policy_ids);

-- Archive access logs indexes
CREATE INDEX IF NOT EXISTS idx_archive_access_logs_archived_data_id ON archive_access_logs(archived_data_id);
CREATE INDEX IF NOT EXISTS idx_archive_access_logs_user_id ON archive_access_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_archive_access_logs_access_type ON archive_access_logs(access_type);
CREATE INDEX IF NOT EXISTS idx_archive_access_logs_requested_at ON archive_access_logs(requested_at);
CREATE INDEX IF NOT EXISTS idx_archive_access_logs_status ON archive_access_logs(status);

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

-- Apply updated_at triggers
CREATE TRIGGER update_retention_policies_updated_at BEFORE UPDATE ON retention_policies
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_legal_holds_updated_at BEFORE UPDATE ON legal_holds
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Trigger to update next execution time
CREATE OR REPLACE FUNCTION update_next_execution()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.schedule_enabled = true AND NEW.schedule_cron IS NOT NULL THEN
        -- This would calculate next execution based on cron expression
        -- For simplicity, setting it to tomorrow at the same time
        NEW.next_execution = NOW() + INTERVAL '1 day';
    ELSE
        NEW.next_execution = NULL;
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_retention_policy_next_execution
    BEFORE INSERT OR UPDATE ON retention_policies
    FOR EACH ROW EXECUTE FUNCTION update_next_execution();

-- =============================================================================
-- FUNCTIONS
-- =============================================================================

-- Function to check if data is subject to legal hold
CREATE OR REPLACE FUNCTION is_data_on_legal_hold(
    entity_type_param VARCHAR,
    entity_id_param VARCHAR,
    data_date TIMESTAMP WITH TIME ZONE
)
RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM legal_holds lh
        WHERE lh.is_active = true
        AND entity_type_param = ANY(lh.entity_types)
        AND (
            ARRAY_LENGTH(lh.entity_ids, 1) IS NULL
            OR entity_id_param = ANY(lh.entity_ids)
        )
        AND (
            lh.date_range_start IS NULL
            OR data_date >= lh.date_range_start
        )
        AND (
            lh.date_range_end IS NULL
            OR data_date <= lh.date_range_end
        )
    );
END;
$$ LANGUAGE plpgsql;

-- Function to get policies due for execution
CREATE OR REPLACE FUNCTION get_policies_due_for_execution()
RETURNS TABLE(
    policy_id UUID,
    policy_name VARCHAR,
    entity_type VARCHAR,
    last_executed TIMESTAMP WITH TIME ZONE,
    next_execution TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        rp.id as policy_id,
        rp.name as policy_name,
        rp.entity_type,
        rp.last_executed,
        rp.next_execution
    FROM retention_policies rp
    WHERE rp.is_active = true
    AND rp.schedule_enabled = true
    AND (
        rp.next_execution IS NULL
        OR rp.next_execution <= NOW()
    )
    ORDER BY rp.next_execution ASC NULLS FIRST;
END;
$$ LANGUAGE plpgsql;

-- Function to calculate storage savings
CREATE OR REPLACE FUNCTION calculate_storage_savings(
    start_date DATE DEFAULT NULL,
    end_date DATE DEFAULT NULL
)
RETURNS TABLE(
    total_original_size BIGINT,
    total_compressed_size BIGINT,
    total_savings BIGINT,
    savings_percentage DECIMAL
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        COALESCE(SUM(ad.data_size), 0) as total_original_size,
        COALESCE(SUM(ad.compressed_size), 0) as total_compressed_size,
        COALESCE(SUM(ad.data_size) - SUM(ad.compressed_size), 0) as total_savings,
        CASE
            WHEN SUM(ad.data_size) > 0
            THEN ((SUM(ad.data_size) - SUM(ad.compressed_size))::DECIMAL / SUM(ad.data_size)) * 100
            ELSE 0
        END as savings_percentage
    FROM archived_data ad
    WHERE (start_date IS NULL OR ad.archived_at::date >= start_date)
    AND (end_date IS NULL OR ad.archived_at::date <= end_date);
END;
$$ LANGUAGE plpgsql;

-- =============================================================================
-- VIEWS
-- =============================================================================

-- View for retention policy dashboard
CREATE OR REPLACE VIEW retention_policy_dashboard AS
SELECT
    rp.id,
    rp.name,
    rp.entity_type,
    rp.policy_type,
    rp.is_active,
    rp.last_executed,
    rp.next_execution,
    CASE
        WHEN rp.next_execution IS NOT NULL AND rp.next_execution < NOW()
        THEN true
        ELSE false
    END as is_overdue,
    COUNT(aj.id) as total_jobs,
    COUNT(aj.id) FILTER (WHERE aj.status = 'COMPLETED') as completed_jobs,
    COUNT(aj.id) FILTER (WHERE aj.status = 'FAILED') as failed_jobs,
    COUNT(aj.id) FILTER (WHERE aj.status IN ('PENDING', 'RUNNING')) as active_jobs,
    COALESCE(SUM(aj.archived_records), 0) as total_archived_records
FROM retention_policies rp
LEFT JOIN archival_jobs aj ON rp.id = aj.policy_id
GROUP BY rp.id, rp.name, rp.entity_type, rp.policy_type, rp.is_active, rp.last_executed, rp.next_execution;

-- View for compliance overview
CREATE OR REPLACE VIEW compliance_overview AS
SELECT
    di.database_name,
    di.schema_name,
    di.table_name,
    di.data_classification,
    di.contains_pii,
    di.contains_phi,
    di.contains_pci,
    COUNT(rp.id) as policies_count,
    COUNT(lh.id) FILTER (WHERE lh.is_active = true) as active_legal_holds,
    EXISTS(
        SELECT 1 FROM retention_policies rp2
        WHERE rp2.entity_type = di.table_name
        AND rp2.is_active = true
    ) as has_active_policy
FROM data_inventory di
LEFT JOIN retention_policies rp ON di.table_name = rp.entity_type
LEFT JOIN legal_holds lh ON di.table_name = ANY(lh.entity_types)
GROUP BY di.database_name, di.schema_name, di.table_name, di.data_classification,
         di.contains_pii, di.contains_phi, di.contains_pci;