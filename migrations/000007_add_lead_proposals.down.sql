ALTER TABLE leads
DROP COLUMN IF EXISTS estimated_price,
DROP COLUMN IF EXISTS arrival_time,
DROP COLUMN IF EXISTS proposal_message;
