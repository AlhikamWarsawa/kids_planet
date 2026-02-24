# API Contract Testing (MVP)

## Source of Truth
This test guide is synchronized to:
- `services/api/openapi/openapi.yaml`

Contract tests should validate behavior against that OpenAPI file first.

## Scope
Covers all API operations under `/api` grouped by domain:
- Public
- Player
- Admin

Base URL conventions:
- Direct API container: `http://localhost:8080/api`
- Via Nginx proxy: `http://localhost/api`

```bash
export BASE_URL="http://localhost"
export API="$BASE_URL/api"
```

## Auth Modes

| Mode | Header | Used by |
| --- | --- | --- |
| None | N/A | public catalog/session/analytics/game detail routes and login routes |
| `BearerAuth` | `Authorization: Bearer <JWT>` | admin routes and player history |
| `PlayTokenAuth` | `Authorization: Bearer <play_token>` | leaderboard submit, optional on leaderboard self |

## Standard Envelopes

Error envelope (standardized):

```json
{
  "error": {
    "code": "STRING_CODE",
    "message": "Human readable",
    "request_id": "req_xxx"
  }
}
```

Success envelope (standard):

```json
{
  "data": {}
}
```

Known success-shape exceptions (intentional):
- `POST /api/auth/player/register` returns top-level `{ token, player }`
- `POST /api/auth/player/login` returns top-level `{ token, player }`
- `GET /api/player/history` returns top-level `{ data, pagination }`
- `POST /api/auth/player/logout` returns `204 No Content`

## Test Bootstrapping

```bash
# admin token
export ADMIN_TOKEN="$(curl -sS -X POST "$API/auth/admin/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@kidsplanet.com","password":"12345678"}' | jq -r '.data.access_token')"

# player token (register once or login)
export PLAYER_TOKEN="$(curl -sS -X POST "$API/auth/player/register" \
  -H 'Content-Type: application/json' \
  -d '{"email":"player.contract@example.com","pin":"123456"}' | jq -r '.token')"

# play token
export PLAY_TOKEN="$(curl -sS -X POST "$API/sessions/start" \
  -H 'Content-Type: application/json' \
  -d '{"game_id":1}' | jq -r '.data.play_token')"
```

## Endpoint Inventory

## Public Operations

| Endpoint | Method | Auth | Request | Response | Expected Errors |
| --- | --- | --- | --- | --- | --- |
| `/api/health` | GET | None | `fail?` query | `{data:{ok,service,time,env}}` | `400` |
| `/api/games` | GET | None | filters + pagination query | `{data:{items,page,limit,total}}` | `400`, `500` |
| `/api/games/{id}` | GET | None | `id` path | `{data:Game}` | `400`, `404`, `500` |
| `/api/categories` | GET | None | `type=age|education` query | `{data:{age_categories,education_categories}}` | `400`, `500` |
| `/api/sessions/start` | POST | Optional player JWT in header | `{game_id}` | `{data:{play_token,expires_at}}` | `400`, `401`, `500` |
| `/api/analytics/event` | POST | None (play token in body) | `{play_token,name,data?}` | `{data:{ok:true}}` | `400`, `401`, `429`, `500` |
| `/api/leaderboard/{game_id}` | GET | None | `period/scope/limit` query | `{data:{game_id,period,scope,limit,items}}` | `400`, `500` |

## Player Operations

| Endpoint | Method | Auth | Request | Response | Expected Errors |
| --- | --- | --- | --- | --- | --- |
| `/api/auth/player/register` | POST | None | `{email,pin}` | `{token,player:{id,email}}` | `400`, `500` |
| `/api/auth/player/login` | POST | None | `{email,pin}` | `{token,player:{id,email}}` | `400`, `401`, `500` |
| `/api/auth/player/logout` | POST | None | none | `204` | n/a |
| `/api/player/history` | GET | `BearerAuth` (player) | `page/limit` query | `{data:[...],pagination:{page,limit,total}}` | `400`, `401`, `500` |
| `/api/leaderboard/submit` | POST | `PlayTokenAuth` | header `X-Guest-Id`, body `{game_id,score}` | `{data:{accepted,best_score}}` | `400`, `401`, `403`, `429`, `500` |
| `/api/leaderboard/{game_id}/self` | GET | `BearerAuth` or `PlayTokenAuth` | `period/scope` query | `{data:{game_id,rank,score,period,scope}}` | `400`, `401`, `403`, `500` |

