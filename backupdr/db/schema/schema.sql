-- =============================================================================
-- BACKUP & DISASTER RECOVERY DATABASE SCHEMA
-- =============================================================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =============================================================================
-- BACKUP POLICIES TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS backup_policies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,

    -- Backup scope
    backup_type VARCHAR(20) NOT NULL DEFAULT 'FULL',
    target_type VARCHAR(20) NOT NULL DEFAULT 'DATABASE',
    target_name VARCHAR(255) NOT NULL,
    include_rules TEXT[] DEFAULT ARRAY[]::TEXT[],
    exclude_rules TEXT[] DEFAULT ARRAY[]::TEXT[],

    -- Schedule configuration
    schedule_type VARCHAR(20) DEFAULT 'CRON',
    cron_expression VARCHAR(100) DEFAULT '0 2 * * *', -- Daily at 2 AM
    interval_minutes INTEGER,
    backup_window JSONB DEFAULT '{}'::jsonb,

    -- Retention policies
    retention_policy JSONB NOT NULL DEFAULT '{
        "keep_daily": 7,
        "keep_weekly": 4,
        "keep_monthly": 12,
        "keep_yearly": 3,
        "max_age": 2555
    }'::jsonb,

    -- Storage configuration
    storage_config JSONB NOT NULL DEFAULT '{
        "storage_type": "LOCAL",
        "local_path": "/backups"
    }'::jsonb,
    compression VARCHAR(20) DEFAULT 'ZSTD',
    encryption JSONB DEFAULT '{
        "enabled": true,
        "algorithm": "AES256",
        "key_source": "KMS"
    }'::jsonb,

    -- Performance settings
    max_parallel_jobs INTEGER DEFAULT 1,
    bandwidth_limit BIGINT DEFAULT 0, -- 0 = unlimited
    timeout_minutes INTEGER DEFAULT 480, -- 8 hours

    -- Notification settings
    notify_on_success BOOLEAN DEFAULT false,
    notify_on_failure BOOLEAN DEFAULT true,
    notify_channels TEXT[] DEFAULT ARRAY['EMAIL']::TEXT[],

    -- Policy metadata
    is_active BOOLEAN DEFAULT true,
    last_backup TIMESTAMP WITH TIME ZONE,
    next_backup TIMESTAMP WITH TIME ZONE,

    created_by UUID NOT NULL,
    updated_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::jsonb,

    -- Constraints
    CONSTRAINT chk_backup_type CHECK (backup_type IN ('FULL', 'INCREMENTAL', 'DIFFERENTIAL', 'SNAPSHOT', 'CONTINUOUS')),
    CONSTRAINT chk_target_type CHECK (target_type IN ('DATABASE', 'FILES', 'APPLICATION', 'SYSTEM', 'VOLUME')),
    CONSTRAINT chk_schedule_type CHECK (schedule_type IN ('CRON', 'INTERVAL', 'MANUAL', 'EVENT')),
    CONSTRAINT chk_compression CHECK (compression IN ('NONE', 'GZIP', 'ZSTD', 'LZ4', 'BZIP2'))
);

