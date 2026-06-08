# FixApp Backend Architecture

## Overview

FixApp is a home repair services marketplace connecting clients with handymen (Kraków, Poland MVP). The backend is a Go monolith using clean architecture with PostgreSQL.

## Tech Stack

- **Language**: Go 1.22+
- **Database**: PostgreSQL 16
- **Auth**: OAuth2 (Google, Facebook) + JWT
- **Docs**: Swagger/OpenAPI (auto-generated)
- **Logging**: Structured (zap)
- **Migrations**: golang-migrate

## Project Structure

```
fixapp/
├── cmd/api/main.go          # Entrypoint, DI wiring
├── internal/
│   ├── domain/              # Entities, value objects, errors (no dependencies)
│   ├── auth/                # OAuth2, JWT, RBAC
│   │   ├── provider/        # Google, Facebook OAuth adapters
│   │   ├── token/           # JWT service
│   │   └── permission/      # Role-based permissions
│   ├── user/                # User CRUD
│   ├── job/                 # Job lifecycle (client creates jobs)
│   ├── lead/                # Lead lifecycle (handyman accepts leads)
│   ├── wallet/              # Credit balance & transactions
│   ├── handyman/            # Handyman profiles
│   ├── scoring/             # CommitScore + ProScore calculators
│   ├── catalog/             # Reference data (categories, districts)
│   ├── dispatch/            # Job→handyman matching engine
│   ├── review/              # Bidirectional reviews
│   └── health/              # Health/readiness checks
├── pkg/
│   ├── database/            # DB connection
│   ├── logger/              # Global logger init
│   ├── middleware/          # JWT, logging, RBAC middleware
│   ├── ctxlog/              # Context-aware logging
│   └── response/            # Standardized HTTP responses
├── migrations/              # SQL migrations (000001-000008)
└── docs/                    # Swagger generated files
```

## Package Dependency Rules

```
domain ← (no imports from internal/)
service ← domain
repository ← domain
handler ← service, domain, pkg/*
cmd/api ← all packages (DI wiring)
```

No circular imports. When two packages need each other (e.g., `job` and `dispatch`), we use interface injection.

---

## Core Flows

### 1. Client Posts a Job

```
Client → POST /jobs (draft) → POST /jobs/{id}/publish (active)
                                       ↓
                              Dispatch service triggered
                                       ↓
                         Find matching handymen (category + district)
                                       ↓
                         Calculate dynamic lead price per handyman
                                       ↓
                         Create leads (one per matching handyman)
```

### 2. Handyman Accepts a Lead

```
Handyman → GET /leads (sees pending leads with price)
         → POST /leads/{id}/accept
                    ↓
           Check wallet balance >= lead price
                    ↓
           Atomic debit (FOR UPDATE lock)
                    ↓
           Lead status → accepted
           Job status → accepted
           Client contact info revealed to handyman
```

### 3. Job Completion & Review

```
Handyman → POST /jobs/{id}/complete (declares final value)
Client   → POST /jobs/{id}/confirm
         → POST /reviews (rates handyman 1-5)
Handyman → POST /reviews (rates client 1-5)
                    ↓
           ProScore recalculated for handyman
```

---

## Domain Entities

### Job (state machine)
```
draft → active → accepted → in_progress → done → (client confirms)
  ↓        ↓        ↓            ↓
cancelled cancelled cancelled  cancelled
```

### Lead (state machine)
```
pending → accepted
        → rejected
        → expired (TTL 24h)
```

### Wallet
- Each user has a wallet with credit balance
- Debit is atomic (PostgreSQL `SELECT ... FOR UPDATE`)
- Every transaction is an immutable audit log entry
- Transaction reasons: lead_accepted, admin_topup, refund, etc.

---

## Scoring System

### Commit Score (Client, 0-100)
Measures client reliability. Affects **lead pricing** (reliable clients = cheaper leads for handymen).

| Factor | Points |
|--------|--------|
| Verified phone | +20 |
| Complete profile | +15 |
| Has avatar | +10 |
| Job history > 0 | +10 |
| No no-shows | +10 |
| No excess cancels | +10 |
| No-show penalty | -20 |
| 2+ cancellations | -15 |

**Per-job bonuses** (computed at dispatch, not stored):
- Description > 250 chars: +20
- At least 1 photo: +15
- Specific time window: +10

**Client Multiplier** (applied to lead price):
- Verified (80-100): ×0.8 (cheaper leads)
- Standard (50-79): ×1.0
- Unverified (0-49): ×1.2 (more expensive leads)

### ProScore (Handyman, 0-1000)
Measures handyman quality. Affects **lead pricing** and **ranking**.

| Factor | Points |
|--------|--------|
| Jobs completed (×50, max 500) | +500 max |
| 5-star reviews (×30, max 300) | +300 max |
| Response time < 1h | +20 |
| Profile 100% complete | +15 |
| Active last 7 days | +10 |
| Portfolio photos (×5, max 50) | +50 max |
| No-show penalty (×100) | -100 each |
| Cancelled after accept (×50) | -50 each |
| Slow response > 24h (×30) | -30 each |
| Low rating 1-2 star (×20) | -20 each |

