package config

import (
	"log"
	"os"
	
	"github.com/joho/godotenv"
)

// AppConfig holds the application configuration
type AppConfig struct {
	Port   string
	AppEnv string // Added APP_ENV
}

// Load loads the application configuration
func Load() *AppConfig {
	// Load environment variables from .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}
	
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	// Get environment setting
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "LOCAL" // Default to LOCAL if not specified
	}
	
	return &AppConfig{
		Port:   port,
		AppEnv: appEnv,
	}
}

// IsProd returns true if running in production environment
func (c *AppConfig) IsProd() bool {
	return c.AppEnv == "PROD"
}
