package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ApiUrl      string
	AuthToken   string
	DatabaseURL string
}

func Load() (*Config, error) {
	var err error = godotenv.Load()

	if err != nil {
		log.Println(".env not found. Using enviroment variables instead")
	}

	var config *Config = &Config{
		ApiUrl:      os.Getenv("API_BASE_URL"),
		AuthToken:   os.Getenv("API_BEARER_TOKEN"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}

	return config, nil
}
