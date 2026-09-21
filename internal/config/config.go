package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port             string
	DatabaseURL      string
	Env              string
	LogLevel         string
	JWTSecret        string
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration
	CORSOrigins      []string
	MigrateOnStartup bool
	UploadDir        string
}

func Load() Config {
	_ = godotenv.Load()

	env := getEnv("APP_ENV", "development")
	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		if env == "production" {
			log.Fatal("JWT_SECRET must be set in production")
		}
		jwtSecret = "dev-only-insecure-secret-change-me"
		log.Println("WARNING: JWT_SECRET not set, using an insecure development default")
	}

	return Config{
		Port:             getEnv("PORT", "8080"),
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/clothing_erp?sslmode=disable"),
		Env:              env,
		LogLevel:         getEnv("LOG_LEVEL", "info"),
		JWTSecret:        jwtSecret,
		AccessTokenTTL:   time.Duration(getEnvInt("ACCESS_TOKEN_TTL_MIN", 15)) * time.Minute,
		RefreshTokenTTL:  time.Duration(getEnvInt("REFRESH_TOKEN_TTL_DAYS", 7)) * 24 * time.Hour,
		CORSOrigins:      getEnvList("CORS_ALLOWED_ORIGINS", []string{"*"}),
		MigrateOnStartup: getEnvBool("MIGRATE_ON_STARTUP", true),
		UploadDir:        getEnv("UPLOAD_DIR", "./uploads"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}

// getEnvList reads a comma-separated env var into a trimmed slice, e.g.
// CORS_ALLOWED_ORIGINS="https://app.example.com,https://admin.example.com".
func getEnvList(key string, fallback []string) []string {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}
