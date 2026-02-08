package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	MongoURI         string
	MongoDB          string
	JWTSecret        string
	JWTRefreshSecret string
	AccessTTLMinutes int
	RefreshTTLDays   int
}

func Load() (*Config, error) {
	cfg := &Config{
		MongoURI:         os.Getenv("MONGO_URI"),
		MongoDB:          os.Getenv("MONGO_DB"),
		JWTSecret:        os.Getenv("JWT_SECRET"),
		JWTRefreshSecret: os.Getenv("JWT_REFRESH_SECRET"),
	}

	if cfg.MongoURI == "" || cfg.MongoDB == "" || cfg.JWTSecret == "" || cfg.JWTRefreshSecret == "" {
		return nil, fmt.Errorf("missing required environment variables")
	}

	accessTTL, err := getEnvInt("ACCESS_TTL_MIN", 15)
	if err != nil {
		return nil, fmt.Errorf("ACCESS_TTL_MIN: %w", err)
	}

	refreshTTL, err := getEnvInt("REFRESH_TTL_DAYS", 30)
	if err != nil {
		return nil, fmt.Errorf("REFRESH_TTL_DAYS: %w", err)
	}

	cfg.AccessTTLMinutes = accessTTL
	cfg.RefreshTTLDays = refreshTTL

	return cfg, nil
}

func getEnvInt(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}

	return parsed, nil
}
