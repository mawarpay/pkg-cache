# github.com/mawarpay/pkg-cache

Shared Redis cache-aside library for IlonaPay microservices. It provides
namespaced key helpers, cache-aside patterns, negative caching and TTL
management, instrumentation (Prometheus + OpenTelemetry), and small
convenience helpers for service health and session validation.

## Goals

- Provide a lightweight, well-instrumented Redis-backed cache-aside library
  usable across IlonaPay microservices.
- Offer safe production defaults: negative caching for not-found results,
  TTL jitter to reduce avalanches, singleflight to deduplicate concurrent
  loads, and optional distributed locking for stampede prevention.
- Expose Prometheus metrics and OpenTelemetry spans so services can observe
  cache behaviour with minimal setup.
- Keep the public API small and testable; avoid embedding business logic.

### Stack
- **Language(s):** Go (100%)
- **Framework / runtime:** Go modules (go.mod)
- **Notable libraries:**
  - github.com/turahe/pkg/redis - shared Redis client/adapter used by the package
  - github.com/prometheus/client_golang - Prometheus metrics
  - go.opentelemetry.io/otel - OpenTelemetry tracing helpers
  - github.com/cenkalti/backoff/v5 - retry/backoff used for Redis setup
  - golang.org/x/sync/singleflight - request coalescing for load deduplication

## Features

- Cache-aside (`LoadJSON`) with hit/miss Prometheus metrics
- Singleflight request coalescing + Redis distributed lock (stampede prevention)
- Negative caching (`__NOT_FOUND__`, configurable TTL)
- TTL jitter 10–20% (avalanche prevention)
- Batch `MGetJSON` / `MSetJSON` via pipeline
- Prometheus metrics (`ilonapay_cache_*`)
- OpenTelemetry spans on GET/SET/DEL/MGET/PIPELINE and DB fallback
- Health check helper and Ping/Close helpers for graceful shutdown
- Auth session validation helpers (`HasUserSession`, `HasAdminSession`) against auth-service Redis keys

## Quick Start

```go
import pkgcache "github.com/mawarpay/pkg-cache"

cfg := pkgcache.LoadConfig()
_ = pkgcache.Setup(ctx, cfg)
defer pkgcache.Close()

store := pkgcache.NewStore("wallet", cfg).WithEntity("uuid")
var wallet Wallet
_ = store.LoadJSON(ctx, pkgcache.Key("wallet", "uuid", id), pkgcache.TTLUserProfile, &wallet, func() (any, error) {
    return repo.FindByUUID(ctx, id)
})
```

## How it's organized

```
.go files at repo root/    library sources and helpers (keys, config, store, metrics, trace)
README.md                 project README and quick start
doc.go                    package-level documentation for pkg.go.dev
config.go                 TTL and environment-driven configuration
client.go                 Redis lifecycle helpers (Setup, Ping, Close)
store.go                  Store type: cache-aside operations and batch helpers
metrics.go                Prometheus metric definitions and increments
trace.go                  OpenTelemetry span helpers (internal)
session.go                auth-service session validation helpers
apikey.go, ipwl.go        key builders for specific domains (API keys, IP whitelist)
middleware.go             Gin helper to register /metrics endpoint
store_test.go, *_test.go  unit tests and benchmarks
```

**How it fits together:** Services call `NewStore(service, cfg)` to get a
Store scoped to a service namespace. Consumers use `LoadJSON` to perform
cache-aside reads that will return cached JSON or call a provided loader
function to fetch from the backing store and populate Redis. Instrumentation
and health helpers are provided at the package level so services can expose
metrics and health endpoints with minimal glue code.

## Contributing / Running tests

```bash
make help          # list targets
make check         # fmt-check + vet + build + unit tests
make lint          # golangci-lint in Docker (v2.13.0)
make test          # go test ./... -race with coverage
make docker-up     # start Redis (:6379)
make docker-test   # build image and run go test ./... against Compose Redis
make docker-down   # stop Compose services
```

Without Make: `gofmt -w .`, `go vet ./...`, `go test ./...`.


