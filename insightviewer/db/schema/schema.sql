-- =============================================================================
-- InsightViewer - Report Viewer Service Database Schema
-- =============================================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =============================================================================
-- Report Runs - Track report executions
-- =============================================================================
CREATE TABLE report_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID NOT NULL, -- References insighthub.reports.id
    report_version INTEGER NOT NULL DEFAULT 1,
    run_by UUID NOT NULL,
    run_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) NOT NULL DEFAULT 'running', -- running, completed, failed, cancelled
    duration_ms INTEGER, -- Execution duration in milliseconds
    row_count INTEGER, -- Number of rows returned
    error_message TEXT, -- Error details if failed
    parameters JSONB DEFAULT '{}'::jsonb, -- Parameters used for this run
    execution_plan TEXT, -- SQL query that was executed
    cache_key VARCHAR(255), -- For result caching
    triggered_by VARCHAR(50) DEFAULT 'manual', -- manual, scheduled, api

    -- Metadata
    metadata JSONB DEFAULT '{}'::jsonb,

    -- Constraints
    CONSTRAINT chk_run_status CHECK (status IN ('running', 'completed', 'failed', 'cancelled', 'timeout')),
    CONSTRAINT chk_triggered_by CHECK (triggered_by IN ('manual', 'scheduled', 'api', 'webhook'))
);

-- =============================================================================
-- Report Results - Store report execution results
-- =============================================================================
CREATE TABLE report_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id UUID NOT NULL REFERENCES report_runs(id) ON DELETE CASCADE,
    result_type VARCHAR(20) NOT NULL DEFAULT 'data', -- data, error, metadata

    -- Data storage options
    result_json JSONB, -- For small results (< 1MB)
    result_text TEXT, -- For larger text results
    cache_ref VARCHAR(255), -- Reference to external cache (Redis/S3)
    file_path TEXT, -- Path to result file for very large results

    -- Result metadata
    column_info JSONB DEFAULT '[]'::jsonb, -- Column definitions and types
    row_count INTEGER DEFAULT 0,
    size_bytes INTEGER DEFAULT 0,
    compression VARCHAR(20), -- none, gzip, lz4

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE, -- When this result expires
    access_count INTEGER DEFAULT 0,
    last_accessed_at TIMESTAMP WITH TIME ZONE,

    -- Constraints
    CONSTRAINT chk_result_type CHECK (result_type IN ('data', 'error', 'metadata', 'summary')),
    CONSTRAINT chk_compression CHECK (compression IS NULL OR compression IN ('none', 'gzip', 'lz4')),
    -- Ensure at least one storage method is used
    CONSTRAINT chk_result_storage CHECK (
        result_json IS NOT NULL OR
        result_text IS NOT NULL OR
        cache_ref IS NOT NULL OR
        file_path IS NOT NULL
    )
);

-- =============================================================================
-- Report Exports - Track report export operations
-- =============================================================================
CREATE TABLE report_exports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id UUID NOT NULL REFERENCES report_runs(id) ON DELETE CASCADE,
    format VARCHAR(20) NOT NULL, -- csv, excel, pdf, json
    file_name VARCHAR(255) NOT NULL,
    file_path TEXT NOT NULL,
    file_size BIGINT DEFAULT 0,
    mime_type VARCHAR(100),
    generated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    generated_by UUID NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE, -- When file should be cleaned up
    download_count INTEGER DEFAULT 0,
    last_downloaded_at TIMESTAMP WITH TIME ZONE,

    -- Export configuration
    export_options JSONB DEFAULT '{}'::jsonb, -- Format-specific options

    -- Status tracking
    status VARCHAR(20) DEFAULT 'generating', -- generating, ready, expired, failed
    error_message TEXT,

    -- Constraints
    CONSTRAINT chk_export_format CHECK (format IN ('csv', 'excel', 'xlsx', 'pdf', 'json', 'xml')),
    CONSTRAINT chk_export_status CHECK (status IN ('generating', 'ready', 'expired', 'failed', 'deleted'))
);

