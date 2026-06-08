# FixApp Backend

Home repair services marketplace API (Kraków, Poland). Connects clients with handymen through a lead-based system with dynamic pricing.

## Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.25+ (for local development)
- `migrate` CLI (for running migrations manually)

### 1. Start the database

```bash
docker compose up db -d
```

### 2. Run migrations

```bash
# Install migrate if you don't have it
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations
migrate -path migrations -database "postgres://fixapp:fixapp@localhost:5432/fixapp?sslmode=disable" up
```

### 3. Seed test data

```bash
docker compose exec -T db psql -U fixapp -d fixapp < scripts/seed.sql
```

This creates:
- 1 admin, 2 clients, 3 handymen with profiles
- Wallets with free credits for handymen
- Pricing items, scores, and a sample active job
- 8 service categories + 18 Kraków districts

### 4. Run the server

```bash
# Option A: locally
go run cmd/api/main.go

# Option B: Docker (full stack)
docker compose up --build
```

### 5. Open Swagger UI

http://localhost:8080/swagger/index.html

---

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PORT` | No | `8080` | Server port |
| `LOG_LEVEL` | No | `info` | `debug`, `info`, `warn`, `error` |
| `DB_HOST` | No | `localhost` | PostgreSQL host |
| `DB_PORT` | No | `5432` | PostgreSQL port |
| `DB_USER` | No | `fixapp` | PostgreSQL user |
| `DB_PASSWORD` | No | `fixapp` | PostgreSQL password |
| `DB_NAME` | No | `fixapp` | PostgreSQL database name |
| `DB_SSLMODE` | No | `disable` | PostgreSQL SSL mode |
| `JWT_SECRET` | Yes* | `dev-secret-...` | JWT signing key (*insecure default in dev) |
| `GOOGLE_CLIENT_ID` | No | - | Google OAuth client ID |
| `GOOGLE_CLIENT_SECRET` | No | - | Google OAuth client secret |
| `GOOGLE_REDIRECT_URL` | No | - | Google OAuth redirect URL |
| `FACEBOOK_APP_ID` | No | - | Facebook OAuth app ID |
| `FACEBOOK_APP_SECRET` | No | - | Facebook OAuth app secret |
| `FACEBOOK_REDIRECT_URL` | No | - | Facebook OAuth redirect URL |

---

## Setting Up OAuth (Google)

### 1. Create a Google Cloud project

1. Go to https://console.cloud.google.com/
2. Click the project dropdown (top-left) → **New Project**
3. Name: `FixApp` → Create
4. Make sure the new project is selected

### 2. Enable the People API

1. Go to **APIs & Services → Library**
2. Search for **"Google People API"**
3. Click it → **Enable**

### 3. Configure consent screen

1. Go to **APIs & Services → OAuth consent screen**
2. Choose **External** → Create
3. Fill in:
   - App name: `FixApp`
   - User support email: your email
   - Developer contact: your email
4. Click **Save and Continue**
5. Add scopes: click **Add or Remove Scopes**, select:
   - `email`
   - `profile`
   - `openid`
6. Save and Continue
7. **Test users**: click **Add Users**, add your own email
   - This is required while the app is in "Testing" status
   - Only added test users can log in until you publish the app
8. Save and Continue → Back to Dashboard

### 4. Create OAuth credentials

1. Go to **APIs & Services → Credentials**
2. Click **Create Credentials → OAuth client ID**
3. Application type: **Web application**
4. Name: `FixApp Dev`
5. **Authorized redirect URIs**: click Add URI, enter:
   ```
   http://localhost:8080/auth/google/callback
   ```
6. Click **Create**
7. A popup shows your **Client ID** and **Client Secret** — copy both

### 5. Set environment variables

Create a `.env` file in the project root (already in `.gitignore`):

```env
GOOGLE_CLIENT_ID=123456789-xxxxxxx.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=GOCSPX-xxxxxxxxxx
GOOGLE_REDIRECT_URL=http://localhost:8080/auth/google/callback
```

Or export them before running:

```bash
export GOOGLE_CLIENT_ID="123456789-xxxxxxx.apps.googleusercontent.com"
export GOOGLE_CLIENT_SECRET="GOCSPX-xxxxxxxxxx"
export GOOGLE_REDIRECT_URL="http://localhost:8080/auth/google/callback"
```

### 6. Test the login flow

```bash
# Start the server (with env vars set)
go run cmd/api/main.go

# Check that Google provider is registered (should show "google" in list)
curl http://localhost:8080/auth/providers

