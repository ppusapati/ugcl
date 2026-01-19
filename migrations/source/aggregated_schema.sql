-- Auto-Generated Schema

-- From identity\db\sqlc\schema.sql (identity\db\sqlc\schema.sql)
/* ===========================================================
   EXTENSIONS
   ===========================================================*/
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

/* ===========================================================
   ROLES  (hierarchy enabled)
   ===========================================================*/
CREATE TABLE IF NOT EXISTS roles (
    id          BIGSERIAL  PRIMARY KEY,
    uuid        UUID  UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    parent_id   BIGINT  NULL REFERENCES roles(id) ON DELETE SET NULL ON UPDATE CASCADE,
    name        TEXT  NOT NULL,
    is_preserved BOOLEAN NOT NULL DEFAULT FALSE,
    metadata    JSONB,
    created_at  TIMESTAMP NOT NULL DEFAULT now(),
    updated_at  TIMESTAMP NOT NULL DEFAULT now(),
    UNIQUE(name)  -- one role name only once in single-tenant mode
); 

CREATE INDEX IF NOT EXISTS ix_roles_parent   ON roles(parent_id);

/* ===========================================================
   USERS
   ===========================================================*/
CREATE TABLE IF NOT EXISTS users (
    id              BIGSERIAL PRIMARY KEY,
    uuid            UUID  UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    username        TEXT,
    normalized_username TEXT,
    fullname        TEXT,
    phone           TEXT,
    phone_confirmed BOOLEAN DEFAULT false,
    email           TEXT,
    normalized_email TEXT,
    email_confirmed BOOLEAN DEFAULT false,
    password        TEXT,
    password_hash   TEXT NOT NULL,
    gender          INTEGER,
    avatar          BYTEA,
    two_factor_enabled BOOLEAN DEFAULT false,
    salt            BYTEA,
    two_factor_secret TEXT,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    roles_cache     TEXT[],          -- renamed from roles to avoid junction clash
    created_at      TIMESTAMP NOT NULL DEFAULT now(),
    updated_at      TIMESTAMP NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMP
);

CREATE INDEX IF NOT EXISTS ix_users_username ON users(normalized_username);
CREATE INDEX IF NOT EXISTS ix_users_email    ON users(normalized_email);

/* ===========================================================
   USER ↔ ROLE  (junction)
   ===========================================================*/
CREATE TABLE IF NOT EXISTS user_roles (
    user_id BIGINT NOT NULL,
    role_id   BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id)   REFERENCES roles(id)   ON DELETE CASCADE
);

/* ===========================================================
   PERMISSION DEFINITIONS  (templates)
   ===========================================================*/
CREATE TABLE IF NOT EXISTS permission_defs (
    id            BIGSERIAL PRIMARY KEY,
    uuid          UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    name          TEXT UNIQUE NOT NULL,                -- unique, not PK
    description  TEXT,
    side          INTEGER,
    metadata      JSONB,
    namespace     TEXT NOT NULL,
    resource      TEXT NOT NULL,
    action        TEXT NOT NULL,
    scope         TEXT NOT NULL DEFAULT '*',           -- NEW
    version       INT  NOT NULL DEFAULT 1,             -- NEW
    created_at    TIMESTAMP NOT NULL DEFAULT now(),
    updated_at    TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ix_permdef_lookup
  ON permission_defs(namespace, resource, scope);

/* ===========================================================
   ROLE ↔ PERMISSION_DEF  (junction)
   ===========================================================*/
CREATE TABLE IF NOT EXISTS role_permission_defs (
    role_id             BIGINT NOT NULL,
    permission_def_name TEXT NOT NULL,
    created_at          TIMESTAMP NOT NULL DEFAULT now(),
    PRIMARY KEY (role_id, permission_def_name),
    FOREIGN KEY (role_id)               REFERENCES roles(id)             ON DELETE CASCADE,
    FOREIGN KEY (permission_def_name)   REFERENCES permission_defs(name) ON DELETE CASCADE
);

/* ===========================================================
   CONCRETE PERMISSIONS  (grants / denies)
   ===========================================================*/
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'permission_effect') THEN
        CREATE TYPE permission_effect AS ENUM ('unknown', 'grant', 'forbidden');
    END IF;
END
$$;

CREATE TABLE IF NOT EXISTS permissions (
    id         BIGSERIAL PRIMARY KEY,
    uuid       UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    namespace  TEXT NOT NULL,
    resource   TEXT NOT NULL,
    action     TEXT NOT NULL,
    subject    TEXT NOT NULL,           -- user:uuid   or role:id
    effect     permission_effect NOT NULL DEFAULT 'grant',
    def_name  TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    UNIQUE(namespace, resource, action, subject)  -- fast exact-match look-up
);

