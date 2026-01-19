-- Tenant Management Schema for Multi-tenant System
-- This schema supports tenant creation, management, and configuration

-- Tenant DB Type Enum to match proto
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'tenant_db_type') THEN
        CREATE TYPE tenant_db_type AS ENUM ('shared', 'seperatedb', 'seperateschema');
    END IF;
END
$$;

-- Tenants table - Core tenant information
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) UNIQUE NOT NULL, -- Unique tenant identifier/subdomain
    display_name VARCHAR(255) NOT NULL, -- Human-readable name
    region VARCHAR(100) NOT NULL, -- Geographic region
    logo TEXT, -- Logo URL or base64 data
    tenant_db tenant_db_type DEFAULT 'shared', -- Now uses enum to match proto
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by UUID, -- Reference to user who created the tenant
    subscription_plan VARCHAR(100), -- e.g., 'basic', 'premium', 'enterprise'
    subscription_expires_at TIMESTAMP,
    max_users INTEGER DEFAULT 100,
    max_storage_gb INTEGER DEFAULT 10
);

-- Tenant connection strings for separate database configurations
CREATE TABLE tenant_connection_strings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    key VARCHAR(255) NOT NULL, -- e.g., 'primary_db', 'read_replica', 'cache'
    value TEXT NOT NULL, -- Connection string value
    is_encrypted BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, key)
);

-- Tenant features/flags for feature toggling per tenant
CREATE TABLE tenant_features (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    key VARCHAR(255) NOT NULL, -- Feature flag name
    value TEXT NOT NULL, -- Feature value (boolean, string, json)
    value_type VARCHAR(50) DEFAULT 'string', -- 'boolean', 'string', 'json', 'number'
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, key)
);

-- Tenant metadata for custom fields and configuration
CREATE TABLE tenant_metadata (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    data JSONB NOT NULL DEFAULT '{}', -- Custom tenant data
    schema_version INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tenant domains for custom domain support
CREATE TABLE tenant_domains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    domain VARCHAR(255) UNIQUE NOT NULL, -- Custom domain
    is_primary BOOLEAN DEFAULT false,
    is_verified BOOLEAN DEFAULT false,
    ssl_certificate TEXT, -- SSL certificate data
    ssl_private_key TEXT, -- SSL private key (encrypted)
    verification_token VARCHAR(255), -- Domain verification token
    verified_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tenant admin users (initial admin users for tenant setup)
CREATE TABLE tenant_admin_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL, -- Reference to identity.user service
    username VARCHAR(255),
    email VARCHAR(255),
    is_primary BOOLEAN DEFAULT false, -- Primary admin for the tenant
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, user_id)
);

-- Tenant database schemas (for SEPERATESCHEMA mode)
CREATE TABLE tenant_database_schemas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    schema_name VARCHAR(255) NOT NULL,
    database_name VARCHAR(255) NOT NULL,
    connection_string TEXT NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    migration_version VARCHAR(100), -- Track migration state
    last_migration_at TIMESTAMP,
    UNIQUE(tenant_id)
);

-- Tenant usage metrics for billing and monitoring
CREATE TABLE tenant_usage_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    metric_name VARCHAR(255) NOT NULL, -- e.g., 'storage_used', 'api_calls', 'active_users'
    metric_value BIGINT NOT NULL,
    metric_unit VARCHAR(50), -- e.g., 'bytes', 'count', 'milliseconds'
    recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    period_start TIMESTAMP NOT NULL,
    period_end TIMESTAMP NOT NULL
);

-- Tenant billing information
CREATE TABLE tenant_billing (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    billing_email VARCHAR(255) NOT NULL,
    payment_method_id VARCHAR(255), -- Reference to payment provider
    subscription_id VARCHAR(255), -- Reference to subscription provider
    billing_address JSONB,
    tax_id VARCHAR(100),
    billing_currency VARCHAR(3) DEFAULT 'USD',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id)
);

