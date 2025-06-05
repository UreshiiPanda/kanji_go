package db

import (
	"context"
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

	// Construct connection string
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=require",
		dbUser, dbPassword, dbHost, dbPort, dbName)

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

// For direct pgx usage when needed (e.g., COPY, LISTEN/NOTIFY)
// this just adds extra PGX features which you probably never need
func GetPgxConnection() (*pgx.Conn, error) {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// Construct connection string
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=require",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	// Connect using pgx directly
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return conn, nil
}
