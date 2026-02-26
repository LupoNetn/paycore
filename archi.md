## Paycore — Architecture review, MVP roadmap, and advanced roadmap

Project: simulated fintech backend (wallets, transfers, double-entry ledger).

This document is a manager-level, actionable roadmap to get the project to a secure
and testable MVP, plus a prioritized list of production/advanced improvements.

**Goals:**
- Deliver a safe, atomic transfer API that preserves double-entry ledger invariants.
- Provide deterministic behavior under concurrency (idempotency, no double-spend).
- Ship with tests, observability, safety guards and an operational playbook.

---

**MVP success criteria (concrete)**
- API: `POST /transfer` that performs a wallet-to-wallet transfer and returns a transaction.
- Atomic: the transfer creates a transaction record, two ledger rows (debit/credit), and
  updates both wallet balances inside a single DB transaction or returns the existing
  transaction when the same idempotency key is used.
- Safety: no double-spend under concurrent requests; idempotency is enforced.
- Tests: unit + integration tests that prove correctness (including concurrent cases).

---

**MVP feature checklist (ordered by priority)**

- **Core transfer flow (P0)**: implement end-to-end in `internal/transfer/service.go`.
  - **Validate input**: positive amount, non-empty idempotency key, different wallets, matching currency.
  - **Lightweight DTO in handler**: handler accepts simple JSON types (string/decimal), validate, then map to `pgtype`.
  - **Start TX + safe rollback**: begin TX and immediately `defer` a rollback that ignores ErrTxClosed; commit at end.
  - **Deterministic locking**: SELECT wallets FOR UPDATE in deterministic order (min(id), max(id)).
  - **Balance checks**: use `shopspring/decimal` and `pgtype.Numeric` conversions; return `ErrInsufficientFunds` on shortfall.
  - **Create transaction row**: status `pending`, store `idempotency_key`.
  - **Create ledger rows**: debit (sender) and credit (receiver) with `balance_before` / `balance_after`.
  - **Update wallet balances**: atomic updates inside same TX.
  - **Finalize transaction**: update status `completed` and commit.

- **Idempotency (P0)**
  - Add DB unique constraint on `transactions(idempotency_key)` (migration added).
  - Pattern: attempt insert → if unique_violation (23505) then fetch transaction by idempotency key and return it.
  - Do not perform a pre-check outside TX to avoid race conditions.

- **Domain errors and mapping (P0)**
  - Use typed errors in service: `ErrInsufficientFunds`, `ErrSameWallet`, `ErrCurrencyMismatch`, `ErrInvalidAmount`.
  - Handler maps these to appropriate HTTP codes: 400/409/422 instead of 500.

- **Testing (P0/P1)**
  - Unit tests for service logic with mocked `db.Queries`.
  - Integration test using ephemeral Postgres (testcontainers) exercising full flow.
  - Concurrency test: spawn many goroutines issuing the same idempotency key and different keys to assert no double ledger entries.

---

**Data model & DB constraints**

- **Transactions table**:
  - `idempotency_key VARCHAR NOT NULL UNIQUE`
  - `status` as enum (`pending`, `completed`, `failed`)
  - indexes: `created_at`, `sender_wallet_id`, `receiver_wallet_id` for queries.

- **Wallets**:
  - `balance NUMERIC(18,2) NOT NULL` with a DB CHECK to prevent negative balances where applicable.
  - Foreign keys to users; add index on `user_id`.

- **Ledger**:
  - Immutable rows once inserted.
  - Use `created_at` timestamp and consider partitioning for very large volumes.

---

**API surface (recommended for MVP)**

- `POST /transfer` — create transfer (idempotent via header/body idempotency_key).
- `GET /wallets/:id` — return wallet with current balance.
- `GET /wallets/:id/transactions` — paginated list using ledger or transactions read-side.
- `GET /transactions/:id` — transaction detail.

Notes:
- Keep handler DTOs simple (strings/primitive types). Convert to `pgtype` only in service layer.
- Use headers for `Idempotency-Key` or accept in request body — document chosen pattern.

---

**Operational & safety features (must for production-grade)**

- **Authentication & Authorization**
  - Enforce that the authenticated user owns the `sender_wallet_id`.
  - RBAC for administrative endpoints.