-- =============================================================================
-- BACKUP JOBS TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS backup_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    policy_id UUID NOT NULL REFERENCES backup_policies(id) ON DELETE RESTRICT,

    -- Job details
    job_type VARCHAR(20) NOT NULL,
    status VARCHAR(20) DEFAULT 'SCHEDULED',
    priority VARCHAR(20) DEFAULT 'NORMAL',

    -- Execution timing
    scheduled_at TIMESTAMP WITH TIME ZONE NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    duration INTERVAL,

    -- Progress tracking
    total_size BIGINT DEFAULT 0,
    processed_size BIGINT DEFAULT 0,
    compressed_size BIGINT DEFAULT 0,
    progress_percent DECIMAL(5,2) DEFAULT 0,
    estimated_finish TIMESTAMP WITH TIME ZONE,

    -- File tracking
    total_files BIGINT DEFAULT 0,
    processed_files BIGINT DEFAULT 0,
    skipped_files BIGINT DEFAULT 0,
    errored_files BIGINT DEFAULT 0,

    -- Storage information
    backup_path TEXT,
    backup_size BIGINT DEFAULT 0,
    compression_ratio DECIMAL(5,4) DEFAULT 0,
    checksum VARCHAR(128),

    -- Performance metrics
    throughput_mbps DECIMAL(10,2) DEFAULT 0,
    average_speed DECIMAL(10,2) DEFAULT 0,
    network_usage BIGINT DEFAULT 0,
    cpu_usage DECIMAL(5,2) DEFAULT 0,

    -- Error handling
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,

    -- Dependencies
    parent_job_id UUID REFERENCES backup_jobs(id) ON DELETE SET NULL,
    baseline_job_id UUID REFERENCES backup_jobs(id) ON DELETE SET NULL,

    executed_by UUID NOT NULL,
    job_metadata JSONB DEFAULT '{}'::jsonb,

    CONSTRAINT chk_job_status CHECK (status IN ('SCHEDULED', 'QUEUED', 'RUNNING', 'COMPLETED', 'FAILED', 'CANCELLED', 'PAUSED', 'RETRYING')),
    CONSTRAINT chk_job_priority CHECK (priority IN ('LOW', 'NORMAL', 'HIGH', 'CRITICAL'))
);

-- =============================================================================
-- DISASTER RECOVERY PLANS TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS disaster_recovery_plans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,

    -- Plan classification
    plan_type VARCHAR(30) NOT NULL DEFAULT 'DATA_RECOVERY',
    severity VARCHAR(20) NOT NULL DEFAULT 'MAJOR',
    scope VARCHAR(20) NOT NULL DEFAULT 'APPLICATION',

    -- Recovery objectives (stored as seconds)
    rto BIGINT NOT NULL, -- Recovery Time Objective in seconds
    rpo BIGINT NOT NULL, -- Recovery Point Objective in seconds
    mttr BIGINT, -- Mean Time To Recovery in seconds
    availability DECIMAL(5,4) DEFAULT 0.99, -- Target availability %

    -- Recovery procedures
    pre_recovery_steps JSONB DEFAULT '[]'::jsonb,
    recovery_steps JSONB NOT NULL DEFAULT '[]'::jsonb,
    post_recovery_steps JSONB DEFAULT '[]'::jsonb,
    rollback_steps JSONB DEFAULT '[]'::jsonb,

    -- Resource requirements
    resource_requirements JSONB DEFAULT '{}'::jsonb,
    dependencies UUID[] DEFAULT ARRAY[]::UUID[],
    prerequisites JSONB DEFAULT '[]'::jsonb,

    -- Testing and validation
    last_tested TIMESTAMP WITH TIME ZONE,
    test_schedule VARCHAR(100), -- Cron expression
    validation_criteria JSONB DEFAULT '[]'::jsonb,

    -- Communication plan
    notification_list TEXT[] DEFAULT ARRAY[]::TEXT[],
    escalation_matrix JSONB DEFAULT '{}'::jsonb,
    communication_plan JSONB DEFAULT '{}'::jsonb,

    -- Plan metadata
    is_active BOOLEAN DEFAULT true,
    version INTEGER DEFAULT 1,
    approved_by UUID,
    approved_at TIMESTAMP WITH TIME ZONE,

    created_by UUID NOT NULL,
    updated_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::jsonb,

    CONSTRAINT chk_plan_type CHECK (plan_type IN ('DATA_RECOVERY', 'SYSTEM_FAILOVER', 'SITE_FAILOVER', 'BUSINESS_CONTINUITY', 'CYBER_INCIDENT')),
    CONSTRAINT chk_severity CHECK (severity IN ('MINOR', 'MAJOR', 'CRITICAL', 'CATASTROPHIC')),
    CONSTRAINT chk_scope CHECK (scope IN ('APPLICATION', 'DATABASE', 'SYSTEM', 'DATACENTER', 'ENTERPRISE'))
);

