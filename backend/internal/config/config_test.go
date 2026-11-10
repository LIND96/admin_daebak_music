package config

import "testing"

// Provider resolution: explicit CHAT_PROVIDER_* wins, otherwise fall back
// to the legacy GEMINI_* values so existing deploys keep working untouched.
func TestProviderFallback(t *testing.T) {
	c := &Config{
		GeminiAPIKey: "gk", GeminiModel: "gm",
	}
	if got := c.ProviderBaseURL(); got != DefaultProviderBaseURL {
		t.Fatalf("base default = %q", got)
	}
	if got := c.ProviderAPIKey(); got != "gk" {
		t.Fatalf("key fallback = %q", got)
	}
	if got := c.ProviderModel(); got != "gm" {
		t.Fatalf("model fallback = %q", got)
	}

	c.ChatProviderBaseURL = "http://nine-router:20128/v1/"
	c.ChatProviderAPIKey = "nk"
	c.ChatProviderModel = "nm"
	if got := c.ProviderBaseURL(); got != "http://nine-router:20128/v1" {
		t.Fatalf("trailing slash not trimmed: %q", got)
	}
	if got := c.ProviderAPIKey(); got != "nk" {
		t.Fatalf("explicit key = %q", got)
	}
	if got := c.ProviderModel(); got != "nm" {
		t.Fatalf("explicit model = %q", got)
	}
}
