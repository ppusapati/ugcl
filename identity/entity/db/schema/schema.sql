-- Entity types enumeration
CREATE TYPE entity_type AS ENUM (
    'EMPLOYEE',
    'CONTRACTOR',
    'VENDOR',
    'CLIENT',
    'DEVICE',
    'DRONE',
    'BOT',
    'SYSTEM',
    'ADMIN',
    'AGENT'
);

-- Entity status enumeration
CREATE TYPE entity_status AS ENUM (
    'ACTIVE',
    'INACTIVE',
    'SUSPENDED',
    'ARCHIVED'
);

-- Reference source enumeration
CREATE TYPE reference_source AS ENUM (
    'EMPLOYEE',
    'CONTRACTOR',
    'VENDOR',
    'CLIENT',
    'DEVICE',
    'DRONE',
    'BOT'
);

-- Entities table - unified identity abstraction
CREATE TABLE entities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    entity_type entity_type NOT NULL,
    user_id UUID NOT NULL, -- Links to identity.users table
    reference_id UUID NOT NULL, -- Points to domain-specific record
    reference_source reference_source NOT NULL,
    status entity_status NOT NULL DEFAULT 'ACTIVE',

    -- Organizational context
    division_id UUID,
    branch_id UUID,
    department_id UUID,

    -- Metadata for extensibility
    metadata JSONB DEFAULT '{}',

    -- Audit fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,

    -- Constraints
    CONSTRAINT unique_user_per_tenant UNIQUE(tenant_id, user_id),
    CONSTRAINT unique_reference_per_tenant UNIQUE(tenant_id, reference_id, reference_source)
);

-- Indexes for entities
CREATE INDEX idx_entities_tenant_id ON entities(tenant_id);
CREATE INDEX idx_entities_user_id ON entities(user_id);
CREATE INDEX idx_entities_entity_type ON entities(entity_type);
CREATE INDEX idx_entities_status ON entities(status);
CREATE INDEX idx_entities_reference ON entities(reference_id, reference_source);
CREATE INDEX idx_entities_division ON entities(division_id) WHERE division_id IS NOT NULL;
CREATE INDEX idx_entities_branch ON entities(branch_id) WHERE branch_id IS NOT NULL;
CREATE INDEX idx_entities_department ON entities(department_id) WHERE department_id IS NOT NULL;
CREATE INDEX idx_entities_metadata ON entities USING GIN(metadata);

-- Entity role bindings table
CREATE TABLE entity_role_bindings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    entity_id UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    role_id UUID NOT NULL, -- Links to identity.roles table

    -- Optional scope limitations
    division_id UUID,
    branch_id UUID,
    department_id UUID,

    -- Time-bound access
    valid_from TIMESTAMPTZ,
    valid_until TIMESTAMPTZ,

    -- Audit fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,

    -- Constraints
    CONSTRAINT unique_entity_role_scope UNIQUE(tenant_id, entity_id, role_id, division_id, branch_id, department_id)
);

-- Indexes for entity_role_bindings
CREATE INDEX idx_entity_role_bindings_tenant_id ON entity_role_bindings(tenant_id);
CREATE INDEX idx_entity_role_bindings_entity_id ON entity_role_bindings(entity_id);
CREATE INDEX idx_entity_role_bindings_role_id ON entity_role_bindings(role_id);
CREATE INDEX idx_entity_role_bindings_division ON entity_role_bindings(division_id) WHERE division_id IS NOT NULL;
CREATE INDEX idx_entity_role_bindings_branch ON entity_role_bindings(branch_id) WHERE branch_id IS NOT NULL;
CREATE INDEX idx_entity_role_bindings_department ON entity_role_bindings(department_id) WHERE department_id IS NOT NULL;
CREATE INDEX idx_entity_role_bindings_validity ON entity_role_bindings(valid_from, valid_until)
    WHERE valid_from IS NOT NULL OR valid_until IS NOT NULL;

-- Comments for documentation
COMMENT ON TABLE entities IS 'Unified identity abstraction for all actors (human, machine, system) in the platform';
COMMENT ON COLUMN entities.user_id IS 'Links to the identity.users table for authentication';
COMMENT ON COLUMN entities.reference_id IS 'Points to domain-specific record (employee_id, device_id, etc.)';
COMMENT ON COLUMN entities.reference_source IS 'Indicates which domain service owns the reference_id';
COMMENT ON COLUMN entities.metadata IS 'Extensible JSON field for entity-specific attributes';

COMMENT ON TABLE entity_role_bindings IS 'Manages role assignments to entities with optional organizational and temporal scoping';
COMMENT ON COLUMN entity_role_bindings.valid_from IS 'Optional start date for time-bound role access';
COMMENT ON COLUMN entity_role_bindings.valid_until IS 'Optional end date for time-bound role access';
