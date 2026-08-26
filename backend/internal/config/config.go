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
	NineRouterAPIKey   string
	NineRouterBaseURL  string
	OpenCodeAPIKey     string
	OpenRouterModel    string
	AIBaseURL          string
	Env                string
	CORSAllowedOrigins []string
}

func Load() (*Config, error) {
	apiKey := strings.TrimSpace(os.Getenv("OPENROUTER_API_KEY"))
	if apiKey == "" {
		apiKey = strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	}
	if apiKey == "" {
		apiKey = strings.TrimSpace(os.Getenv("AI_API_KEY"))
	}
	openCodeKey := strings.TrimSpace(os.Getenv("OPENCODE_API_KEY"))
	if openCodeKey == "" {
		openCodeKey = tryLoadOpencodeKey()
	}
	model := strings.TrimSpace(os.Getenv("OPENROUTER_MODEL"))
	if model == "" {
		model = strings.TrimSpace(os.Getenv("OPENAI_MODEL"))
	}
	if model == "" {
		model = "stealth/ox-alpha"
	}
	cfg := &Config{
		Port:              getenv("PORT", "8080"),
		DatabaseURL:       strings.TrimSpace(os.Getenv("DATABASE_URL")),
		StoragePath:       getenv("STORAGE_PATH", "./data/uploads"),
		AIAPIKey:          apiKey,
		NineRouterAPIKey:  strings.TrimSpace(os.Getenv("NINEROUTER_API_KEY")),
		NineRouterBaseURL: strings.TrimRight(getenv("NINEROUTER_BASE_URL", "https://9router.103-59-94-121.nip.io/v1"), "/"),
		OpenCodeAPIKey:    openCodeKey,
		OpenRouterModel:   model,
		AIBaseURL:         strings.TrimRight(firstNonEmptyEnv("OPENROUTER_BASE_URL", "OPENAI_BASE_URL"), "/"),
		Env:               getenv("ENV", "development"),
	}

	origins := getenv("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:8080,http://127.0.0.1:8080,http://127.0.0.1:5173")
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

func firstNonEmptyEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
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
	candidates = append(candidates,
		"/root/.local/share/opencode/auth.json",
		"/home/app/.local/share/opencode/auth.json",
		"/app/.local/share/opencode/auth.json",
		"./opencode-auth.json",
		"/tmp/opencode-auth.json",
	)
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
