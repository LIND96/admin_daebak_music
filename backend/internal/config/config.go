package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	DBSSLMode   string
	JWTSecret   string
	JWTExpiry   string
	RefreshExpiryAdmin    string
	RefreshExpiryCustomer string
	GoogleAndroidClientID string
	GoogleIOSClientID     string
	BootstrapSecret string
	FrontendURL string
	CookieSecure bool
	// AI chat relay (POST /api/chat). Server holds the provider key so the
	// APK never ships one. Empty GeminiAPIKey = relay disabled (503).
	GeminiAPIKey   string
	GeminiModel   string
	ChatDailyLimit int
	ChatTimeoutSec int
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
		Port:        get("PORT", "8080"),
		DBHost:      get("DB_HOST", "localhost"),
		DBPort:      get("DB_PORT", "5432"),
		DBUser:      get("DB_USER", "postgres"),
		DBPassword:  get("DB_PASSWORD", "postgres"),
		DBName:      get("DB_NAME", "admin_daebak_music"),
		DBSSLMode:   get("DB_SSLMODE", "disable"),
		JWTSecret:   get("JWT_SECRET", "change-me-in-production"),
		JWTExpiry:   get("JWT_EXPIRY", "1h"),
		RefreshExpiryAdmin:    get("REFRESH_EXPIRY_ADMIN", "24h"),
		RefreshExpiryCustomer: get("REFRESH_EXPIRY_CUSTOMER", "2160h"),
		GoogleAndroidClientID: get("GOOGLE_ANDROID_CLIENT_ID", ""),
		GoogleIOSClientID:     get("GOOGLE_IOS_CLIENT_ID", ""),
		BootstrapSecret: get("BOOTSTRAP_SECRET", ""),
		FrontendURL: get("FRONTEND_URL", "http://localhost:3000"),
		CookieSecure: get("COOKIE_SECURE", "false") == "true" || os.Getenv("COOKIE_SECURE") == "1",
		GeminiAPIKey:   get("GEMINI_API_KEY", ""),
		GeminiModel:     get("GEMINI_MODEL", "gemini-2.0-flash"),
		ChatDailyLimit:  getInt("CHAT_DAILY_LIMIT", 30),
		ChatTimeoutSec:  getInt("CHAT_TIMEOUT_SEC", 45),
		ChatAppSecret:   get("CHAT_APP_SECRET", ""),
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