package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DatabaseModule = fx.Provide(NewDB)

// Provide a GORM DB for Fx
func NewDB() (*gorm.DB, error) {
	// Load .env if present (only once at startup)
	_ = godotenv.Load() // ignore error, fallback to env

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN not set in environment")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}
