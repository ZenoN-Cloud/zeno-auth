ALTER TABLE refresh_tokens DROP COLUMN IF EXISTS fingerprint_hash;
ALTER TABLE refresh_tokens DROP COLUMN IF EXISTS device_info;
ALTER TABLE refresh_tokens DROP COLUMN IF EXISTS last_used_at;
