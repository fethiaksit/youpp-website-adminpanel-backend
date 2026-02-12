package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	MongoURI           string
	MongoDB            string
	JWTSecret          string
	JWTRefreshSecret   string
	ProvisionAPIKey    string
	FrontendOrigins    []string
	AccessTTLMinutes   int
	RefreshTTLDays     int
	SuperAdminEmail    string
	SuperAdminPassword string
	DemoEmail          string
	DemoPassword       string
	DemoSiteSlug       string
}

func Load() (*Config, error) {
	cfg := &Config{
		MongoURI:           os.Getenv("MONGO_URI"),
		MongoDB:            os.Getenv("MONGO_DB"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		JWTRefreshSecret:   os.Getenv("JWT_REFRESH_SECRET"),
		ProvisionAPIKey:    os.Getenv("PROVISION_API_KEY"),
		FrontendOrigins:    getFrontendOrigins(),
		SuperAdminEmail:    os.Getenv("SUPERADMIN_EMAIL"),
		SuperAdminPassword: os.Getenv("SUPERADMIN_PASSWORD"),
		DemoEmail:          os.Getenv("DEMO_EMAIL"),
		DemoPassword:       os.Getenv("DEMO_PASSWORD"),
		DemoSiteSlug:       os.Getenv("DEMO_SITE_SLUG"),
	}

	if cfg.MongoURI == "" || cfg.MongoDB == "" || cfg.JWTSecret == "" || cfg.JWTRefreshSecret == "" {
		return nil, fmt.Errorf("missing required environment variables")
	}
	if cfg.DemoSiteSlug == "" {
		cfg.DemoSiteSlug = "demo-site"
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

func getFrontendOrigins() []string {
	frontendOrigins := os.Getenv("FRONTEND_ORIGINS")
	if frontendOrigins == "" {
		return nil
	}

	rawOrigins := strings.Split(frontendOrigins, ",")
	origins := make([]string, 0, len(rawOrigins))
	for _, origin := range rawOrigins {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}

	if len(origins) == 0 {
		return nil
	}

	return origins
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
