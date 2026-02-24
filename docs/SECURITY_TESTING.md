# Security Testing Suite (MVP)

## Purpose
This suite validates security controls for the Kids Planet MVP architecture:
- `SvelteKit` web front end behind `Nginx`
- `Go Fiber` API under `/api`
- game delivery under `/games` from MinIO games bucket
- JWT auth (`admin` and optional `player`), short-lived `play_token`
- Postgres (`sessions`, `analytics_events`, `leaderboard_submissions`), Valkey leaderboard and rate limiting

## Threat Model Overview

### Protected Assets
- Admin credentials and admin-only endpoints
- JWT secrets and token integrity
- Game package integrity and execution boundary
- Player/session identifiers and request correlation (`request_id`)
- Leaderboard fairness and anti-cheat controls
- Analytics pipeline integrity
- Object storage data exposure scope

### Trust Boundaries
- Internet client -> Nginx (`/api`, `/games`, `/`)
- Nginx -> API service
- API -> Postgres / Valkey / MinIO
- Admin browser -> upload pipeline (ZIP -> extract -> validate -> MinIO)

### Primary Threat Scenarios
- untrusted ZIP upload abuse (zip-slip, zip bomb, malicious file types)
- auth bypass, JWT tampering, role confusion
- leaderboard abuse (flood, forged tokens, game mismatch)
- analytics ingestion abuse
- CORS/CSP misconfiguration
- path traversal at proxy/object route layer
- object bucket enumeration/data exposure
- error/request trace leakage

## Attack Surface Inventory

| Surface | Entry Points | Main Risks | Existing Controls |
| --- | --- | --- | --- |
| Upload ZIP pipeline | `POST /api/admin/games/{id}/upload` | zip-slip, decompression bomb, non-web payload upload, oversized payload | 50MB body limit, `.zip` validation, safe path checks, extension allowlist, uncompressed size/file-count guard, root `index.html` required |
| Public API | `/api/games`, `/api/categories`, `/api/sessions/start`, `/api/leaderboard/{game_id}` | parameter abuse, scraping, malformed JSON | input validation, typed parsing, error envelope |
| Auth (JWT admin/player) | `/api/auth/admin/login`, `/api/auth/player/*`, protected endpoints | token forgery/tampering/replay, role bypass | HS256 validation, issuer check, role checks, middleware separation |
| Leaderboard submit | `POST /api/leaderboard/submit` | score fraud, token mismatch, request flood | `play_token` middleware, game-id match, Valkey rate limit, anti-cheat flags |
| Analytics ingest | `POST /api/analytics/event` | event spam, invalid token/data poisoning | play token parse and claim checks, per-session in-memory rate limit, JSON validation |
| Nginx proxy | `/api/*`, `/games/*` | header misconfig, traversal attempts, cache abuse | proxy isolation, security headers, explicit locations |
| MinIO games bucket | `/games/{id}/...` | bucket listing/object exposure | anonymous download policy (intentional read), route scoping, non-guessable upload key suffix |
| Error tracing | API error responses/logging | sensitive data leakage, trace injection | structured error envelope with `request_id`, generic error messages |

## Environment and Prerequisites

```bash
# start stack
make dev

# baseline variables
export BASE_URL="http://localhost"
export API="$BASE_URL/api"
export ADMIN_EMAIL="admin@kidsplanet.com"
export ADMIN_PASSWORD="12345678"
export GAME_ID="1"
```

```bash
# admin token
export ADMIN_TOKEN="$(curl -sS -X POST "$API/auth/admin/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"'"$ADMIN_EMAIL"'","password":"'"$ADMIN_PASSWORD"'"}' | jq -r '.data.access_token')"

# play token for leaderboard/analytics tests
export PLAY_TOKEN="$(curl -sS -X POST "$API/sessions/start" \
  -H 'Content-Type: application/json' \
  -d '{"game_id":'"$GAME_ID"'}' | jq -r '.data.play_token')"

# optional player token
export PLAYER_TOKEN="$(curl -sS -X POST "$API/auth/player/register" \
  -H 'Content-Type: application/json' \
  -d '{"email":"security.player@example.com","pin":"123456"}' | jq -r '.token')"
```

## Error Contract and Status Expectations

### Error Envelope
```json
{
  "error": {
    "code": "BAD_REQUEST",
    "message": "...",
    "request_id": "..."
  }
}
```

### Expected Security-Relevant Codes

| HTTP | `error.code` |
| --- | --- |
| 400 | `BAD_REQUEST`, `INVALID_ZIP`, `ZIP_TOO_LARGE_UNCOMPRESSED`, `ZIP_TOO_MANY_FILES`, `MISSING_INDEX_HTML` |
| 401 | `UNAUTHORIZED` |
| 403 | `FORBIDDEN` |
| 404 | `RESOURCE_NOT_FOUND` |
| 413 | `ZIP_TOO_LARGE` |
| 422 | `INVALID_ZIP_PATH`, `INVALID_FILE_TYPE` |
| 429 | `RATE_LIMITED` |
| 500 | `INTERNAL_ERROR` |

## Security Test Cases