CREATE INDEX IF NOT EXISTS ix_permissions_match
  ON permissions(namespace, resource, action, subject);

/* ===========================================================
   USER ↔ PERMISSION  (direct overrides)
   ===========================================================*/
CREATE TABLE IF NOT EXISTS user_permissions (
    user_id     BIGINT NOT NULL,
    permission_id BIGINT  NOT NULL,
    granted       BOOLEAN DEFAULT true,
    PRIMARY KEY (user_id, permission_id),
    FOREIGN KEY (user_id)     REFERENCES users(id)       ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id)   ON DELETE CASCADE
);

/* ===========================================================
   ROLE ↔ PERMISSION  (materialised grants)
   ===========================================================*/
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id       BIGINT NOT NULL,
    permission_id BIGINT  NOT NULL,
    granted       BOOLEAN DEFAULT true,
    created_at    TIMESTAMP NOT NULL DEFAULT now(),
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id)       REFERENCES roles(id)         ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id)   ON DELETE CASCADE
);

/* ===========================================================
   PERMISSION DEF GROUPS  (unchanged except timestamps)
   ===========================================================*/
CREATE TABLE IF NOT EXISTS permission_def_groups (
    name         TEXT PRIMARY KEY,
    display_name TEXT,
    side         INTEGER NOT NULL,
    priority     INTEGER DEFAULT 0,
    metadata     JSONB,
    created_at   TIMESTAMP NOT NULL DEFAULT now(),
    updated_at   TIMESTAMP NOT NULL DEFAULT now()
);

/* ===========================================================
   UPDATED_AT AUTOMATION
   ===========================================================*/
CREATE OR REPLACE FUNCTION touch_updated_at()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;

-- DO $$
-- DECLARE t TEXT;
-- BEGIN
--   FOR t IN
--     SELECT table_name
--     FROM information_schema.columns
--     WHERE column_name = 'updated_at'
--       AND table_schema = 'public'
--   LOOP
--     EXECUTE format($$
--       DROP TRIGGER IF EXISTS tr_%I_touch ON %I;
--       CREATE TRIGGER tr_%I_touch
--       BEFORE UPDATE ON %I
--       FOR EACH ROW EXECUTE PROCEDURE touch_updated_at();
--     $$, t, t, t, t);
--   END LOOP;
-- END$$;

/* ===========================================================
   END OF SCHEMA
   ===========================================================*/


-- From vendors\db\schema\contractors.sql (vendors\db\schema\contractors.sql)
-- schema.sql

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS contractors (
     id         BIGSERIAL PRIMARY KEY,
    uuid       UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    company_name VARCHAR(255) NOT NULL,
    company_type VARCHAR(100),
    gst VARCHAR(50),
    pan VARCHAR(50) UNIQUE,
    category VARCHAR(100),
    person_id BIGINT REFERENCES users(id),
    associated_project TEXT[], -- array of project IDs
    working_site TEXT[], -- array of site IDs
    contract_start_date TIMESTAMP,
    contract_end_date TIMESTAMP,
    status VARCHAR(50) DEFAULT 'active',
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT chk_contract_dates CHECK (contract_end_date IS NULL OR contract_end_date > contract_start_date)
);

-- Indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_contractors_person_id ON contractors(person_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_contractors_status ON contractors(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_contractors_company ON contractors(company_name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_contractors_category ON contractors(category) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_contractors_pan ON contractors(pan) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_contractors_gst ON contractors(gst) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_contractors_projects ON contractors USING GIN(associated_project) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_contractors_sites ON contractors USING GIN(working_site) WHERE deleted_at IS NULL;

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_trigger WHERE tgname = 'update_contractors_updated_at'
    ) THEN
        CREATE TRIGGER update_contractors_updated_at 
            BEFORE UPDATE ON contractors
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;

-- From projects\db\schema\dairy_sites.sql (projects\db\schema\dairy_sites.sql)
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS dairy_sites (
     id         BIGSERIAL PRIMARY KEY,
    uuid       UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    name_of_site TEXT NOT NULL,
    todays_work TEXT,
    employee_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    submitted_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

-- Optional: Index for employee lookups
CREATE INDEX IF NOT EXISTS idx_dairy_sites_employee_id ON dairy_sites (employee_id);

-- Optional: Index for soft-delete filtering
CREATE INDEX IF NOT EXISTS idx_dairy_sites_not_deleted ON dairy_sites (deleted_at) WHERE deleted_at IS NULL;


CREATE OR REPLACE VIEW dairy_sites_with_user AS
SELECT
    ds.id,
    ds.uuid,
    ds.name_of_site,
    ds.todays_work,
    u.username,
    u.fullname,
    ds.latitude,
    ds.longitude,
    ds.submitted_at,
    ds.created_at,
    ds.updated_at
FROM dairy_sites ds
JOIN users u ON ds.employee_id = u.id
WHERE ds.deleted_at IS NULL
  AND u.deleted_at IS NULL;


-- From formbuilder\db\schema\schema.sql (formbuilder\db\schema\schema.sql)
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

-- From formbuilder\db\schema\sla_tracking_schema.sql (formbuilder\db\schema\sla_tracking_schema.sql)
-- =============================================================================
-- SLA Tracking Schema - Additional tables for SLA monitoring and tracking
-- =============================================================================

-- =============================================================================
-- SLA instances table - tracks SLA performance for each form instance
-- =============================================================================
CREATE TABLE sla_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id UUID NOT NULL REFERENCES form_instances(id) ON DELETE CASCADE,
    sla_rule_id UUID NOT NULL REFERENCES sla_rules(id) ON DELETE CASCADE,
    state VARCHAR(100) NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    due_time TIMESTAMP WITH TIME ZONE NOT NULL,
    completion_time TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, completed, breached, paused
    breach_time TIMESTAMP WITH TIME ZONE,
    breach_duration INTERVAL,
    assigned_to VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_sla_instances_status CHECK (status IN ('active', 'completed', 'breached', 'paused', 'cancelled')),
    CONSTRAINT chk_sla_instances_state_not_empty CHECK (length(trim(state)) > 0),
    CONSTRAINT chk_sla_instances_due_after_start CHECK (due_time > start_time),
    CONSTRAINT chk_sla_instances_completion_after_start CHECK (completion_time IS NULL OR completion_time >= start_time)
);

-- =============================================================================
-- SLA violations table - records when SLAs are breached
-- =============================================================================
CREATE TABLE sla_violations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sla_instance_id UUID NOT NULL REFERENCES sla_instances(id) ON DELETE CASCADE,
    instance_id UUID NOT NULL REFERENCES form_instances(id) ON DELETE CASCADE,
    sla_rule_id UUID NOT NULL REFERENCES sla_rules(id) ON DELETE CASCADE,
    violation_time TIMESTAMP WITH TIME ZONE NOT NULL,
    breach_duration INTERVAL NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'medium', -- low, medium, high, critical
    resolved BOOLEAN DEFAULT false,
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolved_by VARCHAR(255),
    resolution_notes TEXT,
    notification_sent BOOLEAN DEFAULT false,
    escalation_triggered BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_sla_violations_severity CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    CONSTRAINT chk_sla_violations_resolved_consistency CHECK (
        (resolved = false AND resolved_at IS NULL AND resolved_by IS NULL) OR
        (resolved = true AND resolved_at IS NOT NULL AND resolved_by IS NOT NULL)
    )
);

-- =============================================================================
-- SLA escalation instances table - tracks escalation executions
-- =============================================================================
CREATE TABLE sla_escalation_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sla_instance_id UUID NOT NULL REFERENCES sla_instances(id) ON DELETE CASCADE,
    escalation_level INTEGER NOT NULL CHECK (escalation_level > 0),
    triggered_at TIMESTAMP WITH TIME ZONE NOT NULL,
    actions_executed JSONB DEFAULT '[]'::jsonb,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, executing, completed, failed
    error_message TEXT,
    retry_count INTEGER DEFAULT 0 CHECK (retry_count >= 0),
    next_retry_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_sla_escalation_status CHECK (status IN ('pending', 'executing', 'completed', 'failed'))
);

-- =============================================================================
-- SLA notifications table - tracks all SLA-related notifications
-- =============================================================================
CREATE TABLE sla_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sla_instance_id UUID REFERENCES sla_instances(id) ON DELETE CASCADE,
    sla_violation_id UUID REFERENCES sla_violations(id) ON DELETE CASCADE,
    escalation_instance_id UUID REFERENCES sla_escalation_instances(id) ON DELETE CASCADE,
    notification_type VARCHAR(50) NOT NULL, -- warning, breach, escalation, reminder
    recipient VARCHAR(255) NOT NULL,
    channel VARCHAR(20) NOT NULL, -- email, sms, push, webhook
    subject VARCHAR(500),
    message TEXT NOT NULL,
    template_used VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, sent, failed, delivered
    sent_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    retry_count INTEGER DEFAULT 0 CHECK (retry_count >= 0),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_sla_notifications_type CHECK (notification_type IN ('warning', 'breach', 'escalation', 'reminder', 'resolution')),
    CONSTRAINT chk_sla_notifications_channel CHECK (channel IN ('email', 'sms', 'push', 'webhook', 'in_app')),
    CONSTRAINT chk_sla_notifications_status CHECK (status IN ('pending', 'sent', 'failed', 'delivered')),
    CONSTRAINT chk_sla_notifications_recipient_not_empty CHECK (length(trim(recipient)) > 0),
    CONSTRAINT chk_sla_notifications_at_least_one_reference CHECK (
        sla_instance_id IS NOT NULL OR sla_violation_id IS NOT NULL OR escalation_instance_id IS NOT NULL
    )
);

