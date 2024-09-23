package config

import (
	"log"
	"os"
	"strconv"
)

// Config struct holds all the configuration values
type Config struct {
	SERPort      string
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	REDISAddr    string
	REDISPwd     string
	REDISPort    string
	APIKey       string
	Seed         int64
	MaxRetries   int
	RetryDelayMs int
}

// AppConfig holds the global configuration
var AppConfig *Config

// LoadConfig loads the environment variables or default values for the application
func LoadConfig() {
	AppConfig = &Config{
		SERPort:      getEnv("SERVER_PORT", ":8080"),
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnv("DB_PORT", "5433"),
		DBUser:       getEnv("DB_USER", "llmsuser"),
		DBPassword:   getEnv("DB_PASSWORD", "llmspasswd"),
		DBName:       getEnv("DB_NAME", "llm_stats"),
		REDISAddr:    getEnv("REDISAddr", "localhost:6379"),
		REDISPwd:     getEnv("REDISPwd", "redis1234"),
		APIKey:       getEnv("API_KEY", "GPA-prince-edusei-2024"),
		Seed:         getEnvAsInt64("SEED", 42),
		MaxRetries:   getEnvAsInt("MAX_RETRIES", 3),
		RetryDelayMs: getEnvAsInt("RETRY_DELAY_MS", 1000),
	}
}

// getEnv fetches an environment variable, or returns a default value if not present
func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

// getEnvAsInt fetches an integer environment variable or returns a default value
func getEnvAsInt(key string, defaultVal int) int {
	if valueStr, exists := os.LookupEnv(key); exists {
		value, err := strconv.Atoi(valueStr)
		if err != nil {
			log.Fatalf("Invalid value for %s: %v", key, err)
		}
		return value
	}
	return defaultVal
}

// getEnvAsInt64 fetches an int64 environment variable or returns a default value
func getEnvAsInt64(key string, defaultVal int64) int64 {
	if valueStr, exists := os.LookupEnv(key); exists {
		value, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			log.Fatalf("Invalid value for %s: %v", key, err)
		}
		return value
	}
	return defaultVal
}
