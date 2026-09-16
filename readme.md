# URL Shortener

An **event-driven** URL shortening platform built in Go, split into two independent services that collaborate over a message broker:

- **`url-service`** — public HTTP/REST API: user auth, link CRUD, resolve + caching, and per-user click reporting.
- **`analytics-service`** — background consumer that ingests click events from RabbitMQ and persists them for analytics.

Both services share a single PostgreSQL database and talk to each other **only** through RabbitMQ — the analytics service exposes no HTTP surface.

---

## Table of Contents

- [Architecture](#architecture)
- [Tech Stack](#tech-stack)
- [Services](#services)
- [Event-Driven Flow](#event-driven-flow)
- [Resilience & Degradation](#resilience--degradation)
- [Database Schema](#database-schema)
- [Repository Layout](#repository-layout)
- [Layering & Code Style](#layering--code-style)
- [API Reference](#api-reference)
- [Configuration](#configuration)
- [How to Run](#how-to-run)
- [Local Development & Testing](#local-development--testing)
- [Known Gaps & Roadmap](#known-gaps--roadmap)

---

## Architecture
```text
URL SHORTENER — DATA FLOW

  Client
    │  HTTP :8080
    ▼
  url-service
    ├─ POST /users/register  |  POST /users/login          → JWT auth
    ├─ POST /links  |  GET /links  |  PATCH /links/:code   → PostgreSQL
    ├─ GET /:short_code   (cache: Redis → fallback DB)     → 302 redirect
    │     └── publish "link.clicked"  →  RabbitMQ
    │            exchange: link.events (topic)   queue: link.clicks (durable)
    │                                                │
    │                                                ▼
    └─ GET /reports/overview  →  PostgreSQL  ◀── analytics-service (consumer)
                                                        │
                                                        ▼
                                                  click_events table
```

**Traffic flow at a glance**

1. A user registers/logs in; every protected route requires a JWT (`Authorization: Bearer <token>`).
2. Creating a link generates a short code (base-62 sequence counter + random suffix) and persists it.
3. Resolving `GET /:short_code` checks the Redis cache first (fallback: DB), then 302-redirects and **publishes a `link.clicked` event**.
4. The analytics consumer declares the same topic exchange + queue, ingests the event, and upserts it into `click_events` (idempotent via event id).
5. `GET /reports/overview` returns total/active links and total clicks in **one** aggregated SQL query — the report reads `click_events` directly because both services share the database.

---

## Tech Stack

| Concern                | Choice                                   | Where                                  |
|------------------------|------------------------------------------|----------------------------------------|
| Language               | Go 1.25 (separate modules per service)   | `url-service/go.mod`, `analytics-service/go.mod` |
| HTTP framework         | Gin                                  | `url-service/controller/**`            |
| Auth                   | JWT (HS256) + bcrypt                | `url-service/middleware`, `application/user` |
| Database driver        | pgx v5 (via `database/sql`)         | `*/config/db.go`                       |
| Cache                  | Redis 7 (go-redis v9)               | `url-service/repository/cache`         |
| Message broker         | RabbitMQ 3 (amqp091-go) **topic exchange** | `*/message-queue`                      |
| DB migrations          | Idempotent DDL at startup (no tool)  | `*/database/migrate.go`                |

---

## Services

### `url-service` (port 8080)

Public REST API. Owns `users`, `links`, and the `link_code_seq`; serves the click-report endpoint.

- **Auth:** registration (bcrypt-hashed passwords, unique email) and login (24h JWT).
- **Link management:** create, paginated list, disable (invalidates the cache entry).
- **Resolve:** Redis-first lookup, DB fallback, availability check (active + not expired), then **publish a click event**.
- **Report:** one query aggregating total links, active links, and total clicks per user.

### `analytics-service` (no public port)

Single-responsibility worker that consumes `link.clicked` events and persists them.

- Declares the topic exchange and durable queue idempotently at startup.
- Validates payloads into the `ClickEvent` entity; writes with **idempotent dedupe** (`ON CONFLICT (event_id) DO NOTHING`).
- Graceful shutdown on `SIGINT`/`SIGTERM` via `signal.NotifyContext`.

---

## Event-Driven Flow

|                 | Detail                                              |
|-----------------|-----------------------------------------------------|
| Exchange        | `link.events` (topic, durable)                     |
| Routing key     | `link.clicked`                                      |
| Queue           | `link.clicks` (durable)                            |
| Message model   | `link_clicked_event` DTO — `event_id`, `user_id`, `short_code`, `original_url`, `clicked_at` |
| Concurrency     | Single consumer goroutine at startup               |
| Delivery        | `DeliveryMode: Persistent`, 1s publish timeout     |

**Handler contract (`MessageHandler func(payload []byte) error`)**

- Valid JSON + valid entity → insert (duplicate `event_id` is a no-op) → **ACK**.
- Invalid payload/entity → wrapped as `ErrInvalidPayload` → logged as "dropping invalid click event" → **ACK** (poison message skipped).
- Any other error (DB/repo failure) → propagates → **NACK with requeue**.

The consumer safely exits when the parent context is cancelled; `ErrQueueUnavailable` (no RabbitMQ connection) disables consumption gracefully rather than crashing.

---

## Resilience & Degradation

Everything is environment-driven and **degrades instead of failing fast** when a dependency is unavailable.

- **Database connection** — required. Startup fails if PostgreSQL is unreachable; migrations run before any request is served.
- **Redis** — optional. A ping failure logs `warn: redis unavailable, continuing without cache` and sets the client to `nil`; the resolve use case transparently falls back to the DB.
- **RabbitMQ** — optional with a bounded retry: 5s dial timeout, up to **15 attempts × 3s delay (~45s window)**, then degrades to `nil` (publisher stays usable, publishes return `ErrQueueUnavailable`; the consumer disables itself). There is **no runtime auto-reconnect** — this is a deliberate scope decision (see Roadmap).
- **Startup ordering (Docker Compose)** — both services wait on the `rabbitmq` **healthcheck** (`</dev/tcp` + `rabbitmq-diagnostics ping`, `start_period: 45s`) so the compose-managed boot-race is avoided.
- **Publish failures during resolve** are non-blocking: the redirect still succeeds, the event is dropped (loss is acceptable per design).

---

## Database Schema

```text
DATABASE SCHEMA (PostgreSQL 16 — shared DB "test_api")

users (1) ────── owns ──────▶ (0..N) links ────── counted by ──────▶ click_events
                                  ▲                                      │
                                  │  (no FK — deliberately decoupled)    │

users
  id            serial       PK
  username      varchar(255)
  email         varchar(255) UK
  phonenumber   varchar(20)
  password      varchar(255)   -- bcrypt hash
  created_at    timestamptz    -- DEFAULT NOW()

links
  id            serial       PK
  user_id       integer      FK → users (ON DELETE CASCADE)
  original_url  text
  short_code    varchar(64)  UK   -- base-62 code
  created_at    timestamptz  DEFAULT NOW()
  expires_at    timestamptz  nullable
  is_active     boolean      DEFAULT true

  INDEX idx_links_user_id (user_id)

click_events            -- written by analytics-service consumer
  event_id      varchar(36)  PK        -- idempotent dedupe
  user_id       integer               -- no FK
  short_code    varchar(64)
  original_url  text
  clicked_at    timestamptz
  consumed_at   timestamptz  DEFAULT NOW()

  INDEX idx_click_events_user_id (user_id)
  INDEX idx_click_events_short_code (short_code)

link_code_seq -- sequence driving the base-62 short-code counter
```

Additional objects:

- `link_code_seq` — sequence backing the base-62 short-code counter (`idx_links_user_id` on `links`).
- `click_events` — **no FK** (deliberately decoupled from transactional core): `event_id VARCHAR(36) PRIMARY KEY`, `user_id`, `short_code`, `original_url`, `clicked_at`, `consumed_at`, with indexes `idx_click_events_user_id` and `idx_click_events_short_code`.

| Metric                                    | Query shape (per user)                                                      |
|-------------------------------------------|-----------------------------------------------------------------------------|
| Total links                               | `COUNT(*)` over `links` filtered by `user_id`                               |
| Active links                              | `COUNT(*)` where `is_active = true AND (expires_at IS NULL OR expires_at > NOW())` |
| Total clicks                              | `COUNT(*)` over `click_events` filtered by `user_id`                        |

All three aggregates are computed in **a single `SELECT`** with scalar subqueries — one round trip, one consistent snapshot.

---

## Repository Layout

```
.
├── deployment/
│   └── docker-compose.yml        # Postgres + Redis + RabbitMQ + both services
├── url-service/                  # module: url-shortener
│   ├── main.go                   # wiring, routes, graceful shutdown
│   ├── Dockerfile
│   ├── application/              # use cases (business rules, no I/O deps exposed)
│   │   ├── link/                 # create / resolve / list / disable + click-event publish
│   │   ├── report/               # per-user overview aggregation (+ dto/)
│   │   └── user/                 # register / login (+ dto/)
│   ├── config/                   # env-driven DB / Redis / RabbitMQ / JWT / cache TTL
│   ├── container/                # manual DI: construct repositories→usecases→controllers
│   ├── controller/               # Gin handlers (request binding, status codes)
│   ├── database/                 # idempotent migrations
│   ├── doc/                      # design notes
│   ├── entities/                 # JSON-free domain objects (unexported fields)
│   ├── message-queue/            # RabbitMQ publisher + interface
│   ├── middleware/               # JWT auth middleware
│   ├── repository/               # cache/ (Redis), link/, report/, user/ (Postgres)
│   └── service/                  # shortcode-generator (base-62 + CSPRNG suffix)
└── analytics-service/            # module: analytics-service
    ├── main.go                   # consumer runner + handler wiring + graceful shutdown
    ├── Dockerfile
    ├── application/click/        # ConsumeClickEvent use case (+ dto/)
    ├── config/                   # DB + RabbitMQ (same retry policy)
    ├── container/                # manual DI
    ├── database/                 # click_events migration
    ├── entities/                 # ClickEvent domain object
    ├── message-queue/            # RabbitMQ consumer + interfaces & constants
    └── repository/click-event/   # idempotent Postgres writer
```

---

## Layering & Code Style

The codebase follows a strict **onion-ish / clean-architecture-lite** layering. Dependencies point downward, never across siblings.

**Dependency rule**

```
Controller → Use Case → (Repo Interface) → Repository Impl → SQL/Redis
                          ↘  Service / Message Queue
```

Key conventions a reviewer will notice (and should preserve):

- **Entities are JSON-free.** They expose `unexported` fields with getters and validating constructors (e.g. `entities.NewLink`, `NewClickEvent`). No `json`/`db` tags on the domain — every violation is rejected by design.
- **DTOs and entities are separated.** Requests bind into inline anonymous structs at the controller; responses and cross-service messages use DTO packages (`application/*/dto`).
- **Interfaces live at the consuming side.** `application/link.ShortCodeGenerator`, `messagequeue.Rabbitmq` / `messagequeue.Consumer`, `repository/*` interfaces — each dependency is defined where it is used, which keeps tests cheap (hand-rolled mocks, no mock frameworks).
- **Manual DI in `container/`.** Each service composes its own object graph (`*Container` per bounded context + a root `Container`), making construction explicit and unit-testing trivial.
- **Repository implementations** return domain entities directly, use `RETURNING` clauses, handle `sql.ErrNoRows` explicitly, and are runtime-checked with `var _ Interface = (*Impl)(nil)`.
- **Zero-comment policy.** Code is self-documenting; doc-level intent lives in `url-service/doc/`.
- **Module discipline.** The two services are **independent Go modules** (no cross-module `replace`/imports); sharing only happens at the messaging/SQL level.

---

## API Reference

Base URL: `http://localhost:8080`

| Method | Path                     | Auth        | Description                                   | Success |
|--------|--------------------------|-------------|-----------------------------------------------|---------|
| POST   | `/users/register`        | —           | Register a user (`username`, `email`, `phonenumber`, `password`) | `201` |
| POST   | `/users/login`           | —           | Login, returns JWT                            | `200`  |
| GET    | `/:short_code`           | —           | Resolve link → `302 Location: original_url`   | `302`  |
| POST   | `/links`                 | Bearer      | Create short link (`original_url`)            | `201`  |
| GET    | `/links`                 | Bearer      | List links, `?offset=&limit=` (defaults `0,15`) | `200` |
| PATCH  | `/links/:short_code`     | Bearer      | Disable a link                                | `200`  |
| GET    | `/reports/overview`      | Bearer      | `{total_links, active_links, total_clicks}`   | `200`  |
| GET    | `/ping`                  | Bearer      | Auth probe (`user_id` echo)                   | `200`  |

**Sample contract — report overview**

```http
GET /reports/overview
Authorization: Bearer <jwt>

200
{"total_links": 3, "active_links": 2, "total_clicks": 10}
```

`user_id` is always sourced from the verified JWT claim (set by `middleware/auth.go`), never from the body or query — so reports are scoped to the authenticated caller.

---

## Configuration

All settings come from environment variables, with sane local defaults.

| Service          | Variable        | Default             | Notes                          |
|------------------|-----------------|---------------------|--------------------------------|
| both             | `JWT_SECRET`    | `your-secret-key`   | **must change in production**  |
| url-service      | `PGUSER/PGPASSWORD/PGHOST/PGPORT/PGDATABASE` | `postgres/postgres/localhost/5432/test_api` | pgx DSN |
| url-service      | `REDISHOST/REDISPORT/REDISPASSWORD` | `localhost/6379/redis` | — |
| both             | `RABBITMQUSER/RABBITMQPASSWORD/RABBITMQHOST/RABBITMQPORT/RABBITMQVHOST` | `guest/guest/localhost/5672/` | retry 15×3s |
| url-service      | `CACHETTL`      | `3600` (seconds)    | resolve-cache TTL              |

---

## How to Run

### Docker Compose (recommended)

```bash
cd deployment
docker compose up -d --build
```

Containers orchestrated:

| Container                    | Image              | Port(s)                              |
|------------------------------|--------------------|--------------------------------------|
| `postgres` (url-shortener-db)| `postgres:16`      | `5432`                               |
| `redis`                      | `redis:7-alpine`   | `6379`                               |
| `rabbitmq`                   | `rabbitmq:3-management` | `5672` (AMQP), `15672` (Management UI) |
| `url-service`                | built from `url-service/Dockerfile` | `8080`               |
| `analytics-service`          | built from `analytics-service/Dockerfile` | — (no exposed port) |

Compose creds: Postgres `postgres/postgres` (DB `test_api`), Redis password `redis`, RabbitMQ `rabbitmq/rabbitmq` (management UI at `http://localhost:15672`).

> **Why the RabbitMQ healthcheck matters:** both services bind to RabbitMQ at startup. If `depends_on` only waited for container start, they could dial before the broker exposes its AMQP listener, hit the ~15-attempt retry window, and permanently degrade. The healthcheck (`</dev/tcp` probe + `rabbitmq-diagnostics ping`, `start_period: 45s`) guarantees a ready broker before either service resolves.

Services are built as **static binaries in scratch-like `alpine:3.21` images** via multi-stage Dockerfiles (`CGO_ENABLED=0`).

### Smoke test

```bash
export BASE=http://localhost:8080

curl -s -X POST $BASE/users/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"demo","email":"demo@example.com","phonenumber":"09123456789","password":"secret123"}'

TOKEN=$(curl -s -X POST $BASE/users/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"secret123"}' | jq -r .token)

CODE=$(curl -s -X POST $BASE/links -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"original_url":"https://example.com"}' | jq -r .short_code)

curl -sI $BASE/$CODE          # expect: 302 Location: https://example.com
curl -s  $BASE/reports/overview -H "Authorization: Bearer $TOKEN"
# {"total_links":1,"active_links":1,"total_clicks":1}  (once the consumer persists the event)
```

---

## Local Development & Testing

Both modules target Go 1.25 locally; commands are run from each service directory.

```bash
cd url-service        # or analytics-service
GOTOOLCHAIN=go1.25.0 go build ./...   # vet
GOTOOLCHAIN=go1.25.0 go vet ./...
GOTOOLCHAIN=go1.25.0 go test ./...
```

- Use cases are unit-tested with **hand-rolled mocks** (no mockgen/stretchr): e.g. `mock_link_repo_test.go`, `mock_link_cache_test.go`, `mock_click_repo_test.go`, `mock_report_repo_test.go`.
- Covered today: create/resolve/list/disable link use cases, click-event consuming (valid/invalid/duplicate/repo-error), report overview, and the RabbitMQ publisher (nil-connection → `ErrQueueUnavailable`).

> `GOTOOLCHAIN=go1.25.0` pins against a cached toolchain when the default Go on the machine is older.

---

## Known Gaps & Roadmap

Deliberate scope decisions and the next things worth building:

1. **Observability** — no `/metrics`, structured logs, or tracing yet. Priorities: resolve hit-rate / cache-miss / degraded counters, consumer lag (`RabbitMQ management API`), and a small Prometheus endpoint.
2. **RabbitMQ resilience** — startup retry only; no runtime auto-reconnect or publisher-confirms. Add connection monitoring + channel recovery for long-lived deploys.
3. **Secrets** — `JWT_SECRET` is a placeholder. Move to secrets injection (docker secrets / vault).
4. **Migrations** — current DDL is app-start idempotent; adopt a versioned migration tool (golang-migrate) for many environments.
5. **Dedupe persistence** — event id dedupe relies on the PK; would prefer an outbox pattern (transactional publish) for exactly-once windowed guarantees.
6. **Report pagination/rollups** — `reports/overview` is point-in-time; hourly/daily rollup tables become useful at volume.

---

## License

This project is licensed under the [MIT License](LICENSE).

Copyright (c) 2026 Abolfazl Ganjtabesh

Contributions are welcome — see the repository's contributing flow. By submitting a pull request, you agree that your contributions are licensed under the same MIT terms.
