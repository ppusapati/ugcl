-- =============================================================================
-- SEARCH SERVICE DATABASE SCHEMA
-- =============================================================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =============================================================================
-- SEARCH ANALYTICS TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS search_analytics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    search_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255),
    session_id VARCHAR(255),
    query_text TEXT NOT NULL,
    normalized_query TEXT, -- normalized version for analytics
    indices_searched JSONB, -- array of index names
    filters_applied JSONB, -- filters that were applied
    result_count BIGINT DEFAULT 0,
    response_time_ms BIGINT DEFAULT 0,
    page_number INTEGER DEFAULT 1,
    page_size INTEGER DEFAULT 20,

    -- User context
    ip_address INET,
    user_agent TEXT,
    referer TEXT,

    -- Search result interaction
    clicked_results JSONB, -- array of clicked result IDs
    selected_facets JSONB, -- facets that were selected
    scroll_depth INTEGER DEFAULT 0, -- how far user scrolled in results
    time_on_results BIGINT DEFAULT 0, -- time spent viewing results (ms)

    -- Search success metrics
    has_results BOOLEAN DEFAULT false,
    user_refined_search BOOLEAN DEFAULT false, -- did user modify the search
    bounce_rate BOOLEAN DEFAULT false, -- did user leave immediately

    -- Geographic and temporal data
    country_code VARCHAR(2),
    timezone VARCHAR(50),
    search_timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for search_analytics
CREATE INDEX IF NOT EXISTS idx_search_analytics_user_id ON search_analytics(user_id);
CREATE INDEX IF NOT EXISTS idx_search_analytics_query_text ON search_analytics(query_text);
CREATE INDEX IF NOT EXISTS idx_search_analytics_normalized_query ON search_analytics(normalized_query);
CREATE INDEX IF NOT EXISTS idx_search_analytics_search_timestamp ON search_analytics(search_timestamp);
CREATE INDEX IF NOT EXISTS idx_search_analytics_has_results ON search_analytics(has_results);
CREATE INDEX IF NOT EXISTS idx_search_analytics_response_time ON search_analytics(response_time_ms);
CREATE INDEX IF NOT EXISTS idx_search_analytics_result_count ON search_analytics(result_count);
CREATE INDEX IF NOT EXISTS idx_search_analytics_indices ON search_analytics USING GIN(indices_searched);

-- =============================================================================
-- SEARCH SUGGESTIONS TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS search_suggestions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    suggestion_text VARCHAR(500) NOT NULL,
    normalized_text VARCHAR(500) NOT NULL, -- for matching
    category VARCHAR(100), -- 'query', 'entity', 'document_title', etc.
    source VARCHAR(100), -- 'manual', 'auto_generated', 'user_queries'
    popularity_score DECIMAL(10,4) DEFAULT 0,
    click_count BIGINT DEFAULT 0,
    impression_count BIGINT DEFAULT 0,
    conversion_rate DECIMAL(5,4) DEFAULT 0, -- clicks / impressions

    -- Context and metadata
    related_indices JSONB, -- which indices this suggestion is relevant for
    language VARCHAR(10) DEFAULT 'en',
    metadata JSONB,

    -- Quality control
    is_active BOOLEAN DEFAULT true,
    is_approved BOOLEAN DEFAULT true,
    quality_score DECIMAL(5,4) DEFAULT 0,

    -- Lifecycle
    created_by VARCHAR(255),
    last_shown_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for search_suggestions
CREATE INDEX IF NOT EXISTS idx_search_suggestions_normalized_text ON search_suggestions(normalized_text);
CREATE INDEX IF NOT EXISTS idx_search_suggestions_category ON search_suggestions(category);
CREATE INDEX IF NOT EXISTS idx_search_suggestions_popularity_score ON search_suggestions(popularity_score DESC);
CREATE INDEX IF NOT EXISTS idx_search_suggestions_is_active ON search_suggestions(is_active) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_search_suggestions_language ON search_suggestions(language);
CREATE INDEX IF NOT EXISTS idx_search_suggestions_related_indices ON search_suggestions USING GIN(related_indices);

