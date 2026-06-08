-- Leads table: job opportunities sent to handymen
-- A single job can generate multiple leads (one per matching handyman)

CREATE TABLE IF NOT EXISTS leads (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id          UUID NOT NULL REFERENCES jobs(id),
    handyman_id     UUID NOT NULL REFERENCES users(id),

    status          VARCHAR(20) NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'accepted', 'rejected', 'expired')),

    -- Pricing (snapshot at creation time)
    price           INTEGER NOT NULL,

    -- Client quality snapshot
    client_commit_score INTEGER NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL,
    accepted_at     TIMESTAMPTZ,
    rejected_at     TIMESTAMPTZ,

    -- Prevent duplicate leads for same job+handyman
    UNIQUE(job_id, handyman_id)
);

-- Indexes for common queries
CREATE INDEX idx_leads_job ON leads(job_id);
CREATE INDEX idx_leads_handyman ON leads(handyman_id);
CREATE INDEX idx_leads_status ON leads(status);
CREATE INDEX idx_leads_handyman_pending ON leads(handyman_id, status) WHERE status = 'pending';
CREATE INDEX idx_leads_handyman_accepted ON leads(handyman_id, status) WHERE status = 'accepted';
CREATE INDEX idx_leads_expires ON leads(expires_at) WHERE status = 'pending';
CREATE INDEX idx_leads_created_at ON leads(created_at DESC);

-- Reuse the update_updated_at_column() function
CREATE TRIGGER update_leads_updated_at
    BEFORE UPDATE ON leads
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE leads IS 'Job leads sent to handymen - the core monetization entity';
COMMENT ON COLUMN leads.price IS 'Cost in credits to accept this lead (dynamic pricing snapshot)';
COMMENT ON COLUMN leads.client_commit_score IS 'Client Commit Score at time of lead creation';
COMMENT ON COLUMN leads.expires_at IS 'Lead expires if handyman does not act by this time';