-- =============================================================================
-- Report Schedules - Automated report execution
-- =============================================================================
CREATE TABLE report_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID NOT NULL, -- References insighthub.reports.id
    name VARCHAR(255) NOT NULL,
    description TEXT,

    -- Schedule configuration
    cron_expression VARCHAR(100) NOT NULL, -- Cron expression for scheduling
    timezone VARCHAR(50) DEFAULT 'UTC',
    is_active BOOLEAN NOT NULL DEFAULT true,

    -- Execution tracking
    next_run_at TIMESTAMP WITH TIME ZONE,
    last_run_at TIMESTAMP WITH TIME ZONE,
    last_run_status VARCHAR(20), -- success, failed, skipped
    last_run_id UUID, -- References report_runs.id
    run_count INTEGER DEFAULT 0,
    failure_count INTEGER DEFAULT 0,

    -- Parameters and configuration
    parameters JSONB DEFAULT '{}'::jsonb, -- Default parameters for scheduled runs
    export_formats TEXT[] DEFAULT '{}', -- Formats to auto-export (csv, excel, pdf)

    -- Notification settings
    notify_on_success BOOLEAN DEFAULT false,
    notify_on_failure BOOLEAN DEFAULT true,
    notification_recipients UUID[], -- User IDs to notify
    notification_channels TEXT[] DEFAULT '{"email"}', -- email, slack, webhook

    -- Retention settings
    max_results_retained INTEGER DEFAULT 30, -- Keep last N results
    auto_cleanup_days INTEGER DEFAULT 90, -- Cleanup results after N days

    -- Audit fields
    created_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_by UUID,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT chk_cron_expression CHECK (length(cron_expression) > 0),
    CONSTRAINT chk_last_run_status CHECK (last_run_status IS NULL OR last_run_status IN ('success', 'failed', 'skipped')),
    CONSTRAINT chk_max_results_retained CHECK (max_results_retained > 0),
    CONSTRAINT chk_auto_cleanup_days CHECK (auto_cleanup_days > 0)
);

-- =============================================================================
-- Report Subscriptions - User subscriptions to reports
-- =============================================================================
CREATE TABLE report_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID NOT NULL, -- References insighthub.reports.id
    user_id UUID NOT NULL,

    -- Subscription settings
    notification_frequency VARCHAR(20) DEFAULT 'immediate', -- immediate, daily, weekly, monthly
    notification_channels TEXT[] DEFAULT '{"email"}', -- email, in_app, webhook

    -- Filters and conditions
    notify_on_changes BOOLEAN DEFAULT true,
    notify_on_errors BOOLEAN DEFAULT true,
    threshold_conditions JSONB DEFAULT '{}'::jsonb, -- Custom notification triggers

    -- Status
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_notification_at TIMESTAMP WITH TIME ZONE,

    -- Audit
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT uq_report_subscription UNIQUE (report_id, user_id),
    CONSTRAINT chk_notification_frequency CHECK (
        notification_frequency IN ('immediate', 'daily', 'weekly', 'monthly', 'never')
    )
);

-- =============================================================================
-- Report Cache - Cache frequently accessed report metadata and results
-- =============================================================================
CREATE TABLE report_cache (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cache_key VARCHAR(255) NOT NULL UNIQUE,
    cache_type VARCHAR(50) NOT NULL, -- result, metadata, query_plan
    report_id UUID, -- References insighthub.reports.id

    -- Cache data
    data JSONB,
    data_text TEXT, -- For larger cache entries
    size_bytes INTEGER DEFAULT 0,

    -- Cache metadata
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    last_accessed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    access_count INTEGER DEFAULT 0,

    -- Cache statistics
    hit_count INTEGER DEFAULT 0,
    miss_count INTEGER DEFAULT 0,

    -- Constraints
    CONSTRAINT chk_cache_type CHECK (cache_type IN ('result', 'metadata', 'query_plan', 'field_values')),
    CONSTRAINT chk_cache_data CHECK (data IS NOT NULL OR data_text IS NOT NULL)
);

