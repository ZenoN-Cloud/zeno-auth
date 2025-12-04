-- Rollback performance indexes

DROP INDEX CONCURRENTLY IF EXISTS idx_users_email_active;
DROP INDEX CONCURRENTLY IF EXISTS idx_users_created_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_organizations_status;
DROP INDEX CONCURRENTLY IF EXISTS idx_organizations_created_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_org_memberships_user_org;
DROP INDEX CONCURRENTLY IF EXISTS idx_org_memberships_org_role;
DROP INDEX CONCURRENTLY IF EXISTS idx_refresh_tokens_user_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_refresh_tokens_expires_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_refresh_tokens_fingerprint;
DROP INDEX CONCURRENTLY IF EXISTS idx_email_verifications_user_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_email_verifications_expires_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_audit_logs_user_id_created;
DROP INDEX CONCURRENTLY IF EXISTS idx_audit_logs_action_created;
DROP INDEX CONCURRENTLY IF EXISTS idx_audit_logs_created_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_password_reset_user_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_password_reset_expires_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_user_consents_user_type;
DROP INDEX CONCURRENTLY IF EXISTS idx_user_consents_created_at;
