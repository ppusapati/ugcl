-- =============================================================================
-- Scheduler Database Schema
-- =============================================================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =============================================================================
-- Jobs Table
-- =============================================================================
CREATE TABLE IF NOT EXISTS jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    cron_expression VARCHAR(100) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT false,

    -- Target service configuration
    target_service VARCHAR(100) NOT NULL, -- insightviewer, masters, etc.
    target_method VARCHAR(100) NOT NULL,  -- ExecuteReport, SyncMetadata, etc.
    target_payload JSONB,                 -- Service-specific parameters

    -- Execution settings
    timeout_seconds INTEGER NOT NULL DEFAULT 300,  -- 5 minutes default
    max_retries INTEGER NOT NULL DEFAULT 3,
    retry_interval INTEGER NOT NULL DEFAULT 60,    -- 1 minute default

    -- Scheduling info
    next_run_at TIMESTAMPTZ,
    last_run_at TIMESTAMPTZ,
    last_run_status VARCHAR(50),           -- success, failed, timeout
    run_count INTEGER NOT NULL DEFAULT 0,
    failure_count INTEGER NOT NULL DEFAULT 0,

    -- Notification settings
    notify_on_success BOOLEAN NOT NULL DEFAULT false,
    notify_on_failure BOOLEAN NOT NULL DEFAULT true,
    notification_channels TEXT[],          -- email, slack, webhook
    notification_recipients TEXT[],

    -- Metadata
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by VARCHAR(255),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for jobs table
CREATE INDEX IF NOT EXISTS idx_jobs_active ON jobs(is_active) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_jobs_next_run ON jobs(next_run_at) WHERE next_run_at IS NOT NULL AND is_active = true;
CREATE INDEX IF NOT EXISTS idx_jobs_target_service ON jobs(target_service);
CREATE INDEX IF NOT EXISTS idx_jobs_created_by ON jobs(created_by);

-- =============================================================================
-- Job Executions Table
-- =============================================================================
CREATE TABLE IF NOT EXISTS job_executions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL,          -- running, success, failed, timeout, cancelled
    started_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ,
    duration_ms INTEGER,

    -- Request/Response tracking
    request_payload JSONB,
    response_data JSONB,
    error_message TEXT,
    retry_attempt INTEGER NOT NULL DEFAULT 0,

    -- Execution context
    triggered_by VARCHAR(100) NOT NULL,   -- scheduler, manual, webhook
    host_name VARCHAR(255),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for job_executions table
CREATE INDEX IF NOT EXISTS idx_job_executions_job_id ON job_executions(job_id);
CREATE INDEX IF NOT EXISTS idx_job_executions_status ON job_executions(status);
CREATE INDEX IF NOT EXISTS idx_job_executions_started_at ON job_executions(started_at DESC);
CREATE INDEX IF NOT EXISTS idx_job_executions_triggered_by ON job_executions(triggered_by);

-- =============================================================================
-- Job Execution Logs Table (Optional - for detailed logging)
-- =============================================================================
CREATE TABLE IF NOT EXISTS job_execution_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    execution_id UUID NOT NULL REFERENCES job_executions(id) ON DELETE CASCADE,
    level VARCHAR(20) NOT NULL,           -- info, warn, error, debug
    message TEXT NOT NULL,
    details JSONB,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for job_execution_logs table
CREATE INDEX IF NOT EXISTS idx_job_execution_logs_execution_id ON job_execution_logs(execution_id);
CREATE INDEX IF NOT EXISTS idx_job_execution_logs_level ON job_execution_logs(level);
CREATE INDEX IF NOT EXISTS idx_job_execution_logs_timestamp ON job_execution_logs(timestamp DESC);

-- =============================================================================
-- Triggers for automatic timestamp updates
-- =============================================================================

-- Update jobs.updated_at on row updates
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_jobs_updated_at
    BEFORE UPDATE ON jobs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- =============================================================================
-- Views for monitoring and statistics
-- =============================================================================

-- Job statistics view
CREATE OR REPLACE VIEW job_statistics AS
SELECT
    j.id,
    j.name,
    j.target_service,
    j.target_method,
    j.is_active,
    j.run_count,
    j.failure_count,
    j.last_run_at,
    j.last_run_status,
    j.next_run_at,
    COALESCE(stats.total_executions, 0) as total_executions,
    COALESCE(stats.successful_executions, 0) as successful_executions,
    COALESCE(stats.failed_executions, 0) as failed_executions,
    CASE
        WHEN COALESCE(stats.total_executions, 0) > 0
        THEN ROUND((COALESCE(stats.successful_executions, 0)::numeric / stats.total_executions::numeric) * 100, 2)
        ELSE 0
    END as success_rate_percent,
    COALESCE(stats.avg_duration_ms, 0) as avg_duration_ms,
    stats.last_execution_at
FROM jobs j
LEFT JOIN (
    SELECT
        job_id,
        COUNT(*) as total_executions,
        COUNT(*) FILTER (WHERE status = 'success') as successful_executions,
        COUNT(*) FILTER (WHERE status IN ('failed', 'timeout')) as failed_executions,
        ROUND(AVG(duration_ms)) as avg_duration_ms,
        MAX(completed_at) as last_execution_at
    FROM job_executions
    WHERE completed_at IS NOT NULL
    GROUP BY job_id
) stats ON j.id = stats.job_id;

-- Recent executions view
CREATE OR REPLACE VIEW recent_job_executions AS
SELECT
    je.id,
    je.job_id,
    j.name as job_name,
    j.target_service,
    j.target_method,
    je.status,
    je.started_at,
    je.completed_at,
    je.duration_ms,
    je.error_message,
    je.triggered_by,
    je.retry_attempt
FROM job_executions je
JOIN jobs j ON je.job_id = j.id
ORDER BY je.started_at DESC
LIMIT 1000;

-- =============================================================================
-- Sample Data (Optional - for testing)
-- =============================================================================

-- Insert a sample report execution job
INSERT INTO jobs (
    name,
    description,
    cron_expression,
    is_active,
    target_service,
    target_method,
    target_payload,
    timeout_seconds,
    notify_on_failure,
    notification_channels,
    notification_recipients,
    created_by
) VALUES (
    'Daily Sales Report',
    'Generate daily sales report every morning at 8 AM',
    '0 8 * * *',  -- Every day at 8 AM
    false,        -- Start disabled
    'insightviewer',
    'ExecuteReport',
    '{"report_id": "report-123", "parameters": {"date_range": "yesterday"}}',
    600,          -- 10 minutes timeout
    true,
    ARRAY['email', 'slack'],
    ARRAY['admin@company.com', 'sales-team@company.com'],
    'system'
) ON CONFLICT DO NOTHING;

-- Insert a sample cache cleanup job
INSERT INTO jobs (
    name,
    description,
    cron_expression,
    is_active,
    target_service,
    target_method,
    target_payload,
    timeout_seconds,
    notify_on_success,
    notify_on_failure,
    notification_channels,
    notification_recipients,
    created_by
) VALUES (
    'Cache Cleanup',
    'Clean up expired cache entries every hour',
    '0 */1 * * *', -- Every hour
    false,         -- Start disabled
    'insightviewer',
    'CleanupCache',
    '{"max_age_hours": 24}',
    300,           -- 5 minutes timeout
    false,
    true,
    ARRAY['email'],
    ARRAY['ops@company.com'],
    'system'
) ON CONFLICT DO NOTHING;