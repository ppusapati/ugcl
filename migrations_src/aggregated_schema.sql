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


