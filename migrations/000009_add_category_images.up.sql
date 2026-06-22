ALTER TABLE service_categories ADD COLUMN image_url VARCHAR(500);

UPDATE service_categories SET image_url = 'https://images.unsplash.com/photo-1584622650111-993a426fbf0a?auto=format&fit=crop&w=800&q=80' WHERE slug = 'hydraulik';
UPDATE service_categories SET image_url = 'https://images.unsplash.com/photo-1621905252507-b35492cc74b4?auto=format&fit=crop&w=800&q=80' WHERE slug = 'elektryk';
UPDATE service_categories SET image_url = 'https://images.unsplash.com/photo-1581783898377-1c85bf937427?auto=format&fit=crop&w=800&q=80' WHERE slug = 'zlota-raczka';
UPDATE service_categories SET image_url = 'https://images.unsplash.com/photo-1585771724684-38269d6639fd?auto=format&fit=crop&w=800&q=80' WHERE slug = 'agd';
UPDATE service_categories SET image_url = 'https://images.unsplash.com/photo-1562259929-b4e1fd3aef09?auto=format&fit=crop&w=800&q=80' WHERE slug = 'malarz';
UPDATE service_categories SET image_url = 'https://images.unsplash.com/photo-1504148455328-c376907d081c?auto=format&fit=crop&w=800&q=80' WHERE slug = 'stolarz';
UPDATE service_categories SET image_url = 'https://images.unsplash.com/photo-1558618666-fcd25c85f82e?auto=format&fit=crop&w=800&q=80' WHERE slug = 'slusarz';
UPDATE service_categories SET image_url = 'https://images.unsplash.com/photo-1416879595882-3373a0480b5b?auto=format&fit=crop&w=800&q=80' WHERE slug = 'ogrodnik';
UPDATE service_categories SET image_url = 'https://images.unsplash.com/photo-1631545806609-16d423d73587?auto=format&fit=crop&w=800&q=80' WHERE slug = 'klimatyzacja';
UPDATE service_categories SET image_url = 'https://images.unsplash.com/photo-1581578731548-c64695cc6952?auto=format&fit=crop&w=800&q=80' WHERE slug = 'sprzatanie';
UPDATE service_categories SET image_url = 'https://images.unsplash.com/photo-1600518464441-9154a4dea21b?auto=format&fit=crop&w=800&q=80' WHERE slug = 'przeprowadzki';
UPDATE service_categories SET image_url = 'https://images.unsplash.com/photo-1632759145351-1d592919f522?auto=format&fit=crop&w=800&q=80' WHERE slug = 'dacharz';
