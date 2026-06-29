package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	CORSOrigins []string
	LLMProvider string
	LLMModel    string
	LLMBaseURL  string
	PiperBin    string
	PiperVoice  string
	VoicesDir   string
}

// Load reads configuration from .env (if present) and the environment.
func Load() (Config, error) {
	_ = godotenv.Load() // ignore error: env vars may be set externally

	cfg := Config{
		Port:        getenv("PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		CORSOrigins: splitAndTrim(getenv("CORS_ORIGINS", "http://localhost:8081,http://localhost:19006")),
		LLMProvider: getenv("LLM_PROVIDER", "mock"),
		LLMModel:    os.Getenv("LLM_MODEL"),
		LLMBaseURL:  os.Getenv("LLM_BASE_URL"),
		PiperBin:    getenv("PIPER_BIN", "auto"),
		PiperVoice:  os.Getenv("PIPER_VOICE"),
		VoicesDir:   getenv("VOICES_DIR", "voices"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
