**System Overview**

This document summarizes the data model, request flow for sending a campaign, how the background queue worker processes messages, pagination strategy, and the personalization/template approach used by the service.

**Architecture**
This project employs the Hexagonal Architecture (Ports & Adapters) to create a loosely coupled system that enhances modularity, testability and flexibility. This project is organized into:

- Domain: includes entities, payloads and base types
- App/Service Layer: houses the business logic i.e data validation, feature workflows
- Ports: houses the rules of engagement by use of Go interfaces. 
- Adapters: houses the driving adapters (HTTP Controller) and driven adapters (Postgres Integration)

**Data Model & Relationships**

- Customers
	- Table: `customers`
	- Columns: `id` (PK, BIGSERIAL), `phone` (VARCHAR(32), indexed), `first_name` (VARCHAR), `last_name`, `location`, `preferred_product`, `created_at`, `updated_at`
	- Indexes: `idx_customers_phone` on `phone` (just in case)

- Campaigns
	- Table: `campaigns`
	- Columns: `id` (PK, BIGSERIAL), `name`, `channel` (ENUM-like via CHECK: 'sms'|'whatsapp'), `status` (CHECK: 'draft'|'scheduled'|'sending'|'sent'|'failed'), `base_template` (TEXT), `scheduled_at` (TIMESTAMP nullable), `created_at`, `updated_at`
	- Indexes: `idx_campaigns_channel`, `idx_campaigns_status`

- OutboundMessages
	- Table: `outbound_messages`
	- Columns: `id` (PK), `campaign_id` (FK -> campaigns.id), `customer_id` (FK -> customers.id), `status` ('pending'|'sent'|'failed'), `rendered_content` (TEXT), `last_error` (TEXT), `retry_count` (int, default 0), `created_at`, `updated_at`
	- Indexes: `idx_outbound_messages_campaign_id`, `idx_outbound_messages_customer_id`, `idx_outbound_messages_status`

Relationships:
- `campaigns` 1 — * `outbound_messages` (cascade delete)
- `customers` 1 — * `outbound_messages` (cascade delete)

**Request flow: POST /campaigns/{id}/send**
- Client calls `POST /campaigns/{id}/send` with a payload containing `customers` (list of customer IDs).
- API validation is performed using `go-playground/validator` to ensure list of customer IDs is not empty.
- For each target customer the service:
	1. Retrieves the campaign and customer records from the repository (Postgres).
	2. Calls `renderTemplate(campaign.BaseTemplate, customer)` to produce `rendered_content`.
	3. Inserts a record into `outbound_messages` with status `pending` and the rendered content inside a transaction.
	4. Enqueues an `asynq` task (`SendMessageTask`) that contains the `message_id` and is routed to the worker queue. If the campaign is scheduled, the enqueue uses `ProcessAt` to schedule execution.

Concurrency & rate control:
- The HTTP handler uses a bounded worker pool (errgroup + semaphore) to limit concurrent DB/worker operations to a configured parallelism (10 by default).

Response semantics:
- The POST returns an immediate response indicating messages queued count and campaign status. Enqueued tasks asynchronously drive delivery.

**Worker Processing & Retry Logic**
Worker: an asynq worker subscribes to the queue and handles `SendMessageTask` tasks. Current implementation only receives the tasks and logs the intention to send a message. The concrete implementation would look as follows: 
- For each task:
	1. Load the `outbound_messages` record by `message_id` from task payload and attempt delivery via the appropriate channel adapter (SMS/WhatsApp).
	2. On success: update `outbound_messages.status = 'sent'` and `updated_at`.
	3. On failure: increment `retry_count`, set `last_error`, and set status to `'failed'` only after exceeding a retry threshold; otherwise re-enqueue (asynq provides retry/backoff controls).

Retry policy:
- Use a configurable retry limit and backoff (leveraging asynq's retry/backoff configuration). Each failure increments the message `retry_count`.
- After the retry limit, mark the message `failed` and optionally an alert can be emitted.

Idempotency and duplicate protection:
- Use message IDs (primary key of `outbound_messages`) as the canonical identifier for the delivery attempt. Perform database updates in transactions to ensure idempotent state transitions (e.g., check existing status before updating to `sent`).

**Pagination Strategy**
The project uses the following pagination strategy:
- Campaigns listing endpoints use `pageNumber` and `pageSize` with server-side limits to prevent large responses (default page size 10, max 100).
- To avoid duplicates/missing records during pagination, we use stable ordering using campaign ID `id DESC` when fetching pages so that results don't shift between requests.

Future enhancement:

I would prefer cursor-based pagination for high-throughput lists. With that said, current offset-based approach is acceptable for moderate datasets and simpler to implement.

**Personalization & Template System**
- Template format: simple placeholder syntax using `{FieldName}`. The `renderTemplate` function accepts a template string and a struct (or pointer to struct) as data.
- How it works:
	1. The function reflects the provided data value and matches placeholders with the struct field names using a regex (`\\{([^}]+)\\}`).
	2. It handles pointer fields, `pgtype` fields (e.g., `pgtype.Text` by reading the `String` field, `pgtype.Timestamp` by reading `Time` and formatting as RFC3339), and values that implement `fmt.Stringer`.
	3. When a field is missing, the placeholder is left unchanged (fallback) so templates are resilient to missing data.

Future enhancements:
- Richer template language: swap the simple placeholder engine for a templating engine (e.g., Go `text/template` or `sprig`) to support conditionals, loops and formatting.
- AI-driven content: add a personalization pipeline step that can call an external model (or local model) to generate or augment message text before persisting `rendered_content`.
- Localization: support locale-aware templates and date/number formatting.
- Safe fallback & default values: provide a way to include default values e.g `{FirstName|Guest}` to avoid annoying whitespaces in `rendered_content`.

---