-- =============================================================================
-- RECOVERY EXECUTIONS TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS recovery_executions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    plan_id UUID NOT NULL REFERENCES disaster_recovery_plans(id) ON DELETE RESTRICT,
    execution_type VARCHAR(20) NOT NULL DEFAULT 'ACTUAL',
    trigger_reason TEXT NOT NULL,

    -- Execution details
    status VARCHAR(20) DEFAULT 'INITIATED',
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    estimated_rto BIGINT, -- in seconds
    actual_rto BIGINT, -- in seconds

    -- Progress tracking
    total_steps INTEGER DEFAULT 0,
    completed_steps INTEGER DEFAULT 0,
    failed_steps INTEGER DEFAULT 0,
    current_step TEXT,

    -- Results and metrics
    success_rate DECIMAL(5,2) DEFAULT 0,
    data_recovered BIGINT DEFAULT 0,
    systems_recovered INTEGER DEFAULT 0,

    -- Communication tracking
    notifications_sent JSONB DEFAULT '[]'::jsonb,
    stakeholders_notified TEXT[] DEFAULT ARRAY[]::TEXT[],

    -- Issues and resolutions
    issues_encountered JSONB DEFAULT '[]'::jsonb,
    resolution_actions JSONB DEFAULT '[]'::jsonb,

    executed_by UUID NOT NULL,
    execution_log JSONB DEFAULT '[]'::jsonb,
    metadata JSONB DEFAULT '{}'::jsonb,

    CONSTRAINT chk_execution_type CHECK (execution_type IN ('TEST', 'ACTUAL', 'DRILL', 'PARTIAL')),
    CONSTRAINT chk_execution_status CHECK (status IN ('INITIATED', 'IN_PROGRESS', 'COMPLETED', 'FAILED', 'ABORTED', 'ROLLED_BACK'))
);

-- =============================================================================
-- SYSTEM HEALTH TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS system_health (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    system_name VARCHAR(255) NOT NULL,
    system_type VARCHAR(100) NOT NULL,
    health_status VARCHAR(20) NOT NULL DEFAULT 'UNKNOWN',

    -- Health metrics
    cpu_usage DECIMAL(5,2) DEFAULT 0,
    memory_usage DECIMAL(5,2) DEFAULT 0,
    disk_usage DECIMAL(5,2) DEFAULT 0,
    network_latency DECIMAL(10,2) DEFAULT 0,
    response_time DECIMAL(10,2) DEFAULT 0,

    -- Availability metrics
    uptime BIGINT DEFAULT 0, -- in seconds
    last_downtime TIMESTAMP WITH TIME ZONE,
    availability_pct DECIMAL(5,4) DEFAULT 0,

    -- Backup status
    last_backup_time TIMESTAMP WITH TIME ZONE,
    backup_status VARCHAR(50),
    backup_size BIGINT DEFAULT 0,

    -- Alerts and thresholds
    active_alerts JSONB DEFAULT '[]'::jsonb,
    threshold_breaches JSONB DEFAULT '[]'::jsonb,

    checked_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::jsonb,

    CONSTRAINT chk_health_status CHECK (health_status IN ('HEALTHY', 'WARNING', 'CRITICAL', 'UNKNOWN', 'DOWN')),
    UNIQUE(system_name, checked_at)
);

