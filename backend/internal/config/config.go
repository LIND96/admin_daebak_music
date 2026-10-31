package config

import (
	"os"

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
	FrontendURL string
	CookieSecure bool
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
		JWTExpiry:   get("JWT_EXPIRY", "24h"),
		FrontendURL: get("FRONTEND_URL", "http://localhost:3000"),
		CookieSecure: get("COOKIE_SECURE", "false") == "true" || os.Getenv("COOKIE_SECURE") == "1",
	}
}

func (c *Config) DSN() string {
	return "host=" + c.DBHost + " port=" + c.DBPort + " user=" + c.DBUser +
		" password=" + c.DBPassword + " dbname=" + c.DBName + " sslmode=" + c.DBSSLMode
}