package config

import (
	"os"
	"strings"
)

type Config struct {
	DatabasePath string
	JWTSecret    string
	Port         string
	CORSOrigins  []string
	LogoDir      string
}

func Load() Config {
	return Config{
		DatabasePath: envOrDefault("DATABASE_PATH", "/data/radar.db"),
		JWTSecret:    envOrDefault("JWT_SECRET", "change-me-radar-jwt"),
		Port:         envOrDefault("PORT", "8088"),
		CORSOrigins:  splitOrigins(envOrDefault("CORS_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173")),
		LogoDir:      envOrDefault("LOGO_DIR", "/data/logos"),
	}
}

func envOrDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func splitOrigins(value string) []string {
	parts := strings.Split(value, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		if origin := strings.TrimSpace(part); origin != "" {
			origins = append(origins, origin)
		}
	}
	return origins
}