| ID | Test | Severity | Expected Result |
| --- | --- | --- | --- |
| SEC-01 | Zip-slip payload rejected | Critical | `422 INVALID_ZIP_PATH` |
| SEC-02 | Zip bomb (uncompressed > 200MB) rejected | Critical | `400 ZIP_TOO_LARGE_UNCOMPRESSED` |
| SEC-03 | Too many files in ZIP rejected | High | `400 ZIP_TOO_MANY_FILES` |
| SEC-04 | Disallowed file extension in ZIP rejected | High | `422 INVALID_FILE_TYPE` |
| SEC-05 | Missing root `index.html` rejected | High | `400 MISSING_INDEX_HTML` |
| SEC-06 | Oversized upload (`>50MB`) blocked | High | `413 ZIP_TOO_LARGE` |
| SEC-07 | Disallowed CORS origin blocked | High | `403 FORBIDDEN` |
| SEC-08 | CSP header present and restrictive | High | `200`, CSP header contains restrictive directives |
| SEC-09 | Admin auth bypass blocked | Critical | `401 UNAUTHORIZED` or `403 FORBIDDEN` |
| SEC-10 | JWT tampering blocked | Critical | `401 UNAUTHORIZED` |
| SEC-11 | Leaderboard submit flood rate-limited | High | `429 RATE_LIMITED` after threshold |
| SEC-12 | Path traversal via `/games` blocked | High | no file disclosure (`404`/non-success) |
| SEC-13 | MinIO bucket/object enumeration blocked | High | listing denied, unknown object not exposed |
| SEC-14 | `request_id` correlation without sensitive leakage | Medium | request id echoed only as trace metadata |
| SEC-15 | Analytics invalid token blocked | High | `401 UNAUTHORIZED` |
| SEC-16 | Leaderboard game mismatch blocked | High | `403 FORBIDDEN` |

## Critical Test Execution (Curl)

### SEC-01 Zip-slip
Description: verifies extraction path traversal defense in upload pipeline.

```bash
python3 - <<'PY'
import zipfile
with zipfile.ZipFile('/tmp/zip-slip.zip', 'w', zipfile.ZIP_DEFLATED) as z:
    z.writestr('index.html', '<html>ok</html>')
    z.writestr('../escape.txt', 'blocked')
PY

curl -si -X POST "$API/admin/games/$GAME_ID/upload" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -F "file=@/tmp/zip-slip.zip;type=application/zip"
```

Expected:
- HTTP `422`
- `error.code = INVALID_ZIP_PATH`

### SEC-02 Zip bomb
Description: verifies uncompressed-size safety guard.

```bash
python3 - <<'PY'
import zipfile
payload = b'A' * (210 * 1024 * 1024)
with zipfile.ZipFile('/tmp/zip-bomb.zip', 'w', zipfile.ZIP_DEFLATED) as z:
    z.writestr('index.html', '<html>ok</html>')
    z.writestr('assets/huge.txt', payload)
PY

curl -si -X POST "$API/admin/games/$GAME_ID/upload" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -F "file=@/tmp/zip-bomb.zip;type=application/zip"
```

Expected:
- HTTP `400`
- `error.code = ZIP_TOO_LARGE_UNCOMPRESSED`

### SEC-03 Too many files in ZIP
Description: verifies file-count guard.

```bash
python3 - <<'PY'
import zipfile
with zipfile.ZipFile('/tmp/zip-many-files.zip', 'w', zipfile.ZIP_DEFLATED) as z:
    z.writestr('index.html', '<html>ok</html>')
    for i in range(0, 2101):
        z.writestr(f'assets/f{i}.txt', 'x')
PY

curl -si -X POST "$API/admin/games/$GAME_ID/upload" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -F "file=@/tmp/zip-many-files.zip;type=application/zip"
```

Expected:
- HTTP `400`
- `error.code = ZIP_TOO_MANY_FILES`

### SEC-04 Invalid file type in ZIP
Description: blocks non-allowlisted extensions.

```bash
python3 - <<'PY'
import zipfile
with zipfile.ZipFile('/tmp/zip-invalid-ext.zip', 'w', zipfile.ZIP_DEFLATED) as z:
    z.writestr('index.html', '<html>ok</html>')
    z.writestr('payload.exe', 'MZ')
PY

curl -si -X POST "$API/admin/games/$GAME_ID/upload" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -F "file=@/tmp/zip-invalid-ext.zip;type=application/zip"
```

Expected:
- HTTP `422`
- `error.code = INVALID_FILE_TYPE`

### SEC-05 Missing root index.html
Description: ensures playable root entry requirement.

```bash
python3 - <<'PY'
import zipfile
with zipfile.ZipFile('/tmp/zip-no-root-index.zip', 'w', zipfile.ZIP_DEFLATED) as z:
    z.writestr('game/index.html', '<html>nested</html>')
PY

curl -si -X POST "$API/admin/games/$GAME_ID/upload" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -F "file=@/tmp/zip-no-root-index.zip;type=application/zip"
```

Expected:
- HTTP `400`
- `error.code = MISSING_INDEX_HTML`

### SEC-06 Oversized upload > 50MB
Description: validates global body-limit middleware and Nginx limit.

