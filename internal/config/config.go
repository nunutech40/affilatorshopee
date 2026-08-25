package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port               string
	DatabaseURL        string
	AIAPIKey           string
	OpenRouterModel    string
	Env                string
	CORSAllowedOrigins []string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:            getenv("PORT", "8080"),
		DatabaseURL:     strings.TrimSpace(os.Getenv("DATABASE_URL")),
		AIAPIKey:        strings.TrimSpace(os.Getenv("AI_API_KEY")),
		OpenRouterModel: strings.TrimSpace(os.Getenv("OPENROUTER_MODEL")),
		Env:             getenv("ENV", "development"),
	}

	origins := getenv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")
	for _, origin := range strings.Split(origins, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			cfg.CORSAllowedOrigins = append(cfg.CORSAllowedOrigins, origin)
		}
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
