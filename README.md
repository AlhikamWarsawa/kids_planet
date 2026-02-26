<h1 align="center">Kids Planet</h1>

<p align="center">
<a href="LICENSE">
  <img src="https://img.shields.io/badge/license-MIT-blue.svg" />
</a>
  <img src="https://img.shields.io/badge/language-Go-blue.svg" />
</p>

<p align="center">
  <img src="img/icon.png" alt="Logo" width="180">
</p>

## 1. Project Overview

Kids Planet is a web-based educational game platform for children.  
It delivers HTML5 game packages in-browser (desktop and mobile) and provides:

- Player experience: catalog, play, leaderboard, history
- Admin experience: dashboard, game/catalog management, ZIP upload
- API + storage pipeline for sessions, analytics, and leaderboard data

This repository is a production-ready monorepo with:

- `apps/web` (SvelteKit frontend)
- `services/api` (Go Fiber API)
- `db` (baseline schema + seed)
- `infra` (Nginx + production compose)

## 2. Architecture (high-level)

### Components

- Web app (SvelteKit): player UI and admin UI
- API (Go Fiber): public + player + admin endpoints under `/api`
- Postgres: source of truth for relational data
- Valkey: leaderboard sorted sets and request/rate-limit counters
- MinIO: object store for game ZIPs and extracted assets
- Nginx (prod): reverse proxy for `/api/*`, static web, and `/games/*` asset delivery

### Core Domains

- Games: metadata, status lifecycle (`draft`, `active`, `archived`), catalog sorting (`newest`, `popular`)
- Sessions: start gameplay session and issue short-lived `play_token`
- Analytics: ingest gameplay events (for example `game_start`)
- Leaderboards: submit/read scores with daily/weekly and game/global scopes
- Admin: auth, profile, dashboard, game/category management

### Data Stores

- Postgres tables include `games`, `sessions`, `analytics_events`, `leaderboard_submissions`, `users` (plus related categories and player tables)
- Valkey keys include leaderboard (`lb:*`) and rate limit (`rl:*`)
- MinIO game objects are served under `/games/{id}/current/...`

### Trust Boundaries

- Internet client -> Nginx (`/`, `/api`, `/games`)
- Nginx -> API service and MinIO
- API -> Postgres, Valkey, MinIO
- Public endpoints vs admin endpoints (`/api/admin/*` requires admin JWT + role check)

## 3. Core Features

- Public game catalog with age/category filters and sorting (`newest`, `popular`)
- Game detail + playable URL resolution
- Session start with short-lived `play_token`
- Analytics ingestion endpoint (`/api/analytics/event`)
- Leaderboard submit/read/self endpoints with daily/weekly and game/global scopes
- Optional player auth (email + PIN) and player history
- Admin auth + dashboard overview
- Admin game CRUD, publish/unpublish, ZIP upload pipeline
- Admin age/education category management

## 4. Tech Stack

| Layer | Technology |
| --- | --- |
| Frontend | SvelteKit, TypeScript, Vite |
| Backend | Go 1.25, Fiber |
| Database | PostgreSQL 16 |
| Cache / Realtime Index | Valkey (Redis-compatible) |
| Object Storage | MinIO |
| Proxy / Static Serving | Nginx |
| API Contract | OpenAPI 3.0.3 (`services/api/openapi/openapi.yaml`) |
| Local/Infra Orchestration | Docker Compose, Make |

## 5. Repository Structure

```text
kids-planet
├── apps/
│   └── web/                    # SvelteKit frontend (player + admin UI)
├── services/
│   └── api/                    # Go Fiber API
├── db/
│   ├── migrations/             # Baseline schema
│   └── seeds/                  # Bootstrap seed data
├── infra/
│   ├── docker-compose.yml      # Production-style compose (includes Nginx)
│   └── nginx/nginx.conf        # Reverse proxy + security headers
├── docs/                       # Architecture, runbook, testing guides
├── tests/                      # k6 and other test assets
├── tools/                      # Postman collection, utilities
├── docker-compose.yml          # Dev compose
└── README.md
```

## 6. Getting Started (Bootstrap)

### Prerequisites

- Docker + Docker Compose
- Make

### 1) Configure environment

```bash
cp .env.example .env
```

Required environment groups:

- App: `ENV`, `PORT`, `APP_ORIGIN` (recommended for production/browser CORS)
- Postgres: `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_SSLMODE`
- Valkey: `VALKEY_ADDR`, `VALKEY_PASSWORD`, `VALKEY_DB`
- MinIO: `MINIO_ENDPOINT`, `MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY`, `MINIO_BUCKET`, `ZIP_UPLOAD_MAX_BYTES`
- JWT: `JWT_SECRET`, `JWT_ISSUER`, `JWT_EXPIRES_IN`

### 2) Bootstrap database (baseline + seed)

```bash
docker compose up -d postgres
docker exec -i planet_postgres psql -U admin -d kids_planet < db/migrations/000_baseline.sql
docker exec -i planet_postgres psql -U admin -d kids_planet < db/seeds/seed.sql
```

Seed bootstrap includes:

- admin user (`admin@kidsplanet.com`)
- age + education categories
- 3 active games linked to categories

### 3) Start full dev stack

```bash
docker compose up -d --build
docker compose ps
```

Expected services: `postgres`, `valkey`, `minio`, `api`, `web`, `swagger-ui`.

### 4) Verify health

```bash
curl -fsS http://localhost:8080/api/health
curl -I http://localhost:5173
docker exec planet_valkey valkey-cli ping
curl -fsS http://localhost:9000/minio/health/live
```

### Ports

| Service | Port |
| --- | --- |
| Web (dev) | `5173` |
| API | `8080` |
| Swagger UI | `8081` |
| Postgres | `5432` |
| Valkey | `6379` |
| MinIO API | `9000` |
| MinIO Console | `9001` |

### Production-style compose (Nginx entrypoint)

```bash
make prod   # uses infra/docker-compose.yml
make d      # stop infra stack
```

In infra mode, Nginx serves:

- Web on `http://localhost/`
- API proxy on `http://localhost/api/*`
- Game assets on `http://localhost/games/*`

## 7. API Overview

Base path: `/api` (behind Nginx proxy).  
Primary contract: `services/api/openapi/openapi.yaml`.

### Auth types

| Type | How | Used by |
| --- | --- | --- |
| None | no auth header | public catalog/categories/health, login endpoints |
| `BearerAuth` | `Authorization: Bearer <jwt>` | admin endpoints and player history |
| `PlayTokenAuth` | `Authorization: Bearer <play_token>` | leaderboard submit and self-rank (supported with play token) |

### Main API groups

- System: `GET /api/health`
- Public Games/Categories: `GET /api/games`, `GET /api/games/{id}`, `GET /api/categories`
- Sessions/Analytics: `POST /api/sessions/start`, `POST /api/analytics/event`
- Leaderboard: `POST /api/leaderboard/submit`, `GET /api/leaderboard/{game_id}`, `GET /api/leaderboard/{game_id}/self`
- Player Auth/History: `POST /api/auth/player/register`, `POST /api/auth/player/login`, `POST /api/auth/player/logout`, `GET /api/player/history`
- Admin Auth/Profile: `POST /api/auth/admin/login`, `GET /api/admin/ping`, `GET /api/admin/me`
- Admin Dashboard/Games/Categories:
  - `GET /api/admin/dashboard/overview`
  - `GET|POST /api/admin/games`
  - `PUT /api/admin/games/{id}`
  - `POST /api/admin/games/{id}/publish`
  - `POST /api/admin/games/{id}/unpublish`
  - `POST /api/admin/games/{id}/upload`
  - `GET|POST /api/admin/age-categories`, `PUT|DELETE /api/admin/age-categories/{id}`
  - `GET|POST /api/admin/education-categories`, `PUT|DELETE /api/admin/education-categories/{id}`

### Quick verification flows

- Public flow:
  1. `GET /api/games`
  2. `POST /api/sessions/start`
  3. `POST /api/analytics/event`
  4. `POST /api/leaderboard/submit`
  5. `GET /api/leaderboard/{game_id}`
- Admin flow:
  1. `POST /api/auth/admin/login`
  2. `GET /api/admin/dashboard/overview`

## 8. Game Delivery & Assets

Game upload contract (admin ZIP upload):

- ZIP must contain `index.html` at ZIP root
- No absolute paths, drive letters, or `..` segments
- Symlinks are rejected
- Nested directories are allowed (`assets/`, `js/`, `css/`)
- Allowed extension set is enforced by backend
- Limits:
  - request body max: 50MB
  - extracted total max: 200MB
  - file count max: 2000