## Admin Operations

| Endpoint | Method | Auth | Request | Response | Expected Errors |
| --- | --- | --- | --- | --- | --- |
| `/api/auth/admin/login` | POST | None | `{email,password}` | `{data:{access_token,expires_in}}` | `400`, `401`, `403`, `500` |
| `/api/admin/ping` | GET | `BearerAuth` (admin) | none | `{data:{ok:true}}` | `401`, `403` |
| `/api/admin/me` | GET | `BearerAuth` (admin) | none | `{data:{id,email,role}}` | `401`, `403`, `500` |
| `/api/admin/dashboard/overview` | GET | `BearerAuth` (admin) | none | `{data:{sessions_today,total_active_games,total_players,top_games[]}}` | `401`, `403`, `500` |
| `/api/admin/games` | GET | `BearerAuth` (admin) | `status/q/page/limit` query | `{data:{items,page,limit,total}}` | `400`, `401`, `403`, `500` |
| `/api/admin/games` | POST | `BearerAuth` (admin) | create payload | `{data:AdminGame}` | `400`, `401`, `403`, `500` |
| `/api/admin/games/{id}` | PUT | `BearerAuth` (admin) | update payload | `{data:AdminGame}` | `400`, `401`, `403`, `404`, `500` |
| `/api/admin/games/{id}/publish` | POST | `BearerAuth` (admin) | none | `{data:AdminGame}` | `400`, `401`, `403`, `404`, `500` |
| `/api/admin/games/{id}/unpublish` | POST | `BearerAuth` (admin) | none | `{data:AdminGame}` | `400`, `401`, `403`, `404`, `500` |
| `/api/admin/games/{id}/upload` | POST | `BearerAuth` (admin) | multipart `file` | `{data:{object_key,etag,size,game_url}}` | `400`, `401`, `403`, `404`, `413`, `422`, `500` |
| `/api/admin/age-categories` | GET | `BearerAuth` (admin) | `q/page/limit` query | `{data:{items,page,limit}}` | `400`, `401`, `403`, `500` |
| `/api/admin/age-categories` | POST | `BearerAuth` (admin) | `{label,min_age,max_age}` | `{data:AgeCategoryWire}` | `400`, `401`, `403`, `500` |
| `/api/admin/age-categories/{id}` | PUT | `BearerAuth` (admin) | `{label?,min_age?,max_age?}` | `{data:AgeCategoryWire}` | `400`, `401`, `403`, `404`, `500` |
| `/api/admin/age-categories/{id}` | DELETE | `BearerAuth` (admin) | none | `{data:{deleted:true}}` | `400`, `401`, `403`, `404`, `500` |
| `/api/admin/education-categories` | GET | `BearerAuth` (admin) | `q/page/limit` query | `{data:{items,page,limit}}` | `400`, `401`, `403`, `500` |
| `/api/admin/education-categories` | POST | `BearerAuth` (admin) | `{name,icon?,color?}` | `{data:EducationCategoryWire}` | `400`, `401`, `403`, `500` |
| `/api/admin/education-categories/{id}` | PUT | `BearerAuth` (admin) | `{name?,icon?,color?}` | `{data:EducationCategoryWire}` | `400`, `401`, `403`, `404`, `500` |
| `/api/admin/education-categories/{id}` | DELETE | `BearerAuth` (admin) | none | `{data:{deleted:true}}` | `400`, `401`, `403`, `404`, `500` |
| `/api/admin/moderation/flagged-submissions` | GET | `BearerAuth` (admin) | `limit?` query | `{data:{items:[...]}}` | `400`, `401`, `403`, `500` |
| `/api/admin/moderation/flagged` | GET | `BearerAuth` (admin) | `limit?` query | `{data:{items:[...]}}` | `400`, `401`, `403`, `500` |
| `/api/admin/moderation/remove-score` | POST | `BearerAuth` (admin) | `{submission_id:number|string}` | `{data:{ok:true}}` | `400`, `401`, `403`, `404`, `500` |

