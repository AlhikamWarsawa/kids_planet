# CHANGELOG_DEV

## Day 2: Docker Dev Stack & DB Connectivity (DONE)

**Date:** 15 December 2025

### Added

* Docker dev stack with Postgres, Valkey, and MinIO
* MinIO probe sidecar for reliable healthcheck
* Environment configuration via `.env.example`
* Makefile helpers (`make up`, `make ps`)

### Changed

* API startup flow now includes Postgres connection check (fail-fast)

### Verified

* All containers running and healthy via `docker compose ps`
* MinIO Console accessible on port 9001
* API `/health` endpoint responding normally

---

## Day 1: Kickstart & Skeleton (DONE)

**Date:** 14 December 2025

### Added

* Monorepo structure (`apps/web`, `services/api`, `infra`, `db`, `docs`, `tools`)
* Initial Go Fiber API with `/health` endpoint
* Initial SvelteKit web app
* Minimal README for local development
