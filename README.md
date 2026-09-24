
# Sentinel

**Sentinel** is a high-performance, lightweight distributed edge platform and reverse proxy written in Go.

It is built around a fast Data Plane request path with routing, rate limiting, load balancing, health checking, circuit breaking, retries, and reverse proxying.

The project is developed incrementally with a focus on **correctness, measurable performance, concurrency safety, and system-level engineering**.

---

## Architecture

The current architecture separates the **Data Plane** from the **Control Plane**.

```text
                         SENTINEL
                            │
             ┌──────────────┴──────────────┐
             │                             │
        CONTROL PLANE                 DATA PLANE
        Source of Truth              Live Traffic
             │                             │
          SQLite                    Runtime Config
             │                             │
             │                    ┌────────▼────────┐
             │                    │     Router      │
             │                    └────────┬────────┘
             │                             │
             │                    ┌────────▼─────────┐
             │                    │   Rate Limiter   │
             │                    └────────┬──────────┘
             │                             │
             │                    ┌────────▼─────────┐
             │                    │  Load Balancer   │
             │                    └────────┬──────────┘
             │                             │
             │                    ┌────────▼─────────┐
             │                    │ Circuit Breaker  │
             │                    └────────┬──────────┘
             │                             │
             │                    ┌────────▼─────────┐
             │                    │  Retry Engine    │
             │                    └────────┬──────────┘
             │                             │
             │                    ┌────────▼─────────┐
             │                    │ Reverse Proxy    │
             │                    └────────┬──────────┘
             │                             │
             │                    ┌────────▼─────────┐
             │                    │     Backend      │
             │                    └──────────────────┘
             │
             └──────────── gRPC Streaming ────────────►

```

The Data Plane is designed so that the HTTP hot path does not perform Control Plane network, database, or disk operations.

---

## Features

* **High-performance HTTP routing**
* Radix-tree router
* Parameterized routes


* **Per-client token-bucket rate limiting**
* Atomic CAS-based rate limiting


* **Round Robin load balancing**
* **Least Connections load balancing**
* **Power of Two Choices load balancing**
* **Active HTTP/TCP health checking**
* **Circuit breaker**
* Half-open recovery probing


* **Retry engine**
* Exponential backoff
* Full-jitter retry strategy
* Request-body replay safety


* **Reverse proxying**
* **gRPC-based Control Plane**
* Dynamic runtime configuration
* SQLite persistence
* Automatic Control Plane reconnection


* **Graceful shutdown**
* **Health and readiness endpoints**
* **CPU and memory profiling with `pprof**`
* **Unit, integration, concurrency, race, and benchmark testing**

---

## Data Plane

The Data Plane handles live HTTP traffic.

The request lifecycle is:

```text
Client
  │
  ▼
Router
  │
  ▼
Rate Limiter
  │
  ▼
Load Balancer
  │
  ▼
Circuit Breaker
  │
  ▼
Retry Engine
  │
  ▼
Reverse Proxy
  │
  ▼
Backend

```

The hot path is intentionally kept independent of the Control Plane.

### Radix Router

Sentinel uses a radix-tree based router for efficient HTTP route matching.

Supported routing includes:

* `/users`
* `/users/:id`
* `/videos/:id`
* `/files/*`

The router is designed for:

* Fast lookups
* Low allocations
* Parameter extraction
* Wildcard matching
* Concurrent request processing

### Rate Limiter

Sentinel implements a per-client **Token Bucket** rate limiter.

Example:

```yaml
rate_limit:
  enabled: true
  capacity: 100
  refill_rate: 100

```

The token bucket uses atomic Compare-And-Swap operations on the hot path.

```text
Client
   │
   ▼
Bucket Map
   │
   ▼
Client Bucket
   │
   ▼
Atomic CAS
   │
   ▼
Allow / Reject

```

Requests exceeding the configured limit receive:
`HTTP 429 Too Many Requests`

#### Current Scope

The rate limiter is instance-local. Multiple Sentinel instances therefore maintain independent token state. Distributed rate limiting using shared state or coordination is planned for a future stage.

### Load Balancing

Sentinel supports multiple backend-selection strategies:

* **Round Robin**
* **Least Connections**
* **Power of Two Choices**

The selection layer is concurrency-safe and designed to keep backend selection lightweight.

Round Robin provides a simple baseline:

