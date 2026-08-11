-- Migration: Add Pro Registration Fields to handyman_profiles and handyman_pricing

ALTER TABLE handyman_profiles
    ADD COLUMN IF NOT EXISTS company_address               VARCHAR(255),
    ADD COLUMN IF NOT EXISTS pkd_code                      VARCHAR(100),
    ADD COLUMN IF NOT EXISTS gus_status                    VARCHAR(50) DEFAULT 'Aktywna',
    ADD COLUMN IF NOT EXISTS gus_verified                  BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS business_type                 VARCHAR(50) DEFAULT 'jdg',
    ADD COLUMN IF NOT EXISTS is_vat_payer                  BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS consent_identity_verification BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS consent_marketing             BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS experience_years              VARCHAR(50),
    ADD COLUMN IF NOT EXISTS working_hours                 JSONB,
    ADD COLUMN IF NOT EXISTS verification_status           VARCHAR(50) NOT NULL DEFAULT 'unverified',
    ADD COLUMN IF NOT EXISTS ceidg_document_url            VARCHAR(500),
    ADD COLUMN IF NOT EXISTS ceidg_uploaded_at             TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS microtransfer_status          VARCHAR(50) DEFAULT 'none',
    ADD COLUMN IF NOT EXISTS microtransfer_code            VARCHAR(50),
    ADD COLUMN IF NOT EXISTS microtransfer_confirmed_at    TIMESTAMPTZ;

ALTER TABLE handyman_pricing
    ADD COLUMN IF NOT EXISTS estimated_duration            VARCHAR(50);
