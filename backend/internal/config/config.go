package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	Addr        string
	CORSOrigins []string
	JWTSecret   []byte
	TokenTTL    time.Duration
	StorePath   string
}

func Load() (Config, error) {
	cfg := Config{
		Addr:      envOr("APP_ADDR", ":8080"),
		TokenTTL:  envDurationOr("JWT_TTL", 24*time.Hour),
		StorePath: envOr("STORE_PATH", "data/users.json"),
	}

	origins := envOr("CORS_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173")
	for _, origin := range strings.Split(origins, ",") {
		if origin = strings.TrimSpace(origin); origin != "" {
			cfg.CORSOrigins = append(cfg.CORSOrigins, origin)
		}
	}

	cfg.JWTSecret = []byte(os.Getenv("JWT_SECRET"))
	if len(cfg.JWTSecret) == 0 {
		cfg.JWTSecret = []byte("dev-secret-jangan-dipakai-di-production-ubah-me")
		fmt.Fprintln(os.Stderr, "[WARN] JWT_SECRET belum di-set, memakai secret default untuk development.")
	}
	if len(cfg.JWTSecret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET minimal 32 karakter")
	}

	return cfg, nil
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func envDurationOr(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		return fallback
	}
	return d
}