* Request 1 → Backend A
* Request 2 → Backend B
* Request 3 → Backend C
* Request 4 → Backend A

Different strategies can be benchmarked under different backend workloads rather than assuming one strategy is universally optimal.

### Health Checking

Sentinel continuously checks backend health.

Health checking supports:

* HTTP health checks
* TCP fallback
* Backend health tracking
* Removal of unhealthy backends from traffic selection

```text
             Health Checker
                   │
          ┌────────┼────────┐
          ▼        ▼        ▼
       Backend A Backend B Backend C
          │        │        │
          ▼        ▼        ▼
       Healthy  Unhealthy Healthy

```

Health checking operates independently from request retries.

### Circuit Breaker

Sentinel implements a three-state circuit breaker:

```text
              failures
                 │
                 ▼
             ┌────────┐
        ┌───►│ CLOSED │
        │    └────┬───┘
        │         │
        │         ▼
        │    ┌────────┐
        │    │  OPEN  │
        │    └────┬───┘
        │         │
        │       timeout
        │         ▼
        │   ┌───────────┐
        └───│ HALF-OPEN │
            └─────┬─────┘
                  │
             successful
                probe
                  │
                  ▼
               CLOSED

```

The circuit breaker supports:

* Failure thresholds
* Recovery timeout
* Half-open state
* Single concurrent recovery probe

The breaker evaluates the outcome of the **logical request**, rather than treating every retry attempt as an independent request.

### Retry Engine

Sentinel supports retries for transient upstream failures.

Retryable HTTP status codes include:

* `502 Bad Gateway`
* `503 Service Unavailable`
* `504 Gateway Timeout`

By default, retries are enabled for methods considered safe or replayable:

* `GET`
* `HEAD`
* `OPTIONS`
* `PUT`
* `DELETE`

Methods such as `POST` and `PATCH` are not retried by default. Retryable network failures are also supported for eligible requests.

#### Retry + Circuit Breaker

Retries and circuit breaking operate together:

```text
Circuit Breaker
      │
      ▼
 Retry Engine
      │
      ▼
   Backend

```

A logical request may result in multiple upstream attempts. The circuit breaker evaluates the overall logical request outcome instead of counting every retry attempt separately. Retries remain associated with the selected backend.

#### Exponential Backoff + Full Jitter

Retry delays use exponential backoff with full jitter.

Conceptually:

* **Attempt 1** → random delay within Base
* **Attempt 2** → random delay within $2 \times \text{Base}$
* **Attempt 3** → random delay within $4 \times \text{Base}$
* **Attempt 4** → random delay within $8 \times \text{Base}$

The delay is capped by a configured maximum. Full jitter reduces synchronized retry bursts when multiple clients encounter failures simultaneously.

#### Request Body Replay Safety

Sentinel does not blindly retry requests whose bodies have already been consumed. A request is retried only when its body can safely be replayed.

This prevents retries from receiving:

* Empty bodies
* Partially consumed bodies
* Invalid payloads

---

## Control Plane

The Control Plane provides configuration management and distribution to Data Plane nodes.

The current architecture uses:

```text
                 CONTROL PLANE
                       │
                    SQLite
                       │
                 In-Memory Store
                       │
                gRPC Streaming
                       │
                       ▼
                  DATA PLANE
                       │
               Runtime Config
                       │
                atomic.Pointer

```

### Responsibilities

* Service management
* Backend management
* Route management
* Persistent configuration
* Configuration snapshots
* Runtime configuration distribution

The Data Plane maintains its last-known configuration when the Control Plane becomes temporarily unavailable. It automatically reconnects using exponential backoff.

> **Important Design Principle**
> The Data Plane HTTP hot path does **not** access:
> * SQLite
> * Control Plane network calls
> * Configuration storage
> 
> 
> Runtime configuration is held in memory and updated atomically.

---

## Configuration

Sentinel supports YAML configuration for local Data Plane settings.

Example:

```yaml
server:
  port: 8080

rate_limit:
  enabled: true
  capacity: 100
  refill_rate: 100

routes:
  - path: /users
    backend: users

backends:
  - name: users
    url: http://localhost:9001

```

Control Plane managed configuration is distributed dynamically to Data Plane nodes.

---

## Performance

Performance is a core engineering goal of Sentinel. The project uses:

* Go benchmarks
* Allocation measurements
* End-to-end load testing
* CPU profiling
* Memory profiling
* `pprof`
* Race detection
* Concurrency testing

The philosophy is:


$$\text{Measure} \longrightarrow \text{Profile} \longrightarrow \text{Identify Bottleneck} \longrightarrow \text{Optimize} \longrightarrow \text{Benchmark Again}$$

### Router Benchmark

An internal router benchmark on an Apple M2 produced approximately:

* **~22 ns/op**
* **0 B/op**
* **0 allocs/op**

Example benchmark output:

```text
BenchmarkHashMapRouter-8    53357709    22.16 ns/op    0 B/op    0 allocs/op

```

*These numbers represent an isolated routing component and should not be interpreted as complete HTTP request latency.*

### Rate Limiter Benchmark

The token bucket benchmark produced approximately:

* **~60 ns/op**
* **0 B/op**
* **0 allocs/op**

A concurrent benchmark produced approximately:

* **~140 ns/op**

*These measurements represent the rate limiter itself rather than the complete Data Plane.*

### End-to-End Data Plane Benchmark

A complete Data Plane benchmark using a real local HTTP backend produced approximately:

* **~35.8 µs/op**

*This includes HTTP proxying and transport overhead. It therefore should not be directly compared with the nanosecond-level internal component benchmarks. The purpose of this benchmark is to measure the cost of the complete request path.*

### Load Test

A `wrk` load test against a local backend produced:

* **8 threads**
* **100 concurrent connections**
* **30 seconds**
* **~52K requests/sec**
* **1.98 ms average latency**
* **1.61 ms median latency**
* **8.30 ms p99 latency**
* **~1.55M requests**

*Measured on an Apple M2. Exact results vary with hardware, operating system, backend behavior, connection count, and system load.*

---

## Profiling

Sentinel exposes Go `pprof` endpoints during development.

The profiling server runs on:
`localhost:6060`

CPU profiling:

```bash
go tool pprof http://localhost:6060/debug/pprof/profile

```

Heap profiling:

```bash
go tool pprof http://localhost:6060/debug/pprof/heap

```

Profiling is used to identify actual bottlenecks before making performance changes.

---

## Testing

Sentinel includes:

* Unit tests
* Integration tests
* Concurrency tests
* Race-detector tests
* Benchmarks
* Control Plane/Data Plane tests

Run all tests:

```bash
go test ./...

```

Run race detection:

```bash
go test -race ./...

```

Run benchmarks:

```bash
go test -bench=. -benchmem ./...

```

Run router benchmarks:

```bash
go test -bench=. -benchmem ./internal/router

```

---

## Concurrency

Sentinel is designed for concurrent workloads.

Concurrency-sensitive components include:

* Router
* Load balancer
* Rate limiter
* Circuit breaker
* Health checker
* Retry engine
* Runtime configuration
* Control Plane subscriptions

Atomic operations and synchronization primitives are used where appropriate. Race detection is part of the development workflow:

```bash
go test -race ./...

```

---

## Design Principles

### 1. Measure Before Optimizing

Performance decisions should be supported by:

* Benchmarks
* Profiling
* Allocation measurements
* Load testing
* Concurrency testing

### 2. Keep the Hot Path Simple

The request path should minimize unnecessary:

* Locks
* Allocations
* Memory copies
* Coordination
* External dependencies

### 3. Correctness Before Micro-Optimization

Performance improvements must preserve:

* Routing correctness
* Rate-limit behavior
* Load-balancing behavior
* Retry semantics
* Circuit-breaker semantics
* Request-body safety
* Concurrency safety

### 4. Clear Component Boundaries

Routing, rate limiting, load balancing, health checking, circuit breaking, retries, and proxying are implemented as separate components.

This makes each subsystem easier to:

* Test
* Benchmark
* Profile
* Replace
* Optimize

### 5. No Premature Distributed Complexity

Sentinel is being built incrementally.

The approach is:


$$\text{Build} \longrightarrow \text{Test} \longrightarrow \text{Benchmark} \longrightarrow \text{Profile} \longrightarrow \text{Optimize} \longrightarrow \text{Scale}$$

Distributed functionality is introduced only after the underlying Data Plane is understood and measurable.

---

## Project Structure