-- Tenant audit log for tracking changes
CREATE TABLE tenant_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    action VARCHAR(100) NOT NULL, -- 'created', 'updated', 'deleted', 'activated', 'deactivated'
    actor_id VARCHAR(255), -- User who performed the action
    actor_type VARCHAR(50) DEFAULT 'user', -- 'user', 'system', 'api'
    changes JSONB, -- What changed (before/after values)
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_tenants_name ON tenants(name);
CREATE INDEX idx_tenants_region ON tenants(region);
CREATE INDEX idx_tenants_is_active ON tenants(is_active);
CREATE INDEX idx_tenants_created_at ON tenants(created_at);

CREATE INDEX idx_tenant_connection_strings_tenant_id ON tenant_connection_strings(tenant_id);
CREATE INDEX idx_tenant_connection_strings_key ON tenant_connection_strings(key);

CREATE INDEX idx_tenant_features_tenant_id ON tenant_features(tenant_id);
CREATE INDEX idx_tenant_features_key ON tenant_features(key);

CREATE INDEX idx_tenant_metadata_tenant_id ON tenant_metadata(tenant_id);
CREATE INDEX idx_tenant_metadata_data ON tenant_metadata USING GIN(data);

CREATE INDEX idx_tenant_domains_tenant_id ON tenant_domains(tenant_id);
CREATE INDEX idx_tenant_domains_domain ON tenant_domains(domain);
CREATE INDEX idx_tenant_domains_is_primary ON tenant_domains(is_primary);
CREATE INDEX idx_tenant_domains_is_verified ON tenant_domains(is_verified);

CREATE INDEX idx_tenant_admin_users_tenant_id ON tenant_admin_users(tenant_id);
CREATE INDEX idx_tenant_admin_users_user_id ON tenant_admin_users(user_id);

CREATE INDEX idx_tenant_database_schemas_tenant_id ON tenant_database_schemas(tenant_id);
CREATE INDEX idx_tenant_database_schemas_is_active ON tenant_database_schemas(is_active);

CREATE INDEX idx_tenant_usage_metrics_tenant_id ON tenant_usage_metrics(tenant_id);
CREATE INDEX idx_tenant_usage_metrics_metric_name ON tenant_usage_metrics(metric_name);
CREATE INDEX idx_tenant_usage_metrics_recorded_at ON tenant_usage_metrics(recorded_at);

CREATE INDEX idx_tenant_billing_tenant_id ON tenant_billing(tenant_id);

CREATE INDEX idx_tenant_audit_log_tenant_id ON tenant_audit_log(tenant_id);
CREATE INDEX idx_tenant_audit_log_action ON tenant_audit_log(action);
CREATE INDEX idx_tenant_audit_log_created_at ON tenant_audit_log(created_at);

-- Constraints (removed old integer check since we now use enum)
ALTER TABLE tenant_features ADD CONSTRAINT check_value_type CHECK (value_type IN ('boolean', 'string', 'json', 'number'));

-- Triggers for updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_tenants_updated_at BEFORE UPDATE ON tenants FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_tenant_connection_strings_updated_at BEFORE UPDATE ON tenant_connection_strings FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_tenant_features_updated_at BEFORE UPDATE ON tenant_features FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_tenant_metadata_updated_at BEFORE UPDATE ON tenant_metadata FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_tenant_domains_updated_at BEFORE UPDATE ON tenant_domains FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_tenant_billing_updated_at BEFORE UPDATE ON tenant_billing FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Views for common queries
CREATE OR REPLACE VIEW tenant_summary AS
SELECT
    t.id,
    t.name,
    t.display_name,
    t.region,
    t.tenant_db,
    t.is_active,
    t.created_at,
    t.subscription_plan,
    t.subscription_expires_at,
    COUNT(DISTINCT tau.user_id) as admin_count,
    COUNT(DISTINCT td.domain) as domain_count,
    COUNT(DISTINCT tf.key) as feature_count
FROM tenants t
LEFT JOIN tenant_admin_users tau ON t.id = tau.tenant_id
LEFT JOIN tenant_domains td ON t.id = td.tenant_id AND td.is_verified = true
LEFT JOIN tenant_features tf ON t.id = tf.tenant_id
GROUP BY t.id, t.name, t.display_name, t.region, t.tenant_db, t.is_active, t.created_at, t.subscription_plan, t.subscription_expires_at;