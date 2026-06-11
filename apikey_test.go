package cache

import "testing"

func TestApiKeyKeys(t *testing.T) {
	got := ApiKeyByKey("abc123")
	want := "merchant:api_key:abc123"
	if got != want {
		t.Fatalf("ApiKeyByKey() = %q, want %q", got, want)
	}

	gotList := ApiKeyListKey(42)
	wantList := "admin-merchant:api_key:list:42"
	if gotList != wantList {
		t.Fatalf("ApiKeyListKey() = %q, want %q", gotList, wantList)
	}

	keys := ApiKeyKeys("abc123", 42)
	if len(keys) != 2 || keys[0] != want || keys[1] != wantList {
		t.Fatalf("ApiKeyKeys() = %v", keys)
	}
}
