-- Update organization status to match billing states
ALTER TABLE organizations 
DROP CONSTRAINT IF EXISTS organizations_status_check;

ALTER TABLE organizations 
ADD CONSTRAINT organizations_status_check 
CHECK (status IN ('created', 'trialing', 'active', 'past_due', 'canceled'));

-- Add trial tracking fields
ALTER TABLE organizations
ADD COLUMN trial_ends_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN subscription_id UUID;

CREATE INDEX idx_organizations_trial_ends_at ON organizations(trial_ends_at) WHERE trial_ends_at IS NOT NULL;
CREATE INDEX idx_organizations_subscription_id ON organizations(subscription_id) WHERE subscription_id IS NOT NULL;