-- =============================================================================
-- Report Alerts - Alert configurations for report thresholds
-- =============================================================================
CREATE TABLE report_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID NOT NULL, -- References insighthub.reports.id
    name VARCHAR(255) NOT NULL,
    description TEXT,

    -- Alert conditions
    field_id UUID NOT NULL, -- Field to monitor
    operator VARCHAR(20) NOT NULL, -- >, <, =, !=, etc.
    threshold_value TEXT NOT NULL,
    condition_type VARCHAR(20) DEFAULT 'value', -- value, change, trend

    -- Alert settings
    is_active BOOLEAN NOT NULL DEFAULT true,
    severity VARCHAR(20) DEFAULT 'medium', -- low, medium, high, critical

    -- Notification settings
    notification_channels TEXT[] DEFAULT '{"email"}',
    notification_recipients UUID[],
    notification_message TEXT,

    -- Tracking
    last_triggered_at TIMESTAMP WITH TIME ZONE,
    trigger_count INTEGER DEFAULT 0,

    -- Audit
    created_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_by UUID,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT chk_alert_operator CHECK (operator IN ('>', '<', '>=', '<=', '=', '!=', 'BETWEEN', 'CHANGE_GT', 'CHANGE_LT')),
    CONSTRAINT chk_alert_condition_type CHECK (condition_type IN ('value', 'change', 'trend', 'anomaly')),
    CONSTRAINT chk_alert_severity CHECK (severity IN ('low', 'medium', 'high', 'critical'))
);

-- =============================================================================
-- Create indexes for better performance
-- =============================================================================

-- Report runs indexes
CREATE INDEX idx_report_runs_report_id ON report_runs(report_id);
CREATE INDEX idx_report_runs_run_by ON report_runs(run_by);
CREATE INDEX idx_report_runs_status ON report_runs(status);
CREATE INDEX idx_report_runs_run_at ON report_runs(run_at DESC);
CREATE INDEX idx_report_runs_cache_key ON report_runs(cache_key) WHERE cache_key IS NOT NULL;

-- Report results indexes
CREATE INDEX idx_report_results_run_id ON report_results(run_id);
CREATE INDEX idx_report_results_type ON report_results(result_type);
CREATE INDEX idx_report_results_expires ON report_results(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX idx_report_results_cache_ref ON report_results(cache_ref) WHERE cache_ref IS NOT NULL;

-- Report exports indexes
CREATE INDEX idx_report_exports_run_id ON report_exports(run_id);
CREATE INDEX idx_report_exports_format ON report_exports(format);
CREATE INDEX idx_report_exports_status ON report_exports(status);
CREATE INDEX idx_report_exports_generated_by ON report_exports(generated_by);
CREATE INDEX idx_report_exports_expires ON report_exports(expires_at) WHERE expires_at IS NOT NULL;

-- Report schedules indexes
CREATE INDEX idx_report_schedules_report_id ON report_schedules(report_id);
CREATE INDEX idx_report_schedules_active ON report_schedules(is_active);
CREATE INDEX idx_report_schedules_next_run ON report_schedules(next_run_at) WHERE is_active = true;
CREATE INDEX idx_report_schedules_created_by ON report_schedules(created_by);

-- Report subscriptions indexes
CREATE INDEX idx_report_subscriptions_report_id ON report_subscriptions(report_id);
CREATE INDEX idx_report_subscriptions_user_id ON report_subscriptions(user_id);
CREATE INDEX idx_report_subscriptions_active ON report_subscriptions(is_active);

-- Report cache indexes
CREATE INDEX idx_report_cache_cache_key ON report_cache(cache_key);
CREATE INDEX idx_report_cache_type ON report_cache(cache_type);
CREATE INDEX idx_report_cache_report_id ON report_cache(report_id) WHERE report_id IS NOT NULL;
CREATE INDEX idx_report_cache_expires ON report_cache(expires_at);
CREATE INDEX idx_report_cache_accessed ON report_cache(last_accessed_at DESC);

-- Report alerts indexes
CREATE INDEX idx_report_alerts_report_id ON report_alerts(report_id);
CREATE INDEX idx_report_alerts_active ON report_alerts(is_active);
CREATE INDEX idx_report_alerts_field_id ON report_alerts(field_id);
CREATE INDEX idx_report_alerts_severity ON report_alerts(severity);