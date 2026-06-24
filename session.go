package cache

import (
	"context"
	"errors"
	"fmt"

	"github.com/turahe/pkg/config"
	pkgredis "github.com/turahe/pkg/redis"
)

const (
	SessionModelUser  = "User"
	SessionModelAdmin = "Admin"
)

// auth-service session keys (shared with auth-service/internal/cache).
const authSessionKeyPrefix = "auth-service:session:"

// ErrRedisRequired is returned when session validation runs while Redis is disabled or unavailable.
var ErrRedisRequired = errors.New("redis is required for session storage (set REDIS_ENABLED=true)")

func authSessionKey(modelType, modelID, deviceID, ip string) string {
	return fmt.Sprintf("%s%s:%s:%s:%s", authSessionKeyPrefix, modelType, modelID, deviceID, ip)
}

// HasSession checks auth-service session validity in Redis.
func HasSession(ctx context.Context, modelType, modelID, deviceID, ip string) (bool, error) {
	if !config.GetConfig().Redis.Enabled {
		return false, ErrRedisRequired
	}
	if !pkgredis.IsAlive() {
		return false, ErrRedisRequired
	}

	key := authSessionKey(modelType, modelID, deviceID, ip)
	raw, err := pkgredis.Get(ctx, key)
	if err != nil {
		return false, err
	}
	return raw != "", nil
}

// HasUserSession checks portal session validity in Redis (auth-service session store).
func HasUserSession(ctx context.Context, userUUID, deviceID, ip string) (bool, error) {
	return HasSession(ctx, SessionModelUser, userUUID, deviceID, ip)
}

// HasAdminSession checks admin session validity in Redis (auth-service session store).
func HasAdminSession(ctx context.Context, adminUUID, deviceID, ip string) (bool, error) {
	return HasSession(ctx, SessionModelAdmin, adminUUID, deviceID, ip)
}
