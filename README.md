# Sentinel

**Sentinel** is a high-performance, lightweight edge platform and reverse proxy written in Go.

It is designed around a fast data-plane request path with routing, rate limiting, load balancing, circuit breaking, retries, health checking, and proxying.

The current focus is building a **correct, measurable, and high-performance Data Plane** before moving on to distributed control-plane functionality.

---

## Architecture

The current Sentinel Data Plane follows this request path:

```text
                         Sentinel
                            │
Client ────────────────────▼
                     ┌─────────────┐
                     │    Router   │
                     └──────┬──────┘
                            │
                     ┌──────▼──────┐
                     │ Rate Limiter│
                     └──────┬──────┘
                            │
                     ┌──────▼──────┐
                     │Load Balancer│
                     └──────┬──────┘
                            │
                     ┌──────▼──────┐
                     │   Circuit   │
                     │   Breaker   │
                     └──────┬──────┘
                            │
                     ┌──────▼──────┐
                     │    Retry    │
                     │    Engine   │
                     └──────┬──────┘
                            │
                     ┌──────▼──────┐
                     │    Proxy    │
                     └──────┬──────┘
                            │
                    ┌───────▼───────┐
                    │    Backend    │
                    └───────────────┘
```

---

# Features

## Radix Router

Sentinel uses a radix-tree based router for efficient HTTP route matching.

### Features

- Static routes
- Parameterized routes
- Fast lookup
- Low allocation overhead

Example:

```text
/users
/users/:id
/videos/:id
```

The router is designed to keep the request path lightweight and minimize allocations.

---

## Rate Limiter

Sentinel implements a **per-client Token Bucket rate limiter**.

Each client receives an independent bucket with configurable:

- Capacity
- Refill rate

Example configuration:

```yaml
rate_limit:
  enabled: true
  capacity: 100
  refill_rate: 100
```

### Implementation

The token bucket state is updated using **atomic Compare-And-Swap (CAS)**.

This keeps the token bucket's hot path lock-free.

A `sync.Mutex` protects the per-client bucket map during bucket lookup and creation.

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

Requests that exceed the configured limit receive:

```text
HTTP 429 Too Many Requests
```

### Current limitation

The current rate limiter is **instance-local**.

For example:

```text
                    Load Balancer
                   /              \
                  /                \
                 ▼                  ▼
          Sentinel A           Sentinel B
          Bucket A             Bucket B
```

Both Sentinel instances maintain independent token state.

Therefore, if multiple Sentinel instances are running behind a load balancer, the configured limit is not currently a globally shared limit.

Distributed rate limiting using shared state or coordination is planned for a future stage.

---

## Load Balancer

Sentinel currently uses **Round Robin** backend selection.

```text
Request 1 → Backend A
Request 2 → Backend B
Request 3 → Backend C
Request 4 → Backend A
Request 5 → Backend B
```

The implementation uses atomic state for concurrent backend selection.

### Why Round Robin?

Round Robin was selected initially to establish a simple, predictable, low-overhead baseline.

It provides:

- O(1) backend selection
- Minimal state
- Low coordination overhead
- Good concurrency characteristics
- A clean performance baseline

More advanced strategies such as **Least Connections** can be evaluated during the Data Plane optimization phase.

Least Connections may perform better when backend request durations or workloads are uneven, but it also requires maintaining and evaluating active-connection state.

The goal is to benchmark these trade-offs rather than assume one strategy is universally better.

---

## Health Checker

Sentinel continuously checks backend health.

The health checker:

1. Performs HTTP health checks.
2. Falls back to TCP connectivity when appropriate.
3. Tracks backend health.
4. Prevents unhealthy backends from receiving traffic.

Health checking operates independently from the request retry mechanism.

```text
              ┌───────────────┐
              │ Health Checker│
              └───────┬───────┘
                      │
             ┌────────┼────────┐
             ▼        ▼        ▼
          Backend A Backend B Backend C
             │        │        │
             ▼        ▼        ▼
           Healthy  Unhealthy Healthy
```

---

## Circuit Breaker

Sentinel implements a circuit breaker to prevent repeatedly sending traffic to failing backends.

The circuit breaker has three states:

```text
                 Failure Threshold
                       │
                       ▼
                  ┌────────┐
             ┌───►│ CLOSED │
             │    └────┬───┘
             │         │
             │         │ failures
             │         ▼
             │    ┌────────┐
             │    │  OPEN  │
             │    └────┬───┘
             │         │
             │         │ timeout
             │         ▼
             │   ┌───────────┐
             └───│ HALF-OPEN │
                 └─────┬─────┘
                       │
                       │ successful probe
                       ▼
                    CLOSED
```

### States

- **CLOSED** — normal traffic flows.
- **OPEN** — requests are blocked from the failing backend.
- **HALF-OPEN** — a limited recovery probe is allowed.

The circuit breaker supports:

- Failure thresholds
- Recovery timeout
- Half-open state
- Single concurrent recovery probe

The circuit breaker records the outcome of the **logical request**, rather than treating every retry attempt as a separate request.

---

## Retry Engine

Sentinel supports configurable retries for transient upstream failures.

### Retryable HTTP Status Codes

```text
502 Bad Gateway
503 Service Unavailable
504 Gateway Timeout
```

### Retryable Methods

By default:

```text
GET
HEAD
OPTIONS
PUT
DELETE
```

Methods such as:

```text
POST
PATCH
```

are not retried by default.

This avoids automatically replaying potentially non-idempotent operations.

### Network Errors

Retryable network errors are also supported when the request method is eligible for retry.

---

## Retry + Circuit Breaker

Retries and the circuit breaker work together.

The request flow is:

```text
Circuit Breaker
      │
      ▼
 Retry Engine
      │
      ▼
   Backend
```

A single logical request may result in multiple upstream attempts.

The circuit breaker evaluates the overall logical request outcome instead of counting every retry attempt independently.

Retries remain associated with the selected backend rather than selecting a different backend for every attempt.

---

## Exponential Backoff

Retries use exponential backoff with **Full Jitter**.

Conceptually:

```text
Attempt 1 → random delay within Base
Attempt 2 → random delay within 2 × Base
Attempt 3 → random delay within 4 × Base
Attempt 4 → random delay within 8 × Base
...
```

The delay is capped by a configured maximum.

Full jitter helps prevent multiple clients from retrying simultaneously and creating another load spike.

---

## Request Body Replay Safety

Sentinel does not blindly retry requests whose request bodies have already been consumed.

A request is retried only when its body can safely be replayed.

This prevents retry attempts from receiving:

- Empty bodies
- Partially consumed bodies
- Invalid request payloads

This is particularly important when implementing retries around HTTP request streams.

---

# Data Plane

The Data Plane is the primary completed part of Sentinel.

Its responsibility is to process live traffic as efficiently and reliably as possible.

The current request lifecycle is:

```text
Incoming Request
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
     Proxy
       │
       ▼
   Upstream
```

Each component has a focused responsibility.

---

# Configuration

Sentinel currently uses YAML configuration.

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

Configuration is loaded at startup and used to initialize the Data Plane components.

---

# Performance

Performance is a core design goal of Sentinel.

The project uses:

- Go benchmarks
- Allocation measurements
- Race detection
- End-to-end benchmarks
- CPU profiling
- Memory profiling
- `pprof`

The goal is to measure before optimizing.

---

## Router Benchmark

One of the router benchmarks produced approximately:

```text
BenchmarkHashMapRouter-8
53,357,709        22.16 ns/op
0 B/op
0 allocs/op
```

This was measured on an Apple M2.

Exact numbers will vary depending on hardware and benchmark conditions.

---

## Rate Limiter Benchmark

The current token bucket benchmark on an Apple M2 produced approximately:

```text
~60 ns/op
0 B/op
0 allocs/op
```

A concurrent benchmark produced approximately:

```text
~140 ns/op
```

These measurements represent the limiter component rather than the entire HTTP request path.

---

## Full Data Plane Benchmark

A full Data Plane benchmark using a real local HTTP backend produced approximately:

```text
~35.8 µs/op
```

This benchmark includes HTTP proxying and transport overhead.

Therefore, it should not be directly compared with the nanosecond-level internal component benchmarks.

The purpose of the end-to-end benchmark is to understand the cost of the complete request path.

---

# Profiling

Sentinel exposes Go's `pprof` endpoints during development.

The profiling server runs on:

```text
localhost:6060
```

Example CPU profiling command:

```bash
go tool pprof http://localhost:6060/debug/pprof/profile
```

Profiling will be used to identify actual bottlenecks before making performance changes.