## Key Curl Tests for New/Updated APIs

```bash
# admin ping
curl -si "$API/admin/ping" -H "Authorization: Bearer $ADMIN_TOKEN"

# moderation alias endpoint
curl -si "$API/admin/moderation/flagged?limit=20" -H "Authorization: Bearer $ADMIN_TOKEN"

# moderation canonical endpoint
curl -si "$API/admin/moderation/flagged-submissions?limit=20" -H "Authorization: Bearer $ADMIN_TOKEN"

# leaderboard self with play token
curl -si "$API/leaderboard/1/self?period=daily&scope=game" \
  -H "Authorization: Bearer $PLAY_TOKEN"

# upload with expected 200 schema fields
curl -si -X POST "$API/admin/games/1/upload" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -F 'file=@/tmp/k6-light-game.zip;type=application/zip'
```

## Schema-Focused Assertions

## Game list/detail
- Each item includes required fields from OpenAPI `Game` schema.
- `thumbnail` and `game_url` may be `null`.

## Category wire-format (admin)
- Age category item shape follows wire keys: `ID`, `Label`, `MinAge`, `MaxAge`, `CreatedAt`.
- Education category item shape follows wire keys: `ID`, `Name`, `Icon`, `Color`, `CreatedAt`.
- `Icon` and `Color` are SQL-null wrappers: `{String,Valid}`.

## Leaderboard
- `GET /leaderboard/{game_id}` returns `items[]` of `{member,score}`.
- `GET /leaderboard/{game_id}/self` returns `rank` and `score` nullable.

## Upload
- `POST /admin/games/{id}/upload` success includes `object_key`, `etag`, `size`, `game_url`.
- failure contracts include `413 ZIP_TOO_LARGE` and `422` for semantic ZIP validation.

## Negative Contract Tests

## Invalid token
```bash
curl -si "$API/admin/me" -H 'Authorization: Bearer invalid.token.value'
```
Expect: `401 UNAUTHORIZED`, valid error envelope with `request_id`.

## Missing fields
```bash
curl -si -X POST "$API/sessions/start" -H 'Content-Type: application/json' -d '{}'
```
Expect: `400 BAD_REQUEST`.

## Invalid IDs
```bash
curl -si "$API/games/0"
```
Expect: `400 BAD_REQUEST`.

## Permission mismatch
```bash
curl -si "$API/admin/dashboard/overview"
```
Expect: `401 UNAUTHORIZED`.

## Postman / Newman Guidance

## Postman
1. Import `services/api/openapi/openapi.yaml`.
2. Define env vars: `base_url`, `admin_token`, `player_token`, `play_token`, `game_id`.
3. Add auth pre-requests:
- admin login -> `admin_token`
- sessions start -> `play_token`
4. Add tests for:
- status code
- required response fields
- standardized error envelope with `request_id`

## Newman
```bash
newman run ./tests/postman/kids-planet-api.postman_collection.json \
  -e ./tests/postman/local.postman_environment.json \
  --bail \
  --reporters cli,junit
```

## Contract Consistency Checklist

- Every implemented route appears in OpenAPI and in this test matrix.
- Protected routes enforce declared auth mode.
- Error responses include `error.code`, `error.message`, `error.request_id`.
- `X-Request-ID` is present and correlates with error payload request id.
- Status codes match contract (`400/401/403/404/413/422/429/500`).
- Success payload shape matches OpenAPI schema and documented envelope exceptions.
