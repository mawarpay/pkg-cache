package cache

import (
	"context"
	"net/http"
)

// HealthStatus describes Redis connectivity for health endpoints.
type HealthStatus struct {
	Enabled   bool   `json:"enabled"`
	Available bool   `json:"available"`
	Error     string `json:"error,omitempty"`
}

// CheckHealth returns Redis health for inclusion in service health responses.
func CheckHealth(ctx context.Context) HealthStatus {
	if !Enabled() {
		return HealthStatus{Enabled: false, Available: false}
	}
	if err := Ping(ctx); err != nil {
		return HealthStatus{Enabled: true, Available: false, Error: err.Error()}
	}
	return HealthStatus{Enabled: true, Available: true}
}

// WriteHealthResponse sets HTTP status based on Redis health when Redis is required.
func WriteHealthResponse(w http.ResponseWriter, status HealthStatus, redisRequired bool) int {
	if !status.Enabled {
		if redisRequired {
			return http.StatusServiceUnavailable
		}
		return http.StatusOK
	}
	if !status.Available {
		return http.StatusServiceUnavailable
	}
	return http.StatusOK
}