-- =============================================================================
-- RESTORE REQUESTS TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS restore_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    backup_job_id UUID NOT NULL REFERENCES backup_jobs(id) ON DELETE RESTRICT,
    requested_by UUID NOT NULL,

    -- Restore configuration
    restore_type VARCHAR(20) NOT NULL DEFAULT 'FULL',
    restore_scope VARCHAR(20) NOT NULL DEFAULT 'DATABASE',
    target_location TEXT,
    overwrite_policy VARCHAR(20) DEFAULT 'SKIP',

    -- Selection criteria
    include_filters TEXT[] DEFAULT ARRAY[]::TEXT[],
    exclude_filters TEXT[] DEFAULT ARRAY[]::TEXT[],
    point_in_time TIMESTAMP WITH TIME ZONE,

    -- Status and progress
    status VARCHAR(20) DEFAULT 'REQUESTED',
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    progress_percent DECIMAL(5,2) DEFAULT 0,

    -- Results
    restored_files BIGINT DEFAULT 0,
    restored_size BIGINT DEFAULT 0,
    error_count BIGINT DEFAULT 0,
    error_message TEXT,

    -- Approval workflow
    requires_approval BOOLEAN DEFAULT true,
    approval_status VARCHAR(20) DEFAULT 'PENDING',
    approved_by UUID,
    approved_at TIMESTAMP WITH TIME ZONE,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::jsonb,

    CONSTRAINT chk_restore_type CHECK (restore_type IN ('FULL', 'PARTIAL', 'POINT_IN_TIME', 'INSTANT')),
    CONSTRAINT chk_restore_scope CHECK (restore_scope IN ('DATABASE', 'TABLE', 'FILES', 'SYSTEM')),
    CONSTRAINT chk_overwrite_policy CHECK (overwrite_policy IN ('SKIP', 'OVERWRITE', 'RENAME', 'PROMPT')),
    CONSTRAINT chk_restore_status CHECK (status IN ('REQUESTED', 'APPROVED', 'QUEUED', 'RUNNING', 'COMPLETED', 'FAILED', 'CANCELLED'))
);

-- =============================================================================
-- BACKUP STORAGE LOCATIONS TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS backup_storage_locations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    storage_type VARCHAR(20) NOT NULL,

    -- Storage configuration
    connection_config JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_encrypted BOOLEAN DEFAULT true,
    compression_enabled BOOLEAN DEFAULT true,

    -- Capacity and usage
    total_capacity BIGINT, -- in bytes
    used_capacity BIGINT DEFAULT 0,
    available_capacity BIGINT,

    -- Performance characteristics
    read_bandwidth BIGINT, -- bytes per second
    write_bandwidth BIGINT, -- bytes per second
    access_latency DECIMAL(10,2), -- milliseconds

    -- Health and status
    is_active BOOLEAN DEFAULT true,
    is_healthy BOOLEAN DEFAULT true,
    last_health_check TIMESTAMP WITH TIME ZONE,
    health_details JSONB DEFAULT '{}'::jsonb,

    -- Geographic information
    region VARCHAR(100),
    availability_zone VARCHAR(100),
    geographic_location VARCHAR(255),

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::jsonb,

    CONSTRAINT chk_storage_type CHECK (storage_type IN ('LOCAL', 'S3', 'AZURE', 'GCP', 'NETWORK'))
);

-- =============================================================================
-- INDEXES
-- =============================================================================

-- Backup policies indexes
CREATE INDEX IF NOT EXISTS idx_backup_policies_target ON backup_policies(target_type, target_name);
CREATE INDEX IF NOT EXISTS idx_backup_policies_active ON backup_policies(is_active) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_backup_policies_next_backup ON backup_policies(next_backup) WHERE next_backup IS NOT NULL AND is_active = true;
CREATE INDEX IF NOT EXISTS idx_backup_policies_schedule ON backup_policies(schedule_type, is_active);

-- Backup jobs indexes
CREATE INDEX IF NOT EXISTS idx_backup_jobs_policy_id ON backup_jobs(policy_id);
CREATE INDEX IF NOT EXISTS idx_backup_jobs_status ON backup_jobs(status);
CREATE INDEX IF NOT EXISTS idx_backup_jobs_scheduled_at ON backup_jobs(scheduled_at);
CREATE INDEX IF NOT EXISTS idx_backup_jobs_started_at ON backup_jobs(started_at);
CREATE INDEX IF NOT EXISTS idx_backup_jobs_completed_at ON backup_jobs(completed_at);
CREATE INDEX IF NOT EXISTS idx_backup_jobs_parent ON backup_jobs(parent_job_id) WHERE parent_job_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_backup_jobs_baseline ON backup_jobs(baseline_job_id) WHERE baseline_job_id IS NOT NULL;

