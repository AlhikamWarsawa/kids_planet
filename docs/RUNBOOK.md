# MVP Runbook

## Scope
This runbook is for **stabilization + verification only**.
No feature work, no schema changes, no endpoint changes.

## Environment Setup

### 1) Database Setup (MVP)
- [ ] Start Postgres only: `docker compose up -d postgres`
- [ ] Apply baseline schema:

```bash
docker exec -i planet_postgres psql -U admin -d kids_planet < db/migrations/000_baseline.sql
```

- [ ] Apply MVP seed:

```bash
docker exec -i planet_postgres psql -U admin -d kids_planet < db/seeds/seed.sql
```

### 2) Start stack
- [ ] `docker compose up -d --build`
- [ ] `docker compose ps` shows `postgres`, `valkey`, `minio`, `api`, `web` as `Up` (and healthy where defined)

- [ ] Verify key tables exist:

```bash
docker exec -i planet_postgres psql -U admin -d kids_planet -c "\dt"
```

Expected MVP tables include: `games`, `sessions`, `analytics_events`, `leaderboard_submissions`.

### 3) Validate seed data
- [ ] Admin user exists (`admin@kidsplanet.com`)
- [ ] 3 active games exist for catalog/session/leaderboard tests

### 4) Service health
- [ ] API health: `curl -fsS http://localhost:8080/api/health`
- [ ] Web reachable: `curl -I http://localhost:5173`
- [ ] Valkey ping: `docker exec planet_valkey valkey-cli ping` returns `PONG`
- [ ] MinIO health: `curl -fsS http://localhost:9000/minio/health/live`

### 5) Ports
- [ ] Web: `5173`
- [ ] API: `8080`
- [ ] Postgres: `5432`
- [ ] Valkey: `6379`
- [ ] MinIO API: `9000`
- [ ] MinIO Console: `9001`
- [ ] Swagger UI: `8081`

### 6) Required env vars
- [ ] App: `ENV`, `PORT`
- [ ] Web/API origin: `APP_ORIGIN` (production web origin, e.g. `https://kidsplanet.example.com`)
- [ ] Postgres: `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_SSLMODE`
- [ ] Valkey: `VALKEY_ADDR`, `VALKEY_PASSWORD`, `VALKEY_DB`
- [ ] MinIO: `MINIO_ENDPOINT`, `MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY`, `MINIO_BUCKET`, `ZIP_UPLOAD_MAX_BYTES`
- [ ] JWT: `JWT_SECRET`, `JWT_ISSUER`, `JWT_EXPIRES_IN`

## Core Public Flow

### Catalog + Play + Analytics + Leaderboard
- [ ] `GET /api/games` returns `200` with `data.items`
- [ ] Game launch URL available from `GET /api/games/{id}` for active game
- [ ] `POST /api/sessions/start` returns `200` with `data.play_token`
- [ ] `POST /api/analytics/event` with `name=game_start` returns `200`
- [ ] Analytics event persists in DB:

```sql
SELECT id, session_id, game_id, event_name, created_at
FROM analytics_events
ORDER BY id DESC
LIMIT 10;
```

- [ ] `POST /api/leaderboard/submit` returns `200` with `data.accepted=true`
- [ ] `GET /api/leaderboard/{game_id}?period=daily&scope=game&limit=10` returns submitted member

## Admin Flow

### Login + Dashboard
- [ ] `POST /api/auth/admin/login` returns `200` with `data.access_token`
- [ ] `GET /api/admin/dashboard/overview` returns `200` with metrics (`sessions_today`, `top_games`, `total_active_games`, `total_players`)

## Rate Limit

### Burst submit -> 429
- [ ] Start one session and reuse same play token + guest id
- [ ] Send >30 `POST /api/leaderboard/submit` requests within 1 minute
- [ ] At least one request returns `429 RATE_LIMITED`

## Popular Sort

### Newest default
- [ ] `GET /api/games` and `GET /api/games?sort=newest` return same order (created_at desc)

### Popular changes order
- [ ] `GET /api/games?sort=popular` returns `200`
- [ ] With at least 2 active games, order differs from newest after generating analytics activity for one game

### Reflects analytics events
- [ ] Send multiple `game_start` events to target game via `/api/analytics/event`
- [ ] Re-run `GET /api/games?sort=popular`
- [ ] Target game rank increases (popularity derived from 7-day `analytics_events` count for `event_name='game_start'`)

## Smoke Test (Postman)

Collection file: `/tools/postman_collection.json`
Folder: **MVP Smoke Tests**

- [ ] Run smoke folder in order
- [ ] `start session` success (`200/201`) and play token extracted
- [ ] `track event` success (`200/201`)
- [ ] `submit score` success (`200/201`)
- [ ] `leaderboard read` success (`200/201`)
- [ ] No unexpected 5xx responses in smoke run

## Security Verification

### Zip-Slip path traversal
- [ ] Build ZIP with `../evil.txt` entry and upload via admin upload endpoint
- [ ] Upload is rejected with `422` and code `INVALID_ZIP_PATH`

### Oversized extracted ZIP (decompression bomb guard)
- [ ] Build ZIP whose total extracted file size is `> 200MB`
- [ ] Upload is rejected with `400` and code `ZIP_TOO_LARGE_UNCOMPRESSED`

### Too many files in ZIP
- [ ] Build ZIP containing `> 2000` files
- [ ] Upload is rejected with `400` and code `ZIP_TOO_MANY_FILES`

### Invalid file extension
- [ ] Add disallowed extension (example: `.exe`) into ZIP
- [ ] Upload is rejected with `422` and code `INVALID_FILE_TYPE`

### Root index validation
- [ ] ZIP without root-level `index.html` is rejected
- [ ] Error code is `MISSING_INDEX_HTML`

### Request body size limit
- [ ] Send `POST` body larger than `50MB` (upload/analytics/leaderboard submit)
- [ ] Request is rejected with `413` and code `ZIP_TOO_LARGE`

### Strict CORS
- [ ] Send request with disallowed `Origin` to `/api/*`
- [ ] Response is blocked (forbidden)
- [ ] Allowed origins are only `http://localhost`, `http://localhost:5173`, and `APP_ORIGIN`

### Nginx security headers
- [ ] Run `curl -I http://localhost/`
- [ ] Response contains:
  - [ ] `X-Content-Type-Options: nosniff`
  - [ ] `X-Frame-Options: SAMEORIGIN`
  - [ ] `Referrer-Policy: strict-origin-when-cross-origin`
  - [ ] `X-XSS-Protection: 0`
  - [ ] `Permissions-Policy`
  - [ ] `Content-Security-Policy`

## Freeze Exit Criteria
- [ ] All checklist items above pass
- [ ] No code/schema changes required to pass checks
- [ ] MVP is feature complete and frozen for polish/deploy phase
