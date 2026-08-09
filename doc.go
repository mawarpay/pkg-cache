// Package cache provides helpers and instrumentation for a Redis-backed
// namespaced cache used across services.
//
// The package includes key builders, TTL configuration helpers, Redis setup
// and health helpers, and Prometheus metrics instrumentation. The package
// provides helpers only and does not implement business storage logic.
package cache