# Get the Google login URL
curl http://localhost:8080/auth/google/login
# Returns a URL like: https://accounts.google.com/o/oauth2/v2/auth?client_id=...
```

**How the full flow works:**

1. Frontend calls `GET /auth/google/login` → gets a Google OAuth URL
2. Frontend redirects user to that URL
3. User logs in with Google, grants consent
4. Google redirects to `http://localhost:8080/auth/google/callback?code=XXXX`
5. Backend exchanges the `code` for user info (email, name, avatar)
6. Backend creates user (first login) or finds existing user
7. Backend returns a JWT access token + refresh token
8. Frontend stores the JWT and sends it as `Authorization: Bearer <token>` on all requests

**First login creates a user with role `user`.** To get admin access:

```bash
docker compose exec -T db psql -U fixapp -d fixapp \
  -c "UPDATE users SET role = 'admin' WHERE email = 'your-google-email@gmail.com';"
```

---

## Setting Up OAuth (Facebook)

1. Go to https://developers.facebook.com/
2. Click **My Apps → Create App**
3. Choose **Consumer** type → Next
4. App name: `FixApp`, contact email: yours → Create
5. On the app dashboard, click **Add Product** → find **Facebook Login** → **Set Up**
6. Go to **Settings → Basic**: copy **App ID** and **App Secret**
7. Go to **Facebook Login → Settings**:
   - Valid OAuth Redirect URIs: `http://localhost:8080/auth/facebook/callback`
   - Save Changes
8. Go to **App Roles → Roles** and add yourself as a tester (required while app is in Development mode)

```bash
export FACEBOOK_APP_ID="your-app-id"
export FACEBOOK_APP_SECRET="your-app-secret"
export FACEBOOK_REDIRECT_URL="http://localhost:8080/auth/facebook/callback"
```

---

## Test Accounts (from seed)

| Email | Role | Notes |
|-------|------|-------|
| `admin@fixapp.pl` | admin | Full access |
| `klient1@example.com` | user (client) | Commit Score 75, has active job |
| `klient2@example.com` | user (client) | Commit Score 45 |
| `hydraulik.jan@example.com` | handyman | ProScore 720, 500 credits |
| `elektryk.marek@example.com` | handyman | ProScore 550, 500 credits |
| `raczka.tomek@example.com` | handyman | ProScore 350, 300 credits |

> Note: These are seeded via OAuth provider `google` with fake provider IDs. To actually login as them in dev, you'll need to either:
> - Generate a JWT manually (see below)
> - Or login with your real Google account and promote it: `UPDATE users SET role = 'admin' WHERE email = 'your@email.com';`

### Generate a dev JWT (for testing without OAuth)

You can test API endpoints using curl by creating a token directly. Check `internal/auth/token/` for the token service, or use the Swagger UI "Authorize" button with a valid JWT.

---

## Project Structure

```
fixapp/
├── cmd/api/main.go           # Entrypoint, DI wiring, route registration
├── internal/
│   ├── domain/               # Entities, value objects, errors (zero deps)
│   ├── auth/                 # OAuth2, JWT, RBAC, permissions
│   ├── user/                 # User CRUD
│   ├── job/                  # Job lifecycle (client posts jobs)
│   ├── lead/                 # Lead lifecycle (handyman accepts)
│   ├── wallet/               # Credit balance & transactions
│   ├── handyman/             # Profiles, search, portfolio, pricing
│   ├── scoring/              # CommitScore + ProScore calculators
│   ├── catalog/              # Reference data (categories, districts)
│   ├── dispatch/             # Job→handyman matching engine
│   ├── review/               # Bidirectional reviews
│   └── health/               # Health/readiness checks
├── pkg/                      # Shared utilities
├── migrations/               # SQL migrations (000001-000008)
├── scripts/seed.sql          # Dev seed data
├── docs/                     # Swagger + architecture docs
├── Dockerfile
└── docker-compose.yml
```

---

## Common Commands

```bash
# Run server locally
go run cmd/api/main.go

# Build
go build ./...

# Run vet
go vet ./...

# Regenerate Swagger docs
swag init -g cmd/api/main.go -o docs

# Run migrations
migrate -path migrations -database "postgres://fixapp:fixapp@localhost:5432/fixapp?sslmode=disable" up

# Rollback last migration
migrate -path migrations -database "postgres://fixapp:fixapp@localhost:5432/fixapp?sslmode=disable" down 1

# Reset database
migrate -path migrations -database "postgres://fixapp:fixapp@localhost:5432/fixapp?sslmode=disable" drop -f

# Seed data
docker compose exec -T db psql -U fixapp -d fixapp < scripts/seed.sql

# Connect to database
docker compose exec db psql -U fixapp -d fixapp

# Promote your account to admin
docker compose exec -T db psql -U fixapp -d fixapp -c "UPDATE users SET role = 'admin' WHERE email = 'your@email.com';"
```

---

## Architecture

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for detailed documentation on:
- Core business flows (post job → dispatch → accept lead → complete → review)
- Scoring system (CommitScore + ProScore)
- Dynamic lead pricing formula
- State machines (Job, Lead)
- Full API endpoint reference
