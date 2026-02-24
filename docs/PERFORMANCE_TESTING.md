# Performance Testing Suite (MVP)

## Scope
This suite validates performance characteristics for the implemented MVP stack:
- API under `/api` (Go Fiber)
- dashboard/admin endpoints
- leaderboard reads via Valkey
- analytics event ingest path (`sessions` -> `analytics_events`)
- light ZIP upload path to MinIO via admin endpoint

## Target SLOs

| Metric | Target |
| --- | --- |
| API read endpoints p95 (`/api/games`, `/api/leaderboard/{game_id}`) | <= 300 ms |
| Admin dashboard p95 (`/api/admin/dashboard/overview`) | <= 700 ms |
| Upload endpoint p95 (`/api/admin/games/{id}/upload`, light ZIP <= 5MB) | <= 5 s |
| Leaderboard read p95 (`/api/leaderboard/{game_id}`) | <= 150 ms |
| Error rate per scenario | < 1% (`upload_light` < 5%) |

## k6 Scripts (Ready to Run)

Stored under `tests/k6/`:
- `tests/k6/admin_dashboard_overview.js`
- `tests/k6/games_catalog_list.js`
- `tests/k6/leaderboard_read.js`
- `tests/k6/analytics_event_ingest.js`
- `tests/k6/upload_light.js`
- shared helpers: `tests/k6/common.js`

## Prerequisites

```bash
# stack up
make dev

# install k6 (macOS example)
brew install k6
```

### Optional env overrides
- `BASE_URL` default: `http://localhost`
- `ADMIN_TOKEN` (or `ADMIN_EMAIL` + `ADMIN_PASSWORD`)
- `GAME_ID` default: `1`
- `VUS`, `DURATION`

### Upload fixture for `upload_light.js`

```bash
mkdir -p /tmp/k6-light-game
cat > /tmp/k6-light-game/index.html <<'HTML'
<!doctype html><html><head><meta charset="utf-8"><title>k6</title></head><body>k6 upload test</body></html>
HTML
(cd /tmp/k6-light-game && zip -qr /tmp/k6-light-game.zip .)
```

## Scenario Runs

### 1) Admin dashboard overview
```bash
k6 run \
  -e BASE_URL=http://localhost \
  -e ADMIN_EMAIL=admin@kidsplanet.com \
  -e ADMIN_PASSWORD=12345678 \
  tests/k6/admin_dashboard_overview.js
```

### 2) Games catalog list
```bash
k6 run \
  -e BASE_URL=http://localhost \
  tests/k6/games_catalog_list.js
```

### 3) Leaderboard read
```bash
k6 run \
  -e BASE_URL=http://localhost \
  -e GAME_ID=1 \
  tests/k6/leaderboard_read.js
```

### 4) Analytics event ingest
```bash
k6 run \
  -e BASE_URL=http://localhost \
  -e GAME_ID=1 \
  tests/k6/analytics_event_ingest.js
```

### 5) Upload endpoint (light)
```bash
k6 run \
  -e BASE_URL=http://localhost \
  -e ADMIN_EMAIL=admin@kidsplanet.com \
  -e ADMIN_PASSWORD=12345678 \
  -e GAME_ID=1 \
  -e ZIP_FILE=/tmp/k6-light-game.zip \
  tests/k6/upload_light.js
```

## Thresholds and Pass Criteria

| Scenario | Key Thresholds | Pass Criteria |
| --- | --- | --- |
| `admin_dashboard_overview` | p95 < 700ms, avg < 400ms, failed < 1% | all thresholds pass |
| `games_catalog_list` | p95 < 300ms, avg < 180ms, failed < 1% | all thresholds pass |
| `leaderboard_read` | p95 < 150ms, avg < 100ms, failed < 1% | all thresholds pass |
| `analytics_event_ingest` | p95 < 300ms, avg < 200ms, failed < 2% | all thresholds pass |
| `upload_light` | p95 < 5000ms, avg < 3000ms, failed < 5% | all thresholds pass |

Global fail conditions:
- any scenario has sustained `http_req_failed` above threshold
- p95 latency exceeds target in two consecutive test runs
- any unexpected `5xx` burst appears under nominal load

## Result Interpretation Guide

1. `http_req_failed` rises first
- likely causes: auth/test setup errors, rate limits hit unintentionally, dependency outages
- action: inspect response body and `request_id`, verify token/session setup

2. p95 latency high, avg moderate
- likely causes: tail latency from DB calls or object storage round-trips
- action: inspect slow endpoint logs by `request_id`; compare with Postgres and MinIO health

3. high latency plus high failure
- likely causes: infrastructure saturation (DB connection pressure, Valkey pressure, MinIO slow writes)
- action: reduce load to isolate bottleneck, then scale targeted component

4. analytics ingest failures with `429`
- likely cause: repeated events against same play session
- action: ensure each iteration uses a fresh `/sessions/start` token (current script already does this)

## Scaling Notes

### Postgres
- verify index coverage for hot reads:
  - `games(status)` and `games(created_at)`
  - `leaderboard_submissions(game_id, score)`
  - `analytics_events(event_name, created_at, game_id)`
- keep connection pool sized to workload; monitor lock waits and sequential scans

### Valkey
- leaderboard read is Valkey-backed; watch memory and key cardinality (`lb:*`, `rl:*`, `ac:*`)
- validate TTL behavior to avoid stale/high-cardinality key buildup

### MinIO
- upload path writes extracted files plus original archive object
- for higher throughput:
  - keep object keys partitioned by game prefix (`{game_id}/current/...`)
  - monitor disk IOPS and network between API and MinIO
  - keep light upload tests separate from stress tests on large assets

### Nginx/API
- keep `client_max_body_size` and API `BodyLimit` aligned (currently 50MB)
- monitor request timeout settings for upload and long-tail responses

## Reporting Template

For each run capture:
- commit SHA / branch
- environment (`dev compose` or `infra compose`)
- script name + env vars
- k6 summary output
- pass/fail versus SLO table
- regressions with linked issue and `request_id` samples
