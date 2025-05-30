// config/auth.go
package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var JWTSecret string

func init() {
	// load .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found — reading env vars")
	}
	JWTSecret = os.Getenv("JWT_SECRET")
	if JWTSecret == "" {
		log.Fatal("JWT_SECRET must be set")
	}
}
