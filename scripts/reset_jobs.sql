SET client_encoding = 'UTF8';
BEGIN;

DELETE FROM reviews;
DELETE FROM messages;
DELETE FROM leads;
DELETE FROM jobs;

-- Job 1: In progress job (accepted lead for Jan Hydraulik)
INSERT INTO jobs (
  id, client_id, category_id, district_id,
  title, description, urgency, status,
  address, building_type, floor, has_elevator,
  preferred_date1, preferred_time, budget, contact_method,
  created_at, updated_at, expires_at
) VALUES (
  '00000000-0000-0000-0000-000000000401',
  '00000000-0000-0000-0000-000000000010',
  (SELECT id FROM service_categories WHERE slug = 'hydraulik'),
  (SELECT id FROM districts WHERE slug = 'krowodrza'),
  'Wymiana cieknącej baterii umywalkowej i zaworów kątowych',
  'Dzień dobry, bateria umywalkowa w łazience przecieka, a zawory pod umywalką są zapieczone i uszkodzone. Zakupiłam nową baterię marki Grohe oraz zestawy montażowe. Potrzebuję fachowca do demontażu starych elementów, wymiany zaworów kątowych oraz montażu i podłączenia nowej baterii wraz ze sprawdzeniem szczelności.',
  'normal',
  'in_progress',
  'ul. Królewska 15/4, Kraków',
  'apartment',
  2,
  true,
  NOW() + INTERVAL '3 days',
  'afternoon',
  250,
  'phone',
  NOW() - INTERVAL '2 days',
  NOW(),
  NOW() + INTERVAL '5 days'
);

INSERT INTO leads (
  id, job_id, handyman_id, status, price, client_commit_score,
  estimated_price, arrival_time, proposal_message,
  created_at, updated_at, expires_at, accepted_at
) VALUES (
  '00000000-0000-0000-0000-000000000501',
  '00000000-0000-0000-0000-000000000401',
  '00000000-0000-0000-0000-000000000100',
  'accepted',
  28,
  85,
  220,
  TO_CHAR(NOW() + INTERVAL '3 days', 'YYYY-MM-DD 14:00'),
  'Chętnie podejmę się tego zadania. Mam ze sobą komplet profesjonalnych narzędzi oraz zapasowe uszczelnienia. Przyjadę w ustalonym terminie.',
  NOW() - INTERVAL '2 days',
  NOW() - INTERVAL '1 day',
  NOW() + INTERVAL '5 days',
  NOW() - INTERVAL '1 day'
);

INSERT INTO messages (
  id, sender_id, receiver_id, job_id, content, is_read, created_at
) VALUES 
  (gen_random_uuid(), '00000000-0000-0000-0000-000000000010', '00000000-0000-0000-0000-000000000100', '00000000-0000-0000-0000-000000000401', 'Dzień dobry Panie Janie, zaakceptowałam Pana ofertę wymiany baterii.', true, NOW() - INTERVAL '1 day'),
  (gen_random_uuid(), '00000000-0000-0000-0000-000000000100', '00000000-0000-0000-0000-000000000010', '00000000-0000-0000-0000-000000000401', 'Dzień dobry Pani Anno! Bardzo dziękuję. Będę pod podanym adresem zgodnie z terminem. Do zobaczenia!', true, NOW() - INTERVAL '20 hours');

-- Job 2: Urgent active job (New Lead for Jan Hydraulik)
INSERT INTO jobs (
  id, client_id, category_id, district_id,
  title, description, urgency, status,
  address, building_type, floor, has_elevator,
  preferred_date1, preferred_time, budget, contact_method,
  created_at, updated_at, expires_at
) VALUES (
  '00000000-0000-0000-0000-000000000402',
  '00000000-0000-0000-0000-000000000010',
  (SELECT id FROM service_categories WHERE slug = 'hydraulik'),
  (SELECT id FROM districts WHERE slug = 'krowodrza'),
  'Udrożnienie zapchanego odpływu w zlewie kuchennym',
  'Awarie w kuchni - woda w zlewie spływa bardzo wolno, a podczas spuszczania słychać głośne bulgotanie w rurach. Domowe środki udrożniające nie pomogły. Szukam hydraulika ze sprężyną mechaniczną lub ciśnieniową.',
  'urgent',
  'active',
  'ul. Królewska 15/4, Kraków',
  'apartment',
  2,
  true,
  NOW() + INTERVAL '1 day',
  'morning',
  300,
  'phone',
  NOW() - INTERVAL '2 hours',
  NOW(),
  NOW() + INTERVAL '7 days'
);

INSERT INTO leads (
  id, job_id, handyman_id, status, price, client_commit_score,
  created_at, updated_at, expires_at
) VALUES (
  '00000000-0000-0000-0000-000000000502',
  '00000000-0000-0000-0000-000000000402',
  '00000000-0000-0000-0000-000000000100',
  'pending',
  28,
  85,
  NOW() - INTERVAL '2 hours',
  NOW(),
  NOW() + INTERVAL '22 hours'
);

