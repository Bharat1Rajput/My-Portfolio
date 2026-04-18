---
title: "FlowPay"
---

## Problem

Payment processing is deceptively hard because failures are not logical — they are **distributed systems failures**.

A naive approach — a single synchronous HTTP call to a payment gateway — breaks in production in predictable ways:
- The network times out after the user has already been charged
- The gateway responds but the acknowledgment is lost
- A retry results in a **duplicate charge**

These failures create **inconsistent system state**, which is unacceptable in financial systems.

FlowPay solves this by treating payments not as request-response operations, but as **idempotent, stateful workflows driven by events**, where every transition is explicitly recorded and recoverable.

---

## Architecture

FlowPay is composed of four services communicating via Kafka topics:

- **API Service** — accepts inbound payment requests, validates them, assigns an idempotency key, persists the payment intent, and publishes a `payment.initiated` event.
- **Processor Service** — consumes `payment.initiated`, executes the payment via the external gateway, enforces state transitions, and publishes either `payment.succeeded` or `payment.failed`.
- **Ledger Service** — consumes all payment events and maintains a **double-entry ledger** in PostgreSQL, acting as the financial source of truth.
- **Notifier Service** — consumes terminal events (`succeeded`, `failed`) and dispatches webhooks or notifications to the merchant.

Each service owns its own PostgreSQL schema and communicates exclusively through Kafka. There are no direct service-to-service calls; all coordination happens through events.

> This design enforces **loose coupling, independent scalability, and failure isolation**.

---

## Design Decisions

**Why Kafka over RabbitMQ?**  
Kafka provides a durable, append-only event log with consumer-controlled offsets. This enables **event replay**, which is critical in financial systems.  
For example, when a bug in the Ledger Service caused incorrect balance calculations, I was able to fix the logic and replay historical events to rebuild the ledger accurately — something not possible with traditional queue systems like RabbitMQ where messages are removed after acknowledgment.

---

**Why idempotency keys at the API layer?**  
Instead of deduplicating downstream, FlowPay prevents duplicates from entering the system.

The client provides an idempotency key, which is stored with the payment intent before emitting any event. On retries, the API returns the cached response without reprocessing.

> This shifts complexity from distributed deduplication to a **single controlled entry point**, significantly reducing system-wide inconsistency risks.

---

**Why a state machine?**  
Payments follow a strict lifecycle:

`initiated → processing → succeeded | failed → refunded`

This lifecycle is enforced as a **state machine in the Processor Service**, ensuring invalid transitions (e.g., refunding a failed payment) are impossible.

> This makes correctness a **design-time guarantee**, not a runtime assumption.

---

**Why a separate Ledger Service?**  
Instead of storing balances directly, FlowPay uses a **double-entry ledger model**.

Every event results in balanced debit/credit entries, ensuring:
- Financial correctness
- Auditability
- No possibility of “lost money”

> The ledger becomes the **single source of truth**, independent of other services.

---

**Why event-driven communication only?**  
There are no direct gRPC/HTTP calls between services.

> This avoids tight coupling, allows independent scaling, and ensures that failures in one service do not cascade across the system.

---

## Trade-offs

The event-driven approach introduces **eventual consistency**.

The API responds with `202 Accepted` immediately, while the final state is resolved asynchronously. The merchant may receive a webhook after a short delay depending on Kafka lag.

This trade-off favors **reliability and fault tolerance over immediate consistency**, which is acceptable for most online payment systems but not for real-time POS systems.

To mitigate this:
- A polling endpoint is provided
- Webhooks deliver asynchronous updates

---

Kafka also introduces **operational overhead**.

Running a distributed log system for a small workload is heavy, so:
- A single-broker setup is used in development
- A managed Kafka cluster is used in production

This trade-off is justified by the **replay capability and system decoupling** Kafka enables.

---

## Failure Handling

**Processor crashes mid-payment**  
Kafka offsets are committed only after the gateway response is processed and the result event is published.  
If a crash occurs before that, the message is re-delivered.

To ensure safety, the gateway call uses an idempotency key derived from the internal payment ID, making retries safe.

---

**Gateway timeout**  
The Processor retries the request with exponential backoff.  
After a fixed number of retries, it emits a `payment.failed` event with reason `gateway_timeout`.

The Notifier then informs the merchant to take corrective action.

---

**Kafka broker down**  
The API service persists payment intents in PostgreSQL before publishing events.

A background publisher (outbox pattern) asynchronously reads pending records and publishes them to Kafka with **at-least-once delivery semantics**.

Once Kafka recovers, the system automatically catches up without data loss.

---

## Improvements

If I were to build the next version:

- Introduce a **Saga orchestrator** to replace implicit choreography and manage complex multi-step workflows explicitly as the system scales.
- Add **distributed tracing and observability** (OpenTelemetry, metrics, structured logging with correlation IDs) to improve debugging in an event-driven system.
- Provide a **gRPC streaming API** for real-time payment status updates, eliminating the need for polling.
- Explore **partitioning strategies and ordering guarantees** in Kafka to optimize throughput and consistency for high-volume workloads.