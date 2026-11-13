package config

import "testing"

// Relay HANYA via 9Router: tidak ada model tetap / fallback.
// Key wajib diisi (CHAT_PROVIDER_API_KEY), base default ke 9Router internal,
// daftar model diambil otomatis dari 9Router (GET /v1/models).
func TestProviderNineRouterOnly(t *testing.T) {
	c := &Config{}
	if got := c.ProviderBaseURL(); got != DefaultProviderBaseURL {
		t.Fatalf("base default = %q, want %q", got, DefaultProviderBaseURL)
	}
	if c.ChatRelayReady() {
		t.Fatal("empty key must not be ready")
	}

	c.ChatProviderBaseURL = "http://nine-router:20128/v1/"
	c.ChatProviderAPIKey = "nk"
	if got := c.ProviderBaseURL(); got != "http://nine-router:20128/v1" {
		t.Fatalf("trailing slash not trimmed: %q", got)
	}
	if got := c.ProviderAPIKey(); got != "nk" {
		t.Fatalf("key = %q", got)
	}
	if !c.ChatRelayReady() {
		t.Fatal("key set must be ready")
	}
}
