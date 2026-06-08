-- Service categories and districts for Krakow MVP

CREATE TABLE IF NOT EXISTS service_categories (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL,
    slug        VARCHAR(100) NOT NULL UNIQUE,
    icon        VARCHAR(50),
    base_price  INTEGER NOT NULL DEFAULT 20,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS districts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL,
    slug        VARCHAR(100) NOT NULL UNIQUE,
    city_name   VARCHAR(100) NOT NULL DEFAULT 'Krakow',
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed Krakow categories
INSERT INTO service_categories (name, slug, icon, base_price) VALUES
    ('Hydraulik', 'hydraulik', 'wrench', 28),
    ('Elektryk', 'elektryk', 'zap', 28),
    ('Zlota raczka', 'zlota-raczka', 'hammer', 15),
    ('AGD', 'agd', 'settings', 20),
    ('Malarz', 'malarz', 'paintbrush', 22),
    ('Slusarz', 'slusarz', 'key', 25),
    ('Klimatyzacja', 'klimatyzacja', 'thermometer', 30),
    ('Sprzatanie', 'sprzatanie', 'sparkles', 15);

-- Seed Krakow districts
INSERT INTO districts (name, slug, city_name) VALUES
    ('Stare Miasto', 'stare-miasto', 'Krakow'),
    ('Krowodrza', 'krowodrza', 'Krakow'),
    ('Bronowice', 'bronowice', 'Krakow'),
    ('Zwierzyniec', 'zwierzyniec', 'Krakow'),
    ('Debniki', 'debniki', 'Krakow'),
    ('Podgorze', 'podgorze', 'Krakow'),
    ('Nowa Huta', 'nowa-huta', 'Krakow'),
    ('Pradnik Bialy', 'pradnik-bialy', 'Krakow'),
    ('Pradnik Czerwony', 'pradnik-czerwony', 'Krakow'),
    ('Czyzyny', 'czyzyny', 'Krakow'),
    ('Mistrzejowice', 'mistrzejowice', 'Krakow'),
    ('Bienczyce', 'bienczyce', 'Krakow'),
    ('Ruczaj', 'ruczaj', 'Krakow'),
    ('Prokocim', 'prokocim', 'Krakow'),
    ('Lagiewniki', 'lagiewniki', 'Krakow'),
    ('Borek Falecki', 'borek-falecki', 'Krakow'),
    ('Swoszowice', 'swoszowice', 'Krakow'),
    ('Wzgorza Krzeslawickie', 'wzgorza-krzeslawickie', 'Krakow');

-- Indexes
CREATE INDEX idx_categories_slug ON service_categories(slug);
CREATE INDEX idx_categories_active ON service_categories(is_active) WHERE is_active = true;
CREATE INDEX idx_districts_slug ON districts(slug);
CREATE INDEX idx_districts_active ON districts(is_active) WHERE is_active = true;
CREATE INDEX idx_districts_city ON districts(city_name);

COMMENT ON TABLE service_categories IS 'Service types offered on the platform (e.g., Hydraulik, Elektryk)';
COMMENT ON TABLE districts IS 'Geographic areas for service matching';
COMMENT ON COLUMN service_categories.base_price IS 'Base lead fee in credits for this category';