```text
sentinel/
│
├── cmd/
│   ├── sentinel/
│   │   └── main.go
│   │
│   └── sentinel-control/
│       └── main.go
│
├── internal/
│   ├── config/
│   ├── core/
│   ├── dataplane/
│   ├── controlplane/
│   ├── router/
│   ├── ratelimiter/
│   ├── lb/
│   ├── health/
│   ├── circuitbreaker/
│   ├── retry/
│   └── proxy/
│
├── proto/
│   ├── control.proto
│   ├── control.pb.go
│   └── control_grpc.pb.go
│
├── configs/
│
├── benchmarks/
│
├── go.mod
├── go.sum
├── Makefile
└── README.md

```

---

## Current Status

| Component | Status |
| --- | --- |
| Project Foundation | ✅ Complete |
| Configuration | ✅ Complete |
| Radix Router | ✅ Complete |
| Load Balancing | ✅ Complete |
| Health Checking | ✅ Complete |
| Circuit Breaker | ✅ Complete |
| Retry Engine | ✅ Complete |
| Exponential Backoff | ✅ Complete |
| Full Jitter | ✅ Complete |
| Request Body Replay Safety | ✅ Complete |
| Rate Limiter V1 | ✅ Complete |
| Data Plane Integration | ✅ Complete |
| Control Plane | ✅ Complete |
| gRPC Streaming | ✅ Complete |
| SQLite Persistence | ✅ Complete |
| Dynamic Runtime Configuration | ✅ Complete |
| Automatic Reconnection | ✅ Complete |
| Graceful Shutdown | ✅ Complete |
| Health / Readiness | ✅ Complete |
| Unit Tests | ✅ Complete |
| Integration Tests | ✅ Complete |
| Race Detection | ✅ Complete |
| Benchmarking | ✅ Complete |
| Initial Profiling | ✅ Complete |

### Current Focus

Sentinel is currently **feature-complete for the current architecture**.

The next phase is **Data Plane performance engineering**.

Planned optimization work includes:

* CPU profiling
* Memory profiling
* Allocation analysis
* HTTP transport optimization
* Reverse-proxy optimization
* Router optimization
* Rate limiter overhead analysis
* Load-balancer strategy benchmarking
* GC pressure analysis
* Concurrency tuning
* End-to-end latency optimization

The goal is to improve performance based on measured bottlenecks rather than premature micro-optimization.

---

## Future Work

Potential future areas include:

```text
Performance Optimization
        │
        ▼
Advanced Observability
        │
        ├── Metrics
        └── Tracing
        │
        ▼
Service Discovery
        │
        ▼
Distributed Rate Limiting
        │
        ▼
Advanced Control Plane
        │
        ▼
Configuration Distribution
        │
        ▼
Larger-Scale Distributed Edge

```

*These are future directions rather than requirements for the current milestone.*

---

## Technology Stack

* **Go**
* **net/http**
* **gRPC**
* **Protocol Buffers**
* **SQLite**
* **YAML**
* Atomic operations
* Go concurrency primitives
* Go benchmarking
* Go race detector
* Go `pprof`
* Docker for development/testing

---

## Why Sentinel?

Sentinel is an engineering project focused on understanding how high-performance edge infrastructure works internally.

It explores:

* HTTP networking
* Routing
* Load balancing
* Rate limiting
* Fault tolerance
* Retry behavior
* Circuit breaking
* Backend health
* Concurrency
* Reverse proxying
* Distributed configuration
* Performance engineering

The project prioritizes understanding the system from the inside out rather than simply assembling existing infrastructure components.

---

## Development Philosophy

Every major feature follows an engineering loop:

$$\text{Design} \longrightarrow \text{Implement} \longrightarrow \text{Test} \longrightarrow \text{Concurrency Test} \longrightarrow \text{Race Detection} \longrightarrow \text{Benchmark} \longrightarrow \text{Profile} \longrightarrow \text{Optimize}$$

The objective is to make Sentinel **correct first, measurable second, and faster through evidence-driven optimization**.

---

## Status

* **Sentinel:** Feature Complete
* **Current Phase:** Data Plane Performance Optimization
* **Control Plane:** Implemented
* **Next Goal:** Profile $\rightarrow$ Identify Bottlenecks $\rightarrow$ Optimize $\rightarrow$ Benchmark Again

---

## License

License information will be added as the project is finalized.
