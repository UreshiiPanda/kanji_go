package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

// GetDBConnection establishes a connection to the PostgreSQL database
// Returns a standard *sql.DB for compatibility with other packages
func GetDBConnection() (*sql.DB, error) {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	appEnv := os.Getenv("APP_ENV")

	var connStr string

	// Check if we're in production
	if appEnv == "PROD" {
		// Cloud Run with Cloud SQL connection
		// Format: /cloudsql/CONNECTION_NAME
		socketDir := "/cloudsql"
		instanceConnectionName := dbHost
		log.Printf("PROD environment: Using Cloud SQL socket connection")
		
		// For Cloud SQL with Unix socket
		connStr = fmt.Sprintf("host=%s/%s user=%s password=%s dbname=%s sslmode=disable",
			socketDir, instanceConnectionName, dbUser, dbPassword, dbName)
	} else {
		// Direct connection (local development)
		log.Printf("LOCAL environment: Using direct connection to %s:%s", dbHost, dbPort)
		
		// For direct TCP connection
		connStr = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=require",
			dbUser, dbPassword, dbHost, dbPort, dbName)
	}

	// Parse connection string into pgx config
	config, err := pgx.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing connection string: %w", err)
	}

	// Convert to standard sql.DB connection
	db := stdlib.OpenDB(*config)
	
	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database")
	return db, nil
}
