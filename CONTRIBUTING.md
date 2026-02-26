# Contributing

Thank you for contributing to Kids Planet. This guide defines the workflow,
quality bar, and security rules for production-grade contributions.

## Project Overview

Kids Planet is a web-based educational game portal with:
- SvelteKit frontend (`apps/web`)
- Go Fiber API (`services/api`)
- PostgreSQL for source-of-truth data (`games`, `sessions`, `analytics_events`, `leaderboard_submissions`)
- Valkey for leaderboard and rate limiting
- MinIO games bucket for game assets and uploaded ZIP packages
- Nginx reverse proxy for `/api` and `/games`
- Docker Compose for local dev and production-like infra

## Architecture Summary

Primary flows:
- Game catalog and play: `/api/games`, `/api/sessions/start`, `/games/{id}/current/index.html`
- Analytics ingestion: `/api/analytics/event` -> `analytics_events`
- Leaderboard submit/read: `/api/leaderboard/submit`, `/api/leaderboard/{game_id}`
- Admin operations: `/api/admin/*` (JWT admin auth)

Core docs:
- `docs/ARCHITECTURE.md`
- `docs/RUNBOOK.md`
- `docs/GAME_INTEGRATION.md`
- `docs/SECURITY_TESTING.md`
- `docs/PERFORMANCE_TESTING.md`
- `docs/API_TESTING.md`

## Local Setup (Docker Dev Compose)

Prerequisites:
- Docker + Docker Compose
- Make

Steps:

```bash
cp .env.example .env
make dev
```

This starts:
- `postgres` (5432)
- `valkey` (6379)
- `minio` (9000/9001)
- `api` (8080)
- `web` (5173)

Seed/DB bootstrap is documented in `docs/RUNBOOK.md`.

## Coding Standards

## Go (API)
- Run formatting and vet/lint checks before PR.
- Keep handlers thin; put business logic in services and data logic in repos.
- Return consistent error envelope via shared utilities.
- Maintain strict validation for all public/admin inputs.
- Preserve request correlation (`X-Request-ID` and `error.request_id`).

## Svelte + TypeScript (Web)
- Use strict TypeScript types and avoid `any` unless justified.
- Keep API contracts in sync with backend payload shape.
- Reuse shared API client and typed DTOs in `apps/web/src/lib`.
- Avoid introducing new global state when existing stores are sufficient.

## Commit Convention

Use Conventional Commits:

- `feat(api): add admin game publish validation`
- `fix(web): handle player history empty state`
- `docs(security): extend zip-slip test matrix`
- `refactor(api): simplify leaderboard key resolver`
- `test(api): add negative auth coverage`

Rules:
- one logical change per commit
- imperative, concise subject
- include scope when practical (`api`, `web`, `docs`, `infra`, `db`)

## Branch Strategy

- `main` is the protected release branch.
- Work in short-lived branches from `main`:
  - `feature/<topic>`
  - `fix/<topic>`
  - `docs/<topic>`
  - `security/<topic>`
- Keep branches rebased on latest `main` before merge.
- Use draft PRs early for visibility on large changes.

## Pull Request Process

1. Open PR with clear summary and affected areas.
2. Link issue(s) and include migration/security impact when relevant.
3. Attach testing evidence (commands + results).
4. Address review comments with follow-up commits.
5. Merge only after required checks and approvals pass.

Use `.github/pull_request_template.md` checklist for every PR.

## Testing Requirements

Minimum before merge:
- run API and web tests relevant to your change
- run contract checks from `docs/API_TESTING.md` for touched endpoints
- run security checks from `docs/SECURITY_TESTING.md` for auth/upload/proxy changes
- run performance checks from `docs/PERFORMANCE_TESTING.md` for latency-sensitive paths

At minimum include in PR description:
- commands executed
- pass/fail outcome
- any known gap and follow-up issue

## Security Rules (Mandatory)

## Uploads
- only ZIP uploads through `/api/admin/games/{id}/upload`
- keep validation intact: path safety, extension allowlist, uncompressed/file-count limits, root `index.html`
- preserve body-size controls (50MB API/Nginx alignment)

## Auth
- do not bypass or weaken admin/player/play-token middleware
- preserve issuer/signature checks for JWT
- maintain explicit role checks on `/api/admin/*`

## Storage
- do not expose unintended MinIO object paths
- preserve `/games/{id}/current/...` serving model
- never commit secrets, credentials, or real JWT keys

## Reporting Security Issues

Do not open public issues for sensitive vulnerabilities.
Follow `SECURITY.md` for private disclosure workflow.

## Documentation Expectations

When behavior changes, update docs in the same PR:
- API behavior -> `docs/API_TESTING.md`
- security-relevant behavior -> `docs/SECURITY_TESTING.md`
- latency/perf behavior -> `docs/PERFORMANCE_TESTING.md`
- operational behavior -> `docs/RUNBOOK.md`