**Handyman Multiplier** (applied to lead price):
- Pro Partner (800+): ×0.8 (cheaper leads — reward)
- Standard (300-799): ×1.0
- Low (0-299): ×1.1 (slightly more expensive)

### Dynamic Lead Price Formula
```
Lead Price = Category Base Price × Client Multiplier × Handyman Multiplier
```

Each handyman may pay a different price for the same job.

---

## Authentication & Authorization

### Auth Flow
1. Client/handyman authenticates via Google/Facebook OAuth
2. Backend exchanges token, creates/finds user, returns JWT
3. JWT is sent as `Authorization: Bearer <token>` on subsequent requests
4. Middleware extracts user from JWT and adds to request context

### Roles
- **user** — default (client)
- **handyman** — can see leads, has profile
- **admin** — full access

### Permissions (RBAC)
Routes are protected with middleware:
- `RequireAuth` — any authenticated user
- `RequireHandyman` — handyman or admin role
- `HasPermission(ctx, permission.X)` — fine-grained checks

---

## API Endpoints Summary

### Public
| Method | Path | Description |
|--------|------|-------------|
| GET | /health | Health check |
| GET | /ready | Readiness check |
| GET | /categories | List service categories |
| GET | /districts | List districts |
| GET | /users/{id}/reviews | List user's reviews |
| GET | /users/{id}/rating | Get user's average rating |
| GET | /swagger/ | Swagger UI |

### Authenticated (any user)
| Method | Path | Description |
|--------|------|-------------|
| GET | /auth/providers | List OAuth providers |
| POST | /auth/{provider}/login | OAuth login |
| POST | /auth/refresh | Refresh JWT |
| GET | /user/profile | Get my profile |
| PUT | /user/profile | Update my profile |
| POST | /jobs | Create job (draft) |
| POST | /jobs/{id}/publish | Publish job |
| GET | /jobs/{id} | Get job details |
| GET | /jobs/my | List my jobs |
| POST | /jobs/{id}/confirm | Client confirms completion |
| POST | /jobs/{id}/cancel | Cancel job |
| POST | /reviews | Submit a review |
| GET | /jobs/{id}/reviews | List reviews for a job |
| GET | /score/commit | Get my Commit Score |
| GET | /wallet | Get my wallet balance |
| GET | /wallet/transactions | List my transactions |

### Handyman
| Method | Path | Description |
|--------|------|-------------|
| GET | /leads | List my leads |
| GET | /leads/{id} | Get lead details |
| POST | /leads/{id}/accept | Accept lead (debits credits) |
| POST | /leads/{id}/reject | Reject lead |
| POST | /jobs/{id}/complete | Mark job as done |
| GET | /handyman/profile | Get my handyman profile |
| PUT | /handyman/profile | Update profile |
| POST | /handyman/pricing | Add pricing item |
| POST | /handyman/portfolio | Add portfolio item |
| GET | /score/pro | Get my ProScore |

### Admin
| Method | Path | Description |
|--------|------|-------------|
| GET | /admin/users | List all users |
| PUT | /admin/users/{id}/role | Change user role |
| POST | /admin/wallet/{id}/topup | Add credits to user wallet |

---

## Database Migrations

| # | Name | Tables |
|---|------|--------|
| 000001 | create_users_table | users |
| 000002 | create_categories_and_districts | service_categories, districts (seeded) |
| 000003 | create_jobs | jobs |
| 000004 | create_leads | leads |
| 000005 | create_wallets | wallets, wallet_transactions |
| 000006 | create_handyman_profiles | handyman_profiles, pricing_items, portfolio_items |
| 000007 | create_scores | commit_scores, pro_scores |
| 000008 | create_reviews | reviews |

---

## Running

```bash
# Prerequisites: PostgreSQL running, env vars set
export DATABASE_URL=postgres://user:pass@localhost:5432/fixapp?sslmode=disable
export JWT_SECRET=your-secret
export GOOGLE_CLIENT_ID=...
export GOOGLE_CLIENT_SECRET=...

# Run migrations
migrate -path migrations -database $DATABASE_URL up

# Start server
go run cmd/api/main.go

# Swagger UI
open http://localhost:8080/swagger/
```

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| PORT | No | 8080 | Server port |
| LOG_LEVEL | No | info | Log level (debug/info/warn/error) |
| DATABASE_URL | Yes | - | PostgreSQL connection string |
| JWT_SECRET | Yes* | dev-secret | JWT signing key (*insecure default in dev) |
| GOOGLE_CLIENT_ID | No | - | Google OAuth client ID |
| GOOGLE_CLIENT_SECRET | No | - | Google OAuth client secret |
| GOOGLE_REDIRECT_URL | No | - | Google OAuth redirect URL |
| FACEBOOK_APP_ID | No | - | Facebook OAuth app ID |
| FACEBOOK_APP_SECRET | No | - | Facebook OAuth app secret |
| FACEBOOK_REDIRECT_URL | No | - | Facebook OAuth redirect URL |
