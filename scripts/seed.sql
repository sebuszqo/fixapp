-- FixApp Development Seed Data
-- Run: docker compose exec db psql -U fixapp -d fixapp -f /seed.sql
-- Or:  psql -U fixapp -d fixapp -f scripts/seed.sql

BEGIN;

-- ============================================================
-- 1. USERS (admin, 2 clients, 3 handymen)
-- ============================================================

-- Admin user
INSERT INTO users (id, email, name, role, provider, provider_id, phone, is_active, email_verified)
VALUES (
  '00000000-0000-0000-0000-000000000001',
  'admin@fixapp.pl',
  'Admin FixApp',
  'admin',
  'google',
  'admin-seed',
  '+48500000001',
  true, true
) ON CONFLICT (email) DO NOTHING;

-- Client 1
INSERT INTO users (id, email, name, role, provider, provider_id, phone, is_active, email_verified)
VALUES (
  '00000000-0000-0000-0000-000000000010',
  'klient1@example.com',
  'Anna Kowalska',
  'user',
  'google',
  'client1-seed',
  '+48501000001',
  true, true
) ON CONFLICT (email) DO NOTHING;

-- Client 2
INSERT INTO users (id, email, name, role, provider, provider_id, phone, is_active, email_verified)
VALUES (
  '00000000-0000-0000-0000-000000000011',
  'klient2@example.com',
  'Piotr Nowak',
  'user',
  'google',
  'client2-seed',
  '+48501000002',
  true, true
) ON CONFLICT (email) DO NOTHING;

-- Handyman 1 (Hydraulik)
INSERT INTO users (id, email, name, role, provider, provider_id, phone, is_active, email_verified)
VALUES (
  '00000000-0000-0000-0000-000000000100',
  'hydraulik.jan@example.com',
  'Jan Majster',
  'handyman',
  'google',
  'handyman1-seed',
  '+48502000001',
  true, true
) ON CONFLICT (email) DO NOTHING;

-- Handyman 2 (Elektryk)
INSERT INTO users (id, email, name, role, provider, provider_id, phone, is_active, email_verified)
VALUES (
  '00000000-0000-0000-0000-000000000101',
  'elektryk.marek@example.com',
  'Marek Naprawiacz',
  'handyman',
  'google',
  'handyman2-seed',
  '+48502000002',
  true, true
) ON CONFLICT (email) DO NOTHING;

-- Handyman 3 (Zlota raczka)
INSERT INTO users (id, email, name, role, provider, provider_id, phone, is_active, email_verified)
VALUES (
  '00000000-0000-0000-0000-000000000102',
  'raczka.tomek@example.com',
  'Tomek Wszystkomogacy',
  'handyman',
  'google',
  'handyman3-seed',
  '+48502000003',
  true, true
) ON CONFLICT (email) DO NOTHING;

-- ============================================================
-- 2. WALLETS (every user gets a wallet, handymen get free credits)
-- ============================================================

INSERT INTO wallets (id, user_id, balance) VALUES
  (gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 0),
  (gen_random_uuid(), '00000000-0000-0000-0000-000000000010', 0),
  (gen_random_uuid(), '00000000-0000-0000-0000-000000000011', 0),
  ('00000000-0000-0000-0000-000000000200', '00000000-0000-0000-0000-000000000100', 500),
  ('00000000-0000-0000-0000-000000000201', '00000000-0000-0000-0000-000000000101', 500),
  ('00000000-0000-0000-0000-000000000202', '00000000-0000-0000-0000-000000000102', 300)
ON CONFLICT DO NOTHING;

-- Credit transaction records for handyman onboarding
INSERT INTO wallet_transactions (id, wallet_id, type, amount, reason, description, balance_after, created_at) VALUES
  (gen_random_uuid(), '00000000-0000-0000-0000-000000000200', 'credit', 500, 'admin_top_up', 'Onboarding free credits', 500, NOW()),
  (gen_random_uuid(), '00000000-0000-0000-0000-000000000201', 'credit', 500, 'admin_top_up', 'Onboarding free credits', 500, NOW()),
  (gen_random_uuid(), '00000000-0000-0000-0000-000000000202', 'credit', 300, 'admin_top_up', 'Onboarding free credits', 300, NOW())
ON CONFLICT DO NOTHING;

-- ============================================================
-- 3. HANDYMAN PROFILES
-- ============================================================

-- Get category & district IDs for reference
-- Categories: hydraulik, elektryk, zlota-raczka, malarz, etc.
-- Districts: stare-miasto, krowodrza, podgorze, nowa-huta, etc.

