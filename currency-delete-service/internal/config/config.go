package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct{}

func (c *Config) LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func (c *Config) GetEnv(key string, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}
