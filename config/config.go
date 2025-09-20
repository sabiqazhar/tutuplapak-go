package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type DBConfig struct {
	Host string
	Port string
	User string
	Pass string
	Name string
}

type Config struct {
	Database DBConfig
}

// LoadConfig loads from .env if present, else from system env
func LoadConfig() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, falling back to system environment")
	}

	cfg := &Config{
		Database: DBConfig{
			Host: getEnv("DB_HOST", ""),
			Port: getEnv("DB_PORT", ""),
			User: getEnv("DB_USER", ""),
			Pass: getEnv("DB_PASS", ""),
			Name: getEnv("DB_NAME", ""),
		},
	}

	return cfg
}

// Helper function: env var with fallback
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