-- DR plans indexes
CREATE INDEX IF NOT EXISTS idx_dr_plans_type ON disaster_recovery_plans(plan_type);
CREATE INDEX IF NOT EXISTS idx_dr_plans_severity ON disaster_recovery_plans(severity);
CREATE INDEX IF NOT EXISTS idx_dr_plans_scope ON disaster_recovery_plans(scope);
CREATE INDEX IF NOT EXISTS idx_dr_plans_active ON disaster_recovery_plans(is_active) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_dr_plans_last_tested ON disaster_recovery_plans(last_tested);

-- Recovery executions indexes
CREATE INDEX IF NOT EXISTS idx_recovery_executions_plan_id ON recovery_executions(plan_id);
CREATE INDEX IF NOT EXISTS idx_recovery_executions_type ON recovery_executions(execution_type);
CREATE INDEX IF NOT EXISTS idx_recovery_executions_status ON recovery_executions(status);
CREATE INDEX IF NOT EXISTS idx_recovery_executions_started_at ON recovery_executions(started_at);

-- System health indexes
CREATE INDEX IF NOT EXISTS idx_system_health_system_name ON system_health(system_name);
CREATE INDEX IF NOT EXISTS idx_system_health_status ON system_health(health_status);
CREATE INDEX IF NOT EXISTS idx_system_health_checked_at ON system_health(checked_at);
CREATE INDEX IF NOT EXISTS idx_system_health_latest ON system_health(system_name, checked_at DESC);

-- Restore requests indexes
CREATE INDEX IF NOT EXISTS idx_restore_requests_backup_job ON restore_requests(backup_job_id);
CREATE INDEX IF NOT EXISTS idx_restore_requests_requested_by ON restore_requests(requested_by);
CREATE INDEX IF NOT EXISTS idx_restore_requests_status ON restore_requests(status);
CREATE INDEX IF NOT EXISTS idx_restore_requests_created_at ON restore_requests(created_at);

-- Storage locations indexes
CREATE INDEX IF NOT EXISTS idx_storage_locations_type ON backup_storage_locations(storage_type);
CREATE INDEX IF NOT EXISTS idx_storage_locations_active ON backup_storage_locations(is_active) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_storage_locations_region ON backup_storage_locations(region);

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
CREATE TRIGGER update_backup_policies_updated_at BEFORE UPDATE ON backup_policies
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_dr_plans_updated_at BEFORE UPDATE ON disaster_recovery_plans
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_storage_locations_updated_at BEFORE UPDATE ON backup_storage_locations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Trigger to calculate job duration
CREATE OR REPLACE FUNCTION calculate_job_duration()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.completed_at IS NOT NULL AND NEW.started_at IS NOT NULL THEN
        NEW.duration = NEW.completed_at - NEW.started_at;
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_backup_job_duration BEFORE UPDATE ON backup_jobs
    FOR EACH ROW EXECUTE FUNCTION calculate_job_duration();

-- =============================================================================
-- FUNCTIONS
-- =============================================================================

