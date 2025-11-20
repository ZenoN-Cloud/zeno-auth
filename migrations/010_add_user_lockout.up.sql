ALTER TABLE users ADD COLUMN failed_login_attempts INT DEFAULT 0;
ALTER TABLE users ADD COLUMN locked_until TIMESTAMP WITH TIME ZONE;

CREATE INDEX idx_users_locked_until ON users(locked_until);
