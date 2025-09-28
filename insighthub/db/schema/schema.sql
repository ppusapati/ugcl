-- =============================================================================
-- InsightHub - Report Builder Service Database Schema
-- =============================================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =============================================================================
-- Data Sources - Define available data sources from Masters
-- =============================================================================
CREATE TABLE data_sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    connection_string TEXT, -- Connection details if needed
    schema_id UUID NOT NULL, -- References masters.schemas_metadata.id
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_by UUID,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- =============================================================================
-- Reports - Main report definitions
-- =============================================================================
CREATE TABLE reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100), -- e.g., 'financial', 'operational', 'hr'
    tags TEXT[], -- Array of tags for organization
    is_template BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,

    -- Data source configuration
    primary_data_source_id UUID REFERENCES data_sources(id),

    created_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_by UUID,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    version INTEGER NOT NULL DEFAULT 1, -- For version control

    -- Metadata for the report
    metadata JSONB DEFAULT '{}'::jsonb, -- Additional metadata like data source info

    -- Constraints
    CONSTRAINT chk_reports_name_length CHECK (length(name) >= 3)
);

-- =============================================================================
-- Report Tables - Tables used in this report from various data sources
-- =============================================================================
CREATE TABLE report_tables (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID NOT NULL REFERENCES reports(id) ON DELETE CASCADE,
    table_id UUID NOT NULL, -- References masters.tables_metadata.id
    table_alias VARCHAR(255), -- Alias for this table in the report
    data_source_id UUID NOT NULL REFERENCES data_sources(id),
    is_primary BOOLEAN NOT NULL DEFAULT false, -- Is this the primary table
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT uq_report_table_alias UNIQUE (report_id, table_alias),
    -- Ensure only one primary table per report
    CONSTRAINT uq_report_primary_table EXCLUDE (report_id WITH =) WHERE (is_primary = true)
);

-- =============================================================================
-- Report Fields - Fields selected for the report
-- =============================================================================
CREATE TABLE report_fields (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID NOT NULL REFERENCES reports(id) ON DELETE CASCADE,
    column_id UUID NOT NULL, -- References masters.columns_metadata.id
    report_table_id UUID REFERENCES report_tables(id), -- Link to specific table in report
    field_alias VARCHAR(255), -- Custom alias for the field in report
    aggregate_function VARCHAR(50), -- SUM, COUNT, AVG, MAX, MIN, etc.
    order_index INTEGER NOT NULL, -- Order in the report
    is_visible BOOLEAN NOT NULL DEFAULT true,
    formatting_rules JSONB DEFAULT '{}'::jsonb, -- Number format, date format, etc.
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT uq_report_field_order UNIQUE (report_id, order_index),
    CONSTRAINT chk_aggregate_function CHECK (
        aggregate_function IS NULL OR
        aggregate_function IN ('SUM', 'COUNT', 'AVG', 'MAX', 'MIN', 'COUNT_DISTINCT', 'DISTINCT_COUNT')
    )
);

-- =============================================================================
-- Report Filters - Filtering conditions for the report
-- =============================================================================
CREATE TABLE report_filters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID NOT NULL REFERENCES reports(id) ON DELETE CASCADE,
    column_id UUID NOT NULL, -- References masters.columns_metadata.id
    report_table_id UUID REFERENCES report_tables(id), -- Link to specific table in report
    operator VARCHAR(20) NOT NULL, -- =, !=, >, <, >=, <=, LIKE, IN, BETWEEN
    value TEXT, -- Filter value (can be JSON for complex values)
    value_type VARCHAR(20) NOT NULL DEFAULT 'static', -- static, parameter, calculated
    logical_operator VARCHAR(5) DEFAULT 'AND', -- AND, OR
    group_index INTEGER DEFAULT 0, -- For grouping conditions
    order_index INTEGER NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT chk_filter_operator CHECK (
        operator IN ('=', '!=', '>', '<', '>=', '<=', 'LIKE', 'ILIKE', 'IN', 'NOT_IN', 'BETWEEN', 'IS_NULL', 'IS_NOT_NULL')
    ),
    CONSTRAINT chk_logical_operator CHECK (logical_operator IN ('AND', 'OR')),
    CONSTRAINT chk_value_type CHECK (value_type IN ('static', 'parameter', 'calculated'))
);

