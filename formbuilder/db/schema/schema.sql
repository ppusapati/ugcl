-- =============================================================================
-- schema.sql - Complete database schema for form builder
-- =============================================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =============================================================================
-- Forms table - stores form definitions
-- =============================================================================
CREATE TABLE forms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    version VARCHAR(50) NOT NULL DEFAULT '1.0.0',
    created_by VARCHAR(255) NOT NULL,
    allowed_roles JSONB DEFAULT '[]'::jsonb,
    audit BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    table_name VARCHAR(255),
    module VARCHAR(255),
    schema_version INTEGER DEFAULT 1,
    core_fields JSONB DEFAULT '[]'::jsonb,
    steps JSONB NOT NULL DEFAULT '[]'::jsonb,
    dependencies JSONB DEFAULT '[]'::jsonb,
    cross_field_validations JSONB DEFAULT '[]'::jsonb,
    workflow_id UUID
);

-- =============================================================================
-- Form instances table - stores form submissions/instances
-- =============================================================================
CREATE TABLE form_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    form_id UUID NOT NULL REFERENCES forms(id) ON DELETE CASCADE,
    current_state VARCHAR(100) NOT NULL DEFAULT 'draft',
    field_values JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_by VARCHAR(255) NOT NULL,
    assigned_to VARCHAR(255),
    metadata JSONB DEFAULT '{}'::jsonb
);

-- =============================================================================
-- Audit logs table - tracks all changes to form instances
-- =============================================================================
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id UUID NOT NULL REFERENCES form_instances(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    action VARCHAR(100) NOT NULL,
    from_state VARCHAR(100),
    to_state VARCHAR(100),
    changes JSONB DEFAULT '{}'::jsonb,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    ip_address INET,
    user_agent TEXT
);

-- =============================================================================
-- Attachments table - stores file attachments for form fields
-- =============================================================================
CREATE TABLE attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id UUID NOT NULL REFERENCES form_instances(id) ON DELETE CASCADE,
    field_id VARCHAR(255) NOT NULL,
    filename VARCHAR(255) NOT NULL,
    mime_type VARCHAR(100),
    size BIGINT CHECK (size >= 0),
    storage_path TEXT NOT NULL,
    uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    uploaded_by VARCHAR(255) NOT NULL
);

-- =============================================================================
-- Comments table - stores comments on form instances
-- =============================================================================
CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id UUID NOT NULL REFERENCES form_instances(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    text TEXT NOT NULL CHECK (length(trim(text)) > 0),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    internal BOOLEAN DEFAULT false
);

-- =============================================================================
-- Workflows table - defines workflow states and transitions
-- =============================================================================
CREATE TABLE workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    initial_state VARCHAR(100) NOT NULL,
    states JSONB NOT NULL DEFAULT '[]'::jsonb,
    global_transitions JSONB DEFAULT '[]'::jsonb,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- =============================================================================
-- SLA rules table - defines SLA rules for workflow states
-- =============================================================================
CREATE TABLE sla_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL CHECK (length(trim(name)) > 0),
    state VARCHAR(100) NOT NULL,
    duration JSONB NOT NULL,
    escalation_levels JSONB DEFAULT '[]'::jsonb,
    active BOOLEAN DEFAULT true,
    applicable_roles JSONB DEFAULT '[]'::jsonb,
    condition TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- =============================================================================
-- Escalations table - defines escalation rules
-- =============================================================================
CREATE TABLE escalations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL CHECK (length(trim(name)) > 0),
    trigger_condition TEXT,
    from_states JSONB DEFAULT '[]'::jsonb,
    to_state VARCHAR(100) NOT NULL,
    notify_roles JSONB DEFAULT '[]'::jsonb,
    escalation_message TEXT,
    auto_escalate BOOLEAN DEFAULT false,
    after_duration JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- =============================================================================
-- Add foreign key constraints
-- =============================================================================
ALTER TABLE forms ADD CONSTRAINT fk_forms_workflow_id 
    FOREIGN KEY (workflow_id) REFERENCES workflows(id) ON DELETE SET NULL;

-- =============================================================================
-- Create indexes for better performance
-- =============================================================================

-- Forms indexes
CREATE INDEX idx_forms_created_by ON forms(created_by);
CREATE INDEX idx_forms_workflow_id ON forms(workflow_id);
CREATE INDEX idx_forms_title ON forms(title);
CREATE INDEX idx_forms_version ON forms(version);
CREATE INDEX idx_forms_created_by_created_at ON forms(created_by, created_at DESC);
CREATE INDEX idx_forms_deleted_at ON forms(deleted_at);

-- Form instances indexes
CREATE INDEX idx_form_instances_form_id ON form_instances(form_id);
CREATE INDEX idx_form_instances_created_by ON form_instances(created_by);
CREATE INDEX idx_form_instances_assigned_to ON form_instances(assigned_to);
CREATE INDEX idx_form_instances_current_state ON form_instances(current_state);
CREATE INDEX idx_form_instances_form_state ON form_instances(form_id, current_state);
CREATE INDEX idx_form_instances_state_created ON form_instances(current_state, created_at);
CREATE INDEX idx_form_instances_deleted_at ON form_instances(deleted_at);

-- Audit logs indexes
CREATE INDEX idx_audit_logs_instance_id ON audit_logs(instance_id);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp);
CREATE INDEX idx_audit_logs_instance_timestamp ON audit_logs(instance_id, timestamp DESC);

-- Attachments indexes
CREATE INDEX idx_attachments_instance_id ON attachments(instance_id);
CREATE INDEX idx_attachments_field_id ON attachments(field_id);

-- Comments indexes
CREATE INDEX idx_comments_instance_id ON comments(instance_id);
CREATE INDEX idx_comments_created_at ON comments(created_at);

-- Workflows indexes
CREATE INDEX idx_workflows_initial_state ON workflows(initial_state);
CREATE INDEX idx_workflows_deleted_at ON workflows(deleted_at);

-- SLA rules indexes
CREATE INDEX idx_sla_rules_state ON sla_rules(state);
CREATE INDEX idx_sla_rules_active ON sla_rules(active);
CREATE INDEX idx_sla_rules_deleted_at ON sla_rules(deleted_at);

-- Escalations indexes
CREATE INDEX idx_escalations_from_states ON escalations USING GIN(from_states);
CREATE INDEX idx_escalations_auto_escalate ON escalations(auto_escalate);
CREATE INDEX idx_escalations_deleted_at ON escalations(deleted_at);

-- =============================================================================
-- Check constraints for data integrity
-- =============================================================================
ALTER TABLE forms ADD CONSTRAINT chk_forms_version_format 
    CHECK (version ~ '^[0-9]+\.[0-9]+\.[0-9]+$');

ALTER TABLE forms ADD CONSTRAINT chk_forms_schema_version_positive 
    CHECK (schema_version > 0);

ALTER TABLE form_instances ADD CONSTRAINT chk_instances_current_state_not_empty 
    CHECK (length(trim(current_state)) > 0);

ALTER TABLE audit_logs ADD CONSTRAINT chk_audit_action_not_empty 
    CHECK (length(trim(action)) > 0);

ALTER TABLE attachments ADD CONSTRAINT chk_attachments_filename_not_empty 
    CHECK (length(trim(filename)) > 0);