package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                  string
	DBHost                string
	DBPort                string
	DBUser                string
	DBPassword            string
	DBName                string
	DBSSLMode             string
	JWTSecret             string
	JWTExpiry             string
	RefreshExpiryAdmin    string
	RefreshExpiryCustomer string
	GoogleAndroidClientID string
	GoogleIOSClientID     string
	BootstrapSecret       string
	FrontendURL           string
	CookieSecure          bool
	// AI chat relay (POST /api/chat) — HANYA via 9Router (OpenAI-compatible
	// /v1/chat/completions). Server holds the provider key so the APK never
	// ships one. Empty ChatProviderAPIKey = relay disabled (503).
	// Tidak ada fallback langsung ke Gemini: semua chat wajib lewat gateway.
	// Model TIDAK di-pin di backend: daftar model diambil otomatis dari
	// 9Router (GET /v1/models) lalu dipilih round-robin per request, sehingga
	// seluruh model/provider yang tersedia di 9Router ikut terpakai.
	ChatProviderBaseURL string
	ChatProviderAPIKey  string
	ChatTimeoutSec      int
	// Optional shared secret: when set, mobile must send X-App-Secret.
	// Leave empty only in local dev.
	ChatAppSecret string
}

func Load() *Config {
	_ = godotenv.Load()

	get := func(key, def string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return def
	}

	return &Config{
		Port:                  get("PORT", "8080"),
		DBHost:                get("DB_HOST", "localhost"),
		DBPort:                get("DB_PORT", "5432"),
		DBUser:                get("DB_USER", "postgres"),
		DBPassword:            get("DB_PASSWORD", "postgres"),
		DBName:                get("DB_NAME", "admin_daebak_music"),
		DBSSLMode:             get("DB_SSLMODE", "disable"),
		JWTSecret:             get("JWT_SECRET", "change-me-in-production"),
		JWTExpiry:             get("JWT_EXPIRY", "1h"),
		RefreshExpiryAdmin:    get("REFRESH_EXPIRY_ADMIN", "24h"),
		RefreshExpiryCustomer: get("REFRESH_EXPIRY_CUSTOMER", "2160h"),
		GoogleAndroidClientID: get("GOOGLE_ANDROID_CLIENT_ID", ""),
		GoogleIOSClientID:     get("GOOGLE_IOS_CLIENT_ID", ""),
		BootstrapSecret:       get("BOOTSTRAP_SECRET", ""),
		FrontendURL:           get("FRONTEND_URL", "http://localhost:3000"),
		CookieSecure:          get("COOKIE_SECURE", "false") == "true" || os.Getenv("COOKIE_SECURE") == "1",
		ChatProviderBaseURL:   get("CHAT_PROVIDER_BASE_URL", ""),
		ChatProviderAPIKey:    get("CHAT_PROVIDER_API_KEY", ""),
		ChatTimeoutSec:        getInt("CHAT_TIMEOUT_SEC", 45),
		ChatAppSecret:         get("CHAT_APP_SECRET", ""),
	}
}

func getInt(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	var n int
	if _, err := fmt.Sscanf(v, "%d", &n); err != nil || n < 0 {
		return def
	}
	return n
}

func (c *Config) DSN() string {
	return "host=" + c.DBHost + " port=" + c.DBPort + " user=" + c.DBUser +
		" password=" + c.DBPassword + " dbname=" + c.DBName + " sslmode=" + c.DBSSLMode
}

// DefaultProviderBaseURL adalah gateway 9Router internal (compose service
// nine-router). Semua chat relay wajib lewat sini; tidak ada jalur langsung
// ke provider (mis. Gemini) dari backend.
const DefaultProviderBaseURL = "http://nine-router:20128/v1"

// ProviderBaseURL adalah base URL OpenAI-compatible tujuan relay
// (default: 9Router internal, tanpa trailing slash).
func (c *Config) ProviderBaseURL() string {
	if v := strings.TrimSpace(c.ChatProviderBaseURL); v != "" {
		return strings.TrimRight(v, "/")
	}
	return DefaultProviderBaseURL
}

// ProviderAPIKey adalah key /v1 milik 9Router (dibuat di UI 9Router).
func (c *Config) ProviderAPIKey() string {
	return strings.TrimSpace(c.ChatProviderAPIKey)
}

// ChatRelayReady true bila relay bisa jalan: key 9Router terisi.
// Base URL selalu ada (default ke 9Router internal di atas); daftar model
// diambil otomatis dari 9Router, bukan dari env.
func (c *Config) ChatRelayReady() bool {
	return c.ProviderAPIKey() != ""
}
