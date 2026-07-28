-- Ensure all 12 service categories exist with icons and high quality image URLs

INSERT INTO service_categories (name, slug, icon, base_price, image_url) VALUES
    ('Hydraulik', 'hydraulik', 'droplets', 28, 'https://images.unsplash.com/photo-1585704032915-c3400ca199e7?auto=format&fit=crop&w=800&q=80'),
    ('Elektryk', 'elektryk', 'zap', 28, 'https://images.unsplash.com/photo-1621905252507-b35492cc74b4?auto=format&fit=crop&w=800&q=80'),
    ('Złota rączka', 'zlota-raczka', 'hammer', 15, 'https://images.unsplash.com/photo-1581783898377-1c85bf937427?auto=format&fit=crop&w=800&q=80'),
    ('Naprawa AGD', 'agd', 'tv', 20, 'https://images.unsplash.com/photo-1584622650111-993a426fbf0a?auto=format&fit=crop&w=800&q=80'),
    ('Malarz', 'malarz', 'paintbrush', 22, 'https://images.unsplash.com/photo-1562259929-b4e1fd3aef09?auto=format&fit=crop&w=800&q=80'),
    ('Stolarz', 'stolarz', 'box', 22, 'https://images.unsplash.com/photo-1504148455328-c376907d081c?auto=format&fit=crop&w=800&q=80'),
    ('Ślusarz', 'slusarz', 'key', 25, 'https://images.unsplash.com/photo-1582139329536-e7284fece509?auto=format&fit=crop&w=800&q=80'),
    ('Ogrodnik', 'ogrodnik', 'flower', 18, 'https://images.unsplash.com/photo-1416879595882-3373a0480b5b?auto=format&fit=crop&w=800&q=80'),
    ('Klimatyzacja', 'klimatyzacja', 'wind', 30, 'https://images.unsplash.com/photo-1621905251189-08b45d6a269e?auto=format&fit=crop&w=800&q=80'),
    ('Sprzątanie', 'sprzatanie', 'sparkles', 15, 'https://images.unsplash.com/photo-1581578731548-c64695cc6952?auto=format&fit=crop&w=800&q=80'),
    ('Przeprowadzki', 'przeprowadzki', 'truck', 25, 'https://images.unsplash.com/photo-1600518464441-9154a4dea21b?auto=format&fit=crop&w=800&q=80'),
    ('Dekarz', 'dacharz', 'home', 30, 'https://images.unsplash.com/photo-1632759145351-1d592919f522?auto=format&fit=crop&w=800&q=80')
ON CONFLICT (slug) DO UPDATE SET
    name = EXCLUDED.name,
    icon = EXCLUDED.icon,
    image_url = EXCLUDED.image_url,
    base_price = EXCLUDED.base_price;