---

# Testing

Sentinel includes:

- Unit tests
- Integration tests
- Concurrent tests
- Race-detector tests
- Benchmarks

## Run all tests

```bash
go test ./...
```

## Run race detector

```bash
go test -race ./...
```

## Run all benchmarks

```bash
go test -bench=. -benchmem ./...
```

## Run a specific package benchmark

```bash
go test -bench=. -benchmem ./internal/router
```

---

# Concurrency

Sentinel is designed to operate safely under concurrent workloads.

Concurrency-sensitive components include:

- Router
- Load balancer
- Rate limiter
- Circuit breaker
- Health checker
- Retry engine

The project uses Go's concurrency primitives and atomic operations where appropriate.

Race detection is part of the development workflow:

```bash
go test -race ./...
```

---

# Design Principles

## 1. Measure Before Optimizing

Performance decisions should be backed by:

- Benchmarks
- Profiling
- Allocation measurements
- Concurrency testing

The goal is to optimize based on actual bottlenecks rather than assumptions.

---

## 2. Keep the Hot Path Simple

The request path should avoid unnecessary:

- Locks
- Allocations
- Memory copies
- Coordination
- Background work

Where appropriate, atomic operations are used instead of heavier synchronization.

---

## 3. Correctness Before Micro-Optimization

A faster component that produces incorrect traffic behavior is not an optimization.

Every performance improvement must preserve:

- Routing correctness
- Rate-limit behavior
- Load-balancing behavior
- Retry semantics
- Circuit-breaker behavior
- Request-body safety
- Concurrency safety

---

## 4. Components Should Have Clear Responsibilities

Routing, rate limiting, load balancing, retries, circuit breaking, health checking, and proxying are implemented as separate components.

This makes them:

- Easier to test
- Easier to benchmark
- Easier to profile
- Easier to replace or improve

---

# Project Structure

```text
sentinel/
│
├── cmd/
│   ├── sentinel/
│   │   └── main.go
│   │
│   └── sentinel-control/
│
├── internal/
│   ├── config/
│   │
│   ├── core/
│   │
│   ├── dataplane/
│   │
│   ├── router/
│   │
│   ├── ratelimiter/
│   │
│   ├── lb/
│   │
│   ├── health/
│   │
│   ├── circuitbreaker/
│   │
│   ├── retry/
│   │
│   └── proxy/
│
├── configs/
│
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

The internal package structure may evolve as the project is optimized.

---

# Current Data Plane Status

| Component | Status |
|---|---|
| Project Foundation | ✅ Complete |
| Configuration | ✅ Complete |
| Radix Router | ✅ Complete |
| Load Balancer | ✅ Complete |
| Health Checker | ✅ Complete |
| Circuit Breaker | ✅ Complete |
| Retry Engine | ✅ Complete |
| Exponential Backoff | ✅ Complete |
| Full Jitter | ✅ Complete |
| Request Body Replay Safety | ✅ Complete |
| Rate Limiter V1 | ✅ Complete |
| Data Plane Integration | ✅ Complete |
| Unit Tests | ✅ Complete |
| Integration Tests | ✅ Complete |
| Concurrent Tests | ✅ Complete |
| Race Detection | ✅ Complete |
| Initial Benchmarks | ✅ Complete |
| Data Plane | ✅ Complete |

---

# Current Focus

The Data Plane is functionally complete for the current milestone.

The next phase is:

## Data Plane Optimization

Planned areas include:

- Router profiling
- Router optimization
- Load-balancer benchmarking
- Round Robin vs Least Connections
- Rate limiter overhead
- Allocation reduction
- GC pressure
- HTTP proxy performance
- Concurrency behavior
- End-to-end latency
- CPU profiling
- Memory profiling
- Benchmark-driven optimization

The goal is to improve Sentinel's performance while preserving correctness and maintainability.

---

# Control Plane

The Control Plane is **not currently implemented**.

It is intentionally postponed until the Data Plane has been optimized and benchmarked.

The future Control Plane is expected to handle management and configuration rather than live request processing.

Potential responsibilities include:

```text
Configuration Management
        │
        ▼
Route Management
        │
        ▼
Backend Management
        │
        ▼
Node Management
        │
        ▼
