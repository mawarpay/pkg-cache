package cache

import (
	"testing"
	"time"
)

func TestKey(t *testing.T) {
	got := Key("wallet", "user", "123")
	want := "wallet:user:123"
	if got != want {
		t.Fatalf("Key() = %q, want %q", got, want)
	}
}

func TestPrefix(t *testing.T) {
	got := Prefix("merchant", "id")
	want := "merchant:id:"
	if got != want {
		t.Fatalf("Prefix() = %q, want %q", got, want)
	}
}

func TestTTLConfigFor(t *testing.T) {
	cfg := TTLConfig{
		UserProfile:    5 * time.Minute,
		WalletBalance:  30 * time.Second,
		MerchantConfig: 30 * time.Minute,
		FeatureFlags:   time.Hour,
		Analytics:      15 * time.Minute,
		NotFound:       60 * time.Second,
		Default:        5 * time.Minute,
	}

	tests := []struct {
		category TTLCategory
		want     time.Duration
	}{
		{TTLUserProfile, 5 * time.Minute},
		{TTLWalletBalance, 30 * time.Second},
		{TTLMerchantConfig, 30 * time.Minute},
		{TTLFeatureFlags, time.Hour},
		{TTLAnalytics, 15 * time.Minute},
		{TTLNotFound, 60 * time.Second},
		{TTLDefault, 5 * time.Minute},
		{TTLCategory("unknown"), 5 * time.Minute},
	}

	for _, tt := range tests {
		if got := cfg.For(tt.category); got != tt.want {
			t.Errorf("For(%q) = %v, want %v", tt.category, got, tt.want)
		}
	}
}

func TestJitterTTL(t *testing.T) {
	base := 300 * time.Second
	for i := 0; i < 100; i++ {
		got := jitterTTL(base)
		min := base + base/10
		max := base + base/5
		if got < min || got > max {
			t.Fatalf("jitterTTL(%v) = %v, want between %v and %v", base, got, min, max)
		}
	}
}

func TestJitterTTLZero(t *testing.T) {
	if got := jitterTTL(0); got != 0 {
		t.Fatalf("jitterTTL(0) = %v, want 0", got)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	cfg := LoadConfig()
	if cfg.PoolSize <= 0 {
		t.Fatalf("PoolSize = %d, want > 0", cfg.PoolSize)
	}
	if cfg.MaxRetries < 0 {
		t.Fatalf("MaxRetries = %d, want >= 0", cfg.MaxRetries)
	}
	if cfg.TTL.UserProfile != 5*time.Minute {
		t.Fatalf("UserProfile TTL = %v", cfg.TTL.UserProfile)
	}
	if cfg.TTL.WalletBalance != 30*time.Second {
		t.Fatalf("WalletBalance TTL = %v", cfg.TTL.WalletBalance)
	}
}

func TestStoreLoadJSONDisabled(t *testing.T) {
	store := NewStore("test", LoadConfig())
	type item struct {
		ID int `json:"id"`
	}
	var dest item
	err := store.LoadJSON(t.Context(), "test:key:1", TTLDefault, &dest, func() (any, error) {
		return &item{ID: 42}, nil
	})
	if err != nil {
		t.Fatalf("LoadJSON() error = %v", err)
	}
	if dest.ID != 42 {
		t.Fatalf("dest.ID = %d, want 42", dest.ID)
	}
}

func TestStoreLoadJSONNilResult(t *testing.T) {
	store := NewStore("test", LoadConfig())
	type item struct {
		ID int `json:"id"`
	}
	var dest item
	err := store.LoadJSON(t.Context(), "test:key:2", TTLDefault, &dest, func() (any, error) {
		return nil, nil
	})
	if err != nil {
		t.Fatalf("LoadJSON() error = %v", err)
	}
	if dest.ID != 0 {
		t.Fatalf("dest.ID = %d, want 0 for nil result", dest.ID)
	}
}
