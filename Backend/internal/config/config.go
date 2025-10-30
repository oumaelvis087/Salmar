package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Port                string
	Environment         string
	PythonServicesHost  string
	NLPServicePort      string
	ImageServicePort    string
	CodeServicePort     string
	LogLevel            string
	MaxRequestSize      int64
	RequestTimeout      int
	EnableCaching       bool
	CacheExpirationSecs int
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Set defaults and override with environment variables
	cfg := &Config{
		Port:                getEnv("PORT", "8080"),
		Environment:         getEnv("ENVIRONMENT", "development"),
		PythonServicesHost:  getEnv("PYTHON_SERVICES_HOST", "localhost"),
		NLPServicePort:      getEnv("NLP_SERVICE_PORT", "5001"),
		ImageServicePort:    getEnv("IMAGE_SERVICE_PORT", "5002"),
		CodeServicePort:     getEnv("CODE_SERVICE_PORT", "5003"),
		LogLevel:            getEnv("LOG_LEVEL", "info"),
		MaxRequestSize:      getEnvAsInt64("MAX_REQUEST_SIZE", 10*1024*1024), // 10MB default
		RequestTimeout:      getEnvAsInt("REQUEST_TIMEOUT", 30),              // 30 seconds default
		EnableCaching:       getEnvAsBool("ENABLE_CACHING", true),
		CacheExpirationSecs: getEnvAsInt("CACHE_EXPIRATION_SECS", 300), // 5 minutes default
	}

	return cfg, nil
}

// Helper functions to get environment variables with defaults
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}