-- =============================================================================
-- Report Groups - Grouping fields for the report
-- =============================================================================
CREATE TABLE report_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID NOT NULL REFERENCES reports(id) ON DELETE CASCADE,
    column_id UUID NOT NULL, -- References masters.columns_metadata.id
    report_table_id UUID REFERENCES report_tables(id), -- Link to specific table in report
    order_index INTEGER NOT NULL,
    group_type VARCHAR(20) DEFAULT 'standard', -- standard, date_part (for date grouping)
    date_part VARCHAR(20), -- year, month, day, quarter (when group_type = 'date_part')
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT uq_report_group_order UNIQUE (report_id, order_index),
    CONSTRAINT chk_group_type CHECK (group_type IN ('standard', 'date_part')),
    CONSTRAINT chk_date_part CHECK (
        (group_type != 'date_part') OR
        (date_part IN ('year', 'quarter', 'month', 'week', 'day'))
    )
);

-- =============================================================================
-- Report Sorts - Sorting configuration for the report
-- =============================================================================
CREATE TABLE report_sorts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID NOT NULL REFERENCES reports(id) ON DELETE CASCADE,
    column_id UUID, -- References masters.columns_metadata.id
    report_field_id UUID, -- References report_fields.id (for calculated fields)
    report_table_id UUID REFERENCES report_tables(id), -- Link to specific table in report
    sort_direction VARCHAR(5) NOT NULL DEFAULT 'ASC', -- ASC, DESC
    order_index INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT uq_report_sort_order UNIQUE (report_id, order_index),
    CONSTRAINT chk_sort_direction CHECK (sort_direction IN ('ASC', 'DESC')),
    -- Must reference either a column or a report field, not both
    CONSTRAINT chk_sort_reference CHECK (
        (column_id IS NOT NULL AND report_field_id IS NULL) OR
        (column_id IS NULL AND report_field_id IS NOT NULL)
    )
);

-- =============================================================================
-- Report Charts - Chart configurations for the report
-- =============================================================================
CREATE TABLE report_charts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID NOT NULL REFERENCES reports(id) ON DELETE CASCADE,
    chart_type VARCHAR(50) NOT NULL, -- bar, line, pie, scatter, etc.
    title VARCHAR(255),
    x_axis_field_id UUID, -- References report_fields.id
    y_axis_field_id UUID, -- References report_fields.id
    series_field_id UUID, -- References report_fields.id (for grouping)
    config_json JSONB NOT NULL DEFAULT '{}'::jsonb, -- ECharts configuration
    order_index INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT chk_chart_type CHECK (
        chart_type IN ('bar', 'line', 'pie', 'doughnut', 'scatter', 'area', 'radar', 'funnel', 'gauge', 'heatmap')
    )
);

-- =============================================================================
-- Report Permissions - Access control for reports
-- =============================================================================
CREATE TABLE report_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID NOT NULL REFERENCES reports(id) ON DELETE CASCADE,
    principal_type VARCHAR(20) NOT NULL, -- user, role, team
    principal_id UUID NOT NULL, -- ID of user, role, or team
    access_level VARCHAR(20) NOT NULL, -- read, write, admin
    granted_by UUID NOT NULL, -- User who granted the permission
    granted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE, -- Optional expiration
    is_active BOOLEAN NOT NULL DEFAULT true,

    -- Constraints
    CONSTRAINT uq_report_permission UNIQUE (report_id, principal_type, principal_id),
    CONSTRAINT chk_principal_type CHECK (principal_type IN ('user', 'role', 'team')),
    CONSTRAINT chk_access_level CHECK (access_level IN ('read', 'write', 'admin'))
);

-- =============================================================================
-- Report Parameters - Dynamic parameters for reports
-- =============================================================================
CREATE TABLE report_parameters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID NOT NULL REFERENCES reports(id) ON DELETE CASCADE,
    parameter_name VARCHAR(100) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    data_type VARCHAR(50) NOT NULL, -- text, number, date, boolean, list
    default_value TEXT,
    is_required BOOLEAN NOT NULL DEFAULT false,
    validation_rules JSONB DEFAULT '{}'::jsonb,
    options JSONB DEFAULT '[]'::jsonb, -- For list type parameters
    order_index INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT uq_report_parameter_name UNIQUE (report_id, parameter_name),
    CONSTRAINT uq_report_parameter_order UNIQUE (report_id, order_index),
    CONSTRAINT chk_parameter_data_type CHECK (
        data_type IN ('text', 'number', 'integer', 'decimal', 'date', 'datetime', 'boolean', 'list', 'multiselect')
    )
);