-- =============================================================================
-- SEARCH QUERY HISTORY TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS search_query_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(255) NOT NULL,
    query_text TEXT NOT NULL,
    normalized_query TEXT NOT NULL,
    query_hash VARCHAR(64) NOT NULL, -- for deduplication
    search_type VARCHAR(50) DEFAULT 'text', -- 'text', 'faceted', 'advanced'

    -- Query context
    indices_searched JSONB,
    filters JSONB,
    sort_criteria JSONB,

    -- Results and interaction
    result_count BIGINT DEFAULT 0,
    selected_results JSONB, -- IDs of results user clicked on
    satisfaction_score INTEGER, -- 1-5 if user provides feedback

    -- Frequency tracking
    query_frequency INTEGER DEFAULT 1, -- how many times this exact query
    last_executed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    first_executed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(user_id, query_hash)
);

-- Indexes for search_query_history
CREATE INDEX IF NOT EXISTS idx_search_query_history_user_id ON search_query_history(user_id);
CREATE INDEX IF NOT EXISTS idx_search_query_history_normalized_query ON search_query_history(normalized_query);
CREATE INDEX IF NOT EXISTS idx_search_query_history_query_frequency ON search_query_history(query_frequency DESC);
CREATE INDEX IF NOT EXISTS idx_search_query_history_last_executed ON search_query_history(last_executed_at DESC);

-- =============================================================================
-- SEARCH INDEX METADATA TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS search_index_metadata (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    index_name VARCHAR(255) UNIQUE NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    index_type VARCHAR(100), -- 'documents', 'forms', 'users', etc.

    -- Configuration
    elasticsearch_settings JSONB,
    elasticsearch_mappings JSONB,
    indexing_config JSONB,

    -- Statistics
    document_count BIGINT DEFAULT 0,
    index_size_bytes BIGINT DEFAULT 0,
    last_indexed_at TIMESTAMP WITH TIME ZONE,
    last_optimized_at TIMESTAMP WITH TIME ZONE,

    -- Status and health
    status VARCHAR(50) DEFAULT 'active', -- 'active', 'inactive', 'error', 'rebuilding'
    health_score DECIMAL(5,4) DEFAULT 1.0,
    error_message TEXT,

    -- Access control
    is_public BOOLEAN DEFAULT false,
    allowed_roles JSONB, -- array of role names that can search this index
    created_by VARCHAR(255) NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for search_index_metadata
CREATE INDEX IF NOT EXISTS idx_search_index_metadata_index_name ON search_index_metadata(index_name);
CREATE INDEX IF NOT EXISTS idx_search_index_metadata_index_type ON search_index_metadata(index_type);
CREATE INDEX IF NOT EXISTS idx_search_index_metadata_status ON search_index_metadata(status);
CREATE INDEX IF NOT EXISTS idx_search_index_metadata_is_public ON search_index_metadata(is_public);

-- =============================================================================
-- SEARCH JOBS TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS search_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    job_type VARCHAR(100) NOT NULL, -- 'reindex', 'optimize', 'backup', 'migrate'
    index_name VARCHAR(255),
    status VARCHAR(50) DEFAULT 'pending', -- 'pending', 'running', 'completed', 'failed'
    priority INTEGER DEFAULT 0,

    -- Job configuration
    job_config JSONB,

    -- Progress tracking
    progress_percentage INTEGER DEFAULT 0,
    records_processed BIGINT DEFAULT 0,
    total_records BIGINT DEFAULT 0,

    -- Timing
    scheduled_at TIMESTAMP WITH TIME ZONE,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    estimated_completion TIMESTAMP WITH TIME ZONE,

    -- Results and errors
    result_data JSONB,
    error_message TEXT,
    error_count INTEGER DEFAULT 0,

    -- Metadata
    created_by VARCHAR(255),
    tags JSONB,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for search_jobs
CREATE INDEX IF NOT EXISTS idx_search_jobs_job_type ON search_jobs(job_type);
CREATE INDEX IF NOT EXISTS idx_search_jobs_index_name ON search_jobs(index_name);
CREATE INDEX IF NOT EXISTS idx_search_jobs_status ON search_jobs(status);
CREATE INDEX IF NOT EXISTS idx_search_jobs_priority ON search_jobs(priority DESC);
CREATE INDEX IF NOT EXISTS idx_search_jobs_scheduled_at ON search_jobs(scheduled_at);

-- =============================================================================
-- SEARCH FACET CONFIGURATIONS TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS search_facet_configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    index_name VARCHAR(255) NOT NULL,
    facet_name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    field_name VARCHAR(255) NOT NULL,
    facet_type VARCHAR(50) NOT NULL, -- 'terms', 'range', 'date_histogram', 'nested'

    -- Configuration
    facet_config JSONB, -- type-specific configuration
    sort_order INTEGER DEFAULT 0,
    max_buckets INTEGER DEFAULT 20,
    min_doc_count INTEGER DEFAULT 1,

    -- Display options
    is_enabled BOOLEAN DEFAULT true,
    is_collapsible BOOLEAN DEFAULT true,
    default_expanded BOOLEAN DEFAULT true,
    show_counts BOOLEAN DEFAULT true,

    -- Access control
    required_roles JSONB,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(index_name, facet_name)
);