INSERT INTO handyman_profiles (id, user_id, company_name, bio, categories, districts, is_available, emergency_available, phone)
VALUES (
  '00000000-0000-0000-0000-000000000300',
  '00000000-0000-0000-0000-000000000100',
  'Jan Hydraulik',
  'Doswiadczony hydraulik z 15-letnim stazem. Specjalizuje sie w naprawach instalacji wodnych, kanalizacji i ogrzewania. Szybko, tanio, solidnie.',
  (SELECT ARRAY_AGG(id) FROM service_categories WHERE slug IN ('hydraulik', 'klimatyzacja')),
  (SELECT ARRAY_AGG(id) FROM districts WHERE slug IN ('stare-miasto', 'krowodrza', 'podgorze', 'debniki')),
  true,
  true,
  '+48502000001'
) ON CONFLICT DO NOTHING;

INSERT INTO handyman_profiles (id, user_id, company_name, bio, categories, districts, is_available, emergency_available, phone)
VALUES (
  '00000000-0000-0000-0000-000000000301',
  '00000000-0000-0000-0000-000000000101',
  'Elektro-Marek',
  'Uprawniony elektryk z certyfikatem SEP. Wykonuje instalacje elektryczne, naprawy, montaz oswietlenia i gniazd. Gwarancja na uslugi.',
  (SELECT ARRAY_AGG(id) FROM service_categories WHERE slug IN ('elektryk', 'klimatyzacja')),
  (SELECT ARRAY_AGG(id) FROM districts WHERE slug IN ('krowodrza', 'bronowice', 'zwierzyniec', 'stare-miasto')),
  true,
  false,
  '+48502000002'
) ON CONFLICT DO NOTHING;

INSERT INTO handyman_profiles (id, user_id, company_name, bio, categories, districts, is_available, emergency_available, phone)
VALUES (
  '00000000-0000-0000-0000-000000000302',
  '00000000-0000-0000-0000-000000000102',
  'Tomek Zlota Raczka',
  'Zajmuje sie wszystkim po trochu - drobne naprawy, montaz mebli, malowanie, wieszanie obrazow. Szybka realizacja, uczciwe ceny.',
  (SELECT ARRAY_AGG(id) FROM service_categories WHERE slug IN ('zlota-raczka', 'malarz', 'agd')),
  (SELECT ARRAY_AGG(id) FROM districts WHERE slug IN ('podgorze', 'debniki', 'nowa-huta', 'mistrzejowice', 'bienczyce')),
  true,
  false,
  '+48502000003'
) ON CONFLICT DO NOTHING;

-- ============================================================
-- 4. PRICING ITEMS
-- ============================================================

INSERT INTO handyman_pricing (id, profile_id, service_name, price_from, price_to, unit, sort_order) VALUES
  (gen_random_uuid(), '00000000-0000-0000-0000-000000000300', 'Naprawa kranu', 80, 150, 'za usluge', 1),
  (gen_random_uuid(), '00000000-0000-0000-0000-000000000300', 'Udrożnienie rury', 150, 400, 'za usluge', 2),
  (gen_random_uuid(), '00000000-0000-0000-0000-000000000300', 'Montaż baterii', 100, 200, 'za usluge', 3),
  (gen_random_uuid(), '00000000-0000-0000-0000-000000000301', 'Wymiana gniazdka', 60, 100, 'za sztuke', 1),
  (gen_random_uuid(), '00000000-0000-0000-0000-000000000301', 'Montaż lampy', 80, 200, 'za sztuke', 2),
  (gen_random_uuid(), '00000000-0000-0000-0000-000000000301', 'Diagnoza instalacji', 150, 300, 'za wizyte', 3),
  (gen_random_uuid(), '00000000-0000-0000-0000-000000000302', 'Montaż mebli', 50, 150, 'za mebel', 1),
  (gen_random_uuid(), '00000000-0000-0000-0000-000000000302', 'Malowanie ściany', 25, 40, 'za m2', 2),
  (gen_random_uuid(), '00000000-0000-0000-0000-000000000302', 'Drobne naprawy', 60, 120, 'za godzine', 3)
ON CONFLICT DO NOTHING;

-- ============================================================
-- 5. SCORES
-- ============================================================

-- Commit scores for clients
INSERT INTO commit_scores (id, user_id, score, phone_verified, profile_complete, has_avatar, has_job_history, no_no_shows, no_excess_cancels, jobs_completed, jobs_cancelled, no_show_count, updated_at) VALUES
  (gen_random_uuid(), '00000000-0000-0000-0000-000000000010', 75, true, true, true, true, true, true, 3, 0, 0, NOW()),
  (gen_random_uuid(), '00000000-0000-0000-0000-000000000011', 45, true, true, false, false, true, true, 0, 0, 0, NOW())
