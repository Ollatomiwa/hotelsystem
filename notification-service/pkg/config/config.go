package config

import (

	"os"
	"strconv"

	
)

type Config struct {
	ServerPort string
	DatabaseURL string
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

	//security configuration
	MaxRequestBodySize int
	AllowedOrigins string

	//logging config
	LogFormat string
}


func Load() *Config {
    // Railway provides PORT environment variable
    serverPort := getEnv("PORT", "8080")
    if serverPort == "8080" {
        serverPort = getEnv("SERVER_PORT", "8080")
    }

    return &Config{
        ServerPort:  serverPort,
        DatabaseURL: getEnv("DATABASE_URL", ""), // FIXED: Database_URL → DATABASE_URL
        Environment: getEnv("ENVIRONMENT", "development"), // FIXED: Environment → ENVIRONMENT
        LogLevel:    getEnv("LOG_LEVEL", "info"), // FIXED: Log_Level → LOG_LEVEL
    
        // Email configuration with defaults
        SMTPHost:      getEnv("SMTP_HOST", "smtp.gmail.com"),
        SMTPPort:      getEnvInt("SMTP_PORT", 465), // FIXED: 587 → 465
        SMTPUsername:  getEnv("SMTP_USERNAME", ""),
        SMTPPassword:  getEnv("SMTP_PASSWORD", ""),
        FromeEmail:     getEnv("FROM_EMAIL", "noreply@example.com"), // FIXED: FromeEmail → FromEmail

        // Rate limiting configuration
        RateLimitRequest: getEnvInt("RATE_LIMIT_REQUESTS", 5), // FIXED: RateLimitRequest → RateLimitRequests
        RateLimitMinutes:  getEnvInt("RATE_LIMIT_MINUTES", 1),

        // Security configuration
        MaxRequestBodySize: getEnvInt("MAX_REQUEST_SIZE", 1*1024*1024),
        AllowedOrigins:     getEnv("ALLOWED_ORIGINS", "*"),

        // Logging configuration
        LogFormat: getEnv("LOG_FORMAT", "json"),
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