-- Indexes for search_facet_configs
CREATE INDEX IF NOT EXISTS idx_search_facet_configs_index_name ON search_facet_configs(index_name);
CREATE INDEX IF NOT EXISTS idx_search_facet_configs_is_enabled ON search_facet_configs(is_enabled) WHERE is_enabled = true;
CREATE INDEX IF NOT EXISTS idx_search_facet_configs_sort_order ON search_facet_configs(sort_order);

-- =============================================================================
-- SEARCH RESULT CLICK TRACKING TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS search_result_clicks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    search_analytics_id UUID REFERENCES search_analytics(id) ON DELETE CASCADE,
    result_id VARCHAR(255) NOT NULL, -- the document/item ID that was clicked
    result_index VARCHAR(255) NOT NULL, -- which Elasticsearch index
    result_position INTEGER NOT NULL, -- position in search results (1-based)
    result_page INTEGER DEFAULT 1,

    -- Click context
    click_timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    time_to_click BIGINT, -- milliseconds from search to click

    -- User session
    user_id VARCHAR(255),
    session_id VARCHAR(255),

    -- Result metadata at time of click
    result_title TEXT,
    result_score DECIMAL(10,6),

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for search_result_clicks
CREATE INDEX IF NOT EXISTS idx_search_result_clicks_analytics_id ON search_result_clicks(search_analytics_id);
CREATE INDEX IF NOT EXISTS idx_search_result_clicks_result_id ON search_result_clicks(result_id);
CREATE INDEX IF NOT EXISTS idx_search_result_clicks_user_id ON search_result_clicks(user_id);
CREATE INDEX IF NOT EXISTS idx_search_result_clicks_result_position ON search_result_clicks(result_position);
CREATE INDEX IF NOT EXISTS idx_search_result_clicks_click_timestamp ON search_result_clicks(click_timestamp);

-- =============================================================================
-- SEARCH SAVED QUERIES TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS search_saved_queries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(255) NOT NULL,
    query_name VARCHAR(255) NOT NULL,
    description TEXT,

    -- Query definition
    query_text TEXT NOT NULL,
    indices JSONB,
    filters JSONB,
    sort_criteria JSONB,
    facets JSONB,

    -- Notifications
    enable_alerts BOOLEAN DEFAULT false,
    alert_frequency VARCHAR(50), -- 'immediate', 'daily', 'weekly'
    last_alert_sent TIMESTAMP WITH TIME ZONE,

    -- Usage tracking
    execution_count INTEGER DEFAULT 0,
    last_executed_at TIMESTAMP WITH TIME ZONE,

    -- Sharing
    is_public BOOLEAN DEFAULT false,
    shared_with JSONB, -- array of user IDs

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(user_id, query_name)
);

-- Indexes for search_saved_queries
CREATE INDEX IF NOT EXISTS idx_search_saved_queries_user_id ON search_saved_queries(user_id);
CREATE INDEX IF NOT EXISTS idx_search_saved_queries_is_public ON search_saved_queries(is_public) WHERE is_public = true;
CREATE INDEX IF NOT EXISTS idx_search_saved_queries_enable_alerts ON search_saved_queries(enable_alerts) WHERE enable_alerts = true;

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
CREATE TRIGGER update_search_suggestions_updated_at BEFORE UPDATE ON search_suggestions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_search_query_history_updated_at BEFORE UPDATE ON search_query_history
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_search_index_metadata_updated_at BEFORE UPDATE ON search_index_metadata
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_search_jobs_updated_at BEFORE UPDATE ON search_jobs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_search_facet_configs_updated_at BEFORE UPDATE ON search_facet_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_search_saved_queries_updated_at BEFORE UPDATE ON search_saved_queries
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =============================================================================
-- FUNCTIONS
-- =============================================================================

