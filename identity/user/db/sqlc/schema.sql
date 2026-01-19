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
-- User Status Enum to match proto
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_status') THEN
        CREATE TYPE user_status AS ENUM (
            'unspecified',
            'active',
            'inactive',
            'suspended',
            'pending_verification',
            'locked'
        );
    END IF;
END
$$;

-- Gender Enum to match proto
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'gender_type') THEN
        CREATE TYPE gender_type AS ENUM ('unknown', 'male', 'female', 'other');
    END IF;
END
$$;

CREATE TABLE IF NOT EXISTS users (
    id              BIGSERIAL PRIMARY KEY,
    uuid            UUID  UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    username        TEXT,
    normalized_username TEXT,
    fullname        TEXT,
    phone           TEXT,
    phone_verified  BOOLEAN DEFAULT false,  -- Renamed to match proto
    email           TEXT,
    normalized_email TEXT,
    email_verified  BOOLEAN DEFAULT false,  -- Renamed to match proto
    -- Removed password field (should never be stored, only password_hash)
    password_hash   TEXT NOT NULL,
    gender          gender_type DEFAULT 'unknown',  -- Now uses enum
    avatar          BYTEA,
    two_factor_enabled BOOLEAN DEFAULT false,
    salt            BYTEA,
    two_factor_secret TEXT,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    status          user_status DEFAULT 'active',  -- New field to match proto
    roles_cache     TEXT[],          -- renamed from roles to avoid junction clash
    created_at      TIMESTAMP NOT NULL DEFAULT now(),
    updated_at      TIMESTAMP NOT NULL DEFAULT now(),
    last_login      TIMESTAMP,       -- New field to match proto
    deleted_at      TIMESTAMP
);

CREATE INDEX IF NOT EXISTS ix_users_username ON users(normalized_username);
CREATE INDEX IF NOT EXISTS ix_users_email    ON users(normalized_email);
CREATE INDEX IF NOT EXISTS ix_users_status ON users(status);
CREATE INDEX IF NOT EXISTS ix_users_last_login ON users(last_login);

/* ===========================================================
   USER TENANT ROLES (New table for multi-tenant support)
   ===========================================================*/
CREATE TABLE IF NOT EXISTS user_tenant_roles (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id   TEXT NOT NULL,  -- Reference to tenant service
    roles       TEXT[] NOT NULL DEFAULT '{}',  -- Array of role names for this tenant
    is_active   BOOLEAN DEFAULT true,
    assigned_at TIMESTAMP NOT NULL DEFAULT now(),
    created_at  TIMESTAMP NOT NULL DEFAULT now(),
    updated_at  TIMESTAMP NOT NULL DEFAULT now(),
    UNIQUE(user_id, tenant_id)
);

CREATE INDEX IF NOT EXISTS ix_user_tenant_roles_user_id ON user_tenant_roles(user_id);
CREATE INDEX IF NOT EXISTS ix_user_tenant_roles_tenant_id ON user_tenant_roles(tenant_id);
CREATE INDEX IF NOT EXISTS ix_user_tenant_roles_is_active ON user_tenant_roles(is_active);

/* ===========================================================
   USER PREFERENCES (To match proto UserPreferences)
   ===========================================================*/
CREATE TABLE IF NOT EXISTS user_preferences (
    id                  BIGSERIAL PRIMARY KEY,
    user_id            BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    language           TEXT DEFAULT 'en',
    timezone           TEXT DEFAULT 'UTC',
    date_format        TEXT DEFAULT 'YYYY-MM-DD',
    time_format        TEXT DEFAULT '24h',
    email_notifications BOOLEAN DEFAULT true,
    sms_notifications  BOOLEAN DEFAULT false,
    push_notifications BOOLEAN DEFAULT true,
    custom_preferences JSONB DEFAULT '{}',
    created_at         TIMESTAMP NOT NULL DEFAULT now(),
    updated_at         TIMESTAMP NOT NULL DEFAULT now(),
    UNIQUE(user_id)
);

CREATE INDEX IF NOT EXISTS ix_user_preferences_user_id ON user_preferences(user_id);

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
