# Golang Tutorials & Concurrency Showcase

Welcome to this interactive, step-by-step Golang tutorial workspace. This repository is structured to take you from basic Golang syntax to advanced concurrency patterns, design patterns, and real-world HTTP applications.

---

## 🛠️ Project Structure

Every subdirectory is a self-contained executable demonstration. You can run any tutorial by running `go run ./<directory_name>` from the root.

```
.
├── go.mod                          # Root Go module
├── README.md                       # This document
│
│── 01_basic_syntax/                # Variables, control flow, loops, structs, interfaces
│── 02_goroutines/                  # Concurrency with go keyword, sync.WaitGroup
│── 03_channels_unbuffered/         # Synchronous signaling & blocking channels
│── 04_channels_buffered/           # Asynchronous message queuing & range iteration
│── 05_select_multiplexing/         # Multi-channel select, timeouts, non-blocking ops
│── 06_worker_pool/                 # Thread-safe concurrent task processor pattern
│
│── 07_producer_consumer/           # Multiple producers and consumers using channels
│── 08_fan_out/                     # Distribute work across multiple goroutines
│── 09_fan_in/                      # Merge results from multiple goroutines into one channel
│── 10_pipeline/                    # Multi-stage processing pipeline
│── 11_concurrent_api_calls/        # Call N APIs in parallel and aggregate results
│── 12_context_cancellation/        # Cancel all child goroutines when parent is cancelled
│── 13_context_timeout/             # Abort long-running operations after timeout
│── 14_graceful_shutdown/           # Shutdown HTTP server and workers cleanly
│── 15_semaphore/                   # Limit concurrent execution to N goroutines
│── 16_rate_limiter/                # Token bucket request throttling
│
│── 17_batch_processor/             # Collect events and process in batches
│── 18_retry_mechanism/             # Retry with exponential backoff
│── 19_circuit_breaker/             # Prevent cascading failures
│── 20_concurrent_task_queue/       # Thread-safe task processing queue
│── 21_event_bus/                   # Publish-subscribe system inside a process
│── 22_pubsub/                      # Topic-based pub/sub system
│── 23_future_promise/              # Execute async work and retrieve result later
│── 24_safe_counter/                # Goroutine-safe counter (Mutex & Atomic)
│── 25_concurrent_map/              # Thread-safe map implementations
│
│── 26_lru_cache/                   # Thread-safe LRU cache
│── 27_cache_aside/                 # Database + cache pattern
│── 28_repository_pattern/          # Abstract data layer behind interfaces
│── 29_generic_cache/               # Reusable cache using Go generics
│── 30_middleware_chain/            # HTTP middleware execution pipeline
│
│── 31_jwt_auth/                    # JWT generation and validation middleware
│── 32_crud_rest_api/               # Full REST service with validation
│── 33_api_rate_limit/              # Per-IP API rate limiting middleware
│── 34_panic_recovery/              # Recover from panics and return proper responses
└── 35_url_shortener/               # Mini system: API, storage, and business logic
```

---

## 📚 Topics Covered

### 1. Basic Syntax (`01_basic_syntax`)
A quick refresher on Go's syntax.
- Strong static typing and type inference (`:=`).
- Implicit loops: Go only has the `for` keyword, which acts as `while` and classic `for` loops.
- Error handling: Idiomatic Go style returns values and errors explicitly.
- Structs and Methods: Object-oriented concepts without traditional inheritance.
- Interfaces: Duck-typing interfaces for flexible design.

---

### 2. Goroutines (`02_goroutines`)
Goroutines are lightweight threads managed by the Go runtime.
- Spawning a goroutine: Use the prefix keyword `go`.
- Synchronization: Without coordination, the main goroutine will exit immediately before background workers run.
- `sync.WaitGroup`: Uses a thread-safe counter (`Add`, `Done`, `Wait`) to block execution until all workers complete.

---

### 3. Channels: Unbuffered (`03_channels_unbuffered`)
Channels are typed conduits through which you can send and receive values with the channel operator, `<-`.
- **Unbuffered Channels** (`ch := make(chan T)`) have no capacity.
- Sending blocks until a receiver is ready to read, and receiving blocks until a sender writes. This ensures synchronous synchronization.
```
Sender (Blocks)  ───────────> [ Channel ] ───────────> Receiver (Blocks)
                               (Capacity: 0)
```

---

### 4. Channels: Buffered (`04_channels_buffered`)
- **Buffered Channels** (`ch := make(chan T, capacity)`) have a predefined queue capacity.
- Sends are non-blocking as long as the buffer is not full. Receives block only when the buffer is empty.
- Excellent for decoupling senders and receivers (e.g., rate-limiting, job queueing).
```
Sender (No Block) ──> [ [Item] | [Item] | [    ] ] ──> Receiver (Blocks if Empty)
                         (Capacity: 3)
```

