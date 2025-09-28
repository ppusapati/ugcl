-- =============================================================================
-- Masters Module - Metadata Schema for Database Structure Management
-- =============================================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =============================================================================
-- Schemas metadata - stores information about database schemas
-- =============================================================================
CREATE TABLE schemas_metadata (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schema_name VARCHAR(63) NOT NULL UNIQUE, -- PostgreSQL schema name limit
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL,
    updated_by VARCHAR(255)
);

-- =============================================================================
-- Tables metadata - stores information about tables within schemas
-- =============================================================================
CREATE TABLE tables_metadata (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schema_id UUID NOT NULL REFERENCES schemas_metadata(id) ON DELETE CASCADE,
    table_name VARCHAR(63) NOT NULL, -- PostgreSQL table name limit
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    supports_import BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL,
    updated_by VARCHAR(255),

    -- Ensure unique table names within schema
    CONSTRAINT uq_tables_schema_table UNIQUE (schema_id, table_name)
);

-- =============================================================================
-- Columns metadata - stores information about columns within tables
-- =============================================================================
CREATE TABLE columns_metadata (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_id UUID NOT NULL REFERENCES tables_metadata(id) ON DELETE CASCADE,
    column_name VARCHAR(63) NOT NULL, -- PostgreSQL column name limit
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    data_type VARCHAR(50) NOT NULL, -- e.g., 'TEXT', 'INTEGER', 'BOOLEAN', 'TIMESTAMP', 'UUID', 'DECIMAL'
    max_length INTEGER, -- for VARCHAR, CHAR types
    is_required BOOLEAN NOT NULL DEFAULT false,
    is_primary_key BOOLEAN NOT NULL DEFAULT false,
    is_unique BOOLEAN NOT NULL DEFAULT false,
    default_value TEXT,
    validation_rules JSONB DEFAULT '[]'::jsonb, -- Array of validation rules
    example_values TEXT[], -- Sample values for user guidance
    is_active BOOLEAN NOT NULL DEFAULT true,
    supports_import BOOLEAN NOT NULL DEFAULT true,
    column_order INTEGER NOT NULL DEFAULT 0, -- Display order
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL,
    updated_by VARCHAR(255),

    -- Ensure unique column names within table
    CONSTRAINT uq_columns_table_column UNIQUE (table_id, column_name),

    -- Check constraints
    CONSTRAINT chk_columns_data_type CHECK (data_type IN (
        'TEXT', 'VARCHAR', 'CHAR', 'INTEGER', 'BIGINT', 'SMALLINT',
        'DECIMAL', 'NUMERIC', 'REAL', 'DOUBLE_PRECISION', 'BOOLEAN',
        'DATE', 'TIME', 'TIMESTAMP', 'TIMESTAMPTZ', 'UUID', 'JSONB', 'JSON'
    )),
    CONSTRAINT chk_columns_max_length CHECK (max_length > 0 OR max_length IS NULL)
);

-- =============================================================================
-- Create indexes for better performance
-- =============================================================================

-- Schemas metadata indexes
CREATE INDEX idx_schemas_metadata_schema_name ON schemas_metadata(schema_name);
CREATE INDEX idx_schemas_metadata_is_active ON schemas_metadata(is_active);
CREATE INDEX idx_schemas_metadata_created_at ON schemas_metadata(created_at DESC);

-- Tables metadata indexes
CREATE INDEX idx_tables_metadata_schema_id ON tables_metadata(schema_id);
CREATE INDEX idx_tables_metadata_table_name ON tables_metadata(table_name);
CREATE INDEX idx_tables_metadata_is_active ON tables_metadata(is_active);
CREATE INDEX idx_tables_metadata_supports_import ON tables_metadata(supports_import);
CREATE INDEX idx_tables_metadata_schema_active ON tables_metadata(schema_id, is_active);

-- Columns metadata indexes
CREATE INDEX idx_columns_metadata_table_id ON columns_metadata(table_id);
CREATE INDEX idx_columns_metadata_column_name ON columns_metadata(column_name);
CREATE INDEX idx_columns_metadata_data_type ON columns_metadata(data_type);
CREATE INDEX idx_columns_metadata_is_required ON columns_metadata(is_required);
CREATE INDEX idx_columns_metadata_supports_import ON columns_metadata(supports_import);
CREATE INDEX idx_columns_metadata_table_order ON columns_metadata(table_id, column_order);
CREATE INDEX idx_columns_metadata_table_active ON columns_metadata(table_id, is_active);