-- Function to get backup status summary
CREATE OR REPLACE FUNCTION get_backup_status_summary(days_back INTEGER DEFAULT 7)
RETURNS TABLE(
    total_policies INTEGER,
    active_policies INTEGER,
    successful_backups BIGINT,
    failed_backups BIGINT,
    total_backup_size BIGINT,
    avg_backup_duration INTERVAL
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        COUNT(DISTINCT bp.id)::INTEGER as total_policies,
        COUNT(DISTINCT bp.id) FILTER (WHERE bp.is_active = true)::INTEGER as active_policies,
        COUNT(bj.id) FILTER (WHERE bj.status = 'COMPLETED' AND bj.started_at >= NOW() - INTERVAL '%s days', days_back)::BIGINT as successful_backups,
        COUNT(bj.id) FILTER (WHERE bj.status = 'FAILED' AND bj.started_at >= NOW() - INTERVAL '%s days', days_back)::BIGINT as failed_backups,
        COALESCE(SUM(bj.backup_size) FILTER (WHERE bj.status = 'COMPLETED' AND bj.started_at >= NOW() - INTERVAL '%s days', days_back), 0) as total_backup_size,
        AVG(bj.duration) FILTER (WHERE bj.status = 'COMPLETED' AND bj.started_at >= NOW() - INTERVAL '%s days', days_back) as avg_backup_duration
    FROM backup_policies bp
    LEFT JOIN backup_jobs bj ON bp.id = bj.policy_id;
END;
$$ LANGUAGE plpgsql;

-- Function to get overdue backups
CREATE OR REPLACE FUNCTION get_overdue_backups()
RETURNS TABLE(
    policy_id UUID,
    policy_name VARCHAR,
    target_name VARCHAR,
    last_backup TIMESTAMP WITH TIME ZONE,
    next_backup TIMESTAMP WITH TIME ZONE,
    hours_overdue NUMERIC
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        bp.id as policy_id,
        bp.name as policy_name,
        bp.target_name,
        bp.last_backup,
        bp.next_backup,
        EXTRACT(EPOCH FROM (NOW() - bp.next_backup)) / 3600 as hours_overdue
    FROM backup_policies bp
    WHERE bp.is_active = true
    AND bp.next_backup IS NOT NULL
    AND bp.next_backup < NOW()
    ORDER BY bp.next_backup ASC;
END;
$$ LANGUAGE plpgsql;

-- Function to calculate recovery metrics
CREATE OR REPLACE FUNCTION calculate_recovery_metrics(
    start_date DATE DEFAULT NULL,
    end_date DATE DEFAULT NULL
)
RETURNS TABLE(
    total_executions BIGINT,
    successful_executions BIGINT,
    failed_executions BIGINT,
    avg_rto_seconds NUMERIC,
    success_rate DECIMAL
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        COUNT(*)::BIGINT as total_executions,
        COUNT(*) FILTER (WHERE status = 'COMPLETED')::BIGINT as successful_executions,
        COUNT(*) FILTER (WHERE status = 'FAILED')::BIGINT as failed_executions,
        AVG(actual_rto) FILTER (WHERE actual_rto IS NOT NULL) as avg_rto_seconds,
        (COUNT(*) FILTER (WHERE status = 'COMPLETED')::DECIMAL / NULLIF(COUNT(*), 0)) * 100 as success_rate
    FROM recovery_executions
    WHERE (start_date IS NULL OR started_at::date >= start_date)
    AND (end_date IS NULL OR started_at::date <= end_date);
END;
$$ LANGUAGE plpgsql;

-- =============================================================================
-- VIEWS
-- =============================================================================

-- View for backup dashboard
CREATE OR REPLACE VIEW backup_dashboard AS
SELECT
    bp.id,
    bp.name,
    bp.target_type,
    bp.target_name,
    bp.backup_type,
    bp.is_active,
    bp.last_backup,
    bp.next_backup,
    CASE
        WHEN bp.next_backup IS NOT NULL AND bp.next_backup < NOW()
        THEN true
        ELSE false
    END as is_overdue,
    COUNT(bj.id) as total_jobs,
    COUNT(bj.id) FILTER (WHERE bj.status = 'COMPLETED') as successful_jobs,
    COUNT(bj.id) FILTER (WHERE bj.status = 'FAILED') as failed_jobs,
    COUNT(bj.id) FILTER (WHERE bj.status IN ('SCHEDULED', 'QUEUED', 'RUNNING')) as active_jobs,
    COALESCE(SUM(bj.backup_size) FILTER (WHERE bj.status = 'COMPLETED'), 0) as total_backup_size
FROM backup_policies bp
LEFT JOIN backup_jobs bj ON bp.id = bj.policy_id
GROUP BY bp.id, bp.name, bp.target_type, bp.target_name, bp.backup_type, bp.is_active, bp.last_backup, bp.next_backup;

-- View for system health summary
CREATE OR REPLACE VIEW system_health_current AS
SELECT DISTINCT ON (sh.system_name)
    sh.system_name,
    sh.system_type,
    sh.health_status,
    sh.cpu_usage,
    sh.memory_usage,
    sh.disk_usage,
    sh.availability_pct,
    sh.last_backup_time,
    sh.backup_status,
    sh.checked_at
FROM system_health sh
ORDER BY sh.system_name, sh.checked_at DESC;