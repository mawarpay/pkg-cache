package cache

import "testing"

func TestIPWhitelistKeys(t *testing.T) {
	got := IPWhitelistActiveKey(42)
	want := "merchant:ipwl:active:42"
	if got != want {
		t.Fatalf("IPWhitelistActiveKey() = %q, want %q", got, want)
	}

	gotAll := IPWhitelistAllKey(42)
	wantAll := "merchant:ipwl:all:42"
	if gotAll != wantAll {
		t.Fatalf("IPWhitelistAllKey() = %q, want %q", gotAll, wantAll)
	}

	keys := IPWhitelistKeys(42)
	if len(keys) != 2 || keys[0] != want || keys[1] != wantAll {
		t.Fatalf("IPWhitelistKeys() = %v", keys)
	}
}
