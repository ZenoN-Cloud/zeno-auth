-- Additional security indexes for performance

-- Composite index for refresh token lookup (most common query)
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_token ON refresh_tokens(user_id, token_hash) WHERE revoked_at IS NULL;

-- Composite index for audit log queries
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_created ON audit_logs(user_id, created_at DESC);

-- Index for cleanup queries
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_cleanup ON refresh_tokens(expires_at) WHERE revoked_at IS NULL;
