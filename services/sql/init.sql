-- AgentMesh Public Schema Initialization
-- Database: agentmesh_public

-- Create database (run separately if needed)
-- CREATE DATABASE agentmesh_public;

-- Tenants table
CREATE TABLE IF NOT EXISTS tenants (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    schema_name VARCHAR(63) NOT NULL UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'deleted')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tenants_schema_name ON tenants(schema_name);
CREATE INDEX idx_tenants_status ON tenants(status);

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    tenant_id INTEGER REFERENCES tenants(id) ON DELETE SET NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user' CHECK (role IN ('platform_admin', 'tenant_admin', 'user')),
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended', 'deleted')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_tenant_id ON users(tenant_id);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_status ON users(status);

-- Plugins registry table
CREATE TABLE IF NOT EXISTS plugins (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    version VARCHAR(50) NOT NULL,
    description TEXT,
    plugin_type VARCHAR(20) NOT NULL CHECK (plugin_type IN ('backend', 'frontend', 'full')),
    backend_path VARCHAR(500),
    frontend_entry VARCHAR(500),
    config_schema JSONB DEFAULT '{}',
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'deprecated')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_plugins_name ON plugins(name);
CREATE INDEX idx_plugins_type ON plugins(plugin_type);
CREATE INDEX idx_plugins_status ON plugins(status);

-- Tenant-Plugin association table
CREATE TABLE IF NOT EXISTS tenant_plugins (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    plugin_id INTEGER NOT NULL REFERENCES plugins(id) ON DELETE CASCADE,
    config JSONB DEFAULT '{}',
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'error')),
    installed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, plugin_id)
);

CREATE INDEX idx_tenant_plugins_tenant_id ON tenant_plugins(tenant_id);
CREATE INDEX idx_tenant_plugins_plugin_id ON tenant_plugins(plugin_id);
CREATE INDEX idx_tenant_plugins_status ON tenant_plugins(status);

-- Insert default admin user
-- Password is 'admin123' (bcrypt hash)
INSERT INTO users (username, password, email, tenant_id, role, status)
VALUES (
    'admin',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
    'admin@agentmesh.io',
    NULL,
    'platform_admin',
    'active'
) ON CONFLICT (username) DO NOTHING;

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for updated_at
CREATE TRIGGER update_tenants_updated_at
    BEFORE UPDATE ON tenants
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_plugins_updated_at
    BEFORE UPDATE ON plugins
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
