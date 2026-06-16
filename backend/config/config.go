package config

import (
	"os"
	"time"
)

type appConfig struct {
	MongoURI     string
	DatabaseName string
	JWTSecret    string
	JWTExpiry    time.Duration
	ServerPort   string
}

var C = appConfig{
	MongoURI:     getEnv("MONGODB_URI", "mongodb://localhost:27017"),
	DatabaseName: getEnv("DB_NAME", "stock_management"),
	JWTSecret:    getEnv("JWT_SECRET", "stock_management_secret_key_2026"),
	JWTExpiry:    24 * time.Hour,
	ServerPort:   getEnv("PORT", "8080"),
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