Storage and delivery model:

- Extracted files are uploaded to `games/{id}/current/{relative_path}`
- Playable URL is `/games/{id}/current/index.html`
- Original ZIP archive is stored under `{id}/upload/<timestamp>_<random>.zip`

Common upload error codes: `INVALID_ZIP`, `INVALID_ZIP_PATH`, `ZIP_TOO_LARGE`, `ZIP_TOO_LARGE_UNCOMPRESSED`, `ZIP_TOO_MANY_FILES`, `INVALID_FILE_TYPE`, `MISSING_INDEX_HTML`.

## 9. Security Highlights

- JWT validation with issuer + signature checks for admin/player tokens
- Short-lived `play_token` with type and game/session claims for gameplay actions
- Strict admin boundary: `/api/admin/*` requires JWT + `admin` role
- CORS allowlist enforcement (`http://localhost`, `http://localhost:5173`, and `APP_ORIGIN` when set)
- Upload hardening against zip-slip, oversized payloads, decompression bombs, and disallowed file types
- Rate limiting:
  - leaderboard submit: Valkey-backed (30/min window)
  - analytics ingest: in-memory per-session limiter
- Nginx security headers on `/`, `/api/`, and `/games/` (CSP, `X-Content-Type-Options`, `X-Frame-Options`, etc.)
- Correlated error envelope with `request_id` and `X-Request-ID`

## 10. Performance Highlights

- Leaderboard reads are Valkey-backed (`ZREVRANGE` with score) for low-latency top-N retrieval
- Daily/weekly leaderboard keys use TTL (`DailyTTL`, `WeeklyTTL`) to control cardinality
- Postgres remains source of truth for sessions, analytics, and submissions
- Popular sort uses recent analytics (`game_start`) over a 7-day window
- Indexed schema for hot paths (games, submissions, analytics, sessions)
- Performance test suite (`tests/k6`) defines MVP SLO targets, including:
  - API read p95 <= 300ms
  - leaderboard read p95 <= 150ms
  - admin dashboard p95 <= 700ms
  - light upload p95 <= 5s

## 11. Player & Admin Roles

- Guest:
  - Can browse and play without registering
  - Uses session play token + `X-Guest-Id` for score submit
  - Appears in leaderboard as `g:<guest_id>`
- Player (optional account):
  - Registers/logs in with email + 6-digit PIN
  - Receives player JWT
  - Can access `GET /api/player/history`
  - Can be represented as `p:<player_id>` in leaderboard identity resolution
- Admin:
  - Logs in with admin credentials
  - Accesses `/api/admin/*` operations for dashboard, game lifecycle, category management, and ZIP upload

## 12. Development Workflow

Typical loop:

1. Start dependencies and app stack (`docker compose up -d --build`)
2. Apply baseline migration and seed if DB is fresh
3. Implement changes in:
   - `apps/web` (SvelteKit UI)
   - `services/api` (handlers/services/repos)
4. Validate behavior against API contract and runbook flows
5. Run checks relevant to your change:
   - API contract checks: `docs/API_TESTING.md`
   - Security checks: `docs/SECURITY_TESTING.md`
   - Performance checks: `docs/PERFORMANCE_TESTING.md`

Useful commands:

```bash
make dev         # dev compose stack
make ps          # list containers
make psql        # open Postgres shell
```

Frontend checks (inside `apps/web`):

```bash
npm run check
npm run lint
```

Backend checks (inside `services/api`):

```bash
go test ./...
```

## 13. Documentation Index

- [Architecture](docs/ARCHITECTURE.md)
- [Runbook](docs/RUNBOOK.md)
- [Game ZIP Integration](docs/GAME_INTEGRATION.md)
- [API Contract Testing](docs/API_TESTING.md)
- [Security Testing](docs/SECURITY_TESTING.md)
- [Performance Testing](docs/PERFORMANCE_TESTING.md)
- [OpenAPI Spec](services/api/openapi/openapi.yaml)
- [Contributing Guide](CONTRIBUTING.md)
- [Security Policy](SECURITY.md)
---
## 14. Technical Whitepaper

* [Architecture of a Scalable Web Based
  Educational Game Platform for Kids
  Planet](https://zenodo.org/records/18778097)

---

## 15. License

This project is licensed under the MIT License. See [LICENSE](LICENSE).
