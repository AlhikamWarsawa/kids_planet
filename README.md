# Kids Planet

This platform is a web based educational gaming portal for children, featuring a collection of HTML5 games that can be played directly in the browser (both desktop and mobile).

---

## 📁 Repository Structure (Overview)

```
.
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

## 🧠 Tech Stack

### Frontend

* **TypeScript**
* **SvelteKit**
* **Vite**
* **Nginx**

### Backend

* **Go (Golang)**
* **Fiber** (HTTP framework)
* **PostgreSQL** (Main Database)

### Dev & Infrastructure

* **Docker & Docker Compose**
* Environment-based configuration
* Planned integrations: **Valkey**, **MinIO**

---
## License

This project is licensed under the MIT License. See the [LICENCE](https://github.com/ZygmaCore/kids_planet/blob/main/LICENSE) file for full details.
