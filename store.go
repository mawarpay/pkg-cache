package cache

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

// Store provides cache-aside operations for a service namespace.
type Store struct {
	service string
	entity  string
	cfg     Config
	group   singleflight.Group
}

// NewStore creates a cache store scoped to service and default entity label for metrics.
func NewStore(service string, cfg Config) *Store {
	return &Store{service: service, entity: "default", cfg: cfg}
}

// WithEntity returns a copy scoped to a specific entity for metrics labeling.
func (s *Store) WithEntity(entity string) *Store {
	cp := *s
	cp.entity = entity
	return &cp
}

// LoadJSON implements cache-aside with singleflight, negative caching, and TTL jitter.
func (s *Store) LoadJSON(ctx context.Context, key string, category TTLCategory, dest any, load func() (any, error)) error {
	if !Enabled() {
		return s.loadDirect(ctx, key, dest, load)
	}

	entity := s.entity
	if hit, err := s.getJSON(ctx, key, dest, entity); err != nil {
		return err
	} else if hit {
		incHit(s.service, entity)
		return nil
	}

	val, err, _ := s.group.Do(key, func() (any, error) {
		if hit, err := s.getJSON(ctx, key, dest, entity); err != nil {
			return nil, err
		} else if hit {
			incHit(s.service, entity)
			return nil, nil
		}

		lockKey := key + ":lock"
		acquired, lockErr := pkgAcquireLock(ctx, lockKey, 5*time.Second)
		if lockErr == nil && !acquired {
			time.Sleep(10 * time.Millisecond)
			if hit, err := s.getJSON(ctx, key, dest, entity); err != nil {
				return nil, err
			} else if hit {
				incHit(s.service, entity)
				return nil, nil
			}
		}
		if acquired {
			defer func() { _ = pkgReleaseLock(ctx, lockKey) }()
		}

		incMiss(s.service, entity)
		loaded, err := s.loadWithTrace(ctx, key, load)
		if err != nil {
			return nil, err
		}
		if loaded == nil {
			_ = s.setNotFound(ctx, key, entity)
			return nil, nil
		}
		if err := s.setJSON(ctx, key, loaded, category, entity); err != nil {
			return loaded, nil
		}
		return loaded, nil
	})
	if err != nil {
		return err
	}
	if val == nil {
		return nil
	}
	raw, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, dest)
}

func (s *Store) loadDirect(ctx context.Context, key string, dest any, load func() (any, error)) error {
	ctx, span := startSpan(ctx, "db.query", key)
	defer span.End()
	setSource(span, "database")

	val, err := load()
	recordError(span, err)
	if err != nil {
		return err
	}
	if val == nil {
		return nil
	}
	raw, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, dest)
}

func (s *Store) loadWithTrace(ctx context.Context, key string, load func() (any, error)) (any, error) {
	ctx, span := startSpan(ctx, "db.query", key)
	defer span.End()
	setSource(span, "database")
	setHit(span, false)

	val, err := load()
	recordError(span, err)
	return val, err
}

func (s *Store) getJSON(ctx context.Context, key string, dest any, entity string) (bool, error) {
	ctx, span := startSpan(ctx, "GET", key)
	defer span.End()

	start := time.Now()
	raw, err := Client().Get(ctx, key).Result()
	observeLatency(s.service, "GET", start)

	if err == redis.Nil {
		setHit(span, false)
		return false, nil
	}
	if err != nil {
		incRedisError(s.service, "GET")
		recordError(span, err)
		return false, err
	}

	if raw == notFoundMarker {
		setHit(span, true)
		setSource(span, "cache")
		return true, nil
	}

	if err := json.Unmarshal([]byte(raw), dest); err != nil {
		incRedisError(s.service, "GET")
		recordError(span, err)
		return false, err
	}
	setHit(span, true)
	setSource(span, "cache")
	return true, nil
}

func (s *Store) setJSON(ctx context.Context, key string, val any, category TTLCategory, entity string) error {
	ctx, span := startSpan(ctx, "SET", key)
	defer span.End()

	ttl := jitterTTL(s.cfg.TTL.For(category))
	setTTL(span, ttl.Seconds())

	raw, err := json.Marshal(val)
	if err != nil {
		recordError(span, err)
		return err
	}

	start := time.Now()
	err = Client().Set(ctx, key, raw, ttl).Err()
	observeLatency(s.service, "SET", start)
	if err != nil {
		incRedisError(s.service, "SET")
		recordError(span, err)
		return err
	}
	incSet(s.service, entity)
	return nil
}

func (s *Store) setNotFound(ctx context.Context, key, entity string) error {
	ctx, span := startSpan(ctx, "SET", key)
	defer span.End()

	ttl := jitterTTL(s.cfg.TTL.For(TTLNotFound))
	setTTL(span, ttl.Seconds())

	start := time.Now()
	err := Client().Set(ctx, key, notFoundMarker, ttl).Err()
	observeLatency(s.service, "SET", start)
	if err != nil {
		incRedisError(s.service, "SET")
		recordError(span, err)
		return err
	}
	incSet(s.service, entity)
	return nil
}

// SetJSON stores a value directly (write-through after mutations).
func (s *Store) SetJSON(ctx context.Context, key string, val any, category TTLCategory) error {
	return s.setJSON(ctx, key, val, category, s.entity)
}