```bash
dd if=/dev/zero of=/tmp/oversize.zip bs=1M count=51

curl -si -X POST "$API/admin/games/$GAME_ID/upload" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -F "file=@/tmp/oversize.zip;type=application/zip"
```

Expected:
- HTTP `413`
- `error.code = ZIP_TOO_LARGE`

### SEC-07 CORS violation
Description: rejects unauthorized browser origins.

```bash
curl -si -X OPTIONS "$API/games" \
  -H 'Origin: https://evil.example' \
  -H 'Access-Control-Request-Method: GET'
```

Expected:
- HTTP `403`
- `error.code = FORBIDDEN`

### SEC-08 CSP enforcement
Description: verifies CSP is set on proxied routes.

```bash
curl -si "$BASE_URL/api/health" | grep -i 'Content-Security-Policy'
curl -si "$BASE_URL/games/$GAME_ID/current/index.html" | grep -i 'Content-Security-Policy'
```

Expected:
- HTTP `200` on reachable target
- response includes `Content-Security-Policy` header
- policy should not include permissive wildcards like `*` for `script-src`/`connect-src`

### SEC-09 Auth bypass attempts
Description: verifies protected endpoints reject missing/wrong auth.

```bash
# no token
curl -si "$API/admin/dashboard/overview"

# player token on admin route
curl -si "$API/admin/me" -H "Authorization: Bearer $PLAYER_TOKEN"
```

Expected:
- HTTP `401` (or `403` when role is known but insufficient)
- `error.code` is `UNAUTHORIZED` or `FORBIDDEN`

### SEC-10 JWT tampering
Description: signature/claims tampering should fail.

```bash
BAD_TOKEN="${ADMIN_TOKEN%?}x"
curl -si "$API/admin/me" -H "Authorization: Bearer $BAD_TOKEN"
```

Expected:
- HTTP `401`
- `error.code = UNAUTHORIZED`

### SEC-11 Rate-limit bypass on leaderboard submit
Description: verifies Valkey-backed rate limiting for submit endpoint.

```bash
for i in $(seq 1 35); do
  curl -s -o /tmp/rl_$i.out -w "%{http_code}\n" -X POST "$API/leaderboard/submit" \
    -H "Authorization: Bearer $PLAY_TOKEN" \
    -H "X-Guest-Id: rl-test" \
    -H "Content-Type: application/json" \
    -d '{"game_id":1,"score":123}'
done

grep -R "RATE_LIMITED" /tmp/rl_*.out
```

Expected:
- at least one request returns HTTP `429`
- response includes `error.code = RATE_LIMITED`

### SEC-12 Path traversal via reverse proxy route
Description: verifies `/games` route cannot read arbitrary filesystem/object paths.

```bash
curl -si "$BASE_URL/games/%2e%2e/%2e%2e/etc/passwd"
curl -si "$BASE_URL/games/../../../../etc/passwd"
```

Expected:
- no sensitive file disclosure
- non-success response (`404`/`400`/`403`), never returns host file contents

### SEC-13 MinIO object exposure
Description: checks public bucket behavior is constrained to intended objects.

```bash
# bucket listing probe
curl -si "$BASE_URL/games/?list-type=2"

# unknown object probe
curl -si "$BASE_URL/games/999999/current/index.html"
```

Expected:
- bucket listing should be denied or non-enumerable (no object index leak)
- unknown object should not return content (`404`/`403`)

### SEC-14 request_id leakage check
Description: ensures trace IDs are correlation-only and do not leak internals.

```bash
curl -si "$API/admin/me" \
  -H 'X-Request-ID: sec-review-req-001'
```

Expected:
- HTTP `401`
- header `X-Request-ID: sec-review-req-001`
- JSON includes `error.request_id = sec-review-req-001`
- `error.message` contains no SQL fragments, stack traces, filesystem paths, or secrets

### SEC-15 Analytics invalid token
Description: analytics requires valid play token.

```bash
curl -si -X POST "$API/analytics/event" \
  -H 'Content-Type: application/json' \
  -d '{"play_token":"not-a-token","name":"game_start"}'
```

Expected:
- HTTP `401`
- `error.code = UNAUTHORIZED`

### SEC-16 Leaderboard token/game mismatch
Description: play token for game A cannot submit for game B.

```bash
curl -si -X POST "$API/leaderboard/submit" \
  -H "Authorization: Bearer $PLAY_TOKEN" \
  -H 'X-Guest-Id: mismatch-test' \
  -H 'Content-Type: application/json' \
  -d '{"game_id":999999,"score":10}'
```

Expected:
- HTTP `403`
- `error.code = FORBIDDEN`

## Pass/Fail Criteria
- `Critical` tests must pass 100% before public release.
- `High` tests must pass or have an accepted, documented mitigation in `SECURITY.md`.
- any unexpected `5xx` during negative security tests is an automatic fail.
- every failed test must capture full response headers/body and `request_id` for traceability.

## Evidence to Store
- command, timestamp, environment (`dev`/`prod` compose)
- exact HTTP response (status, headers, JSON body)
- associated `request_id`
- remediation issue/PR link when failing