- **Rate limiting and per-user quotas**
  - Prevent abuse / replay by limiting transfers per minute/day per user.

- **Context timeouts**
  - Set request-level timeouts (e.g., 5s) and shorter DB context timeouts for TXs.

- **Observability**
  - Structured logs (`slog`) with request id / trace id / idempotency key / tx id.
  - Tracing (OpenTelemetry): trace the lifecycle of a transfer across DB calls, tasks, notifications.
  - Metrics (Prometheus): request latency, tx success/failure counts, concurrent txs, DB errors, idempotency conflicts.

- **Health checks & graceful shutdown**
  - `/healthz` readiness/liveness endpoints; close DB pool on shutdown.

- **Migrations & backups**
  - Use goose or another migration tool; ensure migrations are in CI.
  - Regular DB backups and tested restore playbook.

- **Secrets management**
  - Use env vars + Vault/Secret Manager for DB credentials and API keys.

- **Deployment**
  - Containerize (Docker), provide manifest (Kubernetes) with resources, liveness/readiness probes.
  - CI pipeline: `go vet`, `golangci-lint`, `go test`, `sqlc generate`, run migrations and integration tests before merge.

---

**Resilience & concurrency strategies**

- **Retry strategy**
  - Retry on serialization failures (Postgres 40001) with exponential backoff and jitter — keep retry attempts small (≤3).

- **Lock ordering**
  - Always lock wallets in deterministic order (by UUID bytes) to reduce deadlocks.

- **Idempotency beyond DB**
  - For external side-effects (webhooks, notifications), use an outbox table or task queue that is written inside the same TX and processed after commit.

---

**Testing & QA (detailed)**

- **Unit tests**: service methods with mocked `db.Queries` — assert ledger creation and balance updates.
- **Integration tests**: start Postgres, run migrations, exercise full APIs.
- **Concurrency stress tests**: run 1000+ concurrent transfers in test environment to assert invariants.
- **Load tests**: use k6 or Gatling to measure throughput/latency and find bottlenecks.
- **Property tests**: fuzz amounts, ids and check invariants (sum of balances + ledgers consistency).

---

**Advanced roadmap (post-MVP)**

- **Settlement & rails**: integrate with external payment rails, handle settlement delays and reconciliation.
- **Multi-currency and FX**: atomic conversion, FX table, rounding rules, fees.
- **Event sourcing / CQRS**: separate write model (ledger) and read model for high-performance queries.
- **Archival & partitioning**: archive old ledger rows, partition by date for scale.
- **Disaster recovery**: automated backups, warm standbys, failover automation.
- **Compliance**: audit trails, immutable logs, GDPR data handling, KYC/AML hooks.

---



## Remaining MVP Tasks

- [ ] Handler input validation (DTOs, positive amounts, currency, idempotency key)
- [ ] Proper error mapping (typed errors → HTTP codes)
- [ ] Unit tests for transfer, wallet, ledger flows
- [ ] Integration tests (end-to-end, DB migrations)
- [ ] Concurrency test (no double-spend, idempotency)
- [ ] Observability: logs, metrics, traces
- [ ] Health check endpoint (`/healthz`)
- [ ] Rate limiting and quotas (optional for MVP)
- [ ] API documentation and operational playbook

---

## Advanced Roadmap (post-MVP)

- Settlement rails integration
- Multi-currency and FX
- Event sourcing / CQRS
- Archival & partitioning
- Disaster recovery
- Compliance (audit, GDPR, KYC/AML)

---

**Next steps:**
- Implement handler validation and error mapping
- Add tests (unit, integration, concurrency)
- Add observability and health check
- Document API and ops

Ping me to start any of these tasks and I'll add a targeted todo list and implement the first change.

---

## Advanced Roadmap (post-MVP)

- Settlement rails integration
- Multi-currency and FX
- Event sourcing / CQRS
- Archival & partitioning
- Disaster recovery
- Compliance (audit, GDPR, KYC/AML)

---

**Next steps:**
- Implement handler validation and error mapping
- Add tests (unit, integration, concurrency)
- Add observability and health check
- Document API and ops

Ping me to start any of these tasks and I'll add a targeted todo list and implement the first change.
