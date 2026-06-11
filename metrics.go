package cache

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const metricsNamespace = "ilonapay_cache"

var (
	cacheHitTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Name:      "cache_hit_total",
			Help:      "Total cache hits.",
		},
		[]string{"service", "entity"},
	)

	cacheMissTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Name:      "cache_miss_total",
			Help:      "Total cache misses.",
		},
		[]string{"service", "entity"},
	)

	cacheSetTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Name:      "cache_set_total",
			Help:      "Total cache set operations.",
		},
		[]string{"service", "entity"},
	)

	cacheDeleteTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Name:      "cache_delete_total",
			Help:      "Total cache delete operations.",
		},
		[]string{"service", "entity"},
	)

	redisErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Name:      "redis_errors_total",
			Help:      "Total Redis operation errors.",
		},
		[]string{"service", "operation"},
	)

	redisLatencyMs = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Name:      "redis_latency_ms",
			Help:      "Redis operation latency in milliseconds.",
			Buckets:   []float64{0.5, 1, 2, 5, 10, 25, 50, 100, 250, 500, 1000},
		},
		[]string{"service", "operation"},
	)
)

func incHit(service, entity string) {
	cacheHitTotal.WithLabelValues(service, entity).Inc()
}

func incMiss(service, entity string) {
	cacheMissTotal.WithLabelValues(service, entity).Inc()
}

func incSet(service, entity string) {
	cacheSetTotal.WithLabelValues(service, entity).Inc()
}

func incDelete(service, entity string) {
	cacheDeleteTotal.WithLabelValues(service, entity).Inc()
}

func incRedisError(service, operation string) {
	redisErrorsTotal.WithLabelValues(service, operation).Inc()
}

func observeLatency(service, operation string, started time.Time) {
	redisLatencyMs.WithLabelValues(service, operation).Observe(float64(time.Since(started).Milliseconds()))
}
