package main

import (
	"context"
	"fmt"
	"log"

	"github.com/UreshiiPanda/kanji_go/internal/db"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Get database connection
	conn, err := db.GetPgxConnection()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close(context.Background())

	// Test connection with simple query
	var count int
	err = conn.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM kanji_go.kanji",
	).Scan(&count)

	if err != nil {
		log.Fatalf("Failed to query database: %v", err)
	}

	fmt.Printf("Database connection successful! Found %d kanji in the database.\n", count)
	
	// List a few kanji as additional verification
	rows, err := conn.Query(context.Background(),
		"SELECT kanji_char_id, kanji_char FROM kanji_go.kanji LIMIT 5")
	if err != nil {
		log.Fatalf("Failed to query kanji: %v", err)
	}
	defer rows.Close()
	
	fmt.Println("Sample kanji in database:")
	for rows.Next() {
		var id int
		var char string
		if err := rows.Scan(&id, &char); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		fmt.Printf("ID: %d, Character: %s\n", id, char)
	}
}
