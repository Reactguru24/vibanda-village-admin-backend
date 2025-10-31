package config

import (
	"log"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port            string
	GinMode         string
	MongoURI        string
	DatabaseName    string
	JWTSecret       string
	JWTExpirationHours int
	AllowedOrigins  []string
	MaxFileSize     string
	UploadPath      string
}

func Load() *Config {
	return &Config{
		Port:            getEnv("PORT", "8080"),
		GinMode:         getEnv("GIN_MODE", "debug"),
		MongoURI:        getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		DatabaseName:    getEnv("DATABASE_NAME", "vibanda_village"),
		JWTSecret:       getEnv("JWT_SECRET", "your-super-secret-jwt-key-here"),
		JWTExpirationHours: getEnvAsInt("JWT_EXPIRATION_HOURS", 24),
		AllowedOrigins:  getEnvAsSlice("ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://localhost:5173"}),
		MaxFileSize:     getEnv("MAX_FILE_SIZE", "10MB"),
		UploadPath:      getEnv("UPLOAD_PATH", "uploads/"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	log.Printf("Environment variable %s not set, using default: %s", key, defaultValue)
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Invalid integer value for %s: %s, using default: %d", key, value, defaultValue)
	}
	log.Printf("Environment variable %s not set, using default: %d", key, defaultValue)
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	log.Printf("Environment variable %s not set, using default: %v", key, defaultValue)
	return defaultValue
}