---

### 5. Select Multiplexing (`05_select_multiplexing`)
The `select` statement lets a goroutine wait on multiple communication operations.
- A `select` blocks until one of its cases can run, then it executes that case. If multiple are ready, it chooses one pseudo-randomly.
- Crucial for implementing:
  - Timeouts (`time.After`)
  - Non-blocking channel operations (using a `default` case)
  - Graceful cancellation

---

### 6. Worker Pools (`06_worker_pool`)
A classic design pattern to limit resource consumption by spawning a fixed number of workers to process a queue of jobs.
```
                  ┌──────────┐
                  │ Worker 1 │ ───┐
                  └──────────┘    │
                  ┌──────────┐    │
Jobs Channel ───> │ Worker 2 │ ───┼───> Results Channel
                  └──────────┘    │
                  ┌──────────┐    │
                  │ Worker 3 │ ───┘
                  └──────────┘
```

---

### 7. Producer-Consumer (`07_producer_consumer`)
Decouples data production from consumption using a shared buffered channel. Multiple producers generate items independently while multiple consumers process them concurrently.
```
Producer 1 ──┐                  ┌── Consumer 1
Producer 2 ──┼──▶ [Channel] ──▶─┼── Consumer 2
Producer 3 ──┘                  └── Consumer 3
```

---

### 8. Fan-Out Pattern (`08_fan_out`)
Distributes work from a single source across multiple goroutines for parallel processing. Go's channel semantics ensure each task is delivered to exactly one worker.

---

### 9. Fan-In Pattern (`09_fan_in`)
The inverse of Fan-Out. Multiple goroutines each produce data on their own channels, and a fan-in function merges all channels into one unified output.
```
Source 1 ──▶ ch1 ──┐
                   │
Source 2 ──▶ ch2 ──┼──▶ mergedCh ──▶ Consumer
                   │
Source 3 ──▶ ch3 ──┘
```

---

### 10. Pipeline Processing (`10_pipeline`)
A series of stages connected by channels. Each stage receives from upstream, processes data, and sends downstream. All stages run concurrently.
```
generate → [ch1] → square → [ch2] → filterEven → [ch3] → addLabel → consumer
```

---

### 11. Concurrent API Calls (`11_concurrent_api_calls`)
Fire N API calls in parallel using goroutines, aggregate results with a mutex-protected slice, and compare sequential vs concurrent execution times.

---

### 12. Context Cancellation (`12_context_cancellation`)
`context.WithCancel` propagates cancellation signals to all child goroutines. Every long-running goroutine selects on `ctx.Done()` to respond to cancellation promptly.

---

### 13. Context Timeout (`13_context_timeout`)
`context.WithTimeout` automatically cancels after a specified duration. Demonstrates deadline management across sequential operations sharing a budget.

---

### 14. Graceful Shutdown (`14_graceful_shutdown`)
Intercepts OS signals (SIGINT/SIGTERM), stops the HTTP server from accepting new connections, waits for in-flight requests, signals background workers, and cleans up — all with a bounded timeout.

---

### 15. Semaphore using Channels (`15_semaphore`)
Uses a buffered channel of capacity N as a counting semaphore. Sending = acquire, receiving = release. Limits concurrent goroutines to N.

---

### 16. Rate Limiter — Token Bucket (`16_rate_limiter`)
Implements the token bucket algorithm: a bucket holds up to N tokens refilled at a fixed rate. Supports blocking and non-blocking modes.

---

### 17. Batch Processor (`17_batch_processor`)
Collects incoming events and processes them in batches. Flushes when the batch is full OR a timeout expires. Handles graceful shutdown with final flush.

---

### 18. Retry Mechanism (`18_retry_mechanism`)
Retries failed operations with exponential backoff and jitter. Configurable max retries, base delay, multiplier, max delay, and jitter ratio.

---

### 19. Circuit Breaker (`19_circuit_breaker`)
State machine with CLOSED → OPEN → HALF-OPEN transitions. Prevents cascading failures by blocking requests when failure threshold is reached.

---

### 20. Concurrent Task Queue (`20_concurrent_task_queue`)
Thread-safe task queue with Enqueue/Dequeue API, task status tracking (PENDING/RUNNING/COMPLETED/FAILED), configurable worker pool, and graceful shutdown.

---

### 21. Event Bus (`21_event_bus`)
In-process publish-subscribe system. Components subscribe callback functions to event names. Supports sync and async publishing, plus unsubscribe.

---

