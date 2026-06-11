package cache

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// TTLCategory identifies a data class for TTL selection.
type TTLCategory string

const (
	TTLUserProfile    TTLCategory = "user_profile"
	TTLWalletBalance  TTLCategory = "wallet_balance"
	TTLMerchantConfig TTLCategory = "merchant_config"
	TTLFeatureFlags   TTLCategory = "feature_flags"
	TTLAnalytics      TTLCategory = "analytics"
	TTLNotFound       TTLCategory = "not_found"
	TTLDefault        TTLCategory = "default"
)

const notFoundMarker = "__NOT_FOUND__"

// Config holds Redis connection and cache TTL settings.
type Config struct {
	PoolSize   int
	MaxRetries int
	TTL        TTLConfig
}

// TTLConfig maps categories to durations, overridable via environment variables.
type TTLConfig struct {
	UserProfile    time.Duration
	WalletBalance  time.Duration
	MerchantConfig time.Duration
	FeatureFlags   time.Duration
	Analytics      time.Duration
	NotFound       time.Duration
	Default        time.Duration
}

// LoadConfig reads cache configuration from environment variables.
func LoadConfig() Config {
	return Config{
		PoolSize:   intFromEnv("REDIS_POOL_SIZE", 10),
		MaxRetries: intFromEnv("REDIS_MAX_RETRIES", 3),
		TTL: TTLConfig{
			UserProfile:    durationFromEnv("CACHE_TTL_USER_PROFILE", 5*time.Minute),
			WalletBalance:  durationFromEnv("CACHE_TTL_WALLET_BALANCE", 30*time.Second),
			MerchantConfig: durationFromEnv("CACHE_TTL_MERCHANT_CONFIG", 30*time.Minute),
			FeatureFlags:   durationFromEnv("CACHE_TTL_FEATURE_FLAGS", time.Hour),
			Analytics:      durationFromEnv("CACHE_TTL_ANALYTICS", 15*time.Minute),
			NotFound:         durationFromEnv("CACHE_TTL_NOT_FOUND", 60*time.Second),
			Default:          durationFromEnv("CACHE_TTL_DEFAULT", 5*time.Minute),
		},
	}
}

func (c TTLConfig) For(category TTLCategory) time.Duration {
	switch category {
	case TTLUserProfile:
		return c.UserProfile
	case TTLWalletBalance:
		return c.WalletBalance
	case TTLMerchantConfig:
		return c.MerchantConfig
	case TTLFeatureFlags:
		return c.FeatureFlags
	case TTLAnalytics:
		return c.Analytics
	case TTLNotFound:
		return c.NotFound
	default:
		return c.Default
	}
}

func durationFromEnv(key string, fallback time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	if d, err := time.ParseDuration(v); err == nil && d > 0 {
		return d
	}
	return fallback
}

func intFromEnv(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	if n, err := strconv.Atoi(v); err == nil && n >= 0 {
		return n
	}
	return fallback
}