-- =============================================================================
-- Insert sample metadata entries
-- =============================================================================

-- Insert HR schema
INSERT INTO schemas_metadata (schema_name, display_name, description, created_by)
VALUES ('hr', 'Human Resources', 'Human Resources management tables', 'system');

-- Insert inventory schema
INSERT INTO schemas_metadata (schema_name, display_name, description, created_by)
VALUES ('inventory', 'Inventory Management', 'Product and inventory tracking tables', 'system');

-- Get schema IDs for foreign key references
DO $$
DECLARE
    hr_schema_id UUID;
    inventory_schema_id UUID;
    employees_table_id UUID;
    products_table_id UUID;
BEGIN
    -- Get schema IDs
    SELECT id INTO hr_schema_id FROM schemas_metadata WHERE schema_name = 'hr';
    SELECT id INTO inventory_schema_id FROM schemas_metadata WHERE schema_name = 'inventory';

    -- Insert tables
    INSERT INTO tables_metadata (schema_id, table_name, display_name, description, created_by)
    VALUES (hr_schema_id, 'employees', 'Employees', 'Employee information and records', 'system')
    RETURNING id INTO employees_table_id;

    INSERT INTO tables_metadata (schema_id, table_name, display_name, description, created_by)
    VALUES (inventory_schema_id, 'products', 'Products', 'Product catalog and inventory', 'system')
    RETURNING id INTO products_table_id;

    -- Insert columns for hr.employees
    INSERT INTO columns_metadata (table_id, column_name, display_name, description, data_type, is_required, is_primary_key, column_order, created_by)
    VALUES
        (employees_table_id, 'id', 'Employee ID', 'Unique identifier for employee', 'UUID', true, true, 1, 'system'),
        (employees_table_id, 'name', 'Full Name', 'Employee full name', 'VARCHAR', true, false, 2, 'system'),
        (employees_table_id, 'dept', 'Department', 'Employee department', 'VARCHAR', true, false, 3, 'system'),
        (employees_table_id, 'joined_on', 'Join Date', 'Date when employee joined', 'DATE', true, false, 4, 'system');

    -- Insert columns for inventory.products
    INSERT INTO columns_metadata (table_id, column_name, display_name, description, data_type, is_required, is_primary_key, column_order, created_by)
    VALUES
        (products_table_id, 'id', 'Product ID', 'Unique identifier for product', 'UUID', true, true, 1, 'system'),
        (products_table_id, 'sku', 'SKU', 'Stock Keeping Unit code', 'VARCHAR', true, false, 2, 'system'),
        (products_table_id, 'price', 'Price', 'Product price in USD', 'DECIMAL', true, false, 3, 'system');

    -- Update column metadata with additional constraints
    UPDATE columns_metadata
    SET max_length = 255, example_values = ARRAY['John Doe', 'Jane Smith', 'Robert Johnson']
    WHERE table_id = employees_table_id AND column_name = 'name';

    UPDATE columns_metadata
    SET max_length = 100, example_values = ARRAY['Engineering', 'Sales', 'Marketing', 'HR']
    WHERE table_id = employees_table_id AND column_name = 'dept';

    UPDATE columns_metadata
    SET max_length = 50, example_values = ARRAY['PROD-001', 'SERV-001', 'CONS-001']
    WHERE table_id = products_table_id AND column_name = 'sku';

    UPDATE columns_metadata
    SET validation_rules = '[{"type": "min", "value": 0}, {"type": "max", "value": 999999.99}]'::jsonb,
        example_values = ARRAY['29.99', '150.00', '1299.99']
    WHERE table_id = products_table_id AND column_name = 'price';
END $$;

-- =============================================================================
-- Enhanced Metadata for Report Builder Support
-- =============================================================================

-- Table relationships for joins and data modeling
CREATE TABLE table_relationships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255),
    source_table_id UUID NOT NULL REFERENCES tables_metadata(id) ON DELETE CASCADE,
    source_column_id UUID NOT NULL REFERENCES columns_metadata(id) ON DELETE CASCADE,
    target_table_id UUID NOT NULL REFERENCES tables_metadata(id) ON DELETE CASCADE,
    target_column_id UUID NOT NULL REFERENCES columns_metadata(id) ON DELETE CASCADE,
    relationship_type VARCHAR(50) NOT NULL CHECK (relationship_type IN ('one_to_one', 'one_to_many', 'many_to_one', 'many_to_many')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL,
    updated_by VARCHAR(255),
    UNIQUE(source_table_id, source_column_id, target_table_id, target_column_id)
);