Configuration Distribution
```

Future work may include:

- Centralized configuration
- Route management
- Backend management
- Node registration
- Configuration distribution
- Persistent configuration
- Dynamic configuration
- Distributed control

These features will be implemented in later stages.

---

# Observability

Advanced observability is intentionally separated from the current Data Plane milestone.

Future observability work may include:

- Metrics
- Request metrics
- Backend metrics
- Rate-limit metrics
- Circuit-breaker metrics
- Retry metrics
- Tracing
- Distributed tracing
- Advanced profiling

The current priority is to establish a strong performance baseline before adding additional instrumentation.

---

# Roadmap

```text
                    SENTINEL
                       │
                       ▼
              ┌─────────────────┐
              │   DATA PLANE    │
              └────────┬────────┘
                       │
          ┌────────────┼────────────┐
          ▼            ▼            ▼
       Routing     Traffic      Fault
                   Control     Tolerance
          │            │            │
          ▼            ▼            ▼
       Router      Rate Limit   Circuit Breaker
                   Load Balance    Retry
                                  Health
                       │
                       ▼
              ┌─────────────────┐
              │   OPTIMIZATION  │
              └────────┬────────┘
                       │
          ┌────────────┼────────────┐
          ▼            ▼            ▼
      Benchmark     Profiling   Allocations
          │            │            │
          └────────────┼────────────┘
                       ▼
              Performance Tuning
                       │
                       ▼
              ┌─────────────────┐
              │   OBSERVABILITY │
              └────────┬────────┘
                       │
               ┌───────┴───────┐
               ▼               ▼
            Metrics         Tracing
               │               │
               └───────┬───────┘
                       ▼
              ┌─────────────────┐
              │  CONTROL PLANE  │
              └────────┬────────┘
                       │
          ┌────────────┼────────────┐
          ▼            ▼            ▼
       Routes       Backends      Nodes
          │            │            │
          └────────────┼────────────┘
                       ▼
              Config Distribution
                       │
                       ▼
                Distributed Edge
```

### Completed

```text
[x] Project foundation
[x] Configuration
[x] Radix Router
[x] Load Balancer
[x] Health Checker
[x] Circuit Breaker
[x] Retry Engine
[x] Exponential Backoff
[x] Full Jitter
[x] Request Body Replay Safety
[x] Rate Limiter V1
[x] Data Plane Integration
[x] Testing
[x] Race Detection
[x] Initial Benchmarking
```

### Current

```text
[ ] Data Plane Optimization
[ ] Performance Profiling
[ ] Allocation Optimization
[ ] Load Balancer Strategy Benchmarking
[ ] End-to-End Optimization
```

### Future

```text
[ ] Metrics
[ ] Tracing
[ ] Service Discovery
[ ] Dynamic Configuration
[ ] Advanced Observability
[ ] Control Plane
[ ] Route Management
[ ] Backend Management
[ ] Node Management
[ ] Configuration Distribution
[ ] Persistent Configuration
[ ] Distributed Rate Limiting
```

---

# Technology Stack

- **Go**
- `net/http`
- YAML
- Atomic operations
- Go concurrency primitives
- Go benchmarking
- Go race detector
- Go `pprof`
- Docker for development/testing where required

---

# Why Sentinel?

Sentinel is an engineering-focused project for exploring how high-performance edge infrastructure and traffic-management systems work internally.

The project focuses on:

- High-performance HTTP routing
- Rate limiting
- Load balancing
- Backend health
- Circuit breaking
- Retry behavior
- Fault tolerance
- Concurrency
- Performance engineering
- Distributed system architecture

Rather than introducing distributed complexity immediately, Sentinel is being built incrementally.

The current approach is:

```text
Build
  ↓
Test
  ↓
Benchmark
  ↓
Profile
  ↓
Optimize
  ↓
Scale
```

The goal is to understand the performance and correctness characteristics of each layer before moving to the next stage.

---

# Development Philosophy

Sentinel is intentionally being built incrementally.

Each major feature follows the same general process:

```text
Design
  ↓
Implement
  ↓
Test
  ↓
Concurrency Test
  ↓
Race Detection
  ↓
Benchmark
  ↓
Profile
  ↓
Optimize
```

The project prioritizes **measurable engineering decisions** over prematurely adding complexity.

---

# Status

**Sentinel Data Plane: Functionally Complete**

**Current Phase: Data Plane Optimization**

**Control Plane: Planned**

---

# License

License information will be added as the project is finalized.
