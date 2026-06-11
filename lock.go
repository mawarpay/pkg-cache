package cache

import (
	"context"
	"time"

	pkgredis "github.com/turahe/pkg/redis"
)

func pkgAcquireLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return pkgredis.AcquireLock(ctx, key, "1", ttl)
}

func pkgReleaseLock(ctx context.Context, key string) error {
	return pkgredis.ReleaseLock(ctx, key)
}
