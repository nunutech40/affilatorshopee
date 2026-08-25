package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Port               string
	DatabaseURL        string
	StoragePath        string
	AIAPIKey           string
	OpenRouterModel    string
	Env                string
	CORSAllowedOrigins []string
}

func Load() (*Config, error) {
	apiKey := strings.TrimSpace(os.Getenv("AI_API_KEY"))
	if apiKey == "" {
		apiKey = strings.TrimSpace(os.Getenv("OPENCODE_API_KEY"))
	}
	if apiKey == "" {
		apiKey = tryLoadOpencodeKey()
	}
	model := strings.TrimSpace(os.Getenv("OPENROUTER_MODEL"))
	if model == "" {
		model = "opencode/muse-spark-1.2-contributor-free"
	}
	cfg := &Config{
		Port:            getenv("PORT", "8080"),
		DatabaseURL:     strings.TrimSpace(os.Getenv("DATABASE_URL")),
		StoragePath:     getenv("STORAGE_PATH", "./data/uploads"),
		AIAPIKey:        apiKey,
		OpenRouterModel: model,
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

func tryLoadOpencodeKey() string {
	candidates := []string{}
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, ".local", "share", "opencode", "auth.json"))
		candidates = append(candidates, filepath.Join(home, "Library", "Application Support", "ai.opencode.desktop", "opencode.db"))
	}
	candidates = append(candidates, "/root/.local/share/opencode/auth.json", "./opencode-auth.json")
	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		// Try JSON map like {"opencode-go":{"type":"api","key":"sk-..."}}
		var m map[string]map[string]string
		if err := json.Unmarshal(data, &m); err == nil {
			if entry, ok := m["opencode-go"]; ok {
				if key, ok := entry["key"]; ok && strings.TrimSpace(key) != "" {
					return strings.TrimSpace(key)
				}
			}
			// also try generic
			for _, v := range m {
				if key, ok := v["key"]; ok && strings.HasPrefix(strings.TrimSpace(key), "sk-") {
					return strings.TrimSpace(key)
				}
			}
		}
	}
	return ""
}
