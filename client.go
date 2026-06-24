package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/turahe/pkg/config"
	pkgredis "github.com/turahe/pkg/redis"
)

// Enabled reports whether Redis caching is active.
func Enabled() bool {
	return config.GetConfig().Redis.Enabled && pkgredis.IsAlive()
}

// Setup initializes Redis via turahe/pkg/redis with exponential backoff retry.
// Returns nil when Redis is disabled; returns error only when enabled and all retries fail.
func Setup(ctx context.Context, cfg Config) error {
	if !config.GetConfig().Redis.Enabled {
		return nil
	}

	applyPoolConfig(cfg)

	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = 200 * time.Millisecond
	bo.MaxInterval = 5 * time.Second

	_, err := backoff.Retry(ctx, func() (struct{}, error) {
		if err := pkgredis.Setup(); err != nil {
			return struct{}{}, err
		}
		if !pkgredis.IsAlive() {
			return struct{}{}, fmt.Errorf("redis ping failed")
		}
		return struct{}{}, nil
	}, backoff.WithBackOff(bo), backoff.WithMaxTries(uint(cfg.MaxRetries+1)))
	return err
}

func applyPoolConfig(cfg Config) {
	if cfg.PoolSize <= 0 {
		return
	}
	config.GetConfig().Redis.PoolSize = cfg.PoolSize
}

// Ping checks Redis connectivity for health endpoints.
func Ping(ctx context.Context) error {
	if !Enabled() {
		return fmt.Errorf("redis disabled or unavailable")
	}
	start := time.Now()
	err := pkgredis.GetUniversalClient().Ping(ctx).Err()
	observeLatency("global", "ping", start)
	if err != nil {
		incRedisError("global", "ping")
	}
	return err
}

// Close closes the Redis connection pool during graceful shutdown.
func Close() error {
	if !config.GetConfig().Redis.Enabled {
		return nil
	}
	return pkgredis.Close()
}

func delKeys(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	if len(keys) == 1 {
		return pkgredis.Delete(ctx, keys[0])
	}
	return pkgredis.GetUniversalClient().Del(ctx, keys...).Err()
}
