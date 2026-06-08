-- Jobs table: client service requests

CREATE TABLE IF NOT EXISTS jobs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id       UUID NOT NULL REFERENCES users(id),
    category_id     UUID NOT NULL REFERENCES service_categories(id),
    district_id     UUID NOT NULL REFERENCES districts(id),

    title           VARCHAR(200) NOT NULL,
    description     TEXT NOT NULL,
    urgency         VARCHAR(20) NOT NULL DEFAULT 'normal'
                    CHECK (urgency IN ('low', 'normal', 'urgent', 'emergency')),
    status          VARCHAR(20) NOT NULL DEFAULT 'draft'
                    CHECK (status IN ('draft', 'active', 'accepted', 'in_progress', 'done', 'cancelled')),

    -- Location details
    address         VARCHAR(500),
    building_type   VARCHAR(50),
    floor           INTEGER,
    has_elevator    BOOLEAN,

    -- Scheduling
    preferred_date1 TIMESTAMPTZ,
    preferred_date2 TIMESTAMPTZ,
    preferred_time  VARCHAR(20) CHECK (preferred_time IN ('morning', 'afternoon', 'evening', 'flexible')),

    -- Budget
    budget          INTEGER,
    wants_invoice   BOOLEAN NOT NULL DEFAULT false,

    -- Contact
    contact_method  VARCHAR(20) NOT NULL DEFAULT 'any'
                    CHECK (contact_method IN ('phone', 'app', 'any')),

    -- Photos (JSON array of URLs)
    photo_urls      JSONB DEFAULT '[]'::jsonb,

    -- Completion
    final_value       INTEGER,
    completed_at      TIMESTAMPTZ,
    completed_by_id   UUID REFERENCES users(id),
    client_confirmed  BOOLEAN,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ
);

-- Indexes for common queries
CREATE INDEX idx_jobs_client ON jobs(client_id);
CREATE INDEX idx_jobs_category ON jobs(category_id);
CREATE INDEX idx_jobs_district ON jobs(district_id);
CREATE INDEX idx_jobs_status ON jobs(status);
CREATE INDEX idx_jobs_active ON jobs(status, district_id, category_id) WHERE status = 'active';
CREATE INDEX idx_jobs_client_status ON jobs(client_id, status);
CREATE INDEX idx_jobs_completed_by ON jobs(completed_by_id) WHERE completed_by_id IS NOT NULL;
CREATE INDEX idx_jobs_created_at ON jobs(created_at DESC);

-- Reuse the update_updated_at_column() function from migration 000001
CREATE TRIGGER update_jobs_updated_at
    BEFORE UPDATE ON jobs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE jobs IS 'Client service requests (the core marketplace entity)';
COMMENT ON COLUMN jobs.status IS 'Job lifecycle: draft -> active -> accepted -> in_progress -> done/cancelled';
COMMENT ON COLUMN jobs.final_value IS 'Declared value in PLN after job completion (for DAC7 reporting)';
COMMENT ON COLUMN jobs.budget IS 'Optional client budget in PLN';
COMMENT ON COLUMN jobs.photo_urls IS 'JSON array of photo URLs uploaded by client';
