package cache

import "fmt"

// IPWhitelistActiveKey is the cross-service cache key for active IP whitelist entries per merchant.
// Shared by api-gateway, merchant-service, admin-merchant-service, and production services.
// Format: merchant:ipwl:active:{merchantID}
func IPWhitelistActiveKey(merchantID uint64) string {
	return Key("merchant", "ipwl", fmt.Sprintf("active:%d", merchantID))
}

// IPWhitelistAllKey caches all IP whitelist rows for a merchant (admin list reads).
// Format: merchant:ipwl:all:{merchantID}
func IPWhitelistAllKey(merchantID uint64) string {
	return Key("merchant", "ipwl", fmt.Sprintf("all:%d", merchantID))
}

// IPWhitelistKeys returns all cache keys to invalidate when whitelist data changes.
func IPWhitelistKeys(merchantID uint64) []string {
	return []string{IPWhitelistActiveKey(merchantID), IPWhitelistAllKey(merchantID)}
}
