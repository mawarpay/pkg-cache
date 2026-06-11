# apiservice-cache

Shared Redis cache-aside library for IlonaPay microservices. Built on `github.com/turahe/pkg/redis`.

## Features

- Cache-aside (`LoadJSON`) with hit/miss metrics
- Singleflight request coalescing + Redis distributed lock (stampede prevention)
- Negative caching (`__NOT_FOUND__`, 60s default)
- TTL jitter 10–20% (avalanche prevention)
- Batch `MGetJSON` / `MSetJSON` via pipeline
- Prometheus metrics (`ilonapay_cache_*`)
- OpenTelemetry spans on GET/SET/DEL/MGET/PIPELINE and DB fallback
- Health check helper
- Graceful shutdown via `Close()`

## Quick Start

```go
import pkgcache "github.com/writdev-alt/apiservice-cache"

cfg := pkgcache.LoadConfig()
_ = pkgcache.Setup(ctx, cfg)
defer pkgcache.Close()

store := pkgcache.NewStore("wallet", cfg).WithEntity("uuid")
var wallet Wallet
_ = store.LoadJSON(ctx, pkgcache.Key("wallet", "uuid", id), pkgcache.TTLUserProfile, &wallet, func() (any, error) {
    return repo.FindByUUID(ctx, id)
})
```

## Documentation

- [Architecture](../../docs/redis-caching/architecture.md)
- [Migration Guide](../../docs/redis-caching/migration-guide.md)
- [Performance Report](../../docs/redis-caching/performance-report.md)
