-- Users table for authentication
-- Supports SSO providers (Google, Facebook) and optional email/password

CREATE TABLE IF NOT EXISTS users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           VARCHAR(255) NOT NULL UNIQUE,
    name            VARCHAR(255) NOT NULL,
    role            VARCHAR(50) NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'handyman', 'admin')),
    
    -- Authentication
    provider        VARCHAR(50) NOT NULL CHECK (provider IN ('google', 'facebook', 'email')),
    provider_id     VARCHAR(255),                    -- External provider's user ID
    password_hash   VARCHAR(255),                    -- Only for email provider (future)
    
    -- Profile
    avatar_url      VARCHAR(500),
    phone           VARCHAR(50),
    
    -- Status
    is_active       BOOLEAN NOT NULL DEFAULT true,
    email_verified  BOOLEAN NOT NULL DEFAULT false,
    
    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at   TIMESTAMPTZ
);

-- Indexes for common queries
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_provider ON users(provider, provider_id);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_is_active ON users(is_active) WHERE is_active = true;

-- Updated_at trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Comments for documentation
COMMENT ON TABLE users IS 'Application users with SSO and role-based access';
COMMENT ON COLUMN users.provider IS 'Authentication provider: google, facebook, or email';
COMMENT ON COLUMN users.provider_id IS 'User ID from the OAuth provider';
COMMENT ON COLUMN users.role IS 'User role for RBAC: user, handyman, or admin';


