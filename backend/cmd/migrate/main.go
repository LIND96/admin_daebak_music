package main

import (
	"fmt"
	"log"
	"os"

	"admin_daebak_music/backend/internal/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	cfg := config.Load()

	pgURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSSLMode)

	m, err := migrate.New("file://migrations", pgURL)
	if err != nil {
		log.Fatal(err)
	}

	action := "up"
	if len(os.Args) > 1 {
		action = os.Args[1]
	}
	switch action {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
		log.Println("migrations applied")
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
		log.Println("migrations rolled back")
	default:
		log.Fatalf("unknown action: %s (use up|down)", action)
	}
}