-- Function to normalize search queries for analytics
CREATE OR REPLACE FUNCTION normalize_search_query(query_text TEXT)
RETURNS TEXT AS $$
BEGIN
    -- Convert to lowercase, trim whitespace, remove extra spaces
    RETURN TRIM(REGEXP_REPLACE(LOWER(query_text), '\s+', ' ', 'g'));
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Function to get popular search queries
CREATE OR REPLACE FUNCTION get_popular_queries(days_back INTEGER DEFAULT 30, limit_count INTEGER DEFAULT 10)
RETURNS TABLE(
    query_text TEXT,
    search_count BIGINT,
    avg_result_count DECIMAL(10,2),
    avg_response_time DECIMAL(10,2)
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        sa.normalized_query,
        COUNT(*)::BIGINT as search_count,
        AVG(sa.result_count)::DECIMAL(10,2) as avg_result_count,
        AVG(sa.response_time_ms)::DECIMAL(10,2) as avg_response_time
    FROM search_analytics sa
    WHERE sa.search_timestamp >= NOW() - INTERVAL '%s days', days_back
    AND sa.normalized_query IS NOT NULL
    AND LENGTH(sa.normalized_query) > 0
    GROUP BY sa.normalized_query
    HAVING COUNT(*) > 1 -- at least 2 searches
    ORDER BY search_count DESC, avg_result_count DESC
    LIMIT limit_count;
END;
$$ LANGUAGE plpgsql;

-- Function to get search trends over time
CREATE OR REPLACE FUNCTION get_search_trends(days_back INTEGER DEFAULT 30)
RETURNS TABLE(
    search_date DATE,
    total_searches BIGINT,
    unique_users BIGINT,
    avg_response_time DECIMAL(10,2),
    zero_result_searches BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        sa.search_timestamp::DATE as search_date,
        COUNT(*)::BIGINT as total_searches,
        COUNT(DISTINCT sa.user_id)::BIGINT as unique_users,
        AVG(sa.response_time_ms)::DECIMAL(10,2) as avg_response_time,
        COUNT(*) FILTER (WHERE sa.result_count = 0)::BIGINT as zero_result_searches
    FROM search_analytics sa
    WHERE sa.search_timestamp >= NOW() - INTERVAL '%s days', days_back
    GROUP BY sa.search_timestamp::DATE
    ORDER BY search_date DESC;
END;
$$ LANGUAGE plpgsql;

-- Function to update suggestion popularity scores
CREATE OR REPLACE FUNCTION update_suggestion_scores()
RETURNS void AS $$
BEGIN
    -- Update popularity scores based on recent usage
    UPDATE search_suggestions
    SET popularity_score = CASE
        WHEN impression_count > 0 THEN
            (click_count * 10.0 / impression_count) +
            (click_count * 0.1) +
            (CASE WHEN last_shown_at > NOW() - INTERVAL '7 days' THEN 5.0 ELSE 0 END)
        ELSE 0
    END,
    conversion_rate = CASE
        WHEN impression_count > 0 THEN click_count::DECIMAL / impression_count
        ELSE 0
    END,
    updated_at = NOW()
    WHERE is_active = true;
END;
$$ LANGUAGE plpgsql;

-- =============================================================================
-- VIEWS
-- =============================================================================

-- View for search performance metrics
CREATE OR REPLACE VIEW search_performance_metrics AS
SELECT
    DATE_TRUNC('hour', search_timestamp) as time_bucket,
    COUNT(*) as total_searches,
    COUNT(DISTINCT user_id) as unique_users,
    AVG(response_time_ms) as avg_response_time,
    PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY response_time_ms) as median_response_time,
    PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY response_time_ms) as p95_response_time,
    AVG(result_count) as avg_result_count,
    COUNT(*) FILTER (WHERE result_count = 0) as zero_result_searches,
    COUNT(*) FILTER (WHERE result_count = 0)::DECIMAL / COUNT(*) as zero_result_rate
FROM search_analytics
WHERE search_timestamp >= NOW() - INTERVAL '7 days'
GROUP BY DATE_TRUNC('hour', search_timestamp)
ORDER BY time_bucket DESC;

-- View for top search queries
CREATE OR REPLACE VIEW top_search_queries AS
SELECT
    normalized_query,
    COUNT(*) as search_count,
    COUNT(DISTINCT user_id) as unique_users,
    AVG(result_count) as avg_results,
    AVG(response_time_ms) as avg_response_time,
    MAX(search_timestamp) as last_searched
FROM search_analytics
WHERE search_timestamp >= NOW() - INTERVAL '30 days'
AND normalized_query IS NOT NULL
AND LENGTH(normalized_query) > 0
GROUP BY normalized_query
HAVING COUNT(*) >= 5
ORDER BY search_count DESC
LIMIT 100;