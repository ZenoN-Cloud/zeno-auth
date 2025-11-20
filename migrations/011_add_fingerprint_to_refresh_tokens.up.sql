ALTER TABLE refresh_tokens ADD COLUMN fingerprint_hash TEXT;
ALTER TABLE refresh_tokens ADD COLUMN device_info TEXT;
ALTER TABLE refresh_tokens ADD COLUMN last_used_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();

CREATE INDEX idx_refresh_tokens_fingerprint ON refresh_tokens(fingerprint_hash);