-- =============================================================================
-- SLA metrics table - stores aggregated SLA performance metrics
-- =============================================================================
CREATE TABLE sla_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_date DATE NOT NULL,
    sla_rule_id UUID REFERENCES sla_rules(id) ON DELETE CASCADE,
    state VARCHAR(100),
    assigned_role VARCHAR(255),
    total_instances INTEGER NOT NULL DEFAULT 0 CHECK (total_instances >= 0),
    completed_on_time INTEGER NOT NULL DEFAULT 0 CHECK (completed_on_time >= 0),
    breached_instances INTEGER NOT NULL DEFAULT 0 CHECK (breached_instances >= 0),
    avg_completion_time INTERVAL,
    avg_breach_duration INTERVAL,
    compliance_percentage DECIMAL(5,2) CHECK (compliance_percentage >= 0 AND compliance_percentage <= 100),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_sla_metrics_consistency CHECK (
        completed_on_time + breached_instances <= total_instances
    ),
    UNIQUE(metric_date, sla_rule_id, state, assigned_role)
);

-- =============================================================================
-- SLA pause logs table - tracks when SLAs are paused/resumed
-- =============================================================================
CREATE TABLE sla_pause_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sla_instance_id UUID NOT NULL REFERENCES sla_instances(id) ON DELETE CASCADE,
    action VARCHAR(10) NOT NULL, -- pause, resume
    reason TEXT,
    paused_by VARCHAR(255) NOT NULL,
    paused_at TIMESTAMP WITH TIME ZONE NOT NULL,
    resumed_at TIMESTAMP WITH TIME ZONE,
    pause_duration INTERVAL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_sla_pause_action CHECK (action IN ('pause', 'resume')),
    CONSTRAINT chk_sla_pause_consistency CHECK (
        (action = 'pause' AND resumed_at IS NULL) OR
        (action = 'resume' AND resumed_at IS NOT NULL AND resumed_at > paused_at)
    )
);

-- =============================================================================
-- Create indexes for SLA tracking tables
-- =============================================================================

-- SLA instances indexes
CREATE INDEX idx_sla_instances_instance_id ON sla_instances(instance_id);
CREATE INDEX idx_sla_instances_sla_rule_id ON sla_instances(sla_rule_id);
CREATE INDEX idx_sla_instances_state ON sla_instances(state);
CREATE INDEX idx_sla_instances_status ON sla_instances(status);
CREATE INDEX idx_sla_instances_assigned_to ON sla_instances(assigned_to);
CREATE INDEX idx_sla_instances_due_time ON sla_instances(due_time);
CREATE INDEX idx_sla_instances_active_due ON sla_instances(status, due_time) WHERE status = 'active';
CREATE INDEX idx_sla_instances_state_status ON sla_instances(state, status);

-- SLA violations indexes
CREATE INDEX idx_sla_violations_sla_instance_id ON sla_violations(sla_instance_id);
CREATE INDEX idx_sla_violations_instance_id ON sla_violations(instance_id);
CREATE INDEX idx_sla_violations_sla_rule_id ON sla_violations(sla_rule_id);
CREATE INDEX idx_sla_violations_violation_time ON sla_violations(violation_time);
CREATE INDEX idx_sla_violations_severity ON sla_violations(severity);
CREATE INDEX idx_sla_violations_resolved ON sla_violations(resolved);
CREATE INDEX idx_sla_violations_unresolved ON sla_violations(resolved, violation_time) WHERE resolved = false;

-- SLA escalation instances indexes
CREATE INDEX idx_sla_escalation_instances_sla_instance_id ON sla_escalation_instances(sla_instance_id);
CREATE INDEX idx_sla_escalation_instances_level ON sla_escalation_instances(escalation_level);
CREATE INDEX idx_sla_escalation_instances_triggered_at ON sla_escalation_instances(triggered_at);
CREATE INDEX idx_sla_escalation_instances_status ON sla_escalation_instances(status);
CREATE INDEX idx_sla_escalation_instances_retry ON sla_escalation_instances(status, next_retry_at) WHERE status = 'failed';

