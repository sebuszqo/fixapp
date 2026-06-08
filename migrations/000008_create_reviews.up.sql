CREATE TABLE IF NOT EXISTS reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID NOT NULL REFERENCES jobs(id),
    reviewer_id UUID NOT NULL REFERENCES users(id),
    reviewee_id UUID NOT NULL REFERENCES users(id),
    type VARCHAR(30) NOT NULL CHECK (type IN ('client_to_handyman', 'handyman_to_client')),
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- One review per reviewer per job
    CONSTRAINT unique_review_per_job UNIQUE (job_id, reviewer_id)
);

-- Index for looking up reviews received by a user (for score calculation)
CREATE INDEX idx_reviews_reviewee_id ON reviews(reviewee_id);

-- Index for looking up reviews by job
CREATE INDEX idx_reviews_job_id ON reviews(job_id);

-- Index for looking up reviews left by a user
CREATE INDEX idx_reviews_reviewer_id ON reviews(reviewer_id);
