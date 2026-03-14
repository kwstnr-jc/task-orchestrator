package config

import (
	"os"
	"strings"
)

type Config struct {
	Port            string
	DatabaseURL     string
	Auth0Domain     string
	Auth0ClientID   string
	Auth0Secret     string
	Auth0Callback   string
	SessionSecret   string
	CookieDomain    string
	AllowedUsers    []string
	AllowedMachines []string
	SQSQueueURL     string
	AWSRegion       string
	CORSOrigin      string
	DevMode         bool
	APIKey          string
}

func Load() *Config {
	return &Config{
		Port:            envOr("PORT", "8080"),
		DatabaseURL:     envOr("DATABASE_URL", "postgres://orchestrator:changeme@localhost:5432/orchestrator?sslmode=disable"),
		Auth0Domain:     os.Getenv("AUTH0_DOMAIN"),
		Auth0ClientID:   os.Getenv("AUTH0_CLIENT_ID"),
		Auth0Secret:     os.Getenv("AUTH0_CLIENT_SECRET"),
		Auth0Callback:   os.Getenv("AUTH0_CALLBACK_URL"),
		SessionSecret:   os.Getenv("SESSION_SECRET"),
		CookieDomain:    envOr("COOKIE_DOMAIN", "localhost"),
		AllowedUsers:    splitCSV(envOr("ALLOWED_USERS", "")),
		AllowedMachines: splitCSV(envOr("ALLOWED_MACHINES", "")),
		SQSQueueURL:     os.Getenv("SQS_QUEUE_URL"),
		AWSRegion:       envOr("AWS_REGION", "eu-central-1"),
		CORSOrigin:      envOr("CORS_ORIGIN", "http://localhost:5173"),
		DevMode:         os.Getenv("DEV_MODE") == "true",
		APIKey:          os.Getenv("API_KEY"),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