ON CONFLICT DO NOTHING;

-- Pro scores for handymen
INSERT INTO pro_scores (id, user_id, score, jobs_completed, five_star_reviews, avg_response_mins, profile_complete, active_last_7_days, portfolio_count, no_show_count, cancelled_after_accept, slow_response_count, low_rating_count, updated_at) VALUES
  (gen_random_uuid(), '00000000-0000-0000-0000-000000000100', 720, 12, 8, 25, true, true, 5, 0, 0, 0, 0, NOW()),
  (gen_random_uuid(), '00000000-0000-0000-0000-000000000101', 550, 7, 4, 45, true, true, 3, 0, 1, 0, 0, NOW()),
  (gen_random_uuid(), '00000000-0000-0000-0000-000000000102', 350, 4, 2, 90, false, true, 0, 0, 0, 1, 1, NOW())
ON CONFLICT DO NOTHING;

-- ============================================================
-- 6. SAMPLE JOBS & LEADS FOR ANNA KOWALSKA (HYDRAULIK)
-- ============================================================

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
) ON CONFLICT DO NOTHING;

INSERT INTO leads (
  id, job_id, handyman_id, status, price, client_commit_score,
  estimated_price, arrival_time, proposal_message,
  created_at, updated_at, expires_at, accepted_at
) VALUES (
  '00000000-0000-0000-0000-000000000501',
  '00000000-0000-0000-0000-000000000401',
  '00000000-0000-0000-0000-000000000100',
  'accepted',
  15,
  85,
  220,
  TO_CHAR(NOW() + INTERVAL '3 days', 'YYYY-MM-DD 14:00'),
  'Chętnie podejmę się tego zadania. Mam ze sobą komplet profesjonalnych narzędzi oraz zapasowe uszczelnienia. Przyjadę w ustalonym terminie.',
  NOW() - INTERVAL '2 days',
  NOW() - INTERVAL '1 day',
  NOW() + INTERVAL '5 days',
  NOW() - INTERVAL '1 day'
) ON CONFLICT DO NOTHING;

INSERT INTO messages (
  id, sender_id, receiver_id, job_id, content, is_read, created_at
) VALUES 
  (gen_random_uuid(), '00000000-0000-0000-0000-000000000010', '00000000-0000-0000-0000-000000000100', '00000000-0000-0000-0000-000000000401', 'Dzień dobry Panie Janie, zaakceptowałam Pana ofertę wymiany baterii.', true, NOW() - INTERVAL '1 day'),
  (gen_random_uuid(), '00000000-0000-0000-0000-000000000100', '00000000-0000-0000-0000-000000000010', '00000000-0000-0000-0000-000000000401', 'Dzień dobry Pani Anno! Bardzo dziękuję. Będę pod podanym adresem zgodnie z terminem. Do zobaczenia!', true, NOW() - INTERVAL '20 hours')
ON CONFLICT DO NOTHING;

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
) ON CONFLICT DO NOTHING;

INSERT INTO leads (
  id, job_id, handyman_id, status, price, client_commit_score,
  created_at, updated_at, expires_at
) VALUES (
  '00000000-0000-0000-0000-000000000502',
  '00000000-0000-0000-0000-000000000402',
  '00000000-0000-0000-0000-000000000100',
  'pending',
  20,
  85,
  NOW() - INTERVAL '2 hours',
  NOW(),
  NOW() + INTERVAL '22 hours'
) ON CONFLICT DO NOTHING;

-- ============================================================
-- Done!
-- ============================================================

COMMIT;

-- Summary
SELECT 'Users' AS entity, COUNT(*) AS count FROM users
UNION ALL SELECT 'Wallets', COUNT(*) FROM wallets
UNION ALL SELECT 'Handyman Profiles', COUNT(*) FROM handyman_profiles
UNION ALL SELECT 'Pricing Items', COUNT(*) FROM handyman_pricing
UNION ALL SELECT 'Categories', COUNT(*) FROM service_categories
UNION ALL SELECT 'Districts', COUNT(*) FROM districts
UNION ALL SELECT 'Jobs', COUNT(*) FROM jobs
UNION ALL SELECT 'Commit Scores', COUNT(*) FROM commit_scores
UNION ALL SELECT 'Pro Scores', COUNT(*) FROM pro_scores;
