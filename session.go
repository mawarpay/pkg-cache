package cache

import (
	"context"
	"errors"
	"fmt"

	"github.com/turahe/pkg/config"
	pkgredis "github.com/turahe/pkg/redis"
)

const (
	// SessionModelUser is the model type label for user sessions (value "User").
	SessionModelUser = "User"
	
	// SessionModelAdmin is the model type label for admin sessions (value "Admin").
	SessionModelAdmin = "Admin"
)

// auth-service session keys (shared with auth-service/internal/cache).
const authSessionKeyPrefix = "auth-service:session:"

// ErrRedisRequired is returned when session validation runs while Redis is disabled or unavailable.
var ErrRedisRequired = errors.New("redis is required for session storage (set REDIS_ENABLED=true)")

func authSessionKey(modelType, modelID, deviceID, ip string) string {
	return fmt.Sprintf("%s%s:%s:%s:%s", authSessionKeyPrefix, modelType, modelID, deviceID, ip)
}

// HasSession checks whether a session exists in the shared auth-service session store for the
// given modelType, modelID, deviceID and ip. It returns (true, nil) when a session record is
// present, (false, nil) when no record exists, ErrRedisRequired when Redis is disabled or
// unavailable, and any error returned by the underlying Redis client.
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

// HasUserSession checks whether a portal (user) session exists in the auth-service session store.
// It is a convenience wrapper around HasSession using SessionModelUser.
func HasUserSession(ctx context.Context, userUUID, deviceID, ip string) (bool, error) {
	return HasSession(ctx, SessionModelUser, userUUID, deviceID, ip)
}

// HasAdminSession checks whether an admin session exists in the auth-service session store.
// It is a convenience wrapper around HasSession using SessionModelAdmin.
func HasAdminSession(ctx context.Context, adminUUID, deviceID, ip string) (bool, error) {
	return HasSession(ctx, SessionModelAdmin, adminUUID, deviceID, ip)
}
