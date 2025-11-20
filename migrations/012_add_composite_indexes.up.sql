-- Composite index for refresh token lookup (user_id + token_hash)
-- Optimizes: SELECT * FROM refresh_tokens WHERE user_id = ? AND token_hash = ?
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_token ON refresh_tokens(user_id, token_hash);

-- Composite index for audit log queries (user_id + created_at)
-- Optimizes: SELECT * FROM audit_logs WHERE user_id = ? ORDER BY created_at DESC
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_created ON audit_logs(user_id, created_at DESC);

-- Index for cleanup queries (expires_at + revoked_at)
-- Optimizes: DELETE FROM refresh_tokens WHERE expires_at < NOW() OR revoked_at IS NOT NULL
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_cleanup ON refresh_tokens(expires_at, revoked_at);
