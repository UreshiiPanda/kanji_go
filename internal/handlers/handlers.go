package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/csrf"
)

// HomeHandler handles the root route
func HomeHandler(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := map[string]any{
			"Title":     "Kanji Go",
			"Message":   "Welcome to Kanji Go!",
			"csrfToken": csrf.Token(r), // Add CSRF token for HTMX
		}

		err := tmpl.ExecuteTemplate(w, "base.html", data)
		if err != nil {
			log.Printf("Error executing template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}

// KanjiData represents a kanji character's data
type KanjiData struct {
    ID              int
    KanjiChar       string
    RomajiOnyomi    string
    RomajiKunyomi   string
    HiraganaOnyomi  string
    HiraganaKunyomi string
    JLPTLevel       string
}

// GetKanjiHandler returns all kanji from the database
func GetKanjiHandler(db *sql.DB, tmpl *template.Template) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Query the database
        rows, err := db.Query(`
            SELECT kanji_char_id, kanji_char, romaji_onyomi, romaji_kunyomi, 
                   hiragana_onyomi, hiragana_kunyomi, jlpt_level
            FROM kanji_go.kanji
            ORDER BY kanji_char_id
        `)
        if err != nil {
            log.Printf("Error querying kanji: %v", err)
            http.Error(w, "Failed to retrieve kanji", http.StatusInternalServerError)
            return
        }
        defer rows.Close()

        // Create a slice to hold all kanji data
        var kanjiList []KanjiData

        // Process each row
        for rows.Next() {
            var kanji KanjiData
            if err := rows.Scan(&kanji.ID, &kanji.KanjiChar, &kanji.RomajiOnyomi, 
                               &kanji.RomajiKunyomi, &kanji.HiraganaOnyomi, 
                               &kanji.HiraganaKunyomi, &kanji.JLPTLevel); err != nil {
                log.Printf("Error scanning row: %v", err)
                continue
            }
            kanjiList = append(kanjiList, kanji)
        }

        // Check for errors from iterating over rows
        if err := rows.Err(); err != nil {
            log.Printf("Error iterating rows: %v", err)
        }

        // Prepare template data
        data := map[string]any{
            "KanjiList": kanjiList,
        }

        // Execute the template
        w.Header().Set("Content-Type", "text/html")
        if err := tmpl.ExecuteTemplate(w, "kanji-list", data); err != nil {
            log.Printf("Error executing kanji-list template: %v", err)
            http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        }
    }
}

// GetDialogHandler returns a BeerCSS dialog
func GetDialogHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		dialogHTML := `
		<div class="overlay blur active"></div>
		<dialog class="active">
		  <h5>Custom overlay</h5>
		  <div>Some text here</div>
		  <nav class="right-align no-space">
			<button class="transparent link" hx-get="/empty" hx-target="#dialog-container" hx-swap="innerHTML">Cancel</button>
			<button class="transparent link">Confirm</button>
		  </nav>
		</dialog>
		`
		w.Write([]byte(dialogHTML))
	}
}

// EmptyHandler returns an empty response (used to clear content)
func EmptyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(""))
	}
}
