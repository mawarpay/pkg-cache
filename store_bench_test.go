package cache

import (
	"testing"
	"time"
)

func BenchmarkJitterTTL(b *testing.B) {
	base := 300 * time.Second
	for b.Loop() {
		jitterTTL(base)
	}
}

func BenchmarkKey(b *testing.B) {
	for b.Loop() {
		Key("wallet", "balance", "12345")
	}
}

func BenchmarkLoadConfig(b *testing.B) {
	for b.Loop() {
		LoadConfig()
	}
}

func BenchmarkStoreLoadJSONDisabled(b *testing.B) {
	store := NewStore("bench", LoadConfig())
	type item struct {
		ID int `json:"id"`
	}
	b.ReportAllocs()
	for b.Loop() {
		var dest item
		_ = store.LoadJSON(b.Context(), "bench:key:1", TTLDefault, &dest, func() (any, error) {
			return &item{ID: 1}, nil
		})
	}
}