### 22. Pub/Sub System (`22_pubsub`)
Topic-based publish-subscribe with dedicated subscriber channels. Each subscriber has its own buffered channel providing backpressure. Supports multiple topics per subscriber.

---

### 23. Future/Promise Pattern (`23_future_promise`)
Generic `Future[T]` using Go generics. `Async()` starts work and returns a Future immediately. `Get()`, `GetWithTimeout()`, `Then()` chaining, and `AwaitAll()` for parallel collection.

---

### 24. Goroutine-Safe Counter (`24_safe_counter`)
Compares three approaches: unsafe (demonstrates race condition), `sync.Mutex`, and `sync/atomic`. Includes benchmarks showing atomic is faster for simple counters.

---

### 25. Concurrent Map (`25_concurrent_map`)
Generic `ConcurrentMap[K, V]` using `sync.RWMutex` (multiple readers OR one writer). Compared with `sync.Map` from the standard library including performance benchmarks.

---

### 26. LRU Cache (`26_lru_cache`)
Thread-safe Least Recently Used cache using HashMap + Doubly-Linked List for O(1) operations. Tracks hits/misses and supports concurrent access.

---

### 27. Cache-Aside Pattern (`27_cache_aside`)
Application checks cache first; on miss, loads from database, stores in cache. Writes go to DB first then invalidate cache. Includes TTL support.

---

### 28. Repository Pattern (`28_repository_pattern`)
Abstracts data access behind a `UserRepository` interface. Business logic depends on the interface, enabling easy swapping of database backends and unit testing with mocks.

---

### 29. Generic Cache (`29_generic_cache`)
Reusable `Cache[K, V]` using Go generics with TTL per entry, background cleanup of expired entries, `GetOrSet` for lazy loading, and eviction callbacks.

---

### 30. Middleware Chain (`30_middleware_chain`)
Composable HTTP middleware pipeline. Includes logging, CORS, auth, request ID, and panic recovery middlewares. `Chain()` function composes them into a single wrapper.

---

### 31. JWT Authentication Middleware (`31_jwt_auth`)
JWT generation and validation from scratch using HMAC-SHA256. Includes role-based authorization middleware and login endpoint.

---

### 32. CRUD REST API (`32_crud_rest_api`)
Full RESTful API for a "Book" resource: CREATE, READ (list + single), UPDATE, DELETE with input validation, proper HTTP status codes, and JSON responses.

---

### 33. API Rate Limiting Middleware (`33_api_rate_limit`)
Per-IP sliding window rate limiter with proper HTTP headers (`X-RateLimit-*`, `Retry-After`), 429 responses, and automatic cleanup of stale records.

---

### 34. Panic Recovery Middleware (`34_panic_recovery`)
Catches panics from HTTP handlers, logs stack traces, returns proper 500 responses, and keeps the server running. Includes configurable recovery options.

---

### 35. URL Shortener (`35_url_shortener`)
Mini URL shortening system: POST to create short URLs, GET to redirect with click tracking, stats endpoint, and URL validation. Combines API design, storage, and business logic.

---

## 🚀 Running the Tutorials

You can execute any tutorial from the workspace root:

```bash
# Fundamentals
go run ./01_basic_syntax
go run ./02_goroutines
go run ./03_channels_unbuffered
go run ./04_channels_buffered
go run ./05_select_multiplexing
go run ./06_worker_pool

# Concurrency Patterns
go run ./07_producer_consumer
go run ./08_fan_out
go run ./09_fan_in
go run ./10_pipeline
go run ./11_concurrent_api_calls
go run ./12_context_cancellation
go run ./13_context_timeout
go run ./14_graceful_shutdown        # Press Ctrl+C to test graceful shutdown
go run ./15_semaphore
go run ./16_rate_limiter

# Advanced Patterns
go run ./17_batch_processor
go run ./18_retry_mechanism
go run ./19_circuit_breaker
go run ./20_concurrent_task_queue
go run ./21_event_bus
go run ./22_pubsub
go run ./23_future_promise
go run ./24_safe_counter
go run ./25_concurrent_map

# Data & Design Patterns
go run ./26_lru_cache
go run ./27_cache_aside
go run ./28_repository_pattern
go run ./29_generic_cache
go run ./30_middleware_chain          # Starts HTTP server on :8080

# Web / HTTP Patterns
go run ./31_jwt_auth                  # Starts HTTP server on :8080
go run ./32_crud_rest_api             # Starts HTTP server on :8080
go run ./33_api_rate_limit            # Starts HTTP server on :8080
go run ./34_panic_recovery            # Starts HTTP server on :8080
go run ./35_url_shortener             # Starts HTTP server on :8080
```

> **Note**: Tutorials 14, 30–35 start HTTP servers on port `:8080`. Only run one at a time.
