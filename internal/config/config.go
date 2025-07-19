package config

import (
	"log"
	"os"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	HTTP struct {
		Addr            string        `env:"HTTP_ADDR" envDefault:":8080"`
		ShutdownTimeout time.Duration `env:"HTTP_SHUTDOWN_TIMEOUT" envDefault:"5s"`
	}
	DB struct {
		DSN     string `env:"DB_DSN,required"`
		MaxOpen int    `env:"DB_MAX_OPEN" envDefault:"10"`
	}
	Auth struct {
		JWTSecret string        `env:"JWT_SECRET,required"`
		TokenTTL  time.Duration `env:"JWT_TTL" envDefault:"15m"`
	}
	LogLevel string `env:"LOG_LEVEL" envDefault:"INFO"`
}

func MustLoad() *Config {
	// For local dev, load .env file if it exists
	if _, err := os.Stat(".env"); err == nil {
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("Error loading .env file: %v", err)
		}
	}

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}

	return &cfg
}
