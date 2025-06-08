package config

import (
	"log"
	"os"
	
	"github.com/joho/godotenv"
)

// AppConfig holds the application configuration
type AppConfig struct {
	Port string
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
	
	return &AppConfig{
		Port: port,
	}
}
