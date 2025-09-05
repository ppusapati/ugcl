-- Auto-Generated Schema

-- From identity/db/sqlc/schema.sql (identity)
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


-- From vendors/db/schema/contractors.sql (vendors)
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

-- From projects/db/schema/dairy_sites.sql (projects)
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


-- From formbuilder/db/schema/schema.sql (formbuilder)
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

