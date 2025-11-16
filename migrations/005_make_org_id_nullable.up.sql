-- Drop foreign key constraint
ALTER TABLE refresh_tokens DROP CONSTRAINT IF EXISTS refresh_tokens_org_id_fkey;

-- Make org_id nullable
ALTER TABLE refresh_tokens ALTER COLUMN org_id DROP NOT NULL;

-- Add back foreign key but allow NULL
ALTER TABLE refresh_tokens ADD CONSTRAINT refresh_tokens_org_id_fkey 
    FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE;
