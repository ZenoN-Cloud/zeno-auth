DROP INDEX IF EXISTS idx_organizations_subscription_id;
DROP INDEX IF EXISTS idx_organizations_trial_ends_at;

ALTER TABLE organizations
DROP COLUMN IF EXISTS subscription_id,
DROP COLUMN IF EXISTS trial_ends_at;

ALTER TABLE organizations 
DROP CONSTRAINT IF EXISTS organizations_status_check;

ALTER TABLE organizations 
ADD CONSTRAINT organizations_status_check 
CHECK (status IN ('active', 'trial', 'suspended'));
