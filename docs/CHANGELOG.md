# CHANGELOG_DEV

## Day 3: Database Foundation + Repo Scaffolding (DONE)

**Date:** 16 January 2026

### Added

* Core database schema based on ERD:

    * `users`, `players`
    * `education_categories`, `age_categories`
    * `games`, `game_education_categories`
* Initial indexes for performance readiness:

    * `games(status)`
    * `games(age_category_id)`
    * `game_education_categories(game_id)`
    * `game_education_categories(education_category_id)`
* Seed data for faster development:

    * Default admin user
    * Age categories (3+, 5+, 7+, 10+)
    * Education categories (Math, Reading, Logic, Memory, Creativity)
    * One draft game (`Color Match`)
* Repository scaffolding in API layer:

    * `user_repo.go` (find user by email)
    * `game_repo.go` (basic listing)

### Changed

* API wiring updated to inject database connection into repo layer
* Removed any direct SQL usage from `main.go` (clean separation of concerns)

### Verified

* Migrations run cleanly and reproducibly (`migrate up`)
* Database schema and indexes validated via `psql` (`\dt`, `\di`)
* Seed data successfully inserted and queryable
* API starts with Postgres connected (fail-fast behavior)
* `/health` endpoint remains functional

---

## Day 2: Docker Dev Stack & DB Connectivity (DONE)

**Date:** 15 January 2026

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

**Date:** 14 January 2026

### Added

* Monorepo structure (`apps/web`, `services/api`, `infra`, `db`, `docs`, `tools`)
* Initial Go Fiber API with `/health` endpoint
* Initial SvelteKit web app (placeholder home)
* Minimal README for local development
