package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv      string
	BackendPort string
	DatabaseDSN string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string

	AIProvider              string
	DeepSeekAPIKey          string
	DeepSeekBaseURL         string
	DeepSeekModel           string
	DeepSeekTimeoutSeconds  int
	DeepSeekMaxOutputTokens int

	EnableDebugRoutes  bool
	CORSAllowedOrigins []string
	EnableRateLimit    bool
	RateLimitPerMinute int
	SQLDir             string
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("../../.env")

	cfg := &Config{
		AppEnv:      getEnv("APP_ENV", "development"),
		BackendPort: getEnv("BACKEND_PORT", "8080"),
		DatabaseDSN: getEnv("DATABASE_DSN", ""),
		DBHost:      getEnv("DB_HOST", "127.0.0.1"),
		DBPort:      getEnv("DB_PORT", "3306"),
		DBUser:      getEnv("DB_USER", "yijing"),
		DBPassword:  getEnv("DB_PASSWORD", "yijingpass"),
		DBName:      getEnv("DB_NAME", "yijing"),

		AIProvider:              getEnv("AI_PROVIDER", "mock"),
		DeepSeekAPIKey:          getEnv("DEEPSEEK_API_KEY", ""),
		DeepSeekBaseURL:         getEnv("DEEPSEEK_BASE_URL", "https://api.deepseek.com"),
		DeepSeekModel:           getEnv("DEEPSEEK_MODEL", "deepseek-v4-flash"),
		DeepSeekTimeoutSeconds:  ParseIntDefault(getEnv("DEEPSEEK_TIMEOUT_SECONDS", "60"), 60),
		DeepSeekMaxOutputTokens: ParseIntDefault(getEnv("DEEPSEEK_MAX_OUTPUT_TOKENS", "1800"), 1800),

		EnableDebugRoutes:  ParseBoolDefault(getEnv("ENABLE_DEBUG_ROUTES", "false"), false),
		CORSAllowedOrigins: parseCSV(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")),
		EnableRateLimit:    ParseBoolDefault(getEnv("ENABLE_RATE_LIMIT", "true"), true),
		RateLimitPerMinute: ParseIntDefault(getEnv("RATE_LIMIT_PER_MINUTE", "20"), 20),
		SQLDir:             getEnv("SQL_DIR", ""),
	}
	return cfg, nil
}

func (c *Config) DSN() string {
	if strings.TrimSpace(c.DatabaseDSN) != "" {
		return strings.TrimSpace(c.DatabaseDSN)
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=Local",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func ParseIntDefault(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return n
}

func ParseBoolDefault(s string, fallback bool) bool {
	if s == "" {
		return fallback
	}
	v, err := strconv.ParseBool(s)
	if err != nil {
		return fallback
	}
	return v
}

func parseCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
