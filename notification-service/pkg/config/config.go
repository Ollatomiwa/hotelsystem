package config

import (
	"os"
	"strconv"
)

type Config struct {
	ServerPort string
	DatabasePath string
	Environment string
	LogLevel string 
}

func Load() *Config {
	return &Config{
		ServerPort: getEnv("Server_Port", "8080"),
		DatabasePath: getEnv("Database_Path", "./notifications.db"),
		Environment: getEnv("Environment", "development"),
		LogLevel: getEnv("Log_Level", "info"),
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


