package main

import (
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
	dbConn, err := db.GetDBConnection()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()
	
	// Test connection with simple query
	var count int
	err = dbConn.QueryRow("SELECT COUNT(*) FROM kanji_go.kanji").Scan(&count)
	if err != nil {
		log.Fatalf("Failed to query database: %v", err)
	}
	fmt.Printf("✅ Database connection successful! Found %d kanji in the database.\n", count)
	
	// List a few kanji as additional verification
	rows, err := dbConn.Query("SELECT kanji_char_id, kanji_char FROM kanji_go.kanji LIMIT 5")
	if err != nil {
		log.Fatalf("Failed to query kanji: %v", err)
	}
	defer rows.Close()
	
	fmt.Println("\nSample kanji in database:")
	for rows.Next() {
		var id int
		var char string
		if err := rows.Scan(&id, &char); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		fmt.Printf("ID: %d, Character: %s\n", id, char)
	}
	
	// Create a test kanji entry
	fmt.Println("\n🔍 Testing insert, read, and delete operations...")
	
	// Generate a unique test kanji
	testKanji := "試験__試験"
	
	// Insert the test kanji
	insertSQL := `
		INSERT INTO kanji_go.kanji 
		(kanji_char, romaji_onyomi, romaji_kunyomi, hiragana_onyomi, hiragana_kunyomi, jlpt_level) 
		VALUES ($1, 'shiken', 'tesuto', 'テスト', 'しけん', 'n5')
		RETURNING kanji_char_id
	`
	
	var testID int
	err = dbConn.QueryRow(insertSQL, testKanji).Scan(&testID)
	if err != nil {
		log.Fatalf("Failed to insert test kanji: %v", err)
	}
	fmt.Printf("✅ Successfully inserted test kanji '%s' with ID %d\n", testKanji, testID)
	
	// Read it back to verify
	var readChar string
	err = dbConn.QueryRow("SELECT kanji_char FROM kanji_go.kanji WHERE kanji_char_id = $1", testID).Scan(&readChar)
	if err != nil {
		log.Fatalf("Failed to read back test kanji: %v", err)
	}
	
	if readChar != testKanji {
		log.Fatalf("Read value doesn't match: got '%s', expected '%s'", readChar, testKanji)
	}
	fmt.Printf("✅ Successfully read back the test kanji: '%s'\n", readChar)
	
	// Delete the test entry
	_, err = dbConn.Exec("DELETE FROM kanji_go.kanji WHERE kanji_char_id = $1", testID)
	if err != nil {
		log.Fatalf("Failed to delete test kanji: %v", err)
	}
	
	// Verify it was deleted
	var exists bool
	err = dbConn.QueryRow("SELECT EXISTS(SELECT 1 FROM kanji_go.kanji WHERE kanji_char_id = $1)", testID).Scan(&exists)
	if err != nil {
		log.Fatalf("Failed to check if test kanji was deleted: %v", err)
	}
	
	if exists {
		log.Fatalf("Test kanji still exists after deletion!")
	}
	fmt.Printf("✅ Successfully deleted test kanji with ID %d\n", testID)
	
	fmt.Println("\n🎉 All database tests passed successfully!")
}
