---
title: "URL Shortener"
---

## Problem

URL shorteners appear simple until you consider **read-heavy production traffic patterns**.

The system is highly asymmetric:
- Writes are rare (creating short URLs)
- Reads are extremely frequent (redirects)

A naive implementation queries PostgreSQL on every redirect, which works at low scale but fails under high read load due to:
- Database contention
- Increased latency
- Poor scalability

Additionally, production systems must handle:
- Hot keys (popular links receiving massive traffic)
- Efficient expiry of links
- Low-latency redirects under load

This project focuses on optimizing the **read path using layered caching and TTL-based lifecycle management**, ensuring minimal latency and reduced database pressure.

---

## Architecture

The read path is optimized through a three-layer caching strategy:

1. **In-process cache**  
   A `sync.Map` storing the most recently accessed slug → URL mappings (~1,000 entries).  
   - Zero network latency  
   - Handles hot keys efficiently  
   - Achieves high hit rates due to skewed access patterns (long-tail distribution)

2. **Redis (distributed cache)**  
   Stores all active slugs with TTL aligned to link expiry.  
   - Acts as the primary cache layer  
   - Shared across instances  
   - Reduces database load significantly

3. **PostgreSQL (source of truth)**  
   Accessed only on cache misses or after Redis eviction.  
   - Stores canonical mapping and metadata  
   - Ensures durability and correctness

Write path:
- Writes go to PostgreSQL first
- Redis is updated synchronously (write-through)
- In-process cache is populated lazily on read

> This layered design minimizes latency while preserving correctness and durability.

---

## Design Decisions

**Why layered caching instead of a single cache?**  
Different layers solve different latency and consistency problems:

- In-process cache → ultra-fast access for hot keys
- Redis → shared cache across instances
- Postgres → durable source of truth

> This reduces load progressively instead of relying on a single bottleneck.

---

**Why `sync.Map` over a custom LRU?**  
A custom LRU provides precise eviction control and potentially better performance, but introduces complexity.

`sync.Map` offers:
- Thread-safe access out of the box
- Good performance for read-heavy workloads
- Simplicity and reliability

> The decision follows “measure before optimizing” — replace only if contention becomes a bottleneck.

---

**Why TTL-based expiry instead of soft deletes?**  
Expired links should automatically disappear without manual intervention.

- Redis TTL handles cache expiry efficiently
- PostgreSQL uses a periodic cleanup job to remove expired rows

> This keeps the system self-maintaining and prevents unbounded data growth.

---

**Why write-through caching (Postgres → Redis)?**  
Ensures Redis always has the latest data immediately after creation.

> This avoids cache misses immediately after writes and simplifies consistency logic.

---

**Why not use a CDN for redirects?**  
CDNs can cache HTTP 301 responses, but introduce limitations:

- No control over invalidation
- Difficult to update or deactivate links
- Cached responses may persist beyond intended lifecycle

Using HTTP 302 keeps control on the server side.

> This prioritizes **flexibility and correctness over edge caching performance**.

---

## Trade-offs

The in-process cache is instance-local.

In multi-instance deployments:
- Cache invalidation is not synchronized across instances
- A stale entry may exist temporarily

This is acceptable because:
- Entries are short-lived (TTL-bound)
- Staleness impact is minimal for this use case

---

Click analytics are handled asynchronously.

- Events are buffered via channels
- Under extreme load, events may be dropped

> This prioritizes **user-facing latency over analytics accuracy**, ensuring redirects are never blocked.

---

Memory usage increases with cache size.

- In-process cache is bounded (~1,000 entries)
- Redis handles larger datasets efficiently

---

## Failure Handling

**Redis down**  
The system falls back to PostgreSQL.

- Latency increases
- Correctness is preserved

All Redis errors are logged for operational visibility.

---

**PostgreSQL down**  
- Cached slugs (in-memory + Redis) continue to work
- New URL creation fails with 503

> This ensures partial availability for read-heavy workloads.

---

**Slug collision**  
Slug generation uses base62 encoding with 7 characters:

`62^7 ≈ 3.5 trillion combinations`

Collisions are extremely rare and handled via:
- Unique constraint in PostgreSQL
- Retry with a new slug on conflict

---

## Improvements

If I were to evolve this system further:

- Introduce **consistent hashing** to route requests to specific instances, improving in-process cache effectiveness in distributed setups
- Add **Bloom filters** to prevent unnecessary database hits for non-existent slugs
- Implement **write-behind or async cache population strategies** for better write scalability
- Add **rate limiting (e.g., HoldUp)** to protect against abuse and hot-key amplification
- Integrate **observability (metrics, tracing, cache hit ratios)** to monitor performance and identify bottlenecks
- Provide **custom domains and QR code generation** as user-facing features