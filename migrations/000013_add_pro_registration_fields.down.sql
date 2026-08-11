-- Migration Down: Add Pro Registration Fields

ALTER TABLE handyman_profiles
    DROP COLUMN IF EXISTS company_address,
    DROP COLUMN IF EXISTS pkd_code,
    DROP COLUMN IF EXISTS gus_status,
    DROP COLUMN IF EXISTS gus_verified,
    DROP COLUMN IF EXISTS business_type,
    DROP COLUMN IF EXISTS is_vat_payer,
    DROP COLUMN IF EXISTS consent_identity_verification,
    DROP COLUMN IF EXISTS consent_marketing,
    DROP COLUMN IF EXISTS experience_years,
    DROP COLUMN IF EXISTS working_hours,
    DROP COLUMN IF EXISTS verification_status,
    DROP COLUMN IF EXISTS ceidg_document_url,
    DROP COLUMN IF EXISTS ceidg_uploaded_at,
    DROP COLUMN IF EXISTS microtransfer_status,
    DROP COLUMN IF EXISTS microtransfer_code,
    DROP COLUMN IF EXISTS microtransfer_confirmed_at;

ALTER TABLE handyman_pricing
    DROP COLUMN IF EXISTS estimated_duration;
