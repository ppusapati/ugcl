# UGCL Backend v2 - Database Design & Schema Architecture

## Document Information

- **Version:** 2.0
- **Last Updated:** 2025-10-06
- **Status:** Active
- **Database:** PostgreSQL 15+
- **Schema Management:** SQLC + Atlas Migrations

---

## Table of Contents

1. [Database Overview](#database-overview)
2. [Multi-Tenancy Design](#multi-tenancy-design)
3. [Schema Organization](#schema-organization)
4. [Domain Schemas](#domain-schemas)
5. [Index Strategy](#index-strategy)
6. [Data Integrity & Constraints](#data-integrity--constraints)
7. [Audit Trail Design](#audit-trail-design)
8. [Performance Optimization](#performance-optimization)
9. [Migration Strategy](#migration-strategy)
10. [Naming Conventions](#naming-conventions)

---

## Database Overview

### Database Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                     PostgreSQL Database (ugcl)                       │
│                                                                      │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │                  Schema: testing2 (default)                   │  │
│  │                                                               │  │
│  │  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐    │  │
│  │  │   Identity    │  │  Personnel    │  │ Organization  │    │  │
│  │  │   Domain      │  │   Domain      │  │   Domain      │    │  │
│  │  │               │  │               │  │               │    │  │
│  │  │ • users       │  │ • employees   │  │ • divisions   │    │  │
│  │  │ • roles       │  │ • contractors │  │ • branches    │    │  │
│  │  │ • permissions │  │ • vendors     │  │ • departments │    │  │
│  │  │ • tenants     │  │               │  │               │    │  │
│  │  │ • entities    │  │               │  │               │    │  │
│  │  └───────────────┘  └───────────────┘  └───────────────┘    │  │
│  │                                                               │  │
│  │  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐    │  │
│  │  │   Content     │  │  Operations   │  │   Analytics   │    │  │
│  │  │   Domain      │  │   Domain      │  │   Domain      │    │  │
│  │  │               │  │               │  │               │    │  │
│  │  │ • documents   │  │ • forms       │  │ • insights    │    │  │
│  │  │ • doc_shares  │  │ • workflows   │  │ • reports     │    │  │
│  │  │               │  │ • projects    │  │               │    │  │
│  │  │               │  │ • schedules   │  │               │    │  │
│  │  └───────────────┘  └───────────────┘  └───────────────┘    │  │
│  └──────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
```

### Key Design Principles

1. **Multi-Tenancy:** Every table includes `tenant_id` for data isolation
2. **Soft Deletes:** `deleted_at` timestamp instead of hard deletes
3. **Audit Trail:** `created_at`, `updated_at`, `created_by`, `updated_by` on all tables
4. **UUID Primary Keys:** UUIDs for distributed ID generation
5. **JSONB for Flexibility:** Metadata fields for extensibility
6. **Enum Types:** Strong typing for categorical data
7. **Foreign Key Constraints:** Referential integrity enforcement
8. **Partial Indexes:** Performance optimization for filtered queries

### Database Statistics

- **Total Tables:** ~50+ tables across all domains
- **Total Indexes:** ~200+ indexes for query optimization
- **Total Enums:** ~20+ enum types for type safety
- **Average Table Size:** 5-50 columns depending on domain
- **Connection Pool:** 2-10 connections per application instance

---

## Multi-Tenancy Design

### Tenant Isolation Strategy

**Row-Level Isolation:** Every table includes `tenant_id` column with mandatory filtering.

```sql
-- Standard multi-tenant table pattern
CREATE TABLE {table_name} (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,  -- MANDATORY for all tables

    -- Business columns
    ...

    -- Audit columns
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,
    deleted_at TIMESTAMPTZ,  -- Soft delete

    -- Indexes
    ...
);

-- Tenant index on every table
CREATE INDEX idx_{table}_tenant_id ON {table_name}(tenant_id);
```

### Tenant Data Access Pattern

```sql
-- ✅ CORRECT: Always filter by tenant_id
SELECT * FROM documents
WHERE tenant_id = $1  -- Injected from JWT context
  AND deleted_at IS NULL
  AND owner_entity_id = $2;

-- ❌ WRONG: Missing tenant_id filter (security vulnerability)
SELECT * FROM documents
WHERE owner_entity_id = $1;  -- Can access other tenants' data!

-- ✅ CORRECT: Multi-table join with tenant isolation
SELECT
    d.id,
    d.file_name,
    e.entity_type,
    u.username
FROM documents d
INNER JOIN entities e ON e.id = d.owner_entity_id AND e.tenant_id = d.tenant_id
INNER JOIN users u ON u.uuid = e.user_id
WHERE d.tenant_id = $1
  AND d.deleted_at IS NULL;
```

### Tenant Context Injection

**Application Layer:**
```go
// Middleware extracts tenant from JWT
func (m *AuthMiddleware) ExtractTenant(ctx context.Context) (string, error) {
    claims, ok := ClaimsFromContext(ctx)
    if !ok {
        return "", ErrNoTenant
    }
    return claims.TenantID, nil
}

// Repository automatically adds tenant filter
func (r *DocumentRepository) List(ctx context.Context, filters ListFilters) ([]Document, error) {
    tenantID, err := middleware.TenantFromContext(ctx)
    if err != nil {
        return nil, err
    }

    // SQLC query with tenant_id parameter
    return r.queries.ListDocuments(ctx, ListDocumentsParams{
        TenantID: tenantID,
        Limit:    filters.Limit,
        Offset:   filters.Offset,
    })
}
```

**Database Level (Future Enhancement):**
```sql
-- Row Level Security (RLS) for additional safety
ALTER TABLE documents ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON documents
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);
```

### Tenant Schema

```sql
-- Tenants table (in identity schema)
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    domain VARCHAR(255),  -- Custom domain (e.g., acme.ugcl.com)
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',

    -- Configuration
    settings JSONB DEFAULT '{}',  -- Tenant-specific settings
    features JSONB DEFAULT '{}',  -- Enabled features
    limits JSONB DEFAULT '{}',    -- Resource limits (users, storage, etc.)

    -- Branding
    logo_url TEXT,
    primary_color VARCHAR(7),

    -- Subscription
    plan VARCHAR(50) DEFAULT 'FREE',
    subscription_start DATE,
    subscription_end DATE,

    -- Contact
    admin_email VARCHAR(255),
    admin_phone VARCHAR(20),

    -- Audit
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT chk_tenants_status CHECK (status IN ('ACTIVE', 'SUSPENDED', 'TRIAL', 'EXPIRED'))
);

CREATE INDEX idx_tenants_code ON tenants(code);
CREATE INDEX idx_tenants_status ON tenants(status) WHERE status = 'ACTIVE';
CREATE INDEX idx_tenants_domain ON tenants(domain) WHERE domain IS NOT NULL;
```

---

## Schema Organization

### Domain-Driven Schema Design

Each business domain has its own set of related tables:

```
Identity Domain
├── users                    (core user accounts)
├── roles                    (hierarchical roles)
├── permissions              (concrete permissions)
├── permission_defs          (permission templates)
├── user_roles              (user-role assignments)
├── role_permissions        (role-permission mappings)
├── user_tenant_roles       (multi-tenant role assignments)
├── user_preferences        (user settings)
├── tenants                 (tenant management)
├── entities                (universal entity abstraction)
└── entity_role_bindings    (entity-role-org scope)

Organization Domain
├── divisions               (business units)
├── branches               (physical locations)
└── departments            (functional units)

Personnel Domain
├── employees              (employee records)
├── contractors           (contractor records)
└── vendors               (vendor/supplier records)

Content Domain
├── documents             (document metadata)
└── document_shares       (sharing and permissions)

Operations Domain
├── forms                 (form definitions)
├── form_instances        (form submissions)
├── workflows            (workflow definitions)
├── audit_logs           (audit trail)
├── attachments          (file attachments)
├── comments             (instance comments)
├── sla_rules            (SLA definitions)
├── escalations          (escalation rules)
├── approval_actions     (approval history)
├── approval_delegates   (delegation management)
└── approval_reports     (metrics and analytics)

Projects Domain
├── dairy_sites          (project sites)
├── site_details         (site metadata)
└── project_teams        (team assignments)

Analytics Domain
├── insights             (KPIs and metrics)
└── reports              (generated reports)
```

---

## Domain Schemas

### 1. Identity Domain

#### Users Table

```sql
-- Users table - core authentication and user management
CREATE TABLE users (
    id              BIGSERIAL PRIMARY KEY,
    uuid            UUID  UNIQUE NOT NULL DEFAULT gen_random_uuid(),

    -- Authentication
    username        TEXT,
    normalized_username TEXT,
    email           TEXT,
    normalized_email TEXT,
    password_hash   TEXT NOT NULL,

    -- Verification
    phone           TEXT,
    phone_verified  BOOLEAN DEFAULT false,
    email_verified  BOOLEAN DEFAULT false,

    -- Profile
    fullname        TEXT,
    gender          gender_type DEFAULT 'unknown',
    avatar          BYTEA,

    -- Security
    salt            BYTEA,
    two_factor_enabled BOOLEAN DEFAULT false,
    two_factor_secret TEXT,

    -- Status
    is_active       BOOLEAN NOT NULL DEFAULT true,
    status          user_status DEFAULT 'active',
    last_login      TIMESTAMP,

    -- Cache
    roles_cache     TEXT[],  -- Cached role names for quick access

    -- Audit
    created_at      TIMESTAMP NOT NULL DEFAULT now(),
    updated_at      TIMESTAMP NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMP
);

-- Indexes
CREATE INDEX idx_users_username ON users(normalized_username);
CREATE INDEX idx_users_email ON users(normalized_email);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_last_login ON users(last_login);

-- Enums
CREATE TYPE user_status AS ENUM (
    'unspecified',
    'active',
    'inactive',
    'suspended',
    'pending_verification',
    'locked'
);

CREATE TYPE gender_type AS ENUM ('unknown', 'male', 'female', 'other');
```

#### Roles Table (Hierarchical)

```sql
-- Roles table with parent-child hierarchy support
CREATE TABLE roles (
    id          BIGSERIAL  PRIMARY KEY,
    uuid        UUID  UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    parent_id   BIGINT  NULL REFERENCES roles(id) ON DELETE SET NULL,
    name        TEXT  NOT NULL,
    description TEXT,
    is_preserved BOOLEAN NOT NULL DEFAULT FALSE,  -- System roles can't be deleted
    metadata    JSONB DEFAULT '{}',
    created_at  TIMESTAMP NOT NULL DEFAULT now(),
    updated_at  TIMESTAMP NOT NULL DEFAULT now(),
    UNIQUE(name)
);

CREATE INDEX idx_roles_parent ON roles(parent_id);
CREATE INDEX idx_roles_name ON roles(name);

-- Example role hierarchy:
-- Super Admin (parent_id: NULL)
--   └── Admin (parent_id: Super Admin ID)
--         ├── Manager (parent_id: Admin ID)
--         └── Supervisor (parent_id: Admin ID)
--               └── User (parent_id: Supervisor ID)
```

#### Permissions System

```sql
-- Permission definitions (templates)
CREATE TABLE permission_defs (
    id            BIGSERIAL PRIMARY KEY,
    uuid          UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    name          TEXT UNIQUE NOT NULL,  -- e.g., "document:read:own"
    description   TEXT,

    -- Permission structure
    namespace     TEXT NOT NULL,  -- e.g., "document"
    resource      TEXT NOT NULL,  -- e.g., "file"
    action        TEXT NOT NULL,  -- e.g., "read", "write", "delete"
    scope         TEXT NOT NULL DEFAULT '*',  -- e.g., "own", "branch", "division", "*"

    -- Versioning
    version       INT  NOT NULL DEFAULT 1,
    side          INTEGER,  -- UI grouping
    metadata      JSONB,

    created_at    TIMESTAMP NOT NULL DEFAULT now(),
    updated_at    TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_permdef_lookup ON permission_defs(namespace, resource, scope);
CREATE INDEX idx_permdef_name ON permission_defs(name);

-- Concrete permissions (materialized from definitions)
CREATE TYPE permission_effect AS ENUM ('unknown', 'grant', 'forbidden');

CREATE TABLE permissions (
    id         BIGSERIAL PRIMARY KEY,
    uuid       UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    namespace  TEXT NOT NULL,
    resource   TEXT NOT NULL,
    action     TEXT NOT NULL,
    subject    TEXT NOT NULL,  -- "user:uuid" or "role:id"
    effect     permission_effect NOT NULL DEFAULT 'grant',
    def_name   TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    UNIQUE(namespace, resource, action, subject)
);

CREATE INDEX idx_permissions_match ON permissions(namespace, resource, action, subject);
CREATE INDEX idx_permissions_subject ON permissions(subject);

-- Role to permission definition mapping
CREATE TABLE role_permission_defs (
    role_id             BIGINT NOT NULL,
    permission_def_name TEXT NOT NULL,
    created_at          TIMESTAMP NOT NULL DEFAULT now(),
    PRIMARY KEY (role_id, permission_def_name),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_def_name) REFERENCES permission_defs(name) ON DELETE CASCADE
);

-- User to role mapping
CREATE TABLE user_roles (
    user_id BIGINT NOT NULL,
    role_id BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);

-- Multi-tenant role assignments
CREATE TABLE user_tenant_roles (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id   TEXT NOT NULL,
    roles       TEXT[] NOT NULL DEFAULT '{}',
    is_active   BOOLEAN DEFAULT true,
    assigned_at TIMESTAMP NOT NULL DEFAULT now(),
    created_at  TIMESTAMP NOT NULL DEFAULT now(),
    updated_at  TIMESTAMP NOT NULL DEFAULT now(),
    UNIQUE(user_id, tenant_id)
);

CREATE INDEX idx_user_tenant_roles_user_id ON user_tenant_roles(user_id);
CREATE INDEX idx_user_tenant_roles_tenant_id ON user_tenant_roles(tenant_id);
```

#### Entities Table (Universal Identity Abstraction)

```sql
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

CREATE TYPE entity_status AS ENUM (
    'ACTIVE',
    'INACTIVE',
    'SUSPENDED',
    'ARCHIVED'
);

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
    user_id UUID NOT NULL,  -- Links to users.uuid
    reference_id UUID NOT NULL,  -- Points to domain record (employee.id, device.id, etc.)
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

-- Indexes
CREATE INDEX idx_entities_tenant_id ON entities(tenant_id);
CREATE INDEX idx_entities_user_id ON entities(user_id);
CREATE INDEX idx_entities_entity_type ON entities(entity_type);
CREATE INDEX idx_entities_reference ON entities(reference_id, reference_source);
CREATE INDEX idx_entities_metadata ON entities USING GIN(metadata);

COMMENT ON TABLE entities IS 'Unified identity abstraction for all actors (human, machine, system)';
COMMENT ON COLUMN entities.user_id IS 'Links to the users table for authentication';
COMMENT ON COLUMN entities.reference_id IS 'Points to domain-specific record';
```

#### Entity Role Bindings (Scoped Permissions)

```sql
-- Entity role bindings with organizational and temporal scoping
CREATE TABLE entity_role_bindings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    entity_id UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    role_id UUID NOT NULL,  -- Links to roles.uuid

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
    CONSTRAINT unique_entity_role_scope UNIQUE(
        tenant_id, entity_id, role_id,
        division_id, branch_id, department_id
    )
);

-- Indexes
CREATE INDEX idx_entity_role_bindings_entity_id ON entity_role_bindings(entity_id);
CREATE INDEX idx_entity_role_bindings_role_id ON entity_role_bindings(role_id);
CREATE INDEX idx_entity_role_bindings_validity ON entity_role_bindings(valid_from, valid_until)
    WHERE valid_from IS NOT NULL OR valid_until IS NOT NULL;
```

**Entity Relationship Diagram:**

```
┌─────────────┐         ┌─────────────┐         ┌─────────────┐
│   Users     │         │  Entities   │         │  Employees  │
│             │◄────────│             │────────►│             │
│ • uuid      │         │ • user_id   │         │ • id        │
│ • username  │         │ • reference │         │ • name      │
│ • email     │         │   _id       │         │ • emp_code  │
│ • password  │         │ • tenant_id │         │             │
└─────────────┘         └──────┬──────┘         └─────────────┘
                               │
                               │
                        ┌──────▼──────────┐
                        │ Entity Role     │
                        │ Bindings        │
                        │                 │
                        │ • entity_id     │
                        │ • role_id       │
                        │ • division_id   │
                        │ • valid_from    │
                        └─────────────────┘
```

### 2. Organization Domain

```sql
-- Divisions table
CREATE TABLE divisions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    head_user_id VARCHAR(255),  -- Division head (VP, President)
    is_active BOOLEAN DEFAULT true,
    display_order INTEGER DEFAULT 0,
    metadata JSONB DEFAULT '{}',

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by VARCHAR(255),
    updated_by VARCHAR(255),

    UNIQUE(tenant_id, code)
);

CREATE INDEX idx_divisions_tenant_id ON divisions(tenant_id);
CREATE INDEX idx_divisions_is_active ON divisions(is_active) WHERE is_active = true;

-- Branches table
CREATE TABLE branches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    division_id UUID NOT NULL REFERENCES divisions(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    branch_type VARCHAR(50),  -- 'HQ', 'Regional', 'Local', 'Satellite'

    -- Location Details
    address_line1 TEXT,
    city VARCHAR(100),
    state VARCHAR(100),
    country VARCHAR(100) DEFAULT 'India',
    postal_code VARCHAR(20),
    latitude NUMERIC(10, 8),
    longitude NUMERIC(11, 8),

    -- Contact
    phone VARCHAR(20),
    email VARCHAR(255),
    branch_manager_user_id VARCHAR(255),

    is_active BOOLEAN DEFAULT true,
    display_order INTEGER DEFAULT 0,
    metadata JSONB DEFAULT '{}',

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by VARCHAR(255),
    updated_by VARCHAR(255),

    UNIQUE(tenant_id, code)
);

CREATE INDEX idx_branches_division_id ON branches(division_id);
CREATE INDEX idx_branches_tenant_id ON branches(tenant_id);
CREATE INDEX idx_branches_city ON branches(city);

-- Departments table (can be division-level or business-level)
CREATE TABLE departments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    division_id UUID NULL REFERENCES divisions(id) ON DELETE CASCADE,  -- NULL = business-level
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    department_type VARCHAR(50),  -- 'Operational', 'Support', 'Administrative'
    head_user_id VARCHAR(255),
    parent_department_id UUID REFERENCES departments(id),  -- For sub-departments
    is_active BOOLEAN DEFAULT true,
    display_order INTEGER DEFAULT 0,
    metadata JSONB DEFAULT '{}',

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by VARCHAR(255),
    updated_by VARCHAR(255),

    CHECK (
        (division_id IS NULL AND parent_department_id IS NULL) OR
        (division_id IS NOT NULL)
    )
);

CREATE UNIQUE INDEX idx_departments_unique_code ON departments (
    tenant_id,
    code,
    COALESCE(division_id::text, '')
);

CREATE INDEX idx_departments_tenant_id ON departments(tenant_id);
CREATE INDEX idx_departments_division_id ON departments(division_id) WHERE division_id IS NOT NULL;
CREATE INDEX idx_departments_parent_id ON departments(parent_department_id) WHERE parent_department_id IS NOT NULL;
```

**Organization Hierarchy View:**

```sql
-- Recursive CTE for department hierarchy
CREATE VIEW department_hierarchy AS
WITH RECURSIVE dept_tree AS (
    SELECT
        id,
        tenant_id,
        division_id,
        code,
        name,
        parent_department_id,
        1 as level,
        ARRAY[id] as path,
        code as full_path
    FROM departments
    WHERE parent_department_id IS NULL

    UNION ALL

    SELECT
        d.id,
        d.tenant_id,
        d.division_id,
        d.code,
        d.name,
        d.parent_department_id,
        dt.level + 1,
        dt.path || d.id,
        dt.full_path || ' > ' || d.code
    FROM departments d
    INNER JOIN dept_tree dt ON d.parent_department_id = dt.id
)
SELECT * FROM dept_tree;
```

### 3. Content Domain

```sql
-- Document types and categories
CREATE TYPE document_type AS ENUM (
    'PDF', 'IMAGE', 'VIDEO', 'AUDIO', 'SPREADSHEET',
    'PRESENTATION', 'TEXT', 'ARCHIVE', 'OTHER'
);

CREATE TYPE document_category AS ENUM (
    'CONTRACT', 'INVOICE', 'REPORT', 'MEMO', 'LETTER',
    'FORM', 'POLICY', 'PROCEDURE', 'MANUAL', 'PRESENTATION',
    'DRAWING', 'PHOTO', 'VIDEO', 'AUDIO', 'OTHER'
);

CREATE TYPE processing_status AS ENUM (
    'PENDING', 'PROCESSING', 'COMPLETED', 'FAILED'
);

CREATE TYPE ocr_status AS ENUM (
    'PENDING', 'PROCESSING', 'COMPLETED', 'FAILED', 'SKIPPED'
);

CREATE TYPE virus_scan_status AS ENUM (
    'PENDING', 'SCANNING', 'CLEAN', 'INFECTED', 'FAILED'
);

-- Documents table
CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,

    -- Entity-based ownership (polymorphic)
    owner_entity_type VARCHAR(50),
    owner_entity_id UUID,

    -- Organizational context
    division_id UUID,
    branch_id UUID,
    department_id UUID,

    -- Document classification
    document_type document_type NOT NULL DEFAULT 'OTHER',
    document_category document_category NOT NULL DEFAULT 'OTHER',

    -- File information
    file_name VARCHAR(500) NOT NULL,
    original_name VARCHAR(500) NOT NULL,
    title VARCHAR(1000),
    description TEXT,
    mime_type VARCHAR(255) NOT NULL,
    file_extension VARCHAR(10),
    size_bytes BIGINT NOT NULL,
    compressed_size BIGINT DEFAULT 0,
    checksum VARCHAR(64) NOT NULL,

    -- Storage paths (MinIO)
    storage_path TEXT NOT NULL,
    compressed_path TEXT,
    thumbnail_path TEXT,
    preview_path TEXT,

    -- Document metadata
    page_count INTEGER DEFAULT 0,
    word_count INTEGER DEFAULT 0,
    extracted_text TEXT,
    language VARCHAR(10),
    author VARCHAR(255),
    subject VARCHAR(500),
    keywords TEXT,

    -- Processing status
    processing_status processing_status NOT NULL DEFAULT 'PENDING',
    processing_error TEXT,
    processed_at TIMESTAMPTZ,

    -- Compression
    compression_type VARCHAR(50) DEFAULT 'tar_zstd',
    compression_ratio NUMERIC(5, 2) DEFAULT 0,

    -- OCR
    ocr_status ocr_status NOT NULL DEFAULT 'PENDING',
    ocr_confidence NUMERIC(5, 2) DEFAULT 0,
    ocr_processed_at TIMESTAMPTZ,

    -- Security
    virus_scan_status virus_scan_status NOT NULL DEFAULT 'PENDING',
    virus_scan_result VARCHAR(50),
    virus_scanned_at TIMESTAMPTZ,

    -- Watermark
    watermark_config JSONB,
    has_watermark BOOLEAN DEFAULT FALSE,

    -- Metadata and tags
    metadata JSONB DEFAULT '{}',
    tags TEXT[],

    -- Expiration
    expires_at TIMESTAMPTZ,
    is_expired BOOLEAN GENERATED ALWAYS AS (
        expires_at IS NOT NULL AND expires_at < NOW()
    ) STORED,

    -- Access control
    uploaded_by UUID NOT NULL,
    permissions JSONB DEFAULT '{}',

    -- Soft delete
    is_deleted BOOLEAN DEFAULT FALSE,
    deleted_at TIMESTAMPTZ,
    deleted_by UUID,

    -- Audit
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,

    CONSTRAINT unique_checksum_per_tenant UNIQUE(tenant_id, checksum)
);

-- Indexes for documents
CREATE INDEX idx_documents_tenant_id ON documents(tenant_id);
CREATE INDEX idx_documents_owner_entity ON documents(owner_entity_id);
CREATE INDEX idx_documents_file_name ON documents(file_name);
CREATE INDEX idx_documents_document_type ON documents(document_type);
CREATE INDEX idx_documents_tags ON documents USING GIN(tags);
CREATE INDEX idx_documents_metadata ON documents USING GIN(metadata);
CREATE INDEX idx_documents_created_at ON documents(created_at);
CREATE INDEX idx_documents_is_deleted ON documents(is_deleted) WHERE is_deleted = false;

-- Document shares table
CREATE TABLE document_shares (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,

    -- Share target (entity, user, or email)
    shared_with_entity_id UUID,
    shared_with_user_id UUID,
    shared_with_email VARCHAR(255),

    -- Organizational scope
    division_id UUID,
    branch_id UUID,
    department_id UUID,

    -- Permissions
    can_view BOOLEAN DEFAULT TRUE,
    can_download BOOLEAN DEFAULT FALSE,
    can_edit BOOLEAN DEFAULT FALSE,
    can_delete BOOLEAN DEFAULT FALSE,

    -- Share lifecycle
    share_link VARCHAR(255) UNIQUE,
    share_password VARCHAR(255),
    expires_at TIMESTAMPTZ,
    is_expired BOOLEAN GENERATED ALWAYS AS (
        expires_at IS NOT NULL AND expires_at < NOW()
    ) STORED,

    -- Access tracking
    access_count INTEGER DEFAULT 0,
    last_accessed_at TIMESTAMPTZ,

    -- Audit
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,

    CONSTRAINT share_target_check CHECK (
        shared_with_entity_id IS NOT NULL OR
        shared_with_user_id IS NOT NULL OR
        shared_with_email IS NOT NULL
    )
);

CREATE INDEX idx_document_shares_document_id ON document_shares(document_id);
CREATE INDEX idx_document_shares_entity ON document_shares(shared_with_entity_id);
```

### 4. Operations Domain (FormBuilder)

```sql
-- Forms table
CREATE TABLE forms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    version VARCHAR(50) NOT NULL DEFAULT '1.0.0',
    created_by VARCHAR(255) NOT NULL,
    allowed_roles JSONB DEFAULT '[]'::jsonb,
    audit BOOLEAN DEFAULT false,
    table_name VARCHAR(255),
    module VARCHAR(255),
    schema_version INTEGER DEFAULT 1,
    core_fields JSONB DEFAULT '[]'::jsonb,
    steps JSONB NOT NULL DEFAULT '[]'::jsonb,
    dependencies JSONB DEFAULT '[]'::jsonb,
    cross_field_validations JSONB DEFAULT '[]'::jsonb,
    workflow_id UUID,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT chk_forms_version_format CHECK (version ~ '^[0-9]+\.[0-9]+\.[0-9]+$')
);

CREATE INDEX idx_forms_workflow_id ON forms(workflow_id);
CREATE INDEX idx_forms_title ON forms(title);

-- Form instances table
CREATE TABLE form_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    form_id UUID NOT NULL REFERENCES forms(id) ON DELETE CASCADE,
    current_state VARCHAR(100) NOT NULL DEFAULT 'draft',
    field_values JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_by VARCHAR(255) NOT NULL,
    assigned_to VARCHAR(255),
    metadata JSONB DEFAULT '{}'::jsonb,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_form_instances_form_id ON form_instances(form_id);
CREATE INDEX idx_form_instances_current_state ON form_instances(current_state);

-- Workflows table
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

-- Approval actions table
CREATE TABLE approval_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    form_instance_id UUID NOT NULL REFERENCES form_instances(id) ON DELETE CASCADE,
    approver_id VARCHAR(255) NOT NULL,
    action VARCHAR(50) NOT NULL CHECK (
        action IN ('APPROVE', 'REJECT', 'DELEGATE', 'REQUEST_INFO', 'WITHDRAW', 'REASSIGN')
    ),
    comments TEXT,
    acted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    ip_address INET,
    user_agent TEXT,
    delegated_from VARCHAR(255),
    attachment_urls JSONB DEFAULT '[]'::jsonb,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_approval_actions_form_instance_id ON approval_actions(form_instance_id);
CREATE INDEX idx_approval_actions_approver_id ON approval_actions(approver_id);
```

---

## Index Strategy

### Index Types and Use Cases

```sql
-- 1. B-Tree Indexes (default, most common)
CREATE INDEX idx_users_email ON users(email);  -- Equality and range queries

-- 2. Partial Indexes (filtered, smaller, faster)
CREATE INDEX idx_users_active ON users(status) WHERE status = 'active';
CREATE INDEX idx_documents_not_deleted ON documents(id) WHERE deleted_at IS NULL;

-- 3. Composite Indexes (multi-column queries)
CREATE INDEX idx_documents_tenant_owner ON documents(tenant_id, owner_entity_id);
CREATE INDEX idx_form_instances_form_state ON form_instances(form_id, current_state);

-- 4. GIN Indexes (JSONB, arrays, full-text)
CREATE INDEX idx_documents_tags ON documents USING GIN(tags);
CREATE INDEX idx_documents_metadata ON documents USING GIN(metadata);
CREATE INDEX idx_forms_steps ON forms USING GIN(steps);

-- 5. GiST Indexes (geometric data, full-text search)
CREATE INDEX idx_branches_location ON branches USING GIST(
    ll_to_earth(latitude, longitude)
);

-- 6. Unique Indexes (enforce uniqueness)
CREATE UNIQUE INDEX idx_tenants_code ON tenants(code);
CREATE UNIQUE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;

-- 7. Expression Indexes (computed values)
CREATE INDEX idx_users_normalized_email ON users(LOWER(email));
CREATE INDEX idx_documents_year ON documents(EXTRACT(YEAR FROM created_at));
```

### Index Naming Convention

```
idx_{table}_{column(s)}[_{condition}]

Examples:
idx_users_email              -- Single column
idx_documents_tenant_owner   -- Multiple columns
idx_users_active             -- With condition
idx_documents_metadata       -- GIN index
```

### Index Maintenance

```sql
-- Check index usage
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes
WHERE idx_scan = 0
  AND indexname NOT LIKE '%_pkey'
ORDER BY tablename, indexname;

-- Rebuild bloated indexes
REINDEX INDEX idx_documents_tenant_id;
REINDEX TABLE documents;

-- Analyze table statistics
ANALYZE documents;
VACUUM ANALYZE documents;
```

---

## Data Integrity & Constraints

### Foreign Key Constraints

```sql
-- Cascade on delete (dependent data removed)
CREATE TABLE entity_role_bindings (
    entity_id UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    ...
);

-- Set null on delete (optional relationship)
CREATE TABLE divisions (
    head_user_id VARCHAR(255) REFERENCES users(uuid) ON DELETE SET NULL,
    ...
);

-- Restrict delete (prevent orphans)
CREATE TABLE branches (
    division_id UUID NOT NULL REFERENCES divisions(id) ON DELETE RESTRICT,
    ...
);
```

### Check Constraints

```sql
-- Enum-like constraints
ALTER TABLE tenants ADD CONSTRAINT chk_tenants_status
    CHECK (status IN ('ACTIVE', 'SUSPENDED', 'TRIAL', 'EXPIRED'));

-- Range constraints
ALTER TABLE attachments ADD CONSTRAINT chk_attachments_size
    CHECK (size >= 0 AND size <= 104857600);  -- Max 100MB

-- Format constraints
ALTER TABLE forms ADD CONSTRAINT chk_forms_version_format
    CHECK (version ~ '^[0-9]+\.[0-9]+\.[0-9]+$');

-- Conditional constraints
ALTER TABLE departments ADD CONSTRAINT chk_departments_hierarchy
    CHECK (
        (division_id IS NULL AND parent_department_id IS NULL) OR
        (division_id IS NOT NULL)
    );

-- Date range constraints
ALTER TABLE approval_delegates ADD CONSTRAINT chk_delegates_date_range
    CHECK (end_date IS NULL OR end_date > start_date);
```

### Unique Constraints

```sql
-- Simple unique
ALTER TABLE users ADD CONSTRAINT unique_users_email UNIQUE(email);

-- Composite unique
ALTER TABLE entities ADD CONSTRAINT unique_user_per_tenant
    UNIQUE(tenant_id, user_id);

-- Partial unique (with condition)
CREATE UNIQUE INDEX unique_active_users_email
    ON users(email) WHERE deleted_at IS NULL;

-- Unique with NULL handling
CREATE UNIQUE INDEX idx_departments_unique_code ON departments (
    tenant_id,
    code,
    COALESCE(division_id::text, '')  -- Handle NULL division_id
);
```

---

## Audit Trail Design

### Standard Audit Columns

Every table includes these audit columns:

```sql
CREATE TABLE {table_name} (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,

    -- Business columns...

    -- Audit trail (mandatory on all tables)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,  -- Reference to users.uuid
    updated_by UUID NOT NULL,  -- Reference to users.uuid
    deleted_at TIMESTAMPTZ     -- Soft delete timestamp
);
```

### Auto-Update Trigger

```sql
-- Function to auto-update updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply to all tables
CREATE TRIGGER update_divisions_updated_at
    BEFORE UPDATE ON divisions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_branches_updated_at
    BEFORE UPDATE ON branches
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Repeat for all tables...
```

### Detailed Audit Log Table

For critical operations, maintain a separate audit log:

```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,

    -- Actor information
    user_id UUID NOT NULL,
    entity_id UUID,

    -- Action details
    action VARCHAR(100) NOT NULL,  -- CREATE, UPDATE, DELETE, APPROVE, etc.
    resource_type VARCHAR(100) NOT NULL,  -- documents, forms, users, etc.
    resource_id UUID NOT NULL,

    -- Change tracking
    from_state VARCHAR(100),
    to_state VARCHAR(100),
    changes JSONB DEFAULT '{}',  -- {"field": {"from": "old", "to": "new"}}

    -- Context
    ip_address INET,
    user_agent TEXT,
    request_id UUID,

    -- Timestamp
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Indexes
    INDEX idx_audit_tenant_resource (tenant_id, resource_type, resource_id),
    INDEX idx_audit_user (user_id),
    INDEX idx_audit_timestamp (timestamp)
);

-- Example audit entry:
INSERT INTO audit_logs (
    tenant_id, user_id, action, resource_type, resource_id,
    from_state, to_state, changes, ip_address
) VALUES (
    'tenant-123',
    'user-456',
    'UPDATE',
    'form_instances',
    'instance-789',
    'pending',
    'approved',
    '{"status": {"from": "pending", "to": "approved"}, "approved_by": {"from": null, "to": "user-456"}}'::jsonb,
    '192.168.1.100'::inet
);
```

---

## Performance Optimization

### Query Optimization Techniques

```sql
-- 1. Use covering indexes (index-only scans)
CREATE INDEX idx_users_covering ON users(tenant_id, id, email, status);

SELECT id, email, status FROM users
WHERE tenant_id = $1;  -- Can use index-only scan

-- 2. Avoid SELECT * (fetch only needed columns)
-- ❌ Bad
SELECT * FROM documents WHERE tenant_id = $1;

-- ✅ Good
SELECT id, file_name, created_at FROM documents WHERE tenant_id = $1;

-- 3. Use LIMIT for pagination
SELECT * FROM documents
WHERE tenant_id = $1
ORDER BY created_at DESC
LIMIT 20 OFFSET 0;

-- 4. Use EXPLAIN ANALYZE to understand query plans
EXPLAIN ANALYZE
SELECT d.*, e.entity_type
FROM documents d
JOIN entities e ON e.id = d.owner_entity_id
WHERE d.tenant_id = $1;

-- 5. Materialized views for complex aggregations
CREATE MATERIALIZED VIEW daily_document_stats AS
SELECT
    tenant_id,
    DATE(created_at) as date,
    COUNT(*) as total_documents,
    SUM(size_bytes) as total_size
FROM documents
GROUP BY tenant_id, DATE(created_at);

CREATE UNIQUE INDEX idx_daily_doc_stats ON daily_document_stats(tenant_id, date);

-- Refresh periodically
REFRESH MATERIALIZED VIEW CONCURRENTLY daily_document_stats;
```

### Partitioning Strategy (Future)

For very large tables, consider partitioning:

```sql
-- Partition by tenant_id (multi-tenant isolation)
CREATE TABLE documents (
    id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    ...
) PARTITION BY HASH (tenant_id);

CREATE TABLE documents_p0 PARTITION OF documents
    FOR VALUES WITH (MODULUS 4, REMAINDER 0);
CREATE TABLE documents_p1 PARTITION OF documents
    FOR VALUES WITH (MODULUS 4, REMAINDER 1);
CREATE TABLE documents_p2 PARTITION OF documents
    FOR VALUES WITH (MODULUS 4, REMAINDER 2);
CREATE TABLE documents_p3 PARTITION OF documents
    FOR VALUES WITH (MODULUS 4, REMAINDER 3);

-- Partition by date (time-series data)
CREATE TABLE audit_logs (
    id UUID NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    ...
) PARTITION BY RANGE (timestamp);

CREATE TABLE audit_logs_2024_01 PARTITION OF audit_logs
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
CREATE TABLE audit_logs_2024_02 PARTITION OF audit_logs
    FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');
-- ...
```

### Connection Pooling

```go
// Application-level configuration
config := &pgxpool.Config{
    MaxConns:          10,               // Max connections per instance
    MinConns:          2,                // Min idle connections
    MaxConnIdleTime:   30 * time.Minute, // Close idle connections
    MaxConnLifetime:   2 * time.Hour,    // Recycle old connections
    HealthCheckPeriod: 1 * time.Minute,  // Check connection health
}
```

---

## Migration Strategy

### Migration Tools

**Primary Tool:** Atlas CLI (https://atlasgo.io/)

**Benefits:**
- Declarative schema definition
- Automatic migration generation
- Schema versioning
- Rollback support
- Multi-environment support

### Migration Workflow

```bash
# 1. Define schema in SQL files
# migrations/schemas/schema.sql

# 2. Generate migration
atlas migrate diff migration_name \
  --dir "file://migrations" \
  --to "file://migrations/schemas/schema.sql" \
  --dev-url "docker://postgres/15"

# 3. Review generated migration
# migrations/20240101120000_migration_name.sql

# 4. Apply migration
atlas migrate apply \
  --dir "file://migrations" \
  --url "postgres://user:pass@localhost:5432/ugcl"

# 5. Verify schema
atlas schema inspect \
  --url "postgres://user:pass@localhost:5432/ugcl"
```

### Migration File Structure

```
migrations/
├── schemas/
│   ├── identity.sql         # Identity domain schema
│   ├── organization.sql     # Organization domain schema
│   ├── content.sql          # Content domain schema
│   └── operations.sql       # Operations domain schema
├── 20240101120000_initial_schema.sql
├── 20240115100000_add_entities.sql
├── 20240201140000_add_document_shares.sql
└── atlas.sum                # Checksum file
```

### Example Migration

```sql
-- 20240115100000_add_entities.sql

-- Create entity types
CREATE TYPE entity_type AS ENUM (
    'EMPLOYEE', 'CONTRACTOR', 'VENDOR', 'CLIENT',
    'DEVICE', 'DRONE', 'BOT', 'SYSTEM'
);

-- Create entities table
CREATE TABLE entities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    entity_type entity_type NOT NULL,
    user_id UUID NOT NULL,
    reference_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL
);

-- Create indexes
CREATE INDEX idx_entities_tenant_id ON entities(tenant_id);
CREATE INDEX idx_entities_user_id ON entities(user_id);
CREATE INDEX idx_entities_reference ON entities(reference_id);

-- Add trigger
CREATE TRIGGER update_entities_updated_at
    BEFORE UPDATE ON entities
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

### Rollback Strategy

```bash
# Rollback last migration
atlas migrate down \
  --dir "file://migrations" \
  --url "postgres://user:pass@localhost:5432/ugcl"

# Rollback to specific version
atlas migrate down \
  --dir "file://migrations" \
  --url "postgres://user:pass@localhost:5432/ugcl" \
  --to-version 20240101120000
```

---

## Naming Conventions

### Tables

- **Plural nouns:** `users`, `documents`, `entities`
- **Lowercase with underscores:** `entity_role_bindings`, `approval_actions`
- **Avoid abbreviations:** `departments` not `depts`

### Columns

- **Singular nouns:** `user_id`, `tenant_id`, `created_at`
- **Lowercase with underscores:** `normalized_email`, `phone_verified`
- **Boolean prefix:** `is_active`, `has_watermark`, `can_edit`
- **Timestamp suffix:** `created_at`, `updated_at`, `deleted_at`

### Indexes

```
idx_{table}_{column(s)}[_{condition}]

Examples:
idx_users_email
idx_documents_tenant_owner
idx_users_active
```

### Constraints

```
{type}_{table}_{column(s)}

Types:
- pk_   (primary key)
- fk_   (foreign key)
- uk_   (unique)
- chk_  (check)

Examples:
pk_users
fk_entities_user_id
uk_tenants_code
chk_users_status
```

### Enums

- **Uppercase with underscores:** `USER_STATUS`, `ENTITY_TYPE`
- **Descriptive names:** `ENTITY_TYPE_EMPLOYEE` not `EMP`

---

## Conclusion

The UGCL Backend v2 database design follows modern PostgreSQL best practices with a strong focus on:

- **Multi-tenancy** with tenant isolation
- **Audit trails** for compliance
- **Performance** through strategic indexing
- **Data integrity** with constraints
- **Flexibility** using JSONB for extensibility
- **Type safety** with enums and check constraints

This design supports the complex business requirements of the UGCL platform while maintaining scalability, security, and maintainability.

---

**Document Version:** 2.0
**Total Tables:** 50+
**Total Indexes:** 200+
**Lines:** 1400+
**Generated:** 2025-10-06
**Maintained By:** UGCL Database Team
