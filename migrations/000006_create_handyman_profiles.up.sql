-- Handyman profiles: extended profile data for service providers

CREATE TABLE IF NOT EXISTS handyman_profiles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL UNIQUE REFERENCES users(id),

    -- Company info
    company_name    VARCHAR(200),
    nip             VARCHAR(20),
    phone           VARCHAR(50),
    email           VARCHAR(255),

    -- Public profile
    bio             TEXT,
    avatar_url      VARCHAR(500),

    -- Service configuration (stored as UUID arrays)
    categories      UUID[] DEFAULT '{}',
    districts       UUID[] DEFAULT '{}',

    -- Availability
    is_available        BOOLEAN NOT NULL DEFAULT true,
    emergency_available BOOLEAN NOT NULL DEFAULT false,

    -- Verification
    is_verified     BOOLEAN NOT NULL DEFAULT false,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Pricing list for handyman services
CREATE TABLE IF NOT EXISTS handyman_pricing (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id      UUID NOT NULL REFERENCES handyman_profiles(id) ON DELETE CASCADE,
    service_name    VARCHAR(200) NOT NULL,
    price_from      INTEGER NOT NULL CHECK (price_from >= 0),
    price_to        INTEGER CHECK (price_to IS NULL OR price_to >= price_from),
    unit            VARCHAR(50) NOT NULL DEFAULT 'per service',
    sort_order      INTEGER NOT NULL DEFAULT 0
);

-- Portfolio photos
CREATE TABLE IF NOT EXISTS handyman_portfolio (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id      UUID NOT NULL REFERENCES handyman_profiles(id) ON DELETE CASCADE,
    image_url       VARCHAR(500) NOT NULL,
    caption         VARCHAR(300),
    sort_order      INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_handyman_profiles_user ON handyman_profiles(user_id);
CREATE INDEX idx_handyman_profiles_available ON handyman_profiles(is_available) WHERE is_available = true;
CREATE INDEX idx_handyman_profiles_verified ON handyman_profiles(is_verified) WHERE is_verified = true;
CREATE INDEX idx_handyman_profiles_categories ON handyman_profiles USING GIN(categories);
CREATE INDEX idx_handyman_profiles_districts ON handyman_profiles USING GIN(districts);
CREATE INDEX idx_handyman_pricing_profile ON handyman_pricing(profile_id);
CREATE INDEX idx_handyman_portfolio_profile ON handyman_portfolio(profile_id);

-- Updated_at trigger
CREATE TRIGGER update_handyman_profiles_updated_at
    BEFORE UPDATE ON handyman_profiles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Constraint: max 3 categories
ALTER TABLE handyman_profiles ADD CONSTRAINT chk_max_categories
    CHECK (array_length(categories, 1) IS NULL OR array_length(categories, 1) <= 3);

COMMENT ON TABLE handyman_profiles IS 'Extended profiles for handyman users (portfolio, pricing, service areas)';
COMMENT ON COLUMN handyman_profiles.categories IS 'Array of service_category UUIDs this handyman serves (max 3)';
COMMENT ON COLUMN handyman_profiles.districts IS 'Array of district UUIDs this handyman covers';
COMMENT ON COLUMN handyman_profiles.nip IS 'Polish tax ID for B2B invoicing';
COMMENT ON COLUMN handyman_profiles.is_available IS 'false = vacation/pause mode, does not receive new leads';
