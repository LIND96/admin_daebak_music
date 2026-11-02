package database

import (
	"fmt"

	"admin_daebak_music/backend/internal/config"
	"admin_daebak_music/backend/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate for convenience (migrations also provided as SQL files).
	if err := db.AutoMigrate(&models.User{}, &models.RefreshToken{}, &models.AppUsageEvent{}); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	return db, nil
}