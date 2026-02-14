<h1 align="center">Kids_Planet</h1>

<p align="center">
  <img src="https://img.shields.io/badge/license-MIT-blue.svg" />
  <img src="https://img.shields.io/badge/language-Go-blue.svg" />
</p>

<p align="center">
  <img src="img/icon.png" alt="Logo" width="180">
</p>

<h3 align="center">
This platform is a web based educational gaming portal for children, featuring a collection of HTML5 games that can be played directly in the browser<br> (both desktop and mobile).
</h3>


---

## Repository Structure

```
kids-planet
├── apps/
│   └── web/              # Frontend (SvelteKit)
├── services/
│   └── api/              # Backend API (Go Fiber)
├── infra/                # Infrastructure configs (Docker, Nginx, etc.)
├── db/                   # Database migrations & seeds (next phase)
├── docs/                 # Documentation (architecture, runbook, etc.)
├── tools/                # Tooling (Postman, load testing, etc.)
├── docker-compose.yml
└── README.md
```

---

## Tech Stack

### Frontend (Player & Admin)

* **TypeScript**: type-safe frontend development
* **SvelteKit**: main frontend framework (SPA static build)
* **Vite**: fast bundler & dev server
* **Nginx**: static file serving & reverse proxy

### Backend (API Service)

* **Go (Golang)**: core backend language
* **Fiber**: high-performance HTTP framework
* **PostgreSQL**: primary relational database (metadata, events, submissions)

### Dev & Infrastructure

* **Docker & Docker Compose**: local development & deployment
* **Valkey**: in-memory store (leaderboard, rate limiting, cache)
* **MinIO**: object storage (HTML5 game assets & uploads)
* **Environment-based configuration** (`.env`)

---

## Development

```bash
make dev
```

---

## Production
```bash
make prod
```

---

## Documentation

* `docs/ARCHITECTURE.md`: system architecture overview
* `docs/RUNBOOK.md`: how to run & troubleshoot the system
* `docs/GAME_INTEGRATION.md`: A technical guide for game developers to integrate their HTML5 games with the platform.

---
## License

This project is licensed under the MIT License. See the [LICENCE](https://github.com/ZygmaCore/kids_planet/blob/main/LICENSE) file for full details.