-- Field mappings for data transformation and report building
CREATE TABLE field_mappings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    source_column_id UUID NOT NULL REFERENCES columns_metadata(id) ON DELETE CASCADE,
    target_column_id UUID NOT NULL REFERENCES columns_metadata(id) ON DELETE CASCADE,
    mapping_type VARCHAR(50) NOT NULL DEFAULT 'direct' CHECK (mapping_type IN ('direct', 'calculated', 'lookup', 'aggregate')),
    mapping_rules JSONB DEFAULT '{}'::jsonb,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL,
    updated_by VARCHAR(255)
);

-- Business glossary terms for better data understanding
CREATE TABLE business_terms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    term VARCHAR(255) NOT NULL UNIQUE,
    definition TEXT NOT NULL,
    business_context TEXT,
    domain VARCHAR(100),
    owner_user_id UUID,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL,
    updated_by VARCHAR(255)
);

-- Link business terms to columns for semantic context
CREATE TABLE column_business_terms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    column_id UUID NOT NULL REFERENCES columns_metadata(id) ON DELETE CASCADE,
    business_term_id UUID NOT NULL REFERENCES business_terms(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL,
    UNIQUE(column_id, business_term_id)
);

-- Data quality rules and metrics
CREATE TABLE data_quality_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    column_id UUID NOT NULL REFERENCES columns_metadata(id) ON DELETE CASCADE,
    rule_name VARCHAR(255) NOT NULL,
    rule_type VARCHAR(50) NOT NULL CHECK (rule_type IN ('completeness', 'uniqueness', 'validity', 'accuracy', 'consistency')),
    rule_expression TEXT NOT NULL,
    threshold_value DECIMAL(10,4),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL
);

-- Data lineage tracking for impact analysis
CREATE TABLE data_lineage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_table_id UUID REFERENCES tables_metadata(id) ON DELETE CASCADE,
    source_column_id UUID REFERENCES columns_metadata(id) ON DELETE CASCADE,
    target_table_id UUID REFERENCES tables_metadata(id) ON DELETE CASCADE,
    target_column_id UUID REFERENCES columns_metadata(id) ON DELETE CASCADE,
    transformation_type VARCHAR(50) CHECK (transformation_type IN ('copy', 'aggregate', 'calculate', 'join', 'filter')),
    transformation_logic TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL
);

-- =============================================================================
-- Additional indexes for enhanced metadata
-- =============================================================================

-- Table relationships indexes
CREATE INDEX idx_table_relationships_source_table ON table_relationships(source_table_id);
CREATE INDEX idx_table_relationships_target_table ON table_relationships(target_table_id);
CREATE INDEX idx_table_relationships_type ON table_relationships(relationship_type);
CREATE INDEX idx_table_relationships_active ON table_relationships(is_active);

-- Field mappings indexes
CREATE INDEX idx_field_mappings_source_column ON field_mappings(source_column_id);
CREATE INDEX idx_field_mappings_target_column ON field_mappings(target_column_id);
CREATE INDEX idx_field_mappings_type ON field_mappings(mapping_type);
CREATE INDEX idx_field_mappings_active ON field_mappings(is_active);

-- Business terms indexes
CREATE INDEX idx_business_terms_term ON business_terms(term);
CREATE INDEX idx_business_terms_domain ON business_terms(domain);
CREATE INDEX idx_business_terms_active ON business_terms(is_active);

-- Column business terms indexes
CREATE INDEX idx_column_business_terms_column ON column_business_terms(column_id);
CREATE INDEX idx_column_business_terms_term ON column_business_terms(business_term_id);

-- Data quality rules indexes
CREATE INDEX idx_data_quality_rules_column ON data_quality_rules(column_id);
CREATE INDEX idx_data_quality_rules_type ON data_quality_rules(rule_type);
CREATE INDEX idx_data_quality_rules_active ON data_quality_rules(is_active);

-- Data lineage indexes
CREATE INDEX idx_data_lineage_source_table ON data_lineage(source_table_id);
CREATE INDEX idx_data_lineage_target_table ON data_lineage(target_table_id);
CREATE INDEX idx_data_lineage_transformation_type ON data_lineage(transformation_type);