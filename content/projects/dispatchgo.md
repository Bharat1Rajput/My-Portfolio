---
title: "DispatchGo"
---

## Problem

Webhooks look simple but are fundamentally a **reliability problem in distributed systems**.

A naive implementation — sending an HTTP POST after a database write — breaks in real-world scenarios:
- The receiver is down or slow
- The receiver returns `200 OK` but silently drops the payload
- The sender crashes between persisting the job and making the HTTP call
- Retries create duplicate deliveries without guarantees

These issues lead to **lost events, duplicate side effects, and no auditability**.

DispatchGo solves this by treating webhook delivery as a **durable, stateful job lifecycle**, not a fire-and-forget side effect.

---

## Architecture

DispatchGo is a single Go binary with three internal subsystems:

- **Ingestion API** — a chi-based HTTP server that accepts webhook job submissions, validates them, and persists them in PostgreSQL with status `pending`.
- **Dispatcher** — a fixed-size worker pool that continuously polls for `pending` jobs, acquires row-level locks using `SELECT ... FOR UPDATE SKIP LOCKED`, performs delivery, and updates status to `delivered`, `failed`, or `exhausted`.
- **Retry Scheduler** — a background goroutine that periodically scans for retryable jobs and re-queues them based on retry policy and backoff strategy.

`SKIP LOCKED` enables safe concurrent processing without a message broker. Multiple workers can claim jobs in parallel while PostgreSQL guarantees that each job is processed by exactly one worker at a time.

> This design provides **at-least-once delivery guarantees with minimal infrastructure**.

---

## Design Decisions

**Why PostgreSQL as the queue instead of RabbitMQ?**  
DispatchGo is designed for environments where introducing a message broker adds unnecessary operational overhead.

Using PostgreSQL with `FOR UPDATE SKIP LOCKED` allows us to:
- Achieve safe concurrent dequeuing
- Persist job state and delivery history in one place
- Avoid additional infrastructure

> This is a pragmatic trade-off: leveraging an existing system to solve queuing while accepting its limitations at very high scale.

For moderate workloads (hundreds to low thousands of jobs/min), this approach performs well. At higher throughput, a dedicated broker like RabbitMQ or Kafka would be more appropriate.

---

**Why a fixed worker pool?**  
Spawning a goroutine per job creates **unbounded concurrency**, which can exhaust system resources under load.

A fixed worker pool:
- Enforces back-pressure
- Keeps database connections bounded
- Provides predictable resource usage

> This aligns throughput with system capacity instead of letting load dictate behavior.

---

**Why `SKIP LOCKED` over application-level locking?**  
Concurrency control is delegated to the database.

`SELECT ... FOR UPDATE SKIP LOCKED` ensures:
- Each job is claimed by only one worker
- No duplicate processing due to race conditions
- Minimal coordination logic in application code

> This leverages database guarantees instead of re-implementing locking in Go.

---

**Why store full request and response?**  
Webhook failures are notoriously hard to debug.

Storing:
- Full payload sent
- Response body (bounded)
- Status code and attempt metadata

enables:
- Post-mortem debugging
- Replay and verification
- Operational transparency

> This avoids relying solely on logs, which are often incomplete or ephemeral.

---

**Why a job lifecycle model?**  
Each webhook is treated as a state machine:

`pending → processing → delivered | failed → exhausted`

This explicit lifecycle ensures:
- Clear retry semantics
- Visibility into system state
- Safe recovery after crashes

---

## Trade-offs

Polling introduces latency.

Jobs are picked up within the polling interval (default ~1 second), not instantly.

This is acceptable because:
- Webhooks are asynchronous by nature
- Sub-second latency is not critical

If lower latency is required, this can be improved using `LISTEN/NOTIFY` to wake workers immediately.

---

Row-level locking introduces database load.

Under high concurrency, `SELECT ... FOR UPDATE SKIP LOCKED` can become a bottleneck.

Mitigation:
- Partial index on `(status, next_attempt_at)`
- Keeping the working set small
- Limiting worker pool size

---

PostgreSQL is not a purpose-built queue.

At very high throughput (100k+ jobs/min):
- Lock contention increases
- Query performance degrades

> This design prioritizes **simplicity and operability over extreme scalability**.

---

## Failure Handling

**Worker panics**  
Each worker runs inside a recovery wrapper.

If a panic occurs:
- Stack trace is logged
- Job is marked as `failed`
- Worker is restarted automatically

> The system maintains a constant worker pool size and avoids silent failures.

---

**Receiver timeout**  
HTTP delivery uses a configurable timeout (default ~10 seconds).

Timeouts are treated as transient failures and retried using exponential backoff:

`30s → 2m → 10m → 1h → 6h`

After exhausting retries, the job transitions to `exhausted`.

---

**Database unavailability**  
Workers detect connection failures and back off with jitter.

No jobs are lost because:
- Jobs remain persisted in PostgreSQL
- Processing resumes automatically when the database recovers

---

**Duplicate delivery (at-least-once semantics)**  
The system guarantees at-least-once delivery.

This means duplicates are possible in failure scenarios.

Mitigation:
- Downstream consumers are expected to handle idempotency
- Job IDs can be used as idempotency keys

> This avoids the complexity of exactly-once delivery while maintaining correctness.

---

## Improvements

If I were to evolve this system further:

- Introduce **RabbitMQ or Kafka** for higher throughput and reduced database contention as scale increases
- Implement a **dead-letter queue UI** for inspecting and replaying exhausted jobs without requiring direct database access
- Add **webhook signature verification (HMAC)** to ensure only trusted clients can submit jobs
- Integrate **observability (metrics, tracing, structured logging)** to improve debugging and performance monitoring
- Replace polling with **event-driven wake-up (LISTEN/NOTIFY or broker push model)** to reduce latency and database load