package main

import (
	configs "UrlShortenerBackend/config"
	"UrlShortenerBackend/internal/link"
	"UrlShortenerBackend/pkg/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := configs.Init()

	log := logger.NewLogger(cfg)

	log.Info().Msg("Starting auto migration...")
	log.Info().Msg("Connecting to database...")

	db, err := gorm.Open(postgres.Open(cfg.Db.Dsn), &gorm.Config{})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}

	log.Info().Msg("Running migration for Link model...")
	err = db.AutoMigrate(&link.Link{})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}

	log.Info().Msg("Auto migration completed successfully!")
}
