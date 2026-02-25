# Paycore — Work remaining & roadmap (manager summary)

Project: simulated fintech backend (wallets, transfers, double-entry ledger).  
Current focal file inspected: internal/transfer/service.go — transfer flow scaffolded but incomplete and has correctness issues.

---

## High-level objective (MVP)
- Atomic transfers between wallets with:
  - row-level locking (no double spend)
  - idempotency
  - double-entry ledger (immutable debit/credit)
  - cached wallet balances kept consistent with ledger
  - full DB transaction (commit/rollback)
  - tests and deadlock handling

---

## Immediate correctness issues (found / likely)
- Unreachable / incorrect error handling around idempotency lookup:
  - returning nil error on tx begin failure and other wrong returns.
  - generic err checks instead of distinguishing "not found" vs real error.
- Missing tx.Rollback on error paths and missing tx.Commit on success.
- No creation of transaction record (pending/completed), no ledger entries.
- No wallet balance updates; balance comparisons may be using wrong types (decimal handling).
- No prevention of self-transfers, negative amounts or zero-amount transfers.
- No unique constraint handling for idempotency key (race conditions possible).
- No explicit error return when insufficient balance — function continues.
- No proper deadlock/lock-order safeguards beyond ordering by ID but not handling equal IDs or same wallet.
- No tests (unit/integration) for transfer correctness or concurrency.

---

## What needs to be implemented for MVP (priority order)

High
- Implement full CreateTransaction flow in internal/transfer/service.go:
  - validate request (positive amount, different wallets, currency)
  - idempotency lookup: return cached transaction if found; if not found continue
  - begin DB tx, ensure defer tx.Rollback(ctx) unless committed
  - SELECT ... FOR UPDATE both wallets in deterministic order (by wallet ID)
  - check sender balance using decimal.Compare and prevent negatives
  - create transaction row with status = pending
  - insert two ledger rows (debit for sender, credit for receiver) with balance_before/after
  - update wallet balances atomically
  - update transaction status = completed
  - commit tx
- Ensure idempotency key unique constraint in DB and handle conflicts gracefully.

Medium
- Add integration tests that run against a test Postgres instance using real migrations.
- Add unit tests for service logic (mock queries).
- Implement retry/backoff for serialization failures / deadlocks (exponential backoff).

Low
- Currency validation and multi-currency routing (MVP can restrict to single currency).
- Webhook/event emitters after successful transaction.

---

## Safety & data integrity checklist
- Use accurate decimal arithmetic (shopspring/decimal everywhere).
- Use DB-level constraint to prevent negative wallet balances.
- Use unique index on transactions(idempotency_key).
- Always use tx.Rollback on early returns.
- Log and handle transient DB errors (retry on serialization errors).
- Enforce immutable ledger rows (application + DB policy).

---

## Tests to add
- Concurrent transfers between same wallets to validate no double-spend.
- Idempotency test: same idempotency_key twice returns same transaction, no double ledger.
- Insufficient funds test.
- Self-transfer rejection test.
- Integration end-to-end: create wallets -> transfer -> assert balances and ledgers.

---

## What you should learn / deepen
- SQL transactions, isolation levels, and SELECT FOR UPDATE semantics.
- Deadlock avoidance and retry strategies (Postgres serialization errors).
- Double-entry accounting models (balance_before/after).
- Proper use of decimal arithmetic in Go (shopspring/decimal).
- Patterns for idempotency in distributed systems.
- Writing integration tests with ephemeral Postgres (testcontainers or docker).

---

## Hard challenges (advanced tasks)
- Implement a safe compensating/reversal mechanism for failed external side-effects.
- Add multi-currency transfers with FX rates and atomic conversions (ledger correctness).
- Implement an event-sourcing or CQRS read-side for transaction history with eventual consistency.
- Implement high-throughput benchmark with concurrent transfers and measure throughput/latency.

---

## Concrete next steps (short checklist)
- [ ] Fix error returns in CreateTransaction (return err, not nil).
- [ ] Add defer tx.Rollback(ctx) immediately after tx begin.
- [ ] Distinguish "not found" error when checking idempotency.
- [ ] Implement transfer pipeline (create pending tx -> ledger entries -> update balances -> commit).
- [ ] Add DB unique index on idempotency_key and handle conflicts.
- [ ] Add unit tests and 1 integration test for a successful transfer.
- [ ] Add concurrency test to prove no double spending.

---

## Recommended priority for the next 2 weeks
1. Implement correct CreateTransaction flow + DB transaction handling (one PR).
2. Add idempotency DB constraint + tests.
3. Add integration tests for concurrency and insufficient funds.

---

Maintain commit-sized tasks, add tests for each change, and run integration tests in CI before merging.
