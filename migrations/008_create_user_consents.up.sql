CREATE TABLE user_consents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    consent_type TEXT NOT NULL CHECK (consent_type IN ('terms', 'privacy', 'marketing', 'analytics')),
    version TEXT NOT NULL,
    granted BOOLEAN NOT NULL DEFAULT true,
    granted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_consents_user_id ON user_consents(user_id);
CREATE INDEX idx_user_consents_type ON user_consents(consent_type);
CREATE UNIQUE INDEX idx_user_consents_active ON user_consents(user_id, consent_type, version) WHERE revoked_at IS NULL;
