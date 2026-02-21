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

### Login + Dashboard + Moderation action
- [ ] `POST /api/auth/admin/login` returns `200` with `data.access_token`
- [ ] `GET /api/admin/dashboard/overview` returns `200` with metrics (`sessions_today`, `top_games`, `total_active_games`, `total_players`)
- [ ] `GET /api/admin/moderation/flagged-submissions` returns `200` and flagged rows are visible
- [ ] `POST /api/admin/moderation/remove-score` with `submission_id` returns `200` and `{ "data": { "ok": true } }`
- [ ] Leaderboard no longer contains removed member after moderation

## Anti-Cheat & Rate Limit

### Burst submit -> 429
- [ ] Start one session and reuse same play token + guest id
- [ ] Send >30 `POST /api/leaderboard/submit` requests within 1 minute
- [ ] At least one request returns `429 RATE_LIMITED`

### High score -> flagged
- [ ] Submit score `> 1000000` (e.g. `1000001`) with valid play token + guest header
- [ ] Response stays success (`200`) but record is flagged in DB:

```sql
SELECT id, game_id, score, flagged, flag_reason, removed_at
FROM leaderboard_submissions
ORDER BY id DESC
LIMIT 20;
```

- [ ] `flagged = true` and `flag_reason` contains `score_out_of_bounds`

### Flagged appears in admin
- [ ] `GET /api/admin/moderation/flagged-submissions` includes that submission id

## Moderation

### Endpoint checks
- [ ] `GET /api/admin/moderation/flagged-submissions` (or alias `/api/admin/moderation/flagged`) returns `200`
- [ ] `POST /api/admin/moderation/remove-score` with body `{ "submission_id": "<id>" }` returns `200`

### DB fields updated
After remove-score:

```sql
SELECT id, flagged, flag_reason, removed_by_admin_id, removed_at, updated_at
FROM leaderboard_submissions
WHERE id = <submission_id>;
```

- [ ] `flagged = true`
- [ ] `flag_reason = 'removed_by_admin'`
- [ ] `removed_by_admin_id` is set
- [ ] `removed_at` is set

### Valkey ZREM effect
Member format is `g:<guest_id>`.

```bash
docker exec planet_valkey valkey-cli KEYS 'lb:game:*'
docker exec planet_valkey valkey-cli ZSCORE "lb:game:<game_id>:d:<yyyymmdd>" "g:<guest_id>"
docker exec planet_valkey valkey-cli ZSCORE "lb:game:<game_id>:w:<yyyyww>" "g:<guest_id>"
```

- [ ] `ZSCORE` is nil for removed member on active daily/weekly keys

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
- [ ] `admin flagged list` success (`200/201`)
- [ ] `remove score` success (`200/201`)
- [ ] `leaderboard read again` success (`200/201`)
- [ ] No unexpected 5xx responses in smoke run

## Freeze Exit Criteria
- [ ] All checklist items above pass
- [ ] No code/schema changes required to pass checks
- [ ] MVP is feature complete and frozen for polish/deploy phase