// SetJSONWithDuration stores a value with an explicit duration (still applies jitter).
func (s *Store) SetJSONWithDuration(ctx context.Context, key string, val any, ttl time.Duration) error {
	return s.setJSONWithDuration(ctx, key, val, ttl, s.entity)
}

func (s *Store) setJSONWithDuration(ctx context.Context, key string, val any, ttl time.Duration, entity string) error {
	ctx, span := startSpan(ctx, "SET", key)
	defer span.End()

	ttl = jitterTTL(ttl)
	setTTL(span, ttl.Seconds())

	raw, err := json.Marshal(val)
	if err != nil {
		recordError(span, err)
		return err
	}

	start := time.Now()
	err = Client().Set(ctx, key, raw, ttl).Err()
	observeLatency(s.service, "SET", start)
	if err != nil {
		incRedisError(s.service, "SET")
		recordError(span, err)
		return err
	}
	incSet(s.service, entity)
	return nil
}

// GetJSON reads a cached value without loading from DB.
func (s *Store) GetJSON(ctx context.Context, key string, dest any) (bool, error) {
	if !Enabled() {
		return false, nil
	}
	return s.getJSON(ctx, key, dest, s.entity)
}

// Del removes one or more cache keys.
func (s *Store) Del(ctx context.Context, keys ...string) error {
	if !Enabled() || len(keys) == 0 {
		return nil
	}
	ctx, span := startSpan(ctx, "DEL", keys[0])
	defer span.End()

	start := time.Now()
	err := Client().Del(ctx, keys...).Err()
	observeLatency(s.service, "DEL", start)
	if err != nil {
		incRedisError(s.service, "DEL")
		recordError(span, err)
		return err
	}
	incDelete(s.service, s.entity)
	return nil
}

// InvalidatePrefix deletes all keys matching prefix via SCAN + DEL pipeline.
func (s *Store) InvalidatePrefix(ctx context.Context, prefix string) error {
	if !Enabled() {
		return nil
	}
	ctx, span := startSpan(ctx, "DEL", prefix+"*")
	defer span.End()

	iter := Client().Scan(ctx, 0, prefix+"*", 100).Iterator()
	keys := make([]string, 0, 16)
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		incRedisError(s.service, "SCAN")
		recordError(span, err)
		return err
	}
	if len(keys) == 0 {
		return nil
	}

	start := time.Now()
	err := Client().Del(ctx, keys...).Err()
	observeLatency(s.service, "DEL", start)
	if err != nil {
		incRedisError(s.service, "DEL")
		recordError(span, err)
		return err
	}
	incDelete(s.service, s.entity)
	return nil
}

// MGetJSON batch-reads multiple keys. Missing keys are omitted from the result map.
func (s *Store) MGetJSON(ctx context.Context, keys []string, decode func(key, raw string) error) error {
	if !Enabled() || len(keys) == 0 {
		return nil
	}
	ctx, span := startSpan(ctx, "MGET", fmt.Sprintf("%d keys", len(keys)))
	defer span.End()

	start := time.Now()
	vals, err := Client().MGet(ctx, keys...).Result()
	observeLatency(s.service, "MGET", start)
	if err != nil {
		incRedisError(s.service, "MGET")
		recordError(span, err)
		return err
	}

	for i, v := range vals {
		if v == nil {
			incMiss(s.service, s.entity)
			continue
		}
		raw, ok := v.(string)
		if !ok || raw == notFoundMarker {
			incHit(s.service, s.entity)
			continue
		}
		incHit(s.service, s.entity)
		if err := decode(keys[i], raw); err != nil {
			return err
		}
	}
	return nil
}

// MSetJSON batch-writes multiple keys using a pipeline.
func (s *Store) MSetJSON(ctx context.Context, entries map[string]any, category TTLCategory) error {
	if !Enabled() || len(entries) == 0 {
		return nil
	}
	ctx, span := startSpan(ctx, "MSET", fmt.Sprintf("%d keys", len(entries)))
	defer span.End()

	ttl := jitterTTL(s.cfg.TTL.For(category))
	setTTL(span, ttl.Seconds())

	pipe := Client().Pipeline()
	for key, val := range entries {
		raw, err := json.Marshal(val)
		if err != nil {
			recordError(span, err)
			return err
		}
		pipe.Set(ctx, key, raw, ttl)
	}

	start := time.Now()
	_, err := pipe.Exec(ctx)
	observeLatency(s.service, "PIPELINE", start)
	if err != nil {
		incRedisError(s.service, "PIPELINE")
		recordError(span, err)
		return err
	}
	for range entries {
		incSet(s.service, s.entity)
	}
	return nil
}

// jitterTTL adds 10-20% random jitter to prevent cache avalanche.
func jitterTTL(base time.Duration) time.Duration {
	if base <= 0 {
		return base
	}
	maxExtra := int64(base) / 5 // 20% upper bound
	if maxExtra <= 0 {
		return base
	}
	minExtra := maxExtra / 2 // 10% lower bound
	n, err := rand.Int(rand.Reader, big.NewInt(maxExtra-minExtra+1))
	if err != nil {
		return base + time.Duration(minExtra)
	}
	return base + time.Duration(minExtra+n.Int64())
}
