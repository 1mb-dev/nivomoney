# Nivo

A neobank demo showing fintech architecture with Go microservices.

[![Documentation](https://img.shields.io/badge/docs-nivomoney.com-blue)](https://nivomoney.com/docs)
[![Go](https://img.shields.io/badge/go-1.26+-00ADD8)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Status](https://img.shields.io/badge/status-archived%20showcase-orange)]()

## Overview

Nivo is a portfolio project implementing the core systems of a digital bank. It demonstrates microservices architecture, double-entry accounting, and fintech domain patterns in a working, deployable application.

> **Archived Showcase** — The live demo at nivomoney.com has been retired. The domain now serves a [static landing page](https://nivomoney.com). Run the platform locally for the full demo experience.

**What it includes:**
- 8 Go services and an API gateway, with domain-driven boundaries
- Double-entry ledger with balanced journal entries
- JWT authentication with role-based access control
- React frontends for users and admins
- Full observability stack (Prometheus, Grafana)

## Quick Start

### Prerequisites

- Go 1.26+
- Docker & Docker Compose
- Node.js 18+

### Setup

```bash
git clone https://github.com/1mb-dev/nivomoney.git
cd nivomoney
cp .env.example .env
cp docker-compose.override.yml.example docker-compose.override.yml  # publishes ports

# Start all services. On a fresh database they race on migrations:
# repeat this until `docker compose ps` shows them all healthy.
make dev

# Seed database with demo data (reads config from the environment)
set -a; . ./.env; set +a
make seed

# Start frontend (separate terminal)
cd frontend/user-app && npm install && npm run dev
```

Open http://localhost:5173 and login with demo credentials.

### Demo Accounts

| Email | Password | Balance |
|-------|----------|---------|
| raj.kumar@gmail.com | raj123 | ₹50,000 |
| priya.electronics@business.com | priya123 | ₹1,50,000 |

**Admin access:** Run `make seed` locally to generate admin credentials (see `.secrets/credentials.txt`).

All data is synthetic. No real money.

## Architecture

```
services/
├── identity/       # Auth, users, KYC
├── ledger/         # Double-entry accounting
├── wallet/         # Balance management
├── transaction/    # Transfers, payments
├── rbac/           # Roles & permissions
├── risk/           # Fraud detection
├── notification/   # Alerts, messaging
├── simulation/     # Test data generation
└── seed/           # Database seeding

gateway/            # API Gateway with SSE
frontend/
├── user-app/       # Customer React app
└── admin-app/      # Admin dashboard
```

## Tech Stack

| Component | Technology |
|-----------|------------|
| Services | Go 1.26, standard library net/http |
| Database | PostgreSQL 15 |
| Cache | Redis |
| Auth | JWT, bcrypt |
| Frontend | React 19, TypeScript, Vite, TailwindCSS 4 |
| Infrastructure | Docker Compose |
| Monitoring | Prometheus, Grafana |

## Key Patterns

- **Double-entry ledger** - Postgres refuses to post a journal entry that does not balance
- **Idempotent transfers** - The wallet service dedupes internal transfers on transaction ID
- **RBAC** - Granular permissions with role hierarchies
- **Real-time updates** - Server-sent events through the gateway

Transfers do not reach the ledger: the transaction service's ledger client sends no credentials, so the ledger refuses the call. The failure is logged and the transfer still completes.

## Documentation

Full documentation: [nivomoney.com/docs](https://nivomoney.com/docs)

- [Quick Start](https://nivomoney.com/quickstart)
- [Architecture](https://nivomoney.com/architecture)
- [API Flows](https://nivomoney.com/flows)
- [ADRs](https://nivomoney.com/adr)

## Project Scope

| Category | Count |
|----------|-------|
| Services | 8 + gateway |
| API Routes | 111 |
| Frontend Pages | 23 |
| Database Migrations | 8 |

This is a portfolio demonstration, not a production bank. It shows how a neobank *would* be built.

## Contributing

See [CONTRIBUTING.md](.github/CONTRIBUTING.md) for development setup and guidelines.

## License

MIT - see [LICENSE](LICENSE)
