-- Remove nullable foreign key
ALTER TABLE refresh_tokens DROP CONSTRAINT IF EXISTS refresh_tokens_org_id_fkey;

-- Make org_id NOT NULL again
ALTER TABLE refresh_tokens ALTER COLUMN org_id SET NOT NULL;

-- Add back original foreign key
ALTER TABLE refresh_tokens ADD CONSTRAINT refresh_tokens_org_id_fkey 
    FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE;
