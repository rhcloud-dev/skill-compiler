---
# Required: unique name for your skill (used in output filenames and metadata)
name: acme-api

# Spec source(s) — where sc reads your interface definition from.
# Can be a simple string path, an object, or an array of sources.
#
# Simple path:
#   spec: ./openapi.yaml
#
# Object with options:
#   spec:
#     path: ./openapi.yaml
#     type: openapi
#
# Multiple sources:
spec:
  - path: ./openapi.yaml
    type: openapi
  - binary: acme
    type: cli
    help-flag: --help
    max-depth: 2
    exclude:
      - internal

# Output directory (default: ./sc-out/)
out: ./sc-out/

# Per-artifact toggles — disable artifacts you don't need
# artifacts:
#   scripts:
#     enabled: false
#   llms-api-txt:
#     filename: api-reference.txt

# Skill metadata — ends up in the SKILL.md frontmatter
skill:
  license: MIT
  compatibility: claude
  env:
    - ACME_API_KEY
    - ACME_BASE_URL
  allowed-tools: Bash(curl *)
  metadata:
    version: "1.0"
    author: your-org

# LLM provider overrides (optional — can also use CLI flags, env vars, or ~/.config/sc/config.yaml)
# provider:
#   provider: anthropic
#   model: claude-sonnet-4-20250514
---

# Product

Acme API is a REST service for managing widgets and orders. It's used by
internal teams and partners to create, query, and fulfill widget orders.

Key capabilities:
- Widget CRUD (create, read, update, delete)
- Order lifecycle management (draft -> confirmed -> shipped -> delivered)
- Inventory tracking and alerts
- Webhook subscriptions for order status changes

Target users: backend services and AI agents that automate order fulfillment.

# Workflows

## Create and fulfill an order

1. List available widgets: `GET /widgets?in_stock=true`
2. Create a draft order: `POST /orders` with widget IDs and quantities
3. Confirm the order: `POST /orders/{id}/confirm`
4. Poll for fulfillment: `GET /orders/{id}` until status is `shipped`

## Set up webhook notifications

1. Register a webhook: `POST /webhooks` with target URL and event types
2. Verify the webhook with the test ping: `POST /webhooks/{id}/test`
3. Handle incoming `order.status_changed` events

## Inventory check before bulk ordering

1. `GET /widgets?ids=W1,W2,W3` to check current stock levels
2. For each widget with `stock < desired_qty`, skip or reduce quantity
3. `POST /orders` with adjusted quantities
4. Log any widgets that were out of stock

# Guardrails

- Never delete a widget that has open orders — the API returns 409 Conflict
- Rate limit: 100 requests/minute per API key. Implement backoff on 429 responses
- Order confirmation is irreversible — always verify line items before confirming
- Webhook URLs must be HTTPS in production
- Do not store `ACME_API_KEY` in code or logs

# Conventions

- All timestamps are ISO 8601 in UTC (e.g., `2026-01-15T09:30:00Z`)
- Pagination: `?page=1&per_page=50` (max 100 per page)
- Filtering: query params use snake_case (e.g., `?created_after=2026-01-01`)
- IDs are prefixed strings: `wgt_abc123` (widgets), `ord_xyz789` (orders)
- Error responses follow `{ "error": { "code": "not_found", "message": "..." } }`

# Examples

## Quick health check

```
GET /health
Authorization: Bearer $ACME_API_KEY

200 OK
{ "status": "ok", "version": "2.4.1" }
```

## Create a widget

```
POST /widgets
Authorization: Bearer $ACME_API_KEY
Content-Type: application/json

{
  "name": "Turbo Sprocket",
  "sku": "TS-500",
  "price_cents": 2499,
  "stock": 150
}

201 Created
{
  "id": "wgt_abc123",
  "name": "Turbo Sprocket",
  "sku": "TS-500",
  "price_cents": 2499,
  "stock": 150,
  "created_at": "2026-01-15T09:30:00Z"
}
```

# Common patterns

- Use `If-None-Match` / `ETag` headers to avoid redundant fetches on widget listings
- Batch widget lookups with `?ids=W1,W2,W3` instead of individual GET requests
- Subscribe to webhooks instead of polling for order status changes
- Always pass `Idempotency-Key` header on POST requests to prevent duplicate orders
