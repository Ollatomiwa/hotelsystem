package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort string
	DatabasePath string
	Environment string
	LogLevel string 

	//email configurations
	SMTPHost string
	SMTPPort int
	SMTPUsername string
	SMTPPassword string
	FromeEmail string

	//rate limit configuration
	RateLimitRequest int
	RateLimitMinutes int
}


func Load() *Config {
	//load env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: No .env file found, using environment variable only")
	}
	return &Config{
		ServerPort: getEnv("Server_Port", "8080"),
		DatabasePath: getEnv("Database_Path", "./notifications.db"),
		Environment: getEnv("Environment", "development"),
		LogLevel: getEnv("Log_Level", "info"),
	
		//email configs with defaults
		SMTPHost: getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort: getEnvInt("SMTP_PORT", 587),
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		FromeEmail: getEnv("FROM_EMAIL", "noreply@example.com"),

		//rate linit configs
		RateLimitRequest: getEnvInt("RATE_LIMIT_REQUEST", 5),
		RateLimitMinutes: getEnvInt("RATE_LIMIT_MINUTES", 1),
	}
}


//helperfunction
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err != nil{
			return intValue
		}
	}
	return defaultValue
}


