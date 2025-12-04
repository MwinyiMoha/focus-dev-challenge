# Dev Challenge

This repository contains a small campaign sending service (HTTP API + background worker) backed by Postgres and Redis. The project includes a simple template personalization engine and an asynq-based task queue.

## How to run

Prerequisites: Docker & Docker Compose, Go toolchain (for running tests locally).

1. Start the development stack (Postgres + Redis + web + worker):

```bash
docker compose up --build
```

Notes:
- Postgres is exposed on host port `15432` (container port `5432`) to avoid conflicts with a local DB.
- Redis is exposed on host port `16379` (container port `6379`).
- The `docker-compose.yaml` mounts migration and seed SQL files from `schema/migrations` and `schema/scripts` into Postgres' `/docker-entrypoint-initdb.d/` so the DB schema and seeds are applied on first startup.

2. Run the web service or worker locally without Docker (useful for development):

- Rename `config.env.sample` to `config.env`. The environment variables will control the behavior of the application

```bash
# Run web server
APP_TIER=web go run ./cmd

# Run worker
APP_TIER=worker go run ./cmd

# With Live Reload
air
```

3. Run tests:

```bash
make test

# or

go test ./... -v
```

## Assumptions

- The project uses Postgres and Redis for persistence and queuing respectively.
- Seed scripts are used to initialize example customers and campaigns for local development and testing.
- The delivery adapters for SMS/WhatsApp are not implemented; the worker currently logs send attempts (see "Mock sender behavior").
- The service expects templates to reference struct field names directly (e.g., `{FirstName}`).

## Template Handling 

- Template syntax: `{FieldName}`. The `renderTemplate` function reflects the provided data struct and replaces placeholders with the corresponding field values.
- Behavior for missing/null fields:
	- If a struct field is not present, the placeholder is left unchanged in the rendered template (e.g., `{NonExistent}` remains `{NonExistent}`).
	- If a field is a nil pointer, it renders as an empty string.
	- For `pgtype.Text` fields (from `sqlc`), the implementation reads the `String` field (so when `Valid` is false the `String` may be empty).
	- `pgtype.Timestamp` fields are read via their `Time` field and formatted as RFC3339.

This approach makes templates resilient: missing data doesn't break rendering, and unresolved placeholders remain visible for inspection.

## Mock Sender

- The worker handler `SendMessage` currently acts as a no-op/mock: it parses the task payload and logs the `message_id` that would be sent. The actual channel adapters (SMS/WhatsApp) are left as TODOs.
- This mock behavior is intentional considering time constraints: it allows triggering the background task, DB writes, and task scheduling without integrating third-party providers e.g SMS|WhatsApp Gateways.
- Further implementation details in system overview document under section: **Worker Processing & Retry Logic**

## Queue choice

- The project uses Redis (via the `asynq` library) for task queuing.

Why Redis/asynq?
- Lightweight and simple to run locally with Docker (no external broker required).
- `asynq` offers solid Go-native APIs, task scheduling and retry/backoff support out of the box.
- Good developer experience: easy to inspect tasks, retry behavior and to integrate into Go services.
- Reasonable message throughput for low to medium traffic applications

For high throughput applications, a broker with advanced routing guarantees is more desirable. Swapping to RabbitMQ or any other broker is straightforward by implementing a queue interface and swapping the task producer/consumer.

## Where to look next

- `internal/core/app/service.go` — business logic, template rendering and enqueueing tasks.
- `internal/adapters/worker/handlers.go` — background task handler (current mock sender).
- `schema/migrations` and `schema/scripts` — database schema and seed data.
- `docker-compose.yaml` — development stack configuration (note host port remapping).

## Tools Used

- **Golang Migrate:** For managing database schema changes
- **SQLC:** For generating Go queries from SQL
- **Generative AI:** Code autocompletion, performance tricks and writing tests
