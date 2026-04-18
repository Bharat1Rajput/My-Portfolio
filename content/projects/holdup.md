---
title: "HoldUp"
---

## Problem

Rate limiting is not just about throttling traffic — it is about **protecting system stability under unpredictable load**.

Most implementations fail in one of two ways:
- They depend on external systems (e.g., Redis), introducing latency and a potential single point of failure
- They use naive in-memory counters that break under concurrency or do not handle burst traffic correctly

Additionally, many implementations ignore critical concerns like:
- Unbounded memory growth
- Goroutine explosion under high cardinality (many clients)
- Race conditions in concurrent updates

HoldUp solves this by providing a **deterministic, in-process, concurrency-safe rate limiter** based on the token bucket algorithm, designed to be lightweight, predictable, and verifiable.

---

## Architecture

HoldUp is a single Go package with no external dependencies. Core components:

- **Bucket** — maintains current token count, capacity, refill rate, and last refill timestamp. Protected by a `sync.Mutex` to ensure atomic updates.
- **Limiter** — a map of `key → *Bucket` (e.g., client IP or API key). Uses `sync.RWMutex` to allow concurrent reads while synchronizing bucket creation.
- **Middleware** — a standard `func(http.Handler) http.Handler` that extracts the key, checks allowance, and either forwards the request or returns HTTP 429.

Token refill is computed lazily on each request using elapsed time since the last update.

> This design avoids background workers entirely — **zero goroutines and zero allocations per request path**.

---

## Design Decisions

**Why token bucket over leaky bucket or sliding window?**  
Token bucket allows **controlled bursts**, which aligns with real-world API usage patterns.

For example, a client that has been idle should be allowed to make multiple requests immediately (e.g., dashboard loading multiple resources). Sliding window enforces strict fairness but penalizes such burst behavior, degrading user experience.

> The goal is not just fairness — it is **practical usability under bursty workloads**.

---

**Why lazy refill instead of a background ticker?**  
A ticker-based refill introduces a **goroutine-per-bucket problem**.

At scale (e.g., 10k+ unique clients), this leads to:
- Excessive goroutine overhead
- Increased scheduling pressure
- Memory inefficiency

Lazy refill computes tokens only when needed using simple arithmetic.

> This achieves the same correctness with **O(1) computation and zero background overhead**.

---

**Why `sync.Mutex` per bucket instead of atomic operations?**  
Token updates involve multiple fields:
- Token count
- Last refill timestamp

These must be updated atomically as a unit.

Using atomics would require complex CAS loops, increasing code complexity without meaningful performance gains at typical API concurrency levels.

> A mutex provides **clarity, correctness, and sufficient performance**.

---

**Why in-process design (no Redis)?**  
HoldUp intentionally avoids external dependencies to:
- Minimize latency
- Remove network failure modes
- Keep deployment simple

> This makes it suitable as a **first line of defense** in the request path.

---

## Trade-offs

HoldUp is **instance-local**.

In a horizontally scaled system:
- Each instance maintains its own rate limits
- A client can exceed global limits by distributing requests across instances

This is acceptable for many systems where:
- Rate limiting is primarily a protection mechanism
- Not a strict billing or quota enforcement tool

For strict global limits, a shared store like Redis is required.

---

Memory usage grows with the number of unique clients.

Each new key creates a bucket, leading to linear growth.

Mitigation:
- A `Cleanup(maxAge)` function removes stale buckets
- Responsibility is delegated to the caller to run periodically

> This keeps the library lightweight and avoids hidden background behavior.

---

## Failure Handling

**Concurrent bucket creation**  
The Limiter uses a double-checked locking pattern:
- First checks under read lock
- If missing, acquires write lock and checks again before inserting

This prevents duplicate bucket creation under race conditions while keeping the common path fast.

---

**Clock skew**  
If the system clock moves backward (e.g., NTP adjustment), elapsed time may become negative.

We clamp elapsed time to zero to prevent token subtraction.

> This ensures the system degrades safely instead of corrupting state.

---

**High concurrency contention**  
Under heavy load on a single key, the bucket mutex can become a contention point.

This is acceptable because:
- Rate limiting is inherently serialized per key
- Contention reflects real pressure from that client

---

## Improvements

If I were to evolve this further:

- Add a **distributed mode (Redis-backed sliding window or token bucket)** for enforcing global limits across instances
- Combine local (in-process) + global limiter to reduce Redis load (local acts as a fast pre-filter)
- Add **Prometheus metrics** (request rate, 429 responses, token levels) for observability and alerting
- Introduce **adaptive rate limiting** (dynamic limits based on system load)
- Support **pluggable key strategies** (IP, API key, user ID, custom extractor functions)