package cache

import "fmt"

// ApiKeyByKey is the cross-service cache key for API key lookup by key string.
// Format: merchant:api_key:{apiKey}
func ApiKeyByKey(apiKey string) string {
	return Key("merchant", "api_key", apiKey)
}

// ApiKeyListKey caches API key list rows per merchant (admin + merchant list reads).
// Format: admin-merchant:api_key:list:{merchantID}
func ApiKeyListKey(merchantID uint64) string {
	return Key("admin-merchant", "api_key", fmt.Sprintf("list:%d", merchantID))
}

// ApiKeyKeys returns all cache keys to invalidate when API key data changes.
func ApiKeyKeys(apiKey string, merchantID uint64) []string {
	return []string{ApiKeyByKey(apiKey), ApiKeyListKey(merchantID)}
}