-- Job 3: Normal active job (New Lead for Jan Hydraulik)
INSERT INTO jobs (
  id, client_id, category_id, district_id,
  title, description, urgency, status,
  address, building_type, floor, has_elevator,
  preferred_date1, preferred_time, budget, contact_method,
  created_at, updated_at, expires_at
) VALUES (
  '00000000-0000-0000-0000-000000000403',
  '00000000-0000-0000-0000-000000000010',
  (SELECT id FROM service_categories WHERE slug = 'hydraulik'),
  (SELECT id FROM districts WHERE slug = 'stare-miasto'),
  'Podłączenie pralki automatycznej i montaż zaworu odpływowego',
  'Zakupiliśmy nową pralkę do mieszkania na wynajem. Wymagane jest podłączenie dopływu wody z montażem dodatkowego zaworu oraz prawidłowe uszczelnienie odpływu w ścianie.',
  'normal',
  'active',
  'ul. Floriańska 22/6, Kraków',
  'apartment',
  1,
  false,
  NOW() + INTERVAL '5 days',
  'flexible',
  180,
  'app',
  NOW() - INTERVAL '5 hours',
  NOW(),
  NOW() + INTERVAL '7 days'
);

INSERT INTO leads (
  id, job_id, handyman_id, status, price, client_commit_score,
  created_at, updated_at, expires_at
) VALUES (
  '00000000-0000-0000-0000-000000000503',
  '00000000-0000-0000-0000-000000000403',
  '00000000-0000-0000-0000-000000000100',
  'pending',
  28,
  85,
  NOW() - INTERVAL '5 hours',
  NOW(),
  NOW() + INTERVAL '19 hours'
);

-- Job 4: New Lead for Jan Hydraulik
INSERT INTO jobs (
  id, client_id, category_id, district_id,
  title, description, urgency, status,
  address, building_type, floor, has_elevator,
  preferred_date1, preferred_time, budget, contact_method,
  created_at, updated_at, expires_at
) VALUES (
  '00000000-0000-0000-0000-000000000404',
  '00000000-0000-0000-0000-000000000010',
  (SELECT id FROM service_categories WHERE slug = 'hydraulik'),
  (SELECT id FROM districts WHERE slug = 'podgorze'),
  'Montaż i podłączenie kabiny prysznicowej z brodzikiem',
  'Poszukuję sprawnego hydraulika do złożenia i montażu nowej kabiny prysznicowej 90x90 z głębokim brodzikiem. Wymagane podłączenie odpływu, silikonowanie oraz podłączenie deszczownicy.',
  'normal',
  'active',
  'ul. Wielicka 42/12, Kraków',
  'apartment',
  3,
  true,
  NOW() + INTERVAL '4 days',
  'morning',
  450,
  'phone',
  NOW() - INTERVAL '1 hour',
  NOW(),
  NOW() + INTERVAL '7 days'
);

INSERT INTO leads (
  id, job_id, handyman_id, status, price, client_commit_score,
  created_at, updated_at, expires_at
) VALUES (
  '00000000-0000-0000-0000-000000000504',
  '00000000-0000-0000-0000-000000000404',
  '00000000-0000-0000-0000-000000000100',
  'pending',
  28,
  90,
  NOW() - INTERVAL '1 hour',
  NOW(),
  NOW() + INTERVAL '23 hours'
);

-- Job 5: New Lead for Jan Hydraulik
INSERT INTO jobs (
  id, client_id, category_id, district_id,
  title, description, urgency, status,
  address, building_type, floor, has_elevator,
  preferred_date1, preferred_time, budget, contact_method,
  created_at, updated_at, expires_at
) VALUES (
  '00000000-0000-0000-0000-000000000405',
  '00000000-0000-0000-0000-000000000010',
  (SELECT id FROM service_categories WHERE slug = 'hydraulik'),
  (SELECT id FROM districts WHERE slug = 'krowodrza'),
  'Wymiana syfonu umywalkowego oraz uszczelnienie zaworów w łazience',
  'Syfon pod umywalką przecieka przy połączeniu z rurą odpłyową. Potrzebna wymiana na nowy syfon mosiężny oraz sprawdzenie uszczelnień przy zaworach.',
  'urgent',
  'active',
  'ul. Lea 8/2, Kraków',
  'apartment',
  1,
  false,
  NOW() + INTERVAL '2 days',
  'afternoon',
  200,
  'app',
  NOW() - INTERVAL '30 minutes',
  NOW(),
  NOW() + INTERVAL '7 days'
);

INSERT INTO leads (
  id, job_id, handyman_id, status, price, client_commit_score,
  created_at, updated_at, expires_at
) VALUES (
  '00000000-0000-0000-0000-000000000505',
  '00000000-0000-0000-0000-000000000405',
  '00000000-0000-0000-0000-000000000100',
  'pending',
  28,
  88,
  NOW() - INTERVAL '30 minutes',
  NOW(),
  NOW() + INTERVAL '23 hours'
);

COMMIT;
