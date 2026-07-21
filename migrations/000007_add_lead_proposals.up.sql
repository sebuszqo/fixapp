-- Add proposal fields to leads table
ALTER TABLE leads
ADD COLUMN IF NOT EXISTS estimated_price INTEGER,
ADD COLUMN IF NOT EXISTS arrival_time VARCHAR(255),
ADD COLUMN IF NOT EXISTS proposal_message TEXT;
