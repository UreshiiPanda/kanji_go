package models

import (
	"database/sql"
	"fmt"
	"time"
)

// User represents a user of the application
type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // Never expose in JSON
	KanjiPacks   []string  `json:"kanji_packs"`
	StarredKanji []int     `json:"starred_kanji"`
	SavedKanji   []int     `json:"saved_kanji"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Session represents a user session
type Session struct {
	SessionID          string     `json:"session_id"`
	CurrentUser        *string    `json:"curr_user"` // Pointer to allow NULL
	CurrentJLPTLevel   string     `json:"curr_jlpt_level"`
	CurrentPage        string     `json:"curr_page"`
	ContactPopupActive bool       `json:"contact_popup_active"`
	LoginPopupActive   bool       `json:"login_popup_active"`
	PaymentPopupActive bool       `json:"payment_popup_active"`
	LeftSidebarActive  bool       `json:"left_sidebar_active"`
	DarkModeActive     bool       `json:"dark_mode_active"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// Kanji represents a Japanese kanji character
type Kanji struct {
	KanjiCharID      int       `json:"kanji_char_id"`
	KanjiChar        string    `json:"kanji_char"`
	RomajiOnyomi     string    `json:"romaji_onyomi"`
	RomajiKunyomi    string    `json:"romaji_kunyomi"`
	HiraganaOnyomi   string    `json:"hiragana_onyomi"`
	HiraganaKunyomi  string    `json:"hiragana_kunyomi"`
	JLPTLevel        string    `json:"jlpt_level"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// KanjiCreation represents a user-created explanation for a kanji
type KanjiCreation struct {
	KanjiCreationID int        `json:"kanji_creation_id"`
	KanjiCharID     int        `json:"kanji_char_id"`
	CreatedBy       string     `json:"created_by"`
	CreatedDate     time.Time  `json:"created_date"`
	ImageURL        *string    `json:"image_url"` // Pointer to allow NULL
	MappingURL      *string    `json:"mapping_url"` // Pointer to allow NULL
	Explanation     string     `json:"explanation"`
	IsPublic        bool       `json:"is_public"`
	Stars           int        `json:"stars"`
	Flags           int        `json:"flags"`
	UpdatedAt       time.Time  `json:"updated_at"`
	Kanji           *Kanji     `json:"kanji,omitempty"` // For joins
}

// TempCreation represents a temporary draft of a kanji creation
type TempCreation struct {
	TempID      int       `json:"temp_id"`
	KanjiCharID int       `json:"kanji_char_id"`
	ImageURL    *string   `json:"image_url"` // Pointer to allow NULL
	MappingURL  *string   `json:"mapping_url"` // Pointer to allow NULL
	Explanation string    `json:"explanation"`
	CreatedAt   time.Time `json:"created_at"`
	Kanji       *Kanji    `json:"kanji,omitempty"` // For joins
}


////////// Functions for database operations //////////


// AddKanji adds a new kanji to the database
func AddKanji(db *sql.DB, kanji *Kanji) error {
    // Start a transaction
    tx, err := db.Begin()
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer func() {
        if err != nil {
            tx.Rollback()
            return
        }
    }()

    // Insert the kanji
    query := `
        INSERT INTO kanji_go.kanji 
        (kanji_char, romaji_onyomi, romaji_kunyomi, hiragana_onyomi, hiragana_kunyomi, jlpt_level)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING kanji_char_id, created_at, updated_at
    `

    err = tx.QueryRow(
        query, 
        kanji.KanjiChar, 
        kanji.RomajiOnyomi,
        kanji.RomajiKunyomi, 
        kanji.HiraganaOnyomi, 
        kanji.HiraganaKunyomi,
        kanji.JLPTLevel,
    ).Scan(&kanji.KanjiCharID, &kanji.CreatedAt, &kanji.UpdatedAt)

    if err != nil {
        return fmt.Errorf("failed to insert kanji: %w", err)
    }

    // Commit the transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

    return nil
}
