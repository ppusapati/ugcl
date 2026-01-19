-- Authentication Schema for Multi-tenant System
-- This schema supports authentication, session management, and multi-tenant access control

-- Users table (references tenant users)
CREATE TABLE auth_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL, -- Reference to identity.user service
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    password_hash VARCHAR(255) NOT NULL,
    password_salt VARCHAR(255) NOT NULL,
    password_changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    password_expires_at TIMESTAMP,
    failed_login_attempts INTEGER DEFAULT 0,
    account_locked_until TIMESTAMP,
    is_active BOOLEAN DEFAULT true,
    email_verified BOOLEAN DEFAULT false,
    phone_verified BOOLEAN DEFAULT false,
    two_factor_enabled BOOLEAN DEFAULT false,
    two_factor_secret VARCHAR(255),
    backup_codes TEXT[], -- Array of backup codes
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP,
    last_password_change TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- User Tenants mapping (which tenants a user belongs to)
CREATE TABLE auth_user_tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES auth_users(id) ON DELETE CASCADE,
    tenant_id VARCHAR(255) NOT NULL,
    is_default BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    roles TEXT[], -- Array of role IDs for this tenant
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, tenant_id)
);

-- Sessions table for tracking user sessions
CREATE TABLE auth_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id VARCHAR(255) UNIQUE NOT NULL,
    user_id UUID NOT NULL REFERENCES auth_users(id) ON DELETE CASCADE,
    tenant_id VARCHAR(255),
    refresh_token_hash VARCHAR(255) NOT NULL,
    device_info JSONB,
    ip_address INET,
    user_agent TEXT,
    is_active BOOLEAN DEFAULT true,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_accessed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMP,
    revoked_reason VARCHAR(255)
);

-- Password reset tokens
CREATE TABLE auth_password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES auth_users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ip_address INET,
    user_agent TEXT
);

-- Email verification tokens
CREATE TABLE auth_email_verification_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES auth_users(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    verified_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Phone verification tokens (OTP)
CREATE TABLE auth_phone_verification_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES auth_users(id) ON DELETE CASCADE,
    phone VARCHAR(50) NOT NULL,
    otp_code VARCHAR(10) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    verified_at TIMESTAMP,
    attempts INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Two-factor authentication backup codes
CREATE TABLE auth_two_factor_backup_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES auth_users(id) ON DELETE CASCADE,
    code_hash VARCHAR(255) NOT NULL,
    used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Login attempts tracking for security
CREATE TABLE auth_login_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    identifier VARCHAR(255) NOT NULL, -- username, email, or phone
    ip_address INET NOT NULL,
    user_agent TEXT,
    success BOOLEAN NOT NULL,
    failure_reason VARCHAR(255),
    tenant_id VARCHAR(255),
    user_id UUID REFERENCES auth_users(id),
    attempted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Security events log
CREATE TABLE auth_security_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES auth_users(id),
    event_type VARCHAR(100) NOT NULL, -- 'login', 'logout', 'password_change', 'two_factor_enable', etc.
    event_data JSONB,
    ip_address INET,
    user_agent TEXT,
    tenant_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Revoked tokens (blacklist)
CREATE TABLE auth_revoked_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token_jti VARCHAR(255) UNIQUE NOT NULL, -- JWT ID
    token_type VARCHAR(50) NOT NULL, -- 'access', 'refresh'
    user_id UUID REFERENCES auth_users(id),
    revoked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    reason VARCHAR(255)
);

-- API Keys for service-to-service authentication
CREATE TABLE auth_api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_id VARCHAR(255) UNIQUE NOT NULL,
    key_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255),
    scopes TEXT[], -- Array of allowed scopes/permissions
    is_active BOOLEAN DEFAULT true,
    last_used_at TIMESTAMP,
    expires_at TIMESTAMP,
    created_by UUID REFERENCES auth_users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_auth_users_username ON auth_users(username);
CREATE INDEX idx_auth_users_email ON auth_users(email);
CREATE INDEX idx_auth_users_phone ON auth_users(phone);
CREATE INDEX idx_auth_users_user_id ON auth_users(user_id);
CREATE INDEX idx_auth_users_is_active ON auth_users(is_active);

CREATE INDEX idx_auth_user_tenants_user_id ON auth_user_tenants(user_id);
CREATE INDEX idx_auth_user_tenants_tenant_id ON auth_user_tenants(tenant_id);
CREATE INDEX idx_auth_user_tenants_is_default ON auth_user_tenants(is_default);

CREATE INDEX idx_auth_sessions_session_id ON auth_sessions(session_id);
CREATE INDEX idx_auth_sessions_user_id ON auth_sessions(user_id);
CREATE INDEX idx_auth_sessions_tenant_id ON auth_sessions(tenant_id);
CREATE INDEX idx_auth_sessions_is_active ON auth_sessions(is_active);
CREATE INDEX idx_auth_sessions_expires_at ON auth_sessions(expires_at);

CREATE INDEX idx_auth_password_reset_tokens_token_hash ON auth_password_reset_tokens(token_hash);
CREATE INDEX idx_auth_password_reset_tokens_user_id ON auth_password_reset_tokens(user_id);
CREATE INDEX idx_auth_password_reset_tokens_expires_at ON auth_password_reset_tokens(expires_at);

CREATE INDEX idx_auth_login_attempts_identifier ON auth_login_attempts(identifier);
CREATE INDEX idx_auth_login_attempts_ip_address ON auth_login_attempts(ip_address);
CREATE INDEX idx_auth_login_attempts_attempted_at ON auth_login_attempts(attempted_at);

CREATE INDEX idx_auth_security_events_user_id ON auth_security_events(user_id);
CREATE INDEX idx_auth_security_events_event_type ON auth_security_events(event_type);
CREATE INDEX idx_auth_security_events_created_at ON auth_security_events(created_at);

CREATE INDEX idx_auth_revoked_tokens_token_jti ON auth_revoked_tokens(token_jti);
CREATE INDEX idx_auth_revoked_tokens_expires_at ON auth_revoked_tokens(expires_at);

CREATE INDEX idx_auth_api_keys_key_id ON auth_api_keys(key_id);
CREATE INDEX idx_auth_api_keys_tenant_id ON auth_api_keys(tenant_id);
CREATE INDEX idx_auth_api_keys_is_active ON auth_api_keys(is_active);

-- Triggers for updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_auth_users_updated_at BEFORE UPDATE ON auth_users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_auth_user_tenants_updated_at BEFORE UPDATE ON auth_user_tenants FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_auth_api_keys_updated_at BEFORE UPDATE ON auth_api_keys FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();