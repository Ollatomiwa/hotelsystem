package config

import (
	"os"
	"strconv"
	"time"
	"strings"
)

type Config struct {
	Server ServerConfig
	Database DatabaseConfig
	Security SecurityConfig
	Redis RedisConfig
}

type ServerConfig struct {
	Port string
	Env string
	ReadTimeout time.Duration
	WriteTimeout time.Duration
	IdleTimeout time.Duration
}
type DatabaseConfig struct {
	Host string
	Port string
	User string
	Password string
	DBName string
	SSLMode string
}

type SecurityConfig struct {
	JWTSecretKey string
	JWTRefreshKey string
	AccessTokenDuration time.Duration
	RefreshToken time.Duration
	BCryptCost int
	CORSAllowedOrigins []string
	CSRFSecret string
	RateLimitRequests int
	RateLimitWindow time.Duration
}

type RedisConfig struct {
	Host string
	Port string
	Password string
	DB int
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8082"),
			Env: getEnv("ENV", "development"),
			ReadTimeout: getEnvDuration("READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getEnvDuration("WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout: getEnvDuration("IDLE_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			Host: getEnv("DB_HOST","localhost"),
			Port: getEnv("DB_PORT", "5342"),
			User: getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName: getEnv("DB_NAME", "user_service"),
			SSLMode: getEnv("DB_SSL_MODE", "disable"),
		},Security: SecurityConfig{
			JWTSecretKey: getEnv("JWT_SECRET_KEY", "256-bit-secret"),
			JWTRefreshKey: getEnv("JWT_REFRESH_KEY", "refrsh-token"),
			AccessTokenDuration: getEnvDuration("ACCESS_TOKEN_DURATION", 15*time.Minute),
			RefreshToken: getEnvDuration("REFRESH_TOKEN", 7*24*time.Hour),
			BCryptCost: getEnvInt("BCRYPT_COST", 12),
			CORSAllowedOrigins: getEnvSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
			CSRFSecret: getEnv("CSRF_SECRET", "scfr-key"),
			RateLimitRequests: getEnvInt("RATE_LIMIT_REQUESTS", 100),
			RateLimitWindow: getEnvDuration("RATE_LIMIT_WINDOW", 1*time.Minute),
		},
		Redis: RedisConfig{
			Host: getEnv("REDIS_HOST", "localhost"),
			Port: getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB: getEnvInt("REDIS_DB", 0),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return  defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != ""{
		if duration, err := time.ParseDuration(value); err == nil {
			return duration 
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}