-- Add performance indexes for audit logs
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_action 
ON audit_logs(user_id, action, created_at DESC);

-- Add index for refresh tokens cleanup
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires 
ON refresh_tokens(expires_at) WHERE revoked_at IS NULL;