-- SLA notifications indexes
CREATE INDEX idx_sla_notifications_sla_instance_id ON sla_notifications(sla_instance_id);
CREATE INDEX idx_sla_notifications_sla_violation_id ON sla_notifications(sla_violation_id);
CREATE INDEX idx_sla_notifications_escalation_instance_id ON sla_notifications(escalation_instance_id);
CREATE INDEX idx_sla_notifications_type ON sla_notifications(notification_type);
CREATE INDEX idx_sla_notifications_recipient ON sla_notifications(recipient);
CREATE INDEX idx_sla_notifications_status ON sla_notifications(status);
CREATE INDEX idx_sla_notifications_channel ON sla_notifications(channel);
CREATE INDEX idx_sla_notifications_pending ON sla_notifications(status, created_at) WHERE status = 'pending';

-- SLA metrics indexes
CREATE INDEX idx_sla_metrics_date ON sla_metrics(metric_date);
CREATE INDEX idx_sla_metrics_sla_rule_id ON sla_metrics(sla_rule_id);
CREATE INDEX idx_sla_metrics_state ON sla_metrics(state);
CREATE INDEX idx_sla_metrics_role ON sla_metrics(assigned_role);
CREATE INDEX idx_sla_metrics_date_rule ON sla_metrics(metric_date, sla_rule_id);

-- SLA pause logs indexes
CREATE INDEX idx_sla_pause_logs_sla_instance_id ON sla_pause_logs(sla_instance_id);
CREATE INDEX idx_sla_pause_logs_action ON sla_pause_logs(action);
CREATE INDEX idx_sla_pause_logs_paused_by ON sla_pause_logs(paused_by);
CREATE INDEX idx_sla_pause_logs_paused_at ON sla_pause_logs(paused_at);

-- =============================================================================
-- Create views for common SLA queries
-- =============================================================================

-- Active SLA instances with time remaining
CREATE OR REPLACE VIEW active_sla_instances AS
SELECT 
    si.*,
    fi.form_id,
    fi.created_by as instance_created_by,
    sr.name as sla_rule_name,
    sr.duration as sla_duration,
    EXTRACT(EPOCH FROM (si.due_time - NOW())) / 3600 as hours_remaining,
    CASE 
        WHEN NOW() > si.due_time THEN true 
        ELSE false 
    END as is_overdue,
    EXTRACT(EPOCH FROM (NOW() - si.due_time)) / 3600 as hours_overdue
FROM sla_instances si
JOIN form_instances fi ON si.instance_id = fi.id
JOIN sla_rules sr ON si.sla_rule_id = sr.id
WHERE si.status = 'active' AND fi.deleted_at IS NULL;

-- SLA compliance summary by rule
CREATE OR REPLACE VIEW sla_compliance_summary AS
SELECT 
    sr.id as sla_rule_id,
    sr.name as sla_rule_name,
    sr.state,
    COUNT(si.id) as total_instances,
    COUNT(CASE WHEN si.status = 'completed' THEN 1 END) as completed_instances,
    COUNT(CASE WHEN si.status = 'breached' THEN 1 END) as breached_instances,
    COUNT(CASE WHEN si.status = 'active' AND NOW() > si.due_time THEN 1 END) as overdue_instances,
    ROUND(
        (COUNT(CASE WHEN si.status = 'completed' THEN 1 END) * 100.0) / 
        NULLIF(COUNT(si.id), 0), 2
    ) as compliance_percentage
FROM sla_rules sr
LEFT JOIN sla_instances si ON sr.id = si.sla_rule_id
WHERE sr.deleted_at IS NULL
GROUP BY sr.id, sr.name, sr.state;

-- Recent SLA violations
CREATE OR REPLACE VIEW recent_sla_violations AS
SELECT 
    sv.*,
    si.instance_id as si_instance_id,
    si.state,
    si.assigned_to,
    sr.name as sla_rule_name,
    fi.form_id,
    fi.created_by as instance_created_by
FROM sla_violations sv
JOIN sla_instances si ON sv.sla_instance_id = si.id
JOIN sla_rules sr ON sv.sla_rule_id = sr.id
JOIN form_instances fi ON sv.instance_id = fi.id
WHERE sv.created_at >= NOW() - INTERVAL '30 days'
ORDER BY sv.violation_time DESC;


