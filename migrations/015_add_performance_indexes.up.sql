-- Performance indexes for zeno-auth
-- Add indexes concurrently to avoid locking tables in production

-- Users table
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_email_active ON users(email) WHERE is_active = true;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_created_at ON users(created_at DESC);

-- Organizations table
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_organizations_status ON organizations(status);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_organizations_created_at ON organizations(created_at DESC);

-- Org memberships table
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_org_memberships_user_org ON org_memberships(user_id, org_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_org_memberships_org_role ON org_memberships(org_id, role);

-- Refresh tokens table
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at) WHERE revoked_at IS NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_refresh_tokens_fingerprint ON refresh_tokens(fingerprint);

-- Email verifications table
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_email_verifications_user_id ON email_verifications(user_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_email_verifications_expires_at ON email_verifications(expires_at) WHERE verified_at IS NULL;

-- Audit logs table
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_user_id_created ON audit_logs(user_id, created_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_action_created ON audit_logs(action, created_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC);

-- Password reset tokens table
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_password_reset_user_id ON password_reset_tokens(user_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_password_reset_expires_at ON password_reset_tokens(expires_at) WHERE used_at IS NULL;

-- User consents table
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_consents_user_type ON user_consents(user_id, consent_type);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_consents_created_at ON user_consents(created_at DESC);
