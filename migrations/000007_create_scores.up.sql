-- Scoring system: Commit Score (clients) and ProScore (handymen)

CREATE TABLE IF NOT EXISTS commit_scores (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL UNIQUE REFERENCES users(id),
    score           INTEGER NOT NULL DEFAULT 0 CHECK (score >= 0 AND score <= 100),

    -- Score factors
    phone_verified      BOOLEAN NOT NULL DEFAULT false,
    profile_complete    BOOLEAN NOT NULL DEFAULT false,
    has_avatar          BOOLEAN NOT NULL DEFAULT false,
    has_job_history     BOOLEAN NOT NULL DEFAULT false,
    no_no_shows         BOOLEAN NOT NULL DEFAULT true,
    no_excess_cancels   BOOLEAN NOT NULL DEFAULT true,

    -- Stats
    jobs_completed  INTEGER NOT NULL DEFAULT 0,
    jobs_cancelled  INTEGER NOT NULL DEFAULT 0,
    no_show_count   INTEGER NOT NULL DEFAULT 0,

    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS pro_scores (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL UNIQUE REFERENCES users(id),
    score           INTEGER NOT NULL DEFAULT 0 CHECK (score >= 0 AND score <= 1000),

    -- Positive factors
    jobs_completed      INTEGER NOT NULL DEFAULT 0,
    five_star_reviews   INTEGER NOT NULL DEFAULT 0,
    avg_response_mins   INTEGER NOT NULL DEFAULT 0,
    profile_complete    BOOLEAN NOT NULL DEFAULT false,
    active_last_7_days  BOOLEAN NOT NULL DEFAULT false,
    portfolio_count     INTEGER NOT NULL DEFAULT 0,

    -- Penalty factors
    no_show_count           INTEGER NOT NULL DEFAULT 0,
    cancelled_after_accept  INTEGER NOT NULL DEFAULT 0,
    slow_response_count     INTEGER NOT NULL DEFAULT 0,
    low_rating_count        INTEGER NOT NULL DEFAULT 0,

    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_commit_scores_user ON commit_scores(user_id);
CREATE INDEX idx_commit_scores_score ON commit_scores(score DESC);
CREATE INDEX idx_pro_scores_user ON pro_scores(user_id);
CREATE INDEX idx_pro_scores_score ON pro_scores(score DESC);

COMMENT ON TABLE commit_scores IS 'Client reliability scores (0-100) - affects lead visibility and pricing';
COMMENT ON TABLE pro_scores IS 'Handyman reputation scores (0-1000) - affects ranking, pricing, and badges';
COMMENT ON COLUMN commit_scores.score IS '0-49: unverified, 50-79: standard, 80-100: verified';
COMMENT ON COLUMN pro_scores.score IS '0-299: low, 300-799: standard, 800+: Pro Partner';