-- =============================================================================
-- Report Joins - Table joins configuration
-- =============================================================================
CREATE TABLE report_joins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID NOT NULL REFERENCES reports(id) ON DELETE CASCADE,
    source_table_id UUID NOT NULL, -- References masters.tables_metadata.id
    target_table_id UUID NOT NULL, -- References masters.tables_metadata.id
    join_type VARCHAR(20) NOT NULL DEFAULT 'INNER', -- INNER, LEFT, RIGHT, FULL
    source_field_id UUID NOT NULL, -- References masters.columns_metadata.id
    target_field_id UUID NOT NULL, -- References masters.columns_metadata.id
    order_index INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT chk_join_type CHECK (join_type IN ('INNER', 'LEFT', 'RIGHT', 'FULL', 'CROSS'))
);

-- =============================================================================
-- Create indexes for better performance
-- =============================================================================

-- Data sources indexes
CREATE INDEX idx_data_sources_schema_id ON data_sources(schema_id);
CREATE INDEX idx_data_sources_is_active ON data_sources(is_active);
CREATE INDEX idx_data_sources_name ON data_sources(name);

-- Reports indexes
CREATE INDEX idx_reports_created_by ON reports(created_by);
CREATE INDEX idx_reports_category ON reports(category) WHERE category IS NOT NULL;
CREATE INDEX idx_reports_is_active ON reports(is_active);
CREATE INDEX idx_reports_created_at ON reports(created_at DESC);
CREATE INDEX idx_reports_tags ON reports USING GIN(tags) WHERE tags IS NOT NULL;
CREATE INDEX idx_reports_primary_data_source ON reports(primary_data_source_id);

-- Report tables indexes
CREATE INDEX idx_report_tables_report_id ON report_tables(report_id);
CREATE INDEX idx_report_tables_table_id ON report_tables(table_id);
CREATE INDEX idx_report_tables_data_source_id ON report_tables(data_source_id);
CREATE INDEX idx_report_tables_primary ON report_tables(report_id, is_primary) WHERE is_primary = true;

-- Report fields indexes
CREATE INDEX idx_report_fields_report_id ON report_fields(report_id);
CREATE INDEX idx_report_fields_column_id ON report_fields(column_id);
CREATE INDEX idx_report_fields_report_table_id ON report_fields(report_table_id);
CREATE INDEX idx_report_fields_order ON report_fields(report_id, order_index);

-- Report filters indexes
CREATE INDEX idx_report_filters_report_id ON report_filters(report_id);
CREATE INDEX idx_report_filters_column_id ON report_filters(column_id);
CREATE INDEX idx_report_filters_report_table_id ON report_filters(report_table_id);
CREATE INDEX idx_report_filters_active ON report_filters(report_id, is_active);

-- Report groups indexes
CREATE INDEX idx_report_groups_report_id ON report_groups(report_id);
CREATE INDEX idx_report_groups_column_id ON report_groups(column_id);
CREATE INDEX idx_report_groups_report_table_id ON report_groups(report_table_id);

-- Report sorts indexes
CREATE INDEX idx_report_sorts_report_id ON report_sorts(report_id);
CREATE INDEX idx_report_sorts_column_id ON report_sorts(column_id);
CREATE INDEX idx_report_sorts_report_field_id ON report_sorts(report_field_id);
CREATE INDEX idx_report_sorts_report_table_id ON report_sorts(report_table_id);

-- Report charts indexes
CREATE INDEX idx_report_charts_report_id ON report_charts(report_id);
CREATE INDEX idx_report_charts_active ON report_charts(report_id, is_active);

-- Report permissions indexes
CREATE INDEX idx_report_permissions_report_id ON report_permissions(report_id);
CREATE INDEX idx_report_permissions_principal ON report_permissions(principal_type, principal_id);
CREATE INDEX idx_report_permissions_active ON report_permissions(is_active);

-- Report parameters indexes
CREATE INDEX idx_report_parameters_report_id ON report_parameters(report_id);
CREATE INDEX idx_report_parameters_order ON report_parameters(report_id, order_index);

-- Report joins indexes
CREATE INDEX idx_report_joins_report_id ON report_joins(report_id);
CREATE INDEX idx_report_joins_tables ON report_joins(source_table_id, target_table_id);