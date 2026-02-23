package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL      string
	Port             string
	JWTAccessSecret  string
	JWTRefreshSecret string
}

func LoadConfig() (*Config, error) {
	var cfg Config
	var err error

	godotenv.Load()

	cfg.DatabaseURL, err = getEnv("DATABASE_URL")
	if err != nil {
		return nil, err
	}

	cfg.Port, err = getEnv("PORT")
	if err != nil {
		return nil, err
	}

	cfg.JWTAccessSecret, err = getEnv("JWT_ACCESS_SECRET")
	if err != nil {
		return nil, err
	}

	cfg.JWTRefreshSecret, err = getEnv("JWT_REFRESH_SECRET")
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func getEnv(key string) (string, error) {
	envStr := os.Getenv(key)
	if envStr == "" {
		return "", fmt.Errorf("environment variable %s not set", key)
	}
	return envStr, nil
}
